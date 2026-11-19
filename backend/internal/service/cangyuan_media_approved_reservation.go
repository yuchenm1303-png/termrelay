package service

import (
	"context"
	"errors"
	"strings"
)

// ReserveWithApprovedPricing is the only intended entrypoint for future paid
// Cangyuan media routes. It derives a trusted quote from persistent admin
// settings; never accepts VerifiedByAdmin or USD amounts from a user request.
//
// This helper does NOT authorize an HTTP request. The future public gateway
// must first check group entitlements, account scheduling state, concurrency,
// subscription/quota eligibility and a pre-generation balance reservation.
func (s *CangyuanMediaJobService) ReserveWithApprovedPricing(
	ctx context.Context,
	owner CangyuanMediaJobOwner,
	account *Account,
	kind CangyuanMediaKind,
	model string,
	payload []byte,
	idempotencyKey string,
	prices *CangyuanMediaPriceRegistry,
) (*CangyuanMediaJob, error) {
	if prices == nil {
		return nil, errors.New("cangyuan media pricing registry is unavailable")
	}
	if account == nil || account.ID <= 0 || account.Status != StatusActive ||
		!isCangyuanMediaOnlyAccount(account) || !account.AllowsCangyuanMediaModel(kind, model) ||
		strings.TrimSpace(account.GetCredential("api_key")) == "" {
		return nil, errors.New("cangyuan media account is not eligible for this model")
	}
	quote, err := prices.Quote(ctx, kind, model, payload)
	if err != nil {
		return nil, err
	}
	return s.ReserveIdempotent(ctx, owner, account.ID, kind, model, payload, idempotencyKey, *quote)
}
