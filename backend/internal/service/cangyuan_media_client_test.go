package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type cangyuanMediaRoundTripFunc func(*http.Request) (*http.Response, error)

func (f cangyuanMediaRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func cangyuanMediaTestHTTPClient(fn cangyuanMediaRoundTripFunc) *http.Client {
	return &http.Client{Transport: fn}
}

func cangyuanMediaTestResponse(status int, content string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(content)),
	}
}

func TestCangyuanMediaSubmitImageForcesAsyncAndDoesNotRepeatPost(t *testing.T) {
	var posts atomic.Int32
	client := &CangyuanMediaClient{
		APIKey: "test-only-secret",
		HTTP: cangyuanMediaTestHTTPClient(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodPost || r.URL.Host != "ai.cangyuansuanli.cn" || r.URL.Path != "/v1/images/generations" {
				t.Errorf("unexpected media submit %s %s", r.Method, r.URL.String())
			}
			if r.Header.Get("Authorization") != "Bearer test-only-secret" {
				t.Error("missing upstream authorization")
			}
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"async":true`) || !strings.Contains(string(body), `"response_format":"url"`) {
				t.Errorf("missing required async image contract: %s", body)
			}
			posts.Add(1)
			return cangyuanMediaTestResponse(200, `{"id":"task_image_1","status":"queued"}`), nil
		}),
	}
	task, err := client.Submit(context.Background(), CangyuanImageGeneration, []byte(`{"model":"img","prompt":"orange cat","async":false}`))
	if err != nil {
		t.Fatal(err)
	}
	if task.IDValue() != "task_image_1" || posts.Load() != 1 {
		t.Fatalf("unexpected task/POST count: %+v %d", task, posts.Load())
	}
}

func TestCangyuanMediaPollUsesEndpointSpecificPathAndResultURL(t *testing.T) {
	for _, tc := range []struct {
		kind CangyuanMediaKind
		path string
		body string
		url  string
	}{
		{CangyuanImageGeneration, "/v1/images/generations/task_1", `{"id":"task_1","status":"completed","data":[{"url":"https://images.example.com/a.png"}]}`, "https://images.example.com/a.png"},
		{CangyuanImageEdit, "/v1/images/edits/task_1", `{"id":"task_1","status":"completed","data":[{"url":"https://images.example.com/b.png"}]}`, "https://images.example.com/b.png"},
		{CangyuanVideoGeneration, "/v1/videos/task_1", `{"id":"task_1","status":"completed","video_url":"https://videos.example.com/a.mp4"}`, "https://videos.example.com/a.mp4"},
		{CangyuanVideoGeneration, "/v1/videos/task_1", `{"id":"task_1","status":"completed","data":[{"url":"https://videos.example.com/b.mp4"}]}`, "https://videos.example.com/b.mp4"},
	} {
		t.Run(string(tc.kind)+" "+tc.url, func(t *testing.T) {
			client := &CangyuanMediaClient{
				APIKey: "test-key",
				HTTP: cangyuanMediaTestHTTPClient(func(r *http.Request) (*http.Response, error) {
					if r.Method != http.MethodGet || r.URL.Path != tc.path {
						t.Fatalf("unexpected poll endpoint: %s %s", r.Method, r.URL.Path)
					}
					return cangyuanMediaTestResponse(http.StatusOK, tc.body), nil
				}),
			}
			task, err := client.Poll(context.Background(), tc.kind, "task_1")
			if err != nil {
				t.Fatal(err)
			}
			if len(task.ResultURLs(tc.kind)) < 1 || task.ResultURLs(tc.kind)[0] != tc.url {
				t.Fatalf("unexpected completed result: %#v", task)
			}
		})
	}
}

func TestCangyuanMediaWaitOnlyPollsAndHandlesFailure(t *testing.T) {
	var polls atomic.Int32
	client := &CangyuanMediaClient{
		APIKey: "test-key",
		HTTP: cangyuanMediaTestHTTPClient(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodGet {
				t.Fatal("poll loop must never repeat the upstream POST")
			}
			if polls.Add(1) == 1 {
				return cangyuanMediaTestResponse(200, `{"id":"task_2","status":"in_progress"}`), nil
			}
			return cangyuanMediaTestResponse(200, `{"id":"task_2","status":"completed","data":[{"url":"https://images.example.com/2.png"}]}`), nil
		}),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	task, err := client.Wait(ctx, CangyuanImageGeneration, "task_2", time.Millisecond)
	if err != nil || task.Status != "completed" || polls.Load() != 2 {
		t.Fatalf("unexpected wait outcome: task=%+v err=%v polls=%d", task, err, polls.Load())
	}

	failed := &CangyuanMediaClient{
		APIKey: "test-key",
		HTTP: cangyuanMediaTestHTTPClient(func(_ *http.Request) (*http.Response, error) {
			return cangyuanMediaTestResponse(200, `{"id":"task_failed","status":"failed"}`), nil
		}),
	}
	_, err = failed.Wait(ctx, CangyuanVideoGeneration, "task_failed", time.Millisecond)
	if err == nil {
		t.Error("a failed video must not be recorded as a successful generation")
	}
}

func TestCangyuanMediaFailClosedOnUnsafeIdsAndIncompleteTasks(t *testing.T) {
	var calls atomic.Int32
	client := &CangyuanMediaClient{
		APIKey: "test-key",
		HTTP: cangyuanMediaTestHTTPClient(func(_ *http.Request) (*http.Response, error) {
			calls.Add(1)
			return cangyuanMediaTestResponse(200, `{"id":"task_ok","status":"completed"}`), nil
		}),
	}
	for _, id := range []string{"../admin", "", "a/b", "https://localhost", "task?x=1"} {
		if _, err := client.Poll(context.Background(), CangyuanImageGeneration, id); err == nil {
			t.Errorf("unsafe task ID was accepted: %q", id)
		}
	}
	if calls.Load() != 0 {
		t.Error("unsafe task IDs must fail before network access")
	}
	if _, err := client.Poll(context.Background(), CangyuanImageGeneration, "task_ok"); err == nil {
		t.Error("completed task with no image URL must fail closed")
	}
	if calls.Load() != 1 {
		t.Errorf("expected one real GET, got %d", calls.Load())
	}
}

func TestCangyuanMediaDoesNotFollowRedirectWithCredential(t *testing.T) {
	var calls atomic.Int32
	client := &CangyuanMediaClient{
		APIKey: "test-key",
		HTTP: cangyuanMediaTestHTTPClient(func(_ *http.Request) (*http.Response, error) {
			calls.Add(1)
			r := cangyuanMediaTestResponse(http.StatusFound, "")
			r.Header.Set("Location", "https://redirect.example.com/capture-key")
			return r, nil
		}),
	}
	_, err := client.Submit(context.Background(), CangyuanVideoGeneration, []byte(`{"model":"video","prompt":"scene"}`))
	if err == nil || calls.Load() != 1 {
		t.Fatalf("redirect must fail before forwarding credential: %v, requests=%d", err, calls.Load())
	}
}

func TestCangyuanMediaRejectsMalformedRequestsAndTransportErrors(t *testing.T) {
	var networkCalls atomic.Int32
	client := &CangyuanMediaClient{
		APIKey: "test-key",
		HTTP: cangyuanMediaTestHTTPClient(func(_ *http.Request) (*http.Response, error) {
			networkCalls.Add(1)
			return nil, errors.New("secret-upstream-proxy-URL")
		}),
	}
	for _, raw := range []string{`{}`, `[]`, `{"model":"img"}`, `{"model":12,"prompt":"cat"}`, `{"model":"img","prompt":null}`} {
		if _, err := client.Submit(context.Background(), CangyuanImageEdit, []byte(raw)); err == nil {
			t.Errorf("invalid payload accepted: %q", raw)
		}
	}
	if networkCalls.Load() != 0 {
		t.Fatalf("malformed requests must not reach upstream: %d", networkCalls.Load())
	}
	_, err := client.Submit(context.Background(), CangyuanImageEdit, []byte(`{"model":"img","prompt":"cat"}`))
	if err == nil || strings.Contains(err.Error(), "secret-upstream") {
		t.Fatalf("unsafe transport error: %v", err)
	}
}

func TestCangyuanMediaReferencesRejectLocalAndInlineContent(t *testing.T) {
	t.Parallel()
	var networkCalls atomic.Int32
	client := &CangyuanMediaClient{
		APIKey: "test-key",
		HTTP: cangyuanMediaTestHTTPClient(func(_ *http.Request) (*http.Response, error) {
			networkCalls.Add(1)
			return cangyuanMediaTestResponse(http.StatusOK, `{"id":"task_1","status":"queued"}`), nil
		}),
	}
	for _, invalid := range []string{
		`{"model":"img","prompt":"scene","images":["data:image/png;base64,AA=="]}`,
		`{"model":"img","prompt":"scene","mask":"http://example.com/mask.png"}`,
		`{"model":"img","prompt":"scene","images":["https://127.0.0.1/photo.png"]}`,
		`{"model":"img","prompt":"scene","images":["https://localhost/photo.png"]}`,
		`{"model":"img","prompt":"scene","images":["https://10.0.0.2/private.png"]}`,
		`{"model":"vid","prompt":"scene","reference_image_urls":["file:///tmp/pic.png"]}`,
		`{"model":"vid","prompt":"scene","reference_videos":["https://media.internal/clip.mp4"]}`,
		`{"model":"vid","prompt":"scene","first_image_url":"https://user:pw@example.com/pic.png"}`,
	} {
		if _, err := client.Submit(context.Background(), CangyuanImageEdit, []byte(invalid)); err == nil {
			t.Errorf("unsafe reference unexpectedly accepted: %s", invalid)
		}
	}
	if networkCalls.Load() != 0 {
		t.Fatalf("unsafe references must never be submitted: %d", networkCalls.Load())
	}
	_, err := client.Submit(context.Background(), CangyuanImageEdit, []byte(
		`{"model":"img","prompt":"scene","images":["https://cdn.example.com/approved.png?signature=test"]}`,
	))
	if err != nil || networkCalls.Load() != 1 {
		t.Fatalf("safe signed HTTPS reference should pass validation: err=%v calls=%d", err, networkCalls.Load())
	}
}
