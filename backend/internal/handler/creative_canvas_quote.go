package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type creativeMediaQuoteRequest struct {
	APIKeyID int64           `json:"api_key_id" binding:"required"`
	Kind     string          `json:"kind" binding:"required"`
	Model    string          `json:"model" binding:"required"`
	Payload  json.RawMessage `json:"payload" binding:"required"`
}

// A preflight quote is not permission to charge: the real generation route
// rechecks account authorization, price and quota under the billing lock.
// No provider credentials or provider cost are exposed in the response.
func (h *CreativeCanvasHandler) QuoteMedia(c *gin.Context) {
	userID, ok := creativeAuthUser(c)
	if !ok {
		return
	}
	if h == nil || h.apiKeys == nil || h.prices == nil || h.media == nil {
		creativeFail(c, http.StatusServiceUnavailable, "Media price quote is unavailable")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	var req creativeMediaQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.APIKeyID <= 0 ||
		len(req.Model) == 0 || len(req.Model) > 200 || strings.TrimSpace(req.Model) != req.Model {
		creativeFail(c, http.StatusBadRequest, "Invalid media quote request")
		return
	}
	var kind service.CangyuanMediaKind
	switch req.Kind {
	case "image":
		kind = service.CangyuanImageGeneration
	case "video":
		kind = service.CangyuanVideoGeneration
	default:
		creativeFail(c, http.StatusBadRequest, "Unsupported media type")
		return
	}
	key, err := h.apiKeys.GetByID(c.Request.Context(), req.APIKeyID)
	if err != nil || key == nil || key.UserID != userID {
		creativeFail(c, http.StatusNotFound, "API Key not found")
		return
	}
	if !key.IsActive() || key.IsExpired() || key.IsQuotaExhausted() ||
		key.GroupID == nil || *key.GroupID <= 0 || key.Group == nil ||
		key.Group.ID != *key.GroupID || !key.Group.IsActive() ||
		key.Group.Platform != service.PlatformOpenAI ||
		(kind == service.CangyuanImageGeneration && !service.GroupAllowsImageGeneration(key.Group)) ||
		key.Group.IsSubscriptionType() {
		creativeFail(c, http.StatusForbidden, "API Key or group is not eligible for media billing")
		return
	}
	candidates, err := h.media.Candidates(c.Request.Context(), *key.GroupID, kind, req.Model)
	if err != nil || len(candidates) == 0 {
		creativeFail(c, http.StatusServiceUnavailable, "No schedulable media account for this model and group")
		return
	}
	approved, err := h.prices.Quote(c.Request.Context(), kind, req.Model, req.Payload)
	if errors.Is(err, service.ErrCangyuanMediaPriceNotFound) {
		creativeFail(c, http.StatusUnprocessableEntity, "No enabled administrator-approved price matches these exact options")
		return
	}
	if err != nil || approved == nil || !approved.VerifiedByAdmin || approved.Currency != "USD" {
		creativeFail(c, http.StatusServiceUnavailable, "Media pricing unavailable")
		return
	}
	unit, err := decimal.NewFromString(approved.UnitSalePriceUSD)
	if err != nil || !unit.IsPositive() || approved.Units <= 0 {
		creativeFail(c, http.StatusServiceUnavailable, "Approved media price is invalid")
		return
	}
	total := unit.Mul(decimal.NewFromInt(int64(approved.Units))).StringFixed(8)
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"currency":            "USD",
		"kind":                req.Kind,
		"model":               req.Model,
		"api_key_id":          req.APIKeyID,
		"group_id":            *key.GroupID,
		"units":               approved.Units,
		"unit_sale_price_usd": approved.UnitSalePriceUSD,
		"total_sale_usd":      total,
		"tier":                approved.Tier,
		"pricing_version":     approved.PricingVersion,
		"approved":            true,
		"disclaimer":          "Final execution checks the same quote and rejects stale or changed pricing",
	}})
}

// matchCanvasPriceExpectation guarantees an opt-in canvas POST can never be
// charged at a different price/version than the quote the user confirmed.
// All non-canvas SDK clients continue to use the existing pricing lifecycle.
func matchCanvasPriceExpectation(h *CangyuanMediaHandler, c *gin.Context,
	kind service.CangyuanMediaKind, model string, payload []byte) bool {
	expected := strings.TrimSpace(c.GetHeader("X-Canvas-Quoted-Sale-USD"))
	version := strings.TrimSpace(c.GetHeader("X-Canvas-Quoted-Pricing-Version"))
	if expected == "" && version == "" {
		return true // ordinary API clients are unchanged
	}
	if expected == "" || version == "" || len(expected) > 24 || len(version) > 128 ||
		h == nil || h.public == nil || h.public.Prices == nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{
			"type": "price_changed", "message": "Canvas pricing quote must be refreshed"}})
		return false
	}
	approved, err := h.public.Prices.Quote(c.Request.Context(), kind, model, payload)
	if err != nil || approved == nil || !approved.VerifiedByAdmin ||
		approved.Currency != "USD" || approved.PricingVersion != version ||
		approved.Units <= 0 {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{
			"type": "price_changed", "message": "Media price or options changed. Refresh the quote before retrying"}})
		return false
	}
	unit, err := decimal.NewFromString(approved.UnitSalePriceUSD)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"type": "price_changed", "message": "Media quote invalid"}})
		return false
	}
	price := unit.Mul(decimal.NewFromInt(int64(approved.Units)))
	claimed, err := decimal.NewFromString(expected)
	if err != nil || !price.Equal(claimed) {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{
			"type": "price_changed", "message": "Media price changed. Confirm the updated amount before generating"}})
		return false
	}
	// The claimed amount is not used for billing; server-side approved
	// snapshot remains the only source of truth.
	return true
}

// The quote response is never cached by intermediaries because its eligibility
// and price are bound to a signed-in user and a particular API key ID.
