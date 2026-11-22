package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// recordSettledCangyuanMediaUsage is part of the SAME SQL transaction as the
// once-only quota settlement. Balance capture already happened separately:
// this routine must never debit money or increment usage counters.
//
// Only the dedicated, fingerprinted Cangyuan media settlement request ids
// qualify. Ordinary OpenAI image/video billing is unchanged.
func recordSettledCangyuanMediaUsage(ctx context.Context, tx *sql.Tx, cmd *service.UsageBillingCommand) error {
	if cmd == nil || !strings.HasPrefix(cmd.RequestID, "cangyuan-media:cyjob_") {
		return nil
	}
	if cmd.UserID <= 0 || cmd.APIKeyID <= 0 || cmd.AccountID <= 0 ||
		cmd.Model == "" || cmd.APIKeyQuotaCost <= 0 || cmd.BalanceCost != 0 ||
		cmd.SubscriptionCost != 0 {
		return errors.New("invalid settled Cangyuan media usage command")
	}

	imageCount, videoCount := 0, 0
	var imageSize any
	var imageSizeSource any
	switch cmd.MediaType {
	case "image":
		if cmd.ImageCount < 1 {
			return errors.New("Cangyuan image settlement must specify image count")
		}
		imageCount = cmd.ImageCount
		// The approved billable snapshot currently records the count, not the
		// final pixel dimensions. The SQL schema requires an image-size category
		// for every positive image_count; 'mixed' is the unclassified category,
		// not an assertion of a particular output resolution.
		imageSize = "mixed"
		imageSizeSource = "legacy"
	case "video":
		if cmd.ImageCount != 0 {
			return errors.New("Cangyuan video settlement cannot count images")
		}
		videoCount = 1
	default:
		return errors.New("invalid Cangyuan media kind for usage log")
	}

	// The provider-cost field contains the administrator's frozen estimated USD
	// cost, NOT a provider-verified bill. Reconciliation of upstream CNY invoices
	// must remain a separate process.
	result, err := tx.ExecContext(ctx, `
		INSERT INTO usage_logs (
			user_id, api_key_id, account_id, request_id,
			model, requested_model, upstream_model, group_id,
			total_cost, actual_cost, billing_mode, request_type,
			image_count, image_size, image_size_source, video_count,
			status, created_at, ended_at
		)
		SELECT $1, $2, $3, $4, $5, $5, $5, k.group_id,
			$6, $7, $8, 1, $9, $10, $11, $12,
			'success', NOW(), NOW()
		FROM api_keys k
		WHERE k.id = $2 AND k.user_id = $1 AND k.deleted_at IS NULL
		ON CONFLICT (request_id, api_key_id) DO NOTHING
	`, cmd.UserID, cmd.APIKeyID, cmd.AccountID, cmd.RequestID, cmd.Model,
		cmd.APIKeyQuotaCost, cmd.AccountQuotaCost, cmd.MediaType,
		imageCount, imageSize, imageSizeSource, videoCount)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		// A missing key/owner must not silently complete a financially
		// significant settlement. A preexisting matching log is idempotent.
		var exists bool
		err = tx.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM usage_logs WHERE request_id = $1
				AND api_key_id = $2 AND user_id = $3 AND account_id = $4
				AND model = $5 AND status = 'success'
			)
		`, cmd.RequestID, cmd.APIKeyID, cmd.UserID, cmd.AccountID, cmd.Model).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("settled Cangyuan media usage log was not persisted")
		}
	}
	return nil
}
