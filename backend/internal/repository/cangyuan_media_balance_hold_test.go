package repository

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

// All payments here are simulated locally; the stub enforces SQL-like dedup.
type cangyuanEscrowLedgerStub struct {
	cangyuanBillingStub
	muHold          sync.Mutex
	holdRequests    map[string]bool
	captureRequests map[string]bool
	releaseRequests map[string]bool
	holdCalls       int
	captureCalls    int
	releaseCalls    int
	lastHold        *service.BatchImageBalanceHoldCommand
	lastCapture     *service.BatchImageBalanceHoldCommand
	failHold        bool
}

func (s *cangyuanEscrowLedgerStub) applyHold(cmd *service.BatchImageBalanceHoldCommand, which string) (*service.BatchImageBalanceHoldResult, error) {
	s.muHold.Lock()
	defer s.muHold.Unlock()
	if s.failHold && which == "hold" {
		return nil, service.ErrBatchImageInsufficientBalance
	}
	var m *map[string]bool
	switch which {
	case "hold":
		m = &s.holdRequests
		s.holdCalls++
		s.lastHold = cmd
	case "capture":
		m = &s.captureRequests
		s.captureCalls++
		s.lastCapture = cmd
	case "release":
		m = &s.releaseRequests
		s.releaseCalls++
	}
	if *m == nil {
		*m = make(map[string]bool)
	}
	if (*m)[cmd.RequestID] {
		return &service.BatchImageBalanceHoldResult{Applied: false}, nil
	}
	(*m)[cmd.RequestID] = true
	return &service.BatchImageBalanceHoldResult{Applied: true}, nil
}

func (s *cangyuanEscrowLedgerStub) ReserveBatchImageBalance(_ context.Context, cmd *service.BatchImageBalanceHoldCommand) (*service.BatchImageBalanceHoldResult, error) {
	return s.applyHold(cmd, "hold")
}
func (s *cangyuanEscrowLedgerStub) CaptureBatchImageBalance(_ context.Context, cmd *service.BatchImageBalanceHoldCommand) (*service.BatchImageBalanceHoldResult, error) {
	return s.applyHold(cmd, "capture")
}
func (s *cangyuanEscrowLedgerStub) ReleaseBatchImageBalance(_ context.Context, cmd *service.BatchImageBalanceHoldCommand) (*service.BatchImageBalanceHoldResult, error) {
	return s.applyHold(cmd, "release")
}

func pricedCangyuanEscrowFixture(t *testing.T) (*service.CangyuanMediaJobService, service.CangyuanMediaJobOwner, string) {
	t.Helper()
	svc, _ := newCangyuanMediaJobTestService(t)
	owner := service.CangyuanMediaJobOwner{UserID: 81, APIKeyID: 39}
	payload := []byte(`{"model":"image-model","prompt":"a small harbor","n":1}`)
	job, err := svc.ReserveIdempotent(context.Background(), owner, 7, service.CangyuanImageGeneration,
		"image-model", payload, "paid-test-1", cangyuanIdempotencyQuote())
	require.NoError(t, err)
	return svc, owner, job.ID
}

func TestCangyuanEscrowReservesOnceThenCapturesWithoutDoubleChargingBalance(t *testing.T) {
	ctx := context.Background()
	svc, owner, id := pricedCangyuanEscrowFixture(t)
	billing := &cangyuanEscrowLedgerStub{}

	_, _, err := svc.SettleHeldCangyuanMedia(ctx, owner, id, billing)
	require.ErrorIs(t, err, service.ErrCangyuanMediaNotBillable)
	require.Zero(t, billing.captureCalls)

	job, hold, err := svc.HoldApprovedCangyuanMediaBalance(ctx, owner, id, billing)
	require.NoError(t, err)
	require.True(t, hold.Applied)
	require.Equal(t, "held", job.BillingState)
	require.InDelta(t, 2.5, billing.lastHold.HoldAmount, 0.000001)

	job, hold, err = svc.HoldApprovedCangyuanMediaBalance(ctx, owner, id, billing)
	require.NoError(t, err)
	require.False(t, hold.Applied)
	require.Equal(t, "held", job.BillingState)
	require.Equal(t, 1, billing.holdCalls, "must not make another SQL hold")

	_, err = svc.ClaimSubmission(ctx, owner, id)
	require.NoError(t, err)
	_, err = svc.AttachSubmitted(ctx, owner, id, &service.CangyuanMediaTask{
		ID: "remote_escrow_1", Status: "completed",
		Data: []struct {
			URL string `json:"url"`
		}{{URL: "https://cdn.example.com/one.png"}},
	})
	require.NoError(t, err)

	paid, applied, err := svc.SettleHeldCangyuanMedia(ctx, owner, id, billing)
	require.NoError(t, err)
	require.True(t, applied.Applied)
	require.Equal(t, "settled", paid.BillingState)
	require.Equal(t, 1, billing.captureCalls)
	require.Equal(t, 1, billing.applied)
	require.InDelta(t, 2.5, billing.lastCapture.ActualAmount, 0.000001)
	require.Zero(t, billing.lastCommand.BalanceCost, "capture already consumed held funds")
	require.InDelta(t, 2.5, billing.lastCommand.APIKeyQuotaCost, 0.000001)
	require.InDelta(t, 1, billing.lastCommand.AccountQuotaCost, 0.000001)

	_, applied, err = svc.SettleHeldCangyuanMedia(ctx, owner, id, billing)
	require.NoError(t, err)
	require.False(t, applied.Applied)
	require.Equal(t, 1, billing.captureCalls)
	require.Equal(t, 1, billing.applied)
}

