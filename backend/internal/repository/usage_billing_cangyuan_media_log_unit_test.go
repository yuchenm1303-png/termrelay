//go:build unit

package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestRecordSettledCangyuanMediaUsage_OneRecordByKind(t *testing.T) {
	for _, tc := range []struct {
		name      string
		kind      string
		count     int
		imgCount  int
		videoCount int
		imageSize any
		source    any
	}{
		{name: "image", kind: "image", count: 1, imgCount: 1, videoCount: 0, imageSize: "mixed", source: "legacy"},
		{name: "video", kind: "video", count: 0, imgCount: 0, videoCount: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func() { _ = db.Close() }()
			mock.ExpectBegin()
			tx, err := db.BeginTx(context.Background(), nil)
			require.NoError(t, err)
			mock.ExpectExec("INSERT INTO usage_logs").
				WithArgs(int64(11), int64(22), int64(33),
					"cangyuan-media:cyjob_fixture", "media-test-model",
					0.01, 0.008, tc.kind, tc.imgCount, tc.imageSize, tc.source, tc.videoCount).
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
			err = recordSettledCangyuanMediaUsage(context.Background(), tx, &service.UsageBillingCommand{
				RequestID: "cangyuan-media:cyjob_fixture", UserID: 11,
				APIKeyID: 22, AccountID: 33, Model: "media-test-model",
				MediaType: tc.kind, ImageCount: tc.count,
				APIKeyQuotaCost: 0.01, AccountQuotaCost: 0.008,
			})
			require.NoError(t, err)
			require.NoError(t, tx.Commit())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRecordSettledCangyuanMediaUsage_RejectsMissingLogAndRollsBack(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	mock.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	mock.ExpectExec("INSERT INTO usage_logs").
		WillReturnError(errors.New("db insert failed"))
	mock.ExpectRollback()
	err = recordSettledCangyuanMediaUsage(context.Background(), tx, &service.UsageBillingCommand{
		RequestID: "cangyuan-media:cyjob_fixture", UserID: 11,
		APIKeyID: 22, AccountID: 33, Model: "media-model",
		MediaType: "image", ImageCount: 1, APIKeyQuotaCost: 0.01,
	})
	require.ErrorContains(t, err, "db insert failed")
	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordSettledCangyuanMediaUsage_NoOrdinaryBillingSideEffects(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	mock.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	require.NoError(t, recordSettledCangyuanMediaUsage(context.Background(), tx, &service.UsageBillingCommand{
		RequestID: "ordinary-request", MediaType: "image",
	}))
	require.NoError(t, recordSettledCangyuanMediaUsage(context.Background(), tx, nil))
	mock.ExpectCommit()
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordSettledCangyuanMediaUsage_DuplicateIsIdempotent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	mock.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	mock.ExpectExec("INSERT INTO usage_logs").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("cangyuan-media:cyjob_fixture", int64(22), int64(11), int64(33), "media-model").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectCommit()
	err = recordSettledCangyuanMediaUsage(context.Background(), tx, &service.UsageBillingCommand{
		RequestID: "cangyuan-media:cyjob_fixture", UserID: 11,
		APIKeyID: 22, AccountID: 33, Model: "media-model",
		MediaType: "image", ImageCount: 1, APIKeyQuotaCost: 0.01,
	})
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}
