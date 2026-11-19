package service

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

type cangyuanPriceSettingsFixture struct {
	mu     sync.Mutex
	values map[string]string
	err    error
}

func (f *cangyuanPriceSettingsFixture) Set(_ context.Context, key, value string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	f.values[key] = value
	return nil
}
func (f *cangyuanPriceSettingsFixture) GetAll(_ context.Context) (map[string]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	copy := make(map[string]string, len(f.values))
	for key, value := range f.values {
		copy[key] = value
	}
	return copy, nil
}
func (f *cangyuanPriceSettingsFixture) Delete(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	delete(f.values, key)
	return nil
}

func newCangyuanPriceRegistryTest() *CangyuanMediaPriceRegistry {
	return NewCangyuanMediaPriceRegistry(&cangyuanPriceSettingsFixture{values: make(map[string]string)})
}
func cangyuanPriceRuleTest() CangyuanMediaPriceRule {
	return CangyuanMediaPriceRule{
		ID: "video-720p-5s", Model: "video-model", Kind: CangyuanVideoGeneration,
		Options: map[string]json.RawMessage{"duration": json.RawMessage("5"), "resolution": json.RawMessage(`"720p"`)},
		Snapshot: CangyuanMediaPriceSnapshot{
			PricingVersion: "admin-2026-11", Currency: "USD", Tier: "720p-5s", Units: 1,
			UnitSalePriceUSD: "2.50", UnitProviderCostUSD: "1.25",
		},
		Enabled: true,
	}
}
func TestCangyuanMediaPriceRegistryRejectsUnpricedTiersAndUnknownFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registry := newCangyuanPriceRegistryTest()
	rule := cangyuanPriceRuleTest()
	require.NoError(t, registry.Put(ctx, rule))
	rules, err := registry.List(ctx)
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.True(t, rules[0].Snapshot.VerifiedByAdmin, "verification must come only from admin writer")
	quote, err := registry.Quote(ctx, CangyuanVideoGeneration, "video-model",
		[]byte(`{"model":"video-model","prompt":"a rainbow","resolution":"720p","duration":5}`))
	require.NoError(t, err)
	require.Equal(t, "2.50", quote.UnitSalePriceUSD)
	require.Equal(t, "1.25", quote.UnitProviderCostUSD)
	require.True(t, quote.VerifiedByAdmin)

	for name, raw := range map[string]string{
		"higher-resolution":       `{"model":"video-model","prompt":"a rainbow","resolution":"1080p","duration":5}`,
		"longer-video":            `{"model":"video-model","prompt":"a rainbow","resolution":"720p","duration":10}`,
		"unknown-expensive-field": `{"model":"video-model","prompt":"a rainbow","resolution":"720p","duration":5,"cost_multiplier":5}`,
		"wrong-model":             `{"model":"other-model","prompt":"a rainbow","resolution":"720p","duration":5}`,
		"unsupported-response":    `{"model":"video-model","prompt":"a rainbow","resolution":"720p","duration":5,"response_format":"b64_json"}`,
		"missing-duration":        `{"model":"video-model","prompt":"a rainbow","resolution":"720p"}`,
		"invalid-image-count":     `{"model":"video-model","prompt":"a rainbow","resolution":"720p","duration":5,"n":2}`,
		"unexpected-callback":     `{"model":"video-model","prompt":"a rainbow","resolution":"720p","duration":5,"webhook_url":"https://capture.test"}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := registry.Quote(ctx, CangyuanVideoGeneration, "video-model", []byte(raw))
			require.Error(t, err, "unapproved media variant must never receive a price")
		})
	}
	require.NoError(t, registry.Delete(ctx, rule.ID))
	_, err = registry.Quote(ctx, CangyuanVideoGeneration, "video-model",
		[]byte(`{"model":"video-model","prompt":"a rainbow","resolution":"720p","duration":5}`))
	require.ErrorIs(t, err, ErrCangyuanMediaPriceNotFound)
}

func TestCangyuanMediaPriceRegistryImageCountAndSourceCounts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registry := newCangyuanPriceRegistryTest()
	rule := cangyuanPriceRuleTest()
	rule.ID = "image-two-refs"
	rule.Model = "image-model"
	rule.Kind = CangyuanImageEdit
	rule.Options = map[string]json.RawMessage{
		"n":            json.RawMessage("2"),
		"images_count": json.RawMessage("1"),
		"size":         json.RawMessage(`"1024x1024"`),
	}
	rule.Snapshot.Tier = "1024x1024-two"
	rule.Snapshot.Units = 2
	require.NoError(t, registry.Put(ctx, rule))
	payload := []byte(`{"model":"image-model","prompt":"restyle cat","n":2,"size":"1024x1024","images":["https://cdn.example.com/cat.png"]}`)
	quote, err := registry.Quote(ctx, CangyuanImageEdit, "image-model", payload)
	require.NoError(t, err)
	require.Equal(t, 2, quote.Units)
	for _, raw := range [][]byte{
		[]byte(`{"model":"image-model","prompt":"restyle cat","n":1,"size":"1024x1024","images":["https://cdn.example.com/cat.png"]}`),
		[]byte(`{"model":"image-model","prompt":"restyle cat","n":2,"size":"1024x1024","images":["https://cdn.example.com/a.png","https://cdn.example.com/b.png"]}`),
		[]byte(`{"model":"image-model","prompt":"restyle cat","n":2,"size":"1024x1024","images":["http://localhost/private"]}`),
	} {
		_, err := registry.Quote(ctx, CangyuanImageEdit, "image-model", raw)
		require.Error(t, err)
	}
}

func TestCangyuanMediaPriceRegistryRejectsDuplicateEnabledTierAndBadCatalog(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	registry := newCangyuanPriceRegistryTest()
	rule := cangyuanPriceRuleTest()
	require.NoError(t, registry.Put(ctx, rule))
	rule.ID = "duplicate"
	require.NoError(t, registry.Put(ctx, rule))
	_, err := registry.Quote(ctx, CangyuanVideoGeneration, "video-model",
		[]byte(`{"model":"video-model","prompt":"sunrise","resolution":"720p","duration":5}`))
	require.ErrorIs(t, err, ErrCangyuanMediaPriceNotFound, "overlapping enabled rules must fail closed")

	rule.Enabled = false
	require.NoError(t, registry.Put(ctx, rule))
	_, err = registry.Quote(ctx, CangyuanVideoGeneration, "video-model",
		[]byte(`{"model":"video-model","prompt":"sunrise","resolution":"720p","duration":5}`))
	require.NoError(t, err)

	bad := cangyuanPriceRuleTest()
	bad.ID = "bad"
	bad.Options["unknown_option"] = json.RawMessage(`"unreviewed"`)
	require.Error(t, registry.Put(ctx, bad))
	bad = cangyuanPriceRuleTest()
	bad.Snapshot.UnitSalePriceUSD = "-1"
	require.Error(t, registry.Put(ctx, bad))
	bad = cangyuanPriceRuleTest()
	bad.Model = "video-*"
	require.Error(t, registry.Put(ctx, bad))
	bad = cangyuanPriceRuleTest()
	bad.Kind = CangyuanImageGeneration
	require.Error(t, registry.Put(ctx, bad), "image prices must specify n")

	settings, ok := registry.Settings.(*cangyuanPriceSettingsFixture)
	require.True(t, ok)
	settings.values["cangyuan_media_price_v1:malformed"] = "not JSON"
	_, err = registry.Quote(ctx, CangyuanVideoGeneration, "video-model",
		[]byte(`{"model":"video-model","prompt":"sunrise","resolution":"720p","duration":5}`))
	require.Error(t, err, "invalid stored catalog must fail closed")
	settings.err = errors.New("database unavailable")
	_, err = registry.List(ctx)
	require.Error(t, err)
}
