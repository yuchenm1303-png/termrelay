package repository

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type cangyuanWorkerAccountReader struct {
	account *service.Account
}

func (r *cangyuanWorkerAccountReader) GetByID(_ context.Context, id int64) (*service.Account, error) {
	if r.account != nil && r.account.ID == id {
		return r.account, nil
	}
	return nil, service.ErrCangyuanMediaJobNotFound
}

func testCangyuanRecoveryAccount() *service.Account {
	return &service.Account{
		ID: 7, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey,
		Status: service.StatusActive, Schedulable: true,
		Credentials: map[string]any{
			"api_key": "no-network-test-token",
			"base_url": "https://ai.cangyuansuanli.cn",
		},
		Extra: map[string]any{
			"upstream_vendor": "cangyuan",
			"cangyuan_media_only": true,
			"cangyuan_media_models": map[string]any{"image": []any{"image-model"}},
			"openai_responses_mode": "force_chat_completions",
		},
	}
}

func TestCangyuanRecoveryWorkerPollsHeldJobFromRedisIndex(t *testing.T) {
	ctx := context.Background()
	svc, mr := newCangyuanMediaJobTestService(t)
	owner := service.CangyuanMediaJobOwner{UserID: 81, APIKeyID: 39}
	payload := []byte(`{"model":"image-model","prompt":"test sunrise","n":1}`)
	job, err := svc.ReserveIdempotent(ctx, owner, 7, service.CangyuanImageGeneration,
		"image-model", payload, "worker-key", cangyuanIdempotencyQuote())
	require.NoError(t, err)
	billing := &cangyuanEscrowLedgerStub{}
	_, _, err = svc.HoldApprovedCangyuanMediaBalance(ctx, owner, job.ID, billing)
	require.NoError(t, err)
	_, err = svc.ClaimSubmission(ctx, owner, job.ID)
	require.NoError(t, err)
	_, err = svc.AttachSubmitted(ctx, owner, job.ID,
		&service.CangyuanMediaTask{ID: "remote-worker-1", Status: "queued"})
	require.NoError(t, err)

	queue := svc.Store.(service.CangyuanMediaRecoveryQueue)
	require.NoError(t, queue.Delay(ctx, job.ID, 0))
	worker := service.NewCangyuanMediaRecoveryWorker(svc,
		&cangyuanWorkerAccountReader{account: testCangyuanRecoveryAccount()}, billing)
	worker.ClientFactory = func(_ *service.Account) *service.CangyuanMediaClient {
		return &service.CangyuanMediaClient{
			APIKey: "fake",
			HTTP: &http.Client{Transport: cangyuanCoordinatorTransport(func(req *http.Request) (*http.Response, error) {
				require.Equal(t, http.MethodGet, req.Method)
				require.Equal(t, "/v1/images/generations/remote-worker-1", req.URL.Path)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{"id":"remote-worker-1","status":"completed","data":[{"url":"https://cdn.example.com/sunrise.png"}]}`)),
					Header: make(http.Header),
				}, nil
			})},
		}
	}
	require.NoError(t, worker.RunOnce(ctx))
	current, err := svc.Get(ctx, owner, job.ID)
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobCompleted, current.Status)
	require.Equal(t, "settled", current.BillingState)
	require.Equal(t, 1, billing.captureCalls)
	require.Equal(t, 1, billing.applied)
	require.False(t, mr.Exists(cangyuanMediaActiveIndex))
	require.True(t, mr.TTL(cangyuanMediaJobKeyPrefix+job.ID) > 0)
	require.NoError(t, worker.RunOnce(ctx))
	require.Equal(t, 1, billing.captureCalls)
}

func TestCangyuanRecoveryWorkerNeverResubmitsUnknownTask(t *testing.T) {
	ctx := context.Background()
	svc, mr := newCangyuanMediaJobTestService(t)
	owner := service.CangyuanMediaJobOwner{UserID: 81, APIKeyID: 39}
	payload := []byte(`{"model":"image-model","prompt":"unknown transport","n":1}`)
	job, err := svc.ReserveIdempotent(ctx, owner, 7, service.CangyuanImageGeneration,
		"image-model", payload, "worker-unknown-key", cangyuanIdempotencyQuote())
	require.NoError(t, err)
	billing := &cangyuanEscrowLedgerStub{}
	_, _, err = svc.HoldApprovedCangyuanMediaBalance(ctx, owner, job.ID, billing)
	require.NoError(t, err)
	_, err = svc.ClaimSubmission(ctx, owner, job.ID)
	require.NoError(t, err)
	_, err = svc.MarkSubmissionUnknown(ctx, owner, job.ID)
	require.NoError(t, err)
	queue := svc.Store.(service.CangyuanMediaRecoveryQueue)
	require.NoError(t, queue.Delay(ctx, job.ID, 0))
	worker := service.NewCangyuanMediaRecoveryWorker(svc,
		&cangyuanWorkerAccountReader{account: testCangyuanRecoveryAccount()}, billing)
	worker.ClientFactory = func(_ *service.Account) *service.CangyuanMediaClient {
		t.Fatal("ambiguous POST must never be retried")
		return nil
	}
	require.NoError(t, worker.RunOnce(ctx))
	status, err := svc.Get(ctx, owner, job.ID)
	require.NoError(t, err)
	require.Equal(t, "held", status.BillingState)
	require.Equal(t, service.CangyuanMediaJobSubmitUnknown, status.Status)
	require.Zero(t, billing.captureCalls)
	require.Zero(t, billing.releaseCalls)
	require.Equal(t, time.Duration(0), mr.TTL(cangyuanMediaJobKeyPrefix+job.ID))
}
