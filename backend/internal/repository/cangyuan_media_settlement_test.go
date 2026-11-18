package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type cangyuanBillingStub struct {
	mu          sync.Mutex
	calls       int
	applied     int
	failNext    bool
	requests    map[string]string
	lastCommand *service.UsageBillingCommand
}

func (b *cangyuanBillingStub) Apply(_ context.Context, cmd *service.UsageBillingCommand) (*service.UsageBillingApplyResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.calls++
	if b.failNext {
		b.failNext = false
		return nil, errors.New("temporary SQL transaction failure")
	}
	if b.requests == nil {
		b.requests = make(map[string]string)
	}
	cmd.Normalize()
	key := fmt.Sprintf("%d:%s", cmd.APIKeyID, cmd.RequestID)
	if hash, exists := b.requests[key]; exists {
		if hash != cmd.RequestFingerprint {
			return nil, service.ErrUsageBillingRequestConflict
		}
		return &service.UsageBillingApplyResult{Applied: false}, nil
	}
	b.requests[key] = cmd.RequestFingerprint
	b.lastCommand = cmd
	b.applied++
	return &service.UsageBillingApplyResult{Applied: true}, nil
}

func cangyuanSettlementFixture(t *testing.T, priced bool) (*service.CangyuanMediaJobService, service.CangyuanMediaJobOwner, string) {
	t.Helper()
	svc, _ := newCangyuanMediaJobTestService(t)
	owner := service.CangyuanMediaJobOwner{UserID: 51, APIKeyID: 17}
	var (
		job *service.CangyuanMediaJob
		err error
	)
	if priced {
		job, err = svc.Reserve(context.Background(), owner, 112, service.CangyuanImageGeneration, "image-approved",
			service.CangyuanMediaPriceSnapshot{
				PricingVersion:      "2026-11-18-admin-reviewed",
				Currency:            "USD",
				Tier:                "standard-1024",
				Units:               2,
				UnitSalePriceUSD:    "0.15",
				UnitProviderCostUSD: "0.04",
				VerifiedByAdmin:     true,
			},
		)
	} else {
		job, err = svc.Reserve(context.Background(), owner, 112, service.CangyuanImageGeneration, "image-approved")
	}
	require.NoError(t, err)
	return svc, owner, job.ID
}

func attachCangyuanCompletedMediaTask(t *testing.T, svc *service.CangyuanMediaJobService, owner service.CangyuanMediaJobOwner, id string) {
	t.Helper()
	ctx := context.Background()
	_, err := svc.ClaimSubmission(ctx, owner, id)
	require.NoError(t, err)
	var completed service.CangyuanMediaTask
	err = json.Unmarshal([]byte(`{"id":"upstream_task_1","status":"completed","data":[{"url":"https://cdn.example.com/one.png"},{"url":"https://cdn.example.com/two.png"}]}`), &completed)
	require.NoError(t, err)
	_, err = svc.AttachSubmitted(ctx, owner, id, &completed)
	require.NoError(t, err)
}

func TestCangyuanMediaSettlementChargesOnlyOnceAfterVerifiedCompletion(t *testing.T) {
	svc, owner, id := cangyuanSettlementFixture(t, true)
	billing := &cangyuanBillingStub{}
	ctx := context.Background()

	_, _, err := svc.SettleCompleted(ctx, owner, id, billing)
	require.ErrorIs(t, err, service.ErrCangyuanMediaNotBillable)
	require.Zero(t, billing.calls)
	attachCangyuanCompletedMediaTask(t, svc, owner, id)

	job, result, err := svc.SettleCompleted(ctx, owner, id, billing)
	require.NoError(t, err)
	require.True(t, result.Applied)
	require.Equal(t, "settled", job.BillingState)
	require.Equal(t, 1, billing.calls)
	require.Equal(t, 1, billing.applied)
	require.Equal(t, "cangyuan-media:"+id, billing.lastCommand.RequestID)
	require.Equal(t, int64(51), billing.lastCommand.UserID)
	require.Equal(t, int64(17), billing.lastCommand.APIKeyID)
	require.Equal(t, int64(112), billing.lastCommand.AccountID)
	require.Equal(t, "image-approved", billing.lastCommand.Model)
	require.Equal(t, "image", billing.lastCommand.MediaType)
	require.Equal(t, 2, billing.lastCommand.ImageCount)
	require.InDelta(t, 0.30, billing.lastCommand.BalanceCost, 0.000000001)
	require.InDelta(t, 0.08, billing.lastCommand.AccountQuotaCost, 0.000000001)
	require.InDelta(t, 0.30, billing.lastCommand.APIKeyQuotaCost, 0.000000001)
	require.NotEmpty(t, billing.lastCommand.RequestFingerprint)

	again, secondResult, err := svc.SettleCompleted(ctx, owner, id, billing)
	require.NoError(t, err)
	require.False(t, secondResult.Applied)
	require.Equal(t, job.Version, again.Version)
	require.Equal(t, 1, billing.applied, "a completed media task must never debit twice")
}

