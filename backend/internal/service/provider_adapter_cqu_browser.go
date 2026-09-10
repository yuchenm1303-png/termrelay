package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/tidwall/gjson"
)

const (
	cquBrowserAdapterName  = "cqu_browser"
	cquBrowserSendChatPath = cquSendChatAPIPath
)

// CQUBrowserSender is the transport the adapter drives. It is satisfied by
// *CQUWebBridge; the interface exists so the adapter can be tested without a
// real browser, not so alternative HTTP transports can be plugged in.
//
// There is deliberately no Go-side HTTP client for CQU: the send-chat call must
// originate from the authenticated page so CQU's own runtime supplies the
// dynamic security parameter. TermRelay neither generates, stores, refreshes
// nor logs that parameter.
type CQUBrowserSender interface {
	SendChat(ctx context.Context, request CQUChatRequest) (*http.Response, error)
}

// CQUBrowserAdapterDeps wires the adapter to the browser bridge and to the
// shared configuration. Bridge is resolved lazily so the process can start
// (and every non-CQU provider keep working) while no browser is running.
type CQUBrowserAdapterDeps struct {
	Config        *config.Config
	Bridge        func(ctx context.Context) (CQUBrowserSender, error)
	Conversations *cquConversationStore
}

// CQUBrowserProviderAdapter bridges the authenticated CQU web chat transport
// into the provider-adapter boundary. It owns only CQU-specific
// request/response translation; scheduling, retries, billing, account failover
// and client response writes remain gateway concerns.
type CQUBrowserProviderAdapter struct {
	deps          CQUBrowserAdapterDeps
	conversations *cquConversationStore

	bridgeOnce sync.Once
	bridge     CQUBrowserSender
	bridgeErr  error
}

func NewCQUBrowserProviderAdapter(deps CQUBrowserAdapterDeps) *CQUBrowserProviderAdapter {
	conversations := deps.Conversations
	if conversations == nil {
		conversations = newCQUConversationStore(0, 0)
	}
	return &CQUBrowserProviderAdapter{deps: deps, conversations: conversations}
}

func (a *CQUBrowserProviderAdapter) Name() string {
	return cquBrowserAdapterName
}

// Supports matches CQU browser accounts only. It intentionally never matches on
// platform alone for the shared OpenAI platform: an OpenAI account becomes a
// CQU account only when it is explicitly marked, so no existing provider's
// routing changes.
func (a *CQUBrowserProviderAdapter) Supports(account *Account) bool {
	return IsCQUBrowserAccount(account)
}

// IsCQUBrowserAccount reports whether an account is served by the CQU browser
// bridge.
//
// Two shapes are accepted. The canonical one is platform=cqu (type browser or
// unset). The second is an openai-platform account carrying an explicit
// cqu_browser marker, which lets a CQU account live in an existing OpenAI group
// and reuse the scheduler, quota and billing paths unchanged.
func IsCQUBrowserAccount(account *Account) bool {
	if account == nil {
		return false
	}

	platform := strings.ToLower(strings.TrimSpace(account.Platform))
	accountType := strings.ToLower(strings.TrimSpace(account.Type))

	switch platform {
	case cquBrowserAdapterName:
		return true
	case "cqu":
		return accountType == "" || accountType == "browser" || accountType == cquBrowserAdapterName
	}

	return cquBrowserAccountMarked(account)
}

// cquBrowserAccountMarked looks for the opt-in marker on a non-CQU platform.
// Only an explicit truthy value counts, so an unrelated account can never be
// captured by accident.
func cquBrowserAccountMarked(account *Account) bool {
	if account == nil {
		return false
	}
	for _, source := range []map[string]any{account.Extra, account.Credentials} {
		for _, key := range []string{"cqu_browser", "cquBrowser"} {
			switch value := source[key].(type) {
			case bool:
				if value {
					return true
				}
			case string:
				switch strings.ToLower(strings.TrimSpace(value)) {
				case "1", "true", "yes", "on":
					return true
				}
			}
		}
		if provider, ok := source["provider"].(string); ok {
			if strings.EqualFold(strings.TrimSpace(provider), cquBrowserAdapterName) {
				return true
			}
		}
	}
	return false
}

