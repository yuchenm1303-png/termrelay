package chatgptweb

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNativeSentinelProviderGeneratesRequirementsAndProofTokens(t *testing.T) {
	state := NewClientState("device-test", "ua-test", time.Now())
	provider := NativeSentinelProvider{}
	requirements, err := provider.RequirementsToken(context.Background(), state)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(requirements, requirementsTokenPrefix))
	require.True(t, strings.HasSuffix(requirements, sentinelTokenSuffix))

	proof, err := provider.Solve(context.Background(), ProofRequirement{Required: true, Seed: "seed", Difficulty: "ffffffff"}, state)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(proof, proofTokenPrefix))
	require.True(t, strings.HasSuffix(proof, sentinelTokenSuffix))
}

func TestNativeSentinelProviderHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (NativeSentinelProvider{}).Solve(ctx, ProofRequirement{Required: true, Seed: "seed", Difficulty: "00000000"}, nil)
	require.ErrorIs(t, err, context.Canceled)
}
