package product_repository

import (
	"context"
	"errors"
	"testing"

	"backend-template/domain/entity"
	"backend-template/domain/valueobject"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProductRepository_Update(t *testing.T) {
	ctx := context.Background()
	input := &entity.Product{ID: uuid.New(), Code: "P-001", Name: "Widget", Status: valueobject.ProductStatusInactive}

	// Only mutable columns are written, keyed by primary key — no INSERT
	// fallback and created_* untouched.
	updateSQL := `UPDATE "product" SET "code"=\$1,"name"=\$2,"description"=\$3,"price"=\$4,"status"=\$5,"updated_at"=\$6,"updated_by_name"=\$7 WHERE "id" = \$8`

	t.Run("updates mutable columns of the row", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(updateSQL).WillReturnResult(sqlmockNewResult())
		mock.ExpectCommit()

		assert.NoError(t, NewProductRepository(gdb).Update(ctx, nil, input))
	})

	t.Run("propagates a database error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("connection refused")
		mock.ExpectBegin()
		mock.ExpectExec(updateSQL).WillReturnError(dbErr)
		mock.ExpectRollback()

		assert.ErrorIs(t, NewProductRepository(gdb).Update(ctx, nil, input), dbErr)
	})

	t.Run("uses the provided transaction", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(updateSQL).WillReturnResult(sqlmockNewResult())
		mock.ExpectCommit()

		assert.NoError(t, NewProductRepository(gdb).Update(ctx, wrapTx(gdb), input))
	})

	t.Run("returns an error for an unsupported Tx type", func(t *testing.T) {
		gdb, _ := newMockDB(t)
		assert.Error(t, NewProductRepository(gdb).Update(ctx, unsupportedTx{}, input))
	})
}