// ResolveModel keeps the client-facing model name. CQU's numeric agent/model
// ids are resolved separately when the native body is built, so "cqu-default"
// is never sent upstream as a modelId.
func (a *CQUBrowserProviderAdapter) ResolveModel(account *Account, requestedModel string) (string, error) {
	if account == nil {
		return "", errors.New("account is required")
	}
	if !a.Supports(account) {
		return "", errors.New("account is not supported by cqu browser adapter")
	}
	if model := strings.TrimSpace(requestedModel); model != "" {
		return model, nil
	}
	return CQUDefaultModelAlias, nil
}

// ListModels exposes the logical CQU models. The list is static for now; the
// shape matches ProviderModelLister so a future sync from
// /api/chat-web/chatAgent/list can replace the body without changing callers.
func (a *CQUBrowserProviderAdapter) ListModels(_ context.Context, account *Account) ([]string, error) {
	if !a.Supports(account) {
		return nil, errors.New("account is not supported by cqu browser adapter")
	}
	return []string{CQUDefaultModelAlias}, nil
}

func (a *CQUBrowserProviderAdapter) NormalizeError(resp *http.Response, body []byte) NormalizedProviderError {
	normalized := normalizeGenericProviderError(resp, body)
	if resp == nil {
		return normalized
	}

	lowerBody := strings.ToLower(string(body))
	if resp.StatusCode >= 400 && resp.StatusCode < 500 && cquBrowserLooksLikeSessionFailure(lowerBody) {
		normalized.Type = "authentication_error"
		normalized.AuthFailure = true
		normalized.Retryable = false
		normalized.Failover = false
		normalized.Temporary = false
	}
	return normalized
}

func (a *CQUBrowserProviderAdapter) PrepareRequest(_ context.Context, input ProviderRequestInput) (ProviderRequestInput, error) {
	if input.Account == nil {
		return input, errors.New("account is required")
	}
	if !a.Supports(input.Account) {
		return input, errors.New("account is not supported by cqu browser adapter")
	}
	if input.Protocol != "" && input.Protocol != ProviderProtocolChatCompletions {
		return input, fmt.Errorf("cqu browser adapter does not support protocol %q", input.Protocol)
	}

	model, err := a.ResolveModel(input.Account, input.Model)
	if err != nil {
		return input, err
	}
	input.Model = model
	input.Method = http.MethodPost
	// The client's own stream flag is preserved: CQU only streams upstream, but
	// a non-streaming client must still receive a single JSON completion.
	input.Stream = gjson.GetBytes(input.Body, "stream").Bool()
	return input, nil
}

// cquRequestMeta travels with the descriptor request so SendRequest can render
// the right client wire format without re-parsing the OpenAI body.
type cquRequestMeta struct {
	Stream          bool
	Model           string
	ConversationKey string
}

type cquRequestMetaKey struct{}

