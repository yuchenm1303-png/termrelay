//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// Bug 1 regression: a channel with RestrictModels=true but no pricing rows for
// the requested group/platform used to reject every model, because an empty
// allow-list can never contain the requested model. Auto-generated catalogs hit
// exactly this state and every customer request died with 503
// "channel pricing restriction" before scheduling ever started.
//
// Semantics that must hold after the fix:
//   - no usable pricing row on the channel at all -> fail open (model allowed)
//   - pricing rows exist and contain the model -> allowed
//   - pricing rows exist but do NOT contain the model -> still restricted
//     (this is the whole point of restrict_models, do not regress it)

func TestIsModelRestricted_EmptyPricingFailsOpen(t *testing.T) {
	ch := Channel{
		ID:             1,
		Status:         StatusActive,
		GroupIDs:       []int64{10},
		RestrictModels: true,
	}
	repo := makeStandardRepo(ch, map[int64]string{10: "anthropic"})
	svc := newTestChannelService(repo)

	restricted := svc.IsModelRestricted(context.Background(), 10, "claude-opus-4")
	require.False(t, restricted, "empty pricing allow-list must fail open, not reject every model")
}

func TestIsModelRestricted_EmptyPricingRowFailsOpen(t *testing.T) {
	ch := Channel{
		ID:             1,
		Status:         StatusActive,
		GroupIDs:       []int64{10},
		RestrictModels: true,
		ModelPricing: []ChannelModelPricing{
			{Platform: "anthropic", Models: nil},
		},
	}
	repo := makeStandardRepo(ch, map[int64]string{10: "anthropic"})
	svc := newTestChannelService(repo)

	restricted := svc.IsModelRestricted(context.Background(), 10, "claude-opus-4")
	require.False(t, restricted, "a pricing row with an empty model list is still an empty allow-list")
}

func TestIsModelRestricted_PricingPresentModelListed(t *testing.T) {
	ch := Channel{
		ID:             1,
		Status:         StatusActive,
		GroupIDs:       []int64{10},
		RestrictModels: true,
		ModelPricing: []ChannelModelPricing{
			{Platform: "anthropic", Models: []string{"claude-opus-4"}},
		},
	}
	repo := makeStandardRepo(ch, map[int64]string{10: "anthropic"})
	svc := newTestChannelService(repo)

	restricted := svc.IsModelRestricted(context.Background(), 10, "claude-opus-4")
	require.False(t, restricted)
}

func TestIsModelRestricted_PricingPresentModelMissingStillRestricted(t *testing.T) {
	ch := Channel{
		ID:             1,
		Status:         StatusActive,
		GroupIDs:       []int64{10},
		RestrictModels: true,
		ModelPricing: []ChannelModelPricing{
			{Platform: "anthropic", Models: []string{"claude-opus-4"}},
		},
	}
	repo := makeStandardRepo(ch, map[int64]string{10: "anthropic"})
	svc := newTestChannelService(repo)

	// Red line: the fix must not turn a real allow-list into a no-op.
	restricted := svc.IsModelRestricted(context.Background(), 10, "gpt-5.5")
	require.True(t, restricted)
}

func TestIsModelRestricted_OtherPlatformPricingKeepsIsolation(t *testing.T) {
	// A channel that carries pricing rows for any platform keeps the strict
	// per-(group, platform) semantics: only a channel with no usable pricing
	// row at all may fail open, otherwise an unlisted platform would silently
	// bypass restrict_models and route to an unexpected upstream.
	ch := Channel{
		ID:             1,
		Status:         StatusActive,
		GroupIDs:       []int64{10, 11},
		RestrictModels: true,
		ModelPricing: []ChannelModelPricing{
			{Platform: "anthropic", Models: []string{"claude-opus-4"}},
		},
	}
	repo := makeStandardRepo(ch, map[int64]string{10: "anthropic", 11: "openai"})
	svc := newTestChannelService(repo)

	require.False(t, svc.IsModelRestricted(context.Background(), 10, "claude-opus-4"))
	// Group 11 is on openai and there is no openai pricing row: the allow-list
	// for that platform is empty while the channel does carry pricing, so the
	// strict isolation semantics apply -> restricted.
	require.True(t, svc.IsModelRestricted(context.Background(), 11, "gpt-5.5"),
		"channel carries pricing rows, so an empty allow-list for this platform must not fail open")
}

func TestIsModelRestricted_EmptyPricingFailsOpenForWildcardRequest(t *testing.T) {
	ch := Channel{
		ID:             1,
		Status:         StatusActive,
		GroupIDs:       []int64{10},
		RestrictModels: true,
	}
	repo := makeStandardRepo(ch, map[int64]string{10: "anthropic"})
	svc := newTestChannelService(repo)

	ctx := context.Background()
	for _, model := range []string{"gpt-5.5", "gpt-5.6-sol", "gpt-6-astra", "codex-auto-review", "Model-With-Dots.1"} {
		require.Falsef(t, svc.IsModelRestricted(ctx, 10, model), "model %q must not be blocked", model)
	}
}
