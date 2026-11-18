package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	cangyuanMediaBaseURL        = "https://ai.cangyuansuanli.cn"
	cangyuanMediaResponseLimit  = 2 << 20
	cangyuanMediaRequestLimit   = 1 << 20
	defaultCangyuanPollInterval = 6 * time.Second
)

type CangyuanMediaKind string

const (
	CangyuanImageGeneration CangyuanMediaKind = "image_generation"
	CangyuanImageEdit       CangyuanMediaKind = "image_edit"
	CangyuanVideoGeneration CangyuanMediaKind = "video_generation"
)

var cangyuanTaskIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,256}$`)

// CangyuanMediaClient is deliberately independent of the public gateway and
// billing. A task being queued is not a chargeable completed image/video.
// Callers must implement persistent ownership and success-only settlement
// before exposing the client to user-facing routes.
type CangyuanMediaClient struct {
	APIKey string
	HTTP   *http.Client
}

type CangyuanMediaTask struct {
	ID       string `json:"id"`
	TaskID   string `json:"task_id,omitempty"`
	Status   string `json:"status"`
	Model    string `json:"model,omitempty"`
	VideoURL string `json:"video_url,omitempty"`
	Data     []struct {
		URL string `json:"url"`
	} `json:"data,omitempty"`
	Usage json.RawMessage `json:"usage,omitempty"`
}

func (k CangyuanMediaKind) path() (string, error) {
	switch k {
	case CangyuanImageGeneration:
		return "/v1/images/generations", nil
	case CangyuanImageEdit:
		return "/v1/images/edits", nil
	case CangyuanVideoGeneration:
		return "/v1/videos", nil
	default:
		return "", fmt.Errorf("unsupported cangyuan media capability")
	}
}

func (t *CangyuanMediaTask) IDValue() string {
	if t == nil {
		return ""
	}
	if strings.TrimSpace(t.ID) != "" {
		return t.ID
	}
	return t.TaskID
}

func (t *CangyuanMediaTask) ResultURLs(kind CangyuanMediaKind) []string {
	if t == nil || t.Status != "completed" {
		return nil
	}
	result := make([]string, 0, len(t.Data)+1)
	if kind == CangyuanVideoGeneration && strings.TrimSpace(t.VideoURL) != "" {
		result = append(result, t.VideoURL)
	}
	for _, item := range t.Data {
		if strings.TrimSpace(item.URL) != "" {
			result = append(result, item.URL)
		}
	}
	return result
}

// Submit starts one upstream job exactly once. Never retry an uncertain POST:
// an upstream job may already be running even if the connection was lost.
func (c *CangyuanMediaClient) Submit(ctx context.Context, kind CangyuanMediaKind, payload []byte) (*CangyuanMediaTask, error) {
	path, err := kind.path()
	if err != nil {
		return nil, err
	}
	if len(payload) == 0 || len(payload) > cangyuanMediaRequestLimit {
		return nil, errors.New("cangyuan media request has invalid size")
	}
	var request map[string]json.RawMessage
	if err := json.Unmarshal(payload, &request); err != nil || request == nil {
		return nil, errors.New("cangyuan media request must be a JSON object")
	}
	var model, prompt string
	if err := json.Unmarshal(request["model"], &model); err != nil || strings.TrimSpace(model) == "" {
		return nil, errors.New("cangyuan media request requires model")
	}
	if err := json.Unmarshal(request["prompt"], &prompt); err != nil || strings.TrimSpace(prompt) == "" {
		return nil, errors.New("cangyuan media request requires prompt")
	}
	if err := validateCangyuanMediaReferences(request); err != nil {
		return nil, err
	}
	if kind != CangyuanVideoGeneration {
		// The upstream defaults to an asynchronous media lifecycle. Force the
		// explicit contract so both synchronous and async clients can use the
		// same task engine without a second upstream generation.
		request["async"] = json.RawMessage("true")
		if _, exists := request["response_format"]; !exists {
			request["response_format"] = json.RawMessage(`"url"`)
		}
	}
	encoded, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("encode cangyuan media request: %w", err)
	}
	task, err := c.do(ctx, http.MethodPost, path, encoded)
	if err != nil {
		return nil, err
	}
	if !cangyuanTaskIDPattern.MatchString(task.IDValue()) {
		return nil, errors.New("cangyuan media submission did not return a safe task ID")
	}
	if task.Status != "queued" && task.Status != "in_progress" && task.Status != "completed" && task.Status != "failed" {
		return nil, errors.New("cangyuan media submission returned unknown task status")
	}
	if task.Status == "completed" && len(task.ResultURLs(kind)) == 0 {
		return nil, errors.New("cangyuan media submission completed without result URL")
	}
	return task, nil
}

// Poll checks the original kind-specific path, NOT a generic /tasks or
// /content endpoint. The task ID is strictly validated before it is put in
// the path, preventing a caller-controlled URL from being requested.
func (c *CangyuanMediaClient) Poll(ctx context.Context, kind CangyuanMediaKind, taskID string) (*CangyuanMediaTask, error) {
	path, err := kind.path()
	if err != nil {
		return nil, err
	}
	if !cangyuanTaskIDPattern.MatchString(taskID) {
		return nil, errors.New("invalid cangyuan media task ID")
	}
	task, err := c.do(ctx, http.MethodGet, path+"/"+url.PathEscape(taskID), nil)
	if err != nil {
		return nil, err
	}
	if task.IDValue() != taskID {
		return nil, errors.New("cangyuan media response task ID mismatch")
	}
	if task.Status != "queued" && task.Status != "in_progress" && task.Status != "completed" && task.Status != "failed" {
		return nil, errors.New("cangyuan media task returned unknown status")
	}
	if task.Status == "completed" && len(task.ResultURLs(kind)) == 0 {
		return nil, errors.New("cangyuan media task completed without result URL")
	}
	return task, nil
}

// Wait checks status without resubmitting the prompt. It requires a bounded
// context from its caller; production video requests should allow >=30 minutes.
func (c *CangyuanMediaClient) Wait(ctx context.Context, kind CangyuanMediaKind, taskID string, interval time.Duration) (*CangyuanMediaTask, error) {
	if _, err := kind.path(); err != nil {
		return nil, err
	}
	if !cangyuanTaskIDPattern.MatchString(taskID) {
		return nil, errors.New("invalid cangyuan media task ID")
	}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		return nil, errors.New("cangyuan media polling requires a deadline")
	}
	if interval <= 0 {
		interval = defaultCangyuanPollInterval
	}
	for {
		task, err := c.Poll(ctx, kind, taskID)
		if err != nil {
			return nil, err
		}
		switch task.Status {
		case "completed":
			return task, nil
		case "failed":
			return task, errors.New("cangyuan media task failed")
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

// validateCangyuanMediaReferences fails closed on inline/local/private
// reference material. The upstream expects an independently accessible HTTPS
// URL; signed upload URLs are handled by a separate upload capability.
func validateCangyuanMediaReferences(request map[string]json.RawMessage) error {
	arrayKeys := []string{"images", "reference_image_urls", "reference_videos", "reference_audios"}
	singleKeys := []string{"mask", "first_image_url", "last_image_url"}
	for _, key := range arrayKeys {
		raw, exists := request[key]
		if !exists {
			continue
		}
		var values []string
		if err := json.Unmarshal(raw, &values); err != nil || len(values) == 0 {
			return fmt.Errorf("cangyuan media %s must be a non-empty HTTPS URL array", key)
		}
		for _, value := range values {
			if !isPublicCangyuanMediaReference(value) {
				return fmt.Errorf("cangyuan media %s requires publicly accessible HTTPS references", key)
			}
		}
	}
	for _, key := range singleKeys {
		raw, exists := request[key]
		if !exists {
			continue
		}
		var value string
		if err := json.Unmarshal(raw, &value); err != nil || !isPublicCangyuanMediaReference(value) {
			return fmt.Errorf("cangyuan media %s requires a publicly accessible HTTPS reference", key)
		}
	}
	return nil
}

func isPublicCangyuanMediaReference(value string) bool {
	u, err := url.Parse(strings.TrimSpace(value))
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return false
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if host == "localhost" || strings.HasSuffix(host, ".localhost") ||
		strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return false
	}
	if addr := net.ParseIP(host); addr != nil && (!addr.IsGlobalUnicast() || addr.IsPrivate()) {
		return false
	}
	return true
}

func (c *CangyuanMediaClient) do(ctx context.Context, method, path string, payload []byte) (*CangyuanMediaTask, error) {
	if c == nil || strings.TrimSpace(c.APIKey) == "" {
		return nil, errors.New("cangyuan media API key is required")
	}
	req, err := http.NewRequestWithContext(ctx, method, cangyuanMediaBaseURL+path, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build cangyuan media request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(c.APIKey))
	req.Header.Set("Accept", "application/json")
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 30 * time.Second}
	if c.HTTP != nil {
		*client = *c.HTTP
	}
	// Never forward an API key to an unexpected redirect target.
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		// Do not embed raw HTTP client errors: proxy errors and URLs can
		// include sensitive infrastructure details.
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("cangyuan media upstream transport failed")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("cangyuan media upstream returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, cangyuanMediaResponseLimit+1))
	if err != nil || len(body) > cangyuanMediaResponseLimit {
		return nil, errors.New("cangyuan media upstream response unavailable or too large")
	}
	var task CangyuanMediaTask
	if err := json.Unmarshal(body, &task); err != nil {
		return nil, errors.New("cangyuan media upstream returned malformed JSON")
	}
	return &task, nil
}
