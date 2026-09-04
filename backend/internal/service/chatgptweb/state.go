package chatgptweb

import (
	"crypto/rand"
	"fmt"
	"time"
)

const RootParentMessageID = "client-created-root"

type ClientState struct {
	DeviceID        string    `json:"device_id"`
	SessionID       string    `json:"session_id"`
	UserAgent       string    `json:"user_agent"`
	StartedAt       time.Time `json:"started_at"`
	ConversationID  string    `json:"conversation_id,omitempty"`
	ParentMessageID string    `json:"parent_message_id"`
}

func NewClientState(deviceID, userAgent string, now time.Time) *ClientState {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if deviceID == "" {
		deviceID = newUUID()
	}
	return &ClientState{
		DeviceID:        deviceID,
		SessionID:       newUUID(),
		UserAgent:       userAgent,
		StartedAt:       now,
		ParentMessageID: RootParentMessageID,
	}
}

func (s *ClientState) NoteTurnResult(conversationID, parentMessageID string) {
	if s == nil {
		return
	}
	if conversationID != "" {
		s.ConversationID = conversationID
	}
	if parentMessageID != "" {
		s.ParentMessageID = parentMessageID
	}
}

func (s *ClientState) TimeSinceLoaded(now time.Time) time.Duration {
	if s == nil || s.StartedAt.IsZero() {
		return 0
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if now.Before(s.StartedAt) {
		return 0
	}
	return now.Sub(s.StartedAt)
}

type SessionState struct {
	Route           string    `json:"route"`
	AccountID       int64     `json:"account_id"`
	ConversationID  string    `json:"conversation_id,omitempty"`
	ParentMessageID string    `json:"parent_message_id"`
	ClientSessionID string    `json:"client_session_id"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (s *SessionState) NoteTurnResult(conversationID, parentMessageID string, now time.Time) {
	if s == nil {
		return
	}
	if conversationID != "" {
		s.ConversationID = conversationID
	}
	if parentMessageID != "" {
		s.ParentMessageID = parentMessageID
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	s.UpdatedAt = now
}

func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Sprintf("chatgptweb: generate UUID: %v", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
