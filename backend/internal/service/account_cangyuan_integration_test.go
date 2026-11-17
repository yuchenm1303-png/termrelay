package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/stretchr/testify/require"
)

// Cangyuan is an OpenAI-compatible New API gateway, not the official OpenAI API.
// Its model discovery includes non-chat / asynchronous / per-job products.
// The admin preset must only schedule explicitly approved chat models.
func TestCangyuanUpstreamPresetRoutingAndModelAllowlist(t *testing.T) {
	const base = "https://ai.cangyuansuanli.cn"
	account := &Account{
		Platform: PlatformOpenAI,
		Type: AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key": "test-key-not-a-real-credential",
			"base_url": base,
			"model_mapping": map[string]any{"verified-chat-model": "verified-chat-model"},
		},
		Extra: map[string]any{
			"upstream_vendor": "cangyuan",
			"openai_passthrough": false,
			"openai_responses_mode": "force_chat_completions",
		},
	}

	require.False(t, account.IsOpenAIPassthroughEnabled(),
		"OpenAI passthrough disables the account's model allowlist")
	require.True(t, account.IsModelSupported("verified-chat-model"))
	require.False(t, account.IsModelSupported("unverified-video-or-image-model"))
	require.False(t, openai_compat.ShouldUseResponsesAPI(account.Extra),
		"Cangyuan chat calls must not be sent to an unverified Responses endpoint")
	require.Equal(t, base+"/v1/models", buildOpenAIModelsURL(base))
	require.Equal(t, base+"/v1/chat/completions", buildOpenAIChatCompletionsURL(base))
}
