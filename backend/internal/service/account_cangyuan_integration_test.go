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
	t.Run("official host without vendor marker still fail-closes", func(t *testing.T) {
		account := valid()
		delete(account.Extra, "upstream_vendor")
		require.True(t, isCangyuanAccount(account))
		require.True(t, account.IsModelSupported("safe-model"))
		require.False(t, account.IsModelSupported("unlisted-model"))
		account.Extra["openai_passthrough"] = true
		require.False(t, account.IsOpenAIPassthroughEnabled())
		require.ErrorContains(t, validateCangyuanAccount(account), "unrestricted")
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
		account.Credentials["base_url"] = "https://third-party.example.com"
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

func TestCangyuanMediaOnlyAccountIsolationAndValidation(t *testing.T) {
	t.Parallel()

	media := func() *Account {
		return &Account{
			Platform: PlatformOpenAI,
			Type:     AccountTypeAPIKey,
			Credentials: map[string]any{
				"api_key": "fixture-only",
				"base_url": "https://ai.cangyuansuanli.cn",
				"model_mapping": map[string]any{},
			},
			Extra: map[string]any{
				"upstream_vendor": "cangyuan",
				"cangyuan_media_only": true,
				"cangyuan_media_models": map[string]any{
					"image": []any{"image-model"},
					"video": []any{"video-model"},
				},
				"openai_passthrough": false,
				"openai_responses_mode": "force_chat_completions",
			},
		}
	}

	approved := media()
	require.NoError(t, validateCangyuanAccount(approved))
	require.True(t, isCangyuanMediaOnlyAccount(approved))
	for _, model := range []string{"image-model", "video-model", "arbitrary-chat", "gpt-5.6"} {
		require.False(t, approved.IsModelSupported(model), "media-only accounts must never enter chat scheduling")
	}
	require.False(t, approved.IsOpenAIPassthroughEnabled())
	require.False(t, approved.SupportsOpenAIImageCapability(OpenAIImagesCapabilityBasic))
	require.False(t, approved.SupportsOpenAIImageCapability(OpenAIImagesCapabilityNative))

	input := &CreateAccountInput{
		Name: "Staging media-only",
		Platform: PlatformOpenAI,
		Type: AccountTypeAPIKey,
		Credentials: approved.Credentials,
	}
	created, err := buildAccountForCreate(input, approved.Extra)
	require.NoError(t, err)
	require.False(t, created.Schedulable, "unpriced media-only accounts cannot receive live traffic")
	require.False(t, created.IsModelSupported("image-model"))

	t.Run("no approved media", func(t *testing.T) {
		a := media()
		a.Extra["cangyuan_media_models"] = map[string]any{"image": []any{}, "video": []any{}}
		require.ErrorContains(t, validateCangyuanAccount(a), "at least one model")
	})
	t.Run("media safety survives vendor flag removal", func(t *testing.T) {
		a := media()
		delete(a.Extra, "upstream_vendor")
		require.True(t, isCangyuanAccount(a))
		require.True(t, isCangyuanMediaOnlyAccount(a))
		require.NoError(t, validateCangyuanAccount(a))
		require.False(t, a.IsModelSupported("image-model"))
	})
	t.Run("media wildcard", func(t *testing.T) {
		a := media()
		a.Extra["cangyuan_media_models"] = map[string]any{"image": []any{"image-*"}}
		require.ErrorContains(t, validateCangyuanAccount(a), "non-wildcard")
	})
	t.Run("mixed chat and media", func(t *testing.T) {
		a := media()
		a.Credentials["model_mapping"] = map[string]any{"chat": "chat"}
		require.ErrorContains(t, validateCangyuanAccount(a), "cannot advertise chat models")
		require.False(t, a.IsModelSupported("chat"))
	})
	t.Run("unsupported endpoint", func(t *testing.T) {
		a := media()
		a.Extra["cangyuan_media_models"] = map[string]any{"music": []any{"music-model"}}
		require.ErrorContains(t, validateCangyuanAccount(a), "unsupported capability")
	})
	t.Run("invalid media-only flag type", func(t *testing.T) {
		a := media()
		a.Extra["cangyuan_media_only"] = "true"
		require.ErrorContains(t, validateCangyuanAccount(a), "boolean")
	})
	t.Run("missing metadata", func(t *testing.T) {
		a := media()
		delete(a.Extra, "cangyuan_media_models")
		require.ErrorContains(t, validateCangyuanAccount(a), "explicit image/video")
	})
	t.Run("ordinary chat cannot carry undeclared media", func(t *testing.T) {
		a := media()
		delete(a.Extra, "cangyuan_media_only")
		a.Credentials["model_mapping"] = map[string]any{"chat": "chat"}
		require.ErrorContains(t, validateCangyuanAccount(a), "cannot mix")
	})
}
