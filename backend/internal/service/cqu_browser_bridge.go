package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

const (
	CQUDefaultOrigin   = "https://ai.cqu.edu.cn"
	CQUDefaultAgentID  = "910020080887140352"
	CQUDefaultModelID  = "910023701267746816"
	cquSendChatAPIPath = "/api/chat-web/message-chat/send-chat"
)

// CQUWebBridgeConfig describes an already-running persistent Chromium/Edge
// instance. TermRelay attaches to it through CDP; it never owns, closes or
// restarts the browser process or its user-data directory.
type CQUWebBridgeConfig struct {
	DebugURL        string
	PageURLContains string
	ConnectTimeout  time.Duration
}

// CQUChatRequest is the browser-page request understood by CQU's send-chat API.
// CAqWHAeT is intentionally absent. The live CQU page/security runtime is the
// sole owner of that dynamic parameter and its matching request header.
type CQUChatRequest struct {
	Query                string   `json:"query"`
	ConversationID       string   `json:"conversation_id"`
	CustomConversationID string   `json:"custom_conversation_id"`
	ModelID              string   `json:"modelId"`
	AgentID              string   `json:"agentId"`
	HasNetwork           int      `json:"hasNetwork"`
	ContainerIDs         []string `json:"containerIds"`
	DeepThink            bool     `json:"deepThink"`
}

func (r CQUChatRequest) withDefaults() CQUChatRequest {
	if strings.TrimSpace(r.ModelID) == "" {
		r.ModelID = CQUDefaultModelID
	}
	if strings.TrimSpace(r.AgentID) == "" {
		r.AgentID = CQUDefaultAgentID
	}
	if r.ContainerIDs == nil {
		r.ContainerIDs = []string{}
	}
	return r
}

// CQUWebBridge forwards a request through an already authenticated normal CQU
// browser page. The page's own fetch/security runtime is used deliberately so
// dynamic browser/session authentication such as CAqWHAeT is never generated,
// persisted or hard-coded by TermRelay.
type CQUWebBridge struct {
	debugURL        string
	pageURLContains string
	connectTimeout  time.Duration
	httpClient      *http.Client
	bindingSeq      atomic.Uint64
}

func NewCQUWebBridge(cfg CQUWebBridgeConfig) (*CQUWebBridge, error) {
	debugURL := strings.TrimRight(strings.TrimSpace(cfg.DebugURL), "/")
	if debugURL == "" {
		debugURL = "http://127.0.0.1:9222"
	}
	parsed, err := url.Parse(debugURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid cqu browser debug url: %q", cfg.DebugURL)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported cqu browser debug url scheme: %s", parsed.Scheme)
	}

	pageURLContains := strings.TrimSpace(cfg.PageURLContains)
	if pageURLContains == "" {
		pageURLContains = "ai.cqu.edu.cn"
	}
	connectTimeout := cfg.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = 5 * time.Second
	}

	return &CQUWebBridge{
		debugURL:        debugURL,
		pageURLContains: pageURLContains,
		connectTimeout:  connectTimeout,
		httpClient:      &http.Client{Timeout: connectTimeout},
	}, nil
}

// Ready verifies that CDP is reachable and that an authenticated CQU page is
// present. It does not navigate or mutate the page.
func (b *CQUWebBridge) Ready(ctx context.Context) error {
	if b == nil {
		return errors.New("cqu browser bridge is nil")
	}
	_, err := b.findPageTarget(ctx)
	return err
}

