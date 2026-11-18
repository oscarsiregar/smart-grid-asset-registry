package product_repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	selectSQL := `SELECT \* FROM "product" WHERE id = \$1 ORDER BY "product"\."id" LIMIT \$2`

	t.Run("maps the found row to an entity", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(selectSQL).WithArgs(id, 1).
			WillReturnRows(productRow(sqlmock.NewRows(productColumns), id, "P-001"))

		result, err := NewProductRepository(gdb).FindByID(ctx, nil, id)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, id, result.ID)
		assert.Equal(t, "P-001", result.Code)
	})

	t.Run("returns nil, nil when no row is found", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(selectSQL).WithArgs(id, 1).WillReturnRows(sqlmock.NewRows(productColumns))

		result, err := NewProductRepository(gdb).FindByID(ctx, nil, id)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("propagates a database error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("connection refused")
		mock.ExpectQuery(selectSQL).WithArgs(id, 1).WillReturnError(dbErr)

		result, err := NewProductRepository(gdb).FindByID(ctx, nil, id)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("uses the provided transaction", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(selectSQL).WithArgs(id, 1).
			WillReturnRows(productRow(sqlmock.NewRows(productColumns), id, "P-001"))

		result, err := NewProductRepository(gdb).FindByID(ctx, wrapTx(gdb), id)
		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("returns an error for an unsupported Tx type", func(t *testing.T) {
		gdb, _ := newMockDB(t)
		result, err := NewProductRepository(gdb).FindByID(ctx, unsupportedTx{}, id)
		assert.Nil(t, result)
		assert.Error(t, err)
	})
}
