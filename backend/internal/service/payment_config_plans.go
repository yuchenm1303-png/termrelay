package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionplan"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const planCardMetadataPrefix = "__smirel_plan_card_v1__:"

// PlanCardConfig is the stable API representation of fields that are shown on
// the subscription card but are not part of the legacy plan schema yet.
type PlanCardConfig struct {
	Tier             string `json:"card_tier,omitempty"`
	Badge            string `json:"card_badge,omitempty"`
	Featured         bool   `json:"card_featured,omitempty"`
	Footnote         string `json:"card_footnote,omitempty"`
	SeatLimit        int    `json:"seat_limit,omitempty"`
	ConcurrencyLimit int    `json:"concurrency_limit,omitempty"`
	PurchasePolicy   string `json:"purchase_policy,omitempty"`
}

func EncodePlanFeatures(features string, cfg PlanCardConfig) string {
	raw, err := json.Marshal(cfg)
	if err != nil {
		return features
	}
	clean := DecodePlanFeatures(features)
	return strings.TrimRight(clean.Features, "\n") + "\n" + planCardMetadataPrefix + string(raw)
}

type DecodedPlanFeatures struct {
	Features string
	Card     PlanCardConfig
}

func DecodePlanFeatures(raw string) DecodedPlanFeatures {
	lines := strings.Split(raw, "\n")
	visible := make([]string, 0, len(lines))
	var cfg PlanCardConfig
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, planCardMetadataPrefix) {
			_ = json.Unmarshal([]byte(strings.TrimPrefix(trimmed, planCardMetadataPrefix)), &cfg)
			continue
		}
		if trimmed != "" {
			visible = append(visible, line)
		}
	}
	return DecodedPlanFeatures{Features: strings.Join(visible, "\n"), Card: cfg}
}

// normalizePlanCurrency validates and normalizes the display-only currency label.
// Empty means "no label" and is kept as-is so existing plans stay unchanged.
func normalizePlanCurrency(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", nil
	}
	currency, err := payment.NormalizePaymentCurrency(raw)
	if err != nil {
		return "", infraerrors.BadRequest("PLAN_CURRENCY_INVALID", "currency must be a 3-letter ISO currency code")
	}
	return currency, nil
}

// validatePlanRequired checks that all required fields for a plan are provided.
func validatePlanRequired(name string, groupID int64, price float64, validityDays int, validityUnit string, originalPrice *float64) error {
	if strings.TrimSpace(name) == "" {
		return infraerrors.BadRequest("PLAN_NAME_REQUIRED", "plan name is required")
	}
	if groupID <= 0 {
		return infraerrors.BadRequest("PLAN_GROUP_REQUIRED", "group is required")
	}
	if price <= 0 {
		return infraerrors.BadRequest("PLAN_PRICE_INVALID", "price must be > 0")
	}
	if validityDays <= 0 {
		return infraerrors.BadRequest("PLAN_VALIDITY_REQUIRED", "validity days must be > 0")
	}
	if strings.TrimSpace(validityUnit) == "" {
		return infraerrors.BadRequest("PLAN_VALIDITY_UNIT_REQUIRED", "validity unit is required")
	}
	if originalPrice != nil && *originalPrice < 0 {
		return infraerrors.BadRequest("PLAN_ORIGINAL_PRICE_INVALID", "original price must be >= 0")
	}
	return nil
}

func validatePlanCardConfig(seatLimit, concurrencyLimit int, purchasePolicy string) error {
	if seatLimit < 1 {
		return infraerrors.BadRequest("PLAN_SEAT_LIMIT_INVALID", "seat_limit must be >= 1")
	}
	if concurrencyLimit < 1 {
		return infraerrors.BadRequest("PLAN_CONCURRENCY_LIMIT_INVALID", "concurrency_limit must be >= 1")
	}
	if purchasePolicy != "" && purchasePolicy != "public" && purchasePolicy != "approval" {
		return infraerrors.BadRequest("PLAN_PURCHASE_POLICY_INVALID", "purchase_policy must be public or approval")
	}
	return nil
}

