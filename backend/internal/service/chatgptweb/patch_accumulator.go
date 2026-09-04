package chatgptweb

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Snapshot struct {
	ConversationID  string
	MessageID       string
	ParentMessageID string
	Channel         string
	FinalText       string
	CommentaryText  string
	Finished        bool
	FinishReason    string
}

type PatchAccumulator struct {
	document any
	snapshot Snapshot
}

func NewPatchAccumulator() *PatchAccumulator {
	return &PatchAccumulator{}
}

func (a *PatchAccumulator) Snapshot() Snapshot {
	if a == nil {
		return Snapshot{}
	}
	return a.snapshot
}

func (a *PatchAccumulator) ApplyJSON(data []byte) (Snapshot, error) {
	if a == nil {
		return Snapshot{}, errors.New("chatgptweb: nil patch accumulator")
	}
	var value any
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber()
	if err := dec.Decode(&value); err != nil {
		return a.snapshot, fmt.Errorf("chatgptweb: decode stream frame: %w", err)
	}
	if err := a.applyValue(value); err != nil {
		return a.snapshot, err
	}
	a.refreshSnapshot()
	return a.snapshot, nil
}

func (a *PatchAccumulator) applyValue(value any) error {
	switch v := value.(type) {
	case []any:
		for _, item := range v {
			if err := a.applyValue(item); err != nil {
				return err
			}
		}
		return nil
	case map[string]any:
		if data, ok := v["data"]; ok && len(v) <= 3 {
			if _, hasType := v["type"]; hasType {
				return a.applyValue(data)
			}
		}
		if patches, ok := v["patches"].([]any); ok {
			for _, patch := range patches {
				if err := a.applyValue(patch); err != nil {
					return err
				}
			}
			a.captureEnvelope(v)
			return nil
		}
		if _, hasPath := v["p"]; hasPath {
			if _, hasOp := v["o"]; hasOp {
				return a.applyPatch(v)
			}
		}
		if message, ok := v["message"].(map[string]any); ok {
			a.document = map[string]any{"message": deepCopyMap(message)}
			a.captureEnvelope(v)
			return nil
		}
		if _, hasContent := v["content"]; hasContent {
			if _, hasAuthor := v["author"]; hasAuthor {
				a.document = map[string]any{"message": deepCopyMap(v)}
				return nil
			}
		}
		a.captureEnvelope(v)
		return nil
	default:
		return fmt.Errorf("chatgptweb: unsupported stream frame %T", value)
	}
}

func (a *PatchAccumulator) captureEnvelope(v map[string]any) {
	root, _ := a.document.(map[string]any)
	if root == nil {
		root = map[string]any{}
		a.document = root
	}
	for _, key := range []string{"conversation_id", "parent_message_id", "finish_reason", "finish_details", "finished"} {
		if value, ok := v[key]; ok {
			root[key] = value
		}
	}
}

func (a *PatchAccumulator) applyPatch(patch map[string]any) error {
	path, _ := patch["p"].(string)
	op, _ := patch["o"].(string)
	if path == "" || op == "" {
		return errors.New("chatgptweb: malformed patch")
	}
	if a.document == nil {
		a.document = map[string]any{}
	}
	var err error
	switch strings.ToLower(op) {
	case "add", "replace", "set":
		a.document, err = setAtPointer(a.document, path, patch["v"], false)
	case "append":
		a.document, err = setAtPointer(a.document, path, patch["v"], true)
	case "remove":
		a.document, err = removeAtPointer(a.document, path)
	default:
		return fmt.Errorf("chatgptweb: unsupported patch operation %q", op)
	}
	if err != nil {
		return fmt.Errorf("chatgptweb: apply patch %s %s: %w", op, path, err)
	}
	return nil
}

func (a *PatchAccumulator) refreshSnapshot() {
	root, _ := a.document.(map[string]any)
	if root == nil {
		return
	}
	if s := stringValue(root["conversation_id"]); s != "" {
		a.snapshot.ConversationID = s
	}
	if s := stringValue(root["parent_message_id"]); s != "" {
		a.snapshot.ParentMessageID = s
	}
	if b, ok := root["finished"].(bool); ok {
		a.snapshot.Finished = b
	}
	if s := finishReason(root); s != "" {
		a.snapshot.FinishReason = s
		a.snapshot.Finished = true
	}

	message, _ := root["message"].(map[string]any)
	if message == nil {
		return
	}
	if s := stringValue(message["id"]); s != "" {
		a.snapshot.MessageID = s
		a.snapshot.ParentMessageID = s
	}
	if s := stringValue(message["parent_message_id"]); s != "" && a.snapshot.ParentMessageID == "" {
		a.snapshot.ParentMessageID = s
	}
	channel := messageChannel(message)
	if channel != "" {
		a.snapshot.Channel = channel
	}
	text := messageText(message)
	if text == "" {
		return
	}
	if channel == "analysis" || channel == "commentary" {
		a.snapshot.CommentaryText = text
	} else {
		a.snapshot.FinalText = text
	}
}

func finishReason(root map[string]any) string {
	if s := stringValue(root["finish_reason"]); s != "" {
		return s
	}
	if details, ok := root["finish_details"].(map[string]any); ok {
		if s := stringValue(details["type"]); s != "" {
			return s
		}
		if s := stringValue(details["reason"]); s != "" {
			return s
		}
	}
	if message, ok := root["message"].(map[string]any); ok {
		if metadata, ok := message["metadata"].(map[string]any); ok {
			if details, ok := metadata["finish_details"].(map[string]any); ok {
				if s := stringValue(details["type"]); s != "" {
					return s
				}
			}
		}
	}
	return ""
}

