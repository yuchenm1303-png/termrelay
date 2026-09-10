package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

// fakeCQUSender stands in for the browser bridge. It records what the adapter
// asked the page to send and replays a canned CQU event stream.
type fakeCQUSender struct {
	requests []CQUChatRequest
	stream   string
	status   int
	body     string
	err      error
}

func (f *fakeCQUSender) SendChat(ctx context.Context, request CQUChatRequest) (*http.Response, error) {
	f.requests = append(f.requests, request)
	if f.err != nil {
		return nil, f.err
	}
	status := f.status
	if status == 0 {
		status = http.StatusOK
	}
	body := f.stream
	if f.body != "" {
		body = f.body
	}
	header := make(http.Header)
	header.Set("Content-Type", "text/event-stream")
	return &http.Response{
		StatusCode: status,
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}

func newTestCQUAdapter(sender CQUBrowserSender) *CQUBrowserProviderAdapter {
	return NewCQUBrowserProviderAdapter(CQUBrowserAdapterDeps{
		Config: &config.Config{CQU: config.CQUConfig{
			Enabled:         true,
			BaseURL:         config.DefaultCQUBaseURL,
			DefaultAgentID:  config.DefaultCQUAgentID,
			DefaultModelID:  config.DefaultCQUModelID,
			BrowserDebugURL: config.DefaultCQUBrowserDebugURL,
		}},
		Bridge: func(context.Context) (CQUBrowserSender, error) { return sender, nil },
	})
}

func cquTestAccount() *Account {
	return &Account{ID: 7, Platform: "cqu", Type: "browser"}
}

func cquTestStream(deltas ...string) string {
	frames := []string{
		"event: RUN_STARTED\ndata: {\"threadId\":\"thread-1\",\"runId\":\"run-1\"}\n\n",
		"event: TEXT_MESSAGE_START\ndata: {\"messageId\":\"msg-1\",\"role\":\"assistant\"}\n\n",
	}
	for _, delta := range deltas {
		encoded, _ := json.Marshal(map[string]any{"messageId": "msg-1", "delta": delta})
		frames = append(frames, "event: TEXT_MESSAGE_CONTENT\ndata: "+string(encoded)+"\n\n")
	}
	frames = append(frames,
		"event: TEXT_MESSAGE_END\ndata: {\"messageId\":\"msg-1\"}\n\n",
		"event: RUN_FINISHED\ndata: {\"threadId\":\"thread-1\",\"runId\":\"run-1\"}\n\n",
	)
	return strings.Join(frames, "")
}

func TestCQUBrowserProviderAdapterSupports(t *testing.T) {
	adapter := newTestCQUAdapter(&fakeCQUSender{})

	tests := []struct {
		name    string
		account *Account
		want    bool
	}{
		{name: "nil", account: nil, want: false},
		{name: "canonical platform", account: &Account{Platform: "cqu_browser"}, want: true},
		{name: "cqu browser", account: &Account{Platform: "cqu", Type: "browser"}, want: true},
		{name: "cqu browser canonical type", account: &Account{Platform: "CQU", Type: "CQU_BROWSER"}, want: true},
		{name: "cqu default type", account: &Account{Platform: "cqu"}, want: true},
		{name: "cqu incompatible type", account: &Account{Platform: "cqu", Type: "api_key"}, want: false},
		{name: "plain openai", account: &Account{Platform: "openai", Type: "apikey"}, want: false},
		{name: "plain anthropic", account: &Account{Platform: "anthropic", Type: "oauth"}, want: false},
		{name: "plain gemini", account: &Account{Platform: "gemini", Type: "oauth"}, want: false},
		{
			name:    "openai account marked as cqu browser",
			account: &Account{Platform: "openai", Type: "apikey", Extra: map[string]any{"cqu_browser": true}},
			want:    true,
		},
		{
			name:    "openai account with provider marker",
			account: &Account{Platform: "openai", Type: "apikey", Extra: map[string]any{"provider": "cqu_browser"}},
			want:    true,
		},
		{
			name:    "openai account with falsy marker stays openai",
			account: &Account{Platform: "openai", Type: "apikey", Extra: map[string]any{"cqu_browser": false}},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, adapter.Supports(tt.account))
		})
	}
}

