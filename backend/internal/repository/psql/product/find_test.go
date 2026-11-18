package product_repository

import (
	"context"
	"errors"
	"testing"

	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/valueobject"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductRepository_Find(t *testing.T) {
	ctx := context.Background()
	active := valueobject.ProductStatusActive

	t.Run("applies search, status, order and pagination", func(t *testing.T) {
		gdb, mock := newMockDB(t)

		// GORM adds its own parentheses around an OR expression, hence "((...))".
		mock.ExpectQuery(`SELECT count\(\*\) FROM "product" WHERE \(\(code ILIKE \$1 OR name ILIKE \$2\)\) AND status = \$3`).
			WithArgs(`%50\%%`, `%50\%%`, active).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(11))
		mock.ExpectQuery(`SELECT \* FROM "product" WHERE \(\(code ILIKE \$1 OR name ILIKE \$2\)\) AND status = \$3 ORDER BY name ASC LIMIT \$4 OFFSET \$5`).
			WithArgs(`%50\%%`, `%50\%%`, active, 10, 10).
			WillReturnRows(productRow(sqlmock.NewRows(productColumns), uuid.New(), "P-050"))

		items, total, err := NewProductRepository(gdb).Find(ctx, nil, repository.FindProductFilter{
			Search: "50%", Status: &active, Page: 2, Limit: 10, Sort: "name", Order: "asc",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(11), total)
		assert.Len(t, items, 1)
	})

	t.Run("no filters, unknown sort column falls back to created_at DESC", func(t *testing.T) {
		gdb, mock := newMockDB(t)

		mock.ExpectQuery(`SELECT count\(\*\) FROM "product"$`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery(`SELECT \* FROM "product" ORDER BY created_at DESC LIMIT \$1$`).
			WithArgs(10).
			WillReturnRows(sqlmock.NewRows(productColumns))

		items, total, err := NewProductRepository(gdb).Find(ctx, nil, repository.FindProductFilter{
			Page: 1, Limit: 10, Sort: "password_hash",
		})
		require.NoError(t, err)
		assert.Zero(t, total)
		assert.NotNil(t, items)
		assert.Empty(t, items)
	})

	t.Run("limit <= 0 returns every row", func(t *testing.T) {
		gdb, mock := newMockDB(t)

		mock.ExpectQuery(`SELECT count\(\*\) FROM "product"`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(`SELECT \* FROM "product" ORDER BY created_at DESC$`).
			WillReturnRows(productRow(sqlmock.NewRows(productColumns), uuid.New(), "P-001"))

		items, _, err := NewProductRepository(gdb).Find(ctx, wrapTx(gdb), repository.FindProductFilter{Limit: -1})
		require.NoError(t, err)
		assert.Len(t, items, 1)
	})

	t.Run("propagates a count error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("connection refused")
		mock.ExpectQuery(`SELECT count\(\*\) FROM "product"`).WillReturnError(dbErr)

		items, total, err := NewProductRepository(gdb).Find(ctx, nil, repository.FindProductFilter{Page: 1, Limit: 10})
		assert.ErrorIs(t, err, dbErr)
		assert.Nil(t, items)
		assert.Zero(t, total)
	})

	t.Run("propagates a select error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("connection refused")
		mock.ExpectQuery(`SELECT count\(\*\) FROM "product"`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
		mock.ExpectQuery(`SELECT \* FROM "product"`).WillReturnError(dbErr)

		items, total, err := NewProductRepository(gdb).Find(ctx, nil, repository.FindProductFilter{Page: 1, Limit: 10})
		assert.ErrorIs(t, err, dbErr)
		assert.Nil(t, items)
		assert.Zero(t, total)
	})

	t.Run("returns an error for an unsupported Tx type", func(t *testing.T) {
		gdb, _ := newMockDB(t)
		_, _, err := NewProductRepository(gdb).Find(ctx, unsupportedTx{}, repository.FindProductFilter{})
		assert.Error(t, err)
	})
}
