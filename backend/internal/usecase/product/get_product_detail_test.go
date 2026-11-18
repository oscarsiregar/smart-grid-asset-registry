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

func TestGetProductDetail(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db down")
	product := sampleProduct()
	allowed := []string{valueobject.PermissionGetProductDetail}

	tests := []struct {
		name       string
		input      usecase.GetProductDetailInput
		setup      func(d deps)
		wantOutput usecase.GetProductDetailOutput
		wantErr    error
	}{
		{
			name:  "success",
			input: usecase.GetProductDetailInput{ID: product.ID, Permissions: allowed},
			setup: func(d deps) {
				d.repo.EXPECT().FindByID(mock.Anything, mock.Anything, product.ID).Return(product, nil)
			},
			wantOutput: usecase.GetProductDetailOutput{Product: product},
		},
		{
			name:    "error permission denied",
			input:   usecase.GetProductDetailInput{ID: product.ID},
			setup:   func(d deps) {},
			wantErr: apperr.Forbidden(errorcode.PermissionDenied),
		},
		{
			name:  "error not found",
			input: usecase.GetProductDetailInput{ID: product.ID, Permissions: allowed},
			setup: func(d deps) {
				d.repo.EXPECT().FindByID(mock.Anything, mock.Anything, product.ID).Return(nil, nil)
			},
			wantErr: apperr.NotFound(errorcode.ProductNotFound),
		},
		{
			name:  "error repository find by id",
			input: usecase.GetProductDetailInput{ID: product.ID, Permissions: allowed},
			setup: func(d deps) {
				d.repo.EXPECT().FindByID(mock.Anything, mock.Anything, product.ID).Return(nil, repoErr)
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setup(d)

			output, err := d.usecase().GetProductDetail(ctx, tt.input)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantOutput, output)
		})
	}
}
