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

func TestProductRepository_FindByCode(t *testing.T) {
	ctx := context.Background()
	selectSQL := `SELECT \* FROM "product" WHERE code = \$1 ORDER BY "product"\."id" LIMIT \$2`

	t.Run("maps the found row to an entity", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(selectSQL).WithArgs("P-001", 1).
			WillReturnRows(productRow(sqlmock.NewRows(productColumns), uuid.New(), "P-001"))

		result, err := NewProductRepository(gdb).FindByCode(ctx, nil, "P-001")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "P-001", result.Code)
	})

	t.Run("returns nil, nil when no row is found", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(selectSQL).WithArgs("P-404", 1).WillReturnRows(sqlmock.NewRows(productColumns))

		result, err := NewProductRepository(gdb).FindByCode(ctx, nil, "P-404")
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("propagates a database error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("connection refused")
		mock.ExpectQuery(selectSQL).WithArgs("P-001", 1).WillReturnError(dbErr)

		result, err := NewProductRepository(gdb).FindByCode(ctx, nil, "P-001")
		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("returns an error for an unsupported Tx type", func(t *testing.T) {
		gdb, _ := newMockDB(t)
		result, err := NewProductRepository(gdb).FindByCode(ctx, unsupportedTx{}, "P-001")
		assert.Nil(t, result)
		assert.Error(t, err)
	})
}
