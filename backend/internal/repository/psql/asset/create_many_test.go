package asset_repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/valueobject"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const insertSQL = `INSERT INTO "asset" \("id","asset_id","parent_asset_id","asset_type","asset_name",` +
	`"operational_status","voltage_kv","rating_kva","manufacturer","model","serial_number",` +
	`"commissioned_date","import_id","created_at","created_by_name","updated_at","updated_by_name"\) ` +
	`VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8,\$9,\$10,\$11,\$12,\$13,\$14,\$15,\$16,\$17\),` +
	`\(\$18,\$19,\$20,\$21,\$22,\$23,\$24,\$25,\$26,\$27,\$28,\$29,\$30,\$31,\$32,\$33,\$34\)$`

func newAssets() []*entity.Asset {
	now := time.Date(2026, time.November, 19, 10, 0, 0, 0, time.UTC)
	importID := uuid.New()
	parent := "SS-1"
	return []*entity.Asset{
		{ID: uuid.New(), AssetID: "SS-1", AssetType: valueobject.AssetTypeSubstation, AssetName: "Sub", ImportID: importID,
			CreatedAt: now, CreatedByName: "system", UpdatedAt: now, UpdatedByName: "system"},
		{ID: uuid.New(), AssetID: "TX-1", ParentAssetID: &parent, AssetType: valueobject.AssetTypeTransformer, AssetName: "Tx",
			ImportID: importID, CreatedAt: now, CreatedByName: "system", UpdatedAt: now, UpdatedByName: "system"},
	}
}

func TestAssetRepository_CreateMany(t *testing.T) {
	ctx := context.Background()

	t.Run("inserts every column in one batch, in the given order", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		assets := newAssets()
		mock.ExpectBegin()
		mock.ExpectExec(insertSQL).
			WithArgs(assets[0].ID, "SS-1", nil, valueobject.AssetTypeSubstation, "Sub", nil, nil, nil, nil, nil, nil, nil,
				assets[0].ImportID, assets[0].CreatedAt, "system", assets[0].UpdatedAt, "system",
				assets[1].ID, "TX-1", "SS-1", valueobject.AssetTypeTransformer, "Tx", nil, nil, nil, nil, nil, nil, nil,
				assets[1].ImportID, assets[1].CreatedAt, "system", assets[1].UpdatedAt, "system").
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		assert.NoError(t, NewAssetRepository(gdb).CreateMany(ctx, nil, assets))
	})

	t.Run("runs no query for an empty list", func(t *testing.T) {
		gdb, _ := newMockDB(t)

		assert.NoError(t, NewAssetRepository(gdb).CreateMany(ctx, nil, nil))
	})

	t.Run("propagates a database error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("duplicate key value violates unique constraint")
		mock.ExpectBegin()
		mock.ExpectExec(insertSQL).WillReturnError(dbErr)
		mock.ExpectRollback()

		err := NewAssetRepository(gdb).CreateMany(ctx, nil, newAssets())

		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("uses the given transaction", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(insertSQL).WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		assert.NoError(t, NewAssetRepository(gdb).CreateMany(ctx, wrapTx(gdb), newAssets()))
	})

	t.Run("returns an error for an unsupported Tx type", func(t *testing.T) {
		gdb, _ := newMockDB(t)

		assert.Error(t, NewAssetRepository(gdb).CreateMany(ctx, unsupportedTx{}, newAssets()))
	})
}