func TestCangyuanMediaSettlementResumesAfterSQLFailure(t *testing.T) {
	svc, owner, id := cangyuanSettlementFixture(t, true)
	attachCangyuanCompletedMediaTask(t, svc, owner, id)
	billing := &cangyuanBillingStub{failNext: true}
	_, _, err := svc.SettleCompleted(context.Background(), owner, id, billing)
	require.ErrorContains(t, err, "temporary SQL transaction failure")

	pending, err := svc.Get(context.Background(), owner, id)
	require.NoError(t, err)
	require.Equal(t, "pending", pending.BillingState)
	_, result, err := svc.SettleCompleted(context.Background(), owner, id, billing)
	require.NoError(t, err)
	require.True(t, result.Applied)
	require.Equal(t, 1, billing.applied)
}

type cangyuanSettlementStoreWithOneCASFailure struct {
	service.CangyuanMediaJobStore
	once sync.Once
}

func (s *cangyuanSettlementStoreWithOneCASFailure) CompareAndSwap(ctx context.Context, job *service.CangyuanMediaJob, previous int64) error {
	if job.BillingState == "settled" {
		fail := false
		s.once.Do(func() { fail = true })
		if fail {
			return errors.New("redis write temporarily unavailable")
		}
	}
	return s.CangyuanMediaJobStore.CompareAndSwap(ctx, job, previous)
}

func TestCangyuanMediaSettlementSQLDedupSurvivesRedisFailure(t *testing.T) {
	svc, owner, id := cangyuanSettlementFixture(t, true)
	attachCangyuanCompletedMediaTask(t, svc, owner, id)
	svc.Store = &cangyuanSettlementStoreWithOneCASFailure{CangyuanMediaJobStore: svc.Store}
	billing := &cangyuanBillingStub{}
	_, result, err := svc.SettleCompleted(context.Background(), owner, id, billing)
	require.ErrorContains(t, err, "Redis settlement marker")
	require.True(t, result.Applied)
	require.Equal(t, 1, billing.applied)

	latest, secondResult, err := svc.SettleCompleted(context.Background(), owner, id, billing)
	require.NoError(t, err)
	require.False(t, secondResult.Applied)
	require.Equal(t, "settled", latest.BillingState)
	require.Equal(t, 1, billing.applied)
}

func TestCangyuanMediaSettlementRefusesUnpricedFailedOrTamperedOutput(t *testing.T) {
	svc, owner, id := cangyuanSettlementFixture(t, false)
	attachCangyuanCompletedMediaTask(t, svc, owner, id)
	billing := &cangyuanBillingStub{}
	_, _, err := svc.SettleCompleted(context.Background(), owner, id, billing)
	require.ErrorIs(t, err, service.ErrCangyuanMediaNotBillable)
	require.Zero(t, billing.calls)

	svc2, owner2, id2 := cangyuanSettlementFixture(t, true)
	_, err = svc2.ClaimSubmission(context.Background(), owner2, id2)
	require.NoError(t, err)
	_, err = svc2.AttachSubmitted(context.Background(), owner2, id2, &service.CangyuanMediaTask{
		ID: "task_failed", Status: "failed",
	})
	require.NoError(t, err)
	_, _, err = svc2.SettleCompleted(context.Background(), owner2, id2, billing)
	require.ErrorIs(t, err, service.ErrCangyuanMediaNotBillable)
	require.Zero(t, billing.calls)
}