// validatePlanPatch validates only the non-nil fields in a patch update.
func validatePlanPatch(req UpdatePlanRequest) error {
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return infraerrors.BadRequest("PLAN_NAME_REQUIRED", "plan name is required")
	}
	if req.GroupID != nil && *req.GroupID <= 0 {
		return infraerrors.BadRequest("PLAN_GROUP_REQUIRED", "group is required")
	}
	if req.Price != nil && *req.Price <= 0 {
		return infraerrors.BadRequest("PLAN_PRICE_INVALID", "price must be > 0")
	}
	if req.ValidityDays != nil && *req.ValidityDays <= 0 {
		return infraerrors.BadRequest("PLAN_VALIDITY_REQUIRED", "validity days must be > 0")
	}
	if req.ValidityUnit != nil && strings.TrimSpace(*req.ValidityUnit) == "" {
		return infraerrors.BadRequest("PLAN_VALIDITY_UNIT_REQUIRED", "validity unit is required")
	}
	if req.OriginalPrice != nil && *req.OriginalPrice < 0 {
		return infraerrors.BadRequest("PLAN_ORIGINAL_PRICE_INVALID", "original price must be >= 0")
	}
	if req.SeatLimit != nil && *req.SeatLimit < 1 {
		return infraerrors.BadRequest("PLAN_SEAT_LIMIT_INVALID", "seat_limit must be >= 1")
	}
	if req.ConcurrencyLimit != nil && *req.ConcurrencyLimit < 1 {
		return infraerrors.BadRequest("PLAN_CONCURRENCY_LIMIT_INVALID", "concurrency_limit must be >= 1")
	}
	if req.PurchasePolicy != nil && *req.PurchasePolicy != "public" && *req.PurchasePolicy != "approval" {
		return infraerrors.BadRequest("PLAN_PURCHASE_POLICY_INVALID", "purchase_policy must be public or approval")
	}
	return nil
}

// --- Plan CRUD ---

// PlanGroupInfo holds the group details needed for subscription plan display.
type PlanGroupInfo struct {
	Platform           string   `json:"platform"`
	Name               string   `json:"name"`
	RateMultiplier     float64  `json:"rate_multiplier"`
	PeakRateEnabled    bool     `json:"peak_rate_enabled"`
	PeakStart          string   `json:"peak_start"`
	PeakEnd            string   `json:"peak_end"`
	PeakRateMultiplier float64  `json:"peak_rate_multiplier"`
	DailyLimitUSD      *float64 `json:"daily_limit_usd"`
	WeeklyLimitUSD     *float64 `json:"weekly_limit_usd"`
	MonthlyLimitUSD    *float64 `json:"monthly_limit_usd"`
	ModelScopes        []string `json:"supported_model_scopes"`
}