// A registry holding both the CQU adapter and the existing platform adapters
// must keep resolving every non-CQU account exactly as before.
func TestCQUBrowserProviderAdapterRegistryResolve(t *testing.T) {
	openai := &builtinProviderAdapter{name: PlatformOpenAI, platform: PlatformOpenAI}
	anthropic := &builtinProviderAdapter{name: PlatformAnthropic, platform: PlatformAnthropic}
	registry := MustNewProviderAdapterRegistry(openai, anthropic, newTestCQUAdapter(&fakeCQUSender{}))

	resolved, err := registry.Resolve(cquTestAccount())
	require.NoError(t, err)
	require.Equal(t, cquBrowserAdapterName, resolved.Name())

	resolved, err = registry.Resolve(&Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey})
	require.NoError(t, err)
	require.Equal(t, PlatformOpenAI, resolved.Name())

	resolved, err = registry.Resolve(&Account{Platform: PlatformAnthropic, Type: AccountTypeOAuth})
	require.NoError(t, err)
	require.Equal(t, PlatformAnthropic, resolved.Name())

	// The marked OpenAI account is the only one that changes hands, and only
	// because the operator opted in explicitly.
	resolved, err = registry.Resolve(&Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Extra:    map[string]any{"cqu_browser": true},
	})
	require.NoError(t, err)
	require.Equal(t, cquBrowserAdapterName, resolved.Name())
}

func TestCQUBrowserProviderAdapterCapabilities(t *testing.T) {
	registry := MustNewProviderAdapterRegistry(newTestCQUAdapter(&fakeCQUSender{}))
	capabilities, err := registry.Capabilities(cquBrowserAdapterName)
	require.NoError(t, err)
	require.Contains(t, capabilities, ProviderCapabilityRequestSender)
	require.Contains(t, capabilities, ProviderCapabilityRequestBuilder)
	require.Contains(t, capabilities, ProviderCapabilityStreamingParser)
	require.Contains(t, capabilities, ProviderCapabilityModelLister)
}

func TestCQUBrowserProviderAdapterResolvesDefaultModelAlias(t *testing.T) {
	adapter := newTestCQUAdapter(&fakeCQUSender{})
	account := cquTestAccount()

	model, err := adapter.ResolveModel(account, CQUDefaultModelAlias)
	require.NoError(t, err)
	require.Equal(t, CQUDefaultModelAlias, model)

	model, err = adapter.ResolveModel(account, "")
	require.NoError(t, err)
	require.Equal(t, CQUDefaultModelAlias, model)

	models, err := adapter.ListModels(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, []string{CQUDefaultModelAlias}, models)
}

// cqu-default must resolve to the configured agent/model pair and must never be
// sent upstream as a CQU modelId.
func TestCQUDefaultModelAliasMapsToAgentAndModelIDs(t *testing.T) {
	cfg := &config.Config{CQU: config.CQUConfig{
		DefaultAgentID: config.DefaultCQUAgentID,
		DefaultModelID: config.DefaultCQUModelID,
	}}

	routing := resolveCQURouting(cquTestAccount(), cfg, CQUDefaultModelAlias)
	require.Equal(t, "910020080887140352", routing.AgentID)
	require.Equal(t, "910023701267746816", routing.ModelID)

	// Account-level overrides win over the configured defaults.
	account := cquTestAccount()
	account.Credentials = map[string]any{"agent_id": "agent-override", "model_id": "model-override"}
	routing = resolveCQURouting(account, cfg, CQUDefaultModelAlias)
	require.Equal(t, "agent-override", routing.AgentID)
	require.Equal(t, "model-override", routing.ModelID)

	// A raw numeric CQU model id from the client is honored as-is.
	routing = resolveCQURouting(cquTestAccount(), cfg, "910023701267746817")
	require.Equal(t, "910023701267746817", routing.ModelID)
}