// BuildRequest produces the CQU-native send-chat request.
//
// The result is a descriptor, not something this process sends: it carries no
// Cookie, no Authorization and no CAqWHAeT, because the browser page supplies
// all three. SendRequest hands the decoded body to the bridge.
func (a *CQUBrowserProviderAdapter) BuildRequest(ctx context.Context, input ProviderRequestInput) (*http.Request, error) {
	if input.Account == nil {
		return nil, errors.New("account is required")
	}
	if !a.Supports(input.Account) {
		return nil, errors.New("account is not supported by cqu browser adapter")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	endpoint, err := a.sendChatEndpoint(input.Account)
	if err != nil {
		return nil, err
	}

	conversationKey := CQUConversationKey(input.Body)
	routing := resolveCQURouting(input.Account, a.deps.Config, input.Model)
	chatRequest, err := BuildCQUChatRequest(input.Body, routing, a.conversations.Get(conversationKey))
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(chatRequest)
	if err != nil {
		return nil, fmt.Errorf("encode cqu chat request: %w", err)
	}

	meta := cquRequestMeta{
		Stream:          input.Stream || gjson.GetBytes(input.Body, "stream").Bool(),
		Model:           strings.TrimSpace(input.Model),
		ConversationKey: cquNextConversationKeySeed(input.Body),
	}
	if meta.Model == "" {
		meta.Model = CQUDefaultModelAlias
	}

	req, err := http.NewRequestWithContext(
		context.WithValue(ctx, cquRequestMetaKey{}, meta),
		http.MethodPost,
		endpoint,
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("build cqu browser request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	return req, nil
}

// ApplyAuth is intentionally a no-op.
//
// CQU authentication lives entirely in the signed-in browser profile: session
// cookies, bearer/refresh tokens and the dynamic CAqWHAeT parameter are held
// and rotated by the page. Attaching any of them here would mean persisting
// credentials TermRelay must never hold.
func (a *CQUBrowserProviderAdapter) ApplyAuth(_ context.Context, req *http.Request, input ProviderRequestInput) error {
	if req == nil {
		return errors.New("request is required")
	}
	if input.Account == nil {
		return errors.New("account is required")
	}
	return nil
}

// SendRequest performs the request through the browser bridge and returns a
// response the gateway's existing Chat Completions lifecycle can consume: an
// OpenAI SSE stream for streaming clients, a single chat.completion object
// otherwise. Raw CQU frames never leave this function.
func (a *CQUBrowserProviderAdapter) SendRequest(ctx context.Context, req *http.Request, account *Account) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}
	if !a.Supports(account) {
		return nil, errors.New("account is not supported by cqu browser adapter")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	meta, _ := req.Context().Value(cquRequestMetaKey{}).(cquRequestMeta)
	if strings.TrimSpace(meta.Model) == "" {
		meta.Model = CQUDefaultModelAlias
	}

	chatRequest, err := cquChatRequestFromHTTP(req)
	if err != nil {
		return nil, err
	}

	sender, err := a.sender(ctx)
	if err != nil {
		return nil, err
	}

	upstream, err := sender.SendChat(ctx, chatRequest)
	if err != nil {
		return nil, err
	}
	if upstream.StatusCode >= 400 {
		// Error bodies stay CQU-shaped on purpose: the gateway classifies them
		// through NormalizeError and renders the client-facing error itself.
		return upstream, nil
	}

	opts := cquOpenAIStreamOptions{Model: meta.Model}
	if meta.Stream {
		return a.streamingResponse(ctx, upstream, opts, meta), nil
	}
	return a.bufferedResponse(ctx, upstream, opts, meta)
}

// streamingResponse pipes the translated stream so the client starts receiving
// tokens while the browser is still producing them.
func (a *CQUBrowserProviderAdapter) streamingResponse(
	ctx context.Context,
	upstream *http.Response,
	opts cquOpenAIStreamOptions,
	meta cquRequestMeta,
) *http.Response {
	reader, writer := io.Pipe()
	go func() {
		defer func() { _ = upstream.Body.Close() }()
		outcome, err := WriteCQUStreamAsOpenAISSE(ctx, upstream.Body, writer, nil, opts)
		a.rememberConversation(meta, outcome)
		_ = writer.CloseWithError(err)
	}()

	header := make(http.Header)
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-cache")
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     header,
		Body:       reader,
		Request:    upstream.Request,
	}
}

func (a *CQUBrowserProviderAdapter) bufferedResponse(
	ctx context.Context,
	upstream *http.Response,
	opts cquOpenAIStreamOptions,
	meta cquRequestMeta,
) (*http.Response, error) {
	defer func() { _ = upstream.Body.Close() }()

	payload, outcome, err := AggregateCQUStreamAsOpenAIJSON(ctx, upstream.Body, opts)
	if err != nil {
		return nil, err
	}
	a.rememberConversation(meta, outcome)

	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     header,
		Body:       io.NopCloser(bytes.NewReader(payload)),
		Request:    upstream.Request,
	}, nil
}

