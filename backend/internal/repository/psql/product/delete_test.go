package product_repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProductRepository_Delete(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	deleteSQL := `DELETE FROM "product" WHERE id = \$1`

	t.Run("deletes the row", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(deleteSQL).WithArgs(id).WillReturnResult(sqlmockNewResult())
		mock.ExpectCommit()

		assert.NoError(t, NewProductRepository(gdb).Delete(ctx, nil, id))
	})

	t.Run("propagates a database error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("connection refused")
		mock.ExpectBegin()
		mock.ExpectExec(deleteSQL).WithArgs(id).WillReturnError(dbErr)
		mock.ExpectRollback()

		assert.ErrorIs(t, NewProductRepository(gdb).Delete(ctx, nil, id), dbErr)
	})

	t.Run("uses the provided transaction", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(deleteSQL).WithArgs(id).WillReturnResult(sqlmockNewResult())
		mock.ExpectCommit()

		assert.NoError(t, NewProductRepository(gdb).Delete(ctx, wrapTx(gdb), id))
	})

	t.Run("returns an error for an unsupported Tx type", func(t *testing.T) {
		gdb, _ := newMockDB(t)
		assert.Error(t, NewProductRepository(gdb).Delete(ctx, unsupportedTx{}, id))
	})
}
