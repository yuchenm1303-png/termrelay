package admin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClassifyCangyuanMediaModels_KeyScopedAndBillingGuarded(t *testing.T) {
	t.Parallel()
	raw := []byte(`{
		"success":true,
		"data":[
			{"model_name":"img-per-request","quota_type":1,"supported_endpoint_types":["openai-image"],"billing_mode":"per_request"},
			{"model_name":"img-tiered","quota_type":1,"supported_endpoint_types":["openai-image"],"billing_mode":"tiered_expr","billing_expr":"tier(\"720p\", n * 1.5)"},
			{"model_name":"vid-per-request","quota_type":1,"supported_endpoint_types":["openai-video"],"billing_mode":"per_request"},
			{"model_name":"vid-tiered","quota_type":1,"supported_endpoint_types":["openai-video"],"billing_mode":"tiered_expr","billing_expr":"tier(\"default\", n * 8.0)"},
			{"model_name":"vid-unpriced","quota_type":1,"supported_endpoint_types":["openai-video"],"billing_mode":"tiered_expr"},
			{"model_name":"img-unknown","quota_type":1,"supported_endpoint_types":["openai-image"]},
			{"model_name":"img-no-quota","supported_endpoint_types":["openai-image"],"billing_mode":"per_request"},
			{"model_name":"chat","quota_type":0,"supported_endpoint_types":["openai"],"billing_mode":"tiered_expr","billing_expr":"tier(\"test\",p+c)"},
			{"model_name":"provider-only-video","quota_type":1,"supported_endpoint_types":["openai-video"],"billing_mode":"per_request"}
		]
	}`)
	models, err := classifyCangyuanMediaModels(
		[]string{"vid-per-request", "img-tiered", "img-per-request", "vid-tiered", "vid-unpriced", "img-unknown", "img-no-quota", "chat"},
		raw,
	)
	require.NoError(t, err)
	require.Equal(t, []cangyuanMediaModelPreview{
		{ID: "img-per-request", BillingMode: "per_request", Tiered: false},
		{ID: "img-tiered", BillingMode: "tiered_expr", Tiered: true},
	}, models.Image)
	require.Equal(t, []cangyuanMediaModelPreview{
		{ID: "vid-per-request", BillingMode: "per_request", Tiered: false},
		{ID: "vid-tiered", BillingMode: "tiered_expr", Tiered: true},
	}, models.Video)
}

func TestClassifyCangyuanMediaModels_FailClosedOnInvalidPricing(t *testing.T) {
	t.Parallel()
	for _, raw := range [][]byte{
		[]byte("invalid"),
		[]byte(`{"success":false,"data":[{}]}`),
		[]byte(`{"success":true,"data":[]}`),
	} {
		_, err := classifyCangyuanMediaModels([]string{"image"}, raw)
		require.Error(t, err)
	}
}
