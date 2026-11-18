package service

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
)

// isCangyuanAccount keeps New API vendor-specific guardrails scoped to the
// explicit admin preset. Ordinary OpenAI-compatible accounts are unaffected.
func isCangyuanAccount(account *Account) bool {
	if account == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(account.GetExtraString("upstream_vendor")), "cangyuan") {
		return true
	}
	// The official upstream URL is an independent safety signal. Deleting the
	// vendor flag via an extra-only/bulk update must never silently turn a
	// paused, allowlisted relay into unrestricted OpenAI passthrough.
	base, err := url.Parse(strings.TrimSpace(account.GetCredential("base_url")))
	return err == nil && strings.EqualFold(base.Hostname(), "ai.cangyuansuanli.cn")
}

// isCangyuanMediaOnlyAccount separates staged image/video accounts from
// generic OpenAI chat routing. Even an administrator's later bulk setting
// of schedulable=true must never make a media model eligible for chat.
func isCangyuanMediaOnlyAccount(account *Account) bool {
	if !isCangyuanAccount(account) {
		return false
	}
	mediaOnly, _ := account.Extra["cangyuan_media_only"].(bool)
	return mediaOnly
}

func validateCangyuanMediaOnlyModels(account *Account) error {
	raw, ok := account.Extra["cangyuan_media_models"].(map[string]any)
	if !ok {
		return fmt.Errorf("cangyuan media-only account requires explicit image/video model lists")
	}
	count := 0
	for capability, models := range raw {
		if capability != "image" && capability != "video" {
			return fmt.Errorf("cangyuan media-only account includes unsupported capability")
		}
		var names []string
		switch value := models.(type) {
		case []string:
			names = value
		case []any:
			for _, item := range value {
				name, valid := item.(string)
				if !valid {
					return fmt.Errorf("cangyuan media-only model list contains a non-string ID")
				}
				names = append(names, name)
			}
		default:
			return fmt.Errorf("cangyuan media-only model lists must be arrays")
		}
		for _, name := range names {
			if strings.TrimSpace(name) != name || name == "" || len(name) > 256 ||
				strings.ContainsAny(name, "*?") {
				return fmt.Errorf("cangyuan media-only model IDs must be exact and non-wildcard")
			}
			count++
			if count > 256 {
				return fmt.Errorf("cangyuan media-only account has too many models")
			}
		}
	}
	if count == 0 {
		return fmt.Errorf("cangyuan media-only account requires at least one model")
	}
	return nil
}

// AllowsCangyuanMediaModel is a dedicated media-only allowlist. It must never
// be used as permission to schedule an OpenAI chat/image passthrough request.
// Future media POST handlers must also enforce group entitlement, approved
// prices, account status and balance reservation before invoking the client.
func (account *Account) AllowsCangyuanMediaModel(kind CangyuanMediaKind, model string) bool {
	if !isCangyuanMediaOnlyAccount(account) || strings.TrimSpace(model) != model || model == "" {
		return false
	}
	if err := validateCangyuanAccount(account); err != nil {
		return false
	}
	capability := ""
	switch kind {
	case CangyuanImageGeneration, CangyuanImageEdit:
		capability = "image"
	case CangyuanVideoGeneration:
		capability = "video"
	default:
		return false
	}
	lists, ok := account.Extra["cangyuan_media_models"].(map[string]any)
	if !ok {
		return false
	}
	switch items := lists[capability].(type) {
	case []string:
		for _, candidate := range items {
			if candidate == model {
				return true
			}
		}
	case []any:
		for _, raw := range items {
			if candidate, ok := raw.(string); ok && candidate == model {
				return true
			}
		}
	}
	return false
}

// validateCangyuanAccount prevents accidentally exposing image/video/async
// products through the generic chat relay. These safeguards are enforced on
// the server in addition to the admin form, which can be bypassed by API calls.
func validateCangyuanAccount(account *Account) error {
	if !isCangyuanAccount(account) {
		return nil
	}
	if account.Platform != PlatformOpenAI || account.Type != AccountTypeAPIKey {
		return fmt.Errorf("cangyuan upstream must use the OpenAI API-key protocol")
	}
	base, err := url.Parse(strings.TrimSpace(account.GetCredential("base_url")))
	if err != nil || base.Scheme != "https" || base.Hostname() != "ai.cangyuansuanli.cn" ||
		base.Port() != "" || base.User != nil || (base.Path != "" && base.Path != "/") ||
		base.RawQuery != "" || base.Fragment != "" {
		return fmt.Errorf("cangyuan upstream requires the official HTTPS API root")
	}
	if enabled, _ := account.Extra["openai_passthrough"].(bool); enabled {
		return fmt.Errorf("cangyuan upstream cannot enable unrestricted OpenAI passthrough")
	}
	if openai_compat.ShouldUseResponsesAPI(account.Extra) {
		return fmt.Errorf("cangyuan chat upstream must use force_chat_completions mode")
	}
	mapping := account.GetModelMapping()
	if raw, exists := account.Extra["cangyuan_media_only"]; exists {
		if _, ok := raw.(bool); !ok {
			return fmt.Errorf("cangyuan media-only flag must be a boolean")
		}
	}
	if isCangyuanMediaOnlyAccount(account) {
		if len(mapping) > 0 {
			return fmt.Errorf("cangyuan media-only account cannot advertise chat models")
		}
		return validateCangyuanMediaOnlyModels(account)
	}
	if _, exists := account.Extra["cangyuan_media_models"]; exists {
		return fmt.Errorf("cangyuan chat account cannot mix unpriced media models")
	}
	if len(mapping) == 0 {
		return fmt.Errorf("cangyuan upstream requires at least one explicit model mapping")
	}
	for requested, upstream := range mapping {
		if strings.TrimSpace(requested) == "" || strings.TrimSpace(upstream) == "" ||
			strings.ContainsAny(requested, "*?") || strings.ContainsAny(upstream, "*?") {
			return fmt.Errorf("cangyuan upstream model mapping must use exact, non-wildcard model IDs")
		}
	}
	return nil
}
