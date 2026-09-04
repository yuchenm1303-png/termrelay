package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service/chatgptweb"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestChatGPTWebStickyStoreBindRefreshRelease(t *testing.T) {
	store, mini := newChatGPTWebStickyStoreTest(t)
	ctx := context.Background()
	key := chatgptweb.StickyKey{Kind: chatgptweb.StickyKeyConversation, ID: "conv-1"}

	winner, created, err := store.Bind(ctx, key, 11, time.Minute)
	require.NoError(t, err)
	require.True(t, created)
	require.Equal(t, int64(11), winner)

	winner, created, err = store.Bind(ctx, key, 22, time.Minute)
	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, int64(11), winner)

	mini.FastForward(40 * time.Second)
	refreshed, err := store.Refresh(ctx, key, 11, time.Minute)
	require.NoError(t, err)
	require.True(t, refreshed)

	mini.FastForward(30 * time.Second)
	boundID, found, err := store.GetBinding(ctx, key)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(11), boundID)

	mini.FastForward(31 * time.Second)
	_, found, err = store.GetBinding(ctx, key)
	require.NoError(t, err)
	require.False(t, found)

	winner, created, err = store.Bind(ctx, key, 22, time.Minute)
	require.NoError(t, err)
	require.True(t, created)
	require.Equal(t, int64(22), winner)

	released, err := store.Release(ctx, key, 11)
	require.NoError(t, err)
	require.False(t, released)
	released, err = store.Release(ctx, key, 22)
	require.NoError(t, err)
	require.True(t, released)

	winner, created, err = store.Bind(ctx, key, 33, time.Minute)
	require.NoError(t, err)
	require.True(t, created)
	require.Equal(t, int64(33), winner)
}

func TestChatGPTWebStickyStoreConcurrentBindHasOneWinner(t *testing.T) {
	store, _ := newChatGPTWebStickyStoreTest(t)
	ctx := context.Background()
	key := chatgptweb.StickyKey{Kind: chatgptweb.StickyKeyClientSession, ID: "session-race"}

	const workers = 32
	type result struct {
		winner int64
		err    error
	}
	start := make(chan struct{})
	results := make(chan result, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(accountID int64) {
			defer wg.Done()
			<-start
			winner, _, err := store.Bind(ctx, key, accountID, time.Minute)
			results <- result{winner: winner, err: err}
		}(int64(i + 1))
	}
	close(start)
	wg.Wait()
	close(results)

	var winner int64
	for result := range results {
		require.NoError(t, result.err)
		if winner == 0 {
			winner = result.winner
		}
		require.Equal(t, winner, result.winner)
	}
	require.NotZero(t, winner)

	boundID, found, err := store.GetBinding(ctx, key)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, winner, boundID)
}

func TestChatGPTWebStickyStoreSeparatesConversations(t *testing.T) {
	store, _ := newChatGPTWebStickyStoreTest(t)
	ctx := context.Background()

	for i, accountID := range []int64{11, 22} {
		key := chatgptweb.StickyKey{Kind: chatgptweb.StickyKeyConversation, ID: fmt.Sprintf("conv-%d", i+1)}
		winner, created, err := store.Bind(ctx, key, accountID, time.Minute)
		require.NoError(t, err)
		require.True(t, created)
		require.Equal(t, accountID, winner)
	}

	first, found, err := store.GetBinding(ctx, chatgptweb.StickyKey{Kind: chatgptweb.StickyKeyConversation, ID: "conv-1"})
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(11), first)
	second, found, err := store.GetBinding(ctx, chatgptweb.StickyKey{Kind: chatgptweb.StickyKeyConversation, ID: "conv-2"})
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(22), second)
}

func newChatGPTWebStickyStoreTest(t *testing.T) (chatgptweb.StickyStore, *miniredis.Miniredis) {
	t.Helper()
	mini := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return NewChatGPTWebStickyStore(rdb), mini
}
