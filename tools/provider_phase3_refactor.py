from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[1]
SERVICE = ROOT / "backend" / "internal" / "service"


def replace_once(text: str, old: str, new: str, label: str) -> str:
    count = text.count(old)
    if count != 1:
        raise RuntimeError(f"{label}: expected exactly one match, got {count}")
    return text.replace(old, new, 1)


def regex_once(text: str, pattern: str, replacement: str, label: str) -> str:
    out, count = re.subn(pattern, replacement, text, count=1, flags=re.S)
    if count != 1:
        raise RuntimeError(f"{label}: expected exactly one regex match, got {count}")
    return out


# 1. Make provider and protocol orthogonal in the shared request envelope.
p = SERVICE / "provider_adapter.go"
text = p.read_text()
old = '''// ProviderRequestInput contains only provider/request material. Scheduling,
// failover orchestration, response commit and billing remain gateway concerns.
type ProviderRequestInput struct {
\tAccount       *Account
'''
new = '''// ProviderProtocol is the client/wire protocol axis. Provider identity and
// protocol are intentionally orthogonal: one provider may serve several client
// protocols, and one protocol may be implemented by several providers.
type ProviderProtocol string

const (
\tProviderProtocolAnthropic       ProviderProtocol = "anthropic"
\tProviderProtocolChatCompletions ProviderProtocol = "chat_completions"
\tProviderProtocolResponses       ProviderProtocol = "responses"
\tProviderProtocolGemini          ProviderProtocol = "gemini"
)

// ProviderRequestInput contains only provider/request material. Scheduling,
// failover orchestration, response commit and billing remain gateway concerns.
type ProviderRequestInput struct {
\tAccount       *Account
\tProtocol      ProviderProtocol
'''
text = replace_once(text, old, new, "provider protocol envelope")
p.write_text(text)


