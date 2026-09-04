package chatgptweb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPTransportExecutesAccountScopedConversationFlow(t *testing.T) {
	var conversationCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret-access-token" {
			t.Errorf("authorization header missing")
		}
		if r.Header.Get("Cookie") != "session=secret-cookie" {
			t.Errorf("cookie header missing")
		}
		if r.Header.Get("Chatgpt-Account-Id") != "upstream-account" {
			t.Errorf("account header = %q", r.Header.Get("Chatgpt-Account-Id"))
		}
		if r.Header.Get("Oai-Device-Id") != "browser-device" || r.Header.Get("User-Agent") != "browser-ua" {
			t.Errorf("browser fingerprint = %q %q", r.Header.Get("Oai-Device-Id"), r.Header.Get("User-Agent"))
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/conversation/init":
			_, _ = io.WriteString(w, `{"type":"success","default_model_slug":"gpt-5"}`)
		case "/sentinel/chat-requirements/prepare":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["p"] != "requirements-token" {
				t.Errorf("requirements token missing")
			}
			_, _ = io.WriteString(w, `{"prepare_token":"prepare-1","proofofwork":{"required":false},"turnstile":{"required":false}}`)
		case "/sentinel/chat-requirements/finalize":
			_, _ = io.WriteString(w, `{"token":"sentinel-final"}`)
		case "/f/conversation/prepare":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			stage, _ := body["client_prepare_state"].(string)
			token := map[string]string{"none": "c1", "sent": "c2", "success": "c3"}[stage]
			if token == "" {
				t.Errorf("unexpected prepare stage %q", stage)
			}
			_, _ = fmt.Fprintf(w, `{"conduit_token":%q}`, token)
		case "/f/conversation":
			conversationCalls++
			if r.Header.Get("X-Conduit-Token") != "c3" {
				t.Errorf("conduit token = %q", r.Header.Get("X-Conduit-Token"))
			}
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = io.WriteString(w, "data: {\"conversation_id\":\"conv-new\"}\n\ndata: [DONE]\n\n")
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	transport := newHTTPTransportForTest(t, server, "requirements-token")
	state := NewClientState("downstream-device", "downstream-ua", time.Now().Add(-time.Second))
	state.SessionID = "client-session"
	request := TransportRequest{
		ClientState: *state,
		Conversation: ConversationRequest{
			Action:          "next",
			ParentMessageID: RootParentMessageID,
			Model:           "gpt-5",
			Timezone:        "UTC",
			ConversationMode: map[string]string{"kind": "primary_assistant"},
			SystemHints:      []string{},
			Messages: []ConversationMessage{{
				ID:      "user-1",
				Author:  ConversationAuthor{Role: "user"},
				Content: ConversationContent{ContentType: "text", Parts: []any{"hello"}},
			}},
		},
	}

	snapshot, err := transport.Send(context.Background(), AccountRef{ID: 7}, request)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.ConversationID != "conv-new" || conversationCalls != 1 {
		t.Fatalf("snapshot=%+v conversation_calls=%d", snapshot, conversationCalls)
	}
}

func TestHTTPTransportStopsOnHumanChallenge(t *testing.T) {
	var conversationCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/conversation/init":
			_, _ = io.WriteString(w, `{"type":"success"}`)
		case "/sentinel/chat-requirements/prepare":
			_, _ = io.WriteString(w, `{"prepare_token":"prepare-1","turnstile":{"required":true,"dx":"interactive"}}`)
		case "/f/conversation":
			conversationCalls++
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	transport := newHTTPTransportForTest(t, server, "requirements-token")
	state := NewClientState("device", "ua", time.Now())
	_, err := transport.Send(context.Background(), AccountRef{ID: 7}, TransportRequest{
		ClientState: *state,
		Conversation: ConversationRequest{
			Action:          "next",
			ParentMessageID: RootParentMessageID,
			Model:           "gpt-5",
			Timezone:        "UTC",
			ConversationMode: map[string]string{"kind": "primary_assistant"},
			Messages: []ConversationMessage{{
				ID:      "user-1",
				Author:  ConversationAuthor{Role: "user"},
				Content: ConversationContent{ContentType: "text", Parts: []any{"hello"}},
			}},
		},
	})
	var upstream *UpstreamError
	if !errors.As(err, &upstream) || !upstream.ChallengeRequired {
		t.Fatalf("err = %T %v", err, err)
	}
	if conversationCalls != 0 {
		t.Fatalf("conversation calls = %d", conversationCalls)
	}
}

func newHTTPTransportForTest(t *testing.T, server *httptest.Server, requirementsToken string) Transport {
	t.Helper()
	source := CredentialSourceFunc(func(context.Context, int64) (CredentialRecord, error) {
		return NewCredentialRecord(7, AccountPlatformChatGPTWeb, AccountTypeCookie, map[string]any{
			"access_token": "secret-access-token",
			"cookie":       "session=secret-cookie",
			"account_id":   "upstream-account",
			"device_id":    "browser-device",
			"user_agent":   "browser-ua",
		}), nil
	})
	sessions, err := NewCredentialSessionProvider(source)
	if err != nil {
		t.Fatal(err)
	}
	transport, err := NewHTTPTransport(
		sessions,
		HTTPClientProviderFunc(func(context.Context, AccountRef) (*http.Client, error) { return server.Client(), nil }),
		RequirementsTokenProviderFactoryFunc(func(context.Context, AccountRef, *AccountSession) (RequirementsTokenProvider, error) {
			return RequirementsTokenProviderFunc(func(context.Context, *ClientState) (string, error) { return requirementsToken, nil }), nil
		}),
		nil,
		server.URL,
	)
	if err != nil {
		t.Fatal(err)
	}
	return transport
}
