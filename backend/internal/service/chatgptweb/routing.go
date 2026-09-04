package chatgptweb

import (
	"fmt"
	"strings"
)

// RoutingMode selects the high-level gateway path before provider-specific
// handling starts. The zero/empty input intentionally resolves to the existing
// provider path so introducing ChatGPT Web does not change current API routing.
type RoutingMode string

const (
	RoutingModeProvider   RoutingMode = "provider"
	RoutingModeChatGPTWeb RoutingMode = "chatgpt_web"
)

// ParseRoutingMode keeps the routing boundary explicit. New web-login
// providers can be added here without branching inside the existing provider
// handlers.
func ParseRoutingMode(value string) (RoutingMode, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", string(RoutingModeProvider):
		return RoutingModeProvider, nil
	case string(RoutingModeChatGPTWeb):
		return RoutingModeChatGPTWeb, nil
	default:
		return "", fmt.Errorf("chatgptweb: unsupported routing mode %q", value)
	}
}
