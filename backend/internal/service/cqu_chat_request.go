package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/tidwall/gjson"
)

// CQUDefaultModelAlias is the logical model name clients use. It is not a CQU
// modelId: it resolves to the default agent/model pair below (or to the
// account's own overrides), because CQU's numeric ids are routing internals
// that clients should never have to know.
const CQUDefaultModelAlias = "cqu-default"

// cquRouting is the agent/model pair one CQU request is dispatched to.
type cquRouting struct {
	AgentID string
	ModelID string
}

// resolveCQURouting layers account overrides over the configured defaults.
// A client-supplied model only participates when it is a genuine CQU numeric
// model id; the "cqu-default" alias never leaks into the upstream body.
func resolveCQURouting(account *Account, cfg *config.Config, requestedModel string) cquRouting {
	routing := cquRouting{AgentID: CQUDefaultAgentID, ModelID: CQUDefaultModelID}
	if cfg != nil {
		if agentID := strings.TrimSpace(cfg.CQU.DefaultAgentID); agentID != "" {
			routing.AgentID = agentID
		}
		if modelID := strings.TrimSpace(cfg.CQU.DefaultModelID); modelID != "" {
			routing.ModelID = modelID
		}
	}
	if agentID := cquBrowserAccountString(account, "agent_id", "agentId"); agentID != "" {
		routing.AgentID = agentID
	}
	if modelID := cquBrowserAccountString(account, "model_id", "modelId"); modelID != "" {
		routing.ModelID = modelID
	}
	if model := strings.TrimSpace(requestedModel); isCQUUpstreamModelID(model) {
		routing.ModelID = model
	}
	return routing
}