// GetGroupInfoMap returns a map of group_id → PlanGroupInfo for the given plans.
func (s *PaymentConfigService) GetGroupInfoMap(ctx context.Context, plans []*dbent.SubscriptionPlan) map[int64]PlanGroupInfo {
	ids := make([]int64, 0, len(plans))
	seen := make(map[int64]bool)
	for _, p := range plans {
		if p == nil || !p.GroupBound {
			continue
		}
		if !seen[p.GroupID] {
			seen[p.GroupID] = true
			ids = append(ids, p.GroupID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	groups, err := s.entClient.Group.Query().Where(group.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil
	}
	m := make(map[int64]PlanGroupInfo, len(groups))
	for _, g := range groups {
		m[int64(g.ID)] = PlanGroupInfo{
			Platform:           g.Platform,
			Name:               g.Name,
			RateMultiplier:     g.RateMultiplier,
			PeakRateEnabled:    g.PeakRateEnabled,
			PeakStart:          g.PeakStart,
			PeakEnd:            g.PeakEnd,
			PeakRateMultiplier: g.PeakRateMultiplier,
			DailyLimitUSD:      g.DailyLimitUsd,
			WeeklyLimitUSD:     g.WeeklyLimitUsd,
			MonthlyLimitUSD:    g.MonthlyLimitUsd,
			ModelScopes:        g.SupportedModelScopes,
		}
	}
	return m
}

func (s *PaymentConfigService) ListPlans(ctx context.Context) ([]*dbent.SubscriptionPlan, error) {
	return s.entClient.SubscriptionPlan.Query().Order(subscriptionplan.BySortOrder()).All(ctx)
}

func (s *PaymentConfigService) ListPlansForSale(ctx context.Context) ([]*dbent.SubscriptionPlan, error) {
	return s.entClient.SubscriptionPlan.Query().
		Where(subscriptionplan.ForSaleEQ(true), subscriptionplan.GroupBoundEQ(true)).
		Order(subscriptionplan.BySortOrder()).
		All(ctx)
}

func (s *PaymentConfigService) CreatePlan(ctx context.Context, req CreatePlanRequest) (*dbent.SubscriptionPlan, error) {
	if err := validatePlanRequired(req.Name, req.GroupID, req.Price, req.ValidityDays, req.ValidityUnit, req.OriginalPrice); err != nil {
		return nil, err
	}
	currency, err := normalizePlanCurrency(req.Currency)
	if err != nil {
		return nil, err
	}
	if err := validatePlanCardConfig(planPositiveInt(req.SeatLimit, 1), planPositiveInt(req.ConcurrencyLimit, 5), req.PurchasePolicy); err != nil {
		return nil, err
	}
	if req.ForSale {
		if err := s.validatePlanSaleGroup(ctx, req.GroupID); err != nil {
			return nil, err
		}
	}
	features := EncodePlanFeatures(req.Features, PlanCardConfig{Tier: req.CardTier, Badge: req.CardBadge, Featured: req.CardFeatured, Footnote: req.CardFootnote, SeatLimit: planPositiveInt(req.SeatLimit, 1), ConcurrencyLimit: planPositiveInt(req.ConcurrencyLimit, 5), PurchasePolicy: normalizePurchasePolicy(req.PurchasePolicy)})
	b := s.entClient.SubscriptionPlan.Create().
		SetGroupID(req.GroupID).SetGroupBound(true).SetName(req.Name).SetDescription(req.Description).
		SetPrice(req.Price).SetCurrency(currency).SetValidityDays(req.ValidityDays).SetValidityUnit(req.ValidityUnit).
		SetFeatures(features).SetProductName(req.ProductName).
		SetForSale(req.ForSale).SetSortOrder(req.SortOrder)
	if req.OriginalPrice != nil {
		b.SetOriginalPrice(*req.OriginalPrice)
	}
	return b.Save(ctx)
}

// UpdatePlan updates a subscription plan by ID (patch semantics).
// NOTE: This function exceeds 30 lines due to per-field nil-check patch update boilerplate
// plus a validation guard for non-nil fields.
func (s *PaymentConfigService) UpdatePlan(ctx context.Context, id int64, req UpdatePlanRequest) (*dbent.SubscriptionPlan, error) {
	if err := validatePlanPatch(req); err != nil {
		return nil, err
	}
	u := s.entClient.SubscriptionPlan.UpdateOneID(id)
	existing, err := s.entClient.SubscriptionPlan.Get(ctx, id)
	if err != nil {
		return nil, infraerrors.NotFound("PLAN_NOT_FOUND", "subscription plan not found")
	}
	targetForSale := existing.ForSale
	if req.ForSale != nil {
		targetForSale = *req.ForSale
	}
	if targetForSale {
		targetGroupID := existing.GroupID
		targetBound := existing.GroupBound
		if req.GroupID != nil {
			targetGroupID = *req.GroupID
			targetBound = true
		}
		if !targetBound {
			return nil, infraerrors.Conflict("PLAN_GROUP_UNBOUND", "bind a group before putting this plan on sale")
		}
		if err := s.validatePlanSaleGroup(ctx, targetGroupID); err != nil {
			return nil, err
		}
	}
	if req.GroupID != nil {
		u.SetGroupID(*req.GroupID).SetGroupBound(true)
	}
	if req.Name != nil {
		u.SetName(*req.Name)
	}
	if req.Description != nil {
		u.SetDescription(*req.Description)
	}
	if req.Price != nil {
		u.SetPrice(*req.Price)
	}
	if req.OriginalPrice != nil {
		u.SetOriginalPrice(*req.OriginalPrice)
	}
	if req.Currency != nil {
		currency, err := normalizePlanCurrency(*req.Currency)
		if err != nil {
			return nil, err
		}
		u.SetCurrency(currency)
	}
	if req.ValidityDays != nil {
		u.SetValidityDays(*req.ValidityDays)
	}
	if req.ValidityUnit != nil {
		u.SetValidityUnit(*req.ValidityUnit)
	}
	if req.Features != nil {
		decoded := DecodePlanFeatures(existing.Features)
		decoded.Features = *req.Features
		if req.CardTier != nil {
			decoded.Card.Tier = *req.CardTier
		}
		if req.CardBadge != nil {
			decoded.Card.Badge = *req.CardBadge
		}
		if req.CardFeatured != nil {
			decoded.Card.Featured = *req.CardFeatured
		}
		if req.CardFootnote != nil {
			decoded.Card.Footnote = *req.CardFootnote
		}
		if req.SeatLimit != nil {
			decoded.Card.SeatLimit = *req.SeatLimit
		}
		if req.ConcurrencyLimit != nil {
			decoded.Card.ConcurrencyLimit = *req.ConcurrencyLimit
		}
		if req.PurchasePolicy != nil {
			decoded.Card.PurchasePolicy = *req.PurchasePolicy
		}
		u.SetFeatures(EncodePlanFeatures(decoded.Features, decoded.Card))
	} else if req.CardTier != nil || req.CardBadge != nil || req.CardFeatured != nil || req.CardFootnote != nil || req.SeatLimit != nil || req.ConcurrencyLimit != nil || req.PurchasePolicy != nil {
		decoded := DecodePlanFeatures(existing.Features)
		if req.CardTier != nil {
			decoded.Card.Tier = *req.CardTier
		}
		if req.CardBadge != nil {
			decoded.Card.Badge = *req.CardBadge
		}
		if req.CardFeatured != nil {
			decoded.Card.Featured = *req.CardFeatured
		}
		if req.CardFootnote != nil {
			decoded.Card.Footnote = *req.CardFootnote
		}
		if req.SeatLimit != nil {
			decoded.Card.SeatLimit = *req.SeatLimit
		}
		if req.ConcurrencyLimit != nil {
			decoded.Card.ConcurrencyLimit = *req.ConcurrencyLimit
		}
		if req.PurchasePolicy != nil {
			decoded.Card.PurchasePolicy = *req.PurchasePolicy
		}
		u.SetFeatures(EncodePlanFeatures(decoded.Features, decoded.Card))
	}
	if req.ProductName != nil {
		u.SetProductName(*req.ProductName)
	}
	if req.ForSale != nil {
		u.SetForSale(*req.ForSale)
	}
	if req.SortOrder != nil {
		u.SetSortOrder(*req.SortOrder)
	}
	return u.Save(ctx)
}

func (s *PaymentConfigService) validatePlanSaleGroup(ctx context.Context, groupID int64) error {
	if groupID <= 0 {
		return infraerrors.BadRequest("PLAN_GROUP_REQUIRED", "group is required")
	}
	g, err := s.entClient.Group.Get(ctx, groupID)
	if err != nil {
		return infraerrors.NotFound("PLAN_GROUP_NOT_FOUND", "subscription group not found")
	}
	if g.Status != StatusActive {
		return infraerrors.Conflict("PLAN_GROUP_INACTIVE", "subscription group must be active before the plan can be sold")
	}
	if g.SubscriptionType != SubscriptionTypeSubscription {
		return infraerrors.Conflict("PLAN_GROUP_TYPE_MISMATCH", "plan can only be sold when bound to a subscription group")
	}
	return nil
}

// BindPlanGroup activates the plan-to-group relationship. Binding itself is
// allowed to an inactive/non-subscription group for staging, but such a plan
// is forced off sale until the target group becomes eligible.
func (s *PaymentConfigService) BindPlanGroup(ctx context.Context, id, groupID int64) (*dbent.SubscriptionPlan, error) {
	if groupID <= 0 {
		return nil, infraerrors.BadRequest("PLAN_GROUP_REQUIRED", "group is required")
	}
	plan, err := s.entClient.SubscriptionPlan.Get(ctx, id)
	if err != nil {
		return nil, infraerrors.NotFound("PLAN_NOT_FOUND", "subscription plan not found")
	}
	g, err := s.entClient.Group.Get(ctx, groupID)
	if err != nil {
		return nil, infraerrors.NotFound("PLAN_GROUP_NOT_FOUND", "subscription group not found")
	}
	u := s.entClient.SubscriptionPlan.UpdateOneID(plan.ID).
		SetGroupID(groupID).
		SetGroupBound(true)
	if g.Status != StatusActive || g.SubscriptionType != SubscriptionTypeSubscription {
		u.SetForSale(false)
	}
	return u.Save(ctx)
}

// UnbindPlanGroup detaches the plan from routing without deleting either side.
// Detached plans are always taken off sale; existing user subscriptions keep
// their own group snapshot and are not modified.
func (s *PaymentConfigService) UnbindPlanGroup(ctx context.Context, id int64) (*dbent.SubscriptionPlan, error) {
	plan, err := s.entClient.SubscriptionPlan.Get(ctx, id)
	if err != nil {
		return nil, infraerrors.NotFound("PLAN_NOT_FOUND", "subscription plan not found")
	}
	return s.entClient.SubscriptionPlan.UpdateOneID(plan.ID).
		SetGroupBound(false).
		SetForSale(false).
		Save(ctx)
}

func planPositiveInt(value, fallback int) int {
	if value < 1 {
		return fallback
	}
	return value
}

func normalizePurchasePolicy(value string) string {
	if value == "approval" {
		return value
	}
	return "public"
}

func (s *PaymentConfigService) DeletePlan(ctx context.Context, id int64) error {
	count, err := s.countPendingOrdersByPlan(ctx, id)
	if err != nil {
		return fmt.Errorf("check pending orders: %w", err)
	}
	if count > 0 {
		return infraerrors.Conflict("PENDING_ORDERS",
			fmt.Sprintf("this plan has %d in-progress orders and cannot be deleted — wait for orders to complete first", count))
	}
	return s.entClient.SubscriptionPlan.DeleteOneID(id).Exec(ctx)
}

// GetPlan returns a subscription plan by ID.
func (s *PaymentConfigService) GetPlan(ctx context.Context, id int64) (*dbent.SubscriptionPlan, error) {
	plan, err := s.entClient.SubscriptionPlan.Get(ctx, id)
	if err != nil {
		return nil, infraerrors.NotFound("PLAN_NOT_FOUND", "subscription plan not found")
	}
	return plan, nil
}