# 2. Add the real Gemini request/auth adapter. Gateway lifecycle stays outside.
provider_file = SERVICE / "gemini_provider_request.go"
provider_file.write_text(r'''package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
)

const geminiProviderRequestIDHeader = "x-request-id"

type geminiGatewayProviderAdapter struct {
	service *GeminiMessagesCompatService
}

func newGeminiGatewayProviderAdapter(service *GeminiMessagesCompatService) ProviderAdapter {
	return &geminiGatewayProviderAdapter{service: service}
}

func (a *geminiGatewayProviderAdapter) Name() string { return PlatformGemini }

func (a *geminiGatewayProviderAdapter) Supports(account *Account) bool {
	return account != nil && strings.EqualFold(strings.TrimSpace(account.Platform), PlatformGemini)
}

func (a *geminiGatewayProviderAdapter) ResolveModel(account *Account, requestedModel string) (string, error) {
	if account == nil {
		return "", errors.New("account is required")
	}
	if !a.Supports(account) {
		return "", errors.New("account is not supported by gemini provider adapter")
	}
	if strings.TrimSpace(requestedModel) == "" {
		return "", errors.New("gemini model is required")
	}

	switch account.Type {
	case AccountTypeAPIKey, AccountTypeServiceAccount:
		return account.GetMappedModel(requestedModel), nil
	case AccountTypeOAuth:
		// Preserve existing Gemini OAuth behavior: Code Assist / AI Studio OAuth
		// historically forwards the requested model without account mapping.
		return requestedModel, nil
	default:
		return "", fmt.Errorf("unsupported account type: %s", account.Type)
	}
}

func (a *geminiGatewayProviderAdapter) NormalizeError(resp *http.Response, body []byte) NormalizedProviderError {
	return normalizeGeminiProviderError(resp, body)
}

func (a *geminiGatewayProviderAdapter) BuildRequest(ctx context.Context, input ProviderRequestInput) (*http.Request, error) {
	if a == nil || a.service == nil {
		return nil, errors.New("gemini provider adapter is not configured")
	}
	account := input.Account
	if account == nil {
		return nil, errors.New("gemini account is required")
	}
	if !a.Supports(account) {
		return nil, errors.New("account is not supported by gemini provider adapter")
	}
	if strings.TrimSpace(input.Model) == "" {
		return nil, errors.New("gemini model is required")
	}

	switch input.Protocol {
	case ProviderProtocolAnthropic, ProviderProtocolChatCompletions, ProviderProtocolResponses, ProviderProtocolGemini:
		// Supported wire-protocol contexts. The request body has already been
		// converted to Gemini native shape by the compatibility layer.
	default:
		return nil, fmt.Errorf("unsupported gemini provider protocol: %s", input.Protocol)
	}

	action := strings.TrimSpace(input.Endpoint)
	if action == "" {
		action = "generateContent"
		if input.Stream {
			action = "streamGenerateContent"
		}
	}
	if _, ok := geminiAIStudioActions[action]; !ok {
		return nil, fmt.Errorf("unsupported gemini action: %s", action)
	}

	method := strings.TrimSpace(input.Method)
	if method == "" {
		method = http.MethodPost
	}
	if method != http.MethodPost {
		return nil, fmt.Errorf("unsupported gemini request method: %s", method)
	}

	bodyForREST := input.Body
	if input.Protocol != ProviderProtocolGemini {
		bodyForREST = normalizeGeminiRequestForAIStudio(input.Body)
	}

	forceAIStudio := input.Protocol == ProviderProtocolGemini && action == "countTokens"

	switch account.Type {
	case AccountTypeAPIKey:
		baseURL := account.GetGeminiBaseURL(geminicli.AIStudioBaseURL)
		normalizedBaseURL, err := a.service.validateUpstreamBaseURL(baseURL)
		if err != nil {
			return nil, err
		}
		fullURL, err := buildGeminiAIStudioModelActionURL(normalizedBaseURL, input.Model, action, input.Stream)
		if err != nil {
			return nil, err
		}
		return http.NewRequestWithContext(ctx, method, fullURL, bytes.NewReader(bodyForREST))

	case AccountTypeOAuth:
		projectID := strings.TrimSpace(account.GetCredential("project_id"))
		if projectID != "" && !forceAIStudio {
			baseURL, err := a.service.validateUpstreamBaseURL(geminicli.GeminiCliBaseURL)
			if err != nil {
				return nil, err
			}
			fullURL := fmt.Sprintf("%s/v1internal:%s", strings.TrimRight(baseURL, "/"), action)
			if input.Stream {
				fullURL += "?alt=sse"
			}

			var inner any
			if err := json.Unmarshal(input.Body, &inner); err != nil {
				return nil, fmt.Errorf("failed to parse gemini request: %w", err)
			}
			wrappedBytes, err := json.Marshal(map[string]any{
				"model":   input.Model,
				"project": projectID,
				"request": inner,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to wrap gemini request: %w", err)
			}
			return http.NewRequestWithContext(ctx, method, fullURL, bytes.NewReader(wrappedBytes))
		}

		baseURL := account.GetGeminiBaseURL(geminicli.AIStudioBaseURL)
		normalizedBaseURL, err := a.service.validateUpstreamBaseURL(baseURL)
		if err != nil {
			return nil, err
		}
		fullURL, err := buildGeminiAIStudioModelActionURL(normalizedBaseURL, input.Model, action, input.Stream)
		if err != nil {
			return nil, err
		}
		return http.NewRequestWithContext(ctx, method, fullURL, bytes.NewReader(bodyForREST))

	case AccountTypeServiceAccount:
		fullURL, err := buildVertexGeminiURL(
			account.VertexProjectID(),
			account.VertexLocation(input.Model),
			input.Model,
			action,
			input.Stream,
		)
		if err != nil {
			return nil, err
		}
		return http.NewRequestWithContext(ctx, method, fullURL, bytes.NewReader(bodyForREST))

	default:
		return nil, fmt.Errorf("unsupported account type: %s", account.Type)
	}
}

func (a *geminiGatewayProviderAdapter) ApplyAuth(ctx context.Context, req *http.Request, input ProviderRequestInput) error {
	if a == nil || a.service == nil {
		return errors.New("gemini provider adapter is not configured")
	}
	if req == nil {
		return errors.New("gemini upstream request is required")
	}
	account := input.Account
	if account == nil {
		return errors.New("gemini account is required")
	}

	req.Header.Set("Content-Type", "application/json")
	switch account.Type {
	case AccountTypeAPIKey:
		apiKey := account.GetCredential("api_key")
		if strings.TrimSpace(apiKey) == "" {
			return errors.New("gemini api_key not configured")
		}
		req.Header.Set("x-goog-api-key", apiKey)
		return nil

	case AccountTypeOAuth, AccountTypeServiceAccount:
		if a.service.tokenProvider == nil {
			return errors.New("gemini token provider not configured")
		}
		accessToken, err := a.service.tokenProvider.GetAccessToken(ctx, account)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)

		forceAIStudio := input.Protocol == ProviderProtocolGemini && strings.TrimSpace(input.Endpoint) == "countTokens"
		if account.Type == AccountTypeOAuth && strings.TrimSpace(account.GetCredential("project_id")) != "" && !forceAIStudio {
			req.Header.Set("User-Agent", geminicli.GeminiCLIUserAgent)
		}
		return nil

	default:
		return fmt.Errorf("unsupported account type: %s", account.Type)
	}
}

func (s *GeminiMessagesCompatService) geminiProviderAdapterRegistry() *ProviderAdapterRegistry {
	if s != nil && s.providerAdapters != nil {
		return s.providerAdapters
	}
	return MustNewProviderAdapterRegistry(newGeminiGatewayProviderAdapter(s))
}

func (s *GeminiMessagesCompatService) resolveGeminiProviderModel(account *Account, requestedModel string) (string, error) {
	adapter, err := s.geminiProviderAdapterRegistry().Resolve(account)
	if err != nil {
		return "", err
	}
	return adapter.ResolveModel(account, requestedModel)
}

func (s *GeminiMessagesCompatService) newGeminiProviderRequestFactory(input ProviderRequestInput) (func(context.Context) (*http.Request, string, error), string, error) {
	registry := s.geminiProviderAdapterRegistry()
	adapter, err := registry.Resolve(input.Account)
	if err != nil {
		return nil, "", err
	}
	builder, ok := adapter.(ProviderRequestBuilder)
	if !ok {
		return nil, "", fmt.Errorf("provider %s does not implement request builder", adapter.Name())
	}
	auth, ok := adapter.(ProviderAuthApplier)
	if !ok {
		return nil, "", fmt.Errorf("provider %s does not implement auth applier", adapter.Name())
	}

	return func(ctx context.Context) (*http.Request, string, error) {
		req, err := builder.BuildRequest(ctx, input)
		if err != nil {
			return nil, "", err
		}
		if err := auth.ApplyAuth(ctx, req, input); err != nil {
			return nil, "", err
		}
		return req, geminiProviderRequestIDHeader, nil
	}, geminiProviderRequestIDHeader, nil
}
''')


