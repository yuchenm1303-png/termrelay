package chatgptweb

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
)

const (
	requirementsTokenPrefix = "gAAAAAC"
	proofTokenPrefix        = "gAAAAAB"
	sentinelTokenSuffix     = "~S"
	maxProofAttempts        = 500_000
)

// NativeSentinelProvider generates the browser-shaped requirements token and
// solves server-issued proof-of-work challenges locally. Interactive
// Turnstile/SO challenges are intentionally not bypassed; SentinelFlow returns
// a typed challenge error so the account can be handled explicitly.
type NativeSentinelProvider struct{}

var _ RequirementsTokenProvider = NativeSentinelProvider{}
var _ ProofOfWorkSolver = NativeSentinelProvider{}

func NewNativeRequirementsTokenProviderFactory() RequirementsTokenProviderFactory {
	return RequirementsTokenProviderFactoryFunc(func(context.Context, AccountRef, *AccountSession) (RequirementsTokenProvider, error) {
		return NativeSentinelProvider{}, nil
	})
}

func NewNativeProofOfWorkSolver() ProofOfWorkSolver { return NativeSentinelProvider{} }

func (NativeSentinelProvider) RequirementsToken(_ context.Context, state *ClientState) (string, error) {
	config := sentinelFingerprint(state, 1, rand.Float64())
	encoded, err := encodeSentinelConfig(config)
	if err != nil {
		return "", err
	}
	return requirementsTokenPrefix + encoded + sentinelTokenSuffix, nil
}

func (NativeSentinelProvider) Solve(ctx context.Context, requirement ProofRequirement, state *ClientState) (string, error) {
	seed := strings.TrimSpace(requirement.Seed)
	difficulty := strings.ToLower(strings.TrimSpace(requirement.Difficulty))
	if seed == "" || difficulty == "" || len(difficulty) > 8 {
		return "", errors.New("chatgptweb: invalid proof-of-work challenge")
	}
	started := time.Now()
	for nonce := 0; nonce < maxProofAttempts; nonce++ {
		if nonce%1024 == 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
			}
		}
		config := sentinelFingerprint(state, nonce, float64(time.Since(started).Milliseconds()))
		encoded, err := encodeSentinelConfig(config)
		if err != nil {
			return "", err
		}
		hash := sentinelFNV1a(seed + encoded)
		if hash[:len(difficulty)] <= difficulty {
			return proofTokenPrefix + encoded + sentinelTokenSuffix, nil
		}
	}
	return "", errors.New("chatgptweb: proof-of-work attempt limit exceeded")
}

func sentinelFingerprint(state *ClientState, nonce int, timing float64) []any {
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"
	deviceID := ""
	startedAt := time.Now().UTC()
	if state != nil {
		if strings.TrimSpace(state.UserAgent) != "" {
			userAgent = state.UserAgent
		}
		deviceID = state.DeviceID
		if !state.StartedAt.IsZero() {
			startedAt = state.StartedAt
		}
	}
	now := time.Now()
	return []any{
		3000,
		now.Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)"),
		4294967296,
		nonce,
		userAgent,
		"https://chatgpt.com/backend-api/sentinel/sdk.js",
		"",
		"en-US",
		"en-US,en",
		timing,
		"vendorSub",
		"onpointerrawupdate",
		"requestIdleCallback",
		float64(time.Since(startedAt).Milliseconds()),
		deviceID,
		"",
		8,
		float64(startedAt.UnixMilli()),
		0, 0, 0, 0, 0, 0, 0,
	}
}

func encodeSentinelConfig(config []any) (string, error) {
	payload, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("chatgptweb: encode sentinel fingerprint: %w", err)
	}
	return base64.StdEncoding.EncodeToString(payload), nil
}

func sentinelFNV1a(value string) string {
	hash := uint32(2166136261)
	for _, char := range value {
		hash ^= uint32(char)
		hash *= 16777619
	}
	hash ^= hash >> 16
	hash *= 2246822507
	hash ^= hash >> 13
	hash *= 3266489909
	hash ^= hash >> 16
	return fmt.Sprintf("%08x", hash)
}
