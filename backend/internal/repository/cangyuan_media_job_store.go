package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const cangyuanMediaJobKeyPrefix = "cangyuan_media_job:"
const cangyuanMediaIdempotencyPrefix = "cangyuan_media_idempotency:"

type cangyuanMediaIdempotencyEntry struct {
	JobID       string `json:"job_id"`
	Fingerprint string `json:"fingerprint"`
}

// Redis is the shared task registry for every TermRelay instance. CAS avoids
// processing/completed regressions when two pollers observe the same task.
type cangyuanMediaJobStore struct {
	rdb *redis.Client
}

func NewCangyuanMediaJobStore(rdb *redis.Client) service.CangyuanMediaJobStore {
	return &cangyuanMediaJobStore{rdb: rdb}
}

func (s *cangyuanMediaJobStore) Create(ctx context.Context, job *service.CangyuanMediaJob, ttl time.Duration) error {
	if s == nil || s.rdb == nil || job == nil || job.ID == "" || ttl <= 0 {
		return errors.New("invalid cangyuan media job reservation")
	}
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	created, err := s.rdb.SetNX(ctx, cangyuanMediaJobKeyPrefix+job.ID, data, ttl).Result()
	if err != nil {
		return err
	}
	if !created {
		return service.ErrCangyuanMediaJobConflict
	}
	return nil
}

// CreateIdempotent atomically writes the private job and a client retry key.
// A matching retry returns the original job, including its submission state.
// Redis WATCH + MULTI/EXEC prevents concurrent identical requests from each
// creating a new local task on different TermRelay workers.
func (s *cangyuanMediaJobStore) CreateIdempotent(
	ctx context.Context,
	job *service.CangyuanMediaJob,
	keyHash string,
	fingerprint string,
	ttl time.Duration,
) (*service.CangyuanMediaJob, error) {
	if s == nil || s.rdb == nil || job == nil || job.ID == "" || ttl <= 0 ||
		len(keyHash) != 64 || len(fingerprint) != 64 ||
		strings.ContainsAny(keyHash+fingerprint, ": /") {
		return nil, errors.New("invalid cangyuan media idempotency reservation")
	}
	taskKey := cangyuanMediaJobKeyPrefix + job.ID
	dedupKey := cangyuanMediaIdempotencyPrefix + keyHash
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}
	mappingBytes, err := json.Marshal(cangyuanMediaIdempotencyEntry{JobID: job.ID, Fingerprint: fingerprint})
	if err != nil {
		return nil, err
	}

	for attempt := 0; attempt < 8; attempt++ {
		var recorded *service.CangyuanMediaJob
		err := s.rdb.Watch(ctx, func(tx *redis.Tx) error {
			raw, readErr := tx.Get(ctx, dedupKey).Bytes()
			if readErr == nil {
				var entry cangyuanMediaIdempotencyEntry
				if err := json.Unmarshal(raw, &entry); err != nil {
					return service.ErrCangyuanMediaIdempotencyConflict
				}
				if entry.Fingerprint != fingerprint || entry.JobID == "" {
					return service.ErrCangyuanMediaIdempotencyConflict
				}
				current, err := s.Get(ctx, entry.JobID)
				if err != nil {
					// An expired/corrupt task must never trigger a second POST
					// while its idempotency key still exists.
					return err
				}
				recorded = current
				return nil
			}
			if !errors.Is(readErr, redis.Nil) {
				return readErr
			}
			exists, err := tx.Exists(ctx, taskKey).Result()
			if err != nil {
				return err
			}
			if exists != 0 {
				return service.ErrCangyuanMediaJobConflict
			}
			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Set(ctx, taskKey, jobBytes, ttl)
				pipe.Set(ctx, dedupKey, mappingBytes, ttl)
				return nil
			})
			if err == nil {
				recorded = job
			}
			return err
		}, dedupKey, taskKey)
		if errors.Is(err, redis.TxFailedErr) {
			continue
		}
		return recorded, err
	}
	return nil, service.ErrCangyuanMediaJobConflict
}

func (s *cangyuanMediaJobStore) Get(ctx context.Context, id string) (*service.CangyuanMediaJob, error) {
	if s == nil || s.rdb == nil {
		return nil, errors.New("cangyuan media job store unavailable")
	}
	data, err := s.rdb.Get(ctx, cangyuanMediaJobKeyPrefix+id).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, service.ErrCangyuanMediaJobNotFound
	}
	if err != nil {
		return nil, err
	}
	var job service.CangyuanMediaJob
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, err
	}
	if job.ID != id {
		return nil, service.ErrCangyuanMediaJobNotFound
	}
	return &job, nil
}

