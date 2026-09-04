package service

import "strings"

type OpenAIUpstreamRoute string

type OpenAIRouteMode string

const (
	OpenAIUpstreamRouteCodex      OpenAIUpstreamRoute = "codex"
	OpenAIUpstreamRouteChatGPTWeb OpenAIUpstreamRoute = "chatgpt_web"

	OpenAIRouteModeCodex      OpenAIRouteMode = "codex"
	OpenAIRouteModeChatGPTWeb OpenAIRouteMode = "chatgpt_web"
	OpenAIRouteModeAuto       OpenAIRouteMode = "auto"

	OpenAIRouteModeExtraKey = "openai_route_mode"
)

func ParseOpenAIRouteMode(value string) OpenAIRouteMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(OpenAIRouteModeChatGPTWeb):
		return OpenAIRouteModeChatGPTWeb
	case string(OpenAIRouteModeAuto):
		return OpenAIRouteModeAuto
	default:
		return OpenAIRouteModeCodex
	}
}

func (a *Account) GetOpenAIRouteMode() OpenAIRouteMode {
	if a == nil || a.Platform != PlatformOpenAI || a.Extra == nil {
		return OpenAIRouteModeCodex
	}
	value, _ := a.Extra[OpenAIRouteModeExtraKey].(string)
	return ParseOpenAIRouteMode(value)
}

func (a *Account) SupportsOpenAIUpstreamRoute(route OpenAIUpstreamRoute) bool {
	if a == nil || a.Platform != PlatformOpenAI || !a.IsOAuth() {
		return false
	}
	switch a.GetOpenAIRouteMode() {
	case OpenAIRouteModeChatGPTWeb:
		return route == OpenAIUpstreamRouteChatGPTWeb
	case OpenAIRouteModeAuto:
		return route == OpenAIUpstreamRouteCodex || route == OpenAIUpstreamRouteChatGPTWeb
	default:
		return route == OpenAIUpstreamRouteCodex
	}
}
