package service

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// CangyuanMediaEscrowWriter reuses the existing SQL balance-hold transaction.
// Media job IDs begin with cyjob_, so they cannot collide with batch-image IDs.
// These are internal operations; public submission and subscription billing
// remain disabled until authorization and recovery orchestration are complete.
type CangyuanMediaEscrowWriter interface {
	ReserveBatchImageBalance(context.Context, *BatchImageBalanceHoldCommand) (*BatchImageBalanceHoldResult, error)
	CaptureBatchImageBalance(context.Context, *BatchImageBalanceHoldCommand) (*BatchImageBalanceHoldResult, error)
	ReleaseBatchImageBalance(context.Context, *BatchImageBalanceHoldCommand) (*BatchImageBalanceHoldResult, error)
	Apply(context.Context, *UsageBillingCommand) (*UsageBillingApplyResult, error)
}

func cangyuanEscrowCommand(job *CangyuanMediaJob, requestID string, actualAmount float64) (*BatchImageBalanceHoldCommand, float64, float64, error) {
	if job == nil || job.PriceSnapshot == nil || job.PayloadSHA256 == "" ||
		job.UserID <= 0 || job.APIKeyID <= 0 || job.AccountID <= 0 {
		return nil, 0, 0, ErrCangyuanMediaNotBillable
	}
	quote := job.PriceSnapshot
	if err := validateCangyuanMediaPriceSnapshot(job.Kind, quote); err != nil {
		return nil, 0, 0, err
	}
	sale, err := quote.totalSaleUSD()
	if err != nil || sale <= 0 {
		return nil, 0, 0, ErrCangyuanMediaNotBillable
	}
	cost, err := quote.totalProviderCostUSD()
	if err != nil {
		return nil, 0, 0, err
	}
	// A frozen quote and pinned original payload are the only source of price.
	// No current setting, user-supplied price, or upstream usage can change it.
	hash := HashUsageRequestPayload([]byte(job.ID + "|" + job.PayloadSHA256 +
		"|" + quote.PricingVersion + "|" + quote.Tier + "|" +
		quote.UnitSalePriceUSD + "|" + quote.UnitProviderCostUSD))
	return &BatchImageBalanceHoldCommand{
		RequestID:          requestID,
		APIKeyID:           job.APIKeyID,
		UserID:             job.UserID,
		BatchID:            job.ID,
		HoldAmount:         sale,
		ActualAmount:       actualAmount,
		RequestPayloadHash: hash,
	}, sale, cost, nil
}

// HoldApprovedCangyuanMediaBalance atomically freezes the full approved price
// in SQL before the first paid upstream POST. If Redis acknowledgement fails,
// retries reuse the same SQL hold id; they never freeze the same money twice.
func (s *CangyuanMediaJobService) HoldApprovedCangyuanMediaBalance(
	ctx context.Context, owner CangyuanMediaJobOwner, id string, billing CangyuanMediaEscrowWriter,
) (*CangyuanMediaJob, *BatchImageBalanceHoldResult, error) {
	if billing == nil {
		return nil, nil, errors.New("cangyuan media escrow is unavailable")
	}
	job, err := s.Get(ctx, owner, id)
	if err != nil {
		return nil, nil, err
	}
	if job.BillingState == "held" && job.Status == CangyuanMediaJobReserved {
		return job, &BatchImageBalanceHoldResult{Applied: false}, nil
	}
	if job.BillingState != "pending" || job.Status != CangyuanMediaJobReserved {
		return job, nil, ErrCangyuanMediaNotBillable
	}
	cmd, _, _, err := cangyuanEscrowCommand(job, BatchImageHoldRequestID(id), 0)
	if err != nil {
		return job, nil, err
	}
	sqlCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()
	result, err := billing.ReserveBatchImageBalance(sqlCtx, cmd)
	if err != nil {
		return job, nil, fmt.Errorf("reserve cangyuan media balance: %w", err)
	}
	if result == nil {
		return job, nil, errors.New("cangyuan media balance hold returned no result")
	}
	next := *job
	next.BillingState = "held"
	next.Version++
	next.UpdatedAt = time.Now().UTC().Unix()
	storeCtx, storeCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer storeCancel()
	if err := s.Store.CompareAndSwap(storeCtx, &next, job.Version); err != nil {
		if errors.Is(err, ErrCangyuanMediaJobConflict) {
			current, readErr := s.Get(storeCtx, owner, id)
			if readErr == nil && current.BillingState == "held" {
				return current, result, nil
			}
		}
		return job, result, fmt.Errorf("cangyuan media funds held; Redis state needs reconciliation: %w", err)
	}
	return &next, result, nil
}

