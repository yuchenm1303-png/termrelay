package service

import "testing"

func TestProviderIdentityGroupsAccountsByOriginWithoutSecrets(t *testing.T) {
	first := &Account{
		Platform: PlatformOpenAI,
		Credentials: map[string]any{
			"base_url": "https://relay.example.com/v1/",
			"api_key":  "secret-one",
		},
	}
	second := &Account{
		Platform: PlatformOpenAI,
		Credentials: map[string]any{
			"base_url": "https://RELAY.EXAMPLE.COM/v1",
			"api_key":  "secret-two",
		},
	}

	left := first.ProviderIdentity()
	right := second.ProviderIdentity()
	if left.Key == "" || left.Key != right.Key {
		t.Fatalf("same provider origin should share identity: left=%+v right=%+v", left, right)
	}
	if left.Origin != "https://relay.example.com/v1" {
		t.Fatalf("unexpected normalized origin: %q", left.Origin)
	}
	if left.Key == "secret-one" || right.Key == "secret-two" {
		t.Fatal("provider identity must never contain API keys")
	}
}

func TestProviderIdentitySeparatesDifferentOrigins(t *testing.T) {
	first := (&Account{Platform: PlatformOpenAI, Credentials: map[string]any{"base_url": "https://a.example.com/v1"}}).ProviderIdentity()
	second := (&Account{Platform: PlatformOpenAI, Credentials: map[string]any{"base_url": "https://b.example.com/v1"}}).ProviderIdentity()
	if first.Key == second.Key {
		t.Fatalf("different provider origins must not share identity: %+v %+v", first, second)
	}
}

func TestProviderIdentityExplicitProviderIDOverridesOrigin(t *testing.T) {
	first := (&Account{
		Platform: PlatformOpenAI,
		Extra:    map[string]any{accountPoolProviderIDExtraKey: "llmgw"},
		Credentials: map[string]any{
			"base_url": "https://one.example.com/v1",
		},
	}).ProviderIdentity()
	second := (&Account{
		Platform: PlatformOpenAI,
		Extra:    map[string]any{accountPoolProviderIDExtraKey: "llmgw"},
		Credentials: map[string]any{
			"base_url": "https://two.example.com/v1",
		},
	}).ProviderIdentity()
	if first.Key != "provider:llmgw" || second.Key != first.Key {
		t.Fatalf("explicit provider id should group accounts: first=%+v second=%+v", first, second)
	}
}
