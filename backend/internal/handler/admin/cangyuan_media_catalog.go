package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

type cangyuanMediaCatalogEntry struct {
	ModelName    string      `json:"model_name"`
	QuotaType    *int        `json:"quota_type"`
	BillingMode  string      `json:"billing_mode"`
	ModelPrice   json.Number `json:"model_price"`
	RequestUnit  string      `json:"request_unit"`
	Endpoints    []string    `json:"supported_endpoint_types"`
}

type cangyuanMediaCatalogSource struct {
	Success        bool                        `json:"success"`
	PricingVersion string                      `json:"pricing_version"`
	Data           []cangyuanMediaCatalogEntry `json:"data"`
}

type cangyuanMediaCatalogItem struct {
	ID           string      `json:"id"`
	Kind         string      `json:"kind"`
	BillingMode  string      `json:"billing_mode"`
	ModelPrice   json.Number `json:"model_price"`
	RequestUnit  string      `json:"request_unit"`
	Tiered       bool        `json:"tiered"`
}

type cangyuanMediaCatalogResponse struct {
	PricingVersion string                     `json:"pricing_version"`
	Items          []cangyuanMediaCatalogItem `json:"items"`
}

// parseCangyuanMediaCatalog exposes only public, non-sensitive model facts.
// The upstream model_price field is a provider display number, NOT an
// administrator-approved USD provider cost or retail sale price.
func parseCangyuanMediaCatalog(raw []byte) (*cangyuanMediaCatalogResponse, error) {
	var source cangyuanMediaCatalogSource
	if err := json.Unmarshal(raw, &source); err != nil || !source.Success ||
		len(source.Data) == 0 || len(source.Data) > 2000 {
		return nil, errors.New("cangyuan media catalog is invalid")
	}
	if len(source.PricingVersion) == 0 || len(source.PricingVersion) > 128 {
		return nil, errors.New("cangyuan media catalog version is missing")
	}
	result := &cangyuanMediaCatalogResponse{
		PricingVersion: source.PricingVersion,
		Items:          make([]cangyuanMediaCatalogItem, 0),
	}
	seen := make(map[string]bool)
	for _, entry := range source.Data {
		if entry.QuotaType == nil || *entry.QuotaType != 1 ||
			strings.TrimSpace(entry.ModelName) != entry.ModelName ||
			len(entry.ModelName) == 0 || len(entry.ModelName) > 256 ||
			strings.ContainsAny(entry.ModelName, "*?") {
			continue
		}
		if entry.BillingMode != "per_request" && entry.BillingMode != "tiered_expr" {
			continue
		}
		// A null or missing reference price must not be represented as zero.
		if entry.ModelPrice == "" {
			continue
		}
		price, err := entry.ModelPrice.Float64()
		if err != nil || price < 0 || price > 100000 {
			continue
		}
		for _, endpoint := range entry.Endpoints {
			kind := ""
			switch endpoint {
			case "openai-image":
				kind = "image_generation"
			case "openai-video":
				kind = "video_generation"
			default:
				continue
			}
			if seen[kind+":"+entry.ModelName] {
				continue
			}
			seen[kind+":"+entry.ModelName] = true
			result.Items = append(result.Items, cangyuanMediaCatalogItem{
				ID: entry.ModelName, Kind: kind,
				BillingMode: entry.BillingMode,
				ModelPrice: entry.ModelPrice,
				RequestUnit: entry.RequestUnit,
				Tiered: entry.BillingMode == "tiered_expr",
			})
		}
	}
	if len(result.Items) == 0 {
		return nil, errors.New("cangyuan media catalog has no recognized models")
	}
	sort.Slice(result.Items, func(i, j int) bool {
		if result.Items[i].Kind == result.Items[j].Kind {
			return result.Items[i].ID < result.Items[j].ID
		}
		return result.Items[i].Kind < result.Items[j].Kind
	})
	return result, nil
}

// Catalog is admin-only, reads a fixed official HTTPS origin, and deliberately
// omits provider billing expressions and any API-key scoped entitlements.
func (h *CangyuanMediaPriceHandler) Catalog(c *gin.Context) {
	if h == nil || h.registry == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Cangyuan media catalog unavailable"})
		return
	}
	raw, err := fetchCangyuanPublicPricing(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Cangyuan public pricing temporarily unavailable"})
		return
	}
	catalog, err := parseCangyuanMediaCatalog(raw)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Cangyuan public media model catalog invalid"})
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, catalog)
}
