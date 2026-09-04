package chatgptweb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const DefaultBaseURL = "https://chatgpt.com/backend-api"
const defaultMaxErrorBody = int64(32 << 10)

type Client struct {
	httpClient   *http.Client
	baseURL      string
	identity     *Identity
	maxErrorBody int64
	now          func() time.Time
}

func NewClient(httpClient *http.Client, baseURL string, identity *Identity) (*Client, error) {
	if identity == nil {
		return nil, errors.New("chatgptweb: identity is required")
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultBaseURL
	}
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("chatgptweb: invalid base URL")
	}
	return &Client{httpClient: httpClient, baseURL: strings.TrimRight(baseURL, "/"), identity: identity, maxErrorBody: defaultMaxErrorBody, now: func() time.Time { return time.Now().UTC() }}, nil
}

func (c *Client) newJSONRequest(ctx context.Context, method, path string, payload any, state *ClientState, extra http.Header) (*http.Request, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Oai-Language", "en-US")
	req.Header.Set("Origin", "https://chatgpt.com")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	target := "/backend-api" + path
	req.Header.Set("X-Openai-Target-Path", target)
	req.Header.Set("X-Openai-Target-Route", target)
	if state != nil && state.ConversationID != "" {
		req.Header.Set("Referer", "https://chatgpt.com/c/"+state.ConversationID)
	} else {
		req.Header.Set("Referer", "https://chatgpt.com/")
	}
	c.identity.applyHeaders(req.Header, state)
	for key, values := range extra {
		req.Header.Del(key)
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return req, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, payload any, state *ClientState, extra http.Header) (*http.Response, error) {
	req, err := c.newJSONRequest(ctx, method, path, payload, state, extra)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, &UpstreamError{Kind: ErrorKindClientClosed, Message: "client request canceled"}
		}
		return nil, &UpstreamError{Kind: ErrorKindTransient, Message: "upstream transport failed", RetryNextAccount: true}
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, nil
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, c.maxErrorBody))
	return nil, ClassifyHTTPError(resp.StatusCode, body, parseRetryAfter(resp.Header.Get("Retry-After"), c.now()))
}

func parseRetryAfter(value string, now time.Time) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if at, err := http.ParseTime(value); err == nil && at.After(now) {
		return at.Sub(now)
	}
	return 0
}