func TestBuildCQUChatRequestUsesQueryNotMessages(t *testing.T) {
	body := []byte(`{"model":"cqu-default","messages":[{"role":"user","content":"你好"}],"stream":true}`)
	routing := cquRouting{AgentID: "agent-1", ModelID: "model-1"}

	request, err := BuildCQUChatRequest(body, routing, "")
	require.NoError(t, err)
	require.Equal(t, "你好", request.Query)
	require.Equal(t, "", request.ConversationID)
	require.Equal(t, "", request.CustomConversationID)
	require.Equal(t, "agent-1", request.AgentID)
	require.Equal(t, "model-1", request.ModelID)
	require.Equal(t, 0, request.HasNetwork)
	require.Equal(t, []string{}, request.ContainerIDs)
	require.False(t, request.DeepThink)

	// The wire body carries exactly CQU's documented field set and nothing of
	// the OpenAI request shape.
	encoded, err := json.Marshal(request)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(encoded, &payload))
	require.ElementsMatch(t,
		[]string{"query", "conversation_id", "custom_conversation_id", "modelId", "agentId", "hasNetwork", "containerIds", "deepThink"},
		mapKeys(payload),
	)
	require.NotContains(t, payload, "messages")
	require.NotContains(t, payload, "stream")
}

func TestBuildCQUChatRequestFlattensSystemUserAssistant(t *testing.T) {
	body := []byte(`{"messages":[
		{"role":"system","content":"你是助手"},
		{"role":"user","content":"第一问"},
		{"role":"assistant","content":"第一答"},
		{"role":"user","content":"第二问"}
	]}`)

	request, err := BuildCQUChatRequest(body, cquRouting{AgentID: "a", ModelID: "m"}, "")
	require.NoError(t, err)
	require.Contains(t, request.Query, "System: 你是助手")
	require.Contains(t, request.Query, "User: 第一问")
	require.Contains(t, request.Query, "Assistant: 第一答")
	require.True(t, strings.HasSuffix(request.Query, "第二问"))

	// Once CQU owns the history, only the newest question is resent.
	request, err = BuildCQUChatRequest(body, cquRouting{AgentID: "a", ModelID: "m"}, "conv-9")
	require.NoError(t, err)
	require.Equal(t, "第二问", request.Query)
	require.Equal(t, "conv-9", request.ConversationID)
}

func TestBuildCQUChatRequestReadsMultipartTextContent(t *testing.T) {
	body := []byte(`{"messages":[{"role":"user","content":[{"type":"text","text":"你好"},{"type":"image_url","image_url":{"url":"x"}}]}]}`)

	request, err := BuildCQUChatRequest(body, cquRouting{AgentID: "a", ModelID: "m"}, "")
	require.NoError(t, err)
	require.Equal(t, "你好", request.Query)
}

func TestBuildCQUChatRequestRejectsEmptyMessages(t *testing.T) {
	_, err := BuildCQUChatRequest([]byte(`{"messages":[]}`), cquRouting{}, "")
	require.Error(t, err)

	_, err = BuildCQUChatRequest([]byte(`{"model":"cqu-default"}`), cquRouting{}, "")
	require.Error(t, err)
}

