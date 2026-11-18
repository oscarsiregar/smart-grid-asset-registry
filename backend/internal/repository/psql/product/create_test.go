package product_repository

import (
	"context"
	"errors"
	"testing"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/valueobject"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProductRepository_Create(t *testing.T) {
	ctx := context.Background()
	input := &entity.Product{ID: uuid.New(), Code: "P-001", Name: "Widget", Status: valueobject.ProductStatusActive}

	t.Run("inserts the row", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "product"`).WillReturnResult(sqlmockNewResult())
		mock.ExpectCommit()

		assert.NoError(t, NewProductRepository(gdb).Create(ctx, nil, input))
	})

	t.Run("propagates a database error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("unique violation")
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "product"`).WillReturnError(dbErr)
		mock.ExpectRollback()

		assert.ErrorIs(t, NewProductRepository(gdb).Create(ctx, nil, input), dbErr)
	})

	t.Run("uses the provided transaction", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "product"`).WillReturnResult(sqlmockNewResult())
		mock.ExpectCommit()

		assert.NoError(t, NewProductRepository(gdb).Create(ctx, wrapTx(gdb), input))
	})

	t.Run("returns an error for an unsupported Tx type", func(t *testing.T) {
		gdb, _ := newMockDB(t)
		assert.Error(t, NewProductRepository(gdb).Create(ctx, unsupportedTx{}, input))
	})
}
