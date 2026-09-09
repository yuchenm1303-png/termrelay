package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// CQU emits AG-UI-style Server-Sent Events. Keep the event names here instead
// of scattering string literals through the provider adapter.
const (
	CQUSSEEventRunStarted         = "RUN_STARTED"
	CQUSSEEventRunFinished        = "RUN_FINISHED"
	CQUSSEEventRunError           = "RUN_ERROR"
	CQUSSEEventStepStarted        = "STEP_STARTED"
	CQUSSEEventStepFinished       = "STEP_FINISHED"
	CQUSSEEventTextMessageStart   = "TEXT_MESSAGE_START"
	CQUSSEEventTextMessageContent = "TEXT_MESSAGE_CONTENT"
	CQUSSEEventTextMessageEnd     = "TEXT_MESSAGE_END"
	CQUSSEEventToolCallStart      = "TOOL_CALL_START"
	CQUSSEEventToolCallArgs       = "TOOL_CALL_ARGS"
	CQUSSEEventToolCallEnd        = "TOOL_CALL_END"
	CQUSSEEventStateSnapshot      = "STATE_SNAPSHOT"
	CQUSSEEventStateDelta         = "STATE_DELTA"
	CQUSSEEventMessagesSnapshot   = "MESSAGES_SNAPSHOT"
	CQUSSEEventRaw                = "RAW"
	CQUSSEEventCustom             = "CUSTOM"
)

// CQUSSEEvent is the minimally-normalized representation of one upstream SSE
// frame. Data remains available verbatim so future CQU event variants can be
// supported without changing the transport layer.
type CQUSSEEvent struct {
	Type  string
	Data  json.RawMessage
	Delta string
	Raw   []byte
}

// ParseCQUSSE incrementally parses a CQU event stream. It deliberately does not
// use bufio.Scanner so a large model/tool event is not capped by Scanner's token
// limit.
func ParseCQUSSE(r io.Reader, emit func(CQUSSEEvent) error) error {
	if r == nil {
		return errors.New("cqu sse reader is nil")
	}
	if emit == nil {
		return errors.New("cqu sse emitter is nil")
	}

	reader := bufio.NewReader(r)
	var eventName string
	var dataLines [][]byte
	var raw bytes.Buffer

	dispatch := func() error {
		if eventName == "" && len(dataLines) == 0 {
			raw.Reset()
			return nil
		}

		data := bytes.Join(dataLines, []byte("\n"))
		eventType := strings.TrimSpace(eventName)
		if eventType == "" {
			eventType = cquEventTypeFromJSON(data)
		}

		event := CQUSSEEvent{
			Type:  eventType,
			Data:  append(json.RawMessage(nil), data...),
			Raw:   append([]byte(nil), raw.Bytes()...),
			Delta: cquTextDelta(eventType, data),
		}
		if err := emit(event); err != nil {
			return err
		}

		eventName = ""
		dataLines = dataLines[:0]
		raw.Reset()
		return nil
	}

	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			raw.Write(line)
			trimmed := bytes.TrimRight(line, "\r\n")
			if len(trimmed) == 0 {
				if dispatchErr := dispatch(); dispatchErr != nil {
					return dispatchErr
				}
			} else if trimmed[0] != ':' {
				field, value, found := bytes.Cut(trimmed, []byte(":"))
				if found && len(value) > 0 && value[0] == ' ' {
					value = value[1:]
				}
				switch string(field) {
				case "event":
					eventName = string(value)
				case "data":
					dataLines = append(dataLines, append([]byte(nil), value...))
				}
			}
		}

		if err != nil {
			if !errors.Is(err, io.EOF) {
				return fmt.Errorf("read cqu sse: %w", err)
			}
			if raw.Len() > 0 || eventName != "" || len(dataLines) > 0 {
				if dispatchErr := dispatch(); dispatchErr != nil {
					return dispatchErr
				}
			}
			return nil
		}
	}
}

func cquEventTypeFromJSON(data []byte) string {
	if len(bytes.TrimSpace(data)) == 0 {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}
	for _, key := range []string{"type", "event", "eventType", "event_type"} {
		if value, ok := payload[key].(string); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func cquTextDelta(eventType string, data []byte) string {
	if eventType != CQUSSEEventTextMessageContent || len(bytes.TrimSpace(data)) == 0 {
		return ""
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}

	// CQU currently follows AG-UI where TEXT_MESSAGE_CONTENT carries `delta`.
	// The fallbacks keep the parser tolerant to observed wrapper variants while
	// still only extracting text from a content event.
	for _, key := range []string{"delta", "content", "text"} {
		if value, ok := payload[key].(string); ok {
			return value
		}
	}
	if nested, ok := payload["data"].(map[string]any); ok {
		for _, key := range []string{"delta", "content", "text"} {
			if value, ok := nested[key].(string); ok {
				return value
			}
		}
	}
	return ""
}
