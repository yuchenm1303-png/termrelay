package chatgptweb

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseRoutingModeKeepsProviderDefault(t *testing.T) {
	mode, err := ParseRoutingMode("")
	require.NoError(t, err)
	require.Equal(t, RoutingModeProvider, mode)

	mode, err = ParseRoutingMode("provider")
	require.NoError(t, err)
	require.Equal(t, RoutingModeProvider, mode)

	mode, err = ParseRoutingMode("chatgpt_web")
	require.NoError(t, err)
	require.Equal(t, RoutingModeChatGPTWeb, mode)

	_, err = ParseRoutingMode("unknown")
	require.Error(t, err)
}

func TestGatewayBindsFirstConversationAndReusesAccount(t *testing.T) {
	store := newMemoryStickyStore()
	selector := &fakeAccountSelector{available: map[int64]bool{11: true, 22: true}, selections: []int64{11, 22}}
	transport := &fakeTransport{snapshots: []Snapshot{
		{ConversationID: "conv-1", ParentMessageID: "assistant-1", FinalText: "one", Finished: true},
		{ConversationID: "conv-1", ParentMessageID: "assistant-2", FinalText: "two", Finished: true},
	}}
	gateway, err := NewGateway(store, selector, transport, 30*time.Minute)
	require.NoError(t, err)

	state := NewClientState("device-1", "test", time.Now())
	first, err := gateway.Send(context.Background(), SendRequest{ClientState: state, Conversation: ConversationRequest{Model: "auto"}})
	require.NoError(t, err)
	require.Equal(t, int64(11), first.AccountID)
	require.Equal(t, "conv-1", state.ConversationID)
	require.Equal(t, "assistant-1", state.ParentMessageID)
	require.Equal(t, int64(11), first.Session.AccountID)
	require.Equal(t, string(RoutingModeChatGPTWeb), first.Session.Route)

	second, err := gateway.Send(context.Background(), SendRequest{ClientState: state, Conversation: ConversationRequest{Model: "auto"}})
	require.NoError(t, err)
	require.Equal(t, int64(11), second.AccountID)
	require.Equal(t, "assistant-2", state.ParentMessageID)
	require.Equal(t, []int64{11, 11}, transport.accounts)
	require.Equal(t, 1, selector.selectCalls)

	conversationBinding, found, err := store.GetBinding(context.Background(), StickyKey{Kind: StickyKeyConversation, ID: "conv-1"})
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(11), conversationBinding)
}

func TestGatewayUnavailableSessionBindingReleasesAndReselects(t *testing.T) {
	store := newMemoryStickyStore()
	state := NewClientState("device-2", "test", time.Now())
	sessionKey := StickyKey{Kind: StickyKeyClientSession, ID: state.SessionID}
	_, _, err := store.Bind(context.Background(), sessionKey, 11, time.Hour)
	require.NoError(t, err)

	selector := &fakeAccountSelector{available: map[int64]bool{11: false, 22: true}, selections: []int64{22}}
	transport := &fakeTransport{snapshots: []Snapshot{{ConversationID: "conv-2", ParentMessageID: "assistant-1", Finished: true}}}
	gateway, err := NewGateway(store, selector, transport, time.Hour)
	require.NoError(t, err)

	result, err := gateway.Send(context.Background(), SendRequest{ClientState: state, Conversation: ConversationRequest{Model: "auto"}})
	require.NoError(t, err)
	require.Equal(t, int64(22), result.AccountID)

	boundID, found, err := store.GetBinding(context.Background(), sessionKey)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(22), boundID)
}

