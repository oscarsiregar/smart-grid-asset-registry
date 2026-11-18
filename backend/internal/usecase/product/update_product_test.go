package product_usecase

import (
	"context"
	"errors"
	"testing"

	"backend-template/domain/entity"
	"backend-template/domain/errorcode"
	"backend-template/domain/usecase"
	"backend-template/domain/valueobject"
	"backend-template/internal/apperr"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateProduct(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db down")
	allowed := []string{valueobject.PermissionUpdateProduct}
	id := uuid.New()

	// existing returns a fresh copy per case: the usecase mutates it.
	existing := func() *entity.Product {
		p := sampleProduct()
		p.ID = id
		return p
	}
	input := func(code string) usecase.UpdateProductInput {
		return usecase.UpdateProductInput{
			ID: id, Code: code, Name: "Widget v2", Price: 20,
			Status: valueobject.ProductStatusInactive, ActorName: "bob", Permissions: allowed,
		}
	}

	tests := []struct {
		name    string
		input   usecase.UpdateProductInput
		setup   func(d deps)
		wantErr error
	}{
		{
			name:  "success same code skips uniqueness check",
			input: input("P-001"),
			setup: func(d deps) {
				d.withinTx()
				d.repo.EXPECT().FindByID(mock.Anything, d.trx, id).Return(existing(), nil)
				d.repo.EXPECT().Update(mock.Anything, d.trx, mock.MatchedBy(func(p *entity.Product) bool {
					return p.Name == "Widget v2" && p.Price == 20 &&
						p.Status == valueobject.ProductStatusInactive &&
						p.UpdatedAt.Equal(fixedNow) && p.UpdatedByName == "bob" &&
						p.CreatedByName == "alice" // created_* untouched
				})).Return(nil)
				d.trx.EXPECT().Commit().Return(nil)
			},
		},
		{
			name:  "success new unused code",
			input: input("P-002"),
			setup: func(d deps) {
				d.withinTx()
				d.repo.EXPECT().FindByID(mock.Anything, d.trx, id).Return(existing(), nil)
				d.repo.EXPECT().FindByCode(mock.Anything, d.trx, "P-002").Return(nil, nil)
				d.repo.EXPECT().Update(mock.Anything, d.trx, mock.Anything).Return(nil)
				d.trx.EXPECT().Commit().Return(nil)
			},
		},
		{
			name:    "error permission denied",
			input:   usecase.UpdateProductInput{ID: id},
			setup:   func(d deps) {},
			wantErr: apperr.Forbidden(errorcode.PermissionDenied),
		},
		{
			name:    "error invalid input",
			input:   usecase.UpdateProductInput{ID: id, Code: "P-001", Status: valueobject.ProductStatusActive, Permissions: allowed},
			setup:   func(d deps) {},
			wantErr: apperr.BadRequest(errorcode.InvalidRequest, "name is required"),
		},
		{
			name:  "error not found",
			input: input("P-001"),
			setup: func(d deps) {
				d.withinTx()
				d.repo.EXPECT().FindByID(mock.Anything, d.trx, id).Return(nil, nil)
				d.trx.EXPECT().Rollback().Return(nil)
			},
			wantErr: apperr.NotFound(errorcode.ProductNotFound),
		},
		{
			name:  "error repository find by id",
			input: input("P-001"),
			setup: func(d deps) {
				d.withinTx()
				d.repo.EXPECT().FindByID(mock.Anything, d.trx, id).Return(nil, repoErr)
				d.trx.EXPECT().Rollback().Return(nil)
			},
			wantErr: repoErr,
		},
		{
			name:  "error new code used by another product",
			input: input("P-002"),
			setup: func(d deps) {
				d.withinTx()
				d.repo.EXPECT().FindByID(mock.Anything, d.trx, id).Return(existing(), nil)
				d.repo.EXPECT().FindByCode(mock.Anything, d.trx, "P-002").Return(sampleProduct(), nil)
				d.trx.EXPECT().Rollback().Return(nil)
			},
			wantErr: apperr.Conflict(errorcode.ProductCodeAlreadyExists, "P-002"),
		},
		{
			name:  "error repository find by code",
			input: input("P-002"),
			setup: func(d deps) {
				d.withinTx()
				d.repo.EXPECT().FindByID(mock.Anything, d.trx, id).Return(existing(), nil)
				d.repo.EXPECT().FindByCode(mock.Anything, d.trx, "P-002").Return(nil, repoErr)
				d.trx.EXPECT().Rollback().Return(nil)
			},
			wantErr: repoErr,
		},
		{
			name:  "error repository update",
			input: input("P-001"),
			setup: func(d deps) {
				d.withinTx()
				d.repo.EXPECT().FindByID(mock.Anything, d.trx, id).Return(existing(), nil)
				d.repo.EXPECT().Update(mock.Anything, d.trx, mock.Anything).Return(repoErr)
				d.trx.EXPECT().Rollback().Return(nil)
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setup(d)

			output, err := d.usecase().UpdateProduct(ctx, tt.input)

			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				assert.Nil(t, output.Product)
				return
			}
			require.NotNil(t, output.Product)
			assert.Equal(t, tt.input.Code, output.Product.Code)
		})
	}
}