# 3. Route Gemini Messages/native request construction through the adapter.
p = SERVICE / "gemini_messages_compat_service.go"
text = p.read_text()
text = replace_once(
    text,
    '\tantigravityGatewayService *AntigravityGatewayService\n\tcfg                       *config.Config\n',
    '\tantigravityGatewayService *AntigravityGatewayService\n\tproviderAdapters          *ProviderAdapterRegistry\n\tcfg                       *config.Config\n',
    "gemini service registry field",
)
old_ctor = '''\treturn &GeminiMessagesCompatService{
\t\taccountRepo:               accountRepo,
\t\tgroupRepo:                 groupRepo,
\t\tcache:                     cache,
\t\tschedulerSnapshot:         schedulerSnapshot,
\t\ttokenProvider:             tokenProvider,
\t\trateLimitService:          rateLimitService,
\t\thttpUpstream:              httpUpstream,
\t\tantigravityGatewayService: antigravityGatewayService,
\t\tcfg:                       cfg,
\t\tresponseHeaderFilter:      compileResponseHeaderFilter(cfg),
\t}
'''
new_ctor = '''\tsvc := &GeminiMessagesCompatService{
\t\taccountRepo:               accountRepo,
\t\tgroupRepo:                 groupRepo,
\t\tcache:                     cache,
\t\tschedulerSnapshot:         schedulerSnapshot,
\t\ttokenProvider:             tokenProvider,
\t\trateLimitService:          rateLimitService,
\t\thttpUpstream:              httpUpstream,
\t\tantigravityGatewayService: antigravityGatewayService,
\t\tcfg:                       cfg,
\t\tresponseHeaderFilter:      compileResponseHeaderFilter(cfg),
\t}
\tsvc.providerAdapters = MustNewProviderAdapterRegistry(newGeminiGatewayProviderAdapter(svc))
\treturn svc
'''
text = replace_once(text, old_ctor, new_ctor, "gemini constructor registry")
text = replace_once(
    text,
    '''\toriginalModel := req.Model
\tmappedModel := req.Model
\tif account.Type == AccountTypeAPIKey || account.Type == AccountTypeServiceAccount {
\t\tmappedModel = account.GetMappedModel(req.Model)
\t}
''',
    '''\toriginalModel := req.Model
\tmappedModel, err := s.resolveGeminiProviderModel(account, req.Model)
\tif err != nil {
\t\treturn nil, err
\t}
''',
    "messages model resolution",
)
forward_replacement = '''\tuseUpstreamStream := req.Stream
\tif account.Type == AccountTypeOAuth && !req.Stream && strings.TrimSpace(account.GetCredential("project_id")) != "" {
\t\t// Code Assist's non-streaming generateContent may return no content; use streaming upstream and aggregate.
\t\tuseUpstreamStream = true
\t}

\taction := "generateContent"
\tif useUpstreamStream {
\t\taction = "streamGenerateContent"
\t}
\tbuildReq, requestIDHeader, buildFactoryErr := s.newGeminiProviderRequestFactory(ProviderRequestInput{
\t\tAccount:  account,
\t\tProtocol: ProviderProtocolAnthropic,
\t\tEndpoint: action,
\t\tModel:    mappedModel,
\t\tBody:     geminiReq,
\t\tStream:   useUpstreamStream,
\t})
\tif buildFactoryErr != nil {
\t\treturn nil, buildFactoryErr
\t}
'''
text = regex_once(
    text,
    r'\tvar requestIDHeader string\n\tvar buildReq func\(ctx context\.Context\) \(\*http\.Request, string, error\)\n\tuseUpstreamStream := req\.Stream\n.*?\n\tdefault:\n\t\treturn nil, fmt\.Errorf\("unsupported account type: %s", account\.Type\)\n\t}\n',
    forward_replacement,
    "messages request switch",
)
text = replace_once(
    text,
    '''\tmappedModel := originalModel
\tif account.Type == AccountTypeAPIKey || account.Type == AccountTypeServiceAccount {
\t\tmappedModel = account.GetMappedModel(originalModel)
\t}
''',
    '''\tmappedModel, err := s.resolveGeminiProviderModel(account, originalModel)
\tif err != nil {
\t\treturn nil, s.writeGoogleError(c, http.StatusBadGateway, err.Error())
\t}
''',
    "native model resolution",
)
native_replacement = '''\tbuildReq, requestIDHeader, buildFactoryErr := s.newGeminiProviderRequestFactory(ProviderRequestInput{
\t\tAccount:  account,
\t\tProtocol: ProviderProtocolGemini,
\t\tEndpoint: upstreamAction,
\t\tModel:    mappedModel,
\t\tBody:     body,
\t\tStream:   useUpstreamStream,
\t})
\tif buildFactoryErr != nil {
\t\treturn nil, s.writeGoogleError(c, http.StatusBadGateway, buildFactoryErr.Error())
\t}
'''
text = regex_once(
    text,
    r'\tforceAIStudio := action == "countTokens"\n\n\tvar requestIDHeader string\n\tvar buildReq func\(ctx context\.Context\) \(\*http\.Request, string, error\)\n\n\tswitch account\.Type \{.*?\n\tdefault:\n\t\treturn nil, s\.writeGoogleError\(c, http\.StatusBadGateway, "Unsupported account type: "\+account\.Type\)\n\t}\n',
    native_replacement,
    "native request switch",
)
p.write_text(text)


