package service

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// CangyuanMediaBillingWriter is deliberately a narrow port of the existing
// SQL-backed UsageBillingRepository.Apply transaction. That repository
// deduplicates by (request_id, api_key_id) before mutating balances or quotas.
type CangyuanMediaBillingWriter interface {
	Apply(ctx context.Context, cmd *UsageBillingCommand) (*UsageBillingApplyResult, error)
}

var ErrCangyuanMediaNotBillable = errors.New("cangyuan media job is not ready for billing")

// SettleCompleted is an internal-only settlement primitive. No route invokes
// it until model permission, verified pricing, balance reservation, complete
// usage logging and account eligibility are wired end-to-end.
//
// Charge is derived solely from the immutable admin-approved snapshot saved
// BEFORE the upstream request. GET responses cannot increase the price. The
// SQL repository is the source of truth for idempotency; Redis only mirrors
// the final settled state and must not be relied on for double-charge safety.
func (s *CangyuanMediaJobService) SettleCompleted(
	ctx context.Context,
	owner CangyuanMediaJobOwner,
	id string,
	billing CangyuanMediaBillingWriter,
) (*CangyuanMediaJob, *UsageBillingApplyResult, error) {
	if billing == nil {
		return nil, nil, errors.New("cangyuan media billing repository unavailable")
	}
	job, err := s.Get(ctx, owner, id)
	if err != nil {
		return nil, nil, err
	}
	if job.BillingState == "settled" && job.Status == CangyuanMediaJobCompleted {
		return job, &UsageBillingApplyResult{Applied: false}, nil
	}
	if job.BillingState != "pending" || job.Status != CangyuanMediaJobCompleted || job.PriceSnapshot == nil {
		return nil, nil, ErrCangyuanMediaNotBillable
	}
	quote := job.PriceSnapshot
	if err := validateCangyuanMediaPriceSnapshot(job.Kind, quote); err != nil {
		return nil, nil, fmt.Errorf("invalid frozen cangyuan media price: %w", err)
	}
	if len(job.ResultURLs) != quote.Units {
		return nil, nil, ErrCangyuanMediaNotBillable
	}
	for _, resultURL := range job.ResultURLs {
		if !isPublicCangyuanMediaReference(resultURL) {
			return nil, nil, ErrCangyuanMediaNotBillable
		}
	}
	sale, err := quote.totalSaleUSD()
	if err != nil {
		return nil, nil, err
	}
	cost, err := quote.totalProviderCostUSD()
	if err != nil {
		return nil, nil, err
	}
	mediaType := "image"
	imageCount := quote.Units
	if job.Kind == CangyuanVideoGeneration {
		mediaType = "video"
		imageCount = 0
	}
	command := &UsageBillingCommand{
		RequestID:           "cangyuan-media:" + job.ID,
		APIKeyID:            job.APIKeyID,
		UserID:              job.UserID,
		AccountID:           job.AccountID,
		AccountType:         AccountTypeAPIKey,
		Model:               job.Model,
		MediaType:           mediaType,
		ImageCount:          imageCount,
		BalanceCost:         sale,
		APIKeyQuotaCost:     sale,
		APIKeyRateLimitCost: sale,
		AccountQuotaCost:    cost,
		RequestPayloadHash:  HashUsageRequestPayload([]byte(quote.PricingVersion + "|" + quote.Tier + "|" + quote.UnitSalePriceUSD + "|" + quote.UnitProviderCostUSD)),
	}
	command.Normalize()
	// Client disconnection must not interrupt the atomic SQL transaction
	// after a media generation has already succeeded.
	billingCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()
	result, err := billing.Apply(billingCtx, command)
	if err != nil {
		return job, nil, fmt.Errorf("settle cangyuan media job: %w", err)
	}
	if result == nil {
		return job, nil, errors.New("cangyuan media billing returned no result")
	}

	next := *job
	next.BillingState = "settled"
	next.Version++
	next.UpdatedAt = time.Now().UTC().Unix()
	// A Redis failure here must NOT cause a second debit: retrying always
	// uses the same SQL dedup key and immutable frozen snapshot.
	storeCtx, storeCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer storeCancel()
	if err := s.Store.CompareAndSwap(storeCtx, &next, job.Version); err != nil {
		if errors.Is(err, ErrCangyuanMediaJobConflict) {
			current, readErr := s.Get(storeCtx, owner, id)
			if readErr == nil && current.BillingState == "settled" && current.Status == CangyuanMediaJobCompleted {
				return current, result, nil
			}
		}
		return job, result, fmt.Errorf("media billing committed but Redis settlement marker must be reconciled: %w", err)
	}
	return &next, result, nil
}