// SettleHeldCangyuanMedia uses capture (NOT a second balance deduction) then
// an independently idempotent quota transaction. A failed second step can be
// retried safely: the SQL capture request has its own stable dedup key.
func (s *CangyuanMediaJobService) SettleHeldCangyuanMedia(
	ctx context.Context, owner CangyuanMediaJobOwner, id string, billing CangyuanMediaEscrowWriter,
) (*CangyuanMediaJob, *UsageBillingApplyResult, error) {
	if billing == nil {
		return nil, nil, errors.New("cangyuan media escrow is unavailable")
	}
	job, err := s.Get(ctx, owner, id)
	if err != nil {
		return nil, nil, err
	}
	if job.BillingState == "settled" && job.Status == CangyuanMediaJobCompleted {
		return job, &UsageBillingApplyResult{Applied: false}, nil
	}
	if job.BillingState != "held" || job.Status != CangyuanMediaJobCompleted {
		return job, nil, ErrCangyuanMediaNotBillable
	}
	cmd, sale, cost, err := cangyuanEscrowCommand(job, BatchImageCaptureRequestID(id), 0)
	if err != nil {
		return job, nil, err
	}
	if len(job.ResultURLs) != job.PriceSnapshot.Units {
		return job, nil, ErrCangyuanMediaNotBillable
	}
	for _, resultURL := range job.ResultURLs {
		if !isPublicCangyuanMediaReference(resultURL) {
			return job, nil, ErrCangyuanMediaNotBillable
		}
	}
	cmd.ActualAmount = sale
	sqlCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()
	if _, err := billing.CaptureBatchImageBalance(sqlCtx, cmd); err != nil {
		return job, nil, fmt.Errorf("capture cangyuan media reserved funds: %w", err)
	}
	kind := "image"
	imageCount := job.PriceSnapshot.Units
	if job.Kind == CangyuanVideoGeneration {
		kind = "video"
		imageCount = 0
	}
	usage := &UsageBillingCommand{
		RequestID:           "cangyuan-media:" + job.ID,
		APIKeyID:            job.APIKeyID,
		UserID:              job.UserID,
		AccountID:           job.AccountID,
		AccountType:         AccountTypeAPIKey,
		Model:               job.Model,
		MediaType:           kind,
		ImageCount:          imageCount,
		BalanceCost:         0, // SQL capture already consumed the frozen balance.
		APIKeyQuotaCost:     sale,
		APIKeyRateLimitCost: sale,
		AccountQuotaCost:    cost,
		RequestPayloadHash:  cmd.RequestPayloadHash,
	}
	usage.Normalize()
	result, err := billing.Apply(sqlCtx, usage)
	if err != nil {
		return job, nil, fmt.Errorf("record cangyuan media settled quotas: %w", err)
	}
	if result == nil {
		return job, nil, errors.New("cangyuan media quota settlement returned no result")
	}
	next := *job
	next.BillingState = "settled"
	next.Version++
	next.UpdatedAt = time.Now().UTC().Unix()
	storeCtx, storeCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer storeCancel()
	if err := s.Store.CompareAndSwap(storeCtx, &next, job.Version); err != nil {
		if errors.Is(err, ErrCangyuanMediaJobConflict) {
			current, readErr := s.Get(storeCtx, owner, id)
			if readErr == nil && current.BillingState == "settled" && current.Status == CangyuanMediaJobCompleted {
				return current, result, nil
			}
		}
		return job, result, fmt.Errorf("cangyuan media money captured and quota recorded; Redis needs reconciliation: %w", err)
	}
	return &next, result, nil
}

// ReleaseFailedCangyuanMediaBalance refunds a terminal failed media job. An
// ambiguous POST (submit_unknown) is deliberately excluded from automatic
// refunds until an operator reconciles the upstream acceptance status.
func (s *CangyuanMediaJobService) ReleaseFailedCangyuanMediaBalance(
	ctx context.Context, owner CangyuanMediaJobOwner, id string, billing CangyuanMediaEscrowWriter,
) (*CangyuanMediaJob, *BatchImageBalanceHoldResult, error) {
	if billing == nil {
		return nil, nil, errors.New("cangyuan media escrow is unavailable")
	}
	job, err := s.Get(ctx, owner, id)
	if err != nil {
		return nil, nil, err
	}
	if job.BillingState == "released" && job.Status == CangyuanMediaJobFailed {
		return job, &BatchImageBalanceHoldResult{Applied: false}, nil
	}
	if job.BillingState != "held" || job.Status != CangyuanMediaJobFailed {
		return job, nil, ErrCangyuanMediaNotBillable
	}
	cmd, _, _, err := cangyuanEscrowCommand(job, BatchImageReleaseRequestID(id), 0)
	if err != nil {
		return job, nil, err
	}
	sqlCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()
	result, err := billing.ReleaseBatchImageBalance(sqlCtx, cmd)
	if err != nil {
		return job, nil, fmt.Errorf("release failed cangyuan media balance: %w", err)
	}
	if result == nil {
		return job, nil, errors.New("cangyuan media balance release returned no result")
	}
	next := *job
	next.BillingState = "released"
	next.Version++
	next.UpdatedAt = time.Now().UTC().Unix()
	storeCtx, storeCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer storeCancel()
	if err := s.Store.CompareAndSwap(storeCtx, &next, job.Version); err != nil {
		if errors.Is(err, ErrCangyuanMediaJobConflict) {
			current, readErr := s.Get(storeCtx, owner, id)
			if readErr == nil && current.BillingState == "released" && current.Status == CangyuanMediaJobFailed {
				return current, result, nil
			}
		}
		return job, result, fmt.Errorf("cangyuan media funds released; Redis state needs reconciliation: %w", err)
	}
	return &next, result, nil
}
