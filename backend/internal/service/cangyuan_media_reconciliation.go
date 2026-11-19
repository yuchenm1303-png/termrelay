package service

import (
	"context"
	"errors"
)

// ReconcileHeldCangyuanMediaJob advances an existing, already-funded media job
// by one provider GET and settles (or releases) its frozen SQL balance once it
// reaches a verified terminal state. This method is intentionally independent
// of HTTP routes and background schedulers. The caller must supply the exact
// original media-only upstream account and its credentials after re-checking
// ownership, entitlement, and account eligibility.
//
// It NEVER re-POSTs an upstream generation, even after process crashes.
// An ambiguous submit remains frozen and requires operator reconciliation.
func (s *CangyuanMediaJobService) ReconcileHeldCangyuanMediaJob(
	ctx context.Context,
	owner CangyuanMediaJobOwner,
	id string,
	client *CangyuanMediaClient,
	billing CangyuanMediaEscrowWriter,
) (*CangyuanMediaJob, error) {
	if billing == nil {
		return nil, errors.New("cangyuan media reconciliation requires a billing repository")
	}
	job, err := s.Get(ctx, owner, id)
	if err != nil {
		return nil, err
	}
	if (job.BillingState == "settled" && job.Status == CangyuanMediaJobCompleted) ||
		(job.BillingState == "released" && job.Status == CangyuanMediaJobFailed) {
		return job, nil
	}
	if job.BillingState != "held" || job.PriceSnapshot == nil || job.PayloadSHA256 == "" {
		return job, ErrCangyuanMediaNotBillable
	}

	switch job.Status {
	case CangyuanMediaJobReserved:
		// The initial POST has not happened. Only the authenticated submission
		// path may claim and POST this reservation; the reconciler never does.
		return job, nil
	case CangyuanMediaJobSubmitting, CangyuanMediaJobSubmitUnknown:
		// An in-flight or ambiguous POST must not be sent again or treated as
		// a failed generation merely because the HTTP request disappeared.
		return job, ErrCangyuanMediaJobState
	case CangyuanMediaJobQueued, CangyuanMediaJobProcessing:
		if client == nil {
			return job, errors.New("cangyuan media polling client is unavailable")
		}
		job, err = s.PollJob(ctx, owner, id, client)
		if errors.Is(err, ErrCangyuanMediaJobConflict) {
			// Parallel pollers may both observe the same transition. Read the
			// latest committed state, then only settle an actual terminal job.
			job, err = s.Get(ctx, owner, id)
		}
		if err != nil {
			return job, err
		}
	case CangyuanMediaJobCompleted, CangyuanMediaJobFailed:
		// SQL may already have committed while its Redis marker failed.
		// Retrying here is safe because capture/release/usage are SQL-idempotent.
	default:
		return job, ErrCangyuanMediaJobState
	}

	switch job.Status {
	case CangyuanMediaJobCompleted:
		updated, _, settleErr := s.SettleHeldCangyuanMedia(ctx, owner, id, billing)
		return updated, settleErr
	case CangyuanMediaJobFailed:
		updated, _, refundErr := s.ReleaseFailedCangyuanMediaBalance(ctx, owner, id, billing)
		return updated, refundErr
	default:
		return job, nil
	}
}
