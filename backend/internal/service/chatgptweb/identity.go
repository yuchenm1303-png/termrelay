package chatgptweb

import (
	"errors"
	"net/http"
	"strings"
)

// Identity contains upstream browser/account credentials only in process memory.
// Secret fields are deliberately unexported so generic JSON/log encoders cannot dump them.
type Identity struct {
	accessToken string
	cookie      string
	accountID   string
	deviceID    string
	userAgent   string
}

func NewIdentity(accessToken, accountID, deviceID, userAgent, cookie string) (*Identity, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return nil, errors.New("chatgptweb: access token is required")
	}
	return &Identity{accessToken: accessToken, accountID: strings.TrimSpace(accountID), deviceID: strings.TrimSpace(deviceID), userAgent: strings.TrimSpace(userAgent), cookie: strings.TrimSpace(cookie)}, nil
}
func (i *Identity) String() string   { return "chatgptweb.Identity{redacted}" }
func (i *Identity) GoString() string { return i.String() }
func (i *Identity) DeviceID() string {
	if i == nil {
		return ""
	}
	return i.deviceID
}
func (i *Identity) UserAgent() string {
	if i == nil {
		return ""
	}
	return i.userAgent
}
func (i *Identity) applyHeaders(h http.Header, state *ClientState) {
	if i == nil {
		return
	}
	h.Set("Authorization", "Bearer "+i.accessToken)
	deviceID := i.deviceID
	userAgent := i.userAgent
	if state != nil {
		if state.DeviceID != "" {
			deviceID = state.DeviceID
		}
		if state.UserAgent != "" {
			userAgent = state.UserAgent
		}
		if state.SessionID != "" {
			h.Set("Oai-Session-Id", state.SessionID)
		}
	}
	if deviceID != "" {
		h.Set("Oai-Device-Id", deviceID)
	}
	if userAgent != "" {
		h.Set("User-Agent", userAgent)
	}
	if i.accountID != "" {
		h.Set("Chatgpt-Account-Id", i.accountID)
	}
	if i.cookie != "" {
		h.Set("Cookie", i.cookie)
	}
}
