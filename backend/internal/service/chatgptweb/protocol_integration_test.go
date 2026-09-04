package chatgptweb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func newTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server, *ClientState) {
	t.Helper()
	server := httptest.NewServer(handler)
	identity, err := NewIdentity("secret-access-token", "acct-1", "device-1", "ua-1", "session=secret-cookie")
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(server.Client(), server.URL, identity)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 4, 8, 0, 0, 0, time.UTC)
	client.now = func() time.Time { return now }
	state := NewClientState("device-1", "ua-1", now.Add(-3*time.Second))
	state.SessionID = "session-1"
	return client, server, state
}

func TestIdentityDoesNotFormatSecrets(t *testing.T) {
	identity, err := NewIdentity("top-secret", "acct", "device", "ua", "cookie-secret")
	if err != nil {
		t.Fatal(err)
	}
	rendered := fmt.Sprintf("%v %#v", identity, identity)
	if strings.Contains(rendered, "top-secret") || strings.Contains(rendered, "cookie-secret") {
		t.Fatalf("identity leaked secret: %s", rendered)
	}
	encoded, err := json.Marshal(identity)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "top-secret") || strings.Contains(string(encoded), "cookie-secret") {
		t.Fatalf("json leaked secret: %s", encoded)
	}
}

func TestTransformChatCompletionsContinuedConversationUsesTailUserTurn(t *testing.T) {
	now := time.Date(2026, 9, 4, 8, 0, 0, 0, time.UTC)
	state := NewClientState("device", "ua", now.Add(-2*time.Second))
	state.ConversationID = "conv-1"
	state.ParentMessageID = "msg-prev"
	body := []byte(`{"model":"gpt-5","stream":true,"temperature":0.2,"messages":[{"role":"system","content":"rules"},{"role":"user","content":"old"},{"role":"assistant","content":"answer"},{"role":"user","content":[{"type":"text","text":"new"}]}]}`)
	request, stream, err := TransformChatCompletions(body, state, now)
	if err != nil {
		t.Fatal(err)
	}
	if !stream || request.ConversationID != "conv-1" || request.ParentMessageID != "msg-prev" || !request.HistoryAndTrainingDisabled {
		t.Fatalf("request = %+v stream=%v", request, stream)
	}
	if len(request.Messages) != 1 || request.Messages[0].Author.Role != "user" || request.Messages[0].Content.Parts[0] != "new" {
		t.Fatalf("tail messages = %+v", request.Messages)
	}
}

