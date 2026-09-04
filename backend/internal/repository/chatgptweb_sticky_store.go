package repository

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service/chatgptweb"
	"github.com/redis/go-redis/v9"
)

const chatGPTWebStickyKeyPrefix = "termrelay:chatgptweb:sticky:v1:"

var (
	chatGPTWebStickyBindScript = redis.NewScript(`
local current = redis.call("GET", KEYS[1])
if current then
  return {current, 0}
end
redis.call("PSETEX", KEYS[1], ARGV[2], ARGV[1])
return {ARGV[1], 1}
`)
	chatGPTWebStickyRefreshScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
return 0
`)
	chatGPTWebStickyReleaseScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
`)
)

type chatGPTWebStickyStore struct {
	rdb *redis.Client
}

var _ chatgptweb.StickyStore = (*chatGPTWebStickyStore)(nil)

// NewChatGPTWebStickyStore uses the application's existing Redis client. It
// deliberately does not own Redis initialization or lifecycle.
func NewChatGPTWebStickyStore(rdb *redis.Client) chatgptweb.StickyStore {
	return &chatGPTWebStickyStore{rdb: rdb}
}

func (s *chatGPTWebStickyStore) GetBinding(ctx context.Context, key chatgptweb.StickyKey) (int64, bool, error) {
	redisKey, err := chatGPTWebStickyRedisKey(key)
	if err != nil {
		return 0, false, err
	}
	value, err := s.rdb.Get(ctx, redisKey).Result()
	if errors.Is(err, redis.Nil) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	accountID, err := strconv.ParseInt(value, 10, 64)
	if err != nil || accountID <= 0 {
		return 0, false, fmt.Errorf("chatgptweb sticky store: invalid account id %q", value)
	}
	return accountID, true, nil
}

func (s *chatGPTWebStickyStore) Bind(ctx context.Context, key chatgptweb.StickyKey, accountID int64, ttl time.Duration) (int64, bool, error) {
	redisKey, ttlMillis, err := chatGPTWebStickyArguments(key, accountID, ttl)
	if err != nil {
		return 0, false, err
	}
	result, err := chatGPTWebStickyBindScript.Run(
		ctx,
		s.rdb,
		[]string{redisKey},
		strconv.FormatInt(accountID, 10),
		strconv.FormatInt(ttlMillis, 10),
	).Slice()
	if err != nil {
		return 0, false, err
	}
	if len(result) != 2 {
		return 0, false, errors.New("chatgptweb sticky store: malformed bind result")
	}
	winnerID, err := redisResultAccountID(result[0])
	if err != nil {
		return 0, false, err
	}
	created, err := redisResultInt64(result[1])
	if err != nil {
		return 0, false, err
	}
	return winnerID, created == 1, nil
}

func (s *chatGPTWebStickyStore) Refresh(ctx context.Context, key chatgptweb.StickyKey, accountID int64, ttl time.Duration) (bool, error) {
	redisKey, ttlMillis, err := chatGPTWebStickyArguments(key, accountID, ttl)
	if err != nil {
		return false, err
	}
	result, err := chatGPTWebStickyRefreshScript.Run(
		ctx,
		s.rdb,
		[]string{redisKey},
		strconv.FormatInt(accountID, 10),
		strconv.FormatInt(ttlMillis, 10),
	).Int64()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func (s *chatGPTWebStickyStore) Release(ctx context.Context, key chatgptweb.StickyKey, accountID int64) (bool, error) {
	if accountID <= 0 {
		return false, errors.New("chatgptweb sticky store: account id must be positive")
	}
	redisKey, err := chatGPTWebStickyRedisKey(key)
	if err != nil {
		return false, err
	}
	result, err := chatGPTWebStickyReleaseScript.Run(
		ctx,
		s.rdb,
		[]string{redisKey},
		strconv.FormatInt(accountID, 10),
	).Int64()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func chatGPTWebStickyArguments(key chatgptweb.StickyKey, accountID int64, ttl time.Duration) (string, int64, error) {
	if accountID <= 0 {
		return "", 0, errors.New("chatgptweb sticky store: account id must be positive")
	}
	ttlMillis := ttl.Milliseconds()
	if ttlMillis <= 0 {
		return "", 0, errors.New("chatgptweb sticky store: TTL must be at least one millisecond")
	}
	redisKey, err := chatGPTWebStickyRedisKey(key)
	if err != nil {
		return "", 0, err
	}
	return redisKey, ttlMillis, nil
}

func chatGPTWebStickyRedisKey(key chatgptweb.StickyKey) (string, error) {
	if key.Kind != chatgptweb.StickyKeyConversation && key.Kind != chatgptweb.StickyKeyClientSession {
		return "", fmt.Errorf("chatgptweb sticky store: unsupported key kind %q", key.Kind)
	}
	if strings.TrimSpace(key.ID) == "" {
		return "", errors.New("chatgptweb sticky store: key id is required")
	}
	raw := key.String()
	encoded := base64.RawURLEncoding.EncodeToString([]byte(raw))
	return chatGPTWebStickyKeyPrefix + encoded, nil
}

func redisResultAccountID(value any) (int64, error) {
	var raw string
	switch v := value.(type) {
	case string:
		raw = v
	case []byte:
		raw = string(v)
	default:
		return 0, fmt.Errorf("chatgptweb sticky store: unexpected account result type %T", value)
	}
	accountID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || accountID <= 0 {
		return 0, fmt.Errorf("chatgptweb sticky store: invalid account result %q", raw)
	}
	return accountID, nil
}

func redisResultInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	case []byte:
		return strconv.ParseInt(string(v), 10, 64)
	default:
		return 0, fmt.Errorf("chatgptweb sticky store: unexpected integer result type %T", value)
	}
}
