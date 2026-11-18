package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Media discovery is read-only for now. It MUST NOT produce channel bindings,
// activate accounts or suggest that tiered media pricing can be billed with
// the generic OpenAI image-token calculation.
type cangyuanMediaModelPreview struct {
	ID          string `json:"id"`
	BillingMode string `json:"billing_mode"`
	Tiered      bool   `json:"tiered"`
}

type cangyuanMediaModelGroups struct {
	Image []cangyuanMediaModelPreview `json:"image"`
	Video []cangyuanMediaModelPreview `json:"video"`
}

func classifyCangyuanMediaModels(models []string, rawPricing []byte) (cangyuanMediaModelGroups, error) {
	result := cangyuanMediaModelGroups{
		Image: []cangyuanMediaModelPreview{},
		Video: []cangyuanMediaModelPreview{},
	}
	var catalog cangyuanPricingCatalog
	if err := json.Unmarshal(rawPricing, &catalog); err != nil {
		return result, fmt.Errorf("decode cangyuan media catalog: %w", err)
	}
	if !catalog.Success || len(catalog.Data) == 0 {
		return result, errors.New("cangyuan media pricing catalog unavailable")
	}
	authorized := make(map[string]struct{}, len(models))
	for _, model := range models {
		if strings.TrimSpace(model) != "" {
			authorized[model] = struct{}{}
		}
	}
	seen := map[string]struct{}{}
	for _, entry := range catalog.Data {
		if _, ok := authorized[entry.ModelName]; !ok || entry.QuotaType == nil {
			continue
		}
		// Media can be per-request or expression/tier based. Both require
		// completion-based settlement and explicit price configuration.
		if entry.BillingMode != "per_request" && entry.BillingMode != "tiered_expr" {
			continue
		}
		if entry.BillingMode == "tiered_expr" && strings.TrimSpace(entry.BillingExpr) == "" {
			continue
		}
		media := cangyuanMediaModelPreview{
			ID: entry.ModelName,
			BillingMode: entry.BillingMode,
			Tiered: entry.BillingMode == "tiered_expr",
		}
		for _, endpoint := range entry.SupportedEndpointTypes {
			key := entry.ModelName + ":" + endpoint
			if _, duplicate := seen[key]; duplicate {
				continue
			}
			switch endpoint {
			case "openai-image":
				result.Image = append(result.Image, media)
			case "openai-video":
				result.Video = append(result.Video, media)
			default:
				continue
			}
			seen[key] = struct{}{}
		}
	}
	return result, nil
}
