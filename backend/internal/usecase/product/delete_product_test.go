package product_usecase

import (
	"context"
	"errors"
	"testing"

	"backend-template/domain/errorcode"
	"backend-template/domain/usecase"
	"backend-template/domain/valueobject"
	"backend-template/internal/apperr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteProduct(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db down")
	product := sampleProduct()
	allowed := []string{valueobject.PermissionDeleteProduct}

	tests := []struct {
		name    string
		input   usecase.DeleteProductInput
		setup   func(d deps)
		wantErr error
	}{
		{
			name:  "success",
			input: usecase.DeleteProductInput{ID: product.ID, Permissions: allowed},
			setup: func(d deps) {
				d.repo.EXPECT().FindByID(mock.Anything, mock.Anything, product.ID).Return(product, nil)
				d.repo.EXPECT().Delete(mock.Anything, mock.Anything, product.ID).Return(nil)
			},
		},
		{
			name:    "error permission denied",
			input:   usecase.DeleteProductInput{ID: product.ID},
			setup:   func(d deps) {},
			wantErr: apperr.Forbidden(errorcode.PermissionDenied),
		},
		{
			name:  "error not found",
			input: usecase.DeleteProductInput{ID: product.ID, Permissions: allowed},
			setup: func(d deps) {
				d.repo.EXPECT().FindByID(mock.Anything, mock.Anything, product.ID).Return(nil, nil)
			},
			wantErr: apperr.NotFound(errorcode.ProductNotFound),
		},
		{
			name:  "error repository find by id",
			input: usecase.DeleteProductInput{ID: product.ID, Permissions: allowed},
			setup: func(d deps) {
				d.repo.EXPECT().FindByID(mock.Anything, mock.Anything, product.ID).Return(nil, repoErr)
			},
			wantErr: repoErr,
		},
		{
			name:  "error repository delete",
			input: usecase.DeleteProductInput{ID: product.ID, Permissions: allowed},
			setup: func(d deps) {
				d.repo.EXPECT().FindByID(mock.Anything, mock.Anything, product.ID).Return(product, nil)
				d.repo.EXPECT().Delete(mock.Anything, mock.Anything, product.ID).Return(repoErr)
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setup(d)

			err := d.usecase().DeleteProduct(ctx, tt.input)

			assert.Equal(t, tt.wantErr, err)
		})
	}
}
