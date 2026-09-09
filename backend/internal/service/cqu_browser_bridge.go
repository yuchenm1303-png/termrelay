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
	"sync"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/coder/websocket"
)

const (
	CQUDefaultOrigin   = config.DefaultCQUBaseURL
	CQUDefaultAgentID  = config.DefaultCQUAgentID
	CQUDefaultModelID  = config.DefaultCQUModelID
	cquSendChatAPIPath = "/api/chat-web/message-chat/send-chat"
)

// CQU bridge error kinds. They are the stable vocabulary the provider adapter
// maps onto client-facing error types, so an operator can tell "the browser is
// not running" apart from "the browser is running but signed out".
const (
	CQUErrorKindBrowserUnavailable = "cqu_browser_unavailable"
	CQUErrorKindPageMissing        = "cqu_browser_page_missing"
	CQUErrorKindDisabled           = "cqu_browser_disabled"
	CQUErrorKindTransport          = "cqu_browser_transport_error"
)

// CQUBridgeError carries a bridge failure kind alongside the message. It never
// wraps credential material: the browser owns cookies, bearer tokens and
// CAqWHAeT, and none of them ever reach this process.
type CQUBridgeError struct {
	Kind    string
	Message string
	Err     error
}

func (e *CQUBridgeError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Kind, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func (e *CQUBridgeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func newCQUBridgeError(kind, message string, err error) *CQUBridgeError {
	return &CQUBridgeError{Kind: kind, Message: message, Err: err}
}

// CQUBridgeErrorKind reports the bridge failure kind of err, or "" when err did
// not originate from the bridge.
func CQUBridgeErrorKind(err error) string {
	var bridgeErr *CQUBridgeError
	if errors.As(err, &bridgeErr) && bridgeErr != nil {
		return bridgeErr.Kind
	}
	return ""
}

// CQUWebBridgeConfig describes an already-running persistent Chromium/Edge
// instance. TermRelay attaches to it through CDP; it never owns, closes or
// restarts the browser process or its user-data directory.
type CQUWebBridgeConfig struct {
	DebugURL        string
	BaseURL         string
	PageURLContains string
	ConnectTimeout  time.Duration
}

// NewCQUWebBridgeFromConfig builds a bridge from the shared application config
// (CQU_BROWSER_DEBUG_URL / CQU_BASE_URL). It returns a disabled-kind error when
// the operator has not turned the bridge on, so CQU accounts fail closed rather
// than silently attempting a plain HTTP call that could never authenticate.
func NewCQUWebBridgeFromConfig(cfg *config.Config) (*CQUWebBridge, error) {
	if cfg == nil {
		return nil, newCQUBridgeError(CQUErrorKindDisabled, "cqu browser bridge is not configured", nil)
	}
	if !cfg.CQU.Enabled {
		return nil, newCQUBridgeError(CQUErrorKindDisabled, "cqu browser bridge is disabled (set CQU_ENABLED=true)", nil)
	}
	return NewCQUWebBridge(CQUWebBridgeConfig{
		DebugURL: cfg.CQU.BrowserDebugURL,
		BaseURL:  cfg.CQU.BaseURL,
	})
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
	baseURL         string
	pageURLContains string
	connectTimeout  time.Duration
	httpClient      *http.Client
	bindingSeq      atomic.Uint64
}

func NewCQUWebBridge(cfg CQUWebBridgeConfig) (*CQUWebBridge, error) {
	debugURL := strings.TrimRight(strings.TrimSpace(cfg.DebugURL), "/")
	if debugURL == "" {
		debugURL = config.DefaultCQUBrowserDebugURL
	}
	parsed, err := url.Parse(debugURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid cqu browser debug url: %q", cfg.DebugURL)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported cqu browser debug url scheme: %s", parsed.Scheme)
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = CQUDefaultOrigin
	}
	parsedBase, err := url.Parse(baseURL)
	if err != nil || parsedBase.Scheme == "" || parsedBase.Host == "" {
		return nil, fmt.Errorf("invalid cqu base url: %q", cfg.BaseURL)
	}

	// The page is matched by host, not by full URL, so an operator who is
	// parked on /chat, /agent or the app root is all equally acceptable.
	pageURLContains := strings.TrimSpace(cfg.PageURLContains)
	if pageURLContains == "" {
		pageURLContains = parsedBase.Host
	}
	connectTimeout := cfg.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = 5 * time.Second
	}

	return &CQUWebBridge{
		debugURL:        debugURL,
		baseURL:         baseURL,
		pageURLContains: pageURLContains,
		connectTimeout:  connectTimeout,
		httpClient:      &http.Client{Timeout: connectTimeout},
	}, nil
}

