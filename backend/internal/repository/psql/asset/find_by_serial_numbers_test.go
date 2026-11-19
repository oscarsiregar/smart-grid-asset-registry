package asset_repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetRepository_FindBySerialNumbers(t *testing.T) {
	ctx := context.Background()
	selectSQL := `SELECT \* FROM "asset" WHERE serial_number = ANY\(\$1\) ORDER BY asset_id$`
	serials := []string{"SER-1"}

	t.Run("sends the serials as one array parameter and maps the rows", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		rows := sqlmock.NewRows(assetColumns)
		assetRow(rows, "TX-A", "SS-1", "TRANSFORMER", "SER-1")
		mock.ExpectQuery(selectSQL).WithArgs(`{"SER-1"}`).WillReturnRows(rows)

		result, err := NewAssetRepository(gdb).FindBySerialNumbers(ctx, nil, serials)

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "TX-A", result[0].AssetID)
		require.NotNil(t, result[0].SerialNumber)
		assert.Equal(t, "SER-1", *result[0].SerialNumber)
	})

	t.Run("runs no query for an empty list", func(t *testing.T) {
		gdb, _ := newMockDB(t)

		result, err := NewAssetRepository(gdb).FindBySerialNumbers(ctx, nil, []string{})

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("propagates a database error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("connection refused")
		mock.ExpectQuery(selectSQL).WillReturnError(dbErr)

		result, err := NewAssetRepository(gdb).FindBySerialNumbers(ctx, nil, serials)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("uses the given transaction", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(selectSQL).WillReturnRows(sqlmock.NewRows(assetColumns))

		_, err := NewAssetRepository(gdb).FindBySerialNumbers(ctx, wrapTx(gdb), serials)

		assert.NoError(t, err)
	})

	t.Run("returns an error for an unsupported Tx type", func(t *testing.T) {
		gdb, _ := newMockDB(t)

		result, err := NewAssetRepository(gdb).FindBySerialNumbers(ctx, unsupportedTx{}, serials)

		assert.Nil(t, result)
		assert.Error(t, err)
	})
}
