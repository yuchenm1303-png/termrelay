package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// CangyuanMediaPriceSnapshot is an immutable, admin-verified local sell price.
// It is NOT the provider's public pricing expression. A gateway must resolve
// and approve every tier/duration/size before reserving a paid job.
// Prices are USD decimal strings to avoid floating-point quoting drift.
type CangyuanMediaPriceSnapshot struct {
	PricingVersion      string `json:"pricing_version"`
	Currency            string `json:"currency"`
	Tier                string `json:"tier"`
	Units               int    `json:"units"`
	UnitSalePriceUSD    string `json:"unit_sale_price_usd"`
	UnitProviderCostUSD string `json:"unit_provider_cost_usd"`
	VerifiedByAdmin     bool   `json:"verified_by_admin"`
}

func validateCangyuanMediaPriceSnapshot(kind CangyuanMediaKind, q *CangyuanMediaPriceSnapshot) error {
	if q == nil || !q.VerifiedByAdmin || q.Currency != "USD" ||
		strings.TrimSpace(q.PricingVersion) == "" || len(q.PricingVersion) > 128 ||
		strings.TrimSpace(q.Tier) == "" || len(q.Tier) > 128 ||
		q.Units < 1 || q.Units > 16 {
		return errors.New("cangyuan media requires an approved USD pricing snapshot")
	}
	if kind == CangyuanVideoGeneration && q.Units != 1 {
		return errors.New("cangyuan video pricing requires one completed video per job")
	}
	sale, err := parseCangyuanMediaMoney(q.UnitSalePriceUSD)
	if err != nil || !sale.GreaterThan(decimal.Zero) {
		return errors.New("cangyuan media unit sale price must be positive and finite")
	}
	// A provider cost must be specified for margin accounting. A zero provider
	// cost is allowed for explicitly verified promotions, not for missing data.
	provider, err := parseCangyuanMediaMoney(q.UnitProviderCostUSD)
	if err != nil || provider.IsNegative() {
		return errors.New("cangyuan media provider cost must be explicitly quoted")
	}
	return nil
}

// validateCangyuanMediaRequestQuantity prevents a billable upstream job
// from generating more images than the immutable approved price covers.
// The upstream defaults to one image when "n" is absent. Video jobs always
// represent exactly one output, never an image count.
func validateCangyuanMediaRequestQuantity(kind CangyuanMediaKind, request map[string]json.RawMessage, quote *CangyuanMediaPriceSnapshot) error {
	if quote == nil {
		return errors.New("paid cangyuan media requires a frozen price")
	}
	switch kind {
	case CangyuanImageGeneration, CangyuanImageEdit:
		n := 1
		if raw, exists := request["n"]; exists {
			if err := json.Unmarshal(raw, &n); err != nil || n < 1 || n > 16 {
				return errors.New("cangyuan image count must be an integer from 1 to 16")
			}
		}
		if n != quote.Units {
			return errors.New("cangyuan requested image count does not match approved price")
		}
	case CangyuanVideoGeneration:
		if _, exists := request["n"]; exists {
			return errors.New("cangyuan video requests cannot specify image count")
		}
		if quote.Units != 1 {
			return errors.New("cangyuan video jobs require exactly one priced output")
		}
	default:
		return errors.New("unsupported cangyuan media capability")
	}
	return nil
}

func parseCangyuanMediaMoney(raw string) (decimal.Decimal, error) {
	if strings.TrimSpace(raw) != raw || len(raw) == 0 || len(raw) > 24 {
		return decimal.Zero, errors.New("invalid cangyuan media money representation")
	}
	// Keep prices within USD 100,000 per individual output; rejecting very
	// small sub-cent precision prevents different decimal inputs rounding to
	// the same float-based ledger amount.
	var whole, fractional string
	if i := strings.IndexByte(raw, '.'); i >= 0 {
		whole, fractional = raw[:i], raw[i+1:]
	} else {
		whole = raw
	}
	if len(whole) == 0 || len(fractional) > 8 {
		return decimal.Zero, errors.New("cangyuan media prices support at most eight decimal places")
	}
	for _, ch := range whole + fractional {
		if ch < '0' || ch > '9' {
			return decimal.Zero, errors.New("cangyuan media prices must use unsigned decimal numbers")
		}
	}
	value, err := decimal.NewFromString(raw)
	if err != nil || value.GreaterThan(decimal.NewFromInt(100000)) {
		return decimal.Zero, fmt.Errorf("cangyuan media price is invalid or too large")
	}
	return value, nil
}

func (q *CangyuanMediaPriceSnapshot) totalSaleUSD() (float64, error) {
	if q == nil {
		return 0, errors.New("missing cangyuan media price snapshot")
	}
	unit, err := parseCangyuanMediaMoney(q.UnitSalePriceUSD)
	if err != nil {
		return 0, err
	}
	return unit.Mul(decimal.NewFromInt(int64(q.Units))).InexactFloat64(), nil
}

func (q *CangyuanMediaPriceSnapshot) totalProviderCostUSD() (float64, error) {
	if q == nil {
		return 0, errors.New("missing cangyuan media price snapshot")
	}
	unit, err := parseCangyuanMediaMoney(q.UnitProviderCostUSD)
	if err != nil {
		return 0, err
	}
	return unit.Mul(decimal.NewFromInt(int64(q.Units))).InexactFloat64(), nil
}
