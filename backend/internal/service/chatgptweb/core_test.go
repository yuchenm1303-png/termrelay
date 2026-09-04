package chatgptweb

import (
	"io"
	"strings"
	"testing"
	"time"
)

func TestClientStateNoteTurnResult(t *testing.T) {
	now := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)
	state := NewClientState("device-1", "ua", now)
	if state.ParentMessageID != RootParentMessageID {
		t.Fatalf("parent = %q", state.ParentMessageID)
	}
	state.NoteTurnResult("conv-1", "msg-1")
	if state.ConversationID != "conv-1" || state.ParentMessageID != "msg-1" {
		t.Fatalf("unexpected state: %+v", state)
	}
	if got := state.TimeSinceLoaded(now.Add(3 * time.Second)); got != 3*time.Second {
		t.Fatalf("duration = %v", got)
	}
}

func TestClassifyHTTPError(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		kind      ErrorKind
		retry     bool
		reauth    bool
		challenge bool
	}{
		{"unauthorized", 401, "", ErrorKindAuthentication, false, true, false},
		{"challenge", 403, `{"detail":"turnstile challenge_required"}`, ErrorKindChallenge, false, false, true},
		{"forbidden", 403, `{"detail":"forbidden"}`, ErrorKindAuthentication, false, true, false},
		{"rate-limit", 429, "", ErrorKindRateLimit, true, false, false},
		{"server", 503, "", ErrorKindTransient, true, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ClassifyHTTPError(tt.status, []byte(tt.body), 2*time.Second)
			if err.Kind != tt.kind || err.RetryNextAccount != tt.retry || err.RequiresReauth != tt.reauth || err.ChallengeRequired != tt.challenge {
				t.Fatalf("classification = %+v", err)
			}
		})
	}
}

func TestPatchAccumulatorFullMessageAndPatches(t *testing.T) {
	acc := NewPatchAccumulator()
	full := []byte(`{"conversation_id":"conv-1","message":{"id":"msg-1","author":{"role":"assistant"},"channel":"final","content":{"content_type":"text","parts":["Hel"]}}}`)
	snapshot, err := acc.ApplyJSON(full)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.ConversationID != "conv-1" || snapshot.ParentMessageID != "msg-1" || snapshot.FinalText != "Hel" {
		t.Fatalf("full snapshot = %+v", snapshot)
	}

	snapshot, err = acc.ApplyJSON([]byte(`{"p":"/message/content/parts/0","o":"append","v":"lo"}`))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.FinalText != "Hello" {
		t.Fatalf("text = %q", snapshot.FinalText)
	}

	snapshot, err = acc.ApplyJSON([]byte(`{"patches":[{"p":"/message/channel","o":"replace","v":"analysis"},{"p":"/message/content/parts/0","o":"replace","v":"Thinking"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.CommentaryText != "Thinking" || snapshot.Channel != "analysis" {
		t.Fatalf("analysis snapshot = %+v", snapshot)
	}
}

func TestPatchAccumulatorBatchArray(t *testing.T) {
	acc := NewPatchAccumulator()
	_, err := acc.ApplyJSON([]byte(`[
		{"p":"/message","o":"set","v":{"id":"msg-2","author":{"role":"assistant"},"content":{"parts":[""]}}},
		{"p":"/conversation_id","o":"set","v":"conv-2"},
		{"p":"/message/content/parts/0","o":"append","v":"OK"},
		{"p":"/finish_details","o":"set","v":{"type":"stop"}}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	snapshot := acc.Snapshot()
	if snapshot.ConversationID != "conv-2" || snapshot.FinalText != "OK" || !snapshot.Finished || snapshot.FinishReason != "stop" {
		t.Fatalf("snapshot = %+v", snapshot)
	}
}

func TestStreamParserSSE(t *testing.T) {
	stream := strings.Join([]string{
		`data: {"conversation_id":"conv-1","message":{"id":"msg-1","author":{"role":"assistant"},"channel":"final","content":{"parts":["Hel"]}}}`,
		"",
		`data: {"p":"/message/content/parts/0","o":"append","v":"lo"}`,
		"",
		`data: {"p":"/finish_details","o":"set","v":{"type":"stop"}}`,
		"",
		"data: [DONE]",
		"",
	}, "\n")
	parser := NewStreamParser(strings.NewReader(stream))

	event, err := parser.Next()
	if err != nil || event.FinalDelta != "Hel" {
		t.Fatalf("first event = %+v err=%v", event, err)
	}
	event, err = parser.Next()
	if err != nil || event.FinalDelta != "lo" || event.Snapshot.FinalText != "Hello" {
		t.Fatalf("second event = %+v err=%v", event, err)
	}
	event, err = parser.Next()
	if err != nil || !event.Snapshot.Finished || event.Snapshot.FinishReason != "stop" {
		t.Fatalf("finish event = %+v err=%v", event, err)
	}
	event, err = parser.Next()
	if err != nil || event.Type != StreamEventDone {
		t.Fatalf("done event = %+v err=%v", event, err)
	}
	if _, err = parser.Next(); err != io.EOF {
		t.Fatalf("after done err=%v", err)
	}
}
