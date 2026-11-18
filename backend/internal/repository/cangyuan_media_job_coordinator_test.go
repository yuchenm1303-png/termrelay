package repository

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type cangyuanCoordinatorTransport func(*http.Request) (*http.Response, error)

func (f cangyuanCoordinatorTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func cangyuanCoordinatorResponse(payload string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(payload)),
	}
}

func TestCangyuanMediaCoordinatorSubmitsExactlyOnceAndPollsDurably(t *testing.T) {
	svc, _ := newCangyuanMediaJobTestService(t)
	ctx := context.Background()
	owner := service.CangyuanMediaJobOwner{UserID: 17, APIKeyID: 29}
	job, err := svc.Reserve(ctx, owner, 23, service.CangyuanVideoGeneration, "video-model")
	require.NoError(t, err)

	var submitted, polled atomic.Int32
	client := &service.CangyuanMediaClient{
		APIKey: "fixture-only",
		HTTP: &http.Client{Transport: cangyuanCoordinatorTransport(func(req *http.Request) (*http.Response, error) {
			require.Equal(t, "Bearer fixture-only", req.Header.Get("Authorization"))
			switch req.Method {
			case http.MethodPost:
				submitted.Add(1)
				require.Equal(t, "/v1/videos", req.URL.Path)
				return cangyuanCoordinatorResponse(`{"id":"remote_video_9","status":"queued"}`), nil
			case http.MethodGet:
				polled.Add(1)
				require.Equal(t, "/v1/videos/remote_video_9", req.URL.Path)
				return cangyuanCoordinatorResponse(`{"id":"remote_video_9","status":"completed","video_url":"https://cdn.example.com/video.mp4"}`), nil
			default:
				t.Fatalf("unexpected request %s", req.Method)
				return nil, errors.New("unexpected request")
			}
		})},
	}
	payload := []byte(`{"model":"video-model","prompt":"a night skyline","duration":5}`)
	stored, upstream, err := svc.SubmitReserved(ctx, owner, job.ID, client, payload)
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobQueued, stored.Status)
	require.Equal(t, "remote_video_9", upstream.IDValue())
	require.Equal(t, int32(1), submitted.Load())

	_, _, err = svc.SubmitReserved(ctx, owner, job.ID, client, payload)
	require.ErrorIs(t, err, service.ErrCangyuanMediaJobState)
	require.Equal(t, int32(1), submitted.Load(), "duplicate submit must never reach upstream")

	done, err := svc.PollJob(ctx, owner, job.ID, client)
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobCompleted, done.Status)
	require.Equal(t, []string{"https://cdn.example.com/video.mp4"}, done.ResultURLs)
	require.Equal(t, "disabled", done.BillingState)
	require.Equal(t, int32(1), polled.Load())

	again, err := svc.PollJob(ctx, owner, job.ID, client)
	require.NoError(t, err)
	require.Equal(t, done.Version, again.Version)
	require.Equal(t, int32(1), polled.Load(), "terminal jobs must not be polled or billed again")
}

func TestCangyuanMediaCoordinatorTransportUnknownNeverRetries(t *testing.T) {
	svc, _ := newCangyuanMediaJobTestService(t)
	ctx := context.Background()
	owner := service.CangyuanMediaJobOwner{UserID: 12, APIKeyID: 15}
	job, err := svc.Reserve(ctx, owner, 42, service.CangyuanImageGeneration, "image-model")
	require.NoError(t, err)
	var posts atomic.Int32
	client := &service.CangyuanMediaClient{
		APIKey: "fixture-only",
		HTTP: &http.Client{Transport: cangyuanCoordinatorTransport(func(_ *http.Request) (*http.Response, error) {
			posts.Add(1)
			return nil, errors.New("upstream may have accepted the POST")
		})},
	}

	payload := []byte(`{"model":"image-model","prompt":"blue sky"}`)
	_, _, err = svc.SubmitReserved(ctx, owner, job.ID, client, payload)
	require.Error(t, err)
	require.Equal(t, int32(1), posts.Load())
	unknown, err := svc.Get(ctx, owner, job.ID)
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobSubmitUnknown, unknown.Status)
	require.Empty(t, unknown.UpstreamTaskID)

	_, _, err = svc.SubmitReserved(ctx, owner, job.ID, client, payload)
	require.ErrorIs(t, err, service.ErrCangyuanMediaJobState)
	_, err = svc.PollJob(ctx, owner, job.ID, client)
	require.ErrorIs(t, err, service.ErrCangyuanMediaJobState)
	require.Equal(t, int32(1), posts.Load(), "unknown remote task must never be re-submitted")
}

