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

// Fake transport intentionally rejects POST: reconciliation is GET-only and
// must not duplicate the generation even when a remote query fails.
func cangyuanReconciliationClient(t *testing.T, responses ...string) (*service.CangyuanMediaClient, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	client := &service.CangyuanMediaClient{
		APIKey: "fake-test-key",
		HTTP: &http.Client{Transport: cangyuanCoordinatorTransport(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodGet {
				return nil, errors.New("reconciler must not issue upstream POST")
			}
			if !strings.HasSuffix(req.URL.Path, "/remote_reconcile_task") {
				return nil, errors.New("unexpected upstream task path")
			}
			n := int(calls.Add(1)) - 1
			if n >= len(responses) {
				return nil, errors.New("unexpected additional poll")
			}
			if responses[n] == "transport_failure" {
				return nil, errors.New("simulated network failure")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(responses[n])),
				Header:     make(http.Header),
			}, nil
		})},
	}
	return client, &calls
}

func setupCangyuanReconciliation(t *testing.T) (*service.CangyuanMediaJobService, service.CangyuanMediaJobOwner, string, *cangyuanEscrowLedgerStub) {
	t.Helper()
	svc, owner, id := pricedCangyuanEscrowFixture(t)
	billing := &cangyuanEscrowLedgerStub{}
	_, _, err := svc.HoldApprovedCangyuanMediaBalance(context.Background(), owner, id, billing)
	require.NoError(t, err)
	_, err = svc.ClaimSubmission(context.Background(), owner, id)
	require.NoError(t, err)
	_, err = svc.AttachSubmitted(context.Background(), owner, id,
		&service.CangyuanMediaTask{ID: "remote_reconcile_task", Status: "queued"})
	require.NoError(t, err)
	return svc, owner, id, billing
}

func TestCangyuanReconcilerPollsOnlyOncePerInvocationAndCapturesTerminalResult(t *testing.T) {
	svc, owner, id, billing := setupCangyuanReconciliation(t)
	client, calls := cangyuanReconciliationClient(t,
		`{"id":"remote_reconcile_task","status":"in_progress"}`,
		`{"id":"remote_reconcile_task","status":"completed","data":[{"url":"https://cdn.example.com/result.png"}]}`,
	)
	ctx := context.Background()
	pending, err := svc.ReconcileHeldCangyuanMediaJob(ctx, owner, id, client, billing)
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobProcessing, pending.Status)
	require.Equal(t, "held", pending.BillingState)
	require.Zero(t, billing.captureCalls)

	done, err := svc.ReconcileHeldCangyuanMediaJob(ctx, owner, id, client, billing)
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobCompleted, done.Status)
	require.Equal(t, "settled", done.BillingState)
	require.Equal(t, int32(2), calls.Load())
	require.Equal(t, 1, billing.captureCalls)
	require.Equal(t, 1, billing.applied)
	require.Zero(t, billing.lastCommand.BalanceCost)

	// A terminal, settled job must not poll again or capture again.
	again, err := svc.ReconcileHeldCangyuanMediaJob(ctx, owner, id, client, billing)
	require.NoError(t, err)
	require.Equal(t, done.Version, again.Version)
	require.Equal(t, int32(2), calls.Load())
	require.Equal(t, 1, billing.captureCalls)
}

func TestCangyuanReconcilerRetriesTransientGETButNeverResubmitsOrRefunds(t *testing.T) {
	svc, owner, id, billing := setupCangyuanReconciliation(t)
	client, calls := cangyuanReconciliationClient(t,
		"transport_failure",
		`{"id":"remote_reconcile_task","status":"failed"}`,
	)
	ctx := context.Background()
	_, err := svc.ReconcileHeldCangyuanMediaJob(ctx, owner, id, client, billing)
	require.Error(t, err)
	require.Zero(t, billing.releaseCalls)
	status, err := svc.Get(ctx, owner, id)
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobQueued, status.Status)
	require.Equal(t, "held", status.BillingState)

	failed, err := svc.ReconcileHeldCangyuanMediaJob(ctx, owner, id, client, billing)
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobFailed, failed.Status)
	require.Equal(t, "released", failed.BillingState)
	require.Equal(t, int32(2), calls.Load())
	require.Equal(t, 1, billing.releaseCalls)
	require.Zero(t, billing.captureCalls)
	require.Zero(t, billing.applied)

	again, err := svc.ReconcileHeldCangyuanMediaJob(ctx, owner, id, client, billing)
	require.NoError(t, err)
	require.Equal(t, "released", again.BillingState)
	require.Equal(t, int32(2), calls.Load())
	require.Equal(t, 1, billing.releaseCalls)
}

func TestCangyuanReconcilerDeclinesUnknownAndUnfundedWithoutAnyNetworkCall(t *testing.T) {
	ctx := context.Background()
	svc, owner, id, billing := setupCangyuanReconciliation(t)
	client, calls := cangyuanReconciliationClient(t)
	_, err := svc.MarkSubmissionUnknown(ctx, owner, id)
	// A known queued task cannot be relabeled ambiguous.
	require.ErrorIs(t, err, service.ErrCangyuanMediaJobState)
	require.Zero(t, calls.Load())

	svc2, _ := newCangyuanMediaJobTestService(t)
	unknownOwner := service.CangyuanMediaJobOwner{UserID: 101, APIKeyID: 202}
	payload := []byte(`{"model":"image-model","prompt":"a boat"}`)
	unfunded, err := svc2.ReserveIdempotent(ctx, unknownOwner, 81, service.CangyuanImageGeneration,
		"image-model", payload, "unfunded-unique", cangyuanIdempotencyQuote())
	require.NoError(t, err)
	_, err = svc2.ReconcileHeldCangyuanMediaJob(ctx, unknownOwner, unfunded.ID, client, billing)
	require.ErrorIs(t, err, service.ErrCangyuanMediaNotBillable)
	require.Zero(t, calls.Load())

	_, _, err = svc2.HoldApprovedCangyuanMediaBalance(ctx, unknownOwner, unfunded.ID, billing)
	require.NoError(t, err)
	_, err = svc2.ClaimSubmission(ctx, unknownOwner, unfunded.ID)
	require.NoError(t, err)
	_, err = svc2.MarkSubmissionUnknown(ctx, unknownOwner, unfunded.ID)
	require.NoError(t, err)
	_, err = svc2.ReconcileHeldCangyuanMediaJob(ctx, unknownOwner, unfunded.ID, client, billing)
	require.ErrorIs(t, err, service.ErrCangyuanMediaJobState)
	require.Zero(t, billing.releaseCalls)
	require.Zero(t, calls.Load(), "ambiguous POST must never be retried or polled without ID")
}
