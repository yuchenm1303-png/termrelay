package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const cangyuanMediaJobKeyPrefix = "cangyuan_media_job:"

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
		if current.BillingState != job.BillingState {
			if !(current.BillingState == "pending" && job.BillingState == "settled" &&
				current.Status == service.CangyuanMediaJobCompleted && job.Status == service.CangyuanMediaJobCompleted &&
				current.PriceSnapshot != nil) {
				return service.ErrCangyuanMediaJobConflict
			}
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
