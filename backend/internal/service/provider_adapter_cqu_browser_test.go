package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestCQUBrowserProviderAdapterSupports(t *testing.T) {
	adapter := NewCQUBrowserProviderAdapter()

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
		{name: "other", account: &Account{Platform: "openai"}, want: false},
		{name: "cqu incompatible type", account: &Account{Platform: "cqu", Type: "api_key"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := adapter.Supports(tt.account); got != tt.want {
				t.Fatalf("Supports() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCQUBrowserProviderAdapterBuildRequest(t *testing.T) {
	adapter := NewCQUBrowserProviderAdapter()
	account := &Account{
		Platform: "cqu",
		Type:     "browser",
		Credentials: map[string]any{
			"base_url":  "https://example.cqu.edu.cn",
			"caqwhaet":  "secret-query-token",
			"cookie":    "SESSION=secret-cookie",
			"agent_id":  "agent-123",
			"model_id":  "model-456",
			"user_agent": "TermRelay-CQU-Test/1.0",
		},
	}
	input := ProviderRequestInput{
		Account:  account,
		Protocol: ProviderProtocolChatCompletions,
		Model:    "logical-model",
		Body:     []byte(`{"messages":[{"role":"user","content":"hello"}],"stream":false}`),
	}
	// Raw string fixtures must contain real JSON, not escaped quote bytes.
	input.Body = []byte(`{"messages":[{"role":"user","content":"hello"}],"stream":false}`)
	input.Body = []byte(`{"messages":[{"role":"user","content":"hello"}],"stream":false}`)

	prepared, err := adapter.PrepareRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("PrepareRequest() error = %v", err)
	}
	if prepared.Method != http.MethodPost || !prepared.Stream {
		t.Fatalf("PrepareRequest() method/stream = %q/%v, want POST/true", prepared.Method, prepared.Stream)
	}

	req, err := adapter.BuildRequest(context.Background(), prepared)
	if err != nil {
		t.Fatalf("BuildRequest() error = %v", err)
	}
	if err := adapter.ApplyAuth(context.Background(), req, prepared); err != nil {
		t.Fatalf("ApplyAuth() error = %v", err)
	}

	if req.Method != http.MethodPost {
		t.Fatalf("method = %q, want POST", req.Method)
	}
	if req.URL.Path != cquBrowserSendChatPath {
		t.Fatalf("path = %q, want %q", req.URL.Path, cquBrowserSendChatPath)
	}
	if got := req.URL.Query().Get("CAqWHAeT"); got != "secret-query-token" {
		t.Fatalf("CAqWHAeT = %q", got)
	}
	if got := req.Header.Get("Cookie"); got != "SESSION=secret-cookie" {
		t.Fatalf("Cookie = %q", got)
	}
	if got := req.Header.Get("Accept"); got != "text/event-stream" {
		t.Fatalf("Accept = %q", got)
	}
	if got := req.Header.Get("Origin"); got != "https://example.cqu.edu.cn" {
		t.Fatalf("Origin = %q", got)
	}
	if got := req.Header.Get("Referer"); got != "https://example.cqu.edu.cn/chat" {
		t.Fatalf("Referer = %q", got)
	}
	if got := req.Header.Get("User-Agent"); got != "TermRelay-CQU-Test/1.0" {
		t.Fatalf("User-Agent = %q", got)
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("ReadAll(body) error = %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json.Unmarshal(body) error = %v", err)
	}
	if payload["agentId"] != "agent-123" {
		t.Fatalf("agentId = %#v", payload["agentId"])
	}
	if payload["modelId"] != "model-456" {
		t.Fatalf("modelId = %#v", payload["modelId"])
	}
	if payload["stream"] != true {
		t.Fatalf("stream = %#v, want true", payload["stream"])
	}
	if _, ok := payload["messages"]; !ok {
		t.Fatalf("messages were not preserved")
	}
}

func TestCQUBrowserProviderAdapterBuildRequestUsesExplicitEndpoint(t *testing.T) {
	adapter := NewCQUBrowserProviderAdapter()
	account := &Account{
		Platform: "cqu_browser",
		Credentials: map[string]any{
			"endpoint": "https://example.test/custom/send?keep=1",
			"CAqWHAeT": "token",
		},
	}

	req, err := adapter.BuildRequest(context.Background(), ProviderRequestInput{Account: account, Body: []byte(`{}`)})
	if err != nil {
		t.Fatalf("BuildRequest() error = %v", err)
	}
	if req.URL.Path != "/custom/send" || req.URL.Query().Get("keep") != "1" || req.URL.Query().Get("CAqWHAeT") != "token" {
		t.Fatalf("unexpected URL: %s", req.URL.String())
	}
}

func TestCQUBrowserProviderAdapterBuildRequestDoesNotLeakSecretInErrors(t *testing.T) {
	adapter := NewCQUBrowserProviderAdapter()
	account := &Account{
		Platform: "cqu_browser",
		Credentials: map[string]any{
			"base_url": "not-a-url",
			"caqwhaet": "super-secret",
			"cookie":   "SESSION=also-secret",
		},
	}

	_, err := adapter.BuildRequest(context.Background(), ProviderRequestInput{Account: account, Body: []byte(`{}`)})
	if err == nil {
		t.Fatal("BuildRequest() error = nil, want error")
	}
	if strings.Contains(err.Error(), "super-secret") || strings.Contains(err.Error(), "also-secret") {
		t.Fatalf("error leaked credentials: %v", err)
	}
}

func TestCQUBrowserProviderAdapterParseStreaming(t *testing.T) {
	adapter := NewCQUBrowserProviderAdapter()
	stream := strings.Join([]string{
		"event: message",
		`data: {"delta":"hello","usage":{"prompt_tokens":12}}`,
		"",
		"event: message",
		`data: {"delta":" world","usage":{"completion_tokens":7,"cached_tokens":3}}`,
		"",
		"data: [DONE]",
		"",
	}, "\n")
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(stream)),
	}

	var events []ProviderStreamEvent
	usage, err := adapter.ParseStreaming(context.Background(), resp, func(event ProviderStreamEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("ParseStreaming() error = %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("event count = %d, want 3", len(events))
	}
	if events[0].Type != "message" || !strings.Contains(string(events[0].Data), "hello") {
		t.Fatalf("unexpected first event: %#v", events[0])
	}
	if events[2].Type != "done" || string(events[2].Data) != "[DONE]" {
		t.Fatalf("unexpected done event: %#v", events[2])
	}
	if usage.InputTokens != 12 || usage.OutputTokens != 7 || usage.CachedTokens != 3 {
		t.Fatalf("usage = %#v", usage)
	}
}

func TestCQUBrowserProviderAdapterParseStreamingHonorsCancellation(t *testing.T) {
	adapter := NewCQUBrowserProviderAdapter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("data: hello\n\n"))}

	_, err := adapter.ParseStreaming(ctx, resp, func(ProviderStreamEvent) error { return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ParseStreaming() error = %v, want context.Canceled", err)
	}
}

func TestCQUBrowserProviderAdapterParseStreamingPropagatesEmitterError(t *testing.T) {
	adapter := NewCQUBrowserProviderAdapter()
	wantErr := errors.New("client write failed")
	resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("data: hello\n\n"))}

	_, err := adapter.ParseStreaming(context.Background(), resp, func(ProviderStreamEvent) error { return wantErr })
	if !errors.Is(err, wantErr) {
		t.Fatalf("ParseStreaming() error = %v, want %v", err, wantErr)
	}
}