func TestCangyuanEscrowRetriesAfterQuotaFailureWithoutSecondCapture(t *testing.T) {
	ctx := context.Background()
	svc, owner, id := pricedCangyuanEscrowFixture(t)
	billing := &cangyuanEscrowLedgerStub{}
	_, _, err := svc.HoldApprovedCangyuanMediaBalance(ctx, owner, id, billing)
	require.NoError(t, err)
	_, err = svc.ClaimSubmission(ctx, owner, id)
	require.NoError(t, err)
	_, err = svc.AttachSubmitted(ctx, owner, id, &service.CangyuanMediaTask{
		ID: "remote_retry_1", Status: "completed",
		Data: []struct{ URL string `json:"url"` }{{URL: "https://cdn.example.com/result.png"}},
	})
	require.NoError(t, err)

	billing.failNext = true
	_, _, err = svc.SettleHeldCangyuanMedia(ctx, owner, id, billing)
	require.ErrorContains(t, err, "temporary SQL transaction failure")
	stored, err := svc.Get(ctx, owner, id)
	require.NoError(t, err)
	require.Equal(t, "held", stored.BillingState)

	_, _, err = svc.SettleHeldCangyuanMedia(ctx, owner, id, billing)
	require.NoError(t, err)
	require.Equal(t, 1, len(billing.captureRequests), "SQL capture request ID is stable")
	require.Equal(t, 1, billing.applied)
}

func TestCangyuanEscrowRejectsInsufficientBalanceAndReleasesFailedJob(t *testing.T) {
	ctx := context.Background()
	svc, owner, id := pricedCangyuanEscrowFixture(t)
	billing := &cangyuanEscrowLedgerStub{failHold: true}
	_, _, err := svc.HoldApprovedCangyuanMediaBalance(ctx, owner, id, billing)
	require.ErrorIs(t, err, service.ErrBatchImageInsufficientBalance)
	job, err := svc.Get(ctx, owner, id)
	require.NoError(t, err)
	require.Equal(t, "pending", job.BillingState)

	billing.failHold = false
	_, _, err = svc.HoldApprovedCangyuanMediaBalance(ctx, owner, id, billing)
	require.NoError(t, err)
	_, err = svc.ClaimSubmission(ctx, owner, id)
	require.NoError(t, err)
	_, err = svc.AttachSubmitted(ctx, owner, id, &service.CangyuanMediaTask{
		ID: "remote_failed_1", Status: "failed",
	})
	require.NoError(t, err)
	job, released, err := svc.ReleaseFailedCangyuanMediaBalance(ctx, owner, id, billing)
	require.NoError(t, err)
	require.True(t, released.Applied)
	require.Equal(t, "released", job.BillingState)
	require.Equal(t, 1, billing.releaseCalls)
	_, released, err = svc.ReleaseFailedCangyuanMediaBalance(ctx, owner, id, billing)
	require.NoError(t, err)
	require.False(t, released.Applied)
	require.Equal(t, 1, billing.releaseCalls)
	_, _, err = svc.SettleHeldCangyuanMedia(ctx, owner, id, billing)
	require.ErrorIs(t, err, service.ErrCangyuanMediaNotBillable)
}

func TestCangyuanEscrowRequiresPinnedPricedRequestAndTerminalFailure(t *testing.T) {
	ctx := context.Background()
	svc, owner, id := cangyuanSettlementFixture(t, true)
	billing := &cangyuanEscrowLedgerStub{}
	_, _, err := svc.HoldApprovedCangyuanMediaBalance(ctx, owner, id, billing)
	require.ErrorIs(t, err, service.ErrCangyuanMediaNotBillable)
	require.Zero(t, billing.holdCalls)
	_, _, err = svc.ReleaseFailedCangyuanMediaBalance(ctx, owner, id, billing)
	require.ErrorIs(t, err, service.ErrCangyuanMediaNotBillable)
	require.True(t, errors.Is(err, service.ErrCangyuanMediaNotBillable))
}