# 4. Route Chat Completions -> Gemini through the same adapter.
p = SERVICE / "gemini_chat_completions_compat_service.go"
text = p.read_text()
text = replace_once(text, '\t"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"\n', '', "remove obsolete chat geminicli import")
text = replace_once(
    text,
    '''\tmappedModel := req.Model
\tif account.Type == AccountTypeAPIKey || account.Type == AccountTypeServiceAccount {
\t\tmappedModel = account.GetMappedModel(req.Model)
\t}
''',
    '''\tmappedModel, err := s.resolveGeminiProviderModel(account, req.Model)
\tif err != nil {
\t\treturn nil, s.writeChatCompletionsError(c, http.StatusBadGateway, "upstream_error", err.Error())
\t}
''',
    "chat model resolution",
)
text = replace_once(
    text,
    '''\t\tgeminiReq,
\t\tclientStream,
\t\tuseUpstreamStream,
''',
    '''\t\tgeminiReq,
\t\tuseUpstreamStream,
''',
    "chat builder args",
)
chat_builder = r'''func (s *GeminiMessagesCompatService) buildGeminiChatCompletionsUpstreamRequestFunc(
	account *Account,
	mappedModel string,
	geminiReq []byte,
	useUpstreamStream bool,
) (func(context.Context) (*http.Request, string, error), string) {
	action := "generateContent"
	if useUpstreamStream {
		action = "streamGenerateContent"
	}
	buildReq, requestIDHeader, err := s.newGeminiProviderRequestFactory(ProviderRequestInput{
		Account:  account,
		Protocol: ProviderProtocolChatCompletions,
		Endpoint: action,
		Model:    mappedModel,
		Body:     geminiReq,
		Stream:   useUpstreamStream,
	})
	if err != nil {
		return func(context.Context) (*http.Request, string, error) {
			return nil, "", err
		}, geminiProviderRequestIDHeader
	}
	return buildReq, requestIDHeader
}

'''
text = regex_once(
    text,
    r'func \(s \*GeminiMessagesCompatService\) buildGeminiChatCompletionsUpstreamRequestFunc\(.*?\n}\n\n(?=func \(s \*GeminiMessagesCompatService\) handleChatCompletionsNonStreamingResponseFromGemini)',
    chat_builder,
    "chat request builder",
)
p.write_text(text)