func TestGatewayUnavailableConversationBindingIsReleasedWithoutMigration(t *testing.T) {
	store := newMemoryStickyStore()
	state := NewClientState("device-3", "test", time.Now())
	state.ConversationID = "conv-stale"
	state.ParentMessageID = "assistant-old"
	conversationKey := StickyKey{Kind: StickyKeyConversation, ID: state.ConversationID}
	_, _, err := store.Bind(context.Background(), conversationKey, 11, time.Hour)
	require.NoError(t, err)

	selector := &fakeAccountSelector{available: map[int64]bool{11: false, 22: true}, selections: []int64{22}}
	gateway, err := NewGateway(store, selector, &fakeTransport{}, time.Hour)
	require.NoError(t, err)

	_, err = gateway.Send(context.Background(), SendRequest{ClientState: state, Conversation: ConversationRequest{Model: "auto"}})
	require.ErrorIs(t, err, ErrConversationAccountUnavailable)
	require.Equal(t, 0, selector.selectCalls)

	_, found, err := store.GetBinding(context.Background(), conversationKey)
	require.NoError(t, err)
	require.False(t, found)
}

func TestGatewayMissingConversationBindingDoesNotChooseRandomAccount(t *testing.T) {
	store := newMemoryStickyStore()
	state := NewClientState("device-4", "test", time.Now())
	state.ConversationID = "conv-expired"
	state.ParentMessageID = "assistant-old"
	selector := &fakeAccountSelector{available: map[int64]bool{22: true}, selections: []int64{22}}
	gateway, err := NewGateway(store, selector, &fakeTransport{}, time.Hour)
	require.NoError(t, err)

	_, err = gateway.Send(context.Background(), SendRequest{ClientState: state, Conversation: ConversationRequest{Model: "auto"}})
	require.ErrorIs(t, err, ErrConversationBindingMissing)
	require.Equal(t, 0, selector.selectCalls)
}

type memoryStickyStore struct {
	mu       sync.Mutex
	bindings map[string]int64
}

func newMemoryStickyStore() *memoryStickyStore {
	return &memoryStickyStore{bindings: make(map[string]int64)}
}

func (s *memoryStickyStore) GetBinding(_ context.Context, key StickyKey) (int64, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, found := s.bindings[key.String()]
	return id, found, nil
}

func (s *memoryStickyStore) Bind(_ context.Context, key StickyKey, accountID int64, _ time.Duration) (int64, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if current, found := s.bindings[key.String()]; found {
		return current, false, nil
	}
	s.bindings[key.String()] = accountID
	return accountID, true, nil
}

func (s *memoryStickyStore) Refresh(_ context.Context, key StickyKey, accountID int64, _ time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.bindings[key.String()] == accountID, nil
}

func (s *memoryStickyStore) Release(_ context.Context, key StickyKey, accountID int64) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.bindings[key.String()] != accountID {
		return false, nil
	}
	delete(s.bindings, key.String())
	return true, nil
}

type fakeAccountSelector struct {
	mu          sync.Mutex
	available   map[int64]bool
	selections  []int64
	selectCalls int
}

func (s *fakeAccountSelector) Resolve(_ context.Context, accountID int64) (AccountRef, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return AccountRef{ID: accountID}, s.available[accountID], nil
}

func (s *fakeAccountSelector) Select(_ context.Context, excluded []int64) (AccountRef, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selectCalls++
	excludedSet := make(map[int64]struct{}, len(excluded))
	for _, id := range excluded {
		excludedSet[id] = struct{}{}
	}
	for _, id := range s.selections {
		if _, skip := excludedSet[id]; skip || !s.available[id] {
			continue
		}
		return AccountRef{ID: id}, nil
	}
	return AccountRef{}, errors.New("no account available")
}

type fakeTransport struct {
	mu        sync.Mutex
	snapshots []Snapshot
	errs      []error
	accounts  []int64
}

func (t *fakeTransport) Send(_ context.Context, account AccountRef, _ TransportRequest) (Snapshot, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.accounts = append(t.accounts, account.ID)
	idx := len(t.accounts) - 1
	if idx < len(t.errs) && t.errs[idx] != nil {
		return Snapshot{}, t.errs[idx]
	}
	if idx < len(t.snapshots) {
		return t.snapshots[idx], nil
	}
	return Snapshot{}, nil
}
