package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/require"
)

// The production registry the gateway builds must contain the CQU adapter and
// resolve a CQU account to it.
func TestOpenAIGatewayServiceRegistersCQUBrowserAdapter(t *testing.T) {
	svc := &OpenAIGatewayService{cfg: &config.Config{CQU: config.CQUConfig{
		Enabled:         true,
		BrowserDebugURL: config.DefaultCQUBrowserDebugURL,
		BaseURL:         config.DefaultCQUBaseURL,
		DefaultAgentID:  config.DefaultCQUAgentID,
		DefaultModelID:  config.DefaultCQUModelID,
	}}}

	registry := svc.cquProviderAdapterRegistry()
	require.NotNil(t, registry)
	require.Contains(t, registry.Names(), cquBrowserAdapterName)
	// The registry is built once and reused across requests.
	require.Same(t, registry, svc.cquProviderAdapterRegistry())

	adapter, err := registry.Resolve(&Account{Platform: "cqu", Type: "browser"})
	require.NoError(t, err)
	require.Equal(t, cquBrowserAdapterName, adapter.Name())

	_, isSender := adapter.(ProviderRequestSender)
	require.True(t, isSender, "CQU must supply its own transport rather than the shared HTTP upstream")
}

// A disabled bridge must fail closed with an actionable operator error instead
// of attempting any direct HTTP call to ai.cqu.edu.cn.
func TestCQUAdapterFailsClosedWhenBridgeDisabled(t *testing.T) {
	adapter := NewCQUBrowserProviderAdapter(CQUBrowserAdapterDeps{
		Config: &config.Config{CQU: config.CQUConfig{Enabled: false}},
	})

	_, err := cquSendThroughAdapter(context.Background(), adapter, ProviderRequestInput{
		Account: &Account{Platform: "cqu", Type: "browser"},
		Model:   CQUDefaultModelAlias,
		Body:    []byte(`{"model":"cqu-default","messages":[{"role":"user","content":"你好"}],"stream":true}`),
	})
	require.Error(t, err)
	require.Equal(t, CQUErrorKindDisabled, CQUBridgeErrorKind(err))

	status, errType, message := cquBridgeClientError(CQUBridgeErrorKind(err), err)
	require.Equal(t, 503, status)
	require.Equal(t, "upstream_error", errType)
	require.Contains(t, message, "CQU_ENABLED")
}

// /v1/models must advertise cqu-default for a CQU account without breaking the
// model list of any other account in the same group.
func TestGetAvailableModelsAdvertisesCQUDefault(t *testing.T) {
	groupID := int64(41)
	repo := &modelsListAccountRepoStub{
		byGroup: map[int64][]Account{
			groupID: {
				{
					ID:       1,
					Platform: PlatformOpenAI,
					Type:     AccountTypeAPIKey,
					Credentials: map[string]any{
						"model_mapping": map[string]any{"gpt-5": "gpt-5"},
					},
				},
				{
					ID:       2,
					Platform: PlatformOpenAI,
					Type:     AccountTypeAPIKey,
					Extra:    map[string]any{"cqu_browser": true},
				},
			},
		},
	}
	svc := &GatewayService{
		accountRepo:        repo,
		modelsListCache:    gocache.New(time.Minute, time.Minute),
		modelsListCacheTTL: time.Minute,
	}

	models := svc.GetAvailableModels(context.Background(), &groupID, PlatformOpenAI)
	require.Equal(t, []string{CQUDefaultModelAlias, "gpt-5"}, models)
}

// A group with no CQU account must keep exactly the model list it had before.
func TestGetAvailableModelsUnchangedWithoutCQUAccount(t *testing.T) {
	groupID := int64(42)
	repo := &modelsListAccountRepoStub{
		byGroup: map[int64][]Account{
			groupID: {
				{
					ID:       1,
					Platform: PlatformAnthropic,
					Credentials: map[string]any{
						"model_mapping": map[string]any{"claude-3-5-haiku": "claude-3-5-haiku"},
					},
				},
			},
		},
	}
	svc := &GatewayService{
		accountRepo:        repo,
		modelsListCache:    gocache.New(time.Minute, time.Minute),
		modelsListCacheTTL: time.Minute,
	}

	models := svc.GetAvailableModels(context.Background(), &groupID, PlatformAnthropic)
	require.Equal(t, []string{"claude-3-5-haiku"}, models)
	require.NotContains(t, models, CQUDefaultModelAlias)
}