// SendChat executes CQU's fetch inside the selected browser page and exposes
// the resulting SSE body as a normal http.Response so the existing adapter and
// Gateway streaming lifecycle can consume it without special client writes.
func (b *CQUWebBridge) SendChat(ctx context.Context, request CQUChatRequest) (*http.Response, error) {
	if b == nil {
		return nil, errors.New("cqu browser bridge is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	request = request.withDefaults()
	if strings.TrimSpace(request.Query) == "" {
		return nil, errors.New("cqu query is required")
	}

	target, err := b.findPageTarget(ctx)
	if err != nil {
		return nil, err
	}

	connectCtx, cancel := context.WithTimeout(ctx, b.connectTimeout)
	conn, _, err := websocket.Dial(connectCtx, target.WebSocketDebuggerURL, nil)
	cancel()
	if err != nil {
		return nil, fmt.Errorf("connect cqu page cdp: %w", err)
	}
	conn.SetReadLimit(16 << 20)

	binding := fmt.Sprintf("__termrelay_cqu_emit_%d", b.bindingSeq.Add(1))
	if err := cdpSetupBinding(ctx, conn, binding); err != nil {
		_ = conn.Close(websocket.StatusInternalError, "setup failed")
		return nil, err
	}

	payload, err := json.Marshal(request)
	if err != nil {
		_ = conn.Close(websocket.StatusInternalError, "marshal failed")
		return nil, fmt.Errorf("marshal cqu request: %w", err)
	}
	expression, err := cquBrowserFetchExpression(binding, payload)
	if err != nil {
		_ = conn.Close(websocket.StatusInternalError, "expression failed")
		return nil, err
	}

	reader, writer := io.Pipe()
	metaCh := make(chan cquBridgeMeta, 1)
	errCh := make(chan error, 1)
	go b.runBrowserFetch(ctx, conn, binding, expression, writer, metaCh, errCh)

	select {
	case meta := <-metaCh:
		headers := make(http.Header)
		for _, pair := range meta.Headers {
			if len(pair) == 2 {
				headers.Add(pair[0], pair[1])
			}
		}
		if headers.Get("Content-Type") == "" {
			headers.Set("Content-Type", "text/event-stream")
		}
		return &http.Response{
			StatusCode: meta.Status,
			Status:     fmt.Sprintf("%d %s", meta.Status, meta.StatusText),
			Header:     headers,
			Body:       reader,
			Request:    &http.Request{Method: http.MethodPost},
		}, nil
	case err := <-errCh:
		_ = reader.CloseWithError(err)
		return nil, err
	case <-ctx.Done():
		_ = reader.CloseWithError(ctx.Err())
		return nil, ctx.Err()
	}
}

type cquCDPTarget struct {
	ID                   string `json:"id"`
	Type                 string `json:"type"`
	URL                  string `json:"url"`
	Title                string `json:"title"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

func (b *CQUWebBridge) findPageTarget(ctx context.Context) (cquCDPTarget, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, b.debugURL+"/json/list", nil)
	if err != nil {
		return cquCDPTarget{}, fmt.Errorf("build cqu cdp target request: %w", err)
	}
	response, err := b.httpClient.Do(request)
	if err != nil {
		return cquCDPTarget{}, fmt.Errorf("reach cqu browser cdp at %s: %w", b.debugURL, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return cquCDPTarget{}, fmt.Errorf("cqu browser cdp returned status %d", response.StatusCode)
	}

	var targets []cquCDPTarget
	decoder := json.NewDecoder(io.LimitReader(response.Body, 4<<20))
	if err := decoder.Decode(&targets); err != nil {
		return cquCDPTarget{}, fmt.Errorf("decode cqu cdp targets: %w", err)
	}

	needle := strings.ToLower(b.pageURLContains)
	var fallback *cquCDPTarget
	for i := range targets {
		target := targets[i]
		if target.Type != "page" || target.WebSocketDebuggerURL == "" {
			continue
		}
		if !strings.Contains(strings.ToLower(target.URL), needle) {
			continue
		}
		if fallback == nil {
			copy := target
			fallback = &copy
		}
		if strings.Contains(strings.ToLower(target.URL), "/chat/") {
			return target, nil
		}
	}
	if fallback != nil {
		return *fallback, nil
	}
	return cquCDPTarget{}, fmt.Errorf("no CQU browser page found via CDP; open and sign in to %s in the persistent browser", CQUDefaultOrigin)
}

type cdpCommand struct {
	ID     int64 `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
}

type cdpMessage struct {
	ID     int64           `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func cdpSetupBinding(ctx context.Context, conn *websocket.Conn, binding string) error {
	if err := cdpCommandAndWait(ctx, conn, 1, "Runtime.enable", nil); err != nil {
		return fmt.Errorf("enable cqu page runtime: %w", err)
	}
	if err := cdpCommandAndWait(ctx, conn, 2, "Runtime.addBinding", map[string]any{"name": binding}); err != nil {
		return fmt.Errorf("install cqu page bridge binding: %w", err)
	}
	return nil
}

func cdpCommandAndWait(ctx context.Context, conn *websocket.Conn, id int64, method string, params any) error {
	if err := cdpWriteCommand(ctx, conn, id, method, params); err != nil {
		return err
	}
	for {
		messageType, data, err := conn.Read(ctx)
		if err != nil {
			return err
		}
		if messageType != websocket.MessageText {
			continue
		}
		var message cdpMessage
		if err := json.Unmarshal(data, &message); err != nil || message.ID != id {
			continue
		}
		if message.Error != nil {
			return fmt.Errorf("cdp %s failed (%d): %s", method, message.Error.Code, message.Error.Message)
		}
		return nil
	}
}

func cdpWriteCommand(ctx context.Context, conn *websocket.Conn, id int64, method string, params any) error {
	payload, err := json.Marshal(cdpCommand{ID: id, Method: method, Params: params})
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, payload)
}

type cquBridgeEnvelope struct {
	Kind       string     `json:"kind"`
	Status     int        `json:"status,omitempty"`
	StatusText string     `json:"statusText,omitempty"`
	Headers    [][]string `json:"headers,omitempty"`
	Data       string     `json:"data,omitempty"`
	Message    string     `json:"message,omitempty"`
}

type cquBridgeMeta struct {
	Status     int
	StatusText string
	Headers    [][]string
}

type cdpBindingCalledParams struct {
	Name    string `json:"name"`
	Payload string `json:"payload"`
}

func (b *CQUWebBridge) runBrowserFetch(
	ctx context.Context,
	conn *websocket.Conn,
	binding string,
	expression string,
	writer *io.PipeWriter,
	metaCh chan<- cquBridgeMeta,
	errCh chan<- error,
) {
	defer conn.Close(websocket.StatusNormalClosure, "cqu request finished")

	metaSent := false
	finishWithError := func(err error) {
		if err == nil {
			err = errors.New("cqu browser bridge failed")
		}
		_ = writer.CloseWithError(err)
		if !metaSent {
			select {
			case errCh <- err:
			default:
			}
		}
	}

	if err := cdpWriteCommand(ctx, conn, 3, "Runtime.evaluate", map[string]any{
		"expression":    expression,
		"awaitPromise":  false,
		"returnByValue": true,
	}); err != nil {
		finishWithError(fmt.Errorf("start cqu browser fetch: %w", err))
		return
	}

	for {
		messageType, data, err := conn.Read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				abortCtx, cancel := context.WithTimeout(context.Background(), time.Second)
				_ = cdpWriteCommand(abortCtx, conn, 99, "Runtime.evaluate", map[string]any{
					"expression": "globalThis.__termrelayCQUAbort?.()",
				})
				cancel()
				finishWithError(ctx.Err())
				return
			}
			finishWithError(fmt.Errorf("read cqu browser stream: %w", err))
			return
		}
		if messageType != websocket.MessageText {
			continue
		}

		var message cdpMessage
		if err := json.Unmarshal(data, &message); err != nil {
			continue
		}
		if message.ID == 3 && message.Error != nil {
			finishWithError(fmt.Errorf("evaluate cqu browser fetch (%d): %s", message.Error.Code, message.Error.Message))
			return
		}
		if message.Method != "Runtime.bindingCalled" {
			continue
		}

		var bindingEvent cdpBindingCalledParams
		if err := json.Unmarshal(message.Params, &bindingEvent); err != nil || bindingEvent.Name != binding {
			continue
		}
		var envelope cquBridgeEnvelope
		if err := json.Unmarshal([]byte(bindingEvent.Payload), &envelope); err != nil {
			continue
		}

		switch envelope.Kind {
		case "meta":
			if metaSent {
				continue
			}
			metaSent = true
			metaCh <- cquBridgeMeta{Status: envelope.Status, StatusText: envelope.StatusText, Headers: envelope.Headers}
		case "chunk":
			if _, err := io.WriteString(writer, envelope.Data); err != nil {
				finishWithError(err)
				return
			}
		case "done":
			_ = writer.Close()
			return
		case "error":
			message := strings.TrimSpace(envelope.Message)
			if message == "" {
				message = "CQU browser fetch failed"
			}
			finishWithError(errors.New(message))
			return
		}
	}
}

func cquBrowserFetchExpression(binding string, payload []byte) (string, error) {
	bindingJSON, err := json.Marshal(binding)
	if err != nil {
		return "", err
	}
	pathJSON, err := json.Marshal(cquSendChatAPIPath)
	if err != nil {
		return "", err
	}

	// Use window.fetch from the live application realm rather than a Go HTTP
	// client. CQU's page/security runtime owns CAqWHAeT and may wrap fetch; going
	// through that realm preserves the live browser/session behavior.
	return fmt.Sprintf(`(() => {
  const bindingName = %s;
  const endpoint = %s;
  const requestBody = %s;
  const emit = (value) => globalThis[bindingName](JSON.stringify(value));
  (async () => {
    const controller = new AbortController();
    globalThis.__termrelayCQUAbort = () => controller.abort();
    try {
      const response = await globalThis.fetch(endpoint, {
        method: "POST",
        credentials: "include",
        headers: {
          "accept": "text/event-stream",
          "content-type": "application/json"
        },
        body: JSON.stringify(requestBody),
        signal: controller.signal
      });
      emit({
        kind: "meta",
        status: response.status,
        statusText: response.statusText,
        headers: Array.from(response.headers.entries())
      });
      if (!response.body) {
        emit({kind: "error", message: "CQU response has no stream body"});
        return;
      }
      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      for (;;) {
        const result = await reader.read();
        if (result.done) break;
        const text = decoder.decode(result.value, {stream: true});
        if (text) emit({kind: "chunk", data: text});
      }
      const tail = decoder.decode();
      if (tail) emit({kind: "chunk", data: tail});
      emit({kind: "done"});
    } catch (error) {
      emit({kind: "error", message: String(error && error.message ? error.message : error)});
    } finally {
      delete globalThis.__termrelayCQUAbort;
    }
  })();
  return true;
})()`, string(bindingJSON), string(pathJSON), string(payload)), nil
}
