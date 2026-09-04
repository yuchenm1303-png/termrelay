package chatgptweb

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type PrepareState string

const (
	PrepareStateNone    PrepareState = "none"
	PrepareStateSent    PrepareState = "sent"
	PrepareStateSuccess PrepareState = "success"
)

type ConversationInitResponse struct {
	Type             string `json:"type"`
	DefaultModelSlug string `json:"default_model_slug"`
}
type prepareResponse struct {
	ConduitToken string `json:"conduit_token"`
}

func (c *Client) InitConversation(ctx context.Context, state *ClientState) (*ConversationInitResponse, error) {
	payload := map[string]any{"requested_default_model": nil, "conversation_id": nil, "timezone_offset_min": 0, "conversation_origin": nil}
	resp, err := c.doJSON(ctx, http.MethodPost, "/conversation/init", payload, state, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out ConversationInitResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) PrepareConversation(ctx context.Context, request ConversationRequest, state *ClientState, sentinel *SentinelTokens, turnTraceID string) (string, error) {
	if strings.TrimSpace(turnTraceID) == "" {
		turnTraceID = newUUID()
	}
	previous := ""
	for _, stage := range []PrepareState{PrepareStateNone, PrepareStateSent, PrepareStateSuccess} {
		token, err := c.prepareConversationStage(ctx, request, state, sentinel, stage, previous, turnTraceID)
		if err != nil {
			return "", err
		}
		previous = token
	}
	if previous == "" {
		return "", errors.New("chatgptweb: conversation prepare returned no conduit token")
	}
	return previous, nil
}

func (c *Client) prepareConversationStage(ctx context.Context, request ConversationRequest, state *ClientState, sentinel *SentinelTokens, stage PrepareState, previous, turnTraceID string) (string, error) {
	parent := request.ParentMessageID
	if parent == "" {
		parent = RootParentMessageID
	}
	payload := map[string]any{"action": "next", "parent_message_id": parent, "model": modelOrAuto(request.Model), "client_prepare_state": string(stage), "client_prepare_dispatch": "debounced", "client_prepare_source": "composer_editor_state", "timezone_offset_min": request.TimezoneOffsetMin, "timezone": request.Timezone, "conversation_mode": map[string]string{"kind": "primary_assistant"}, "system_hints": request.SystemHints, "supports_buffering": true, "supported_encodings": []string{"v1"}, "client_contextual_info": prepareClientContext(request, state, c.now()), "local_function_names": []string{"local.continue_in_work"}}
	if request.ConversationID != "" {
		payload["conversation_id"] = request.ConversationID
	}
	if stage == PrepareStateSent || stage == PrepareStateSuccess {
		payload["partial_query"] = map[string]any{"id": newUUID(), "author": map[string]string{"role": "user"}, "content": map[string]any{"content_type": "text", "parts": []string{partialUserText(request)}}}
	}
	headers := conversationHeaders(sentinel, state, "*/*", previous, turnTraceID)
	resp, err := c.doJSON(ctx, http.MethodPost, "/f/conversation/prepare", payload, state, headers)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result prepareResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if strings.TrimSpace(result.ConduitToken) == "" {
		return "", &UpstreamError{Kind: ErrorKindProtocol, Message: "conversation prepare conduit token is missing"}
	}
	return result.ConduitToken, nil
}

func (c *Client) OpenConversation(ctx context.Context, request ConversationRequest, state *ClientState, sentinel *SentinelTokens, conduitToken, turnTraceID string) (*http.Response, error) {
	if strings.TrimSpace(conduitToken) == "" {
		return nil, errors.New("chatgptweb: conduit token is required")
	}
	if strings.TrimSpace(turnTraceID) == "" {
		turnTraceID = newUUID()
	}
	if state != nil {
		if request.ParentMessageID == "" {
			request.ParentMessageID = state.ParentMessageID
		}
		if request.ConversationID == "" {
			request.ConversationID = state.ConversationID
		}
		request.ClientPrepareState = "success"
	}
	request.ClientContextualInfo = nil
	headers := conversationHeaders(sentinel, state, "text/event-stream", conduitToken, turnTraceID)
	headers.Set("Oai-Echo-Logs", "0")
	headers.Set("Oai-Telemetry", "[1,null]")
	return c.doJSON(ctx, http.MethodPost, "/f/conversation", request, state, headers)
}

func conversationHeaders(sentinel *SentinelTokens, state *ClientState, accept, conduitToken, turnTraceID string) http.Header {
	h := make(http.Header)
	h.Set("Accept", accept)
	h.Set("X-Conduit-Token", conduitToken)
	if turnTraceID != "" {
		h.Set("X-Oai-Turn-Trace-Id", turnTraceID)
	}
	if sentinel != nil {
		if sentinel.PrepareToken != "" {
			h.Set("Openai-Sentinel-Chat-Requirements-Prepare-Token", sentinel.PrepareToken)
		}
		if sentinel.ChatRequirementsToken != "" {
			h.Set("Openai-Sentinel-Chat-Requirements-Token", sentinel.ChatRequirementsToken)
		}
		if sentinel.ProofToken != "" {
			h.Set("Openai-Sentinel-Proof-Token", sentinel.ProofToken)
		}
	}
	return h
}
func modelOrAuto(model string) string {
	if strings.TrimSpace(model) == "" {
		return "auto"
	}
	return model
}
func partialUserText(request ConversationRequest) string {
	for i := len(request.Messages) - 1; i >= 0; i-- {
		if request.Messages[i].Author.Role != "user" {
			continue
		}
		for _, part := range request.Messages[i].Content.Parts {
			if s, ok := part.(string); ok && strings.TrimSpace(s) != "" {
				r := []rune(s)
				if len(r) > 5 {
					r = r[:5]
				}
				return string(r)
			}
		}
	}
	return "h"
}
func prepareClientContext(request ConversationRequest, state *ClientState, now time.Time) map[string]any {
	out := map[string]any{"app_name": "chatgpt.com"}
	for k, v := range request.ClientContextualInfo {
		out[k] = v
	}
	if state != nil {
		out["time_since_loaded"] = int(state.TimeSinceLoaded(now).Round(time.Second).Seconds())
	}
	return out
}
