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

func TestAssetRepository_FindByAssetIDs(t *testing.T) {
	ctx := context.Background()
	selectSQL := `SELECT \* FROM "asset" WHERE asset_id = ANY\(\$1\) ORDER BY asset_id$`
	ids := []string{"SS-1", "TX-1"}

	t.Run("sends the IDs as one array parameter and maps the rows", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		rows := sqlmock.NewRows(assetColumns)
		assetRow(rows, "SS-1", nil, "SUBSTATION", "S-1")
		assetRow(rows, "TX-1", "SS-1", "TRANSFORMER", "T-1")
		mock.ExpectQuery(selectSQL).WithArgs(`{"SS-1","TX-1"}`).WillReturnRows(rows)

		result, err := NewAssetRepository(gdb).FindByAssetIDs(ctx, nil, ids)

		require.NoError(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, "SS-1", result[0].AssetID)
		assert.Nil(t, result[0].ParentAssetID)
		assert.Equal(t, valueobject.AssetTypeTransformer, result[1].AssetType)
		require.NotNil(t, result[1].ParentAssetID)
		assert.Equal(t, "SS-1", *result[1].ParentAssetID)
	})

	t.Run("runs no query for an empty list", func(t *testing.T) {
		gdb, _ := newMockDB(t)

		result, err := NewAssetRepository(gdb).FindByAssetIDs(ctx, nil, nil)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("returns an empty, non-nil slice when nothing matches", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(selectSQL).WithArgs(`{"SS-1","TX-1"}`).WillReturnRows(sqlmock.NewRows(assetColumns))

		result, err := NewAssetRepository(gdb).FindByAssetIDs(ctx, nil, ids)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("propagates a database error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("connection refused")
		mock.ExpectQuery(selectSQL).WillReturnError(dbErr)

		result, err := NewAssetRepository(gdb).FindByAssetIDs(ctx, nil, ids)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("uses the given transaction", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(selectSQL).WillReturnRows(sqlmock.NewRows(assetColumns))

		_, err := NewAssetRepository(gdb).FindByAssetIDs(ctx, wrapTx(gdb), ids)

		assert.NoError(t, err)
	})

	t.Run("returns an error for an unsupported Tx type", func(t *testing.T) {
		gdb, _ := newMockDB(t)

		result, err := NewAssetRepository(gdb).FindByAssetIDs(ctx, unsupportedTx{}, ids)

		assert.Nil(t, result)
		assert.Error(t, err)
	})
}