// The adapter must never require, read or transmit CAqWHAeT: the live page owns
// it. The same holds for session cookies.
func TestCQUBrowserProviderAdapterNeedsNoSecurityTokenOrCookie(t *testing.T) {
	adapter := newTestCQUAdapter(&fakeCQUSender{})
	account := cquTestAccount()
	account.Credentials = map[string]any{
		"caqwhaet":      "must-never-be-used",
		"CAqWHAeT":      "must-never-be-used",
		"cookie":        "SESSION=must-never-be-used",
		"authorization": "Bearer must-never-be-used",
		"refresh_token": "must-never-be-used",
	}
	body := []byte(`{"model":"cqu-default","messages":[{"role":"user","content":"你好"}],"stream":true}`)

	prepared, err := adapter.PrepareRequest(context.Background(), ProviderRequestInput{
		Account:  account,
		Protocol: ProviderProtocolChatCompletions,
		Model:    CQUDefaultModelAlias,
		Body:     body,
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, prepared.Method)
	require.True(t, prepared.Stream)

	req, err := adapter.BuildRequest(context.Background(), prepared)
	require.NoError(t, err)
	require.NoError(t, adapter.ApplyAuth(context.Background(), req, prepared))

	require.Equal(t, cquBrowserSendChatPath, req.URL.Path)
	require.Empty(t, req.URL.RawQuery, "no CAqWHAeT may be placed in the URL")
	require.Empty(t, req.Header.Get("Cookie"))
	require.Empty(t, req.Header.Get("Authorization"))

	raw, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.NotContains(t, string(raw), "must-never-be-used")

	// Building must also succeed with an account that has no credentials at all.
	bare, err := adapter.BuildRequest(context.Background(), ProviderRequestInput{
		Account: cquTestAccount(),
		Model:   CQUDefaultModelAlias,
		Body:    body,
	})
	require.NoError(t, err)
	require.Empty(t, bare.URL.RawQuery)
}

func TestCQUBrowserProviderAdapterSendsThroughBridgeAsOpenAISSE(t *testing.T) {
	sender := &fakeCQUSender{stream: cquTestStream("你", "好")}
	adapter := newTestCQUAdapter(sender)
	account := cquTestAccount()
	body := []byte(`{"model":"cqu-default","messages":[{"role":"user","content":"你好"}],"stream":true}`)

	resp, err := cquSendThroughAdapter(context.Background(), adapter, ProviderRequestInput{
		Account:  account,
		Protocol: ProviderProtocolChatCompletions,
		Model:    CQUDefaultModelAlias,
		Body:     body,
		Stream:   true,
	})
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Len(t, sender.requests, 1)
	require.Equal(t, "你好", sender.requests[0].Query)
	require.Equal(t, config.DefaultCQUAgentID, sender.requests[0].AgentID)
	require.Equal(t, config.DefaultCQUModelID, sender.requests[0].ModelID)

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	out := string(raw)

	require.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
	require.Contains(t, out, `"content":"你"`)
	require.Contains(t, out, `"content":"好"`)
	require.Contains(t, out, `"finish_reason":"stop"`)
	require.True(t, strings.HasSuffix(out, "data: [DONE]\n\n"))
	// No raw CQU frame may reach the client.
	require.NotContains(t, out, "TEXT_MESSAGE_CONTENT")
	require.NotContains(t, out, "RUN_FINISHED")
}

func TestCQUBrowserProviderAdapterNonStreamingReturnsChatCompletion(t *testing.T) {
	sender := &fakeCQUSender{stream: cquTestStream("你好", "，世界")}
	adapter := newTestCQUAdapter(sender)
	body := []byte(`{"model":"cqu-default","messages":[{"role":"user","content":"你好"}],"stream":false}`)

	resp, err := cquSendThroughAdapter(context.Background(), adapter, ProviderRequestInput{
		Account:  cquTestAccount(),
		Protocol: ProviderProtocolChatCompletions,
		Model:    CQUDefaultModelAlias,
		Body:     body,
	})
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var completion struct {
		Object  string `json:"object"`
		Model   string `json:"model"`
		Choices []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	require.NoError(t, json.Unmarshal(raw, &completion))
	require.Equal(t, "chat.completion", completion.Object)
	require.Equal(t, CQUDefaultModelAlias, completion.Model)
	require.Len(t, completion.Choices, 1)
	require.Equal(t, "assistant", completion.Choices[0].Message.Role)
	require.Equal(t, "你好，世界", completion.Choices[0].Message.Content)
	require.Equal(t, "stop", completion.Choices[0].FinishReason)
}

// The second turn of a conversation must reuse the CQU-side conversation id
// discovered on the first turn, and stop replaying the transcript.
func TestCQUBrowserProviderAdapterReusesConversationID(t *testing.T) {
	sender := &fakeCQUSender{stream: cquTestStream("第一答")}
	adapter := newTestCQUAdapter(sender)
	first := []byte(`{"model":"cqu-default","messages":[{"role":"user","content":"第一问"}],"stream":false}`)

	resp, err := cquSendThroughAdapter(context.Background(), adapter, ProviderRequestInput{
		Account: cquTestAccount(), Model: CQUDefaultModelAlias, Body: first,
	})
	require.NoError(t, err)
	_, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.Equal(t, "", sender.requests[0].ConversationID)

	second := []byte(`{"model":"cqu-default","messages":[
		{"role":"user","content":"第一问"},
		{"role":"assistant","content":"第一答"},
		{"role":"user","content":"第二问"}
	],"stream":false}`)
	sender.stream = cquTestStream("第二答")
	resp, err = cquSendThroughAdapter(context.Background(), adapter, ProviderRequestInput{
		Account: cquTestAccount(), Model: CQUDefaultModelAlias, Body: second,
	})
	require.NoError(t, err)
	_, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	require.Len(t, sender.requests, 2)
	require.Equal(t, "thread-1", sender.requests[1].ConversationID)
	require.Equal(t, "第二问", sender.requests[1].Query)
}

func TestCQUBrowserProviderAdapterCancellationStopsTranslation(t *testing.T) {
	sender := &fakeCQUSender{stream: cquTestStream("你好")}
	adapter := newTestCQUAdapter(sender)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := cquSendThroughAdapter(ctx, adapter, ProviderRequestInput{
		Account: cquTestAccount(),
		Model:   CQUDefaultModelAlias,
		Body:    []byte(`{"model":"cqu-default","messages":[{"role":"user","content":"你好"}]}`),
	})
	require.ErrorIs(t, err, context.Canceled)
}

func TestCQUBrowserProviderAdapterSurfacesUpstreamErrorBody(t *testing.T) {
	sender := &fakeCQUSender{status: http.StatusUnauthorized, body: `{"message":"登录失效，请重新登录"}`}
	adapter := newTestCQUAdapter(sender)

	resp, err := cquSendThroughAdapter(context.Background(), adapter, ProviderRequestInput{
		Account: cquTestAccount(),
		Model:   CQUDefaultModelAlias,
		Body:    []byte(`{"model":"cqu-default","messages":[{"role":"user","content":"你好"}],"stream":true}`),
	})
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	normalized := adapter.NormalizeError(resp, raw)
	require.True(t, normalized.AuthFailure)
	require.Equal(t, "authentication_error", normalized.Type)
	require.False(t, normalized.Failover)
}

func TestCQUBrowserProviderAdapterNormalizeError(t *testing.T) {
	adapter := newTestCQUAdapter(&fakeCQUSender{})

	rateLimited := adapter.NormalizeError(&http.Response{StatusCode: http.StatusTooManyRequests, Header: make(http.Header)}, []byte(`{}`))
	require.Equal(t, "rate_limit_error", rateLimited.Type)
	require.True(t, rateLimited.RateLimited)

	upstream := adapter.NormalizeError(&http.Response{StatusCode: http.StatusBadGateway, Header: make(http.Header)}, []byte(`{}`))
	require.Equal(t, "upstream_error", upstream.Type)
	require.True(t, upstream.Retryable)

	session := adapter.NormalizeError(
		&http.Response{StatusCode: http.StatusBadRequest, Header: make(http.Header)},
		[]byte(`{"message":"未登录"}`),
	)
	require.Equal(t, "authentication_error", session.Type)
	require.True(t, session.AuthFailure)
	require.False(t, session.Retryable)
}

// A bridge failure must be reported as an actionable operator error, never as a
// generic upstream fault and never with credential material attached.
func TestCQUBridgeErrorsMapToClientErrors(t *testing.T) {
	tests := []struct {
		kind       string
		wantStatus int
		wantType   string
	}{
		{CQUErrorKindDisabled, http.StatusServiceUnavailable, "upstream_error"},
		{CQUErrorKindBrowserUnavailable, http.StatusServiceUnavailable, CQUErrorKindBrowserUnavailable},
		{CQUErrorKindPageMissing, http.StatusServiceUnavailable, CQUErrorKindPageMissing},
		{CQUErrorKindTransport, http.StatusBadGateway, "upstream_error"},
	}
	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			err := newCQUBridgeError(tt.kind, "boom", nil)
			require.Equal(t, tt.kind, CQUBridgeErrorKind(err))
			status, errType, message := cquBridgeClientError(tt.kind, err)
			require.Equal(t, tt.wantStatus, status)
			require.Equal(t, tt.wantType, errType)
			require.NotEmpty(t, message)
			require.NotContains(t, strings.ToLower(message), "caqwhaet")
			require.NotContains(t, strings.ToLower(message), "cookie")
		})
	}
}

