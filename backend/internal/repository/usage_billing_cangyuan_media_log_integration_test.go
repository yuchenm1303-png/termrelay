//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestUsageBillingRepositoryApply_CangyuanMediaWritesOneDurableUsageRow(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewUsageBillingRepository(client, integrationDB)

	user := mustCreateUser(t, client, &service.User{
		Email:        fmt.Sprintf("cangyuan-media-usage-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hash",
		Balance:      100,
	})
	apiKey := mustCreateApiKey(t, client, &service.APIKey{
		UserID: user.ID, Key: "sk-cangyuan-media-" + uuid.NewString(), Name: "media-audit",
		Quota: 0.10,
	})
	account := mustCreateAccount(t, client, &service.Account{
		Name: "cangyuan-media-audit-" + uuid.NewString(), Type: service.AccountTypeAPIKey,
	})

	requestID := "cangyuan-media:cyjob_" + uuid.NewString()
	// This test writes to the shared integration database. Remove its log even
	// on failed assertions so unrelated dashboard/range tests stay isolated.
	t.Cleanup(func() {
		_, cleanupErr := integrationDB.ExecContext(context.Background(),
			"DELETE FROM usage_logs WHERE request_id=$1 AND api_key_id=$2", requestID, apiKey.ID)
		require.NoError(t, cleanupErr)
	})
	command := &service.UsageBillingCommand{
		RequestID: requestID, UserID: user.ID, APIKeyID: apiKey.ID,
		AccountID: account.ID, AccountType: service.AccountTypeAPIKey,
		Model: "gpt-image-2.6", MediaType: "image", ImageCount: 1,
		BalanceCost: 0, APIKeyQuotaCost: 0.01,
		APIKeyRateLimitCost: 0.01, AccountQuotaCost: 0.008,
	}

	first, err := repo.Apply(ctx, command)
	require.NoError(t, err)
	require.True(t, first.Applied)
	again, err := repo.Apply(ctx, command)
	require.NoError(t, err)
	require.False(t, again.Applied)

	var rows, imageCount int
	var sale, estimatedCost, userBalance, usedQuota float64
	var size, mode string
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		SELECT COUNT(*), MAX(image_count), MAX(total_cost), MAX(actual_cost),
			MAX(image_size), MAX(billing_mode)
		FROM usage_logs WHERE request_id=$1 AND api_key_id=$2
	`, requestID, apiKey.ID).Scan(&rows, &imageCount, &sale, &estimatedCost, &size, &mode))
	require.Equal(t, 1, rows)
	require.Equal(t, 1, imageCount)
	require.Equal(t, "mixed", size)
	require.Equal(t, "image", mode)
	require.InDelta(t, 0.01, sale, 0.00000001)
	require.InDelta(t, 0.008, estimatedCost, 0.00000001)

	require.NoError(t, integrationDB.QueryRowContext(ctx,
		"SELECT balance FROM users WHERE id=$1", user.ID).Scan(&userBalance))
	require.NoError(t, integrationDB.QueryRowContext(ctx,
		"SELECT quota_used FROM api_keys WHERE id=$1", apiKey.ID).Scan(&usedQuota))
	// Capturing frozen money is a different idempotent transaction. A usage
	// log and quota settlement must not consume another cent from balance.
	require.InDelta(t, 100, userBalance, 0.00000001)
	require.InDelta(t, 0.01, usedQuota, 0.00000001)
}