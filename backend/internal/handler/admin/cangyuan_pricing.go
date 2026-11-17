package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	cangyuanPublicPricingURL = "https://ai.cangyuansuanli.cn/api/pricing"
	cangyuanPricingBodyLimit = 4 << 20
)

type cangyuanPricingEntry struct {
	ModelName              string   `json:"model_name"`
	QuotaType              *int     `json:"quota_type"`
	SupportedEndpointTypes []string `json:"supported_endpoint_types"`
	BillingMode            string   `json:"billing_mode"`
	BillingExpr            string   `json:"billing_expr"`
}

type cangyuanPricingCatalog struct {
	Success bool                   `json:"success"`
	Data    []cangyuanPricingEntry `json:"data"`
}

type cangyuanChatPricingInfo struct {
	BillingMode string `json:"billing_mode"`
	Tiered      bool   `json:"tiered"`
}

// filterCangyuanChatModels intersects the key-scoped /v1/models list with
// the independent /api/pricing endpoint catalog. It explicitly excludes
// media, per-request and unclassified products from the chat-only preset.
// Pricing here is informational: expressions are NOT evaluated or imported.
func filterCangyuanChatModels(models []string, rawPricing []byte) ([]string, map[string]cangyuanChatPricingInfo, error) {
	var catalog cangyuanPricingCatalog
	if err := json.Unmarshal(rawPricing, &catalog); err != nil {
		return nil, nil, fmt.Errorf("decode cangyuan public pricing: %w", err)
	}
	if !catalog.Success || len(catalog.Data) == 0 {
		return nil, nil, errors.New("cangyuan public pricing catalog is unavailable")
	}

	eligible := make(map[string]cangyuanChatPricingInfo, len(catalog.Data))
	for _, entry := range catalog.Data {
		// quota_type=0 is token-based. A missing quota_type is not evidence
		// of token billing and must fail closed. Some audio products advertise
		// an OpenAI-compatible endpoint but are billed per request.
		if entry.ModelName == "" || entry.QuotaType == nil || *entry.QuotaType != 0 {
			continue
		}
		chatEndpoint := false
		for _, endpoint := range entry.SupportedEndpointTypes {
			if endpoint == "openai" {
				chatEndpoint = true
				break
			}
		}
		if !chatEndpoint {
			continue
		}
		if entry.BillingMode == "per_request" {
			continue
		}
		eligible[entry.ModelName] = cangyuanChatPricingInfo{
			BillingMode: entry.BillingMode,
			Tiered:      entry.BillingMode == "tiered_expr" || strings.TrimSpace(entry.BillingExpr) != "",
		}
	}

	// Preserve upstream model order, deduplicate names, and never add models
	// the supplied API key did not explicitly authorize.
	filtered := make([]string, 0, len(models))
	pricing := make(map[string]cangyuanChatPricingInfo)
	for _, model := range models {
		if info, ok := eligible[model]; ok {
			if _, duplicate := pricing[model]; duplicate {
				continue
			}
			filtered = append(filtered, model)
			pricing[model] = info
		}
	}
	return filtered, pricing, nil
}

func fetchCangyuanPublicPricing(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cangyuanPublicPricingURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	client := &http.Client{
		Timeout: 8 * time.Second,
		// Never follow a redirect when fetching pricing. In particular this
		// prevents a changed provider URL from causing internal network reads.
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch cangyuan public pricing: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cangyuan public pricing returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, cangyuanPricingBodyLimit+1))
	if err != nil {
		return nil, fmt.Errorf("read cangyuan public pricing: %w", err)
	}
	if len(body) > cangyuanPricingBodyLimit {
		return nil, errors.New("cangyuan public pricing exceeded response size limit")
	}
	return body, nil
}