func TestCQUBrowserProviderAdapterExtractUsage(t *testing.T) {
	adapter := newTestCQUAdapter(&fakeCQUSender{})
	body, err := json.Marshal(map[string]any{
		"data": map[string]any{
			"usage": map[string]any{
				"input_tokens":      22,
				"output_tokens":     11,
				"cache_read_tokens": 4,
			},
		},
	})
	require.NoError(t, err)

	usage, err := adapter.ExtractUsage(body)
	require.NoError(t, err)
	require.Equal(t, 22, usage.InputTokens)
	require.Equal(t, 11, usage.OutputTokens)
	require.Equal(t, 4, usage.CachedTokens)
}

func TestCQUBrowserProviderAdapterParseStreamingEmitsNativeEvents(t *testing.T) {
	adapter := newTestCQUAdapter(&fakeCQUSender{})
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(cquTestStream("你好"))),
	}

	var types []string
	_, err := adapter.ParseStreaming(context.Background(), resp, func(event ProviderStreamEvent) error {
		types = append(types, event.Type)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, []string{
		CQUSSEEventRunStarted,
		CQUSSEEventTextMessageStart,
		CQUSSEEventTextMessageContent,
		CQUSSEEventTextMessageEnd,
		CQUSSEEventRunFinished,
	}, types)
}