func TestCangyuanMediaCoordinatorValidatesModelBeforeClaim(t *testing.T) {
	svc, _ := newCangyuanMediaJobTestService(t)
	ctx := context.Background()
	owner := service.CangyuanMediaJobOwner{UserID: 2, APIKeyID: 3}
	job, err := svc.Reserve(ctx, owner, 4, service.CangyuanImageGeneration, "approved-model")
	require.NoError(t, err)
	client := &service.CangyuanMediaClient{APIKey: "fixture-only"}
	_, _, err = svc.SubmitReserved(ctx, owner, job.ID, client, []byte(`{"model":"different-model","prompt":"cat"}`))
	require.Error(t, err)
	stored, err := svc.Get(ctx, owner, job.ID)
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobReserved, stored.Status, "bad request must not consume a paid submission")
}

func TestCangyuanMediaPaidSubmissionValidatesQuantityBeforeClaim(t *testing.T) {
	svc, _ := newCangyuanMediaJobTestService(t)
	ctx := context.Background()
	owner := service.CangyuanMediaJobOwner{UserID: 3, APIKeyID: 5}
	quote := cangyuanIdempotencyQuote()
	quote.Tier = "standard"
	job, err := svc.Reserve(ctx, owner, 12, service.CangyuanImageGeneration, "image-model", quote)
	require.NoError(t, err)
	var posts atomic.Int32
	client := &service.CangyuanMediaClient{
		APIKey: "test-fixture",
		HTTP: &http.Client{Transport: cangyuanCoordinatorTransport(func(_ *http.Request) (*http.Response, error) {
			posts.Add(1)
			return nil, errors.New("unexpected paid request")
		})},
	}
	_, _, err = svc.SubmitReserved(ctx, owner, job.ID, client, []byte(`{"model":"image-model","prompt":"two cats","n":2}`))
	require.ErrorContains(t, err, "does not match approved price")
	require.Equal(t, int32(0), posts.Load(), "underquoted media must never reach upstream")
	stored, err := svc.Get(ctx, owner, job.ID)
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobReserved, stored.Status)
}

func TestCangyuanMediaPricedSubmissionRejectsAlteredPayloadBeforeClaim(t *testing.T) {
	svc, _ := newCangyuanMediaJobTestService(t)
	ctx := context.Background()
	owner := service.CangyuanMediaJobOwner{UserID: 52, APIKeyID: 61}
	approved := []byte(`{"model":"image-model","prompt":"one kitten","n":1,"size":"1024x1024"}`)
	quote := cangyuanIdempotencyQuote()
	job, err := svc.ReserveIdempotent(ctx, owner, 14, service.CangyuanImageGeneration,
		"image-model", approved, "locked-payload-1", quote)
	require.NoError(t, err)

	var posts atomic.Int32
	client := &service.CangyuanMediaClient{
		APIKey: "test-fixture",
		HTTP: &http.Client{Transport: cangyuanCoordinatorTransport(func(_ *http.Request) (*http.Response, error) {
			posts.Add(1)
			return nil, errors.New("unapproved paid submission reached upstream")
		})},
	}
	// Same model and n, but a different priced dimension and prompt.
	mutated := []byte(`{"model":"image-model","prompt":"one kitten in 4K","n":1,"size":"4096x4096"}`)
	_, _, err = svc.SubmitReserved(ctx, owner, job.ID, client, mutated)
	require.ErrorContains(t, err, "differs from the reserved priced request")
	require.Zero(t, posts.Load(), "an altered paid request must not reach the upstream")
	stored, err := svc.Get(ctx, owner, job.ID)
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobReserved, stored.Status)

	// The approved payload is still eligible: a failed local validation has
	// not consumed the one-and-only submission claim.
	_, _, err = svc.SubmitReserved(ctx, owner, job.ID, client, approved)
	require.Error(t, err) // test transport intentionally rejects outbound POST
	require.Equal(t, int32(1), posts.Load())
}