func messageChannel(message map[string]any) string {
	for _, candidate := range []any{message["channel"], message["recipient"]} {
		if s := strings.ToLower(stringValue(candidate)); s != "" {
			if strings.Contains(s, "analysis") || strings.Contains(s, "commentary") {
				if strings.Contains(s, "commentary") {
					return "commentary"
				}
				return "analysis"
			}
			if strings.Contains(s, "final") {
				return "final"
			}
		}
	}
	if metadata, ok := message["metadata"].(map[string]any); ok {
		for _, key := range []string{"channel", "message_type"} {
			if s := strings.ToLower(stringValue(metadata[key])); strings.Contains(s, "analysis") || strings.Contains(s, "commentary") {
				return "analysis"
			}
		}
	}
	if content, ok := message["content"].(map[string]any); ok {
		contentType := strings.ToLower(stringValue(content["content_type"]))
		if strings.Contains(contentType, "thought") || strings.Contains(contentType, "reasoning") {
			return "analysis"
		}
	}
	return "final"
}

func messageText(message map[string]any) string {
	content, _ := message["content"].(map[string]any)
	if content == nil {
		return ""
	}
	parts, _ := content["parts"].([]any)
	if len(parts) == 0 {
		return stringValue(content["text"])
	}
	var b strings.Builder
	for _, part := range parts {
		switch p := part.(type) {
		case string:
			_, _ = b.WriteString(p)
		case map[string]any:
			if text := stringValue(p["text"]); text != "" {
				_, _ = b.WriteString(text)
			}
		}
	}
	return b.String()
}

func stringValue(v any) string {
	s, _ := v.(string)
	return s
}

func deepCopyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	buf, _ := json.Marshal(src)
	var dst map[string]any
	_ = json.Unmarshal(buf, &dst)
	return dst
}

func decodePointer(path string) ([]string, error) {
	if path == "" {
		return nil, nil
	}
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("invalid JSON pointer %q", path)
	}
	parts := strings.Split(path[1:], "/")
	for i := range parts {
		parts[i] = strings.ReplaceAll(strings.ReplaceAll(parts[i], "~1", "/"), "~0", "~")
	}
	return parts, nil
}

func setAtPointer(root any, path string, value any, appendMode bool) (any, error) {
	parts, err := decodePointer(path)
	if err != nil {
		return root, err
	}
	if len(parts) == 0 {
		if appendMode {
			return appendValue(root, value)
		}
		return value, nil
	}
	return setRecursive(root, parts, value, appendMode)
}

func setRecursive(node any, parts []string, value any, appendMode bool) (any, error) {
	if len(parts) == 0 {
		if appendMode {
			return appendValue(node, value)
		}
		return value, nil
	}
	key := parts[0]
	switch current := node.(type) {
	case nil:
		m := map[string]any{}
		child, err := setRecursive(nil, parts[1:], value, appendMode)
		if err != nil {
			return nil, err
		}
		m[key] = child
		return m, nil
	case map[string]any:
		child, err := setRecursive(current[key], parts[1:], value, appendMode)
		if err != nil {
			return nil, err
		}
		current[key] = child
		return current, nil
	case []any:
		if key == "-" {
			if len(parts) != 1 {
				return nil, errors.New("array append token must terminate path")
			}
			return append(current, value), nil
		}
		idx, err := strconv.Atoi(key)
		if err != nil || idx < 0 {
			return nil, fmt.Errorf("invalid array index %q", key)
		}
		for len(current) <= idx {
			current = append(current, nil)
		}
		child, err := setRecursive(current[idx], parts[1:], value, appendMode)
		if err != nil {
			return nil, err
		}
		current[idx] = child
		return current, nil
	default:
		return nil, fmt.Errorf("cannot traverse %T at %q", node, key)
	}
}

func appendValue(current, value any) (any, error) {
	switch c := current.(type) {
	case nil:
		return value, nil
	case string:
		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("cannot append %T to string", value)
		}
		return c + s, nil
	case []any:
		if values, ok := value.([]any); ok {
			return append(c, values...), nil
		}
		return append(c, value), nil
	default:
		return nil, fmt.Errorf("cannot append to %T", current)
	}
}

func removeAtPointer(root any, path string) (any, error) {
	parts, err := decodePointer(path)
	if err != nil {
		return root, err
	}
	if len(parts) == 0 {
		return nil, nil
	}
	return removeRecursive(root, parts)
}

func removeRecursive(node any, parts []string) (any, error) {
	key := parts[0]
	switch current := node.(type) {
	case map[string]any:
		if len(parts) == 1 {
			delete(current, key)
			return current, nil
		}
		child, ok := current[key]
		if !ok {
			return current, nil
		}
		updated, err := removeRecursive(child, parts[1:])
		if err != nil {
			return nil, err
		}
		current[key] = updated
		return current, nil
	case []any:
		idx, err := strconv.Atoi(key)
		if err != nil || idx < 0 || idx >= len(current) {
			return nil, fmt.Errorf("invalid array index %q", key)
		}
		if len(parts) == 1 {
			return append(current[:idx], current[idx+1:]...), nil
		}
		updated, err := removeRecursive(current[idx], parts[1:])
		if err != nil {
			return nil, err
		}
		current[idx] = updated
		return current, nil
	default:
		return nil, fmt.Errorf("cannot traverse %T at %q", node, key)
	}
}
