//go:build integration

package repository

import (
	"context"
	"testing"

	dbmigrations "github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

func TestMigration237RestoresModelDisplayConfigurationWithoutOverwritingAllowlist(t *testing.T) {
	tx := testTx(t)
	ctx := context.Background()

	// Simulate a database that has applied 236 but not 237.
	_, err := tx.ExecContext(ctx, "ALTER TABLE groups DROP COLUMN models_list_config")
	require.NoError(t, err)
	var id int64
	require.NoError(t, tx.QueryRowContext(ctx, `
INSERT INTO groups (name, platform, rate_multiplier, status, model_allowlist)
VALUES ('migration-237-restore', 'openai', 1, 'active', '{"enabled":true,"models":["gpt-test"]}'::jsonb)
RETURNING id`).Scan(&id))

	migrationSQL, err := dbmigrations.FS.ReadFile("237_restore_group_models_list_config.sql")
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, string(migrationSQL))
	require.NoError(t, err)

	var listed, allowed string
	require.NoError(t, tx.QueryRowContext(ctx,
		"SELECT models_list_config::text, model_allowlist::text FROM groups WHERE id = $1", id).
		Scan(&listed, &allowed))
	require.JSONEq(t, `{"enabled":true,"models":["gpt-test"]}`, listed)
	require.JSONEq(t, listed, allowed)

	// Once restored, future applies must preserve distinct administrator edits.
	_, err = tx.ExecContext(ctx,
		"UPDATE groups SET models_list_config = '{\"enabled\":false}'::jsonb WHERE id = $1", id)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, string(migrationSQL))
	require.NoError(t, err)
	require.NoError(t, tx.QueryRowContext(ctx,
		"SELECT models_list_config::text, model_allowlist::text FROM groups WHERE id = $1", id).
		Scan(&listed, &allowed))
	require.JSONEq(t, `{"enabled":false}`, listed)
	require.JSONEq(t, `{"enabled":true,"models":["gpt-test"]}`, allowed)
}