// BaseURL is the CQU origin this bridge drives.
func (b *CQUWebBridge) BaseURL() string {
	if b == nil {
		return CQUDefaultOrigin
	}
	return b.baseURL
}

// Ready verifies that CDP is reachable and that a CQU page is present. It does
// not navigate or mutate the page, and it cannot verify the login state on its
// own — an expired session only surfaces as an upstream 401/403 once a request
// is actually made by the page.
func (b *CQUWebBridge) Ready(ctx context.Context) error {
	if b == nil {
		return newCQUBridgeError(CQUErrorKindDisabled, "cqu browser bridge is not configured", nil)
	}
	_, err := b.findPageTarget(ctx)
	return err
}

// SendChat executes CQU's fetch inside the selected browser page and exposes
// the resulting SSE body as a normal http.Response so the existing adapter and
// Gateway streaming lifecycle can consume it without special client writes.
func (b *CQUWebBridge) SendChat(ctx context.Context, request CQUChatRequest) (*http.Response, error) {
	if b == nil {
		return nil, newCQUBridgeError(CQUErrorKindDisabled, "cqu browser bridge is not configured", nil)
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
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, newCQUBridgeError(CQUErrorKindBrowserUnavailable, "connect cqu page cdp", err)
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
		return cquCDPTarget{}, newCQUBridgeError(CQUErrorKindTransport, "build cqu cdp target request", err)
	}
	response, err := b.httpClient.Do(request)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return cquCDPTarget{}, ctxErr
		}
		return cquCDPTarget{}, newCQUBridgeError(
			CQUErrorKindBrowserUnavailable,
			fmt.Sprintf("cannot reach the CQU browser via CDP at %s; start Edge/Chromium with --remote-debugging-port", b.debugURL),
			err,
		)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return cquCDPTarget{}, newCQUBridgeError(
			CQUErrorKindBrowserUnavailable,
			fmt.Sprintf("cqu browser cdp returned status %d", response.StatusCode),
			nil,
		)
	}

	var targets []cquCDPTarget
	decoder := json.NewDecoder(io.LimitReader(response.Body, 4<<20))
	if err := decoder.Decode(&targets); err != nil {
		return cquCDPTarget{}, newCQUBridgeError(CQUErrorKindBrowserUnavailable, "decode cqu cdp targets", err)
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
	return cquCDPTarget{}, newCQUBridgeError(
		CQUErrorKindPageMissing,
		fmt.Sprintf("no CQU browser page found via CDP; open and sign in to %s in the persistent browser", b.baseURL),
		nil,
	)
}

type cdpCommand struct {
	ID     int64  `json:"id"`
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
	defer func() { _ = conn.Close(websocket.StatusNormalClosure, "cqu request finished") }()

	// The read loop deliberately does not use the request context. Cancelling a
	// websocket read tears the connection down, and the page's fetch would then
	// keep running in the browser with no way left to reach it. Instead a
	// watcher fires the page's AbortController first, and only then stops the
	// read — so a client disconnect really does abort the CQU request.
	readCtx, stopReading := context.WithCancel(context.Background())
	defer stopReading()

	var writeMu sync.Mutex
	write := func(writeCtx context.Context, id int64, method string, params any) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		return cdpWriteCommand(writeCtx, conn, id, method, params)
	}

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

	if err := write(ctx, 3, "Runtime.evaluate", map[string]any{
		"expression":    expression,
		"awaitPromise":  false,
		"returnByValue": true,
	}); err != nil {
		finishWithError(fmt.Errorf("start cqu browser fetch: %w", err))
		return
	}

	go func() {
		select {
		case <-ctx.Done():
			abortCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_ = write(abortCtx, 99, "Runtime.evaluate", map[string]any{
				"expression": "globalThis.__termrelayCQUAbort?.()",
			})
			cancel()
			stopReading()
		case <-readCtx.Done():
		}
	}()

	for {
		messageType, data, err := conn.Read(readCtx)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				finishWithError(ctxErr)
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
