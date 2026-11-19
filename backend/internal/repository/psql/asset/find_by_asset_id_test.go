package asset_repository

import (
	"context"
	"errors"
	"testing"

	"smart-grid-asset-registry/domain/valueobject"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetRepository_FindByAssetID(t *testing.T) {
	ctx := context.Background()
	selectSQL := `SELECT \* FROM "asset" WHERE asset_id = \$1 LIMIT \$2$`

	t.Run("returns the stored asset", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		rows := assetRow(sqlmock.NewRows(assetColumns), "TX-1", "SS-1", "TRANSFORMER", "T-1")
		mock.ExpectQuery(selectSQL).WithArgs("TX-1", 1).WillReturnRows(rows)

		result, err := NewAssetRepository(gdb).FindByAssetID(ctx, nil, "TX-1")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "TX-1", result.AssetID)
		assert.Equal(t, valueobject.AssetTypeTransformer, result.AssetType)
		require.NotNil(t, result.ParentAssetID)
		assert.Equal(t, "SS-1", *result.ParentAssetID)
	})

	t.Run("returns nil, nil when no asset has the ID", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(selectSQL).WithArgs("NOPE", 1).WillReturnRows(sqlmock.NewRows(assetColumns))

		result, err := NewAssetRepository(gdb).FindByAssetID(ctx, nil, "NOPE")

		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("propagates a database error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("connection refused")
		mock.ExpectQuery(selectSQL).WillReturnError(dbErr)

		result, err := NewAssetRepository(gdb).FindByAssetID(ctx, nil, "TX-1")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("uses the given transaction", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(selectSQL).WillReturnRows(sqlmock.NewRows(assetColumns))

		_, err := NewAssetRepository(gdb).FindByAssetID(ctx, wrapTx(gdb), "TX-1")

		assert.NoError(t, err)
	})

	t.Run("returns an error for an unsupported Tx type", func(t *testing.T) {
		gdb, _ := newMockDB(t)

		result, err := NewAssetRepository(gdb).FindByAssetID(ctx, unsupportedTx{}, "TX-1")

		assert.Nil(t, result)
		assert.Error(t, err)
	})
}