# 5. Keep older direct-construction fixtures explicit about their provider.
p = SERVICE / "gemini_messages_compat_service_test.go"
text = p.read_text()
for test_name in (
    "TestGeminiMessagesCompatServiceForward_PreservesRequestedModelAndMappedUpstreamModel",
    "TestGeminiMessagesCompatServiceForward_NormalizesWebSearchToolForAIStudio",
):
    pattern = rf'(func {re.escape(test_name)}\(t \*testing\.T\) \{{.*?account := &Account\{{\n\t\tID:\s+1,\n)(\t\tType:\s+AccountTypeAPIKey,)'
    replacement = r'\1\t\tPlatform: PlatformGemini,\n\2'
    text = regex_once(text, pattern, replacement, f"fixture platform: {test_name}")
p.write_text(text)


# 6. Focused adapter tests for routing/auth/body compatibility.
test_file = SERVICE / "gemini_provider_request_test.go"
test_file.write_text(r'''package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
	"github.com/stretchr/testify/require"
)

func newGeminiAdapterTestService() *GeminiMessagesCompatService {
	return &GeminiMessagesCompatService{
		tokenProvider: &GeminiTokenProvider{},
		cfg:           &config.Config{},
	}
}

func TestGeminiGatewayProviderAdapterAPIKeyBuildsAIStudioRequestAndAuth(t *testing.T) {
	svc := newGeminiAdapterTestService()
	adapter := newGeminiGatewayProviderAdapter(svc)
	builder := adapter.(ProviderRequestBuilder)
	auth := adapter.(ProviderAuthApplier)
	account := &Account{
		Platform: PlatformGemini,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key": "secret-key",
		},
	}
	input := ProviderRequestInput{
		Account:  account,
		Protocol: ProviderProtocolChatCompletions,
		Endpoint: "streamGenerateContent",
		Model:    "gemini-2.5-flash",
		Body:     []byte(`{"tools":[{"googleSearch":{}}]}`),
		Stream:   true,
	}

	req, err := builder.BuildRequest(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, req.Method)
	require.Contains(t, req.URL.String(), "/v1beta/models/gemini-2.5-flash:streamGenerateContent?alt=sse")

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(body, &payload))
	tools := payload["tools"].([]any)
	tool := tools[0].(map[string]any)
	_, hasSnake := tool["google_search"]
	_, hasCamel := tool["googleSearch"]
	require.True(t, hasSnake)
	require.False(t, hasCamel)

	require.NoError(t, auth.ApplyAuth(context.Background(), req, input))
	require.Equal(t, "secret-key", req.Header.Get("x-goog-api-key"))
	require.Empty(t, req.Header.Get("Authorization"))
}

func TestGeminiGatewayProviderAdapterOAuthCodeAssistWrapsAndAuthenticates(t *testing.T) {
	svc := newGeminiAdapterTestService()
	adapter := newGeminiGatewayProviderAdapter(svc)
	builder := adapter.(ProviderRequestBuilder)
	auth := adapter.(ProviderAuthApplier)
	account := &Account{
		Platform: PlatformGemini,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "ya29.adapter-test",
			"project_id":   "project-1",
		},
	}
	input := ProviderRequestInput{
		Account:  account,
		Protocol: ProviderProtocolAnthropic,
		Endpoint: "streamGenerateContent",
		Model:    "gemini-2.5-pro",
		Body:     []byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`),
		Stream:   true,
	}

	req, err := builder.BuildRequest(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, "https://cloudcode-pa.googleapis.com/v1internal:streamGenerateContent?alt=sse", req.URL.String())

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	var wrapped map[string]any
	require.NoError(t, json.Unmarshal(body, &wrapped))
	require.Equal(t, "gemini-2.5-pro", wrapped["model"])
	require.Equal(t, "project-1", wrapped["project"])
	require.NotNil(t, wrapped["request"])

	require.NoError(t, auth.ApplyAuth(context.Background(), req, input))
	require.Equal(t, "Bearer ya29.adapter-test", req.Header.Get("Authorization"))
	require.Equal(t, geminicli.GeminiCLIUserAgent, req.Header.Get("User-Agent"))
}

func TestGeminiGatewayProviderAdapterCountTokensForcesAIStudioForOAuth(t *testing.T) {
	svc := newGeminiAdapterTestService()
	adapter := newGeminiGatewayProviderAdapter(svc)
	builder := adapter.(ProviderRequestBuilder)
	auth := adapter.(ProviderAuthApplier)
	account := &Account{
		Platform: PlatformGemini,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "ya29.adapter-test",
			"project_id":   "project-1",
		},
	}
	input := ProviderRequestInput{
		Account:  account,
		Protocol: ProviderProtocolGemini,
		Endpoint: "countTokens",
		Model:    "gemini-2.5-flash",
		Body:     []byte(`{"contents":[{"parts":[{"text":"hi"}]}]}`),
	}

	req, err := builder.BuildRequest(context.Background(), input)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(req.URL.String(), "https://generativelanguage.googleapis.com/v1beta/models/"))
	require.Contains(t, req.URL.String(), ":countTokens")

	require.NoError(t, auth.ApplyAuth(context.Background(), req, input))
	require.Equal(t, "Bearer ya29.adapter-test", req.Header.Get("Authorization"))
	require.Empty(t, req.Header.Get("User-Agent"))
}

func TestGeminiGatewayProviderAdapterPreservesOAuthModelMappingSemantics(t *testing.T) {
	svc := newGeminiAdapterTestService()
	adapter := newGeminiGatewayProviderAdapter(svc)

	apiKey := &Account{
		Platform: PlatformGemini,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"model_mapping": map[string]any{"requested": "mapped"},
		},
	}
	mapped, err := adapter.ResolveModel(apiKey, "requested")
	require.NoError(t, err)
	require.Equal(t, "mapped", mapped)

	oauth := &Account{
		Platform: PlatformGemini,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"model_mapping": map[string]any{"requested": "mapped"},
		},
	}
	mapped, err = adapter.ResolveModel(oauth, "requested")
	require.NoError(t, err)
	require.Equal(t, "requested", mapped)
}

func TestGeminiGatewayProviderAdapterExposesRequestAndAuthCapabilities(t *testing.T) {
	svc := NewGeminiMessagesCompatService(nil, nil, nil, nil, &GeminiTokenProvider{}, nil, nil, nil, &config.Config{})
	caps, err := svc.providerAdapters.Capabilities(PlatformGemini)
	require.NoError(t, err)
	require.Contains(t, caps, ProviderCapabilityRequestBuilder)
	require.Contains(t, caps, ProviderCapabilityAuthApplier)
}
''')

print("provider adapter phase3 refactor applied")