// rememberConversation stores the CQU-side handle under the transcript the
// client will replay next turn, so a follow-up question continues the same
// upstream conversation instead of starting a new one.
func (a *CQUBrowserProviderAdapter) rememberConversation(meta cquRequestMeta, outcome cquStreamOutcome) {
	conversationID := outcome.Identifiers.ConversationKey()
	if conversationID == "" || meta.ConversationKey == "" {
		return
	}
	a.conversations.Put(cquFinalConversationKey(meta.ConversationKey, outcome.Text), conversationID)
}

func (a *CQUBrowserProviderAdapter) sender(ctx context.Context) (CQUBrowserSender, error) {
	a.bridgeOnce.Do(func() {
		if a.deps.Bridge != nil {
			a.bridge, a.bridgeErr = a.deps.Bridge(ctx)
			return
		}
		bridge, err := NewCQUWebBridgeFromConfig(a.deps.Config)
		if err != nil {
			a.bridgeErr = err
			return
		}
		a.bridge = bridge
	})
	if a.bridgeErr != nil {
		return nil, a.bridgeErr
	}
	if a.bridge == nil {
		return nil, newCQUBridgeError(CQUErrorKindDisabled, "cqu browser bridge is not configured", nil)
	}
	return a.bridge, nil
}

// sendChatEndpoint resolves the absolute send-chat URL. The account may pin a
// different CQU deployment; otherwise the configured base URL is used.
func (a *CQUBrowserProviderAdapter) sendChatEndpoint(account *Account) (string, error) {
	if endpoint := cquBrowserAccountString(account, "endpoint", "send_chat_url"); endpoint != "" {
		u, err := url.Parse(endpoint)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return "", errors.New("cqu browser endpoint must be an absolute URL")
		}
		return u.String(), nil
	}

	baseURL := cquBrowserAccountString(account, "base_url", "baseUrl")
	if baseURL == "" && a.deps.Config != nil {
		baseURL = strings.TrimSpace(a.deps.Config.CQU.BaseURL)
	}
	if baseURL == "" {
		baseURL = CQUDefaultOrigin
	}
	u, err := url.Parse(baseURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", errors.New("cqu browser base_url must be an absolute URL")
	}
	u.Path = strings.TrimRight(u.Path, "/") + cquBrowserSendChatPath
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

// cquChatRequestFromHTTP decodes the native body carried by a descriptor
// request built by BuildRequest.
func cquChatRequestFromHTTP(req *http.Request) (CQUChatRequest, error) {
	if req.Body == nil {
		return CQUChatRequest{}, errors.New("cqu browser request body is required")
	}
	defer func() { _ = req.Body.Close() }()

	payload, err := io.ReadAll(io.LimitReader(req.Body, 8<<20))
	if err != nil {
		return CQUChatRequest{}, fmt.Errorf("read cqu browser request body: %w", err)
	}
	var chatRequest CQUChatRequest
	if err := json.Unmarshal(payload, &chatRequest); err != nil {
		return CQUChatRequest{}, errors.New("cqu browser request body must be a CQU send-chat object")
	}
	if strings.TrimSpace(chatRequest.Query) == "" {
		return CQUChatRequest{}, errors.New("cqu query is required")
	}
	return chatRequest, nil
}

func (a *CQUBrowserProviderAdapter) ParseStreaming(ctx context.Context, resp *http.Response, emit func(ProviderStreamEvent) error) (ProviderUsage, error) {
	if resp == nil || resp.Body == nil {
		return ProviderUsage{}, errors.New("cqu browser response body is required")
	}
	if emit == nil {
		return ProviderUsage{}, errors.New("stream emitter is required")
	}

	usage := ProviderUsage{Extra: map[string]int64{}}
	err := ParseCQUSSEContext(ctx, resp.Body, func(event CQUSSEEvent) error {
		if len(event.Data) > 0 {
			usage = mergeProviderUsage(usage, cquBrowserExtractUsage(event.Data))
		}
		eventType := event.Type
		if eventType == "" {
			if bytes.Equal(bytes.TrimSpace(event.Data), []byte("[DONE]")) {
				eventType = "done"
			} else {
				eventType = "message"
			}
		}
		return emit(ProviderStreamEvent{Type: eventType, Data: event.Data, Raw: event.Raw})
	})
	return usage, err
}

