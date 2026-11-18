package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrCangyuanMediaIdempotencyConflict = errors.New("cangyuan media idempotency key reused with a different request")

// CangyuanMediaIdempotentJobStore creates both a tenant-specific idempotency
// mapping and the private task record in a single Redis transaction.
type CangyuanMediaIdempotentJobStore interface {
	CreateIdempotent(ctx context.Context, job *CangyuanMediaJob, keyHash string, fingerprint string, ttl time.Duration) (*CangyuanMediaJob, error)
}

// ReserveIdempotent makes client retries return the original task rather than
// creating a new upstream generation. The caller must require a non-empty
// Idempotency-Key header before offering a public paid media endpoint.
//
// No raw prompt, key or payload is persisted in the dedup mapping: only
// SHA-256 fingerprints and the existing private job ID are saved.
func (s *CangyuanMediaJobService) ReserveIdempotent(
	ctx context.Context,
	owner CangyuanMediaJobOwner,
	accountID int64,
	kind CangyuanMediaKind,
	model string,
	payload []byte,
	clientKey string,
	quote CangyuanMediaPriceSnapshot,
) (*CangyuanMediaJob, error) {
	store, ok := s.Store.(CangyuanMediaIdempotentJobStore)
	if !ok {
		return nil, errors.New("cangyuan media idempotency storage is required for paid requests")
	}
	if len(clientKey) == 0 || len(clientKey) > 128 || strings.TrimSpace(clientKey) != clientKey {
		return nil, errors.New("cangyuan media requests require a valid Idempotency-Key")
	}
	if len(payload) == 0 || len(payload) > cangyuanMediaRequestLimit {
		return nil, errors.New("cangyuan media request exceeds the supported size")
	}
	var request map[string]json.RawMessage
	if json.Unmarshal(payload, &request) != nil || request == nil {
		return nil, errors.New("cangyuan media request must be a JSON object")
	}
	var inputModel, prompt string
	if json.Unmarshal(request["model"], &inputModel) != nil || inputModel != model ||
		json.Unmarshal(request["prompt"], &prompt) != nil || strings.TrimSpace(prompt) == "" {
		return nil, errors.New("cangyuan media request model or prompt is invalid")
	}
	if err := validateCangyuanMediaReferences(request); err != nil {
		return nil, err
	}
	job, ttl, err := s.newReservation(owner, accountID, kind, model, quote)
	if err != nil {
		return nil, err
	}
	quotedBytes, err := json.Marshal(quote)
	if err != nil {
		return nil, fmt.Errorf("encode media price quote: %w", err)
	}
	hashKey := sha256.Sum256([]byte(fmt.Sprintf("%d:%d:%s", owner.UserID, owner.APIKeyID, clientKey)))
	fingerprintBytes := sha256.Sum256([]byte(
		fmt.Sprintf("%d:%s:%s:", accountID, kind, model) + string(payload) + ":" + string(quotedBytes),
	))
	existing, err := store.CreateIdempotent(ctx, job, hex.EncodeToString(hashKey[:]), hex.EncodeToString(fingerprintBytes[:]), ttl)
	if err != nil {
		return nil, err
	}
	if existing == nil || existing.UserID != owner.UserID || existing.APIKeyID != owner.APIKeyID ||
		existing.AccountID != accountID || existing.Kind != kind || existing.Model != model {
		return nil, ErrCangyuanMediaIdempotencyConflict
	}
	return existing, nil
}