func (s *cangyuanMediaJobStore) CompareAndSwap(ctx context.Context, job *service.CangyuanMediaJob, expectedVersion int64) error {
	if s == nil || s.rdb == nil || job == nil || job.ID == "" || expectedVersion < 1 || job.Version != expectedVersion+1 {
		return service.ErrCangyuanMediaJobConflict
	}
	key := cangyuanMediaJobKeyPrefix + job.ID
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	err = s.rdb.Watch(ctx, func(tx *redis.Tx) error {
		stored, readErr := tx.Get(ctx, key).Bytes()
		if errors.Is(readErr, redis.Nil) {
			return service.ErrCangyuanMediaJobNotFound
		}
		if readErr != nil {
			return readErr
		}
		var current service.CangyuanMediaJob
		if err := json.Unmarshal(stored, &current); err != nil {
			return err
		}
		// Identity, upstream account, pricing gate and original TTL are
		// immutable. The caller can only advance state and attach results.
		if current.Version != expectedVersion ||
			current.ID != job.ID || current.UserID != job.UserID ||
			current.APIKeyID != job.APIKeyID || current.AccountID != job.AccountID ||
			current.Kind != job.Kind || current.Model != job.Model ||
			current.PayloadSHA256 != job.PayloadSHA256 ||
			current.CreatedAt != job.CreatedAt || current.ExpiresAt != job.ExpiresAt {
			return service.ErrCangyuanMediaJobConflict
		}
		// A quote is fixed at reservation time. Neither a concurrent worker
		// nor the settlement path may change its unit amount or pricing tier.
		if (current.PriceSnapshot == nil) != (job.PriceSnapshot == nil) {
			return service.ErrCangyuanMediaJobConflict
		}
		if current.PriceSnapshot != nil && *current.PriceSnapshot != *job.PriceSnapshot {
			return service.ErrCangyuanMediaJobConflict
		}
		// The remote task ID must never be rebound to a different upstream
		// generation. A stale or compromised worker cannot silently replace
		// an already accepted billable task with another task ID.
		if current.UpstreamTaskID != "" && job.UpstreamTaskID != current.UpstreamTaskID {
			return service.ErrCangyuanMediaJobConflict
		}
		// Enforce one-way lifecycle constraints at the shared Redis boundary,
		// not only in the worker. This also protects against accidentally
		// replacing completed outputs after SQL settlement.
		switch current.Status {
		case service.CangyuanMediaJobCompleted, service.CangyuanMediaJobFailed,
			service.CangyuanMediaJobSubmitUnknown:
			if job.Status != current.Status || !equalCangyuanMediaResultURLs(current.ResultURLs, job.ResultURLs) {
				return service.ErrCangyuanMediaJobConflict
			}
		case service.CangyuanMediaJobProcessing:
			if job.Status == service.CangyuanMediaJobQueued ||
				job.Status == service.CangyuanMediaJobSubmitting ||
				job.Status == service.CangyuanMediaJobReserved {
				return service.ErrCangyuanMediaJobConflict
			}
		case service.CangyuanMediaJobQueued:
			if job.Status == service.CangyuanMediaJobSubmitting || job.Status == service.CangyuanMediaJobReserved {
				return service.ErrCangyuanMediaJobConflict
			}
		case service.CangyuanMediaJobSubmitting:
			if job.Status == service.CangyuanMediaJobReserved {
				return service.ErrCangyuanMediaJobConflict
			}
		}
		if current.BillingState != job.BillingState &&
			(current.BillingState != "pending" || job.BillingState != "settled" ||
				current.Status != service.CangyuanMediaJobCompleted || job.Status != service.CangyuanMediaJobCompleted ||
				current.PriceSnapshot == nil) {
			return service.ErrCangyuanMediaJobConflict
		}
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			// Retain TTL; polling must not extend job lifetime indefinitely.
			return pipe.Set(ctx, key, data, redis.KeepTTL).Err()
		})
		return err
	}, key)
	if errors.Is(err, redis.TxFailedErr) {
		return service.ErrCangyuanMediaJobConflict
	}
	return err
}

func equalCangyuanMediaResultURLs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
