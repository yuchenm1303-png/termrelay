package service

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
)

const accountPoolProviderIDExtraKey = "provider_id"

// AccountPoolProviderIdentity is the scheduler-facing provider identity. It is
// deliberately separate from Account: one provider may own many API-key/OAuth
// accounts, while credentials and per-account health remain account-scoped.
type AccountPoolProviderIdentity struct {
	Key      string
	Platform string
	Origin   string
}

// ProviderIdentity derives a stable non-secret provider identity. Operators can
// set Extra["provider_id"] for an explicit logical provider. Otherwise API-key
// accounts sharing the same platform + upstream origin are grouped together.
// API keys, OAuth tokens and other credential material never participate in the
// identity or telemetry key.
func (a *Account) ProviderIdentity() AccountPoolProviderIdentity {
	if a == nil {
		return AccountPoolProviderIdentity{}
	}
	platform := strings.ToLower(strings.TrimSpace(a.Platform))
	if explicit := strings.TrimSpace(a.GetExtraString(accountPoolProviderIDExtraKey)); explicit != "" {
		return AccountPoolProviderIdentity{
			Key:      "provider:" + explicit,
			Platform: platform,
		}
	}

	origin := normalizeAccountPoolProviderOrigin(a.GetCredential("base_url"))
	material := platform
	if origin != "" {
		material += "|" + origin
	}
	digest := sha256.Sum256([]byte(material))
	return AccountPoolProviderIdentity{
		Key:      "derived:" + hex.EncodeToString(digest[:8]),
		Platform: platform,
		Origin:   origin,
	}
}

func normalizeAccountPoolProviderOrigin(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return strings.TrimRight(raw, "/")
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	return parsed.String()
}