// isCQUUpstreamModelID reports whether a requested model is a raw CQU model id
// (a long numeric snowflake) rather than a logical alias such as "cqu-default".
func isCQUUpstreamModelID(model string) bool {
	if len(model) < 10 {
		return false
	}
	for _, r := range model {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// BuildCQUChatRequest translates an OpenAI Chat Completions body into CQU's
// native send-chat body.
//
// The OpenAI `messages` array is never forwarded as-is: CQU's send-chat takes a
// single `query` string plus a server-side conversation handle. When the
// conversation is already known upstream, only the newest user turn is sent and
// CQU supplies the history; otherwise the whole exchange is flattened into one
// query so a stateless client still gets coherent context on the first turn.
func BuildCQUChatRequest(body []byte, routing cquRouting, conversationID string) (CQUChatRequest, error) {
	messages := gjson.GetBytes(body, "messages")
	if !messages.Exists() || !messages.IsArray() || len(messages.Array()) == 0 {
		return CQUChatRequest{}, errors.New("messages is required")
	}

	turns := cquCollectTurns(messages)
	if len(turns) == 0 {
		return CQUChatRequest{}, errors.New("messages contain no usable text content")
	}

	query := cquFlattenTurns(turns)
	if strings.TrimSpace(conversationID) != "" {
		// CQU already holds the history for this conversation; resending it
		// would duplicate the transcript inside the model's context.
		if latest := cquLatestUserTurn(turns); latest != "" {
			query = latest
		}
	}
	if strings.TrimSpace(query) == "" {
		return CQUChatRequest{}, errors.New("messages contain no usable text content")
	}

	return CQUChatRequest{
		Query:                query,
		ConversationID:       strings.TrimSpace(conversationID),
		CustomConversationID: "",
		ModelID:              routing.ModelID,
		AgentID:              routing.AgentID,
		HasNetwork:           0,
		ContainerIDs:         []string{},
		DeepThink:            false,
	}, nil
}

type cquTurn struct {
	Role string
	Text string
}

// cquCollectTurns normalizes OpenAI messages into role/text pairs. Roles other
// than system/user/assistant (tool, function, developer) are folded into the
// nearest equivalent rather than dropped, so no instruction silently vanishes.
func cquCollectTurns(messages gjson.Result) []cquTurn {
	turns := make([]cquTurn, 0, len(messages.Array()))
	for _, message := range messages.Array() {
		role := strings.ToLower(strings.TrimSpace(message.Get("role").String()))
		switch role {
		case "system", "developer":
			role = "system"
		case "assistant":
			role = "assistant"
		case "user", "tool", "function", "":
			role = "user"
		default:
			role = "user"
		}
		text := strings.TrimSpace(cquMessageText(message))
		if text == "" {
			continue
		}
		turns = append(turns, cquTurn{Role: role, Text: text})
	}
	return turns
}

// cquMessageText extracts plain text from either a string content or the
// multipart content array. Non-text parts (images, audio, files) are skipped:
// CQU's send-chat body has no field to carry them, and silently dropping them
// is better than sending a corrupted payload.
func cquMessageText(message gjson.Result) string {
	content := message.Get("content")
	switch {
	case content.Type == gjson.String:
		return content.String()
	case content.IsArray():
		var parts []string
		for _, part := range content.Array() {
			if text := part.Get("text"); text.Type == gjson.String && strings.TrimSpace(text.String()) != "" {
				parts = append(parts, text.String())
			}
		}
		return strings.Join(parts, "\n")
	case !content.Exists() || content.Type == gjson.Null:
		return ""
	default:
		return content.String()
	}
}

// cquFlattenTurns renders the whole exchange as a single labelled transcript.
// The trailing user turn is left unlabelled so the model reads it as the actual
// question rather than as one more transcript line.
func cquFlattenTurns(turns []cquTurn) string {
	if len(turns) == 1 {
		if turns[0].Role == "user" {
			return turns[0].Text
		}
	}

	labels := map[string]string{"system": "System", "user": "User", "assistant": "Assistant"}
	lines := make([]string, 0, len(turns))
	last := len(turns) - 1
	for i, turn := range turns {
		if i == last && turn.Role == "user" {
			lines = append(lines, turn.Text)
			continue
		}
		lines = append(lines, labels[turn.Role]+": "+turn.Text)
	}
	return strings.Join(lines, "\n\n")
}

func cquLatestUserTurn(turns []cquTurn) string {
	for i := len(turns) - 1; i >= 0; i-- {
		if turns[i].Role == "user" {
			return turns[i].Text
		}
	}
	return ""
}

// CQUConversationKey derives the lookup handle for the history a CQU
// conversation already contains.
//
// On a follow-up request the client replays every earlier message plus the
// assistant answer we returned, then appends the new question. Hashing
// everything except that trailing user turn therefore reproduces the key that
// was stored when the previous turn finished.
func CQUConversationKey(body []byte) string {
	messages := gjson.GetBytes(body, "messages")
	if !messages.Exists() || !messages.IsArray() {
		return ""
	}
	turns := cquCollectTurns(messages)
	if len(turns) == 0 || turns[len(turns)-1].Role != "user" {
		return ""
	}
	return cquHashTurns(turns[:len(turns)-1])
}

// cquNextConversationKeySeed hashes every turn the client just sent. Combined
// with the assistant answer by cquFinalConversationKey it yields exactly the
// key the next request will look up, without keeping the transcript around.
func cquNextConversationKeySeed(body []byte) string {
	messages := gjson.GetBytes(body, "messages")
	if !messages.Exists() || !messages.IsArray() {
		return ""
	}
	return cquHashTurns(cquCollectTurns(messages))
}

// cquFinalConversationKey extends a seed with the assistant answer produced for
// it. The rolling construction is what lets the seed be computed while the
// request is being built and finished once the answer is known.
func cquFinalConversationKey(seed, assistantText string) string {
	if strings.TrimSpace(seed) == "" {
		return ""
	}
	if strings.TrimSpace(assistantText) == "" {
		return seed
	}
	return cquRollKey(seed, cquTurn{Role: "assistant", Text: assistantText})
}

// cquRollKey folds one turn into a running transcript key.
func cquRollKey(previous string, turn cquTurn) string {
	digest := sha256.New()
	_, _ = digest.Write([]byte(previous))
	_, _ = digest.Write([]byte{0})
	_, _ = digest.Write([]byte(turn.Role))
	_, _ = digest.Write([]byte{0})
	_, _ = digest.Write([]byte(strings.TrimSpace(turn.Text)))
	return hex.EncodeToString(digest.Sum(nil))
}

func cquHashTurns(turns []cquTurn) string {
	key := ""
	for _, turn := range turns {
		key = cquRollKey(key, turn)
	}
	return key
}

// cquConversationStore maps a conversation prefix to the CQU-side conversation
// handle so multi-turn chats reuse the upstream thread.
//
// It is deliberately an in-process bounded map rather than a new database
// table: the mapping is a cache, losing it only costs one flattened-transcript
// turn, and CQU conversation ids are routing handles, not credentials.
type cquConversationStore struct {
	mu      sync.Mutex
	entries map[string]cquConversationEntry
	ttl     time.Duration
	max     int
}

type cquConversationEntry struct {
	ConversationID string
	ExpiresAt      time.Time
}

func newCQUConversationStore(ttl time.Duration, max int) *cquConversationStore {
	if ttl <= 0 {
		ttl = 2 * time.Hour
	}
	if max <= 0 {
		max = 2048
	}
	return &cquConversationStore{entries: make(map[string]cquConversationEntry), ttl: ttl, max: max}
}

func (s *cquConversationStore) Get(key string) string {
	if s == nil || strings.TrimSpace(key) == "" {
		return ""
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[key]
	if !ok {
		return ""
	}
	if time.Now().After(entry.ExpiresAt) {
		delete(s.entries, key)
		return ""
	}
	return entry.ConversationID
}

func (s *cquConversationStore) Put(key, conversationID string) {
	if s == nil || strings.TrimSpace(key) == "" || strings.TrimSpace(conversationID) == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.entries) >= s.max {
		s.evictExpiredLocked()
	}
	if len(s.entries) >= s.max {
		// Still full: drop an arbitrary entry. Any victim only costs the next
		// turn of that conversation its server-side history.
		for existing := range s.entries {
			delete(s.entries, existing)
			break
		}
	}
	s.entries[key] = cquConversationEntry{ConversationID: conversationID, ExpiresAt: time.Now().Add(s.ttl)}
}

func (s *cquConversationStore) evictExpiredLocked() {
	now := time.Now()
	for key, entry := range s.entries {
		if now.After(entry.ExpiresAt) {
			delete(s.entries, key)
		}
	}
}
