package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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
