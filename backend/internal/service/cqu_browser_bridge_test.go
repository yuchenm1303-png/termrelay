package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/coder/websocket"
	"github.com/stretchr/testify/require"
)

// fakeCDP is a minimal Chrome DevTools endpoint: a /json/list target listing
// plus a websocket per page that records the commands the bridge issues.
type fakeCDP struct {
	server   *httptest.Server
	commands chan cdpCommand
	pageURL  string
}

func newFakeCDP(t *testing.T, pageURL string) *fakeCDP {
	t.Helper()
	fake := &fakeCDP{commands: make(chan cdpCommand, 16), pageURL: pageURL}

	mux := http.NewServeMux()
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		targets := []cquCDPTarget{}
		if fake.pageURL != "" {
			targets = append(targets, cquCDPTarget{
				ID:                   "page-1",
				Type:                 "page",
				URL:                  fake.pageURL,
				Title:                "CQU",
				WebSocketDebuggerURL: "ws" + strings.TrimPrefix(fake.server.URL, "http") + "/devtools/page/page-1",
			})
		}
		targets = append(targets, cquCDPTarget{ID: "other", Type: "page", URL: "https://example.com/", WebSocketDebuggerURL: "ws://unused"})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(targets)
	})
	mux.HandleFunc("/devtools/page/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close(websocket.StatusNormalClosure, "done") }()

		ctx := r.Context()
		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				return
			}
			var command cdpCommand
			if json.Unmarshal(data, &command) != nil {
				continue
			}
			select {
			case fake.commands <- command:
			default:
			}
			// Acknowledge setup commands; never produce a bindingCalled event so
			// the request stays in flight until the test cancels it.
			if command.Method == "Runtime.enable" || command.Method == "Runtime.addBinding" {
				ack, _ := json.Marshal(map[string]any{"id": command.ID, "result": map[string]any{}})
				if conn.Write(ctx, websocket.MessageText, ack) != nil {
					return
				}
			}
		}
	})

	fake.server = httptest.NewServer(mux)
	t.Cleanup(fake.server.Close)
	return fake
}

func TestCQUWebBridgeReadyReportsBrowserUnavailable(t *testing.T) {
	// Port 1 is reserved and never listening, so this is a connect failure
	// rather than a protocol failure.
	bridge, err := NewCQUWebBridge(CQUWebBridgeConfig{
		DebugURL:       "http://127.0.0.1:1",
		ConnectTimeout: time.Second,
	})
	require.NoError(t, err)

	err = bridge.Ready(context.Background())
	require.Error(t, err)
	require.Equal(t, CQUErrorKindBrowserUnavailable, CQUBridgeErrorKind(err))
	require.NotContains(t, strings.ToLower(err.Error()), "cookie")
	require.NotContains(t, strings.ToLower(err.Error()), "caqwhaet")
}

func TestCQUWebBridgeReadyReportsPageMissing(t *testing.T) {
	fake := newFakeCDP(t, "")
	bridge, err := NewCQUWebBridge(CQUWebBridgeConfig{DebugURL: fake.server.URL})
	require.NoError(t, err)

	err = bridge.Ready(context.Background())
	require.Error(t, err)
	require.Equal(t, CQUErrorKindPageMissing, CQUBridgeErrorKind(err))
	require.Contains(t, err.Error(), config.DefaultCQUBaseURL)
}

func TestCQUWebBridgeReadyFindsSignedInPage(t *testing.T) {
	fake := newFakeCDP(t, "https://ai.cqu.edu.cn/chat/123")
	bridge, err := NewCQUWebBridge(CQUWebBridgeConfig{DebugURL: fake.server.URL})
	require.NoError(t, err)

	require.NoError(t, bridge.Ready(context.Background()))
	require.Equal(t, config.DefaultCQUBaseURL, bridge.BaseURL())
}

