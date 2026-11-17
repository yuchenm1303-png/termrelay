package admin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterCangyuanChatModels_PricingAndKeyIntersection(t *testing.T) {
	t.Parallel()

	pricing := []byte(`{
		"success": true,
		"data": [
			{"model_name":"chat-standard","quota_type":0,"supported_endpoint_types":["openai"]},
			{"model_name":"chat-tiered","quota_type":0,"supported_endpoint_types":["anthropic","openai"],"billing_mode":"tiered_expr","billing_expr":"tier(\"fast\", p * 1 + c * 4)"},
			{"model_name":"image","quota_type":1,"supported_endpoint_types":["openai-image"],"billing_mode":"per_request"},
			{"model_name":"video","quota_type":1,"supported_endpoint_types":["openai-video"],"billing_mode":"per_request"},
			{"model_name":"audio-per-request","quota_type":1,"supported_endpoint_types":["openai"],"billing_mode":"per_request"},
			{"model_name":"untyped","supported_endpoint_types":["openai"]},
			{"model_name":"anthropic-only","quota_type":0,"supported_endpoint_types":["anthropic"]},
			{"model_name":"pricing-only","quota_type":0,"supported_endpoint_types":["openai"]}
		]
	}`)
	available := []string{
		"video", "chat-tiered", "chat-standard", "chat-tiered",
		"image", "audio-per-request", "untyped", "anthropic-only", "not-in-pricing",
	}

	models, details, err := filterCangyuanChatModels(available, pricing)
	require.NoError(t, err)
	require.Equal(t, []string{"chat-tiered", "chat-standard"}, models)
	require.Len(t, details, 2)
	require.True(t, details["chat-tiered"].Tiered)
	require.Equal(t, "tiered_expr", details["chat-tiered"].BillingMode)
	require.False(t, details["chat-standard"].Tiered)
	require.NotContains(t, details, "pricing-only", "a public price never grants model authorization")
}

func TestFilterCangyuanChatModels_FailsClosedOnInvalidCatalog(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		raw  string
	}{
		{name: "invalid JSON", raw: "{not json"},
		{name: "false success", raw: `{"success":false,"data":[{"model_name":"m","quota_type":0,"supported_endpoint_types":["openai"]}]}`},
		{name: "empty catalog", raw: `{"success":true,"data":[]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := filterCangyuanChatModels([]string{"m"}, []byte(tc.raw))
			require.Error(t, err)
		})
	}
}

func TestFilterCangyuanChatModels_UnknownModelsStayClosed(t *testing.T) {
	t.Parallel()

	models, details, err := filterCangyuanChatModels(
		[]string{"media-only", "not-listed"},
		[]byte(`{"success":true,"data":[{"model_name":"media-only","quota_type":1,"supported_endpoint_types":["openai-image"]}]}`),
	)
	require.NoError(t, err)
	require.Empty(t, models)
	require.Empty(t, details)
}
