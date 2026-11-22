package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type creativePriceSettingsFixture struct {
	values map[string]string
}

func (f *creativePriceSettingsFixture) GetAll(context.Context) (map[string]string, error) {
	copy := map[string]string{}
	for key, value := range f.values {
		copy[key] = value
	}
	return copy, nil
}
func (f *creativePriceSettingsFixture) Set(_ context.Context, key, value string) error {
	f.values[key] = value
	return nil
}
func (f *creativePriceSettingsFixture) Delete(_ context.Context, key string) error {
	delete(f.values, key)
	return nil
}

func creativePriceFixture() (*service.CangyuanMediaPriceRegistry, service.CangyuanMediaPriceRule) {
	settings := &creativePriceSettingsFixture{values: map[string]string{}}
	registry := service.NewCangyuanMediaPriceRegistry(settings)
	rule := service.CangyuanMediaPriceRule{
		ID: "video-720p", Model: "video-fixture", Kind: service.CangyuanVideoGeneration,
		Options: map[string]json.RawMessage{"duration": json.RawMessage("5")},
		Snapshot: service.CangyuanMediaPriceSnapshot{
			PricingVersion: "price-v1", Currency: "USD", Tier: "5s", Units: 1,
			UnitSalePriceUSD: "2.50", UnitProviderCostUSD: "1.00",
		},
		Enabled: true,
	}
	return registry, rule
}

func TestCreativeCanvasMediaQuoteExpectationRejectsStaleOrChangedPrices(t *testing.T) {
	gin.SetMode(gin.TestMode)
	registry, rule := creativePriceFixture()
	require.NoError(t, registry.Put(context.Background(), rule))
	h := &CangyuanMediaHandler{public: &service.CangyuanMediaPublicService{Prices: registry}}
	payload := []byte(`{"model":"video-fixture","prompt":"sunrise","duration":5}`)

	check := func(expected, version string) (bool, int) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", nil)
		c.Request.Header.Set("X-Canvas-Quoted-Sale-USD", expected)
		c.Request.Header.Set("X-Canvas-Quoted-Pricing-Version", version)
		return matchCanvasPriceExpectation(h, c, service.CangyuanVideoGeneration, rule.Model, payload),
			c.Writer.Status()
	}

	ok, status := check("2.50000000", "price-v1")
	require.True(t, ok)
	require.Equal(t, http.StatusOK, status)

	ok, status = check("0.01", "price-v1")
	require.False(t, ok)
	require.Equal(t, http.StatusConflict, status)

	rule.Snapshot.PricingVersion = "price-v2"
	require.NoError(t, registry.Put(context.Background(), rule))
	ok, status = check("2.50", "price-v1")
	require.False(t, ok)
	require.Equal(t, http.StatusConflict, status, "changing only version still invalidates consent")

	rule.Snapshot.UnitSalePriceUSD = "3.50"
	require.NoError(t, registry.Put(context.Background(), rule))
	ok, status = check("2.50", "price-v2")
	require.False(t, ok)
	require.Equal(t, http.StatusConflict, status, "higher price must never be billed silently")

	rule.Enabled = false
	require.NoError(t, registry.Put(context.Background(), rule))
	ok, status = check("3.50", "price-v2")
	require.False(t, ok)
	require.Equal(t, http.StatusConflict, status, "disabled pricing never authorizes a paid request")
}

func TestCreativeCanvasPriceExpectationDoesNotChangeOrdinarySDKRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", nil)
	require.True(t, matchCanvasPriceExpectation(nil, c, service.CangyuanVideoGeneration,
		"unrelated-model", nil))
}
