package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	cquBrowserAdapterName    = "cqu_browser"
	cquBrowserDefaultAgentID = "910020080887140352"
	cquBrowserDefaultModelID = "910023701267746816"
	cquBrowserSendChatPath   = "/api/chat-web/message-chat/send-chat"
)

// CQUBrowserProviderAdapter bridges the authenticated CQU web chat transport
// into the provider-adapter boundary. It intentionally owns only CQU-specific
// request/auth/response translation; scheduling, retries, billing and client
// response writes remain gateway concerns.
type CQUBrowserProviderAdapter struct{}

func NewCQUBrowserProviderAdapter() *CQUBrowserProviderAdapter {
	return &CQUBrowserProviderAdapter{}
}

func (a *CQUBrowserProviderAdapter) Name() string {
	return cquBrowserAdapterName
}

func (a *CQUBrowserProviderAdapter) Supports(account *Account) bool {
	if account == nil {
		return false
	}

	platform := strings.ToLower(strings.TrimSpace(account.Platform))
	accountType := strings.ToLower(strings.TrimSpace(account.Type))
	if platform == cquBrowserAdapterName {
		return true
	}
	if platform != "cqu" {
		return false
	}
	return accountType == "" || accountType == "browser" || accountType == cquBrowserAdapterName
}

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
	if model := cquBrowserAccountString(account, "model_id", "modelId"); model != "" {
		return model, nil
	}
	return cquBrowserDefaultModelID, nil
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
	input.Stream = true
	return input, nil
}

func (a *CQUBrowserProviderAdapter) BuildRequest(ctx context.Context, input ProviderRequestInput) (*http.Request, error) {
	if input.Account == nil {
		return nil, errors.New("account is required")
	}
	if !a.Supports(input.Account) {
		return nil, errors.New("account is not supported by cqu browser adapter")
	}

	endpoint, err := cquBrowserEndpoint(input.Account)
	if err != nil {
		return nil, err
	}

	token := cquBrowserAccountString(input.Account, "caqwhaet", "CAqWHAeT")
	if token == "" {
		return nil, errors.New("cqu browser account is missing CAqWHAeT session token")
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.New("cqu browser endpoint is invalid")
	}
	q := u.Query()
	q.Set("CAqWHAeT", token)
	u.RawQuery = q.Encode()

	body, err := cquBrowserBuildNativeBody(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build cqu browser request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	origin := cquBrowserAccountString(input.Account, "origin")
	if origin == "" {
		origin = u.Scheme + "://" + u.Host
	}
	if origin != "://" {
		req.Header.Set("Origin", origin)
	}

	referer := cquBrowserAccountString(input.Account, "referer")
	if referer == "" && origin != "" && origin != "://" {
		referer = strings.TrimRight(origin, "/") + "/chat"
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	if ua := cquBrowserAccountString(input.Account, "user_agent", "userAgent"); ua != "" {
		req.Header.Set("User-Agent", ua)
	}

	return req, nil
}

func (a *CQUBrowserProviderAdapter) ApplyAuth(_ context.Context, req *http.Request, input ProviderRequestInput) error {
	if req == nil {
		return errors.New("request is required")
	}
	if input.Account == nil {
		return errors.New("account is required")
	}
	if cookie := cquBrowserAccountString(input.Account, "cookie"); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	return nil
}

func (a *CQUBrowserProviderAdapter) ParseStreaming(ctx context.Context, resp *http.Response, emit func(ProviderStreamEvent) error) (ProviderUsage, error) {
	if resp == nil || resp.Body == nil {
		return ProviderUsage{}, errors.New("cqu browser response body is required")
	}
	if emit == nil {
		return ProviderUsage{}, errors.New("stream emitter is required")
	}

	reader := bufio.NewReader(resp.Body)
	usage := ProviderUsage{Extra: map[string]int64{}}
	var raw bytes.Buffer
	var data bytes.Buffer
	eventType := ""

	emitFrame := func() error {
		if raw.Len() == 0 && data.Len() == 0 && eventType == "" {
			return nil
		}
		payload := bytes.TrimSuffix(data.Bytes(), []byte("\n"))
		if len(payload) > 0 {
			usage = mergeProviderUsage(usage, cquBrowserExtractUsage(payload))
		}
		kind := strings.TrimSpace(eventType)
		if kind == "" {
			if bytes.Equal(bytes.TrimSpace(payload), []byte("[DONE]")) {
				kind = "done"
			} else {
				kind = "message"
			}
		}
		event := ProviderStreamEvent{
			Type: kind,
			Data: append([]byte(nil), payload...),
			Raw:  append([]byte(nil), raw.Bytes()...),
		}
		raw.Reset()
		data.Reset()
		eventType = ""
		return emit(event)
	}

	for {
		if err := ctx.Err(); err != nil {
			return usage, err
		}

		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			raw.WriteString(line)
			trimmed := strings.TrimRight(line, "\r\n")
			if trimmed == "" {
				if emitErr := emitFrame(); emitErr != nil {
					return usage, emitErr
				}
			} else if strings.HasPrefix(trimmed, "event:") {
				eventType = strings.TrimSpace(strings.TrimPrefix(trimmed, "event:"))
			} else if strings.HasPrefix(trimmed, "data:") {
				if data.Len() > 0 {
					data.WriteByte('\n')
				}
				data.WriteString(strings.TrimSpace(strings.TrimPrefix(trimmed, "data:")))
			}
		}

		if err != nil {
			if err == io.EOF {
				if raw.Len() > 0 || data.Len() > 0 || eventType != "" {
					if emitErr := emitFrame(); emitErr != nil {
						return usage, emitErr
					}
				}
				return usage, nil
			}
			return usage, fmt.Errorf("read cqu browser stream: %w", err)
		}
	}
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

func cquBrowserEndpoint(account *Account) (string, error) {
	if endpoint := cquBrowserAccountString(account, "endpoint", "send_chat_url"); endpoint != "" {
		u, err := url.Parse(endpoint)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return "", errors.New("cqu browser endpoint must be an absolute URL")
		}
		return u.String(), nil
	}

	baseURL := cquBrowserAccountString(account, "base_url", "baseUrl")
	if baseURL == "" {
		return "", errors.New("cqu browser account is missing base_url")
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

func cquBrowserBuildNativeBody(input ProviderRequestInput) ([]byte, error) {
	payload := make(map[string]any)
	if len(bytes.TrimSpace(input.Body)) > 0 {
		if err := json.Unmarshal(input.Body, &payload); err != nil {
			return nil, errors.New("cqu browser request body must be a JSON object")
		}
	}

	agentID := cquBrowserAccountString(input.Account, "agent_id", "agentId")
	if agentID == "" {
		agentID = cquBrowserDefaultAgentID
	}
	modelID := cquBrowserAccountString(input.Account, "model_id", "modelId")
	if modelID == "" {
		modelID = cquBrowserDefaultModelID
	}

	payload["agentId"] = agentID
	payload["modelId"] = modelID
	payload["stream"] = true

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode cqu browser request body: %w", err)
	}
	return body, nil
}

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