// A client disconnect must travel all the way to the page: the bridge stops
// waiting and asks the page's AbortController to cancel the in-flight fetch.
func TestCQUWebBridgeSendChatAbortsPageFetchOnCancellation(t *testing.T) {
	fake := newFakeCDP(t, "https://ai.cqu.edu.cn/chat/123")
	bridge, err := NewCQUWebBridge(CQUWebBridgeConfig{DebugURL: fake.server.URL})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, sendErr := bridge.SendChat(ctx, CQUChatRequest{Query: "你好"})
		done <- sendErr
	}()

	// Wait for the page fetch to have been started before cancelling.
	require.Eventually(t, func() bool {
		for {
			select {
			case command := <-fake.commands:
				if command.Method == "Runtime.evaluate" {
					return true
				}
			default:
				return false
			}
		}
	}, 5*time.Second, 10*time.Millisecond)

	cancel()

	select {
	case sendErr := <-done:
		require.ErrorIs(t, sendErr, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("SendChat did not return after cancellation")
	}

	require.Eventually(t, func() bool {
		select {
		case command := <-fake.commands:
			params, _ := json.Marshal(command.Params)
			return strings.Contains(string(params), "__termrelayCQUAbort")
		default:
			return false
		}
	}, 5*time.Second, 10*time.Millisecond, "page fetch was never aborted")
}

func TestCQUWebBridgeSendChatRequiresQuery(t *testing.T) {
	fake := newFakeCDP(t, "https://ai.cqu.edu.cn/chat/1")
	bridge, err := NewCQUWebBridge(CQUWebBridgeConfig{DebugURL: fake.server.URL})
	require.NoError(t, err)

	_, err = bridge.SendChat(context.Background(), CQUChatRequest{})
	require.Error(t, err)
}

// The browser-page expression must carry the CQU body and the abort hook, and
// must never carry credentials — the page supplies those itself.
func TestCQUBrowserFetchExpressionCarriesNoCredentials(t *testing.T) {
	payload, err := json.Marshal(CQUChatRequest{
		Query:        "你好",
		ModelID:      config.DefaultCQUModelID,
		AgentID:      config.DefaultCQUAgentID,
		ContainerIDs: []string{},
	})
	require.NoError(t, err)

	expression, err := cquBrowserFetchExpression("__binding", payload)
	require.NoError(t, err)
	require.Contains(t, expression, cquSendChatAPIPath)
	require.Contains(t, expression, "credentials: \"include\"")
	require.Contains(t, expression, "__termrelayCQUAbort")
	require.Contains(t, expression, "AbortController")
	require.NotContains(t, expression, "CAqWHAeT")
	require.NotContains(t, strings.ToLower(expression), "cookie")
	require.NotContains(t, strings.ToLower(expression), "authorization")
}

func TestNewCQUWebBridgeFromConfigFailsClosedWhenDisabled(t *testing.T) {
	_, err := NewCQUWebBridgeFromConfig(nil)
	require.Equal(t, CQUErrorKindDisabled, CQUBridgeErrorKind(err))

	_, err = NewCQUWebBridgeFromConfig(&config.Config{CQU: config.CQUConfig{Enabled: false}})
	require.Equal(t, CQUErrorKindDisabled, CQUBridgeErrorKind(err))

	bridge, err := NewCQUWebBridgeFromConfig(&config.Config{CQU: config.CQUConfig{
		Enabled:         true,
		BrowserDebugURL: config.DefaultCQUBrowserDebugURL,
		BaseURL:         config.DefaultCQUBaseURL,
	}})
	require.NoError(t, err)
	require.Equal(t, config.DefaultCQUBaseURL, bridge.BaseURL())
}

func TestNewCQUWebBridgeRejectsInvalidURLs(t *testing.T) {
	_, err := NewCQUWebBridge(CQUWebBridgeConfig{DebugURL: "not-a-url"})
	require.Error(t, err)

	_, err = NewCQUWebBridge(CQUWebBridgeConfig{DebugURL: "ftp://127.0.0.1:9222"})
	require.Error(t, err)

	_, err = NewCQUWebBridge(CQUWebBridgeConfig{BaseURL: "not-a-url"})
	require.Error(t, err)
}
