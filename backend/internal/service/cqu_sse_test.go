package service

import (
	"context"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"
)

// chunkedReader replays a payload in fixed-size pieces so a parser that assumes
// frame-aligned reads fails loudly.
type chunkedReader struct {
	data []byte
	size int
}

func (r *chunkedReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	n := r.size
	if n > len(r.data) {
		n = len(r.data)
	}
	if n > len(p) {
		n = len(p)
	}
	copy(p, r.data[:n])
	r.data = r.data[n:]
	return n, nil
}

func collectCQUEvents(t *testing.T, r io.Reader) []CQUSSEEvent {
	t.Helper()
	var events []CQUSSEEvent
	require.NoError(t, ParseCQUSSE(r, func(event CQUSSEEvent) error {
		events = append(events, event)
		return nil
	}))
	return events
}

func TestParseCQUSSEExtractsTextDelta(t *testing.T) {
	stream := strings.Join([]string{
		"event: RUN_STARTED\n",
		"data: {\"threadId\":\"t1\",\"runId\":\"r1\"}\n\n",
		"event: TEXT_MESSAGE_START\n",
		"data: {\"messageId\":\"m1\",\"role\":\"assistant\"}\n\n",
		"event: TEXT_MESSAGE_CONTENT\n",
		"data: {\"messageId\":\"m1\",\"delta\":\"你\"}\n\n",
		"event: TEXT_MESSAGE_CONTENT\n",
		"data: {\"messageId\":\"m1\",\"delta\":\"好\"}\n\n",
		"event: TEXT_MESSAGE_END\n",
		"data: {\"messageId\":\"m1\"}\n\n",
		"event: RUN_FINISHED\n",
		"data: {\"threadId\":\"t1\",\"runId\":\"r1\"}\n\n",
	}, "")

	var events []CQUSSEEvent
	err := ParseCQUSSE(strings.NewReader(stream), func(event CQUSSEEvent) error {
		events = append(events, event)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, events, 6)
	require.Equal(t, CQUSSEEventRunStarted, events[0].Type)
	require.Equal(t, CQUSSEEventTextMessageContent, events[2].Type)
	require.Equal(t, "你", events[2].Delta)
	require.Equal(t, "好", events[3].Delta)
	require.Equal(t, CQUSSEEventRunFinished, events[5].Type)
}

func TestParseCQUSSEInfersTypeFromData(t *testing.T) {
	stream := "data: {\"type\":\"TEXT_MESSAGE_CONTENT\",\"delta\":\"hello\"}\n\n"

	var got CQUSSEEvent
	err := ParseCQUSSE(strings.NewReader(stream), func(event CQUSSEEvent) error {
		got = event
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, CQUSSEEventTextMessageContent, got.Type)
	require.Equal(t, "hello", got.Delta)
}

func TestParseCQUSSESupportsMultilineDataAndEOFDispatch(t *testing.T) {
	stream := "event: CUSTOM\ndata: {\"name\":\"x\",\ndata: \"value\":1}"

	var events []CQUSSEEvent
	err := ParseCQUSSE(strings.NewReader(stream), func(event CQUSSEEvent) error {
		events = append(events, event)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, CQUSSEEventCustom, events[0].Type)
	require.JSONEq(t, "{\"name\":\"x\",\"value\":1}", string(events[0].Data))
}

// A CQU stream arrives over CDP as arbitrary byte chunks, so frame boundaries
// never line up with reads. Parsing must be identical whichever way it splits.
func TestParseCQUSSEHandlesArbitraryChunkBoundaries(t *testing.T) {
	stream := strings.Join([]string{
		"event: TEXT_MESSAGE_CONTENT\ndata: {\"delta\":\"你\"}\n\n",
		"event: TEXT_MESSAGE_CONTENT\ndata: {\"delta\":\"好\"}\n\n",
		"event: RUN_FINISHED\ndata: {\"threadId\":\"t1\"}\n\n",
	}, "")

	for _, size := range []int{1, 3, 7, 64, len(stream)} {
		events := collectCQUEvents(t, &chunkedReader{data: []byte(stream), size: size})
		require.Len(t, events, 3, "chunk size %d", size)
		require.Equal(t, "你", events[0].Delta)
		require.Equal(t, "好", events[1].Delta)
		require.Equal(t, CQUSSEEventRunFinished, events[2].Type)
	}

	events := collectCQUEvents(t, iotest.OneByteReader(strings.NewReader(stream)))
	require.Len(t, events, 3)
}

// One chunk carrying several complete events must produce several events.
func TestParseCQUSSEHandlesMultipleEventsInOneChunk(t *testing.T) {
	stream := "event: STEP_STARTED\ndata: {}\n\nevent: TEXT_MESSAGE_CONTENT\ndata: {\"delta\":\"a\"}\n\nevent: STEP_FINISHED\ndata: {}\n\n"

	events := collectCQUEvents(t, strings.NewReader(stream))
	require.Len(t, events, 3)
	require.Equal(t, CQUSSEEventStepStarted, events[0].Type)
	require.Equal(t, "a", events[1].Delta)
	require.Equal(t, CQUSSEEventStepFinished, events[2].Type)
}

func TestParseCQUSSEHandlesCRLFFraming(t *testing.T) {
	stream := "event: TEXT_MESSAGE_CONTENT\r\ndata: {\"delta\":\"你好\"}\r\n\r\nevent: RUN_FINISHED\r\ndata: {\"threadId\":\"t1\"}\r\n\r\n"

	events := collectCQUEvents(t, strings.NewReader(stream))
	require.Len(t, events, 2)
	require.Equal(t, "你好", events[0].Delta)
	require.Equal(t, CQUSSEEventRunFinished, events[1].Type)
}

// A stream that ends without a trailing blank line must still deliver its last
// frame rather than silently dropping the final token.
func TestParseCQUSSEFlushesTrailingFrameAtEOF(t *testing.T) {
	events := collectCQUEvents(t, strings.NewReader("event: TEXT_MESSAGE_CONTENT\ndata: {\"delta\":\"tail\"}"))
	require.Len(t, events, 1)
	require.Equal(t, "tail", events[0].Delta)
}

func TestParseCQUSSEIgnoresCommentsAndUnknownFields(t *testing.T) {
	stream := ": keep-alive\nid: 42\nretry: 1000\nevent: TEXT_MESSAGE_CONTENT\ndata: {\"delta\":\"x\"}\n\n"

	events := collectCQUEvents(t, strings.NewReader(stream))
	require.Len(t, events, 1)
	require.Equal(t, "x", events[0].Delta)
}

func TestParseCQUSSEContextCancellationStopsImmediately(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ParseCQUSSEContext(ctx, strings.NewReader("event: TEXT_MESSAGE_CONTENT\ndata: {\"delta\":\"x\"}\n\n"), func(CQUSSEEvent) error {
		t.Fatal("emit must not run after cancellation")
		return nil
	})
	require.ErrorIs(t, err, context.Canceled)
}

func TestCQUApplyIdentifiersCollectsConversationAndMessageIDs(t *testing.T) {
	stream := strings.Join([]string{
		"event: RUN_STARTED\ndata: {\"threadId\":\"thread-7\",\"runId\":\"run-7\"}\n\n",
		"event: TEXT_MESSAGE_CONTENT\ndata: {\"delta\":\"hi\"}\n\n",
		"event: TEXT_MESSAGE_END\ndata: {\"messageId\":\"msg-7\"}\n\n",
		"event: RUN_FINISHED\ndata: {\"threadId\":\"thread-7\",\"conversationId\":\"conv-7\"}\n\n",
	}, "")

	var ids CQUStreamIdentifiers
	for _, event := range collectCQUEvents(t, strings.NewReader(stream)) {
		cquApplyIdentifiers(event, &ids)
	}

	require.Equal(t, "thread-7", ids.ThreadID)
	require.Equal(t, "conv-7", ids.ConversationID)
	require.Equal(t, "run-7", ids.RunID)
	require.Equal(t, "msg-7", ids.MessageID)
	require.Equal(t, "conv-7", ids.ConversationKey())

	// With no explicit conversationId, the thread handle is the conversation.
	require.Equal(t, "thread-7", CQUStreamIdentifiers{ThreadID: "thread-7"}.ConversationKey())
}
