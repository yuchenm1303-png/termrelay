package chatgptweb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type OpenAIChatRequest struct {
	Model    string              `json:"model"`
	Messages []OpenAIChatMessage `json:"messages"`
	Stream   bool                `json:"stream,omitempty"`
}
type OpenAIChatMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}
type ConversationAuthor struct {
	Role string `json:"role"`
}
type ConversationContent struct {
	ContentType string `json:"content_type"`
	Parts       []any  `json:"parts"`
}
type ConversationMessage struct {
	ID         string              `json:"id"`
	Author     ConversationAuthor  `json:"author"`
	CreateTime float64             `json:"create_time,omitempty"`
	Content    ConversationContent `json:"content"`
	Metadata   map[string]any      `json:"metadata,omitempty"`
}
type ConversationRequest struct {
	Action                           string                `json:"action"`
	Messages                         []ConversationMessage `json:"messages"`
	ParentMessageID                  string                `json:"parent_message_id"`
	ConversationID                   string                `json:"conversation_id,omitempty"`
	Model                            string                `json:"model"`
	ClientPrepareState               string                `json:"client_prepare_state,omitempty"`
	TimezoneOffsetMin                int                   `json:"timezone_offset_min"`
	Timezone                         string                `json:"timezone"`
	ConversationMode                 map[string]string     `json:"conversation_mode"`
	SystemHints                      []string              `json:"system_hints"`
	HistoryAndTrainingDisabled       bool                  `json:"history_and_training_disabled"`
	ParagenCotSummaryDisplayOverride string                `json:"paragen_cot_summary_display_override"`
	ForceParallelSwitch              string                `json:"force_parallel_switch"`
	ThinkingEffort                   string                `json:"thinking_effort"`
	ClientContextualInfo             map[string]any        `json:"client_contextual_info,omitempty"`
}

func TransformChatCompletions(body []byte, state *ClientState, now time.Time) (ConversationRequest, bool, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	var input OpenAIChatRequest
	if err := dec.Decode(&input); err != nil {
		return ConversationRequest{}, false, fmt.Errorf("chatgptweb: decode chat request: %w", err)
	}
	input.Model = strings.TrimSpace(input.Model)
	if input.Model == "" {
		return ConversationRequest{}, false, errors.New("chatgptweb: model is required")
	}
	if len(input.Messages) == 0 {
		return ConversationRequest{}, false, errors.New("chatgptweb: messages are required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	request := ConversationRequest{Action: "next", ParentMessageID: RootParentMessageID, Model: input.Model, TimezoneOffsetMin: 0, Timezone: "UTC", ConversationMode: map[string]string{"kind": "primary_assistant"}, SystemHints: []string{}, HistoryAndTrainingDisabled: true, ParagenCotSummaryDisplayOverride: "allow", ForceParallelSwitch: "auto", ThinkingEffort: "standard"}
	if state != nil {
		if state.ParentMessageID != "" {
			request.ParentMessageID = state.ParentMessageID
		}
		request.ConversationID = state.ConversationID
		request.ClientPrepareState = "success"
		request.ClientContextualInfo = map[string]any{"app_name": "chatgpt.com", "time_since_loaded": int(state.TimeSinceLoaded(now).Round(time.Second).Seconds())}
	}
	for _, msg := range input.Messages {
		role := strings.ToLower(strings.TrimSpace(msg.Role))
		switch role {
		case "system", "developer", "user", "assistant":
		default:
			return ConversationRequest{}, false, fmt.Errorf("chatgptweb: unsupported message role %q", msg.Role)
		}
		text, err := pureTextContent(msg.Content)
		if err != nil {
			return ConversationRequest{}, false, err
		}
		request.Messages = append(request.Messages, ConversationMessage{ID: newUUID(), Author: ConversationAuthor{Role: role}, CreateTime: float64(now.UnixMilli()) / 1000, Content: ConversationContent{ContentType: "text", Parts: []any{text}}, Metadata: defaultMessageMetadata()})
	}
	if request.ConversationID != "" {
		// A continued web conversation already owns the prior turns upstream. Sending the
		// complete OpenAI history again would duplicate them, so only the new tail turn is sent.
		idx := -1
		for i := len(request.Messages) - 1; i >= 0; i-- {
			if request.Messages[i].Author.Role == "user" {
				idx = i
				break
			}
		}
		if idx < 0 {
			return ConversationRequest{}, false, errors.New("chatgptweb: continued conversation requires a user message")
		}
		request.Messages = []ConversationMessage{request.Messages[idx]}
	}
	return request, input.Stream, nil
}

func pureTextContent(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", errors.New("chatgptweb: message content is required")
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err != nil {
		return "", errors.New("chatgptweb: only text message content is supported")
	}
	var b strings.Builder
	for _, part := range parts {
		switch part.Type {
		case "text", "input_text":
			_, _ = b.WriteString(part.Text)
		default:
			return "", fmt.Errorf("chatgptweb: unsupported content part %q", part.Type)
		}
	}
	return b.String(), nil
}
func defaultMessageMetadata() map[string]any {
	return map[string]any{"developer_mode_connector_ids": []any{}, "selected_sources": []any{}, "selected_github_repos": []any{}, "selected_all_github_repos": false, "serialization_metadata": map[string]any{"custom_symbol_offsets": []any{}}}
}
