package repository

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type cangyuanPublicAccountLister struct {
	accounts []service.Account
	groupID  int64
}

func (s *cangyuanPublicAccountLister) ListSchedulableByGroupIDAndPlatform(
	_ context.Context, groupID int64, platform string,
) ([]service.Account, error) {
	if groupID != s.groupID || platform != service.PlatformOpenAI {
		return nil, nil
	}
	return s.accounts, nil
}

type cangyuanPublicPriceSettings struct {
	values map[string]string
}

func (s *cangyuanPublicPriceSettings) Set(_ context.Context, key, value string) error {
	s.values[key] = value
	return nil
}
func (s *cangyuanPublicPriceSettings) GetAll(_ context.Context) (map[string]string, error) {
	out := make(map[string]string, len(s.values))
	for key, value := range s.values {
		out[key] = value
	}
	return out, nil
}
func (s *cangyuanPublicPriceSettings) Delete(_ context.Context, key string) error {
	delete(s.values, key)
	return nil
}

func cangyuanPublicTestService(t *testing.T) (
	*service.CangyuanMediaPublicService, *cangyuanEscrowLedgerStub, *atomic.Int32,
) {
	t.Helper()
	jobs, _ := newCangyuanMediaJobTestService(t)
	prices := service.NewCangyuanMediaPriceRegistry(
		&cangyuanPublicPriceSettings{values: make(map[string]string)})
	rule := service.CangyuanMediaPriceRule{
		ID: "single-image-test", Model: "image-model", Kind: service.CangyuanImageGeneration,
		Options: map[string]json.RawMessage{"n": json.RawMessage("1")},
		Snapshot: cangyuanIdempotencyQuote(),
		Enabled: true,
	}
	require.NoError(t, prices.Put(context.Background(), rule))
	account := testCangyuanRecoveryAccount()
	lister := &cangyuanPublicAccountLister{
		accounts: []service.Account{*account}, groupID: 3,
	}
	billing := &cangyuanEscrowLedgerStub{}
	svc := &service.CangyuanMediaPublicService{
		Jobs: jobs, Prices: prices, Accounts: lister, Billing: billing,
	}
	var postCalls atomic.Int32
	svc.ClientFactory = func(_ *service.Account) *service.CangyuanMediaClient {
		return &service.CangyuanMediaClient{
			APIKey: "fake",
			HTTP: &http.Client{Transport: cangyuanCoordinatorTransport(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost {
					return nil, errors.New("public submission must not poll synchronously")
				}
				postCalls.Add(1)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`{"id":"upstream_public_test","status":"queued"}`)),
					Header: make(http.Header),
				}, nil
			})},
		}
	}
	return svc, billing, &postCalls
}

func TestCangyuanPublicMediaSubmissionHoldsBalanceBeforeExactlyOnePOST(t *testing.T) {
	svc, billing, posts := cangyuanPublicTestService(t)
	ctx := context.Background()
	owner := service.CangyuanMediaJobOwner{UserID: 101, APIKeyID: 202}
	body := []byte(`{"model":"image-model","prompt":"ocean waves","n":1}`)

	job, err := svc.Submit(ctx, owner, 3, service.CangyuanImageGeneration,
		"image-model", body, "same-public-key")
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobQueued, job.Status)
	require.Equal(t, "held", job.BillingState)
	require.Equal(t, int32(1), posts.Load())
	require.Equal(t, 1, billing.holdCalls)
	require.InDelta(t, 2.5, billing.lastHold.HoldAmount, 0.000001)

	again, err := svc.Submit(ctx, owner, 3, service.CangyuanImageGeneration,
		"image-model", body, "same-public-key")
	require.NoError(t, err)
	require.Equal(t, job.ID, again.ID)
	require.Equal(t, int32(1), posts.Load(), "same client key cannot duplicate a paid generation")
	require.Equal(t, 1, billing.holdCalls, "same client key cannot freeze balance twice")

	_, err = svc.Submit(ctx, owner, 3, service.CangyuanImageGeneration,
		"image-model", []byte(`{"model":"image-model","prompt":"different waves","n":1}`),
		"same-public-key")
	require.ErrorIs(t, err, service.ErrCangyuanMediaIdempotencyConflict)
	require.Equal(t, int32(1), posts.Load())
}