func TestCQUBrowserProviderAdapterNormalizeError(t *testing.T) {
	adapter := NewCQUBrowserProviderAdapter()
	resp := &http.Response{StatusCode: http.StatusBadRequest, Header: make(http.Header)}
	got := adapter.NormalizeError(resp, []byte(`{"message":"登录失效，请重新登录"}`))
	if !got.AuthFailure || got.Type != "authentication_error" {
		t.Fatalf("NormalizeError() = %#v", got)
	}
	if got.Retryable || got.Failover || got.Temporary {
		t.Fatalf("session failure should not be retried/failovered: %#v", got)
	}
}

func TestCQUBrowserProviderAdapterExtractUsage(t *testing.T) {
	adapter := NewCQUBrowserProviderAdapter()
	usage, err := adapter.ExtractUsage([]byte(`{"data":{"usage":{"input_tokens":22,"output_tokens":11,"cache_read_tokens":4}}}`))
	if err != nil {
		t.Fatalf("ExtractUsage() error = %v", err)
	}
	if usage.InputTokens != 22 || usage.OutputTokens != 11 || usage.CachedTokens != 4 {
		t.Fatalf("usage = %#v", usage)
	}
}

func TestCQUBrowserEndpointRejectsRelativeURL(t *testing.T) {
	_, err := cquBrowserEndpoint(&Account{Credentials: map[string]any{"endpoint": "/relative"}})
	if err == nil {
		t.Fatal("cquBrowserEndpoint() error = nil, want error")
	}
}

func TestCQUBrowserEndpointPreservesBasePrefix(t *testing.T) {
	got, err := cquBrowserEndpoint(&Account{Credentials: map[string]any{"base_url": "https://example.test/prefix/"}})
	if err != nil {
		t.Fatalf("cquBrowserEndpoint() error = %v", err)
	}
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	if u.Path != "/prefix"+cquBrowserSendChatPath {
		t.Fatalf("path = %q", u.Path)
	}
}