func (a *CQUBrowserProviderAdapter) ParseNonStreaming(_ context.Context, resp *http.Response) (ProviderResponse, error) {
	if resp == nil || resp.Body == nil {
		return ProviderResponse{}, errors.New("cqu browser response body is required")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ProviderResponse{}, fmt.Errorf("read cqu browser response: %w", err)
	}
	return ProviderResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       body,
		Usage:      cquBrowserExtractUsage(body),
	}, nil
}

func (a *CQUBrowserProviderAdapter) ExtractUsage(payload []byte) (ProviderUsage, error) {
	return cquBrowserExtractUsage(payload), nil
}

// cquBrowserAccountString reads a non-secret routing value (agent id, model id,
// base url) from an account. It is never used for credentials.
func cquBrowserAccountString(account *Account, keys ...string) string {
	if account == nil {
		return ""
	}
	for _, source := range []map[string]any{account.Credentials, account.Extra} {
		for _, key := range keys {
			value, ok := source[key]
			if !ok || value == nil {
				continue
			}
			switch v := value.(type) {
			case string:
				if trimmed := strings.TrimSpace(v); trimmed != "" {
					return trimmed
				}
			case json.Number:
				return v.String()
			case float64:
				return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintf("%.0f", v), ".0"), ".")
			}
		}
	}
	return ""
}

func cquBrowserLooksLikeSessionFailure(body string) bool {
	for _, marker := range []string{
		"caqwhaet",
		"未登录",
		"登录失效",
		"登录过期",
		"会话失效",
		"session expired",
		"invalid session",
		"unauthorized",
		"authentication",
	} {
		if strings.Contains(body, marker) {
			return true
		}
	}
	return false
}

func cquBrowserExtractUsage(payload []byte) ProviderUsage {
	usage := ProviderUsage{Extra: map[string]int64{}}
	var root any
	if json.Unmarshal(payload, &root) != nil {
		return usage
	}
	walkCQUUsage(root, &usage)
	return usage
}

func walkCQUUsage(value any, usage *ProviderUsage) {
	if usage == nil {
		return
	}
	switch node := value.(type) {
	case map[string]any:
		for key, child := range node {
			normalizedKey := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "-", "_"), " ", "_"))
			if n, ok := cquBrowserInt64(child); ok {
				switch normalizedKey {
				case "prompt_tokens", "input_tokens", "prompttokens", "inputtokens":
					if int(n) > usage.InputTokens {
						usage.InputTokens = int(n)
					}
				case "completion_tokens", "output_tokens", "completiontokens", "outputtokens":
					if int(n) > usage.OutputTokens {
						usage.OutputTokens = int(n)
					}
				case "cached_tokens", "cache_read_tokens", "cachedtokens":
					if int(n) > usage.CachedTokens {
						usage.CachedTokens = int(n)
					}
				}
			}
			walkCQUUsage(child, usage)
		}
	case []any:
		for _, child := range node {
			walkCQUUsage(child, usage)
		}
	}
}

func cquBrowserInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case float64:
		return int64(v), true
	case json.Number:
		n, err := v.Int64()
		return n, err == nil
	case int:
		return int64(v), true
	case int64:
		return v, true
	default:
		return 0, false
	}
}

func mergeProviderUsage(current, next ProviderUsage) ProviderUsage {
	if next.InputTokens > current.InputTokens {
		current.InputTokens = next.InputTokens
	}
	if next.OutputTokens > current.OutputTokens {
		current.OutputTokens = next.OutputTokens
	}
	if next.CachedTokens > current.CachedTokens {
		current.CachedTokens = next.CachedTokens
	}
	if current.Extra == nil {
		current.Extra = map[string]int64{}
	}
	for key, value := range next.Extra {
		if value > current.Extra[key] {
			current.Extra[key] = value
		}
	}
	return current
}