func TestCangyuanPublicMediaSubmissionRejectsUnpricedOrUnfundedRequests(t *testing.T) {
	svc, billing, posts := cangyuanPublicTestService(t)
	ctx := context.Background()
	owner := service.CangyuanMediaJobOwner{UserID: 101, APIKeyID: 202}
	payload := []byte(`{"model":"image-model","prompt":"marina","n":1}`)

	_, err := svc.Submit(ctx, owner, 99, service.CangyuanImageGeneration,
		"image-model", payload, "wrong-group")
	require.ErrorIs(t, err, service.ErrCangyuanMediaNoAccount)

	_, err = svc.Submit(ctx, owner, 3, service.CangyuanImageGeneration,
		"image-model", []byte(`{"model":"image-model","prompt":"marina","n":1,"size":"4096x4096"}`),
		"unpriced-4k")
	require.ErrorIs(t, err, service.ErrCangyuanMediaPriceNotFound)
	require.Zero(t, posts.Load())
	require.Zero(t, billing.holdCalls)

	billing.failHold = true
	pending, err := svc.Submit(ctx, owner, 3, service.CangyuanImageGeneration,
		"image-model", payload, "balance-empty")
	require.ErrorIs(t, err, service.ErrBatchImageInsufficientBalance)
	require.NotNil(t, pending)
	require.Equal(t, "pending", pending.BillingState)
	require.Zero(t, posts.Load(), "funding must succeed before any paid POST")

	billing.failHold = false
	queued, err := svc.Submit(ctx, owner, 3, service.CangyuanImageGeneration,
		"image-model", payload, "balance-empty")
	require.NoError(t, err)
	require.Equal(t, pending.ID, queued.ID)
	require.Equal(t, "held", queued.BillingState)
	require.Equal(t, int32(1), posts.Load())
}

func TestCangyuanPublicMediaDefinitiveRejectionRefundsWithoutResubmission(t *testing.T) {
	svc, billing, _ := cangyuanPublicTestService(t)
	ctx := context.Background()
	owner := service.CangyuanMediaJobOwner{UserID: 101, APIKeyID: 202}
	var postCalls atomic.Int32
	svc.ClientFactory = func(_ *service.Account) *service.CangyuanMediaClient {
		return &service.CangyuanMediaClient{
			APIKey: "fake",
			HTTP: &http.Client{Transport: cangyuanCoordinatorTransport(func(req *http.Request) (*http.Response, error) {
				require.Equal(t, http.MethodPost, req.Method)
				postCalls.Add(1)
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body: io.NopCloser(strings.NewReader(`{"error":"bad prompt"}`)),
					Header: make(http.Header),
				}, nil
			})},
		}
	}
	payload := []byte(`{"model":"image-model","prompt":"invalid upstream request","n":1}`)
	job, err := svc.Submit(ctx, owner, 3, service.CangyuanImageGeneration,
		"image-model", payload, "definitely-rejected")
	var rejected *service.CangyuanMediaRejectedError
	require.ErrorAs(t, err, &rejected)
	require.Equal(t, http.StatusBadRequest, rejected.StatusCode)
	require.NotNil(t, job)
	require.Equal(t, service.CangyuanMediaJobFailed, job.Status)
	require.Equal(t, "released", job.BillingState)
	require.Equal(t, 1, billing.holdCalls)
	require.Equal(t, 1, billing.releaseCalls)
	require.Equal(t, int32(1), postCalls.Load())
	again, err := svc.Submit(ctx, owner, 3, service.CangyuanImageGeneration,
		"image-model", payload, "definitely-rejected")
	require.NoError(t, err)
	require.Equal(t, job.ID, again.ID)
	require.Equal(t, int32(1), postCalls.Load())
	require.Equal(t, 1, billing.releaseCalls)
}
