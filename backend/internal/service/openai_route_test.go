package service

import "testing"

func TestOpenAIRouteModeBackwardCompatibility(t *testing.T) {
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	if got := account.GetOpenAIRouteMode(); got != OpenAIRouteModeCodex {
		t.Fatalf("missing mode = %q", got)
	}
	if !account.SupportsOpenAIUpstreamRoute(OpenAIUpstreamRouteCodex) || account.SupportsOpenAIUpstreamRoute(OpenAIUpstreamRouteChatGPTWeb) {
		t.Fatal("legacy account must remain codex-only")
	}
}

func TestOpenAIRouteModeWebAndAuto(t *testing.T) {
	web := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Extra: map[string]any{OpenAIRouteModeExtraKey: "chatgpt_web"}}
	if !web.SupportsOpenAIUpstreamRoute(OpenAIUpstreamRouteChatGPTWeb) || web.SupportsOpenAIUpstreamRoute(OpenAIUpstreamRouteCodex) {
		t.Fatal("web account route isolation failed")
	}
	auto := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Extra: map[string]any{OpenAIRouteModeExtraKey: "auto"}}
	if !auto.SupportsOpenAIUpstreamRoute(OpenAIUpstreamRouteChatGPTWeb) || !auto.SupportsOpenAIUpstreamRoute(OpenAIUpstreamRouteCodex) {
		t.Fatal("auto account must be eligible for either route before session lock")
	}
}