func TestCangyuanMediaReservationRequiresExplicitApprovedPrice(t *testing.T) {
	svc, owner, _ := cangyuanSettlementFixture(t, false)
	quote := service.CangyuanMediaPriceSnapshot{
		PricingVersion:      "admin-reviewed-v1",
		Currency:            "USD",
		Tier:                "default",
		Units:               1,
		UnitSalePriceUSD:    "1.23",
		UnitProviderCostUSD: "0.67",
		VerifiedByAdmin:     true,
	}
	ctx := context.Background()
	cases := []struct {
		name   string
		mutate func(*service.CangyuanMediaPriceSnapshot)
	}{
		{"unverified", func(q *service.CangyuanMediaPriceSnapshot) { q.VerifiedByAdmin = false }},
		{"missing version", func(q *service.CangyuanMediaPriceSnapshot) { q.PricingVersion = "" }},
		{"wrong currency", func(q *service.CangyuanMediaPriceSnapshot) { q.Currency = "CNY" }},
		{"negative price", func(q *service.CangyuanMediaPriceSnapshot) { q.UnitSalePriceUSD = "-1" }},
		{"nan", func(q *service.CangyuanMediaPriceSnapshot) { q.UnitSalePriceUSD = "NaN" }},
		{"infinite", func(q *service.CangyuanMediaPriceSnapshot) { q.UnitSalePriceUSD = fmt.Sprintf("%v", math.Inf(1)) }},
		{"too much precision", func(q *service.CangyuanMediaPriceSnapshot) { q.UnitSalePriceUSD = "0.123456789" }},
		{"zero sale", func(q *service.CangyuanMediaPriceSnapshot) { q.UnitSalePriceUSD = "0" }},
		{"missing provider cost", func(q *service.CangyuanMediaPriceSnapshot) { q.UnitProviderCostUSD = "" }},
		{"missing units", func(q *service.CangyuanMediaPriceSnapshot) { q.Units = 0 }},
		{"too many units", func(q *service.CangyuanMediaPriceSnapshot) { q.Units = 17 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bad := quote
			tc.mutate(&bad)
			_, err := svc.Reserve(ctx, owner, 2, service.CangyuanImageGeneration, "image-approved", bad)
			require.Error(t, err)
		})
	}
	_, err := svc.Reserve(ctx, owner, 2, service.CangyuanVideoGeneration, "video-approved", service.CangyuanMediaPriceSnapshot{
		PricingVersion:      "admin-reviewed-v1",
		Currency:            "USD",
		Tier:                "720p-5s",
		Units:               2,
		UnitSalePriceUSD:    "1.23",
		UnitProviderCostUSD: "0.67",
		VerifiedByAdmin:     true,
	})
	require.Error(t, err, "one video job cannot be billed as two outputs")
}

func TestCangyuanMediaSettlementRejectsPriceTamperingInRedis(t *testing.T) {
	svc, owner, id := cangyuanSettlementFixture(t, true)
	ctx := context.Background()
	job, err := svc.Get(ctx, owner, id)
	require.NoError(t, err)
	require.NotNil(t, job.PriceSnapshot)

	altered := *job
	copiedQuote := *job.PriceSnapshot
	copiedQuote.UnitSalePriceUSD = "0.000001"
	altered.PriceSnapshot = &copiedQuote
	altered.Version++
	require.ErrorIs(t, svc.Store.CompareAndSwap(ctx, &altered, job.Version), service.ErrCangyuanMediaJobConflict)

	altered = *job
	altered.PriceSnapshot = nil
	altered.Version++
	require.ErrorIs(t, svc.Store.CompareAndSwap(ctx, &altered, job.Version), service.ErrCangyuanMediaJobConflict)

	updated, err := svc.Get(ctx, owner, id)
	require.NoError(t, err)
	require.Equal(t, "0.15", updated.PriceSnapshot.UnitSalePriceUSD)
}
