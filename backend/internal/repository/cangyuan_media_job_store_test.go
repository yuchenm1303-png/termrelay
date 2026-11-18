package repository

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func newCangyuanMediaJobTestService(t *testing.T) (*service.CangyuanMediaJobService, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return &service.CangyuanMediaJobService{Store: NewCangyuanMediaJobStore(rdb), TTL: 48 * time.Hour}, mr
}

func TestCangyuanMediaJobStoreOwnershipAndDurableStatus(t *testing.T) {
	svc, mr := newCangyuanMediaJobTestService(t)
	ctx := context.Background()
	owner := service.CangyuanMediaJobOwner{UserID: 13, APIKeyID: 17}
	job, err := svc.Reserve(ctx, owner, 91, service.CangyuanImageGeneration, "image-model")
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobReserved, job.Status)
	require.Equal(t, "disabled", job.BillingState)
	require.Empty(t, job.UpstreamTaskID)
	require.Equal(t, 48*time.Hour, mr.TTL(cangyuanMediaJobKeyPrefix+job.ID))

	_, err = svc.Get(ctx, service.CangyuanMediaJobOwner{UserID: 13, APIKeyID: 18}, job.ID)
	require.ErrorIs(t, err, service.ErrCangyuanMediaJobNotFound)
	_, err = svc.Get(ctx, service.CangyuanMediaJobOwner{UserID: 18, APIKeyID: 17}, job.ID)
	require.ErrorIs(t, err, service.ErrCangyuanMediaJobNotFound)

	attached, err := svc.AttachSubmitted(ctx, owner, job.ID, &service.CangyuanMediaTask{TaskID: "remote_task_123", Status: "queued"})
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobQueued, attached.Status)
	require.Equal(t, "remote_task_123", attached.UpstreamTaskID)
	require.Equal(t, int64(2), attached.Version)
	require.Equal(t, 48*time.Hour, mr.TTL(cangyuanMediaJobKeyPrefix+job.ID), "polling must not refresh TTL")

	_, err = svc.Observe(ctx, owner, job.ID, &service.CangyuanMediaTask{ID: "someone_elses_task", Status: "completed"})
	require.ErrorIs(t, err, service.ErrCangyuanMediaJobState)

	// Use a real decoded completed payload to exercise ResultURLs.
	var done service.CangyuanMediaTask
	require.NoError(t, json.Unmarshal([]byte(`{"id":"remote_task_123","status":"completed","data":[{"url":"https://cdn.example.com/image.png"}]}`), &done))
	completed, err := svc.Observe(ctx, owner, job.ID, &done)
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobCompleted, completed.Status)
	require.Equal(t, []string{"https://cdn.example.com/image.png"}, completed.ResultURLs)
	require.Equal(t, "disabled", completed.BillingState, "completion must not silently charge a user")

	_, err = svc.Observe(ctx, owner, job.ID, &service.CangyuanMediaTask{ID: "remote_task_123", Status: "in_progress"})
	require.ErrorIs(t, err, service.ErrCangyuanMediaJobState)

	again, err := svc.Observe(ctx, owner, job.ID, &done)
	require.NoError(t, err)
	require.Equal(t, completed.Version, again.Version, "duplicate completion must not rewrite the record")
}

func TestCangyuanMediaJobStoreCASRejectsStaleChanges(t *testing.T) {
	svc, mr := newCangyuanMediaJobTestService(t)
	ctx := context.Background()
	owner := service.CangyuanMediaJobOwner{UserID: 1, APIKeyID: 2}
	job, err := svc.Reserve(ctx, owner, 11, service.CangyuanVideoGeneration, "video-model")
	require.NoError(t, err)
	a, err := svc.Store.Get(ctx, job.ID)
	require.NoError(t, err)
	b, err := svc.Store.Get(ctx, job.ID)
	require.NoError(t, err)
	a.Status = service.CangyuanMediaJobQueued
	a.UpstreamTaskID = "task_A"
	a.Version++
	require.NoError(t, svc.Store.CompareAndSwap(ctx, a, 1))

	b.Status = service.CangyuanMediaJobCompleted
	b.UpstreamTaskID = "task_B"
	b.Version++
	require.ErrorIs(t, svc.Store.CompareAndSwap(ctx, b, 1), service.ErrCangyuanMediaJobConflict)

	got, err := svc.Get(ctx, owner, job.ID)
	require.NoError(t, err)
	require.Equal(t, "task_A", got.UpstreamTaskID)
	require.Equal(t, service.CangyuanMediaJobQueued, got.Status)
	require.Equal(t, "disabled", got.BillingState)
	require.Equal(t, 48*time.Hour, mr.TTL(cangyuanMediaJobKeyPrefix+job.ID))

	// Ownership, model and account are immutable even for a CAS with a
	// current version; this is enforced in Redis, not only service memory.
	bad := *got
	bad.AccountID++
	bad.Version++
	require.ErrorIs(t, svc.Store.CompareAndSwap(ctx, &bad, got.Version), service.ErrCangyuanMediaJobConflict)
	bad = *got
	bad.BillingState = "charged"
	bad.Version++
	require.ErrorIs(t, svc.Store.CompareAndSwap(ctx, &bad, got.Version), service.ErrCangyuanMediaJobConflict)
}

func TestCangyuanMediaJobStoreUncertainSubmitAndExpiry(t *testing.T) {
	svc, mr := newCangyuanMediaJobTestService(t)
	ctx := context.Background()
	owner := service.CangyuanMediaJobOwner{UserID: 7, APIKeyID: 8}
	job, err := svc.Reserve(ctx, owner, 6, service.CangyuanVideoGeneration, "video-model")
	require.NoError(t, err)
	unknown, err := svc.MarkSubmissionUnknown(ctx, owner, job.ID)
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobSubmitUnknown, unknown.Status)
	require.Empty(t, unknown.UpstreamTaskID)

	_, err = svc.AttachSubmitted(ctx, owner, job.ID, &service.CangyuanMediaTask{ID: "new_task", Status: "queued"})
	require.ErrorIs(t, err, service.ErrCangyuanMediaJobState, "uncertain POST cannot be retried or replaced")
	_, err = svc.Observe(ctx, owner, job.ID, &service.CangyuanMediaTask{ID: "new_task", Status: "completed"})
	require.ErrorIs(t, err, service.ErrCangyuanMediaJobState)

	mr.FastForward(49 * time.Hour)
	_, err = svc.Get(ctx, owner, job.ID)
	require.ErrorIs(t, err, service.ErrCangyuanMediaJobNotFound)

	_, err = svc.Reserve(ctx, service.CangyuanMediaJobOwner{}, 6, service.CangyuanVideoGeneration, "video-model")
	require.Error(t, err)
	_, err = svc.Reserve(ctx, owner, 6, "unrecognized", "video-model")
	require.Error(t, err)
	if errors.Is(err, service.ErrCangyuanMediaJobNotFound) {
		t.Fatal("validation errors must not masquerade as missing tasks")
	}
}
