package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const cangyuanMediaPriceSettingPrefix = "cangyuan_media_price_v1:"
const cangyuanMediaPriceMaxRequestBytes = 1 << 20

var cangyuanPriceRuleIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,79}$`)

var ErrCangyuanMediaPriceNotFound = errors.New("no verified cangyuan media price matches this exact request")

// CangyuanMediaPriceRule is stored only from the 2FA-protected admin API.
// Client API requests cannot submit a verified sell price or an alternate
// pricing version; both are resolved from this persisted registry.
type CangyuanMediaPriceRule struct {
	ID       string                     `json:"id"`
	Model    string                     `json:"model"`
	Kind     CangyuanMediaKind          `json:"kind"`
	Options  map[string]json.RawMessage `json:"options"`
	Snapshot CangyuanMediaPriceSnapshot `json:"snapshot"`
	Enabled  bool                       `json:"enabled"`
}

// SettingRepository satisfies this deliberately small port.
type CangyuanMediaPriceSettings interface {
	Set(ctx context.Context, key, value string) error
	GetAll(ctx context.Context) (map[string]string, error)
	Delete(ctx context.Context, key string) error
}

type CangyuanMediaPriceRegistry struct {
	Settings CangyuanMediaPriceSettings
}

func NewCangyuanMediaPriceRegistry(settings CangyuanMediaPriceSettings) *CangyuanMediaPriceRegistry {
	return &CangyuanMediaPriceRegistry{Settings: settings}
}

func (s *CangyuanMediaPriceRegistry) Put(ctx context.Context, rule CangyuanMediaPriceRule) error {
	if s == nil || s.Settings == nil {
		return errors.New("cangyuan media price storage is unavailable")
	}
	if err := validateCangyuanPriceRule(&rule); err != nil {
		return err
	}
	// Verification is an assertion by the protected admin writer, not
	// an assertion trusted from the request body or media submitter.
	rule.Snapshot.VerifiedByAdmin = true
	raw, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	return s.Settings.Set(ctx, cangyuanMediaPriceSettingPrefix+rule.ID, string(raw))
}

func (s *CangyuanMediaPriceRegistry) Delete(ctx context.Context, id string) error {
	if s == nil || s.Settings == nil {
		return errors.New("cangyuan media price storage is unavailable")
	}
	if !cangyuanPriceRuleIDPattern.MatchString(id) {
		return errors.New("invalid cangyuan media price rule ID")
	}
	return s.Settings.Delete(ctx, cangyuanMediaPriceSettingPrefix+id)
}

func (s *CangyuanMediaPriceRegistry) List(ctx context.Context) ([]CangyuanMediaPriceRule, error) {
	if s == nil || s.Settings == nil {
		return nil, errors.New("cangyuan media price storage is unavailable")
	}
	raw, err := s.Settings.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	entries := make([]CangyuanMediaPriceRule, 0)
	for key, value := range raw {
		if !strings.HasPrefix(key, cangyuanMediaPriceSettingPrefix) {
			continue
		}
		var rule CangyuanMediaPriceRule
		if err := json.Unmarshal([]byte(value), &rule); err != nil {
			return nil, errors.New("stored cangyuan media price rule is invalid")
		}
		if rule.ID != strings.TrimPrefix(key, cangyuanMediaPriceSettingPrefix) {
			return nil, errors.New("stored cangyuan media price rule identity mismatch")
		}
		if err := validateCangyuanPriceRule(&rule); err != nil || !rule.Snapshot.VerifiedByAdmin {
			return nil, errors.New("stored cangyuan media price rule is unverified")
		}
		entries = append(entries, rule)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	return entries, nil
}

// Quote resolves exactly one enabled admin-reviewed variant from the
// caller's real generation parameters. Prompt and source URL values are
// excluded from price matching; source counts/presence are retained.
// Unknown options fail closed instead of silently inheriting a cheap tier.
func (s *CangyuanMediaPriceRegistry) Quote(ctx context.Context, kind CangyuanMediaKind, model string, payload []byte) (*CangyuanMediaPriceSnapshot, error) {
	options, err := normalizeCangyuanMediaPriceRequest(kind, model, payload)
	if err != nil {
		return nil, err
	}
	rules, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	matches := 0
	var quote CangyuanMediaPriceSnapshot
	for _, rule := range rules {
		if !rule.Enabled || rule.Kind != kind || rule.Model != model {
			continue
		}
		if equalCangyuanMediaPriceOptions(rule.Options, options) {
			matches++
			quote = rule.Snapshot
		}
	}
	if matches != 1 {
		// An overlapping active catalog is unsafe, not a reason to choose the
		// cheapest rule or silently fall back to provider pricing expressions.
		return nil, ErrCangyuanMediaPriceNotFound
	}
	if err := validateCangyuanMediaRequestQuantity(kind, options, &quote); err != nil {
		return nil, err
	}
	return &quote, nil
}

// To be safe, a priced request may only contain enumerated non-secret
// provider options. New upstream features require explicit review.
var cangyuanMediaPriceOptionKeys = map[string]bool{
	"n": true, "size": true, "quality": true, "duration": true,
	"seconds": true, "resolution": true, "aspect_ratio": true,
	"ratio": true, "width": true, "height": true, "fps": true,
	"style": true, "output_format": true, "seed": true, "watermark": true,
	"mode": true, "preset": true,
}

var cangyuanMediaPriceSourceArrays = map[string]bool{
	"images": true, "reference_image_urls": true,
	"reference_videos": true, "reference_audios": true,
}
var cangyuanMediaPriceSourceSingles = map[string]bool{
	"mask": true, "first_image_url": true, "last_image_url": true,
}

func normalizeCangyuanMediaPriceRequest(kind CangyuanMediaKind, model string, payload []byte) (map[string]json.RawMessage, error) {
	if len(payload) == 0 || len(payload) > cangyuanMediaPriceMaxRequestBytes {
		return nil, errors.New("invalid cangyuan media price request size")
	}
	var request map[string]json.RawMessage
	if err := json.Unmarshal(payload, &request); err != nil || request == nil {
		return nil, errors.New("cangyuan media price request must be a JSON object")
	}
	var suppliedModel, prompt string
	if json.Unmarshal(request["model"], &suppliedModel) != nil || suppliedModel != model ||
		json.Unmarshal(request["prompt"], &prompt) != nil || strings.TrimSpace(prompt) == "" {
		return nil, errors.New("cangyuan media price request model or prompt mismatch")
	}
	if _, err := kind.path(); err != nil {
		return nil, err
	}
	if err := validateCangyuanMediaReferences(request); err != nil {
		return nil, err
	}
	opts := make(map[string]json.RawMessage)
	for key, value := range request {
		switch {
		case key == "model" || key == "prompt":
			continue
		case key == "async":
			var enabled bool
			if json.Unmarshal(value, &enabled) != nil || !enabled {
				return nil, errors.New("cangyuan priced media requires async mode")
			}
			continue
		case key == "response_format":
			var format string
			if json.Unmarshal(value, &format) != nil || format != "url" {
				return nil, errors.New("cangyuan priced media only supports URL results")
			}
			continue
		case cangyuanMediaPriceOptionKeys[key]:
			normalized, err := canonicalCangyuanMediaPriceScalar(value)
			if err != nil {
				return nil, fmt.Errorf("invalid cangyuan media price option %s", key)
			}
			opts[key] = normalized
		case cangyuanMediaPriceSourceArrays[key]:
			var items []string
			if json.Unmarshal(value, &items) != nil || len(items) == 0 || len(items) > 16 {
				return nil, errors.New("cangyuan media reference array is invalid")
			}
			opts[key+"_count"] = json.RawMessage(strconv.Itoa(len(items)))
		case cangyuanMediaPriceSourceSingles[key]:
			opts[key+"_present"] = json.RawMessage("true")
		default:
			return nil, fmt.Errorf("unpriced cangyuan media request option: %s", key)
		}
	}
	if kind != CangyuanVideoGeneration {
		if _, ok := opts["n"]; !ok {
			opts["n"] = json.RawMessage("1")
		}
	}
	return opts, nil
}

func canonicalCangyuanMediaPriceScalar(raw json.RawMessage) (json.RawMessage, error) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	if value == nil {
		return nil, errors.New("null price option")
	}
	switch value.(type) {
	case bool, string, json.Number:
	default:
		return nil, errors.New("complex options require explicit capability review")
	}
	return json.Marshal(value)
}

func validateCangyuanPriceRule(rule *CangyuanMediaPriceRule) error {
	if rule == nil || !cangyuanPriceRuleIDPattern.MatchString(rule.ID) ||
		rule.Model == "" || strings.TrimSpace(rule.Model) != rule.Model ||
		len(rule.Model) > 256 || strings.ContainsAny(rule.Model, "*?") {
		return errors.New("invalid cangyuan media price rule identity or model")
	}
	if _, err := rule.Kind.path(); err != nil {
		return err
	}
	// Snapshot verification is set by Put only; values and format must still
	// be validated before allowing an admin to publish them.
	verified := rule.Snapshot
	verified.VerifiedByAdmin = true
	if err := validateCangyuanMediaPriceSnapshot(rule.Kind, &verified); err != nil {
		return err
	}
	if len(rule.Options) > 24 {
		return errors.New("too many cangyuan media price options")
	}
	for key, value := range rule.Options {
		if !cangyuanMediaPriceOptionKeys[key] && !strings.HasSuffix(key, "_count") && !strings.HasSuffix(key, "_present") {
			return errors.New("cangyuan media price rule contains unsupported option")
		}
		if strings.HasSuffix(key, "_count") && !cangyuanMediaPriceSourceArrays[strings.TrimSuffix(key, "_count")] ||
			strings.HasSuffix(key, "_present") && !cangyuanMediaPriceSourceSingles[strings.TrimSuffix(key, "_present")] {
			return errors.New("cangyuan media price rule contains unknown reference option")
		}
		normalized, err := canonicalCangyuanMediaPriceScalar(value)
		if err != nil || !bytes.Equal(normalized, value) {
			return fmt.Errorf("invalid cangyuan media price rule option %s", key)
		}
	}
	if rule.Kind != CangyuanVideoGeneration {
		var n int
		if json.Unmarshal(rule.Options["n"], &n) != nil || n != rule.Snapshot.Units {
			return errors.New("cangyuan image price rule must specify matching n")
		}
	} else if _, n := rule.Options["n"]; n {
		return errors.New("cangyuan video price rule cannot include n")
	}
	return nil
}

func equalCangyuanMediaPriceOptions(a, b map[string]json.RawMessage) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		other, exists := b[key]
		if !exists || !bytes.Equal(value, other) {
			return false
		}
	}
	return true
}
