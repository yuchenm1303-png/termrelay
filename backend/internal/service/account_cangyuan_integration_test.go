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
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":       "test-key-not-a-real-credential",
			"base_url":      base,
			"model_mapping": map[string]any{"verified-chat-model": "verified-chat-model"},
		},
		Extra: map[string]any{
			"upstream_vendor":       "cangyuan",
			"openai_passthrough":    false,
			"openai_responses_mode": "force_chat_completions",
		},
	}

	require.NoError(t, validateCangyuanAccount(account))
	require.False(t, account.IsOpenAIPassthroughEnabled(),
		"OpenAI passthrough disables the account's model allowlist")
	require.True(t, account.IsModelSupported("verified-chat-model"))
	require.False(t, account.IsModelSupported("unverified-video-or-image-model"))
	require.False(t, account.SupportsOpenAIImageCapability(OpenAIImagesCapabilityBasic))
	require.False(t, account.SupportsOpenAIImageCapability(OpenAIImagesCapabilityNative))
	require.False(t, openai_compat.ShouldUseResponsesAPI(account.Extra),
		"Cangyuan chat calls must not be sent to an unverified Responses endpoint")
	require.Equal(t, base+"/v1/models", buildOpenAIModelsURL(base))
	require.Equal(t, base+"/v1/chat/completions", buildOpenAIChatCompletionsURL(base))
}

func TestCangyuanAccountRejectsUnsafeConfiguration(t *testing.T) {
	t.Parallel()

	valid := func() *Account {
		return &Account{
			Platform: PlatformOpenAI,
			Type:     AccountTypeAPIKey,
			Credentials: map[string]any{
				"api_key":       "test-key",
				"base_url":      "https://ai.cangyuansuanli.cn",
				"model_mapping": map[string]any{"safe-model": "safe-model"},
			},
			Extra: map[string]any{
				"upstream_vendor":       "cangyuan",
				"openai_passthrough":    false,
				"openai_responses_mode": "force_chat_completions",
			},
		}
	}

	t.Run("unrestricted passthrough", func(t *testing.T) {
		account := valid()
		account.Extra["openai_passthrough"] = true
		require.ErrorContains(t, validateCangyuanAccount(account), "unrestricted")
		require.False(t, account.IsOpenAIPassthroughEnabled(), "runtime must ignore unsafe extra-only changes")
		require.True(t, account.IsModelSupported("safe-model"))
		require.False(t, account.IsModelSupported("unlisted-model"))
	})
	t.Run("responses mode", func(t *testing.T) {
		account := valid()
		account.Extra["openai_responses_mode"] = "auto"
		require.ErrorContains(t, validateCangyuanAccount(account), "force_chat_completions")
	})
	t.Run("missing explicit model allowlist", func(t *testing.T) {
		account := valid()
		delete(account.Credentials, "model_mapping")
		require.ErrorContains(t, validateCangyuanAccount(account), "explicit")
		require.False(t, account.IsModelSupported("any-model"), "empty Cangyuan mapping must fail closed")
	})
	t.Run("wildcard model mapping", func(t *testing.T) {
		account := valid()
		account.Credentials["model_mapping"] = map[string]any{"*": "safe-model"}
		require.ErrorContains(t, validateCangyuanAccount(account), "non-wildcard")
	})
	t.Run("insecure base url", func(t *testing.T) {
		account := valid()
		account.Credentials["base_url"] = "http://ai.cangyuansuanli.cn"
		require.ErrorContains(t, validateCangyuanAccount(account), "HTTPS")
	})
	t.Run("untrusted HTTPS host", func(t *testing.T) {
		account := valid()
		account.Credentials["base_url"] = "https://unrelated.example.com"
		require.ErrorContains(t, validateCangyuanAccount(account), "official HTTPS API root")
	})
	t.Run("unexpected URL query", func(t *testing.T) {
		account := valid()
		account.Credentials["base_url"] = "https://ai.cangyuansuanli.cn?redirect=elsewhere"
		require.ErrorContains(t, validateCangyuanAccount(account), "official HTTPS API root")
	})
	t.Run("wrong protocol", func(t *testing.T) {
		account := valid()
		account.Platform = PlatformAnthropic
		require.ErrorContains(t, validateCangyuanAccount(account), "OpenAI API-key")
	})
	t.Run("unrelated OpenAI accounts retain existing compatibility", func(t *testing.T) {
		account := valid()
		delete(account.Extra, "upstream_vendor")
		account.Extra["openai_passthrough"] = true
		require.NoError(t, validateCangyuanAccount(account))
		require.True(t, account.SupportsOpenAIImageCapability(OpenAIImagesCapabilityNative))
	})
}

func TestCangyuanAccountCreationStartsUnschedulable(t *testing.T) {
	t.Parallel()
	input := &CreateAccountInput{
		Name:     "Cangyuan staging",
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":       "test-only",
			"base_url":      "https://ai.cangyuansuanli.cn",
			"model_mapping": map[string]any{"safe-chat": "safe-chat"},
		},
	}
	extra := map[string]any{
		"upstream_vendor":       "cangyuan",
		"openai_passthrough":    false,
		"openai_responses_mode": "force_chat_completions",
	}
	account, err := buildAccountForCreate(input, extra)
	require.NoError(t, err)
	require.False(t, account.Schedulable, "an ungrouped Cangyuan account must not receive live traffic")

	input.Platform = PlatformAnthropic
	extra = nil
	account, err = buildAccountForCreate(input, extra)
	require.NoError(t, err)
	require.True(t, account.Schedulable, "existing providers retain their normal defaults")
}
