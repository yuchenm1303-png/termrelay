package chatgptweb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGatewayTransientConversationErrorPreservesBinding(t *testing.T) {
	store := newMemoryStickyStore()
	state := NewClientState("device-transient", "test", time.Now())
	state.ConversationID = "conv-transient"
	state.ParentMessageID = "assistant-old"
	key := StickyKey{Kind: StickyKeyConversation, ID: state.ConversationID}
	_, _, err := store.Bind(context.Background(), key, 11, time.Hour)
	require.NoError(t, err)

	selector := &fakeAccountSelector{available: map[int64]bool{11: true, 22: true}, selections: []int64{22}}
	transport := &fakeTransport{errs: []error{&UpstreamError{Kind: ErrorKindRateLimit, StatusCode: 429, RetryNextAccount: true}}}
	gateway, err := NewGateway(store, selector, transport, time.Hour)
	require.NoError(t, err)

	_, err = gateway.Send(context.Background(), SendRequest{ClientState: state, Conversation: ConversationRequest{Model: "auto"}})
	require.Error(t, err)
	require.Equal(t, 0, selector.selectCalls)

	boundID, found, err := store.GetBinding(context.Background(), key)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(11), boundID)
}

func TestGatewayAuthenticationConversationErrorReleasesBinding(t *testing.T) {
	store := newMemoryStickyStore()
	state := NewClientState("device-auth", "test", time.Now())
	state.ConversationID = "conv-auth"
	state.ParentMessageID = "assistant-old"
	key := StickyKey{Kind: StickyKeyConversation, ID: state.ConversationID}
	_, _, err := store.Bind(context.Background(), key, 11, time.Hour)
	require.NoError(t, err)

	selector := &fakeAccountSelector{available: map[int64]bool{11: true}}
	transport := &fakeTransport{errs: []error{&UpstreamError{Kind: ErrorKindAuthentication, StatusCode: 401, RequiresReauth: true}}}
	gateway, err := NewGateway(store, selector, transport, time.Hour)
	require.NoError(t, err)

	_, err = gateway.Send(context.Background(), SendRequest{ClientState: state, Conversation: ConversationRequest{Model: "auto"}})
	require.Error(t, err)

	_, found, err := store.GetBinding(context.Background(), key)
	require.NoError(t, err)
	require.False(t, found)
}

func TestGatewayInitialSessionTransientErrorCanRotateAccount(t *testing.T) {
	store := newMemoryStickyStore()
	state := NewClientState("device-rotate", "test", time.Now())
	selector := &fakeAccountSelector{available: map[int64]bool{11: true, 22: true}, selections: []int64{11, 22}}
	transport := &fakeTransport{
		errs: []error{&UpstreamError{Kind: ErrorKindTransient, StatusCode: 503, RetryNextAccount: true}},
		snapshots: []Snapshot{
			{},
			{ConversationID: "conv-rotated", ParentMessageID: "assistant-new", Finished: true},
		},
	}
	gateway, err := NewGateway(store, selector, transport, time.Hour)
	require.NoError(t, err)

	result, err := gateway.Send(context.Background(), SendRequest{ClientState: state, Conversation: ConversationRequest{Model: "auto"}})
	require.NoError(t, err)
	require.Equal(t, int64(22), result.AccountID)
	require.Equal(t, []int64{11, 22}, transport.accounts)

	boundID, found, err := store.GetBinding(context.Background(), StickyKey{Kind: StickyKeyConversation, ID: "conv-rotated"})
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(22), boundID)
}
