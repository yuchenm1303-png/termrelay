package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func cangyuanIdempotencyQuote() service.CangyuanMediaPriceSnapshot {
	return service.CangyuanMediaPriceSnapshot{
		PricingVersion:      "approved-2026-11-18",
		Currency:            "USD",
		Tier:                "720p-5s",
		Units:               1,
		UnitSalePriceUSD:    "2.50",
		UnitProviderCostUSD: "1.00",
		VerifiedByAdmin:     true,
	}
}

func TestCangyuanMediaIdempotentReservationReusesOriginalJobAndState(t *testing.T) {
	svc, mr := newCangyuanMediaJobTestService(t)
	ctx := context.Background()
	owner := service.CangyuanMediaJobOwner{UserID: 15, APIKeyID: 36}
	raw := []byte(`{"model":"video-model","prompt":"a quiet harbor","duration":5}`)
	quote := cangyuanIdempotencyQuote()
	first, err := svc.ReserveIdempotent(ctx, owner, 61, service.CangyuanVideoGeneration, "video-model", raw, "request-id-123", quote)
	require.NoError(t, err)
	require.Equal(t, service.CangyuanMediaJobReserved, first.Status)
	require.Equal(t, "pending", first.BillingState)
	require.NotNil(t, first.PriceSnapshot)
	require.Len(t, first.PayloadSHA256, 64, "the priced original request must be pinned")
	require.Equal(t, 48*time.Hour, mr.TTL(cangyuanMediaJobKeyPrefix+first.ID))

	_, err = svc.ClaimSubmission(ctx, owner, first.ID)
	require.NoError(t, err)
	_, err = svc.AttachSubmitted(ctx, owner, first.ID, &service.CangyuanMediaTask{ID: "remote_task_1", Status: "queued"})
	require.NoError(t, err)
	duplicate, err := svc.ReserveIdempotent(ctx, owner, 61, service.CangyuanVideoGeneration, "video-model", raw, "request-id-123", quote)
	require.NoError(t, err)
	require.Equal(t, first.ID, duplicate.ID)
	require.Equal(t, service.CangyuanMediaJobQueued, duplicate.Status)
	require.Equal(t, "remote_task_1", duplicate.UpstreamTaskID, "retry must reuse existing upstream job")

	_, err = svc.ReserveIdempotent(ctx, owner, 61, service.CangyuanVideoGeneration, "video-model",
		[]byte(`{"model":"video-model","prompt":"a DIFFERENT harbor","duration":5}`), "request-id-123", quote)
	require.ErrorIs(t, err, service.ErrCangyuanMediaIdempotencyConflict)
	changedPrice := quote
	changedPrice.UnitSalePriceUSD = "9.99"
	_, err = svc.ReserveIdempotent(ctx, owner, 61, service.CangyuanVideoGeneration, "video-model", raw, "request-id-123", changedPrice)
	require.ErrorIs(t, err, service.ErrCangyuanMediaIdempotencyConflict)

	otherUser := service.CangyuanMediaJobOwner{UserID: 16, APIKeyID: 36}
	another, err := svc.ReserveIdempotent(ctx, otherUser, 61, service.CangyuanVideoGeneration, "video-model", raw, "request-id-123", quote)
	require.NoError(t, err)
	require.NotEqual(t, first.ID, another.ID, "different tenants must not share an idempotency record")
}

func TestCangyuanMediaIdempotencyConcurrentSameKeySingleReservation(t *testing.T) {
	svc, _ := newCangyuanMediaJobTestService(t)
	owner := service.CangyuanMediaJobOwner{UserID: 5, APIKeyID: 9}
	quote := cangyuanIdempotencyQuote()
	raw := []byte(`{"model":"video-model","prompt":"morning commute","duration":5}`)
	const count = 12
	var wg sync.WaitGroup
	ids := make(chan string, count)
	errs := make(chan error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			job, err := svc.ReserveIdempotent(context.Background(), owner, 81, service.CangyuanVideoGeneration, "video-model",
				raw, "same-client-request", quote)
			if err != nil {
				errs <- err
				return
			}
			ids <- job.ID
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	var original string
	for id := range ids {
		if original == "" {
			original = id
		}
		require.Equal(t, original, id, "all concurrent client retries must reuse the same local task")
	}
	require.NotEmpty(t, original)
}

func TestCangyuanMediaIdempotencyValidatesKeyAndExpiresTogether(t *testing.T) {
	svc, mr := newCangyuanMediaJobTestService(t)
	owner := service.CangyuanMediaJobOwner{UserID: 7, APIKeyID: 8}
	raw := []byte(`{"model":"video-model","prompt":"scene","duration":5}`)
	quote := cangyuanIdempotencyQuote()
	for _, key := range []string{"", "  bad ", string(make([]byte, 129))} {
		_, err := svc.ReserveIdempotent(context.Background(), owner, 11, service.CangyuanVideoGeneration, "video-model", raw, key, quote)
		require.Error(t, err)
	}
	first, err := svc.ReserveIdempotent(context.Background(), owner, 11, service.CangyuanVideoGeneration, "video-model", raw, "unique-key", quote)
	require.NoError(t, err)
	mr.FastForward(49 * time.Hour)
	_, err = svc.Get(context.Background(), owner, first.ID)
	require.ErrorIs(t, err, service.ErrCangyuanMediaJobNotFound)
	next, err := svc.ReserveIdempotent(context.Background(), owner, 11, service.CangyuanVideoGeneration, "video-model", raw, "unique-key", quote)
	require.NoError(t, err)
	require.NotEqual(t, first.ID, next.ID, "dedup lifespan is bounded by the retention TTL")
}

func TestCangyuanMediaPaidImageReservationRejectsUnderquotedOutputs(t *testing.T) {
	svc, mr := newCangyuanMediaJobTestService(t)
	owner := service.CangyuanMediaJobOwner{UserID: 83, APIKeyID: 91}
	quote := cangyuanIdempotencyQuote()
	quote.Tier = "standard"
	raw := []byte(`{"model":"image-model","prompt":"three cats","n":3}`)
	_, err := svc.ReserveIdempotent(context.Background(), owner, 61, service.CangyuanImageGeneration, "image-model", raw, "request-3-images", quote)
	require.ErrorContains(t, err, "does not match approved price")
	require.Empty(t, mr.Keys(), "a mispriced request must not reserve a paid task")

	quote.Units = 3
	job, err := svc.ReserveIdempotent(context.Background(), owner, 61, service.CangyuanImageGeneration, "image-model", raw, "request-3-images", quote)
	require.NoError(t, err)
	require.Equal(t, "pending", job.BillingState)
	require.Equal(t, 3, job.PriceSnapshot.Units)
}