func TestSentinelNoChallenge(t *testing.T) {
	var paths []string
	client, server, state := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if r.Header.Get("Authorization") != "Bearer secret-access-token" || r.Header.Get("Oai-Device-Id") != "device-1" || r.Header.Get("Oai-Session-Id") != "session-1" {
			t.Errorf("identity headers missing")
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/sentinel/chat-requirements/prepare":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["p"] != "requirements-token" {
				t.Errorf("prepare body = %+v", body)
			}
			_, _ = io.WriteString(w, `{"prepare_token":"prepare-1","proofofwork":{"required":false},"turnstile":{"required":false}}`)
		case "/sentinel/chat-requirements/finalize":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["prepare_token"] != "prepare-1" {
				t.Errorf("finalize body = %+v", body)
			}
			_, _ = io.WriteString(w, `{"token":"final-1"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	flow, err := NewSentinelFlow(client, RequirementsTokenProviderFunc(func(context.Context, *ClientState) (string, error) { return "requirements-token", nil }), nil)
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := flow.Run(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if tokens.PrepareToken != "prepare-1" || tokens.ChatRequirementsToken != "final-1" || tokens.ProofToken != "" {
		t.Fatalf("tokens=%+v", tokens)
	}
	if !reflect.DeepEqual(paths, []string{"/sentinel/chat-requirements/prepare", "/sentinel/chat-requirements/finalize"}) {
		t.Fatalf("paths=%v", paths)
	}
}

func TestSentinelPoWRequiredUsesInjectedSolver(t *testing.T) {
	solverCalled := false
	client, server, state := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/sentinel/chat-requirements/prepare":
			_, _ = io.WriteString(w, `{"prepare_token":"prepare-1","proofofwork":{"required":true,"seed":"seed-1","difficulty":"hard"},"turnstile":{"required":false}}`)
		case "/sentinel/chat-requirements/finalize":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["proofofwork"] != "proof-1" {
				t.Errorf("finalize=%+v", body)
			}
			_, _ = io.WriteString(w, `{"token":"final-1"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	flow, err := NewSentinelFlow(client, RequirementsTokenProviderFunc(func(context.Context, *ClientState) (string, error) { return "requirements-token", nil }), ProofOfWorkSolverFunc(func(_ context.Context, p ProofRequirement, _ *ClientState) (string, error) {
		solverCalled = true
		if p.Seed != "seed-1" || p.Difficulty != "hard" {
			t.Errorf("proof=%+v", p)
		}
		return "proof-1", nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := flow.Run(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if !solverCalled || tokens.ProofToken != "proof-1" {
		t.Fatalf("solver=%v tokens=%+v", solverCalled, tokens)
	}
}

func TestSentinelHumanChallengeStopsBeforeFinalize(t *testing.T) {
	finalizeCalls := 0
	client, server, state := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/sentinel/chat-requirements/finalize" {
			finalizeCalls++
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"prepare_token":"prepare-1","turnstile":{"required":true,"dx":"human"}}`)
	}))
	defer server.Close()
	flow, _ := NewSentinelFlow(client, RequirementsTokenProviderFunc(func(context.Context, *ClientState) (string, error) { return "requirements-token", nil }), nil)
	_, err := flow.Run(context.Background(), state)
	upstream, ok := err.(*UpstreamError)
	if !ok || !upstream.ChallengeRequired || upstream.RetryNextAccount || finalizeCalls != 0 {
		t.Fatalf("err=%T %+v finalize=%d", err, err, finalizeCalls)
	}
}

func TestConversationPrepareThreeStagesAndOpen(t *testing.T) {
	var mu sync.Mutex
	var stages []string
	var conduits []string
	client, server, state := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/conversation/init":
			_, _ = io.WriteString(w, `{"type":"success","default_model_slug":"gpt-5"}`)
		case "/f/conversation/prepare":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			stage, _ := body["client_prepare_state"].(string)
			stages = append(stages, stage)
			conduits = append(conduits, r.Header.Get("X-Conduit-Token"))
			if stage != "none" {
				if _, ok := body["partial_query"]; !ok {
					t.Errorf("stage %s missing partial_query", stage)
				}
			}
			token := map[string]string{"none": "c1", "sent": "c2", "success": "c3"}[stage]
			_, _ = fmt.Fprintf(w, `{"conduit_token":%q}`, token)
		case "/f/conversation":
			if got := r.Header.Get("X-Conduit-Token"); got != "c3" {
				t.Errorf("conduit=%q", got)
			}
			if got := r.Header.Get("Openai-Sentinel-Chat-Requirements-Token"); got != "sentinel-final" {
				t.Errorf("sentinel=%q", got)
			}
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = io.WriteString(w, "data: {\"conversation_id\":\"conv-new\"}\n\ndata: [DONE]\n\n")
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	request := ConversationRequest{Action: "next", Messages: []ConversationMessage{{ID: "u1", Author: ConversationAuthor{Role: "user"}, Content: ConversationContent{ContentType: "text", Parts: []any{"hello"}}}}, ParentMessageID: RootParentMessageID, Model: "gpt-5", Timezone: "UTC", ConversationMode: map[string]string{"kind": "primary_assistant"}, SystemHints: []string{}, HistoryAndTrainingDisabled: true}
	if _, err := client.InitConversation(context.Background(), state); err != nil {
		t.Fatal(err)
	}
	sentinel := &SentinelTokens{PrepareToken: "sentinel-prepare", ChatRequirementsToken: "sentinel-final", ProofToken: "proof"}
	conduit, err := client.PrepareConversation(context.Background(), request, state, sentinel, "trace-1")
	if err != nil {
		t.Fatal(err)
	}
	if conduit != "c3" {
		t.Fatalf("conduit=%q", conduit)
	}
	if !reflect.DeepEqual(stages, []string{"none", "sent", "success"}) || !reflect.DeepEqual(conduits, []string{"", "c1", "c2"}) {
		t.Fatalf("stages=%v conduits=%v", stages, conduits)
	}
	resp, err := client.OpenConversation(context.Background(), request, state, sentinel, conduit, "trace-1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "[DONE]") {
		t.Fatalf("body=%q", body)
	}
}

func TestHTTPErrorDoesNotExposeUpstreamBody(t *testing.T) {
	client, server, state := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"token":"super-secret-upstream-body"}`)
	}))
	defer server.Close()
	_, err := client.InitConversation(context.Background(), state)
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "super-secret-upstream-body") {
		t.Fatalf("error leaked body: %v", err)
	}
}