func TestCQUBrowserProviderAdapterParseStreamingHonorsCancellation(t *testing.T) {
	adapter := newTestCQUAdapter(&fakeCQUSender{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("data: hello\n\n"))}

	_, err := adapter.ParseStreaming(ctx, resp, func(ProviderStreamEvent) error { return nil })
	require.ErrorIs(t, err, context.Canceled)
}

func TestCQUBrowserSendChatEndpointResolution(t *testing.T) {
	adapter := newTestCQUAdapter(&fakeCQUSender{})

	endpoint, err := adapter.sendChatEndpoint(cquTestAccount())
	require.NoError(t, err)
	require.Equal(t, config.DefaultCQUBaseURL+cquBrowserSendChatPath, endpoint)

	account := cquTestAccount()
	account.Credentials = map[string]any{"base_url": "https://example.test/prefix/"}
	endpoint, err = adapter.sendChatEndpoint(account)
	require.NoError(t, err)
	require.Equal(t, "https://example.test/prefix"+cquBrowserSendChatPath, endpoint)

	account.Credentials = map[string]any{"endpoint": "/relative"}
	_, err = adapter.sendChatEndpoint(account)
	require.Error(t, err)
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func TestCQUDefaultModelIsIsolatedToBrowserAccounts(t *testing.T) {
	t.Parallel()

	cquAccount := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Extra:    map[string]any{"cqu_browser": true},
	}
	ordinaryOpenAI := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
	}

	require.True(t, isOpenAICompatibleAccountModelSupported(cquAccount, CQUDefaultModelAlias))
	require.True(t, isOpenAICompatibleAccountModelSupported(cquAccount, CQUDefaultModelID))
	require.False(t, isOpenAICompatibleAccountModelSupported(cquAccount, "gpt-5"))
	require.False(t, isOpenAICompatibleAccountModelSupported(ordinaryOpenAI, CQUDefaultModelAlias))
	require.True(t, isOpenAICompatibleAccountModelSupported(ordinaryOpenAI, "gpt-5"))
}
