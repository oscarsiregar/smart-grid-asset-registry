package product_usecase

import (
	"context"
	"errors"
	"testing"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetListProduct(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db down")
	items := []*entity.Product{sampleProduct(), sampleProduct()}
	active := valueobject.ProductStatusActive
	allowed := []string{valueobject.PermissionGetListProduct}

	tests := []struct {
		name       string
		input      usecase.GetListProductInput
		setup      func(d deps)
		wantOutput usecase.GetListProductOutput
		wantErr    error
	}{
		{
			name:  "success applies defaults and passes filters through",
			input: usecase.GetListProductInput{Search: "wid", Status: &active, Permissions: allowed},
			setup: func(d deps) {
				d.repo.EXPECT().Find(mock.Anything, mock.Anything, repository.FindProductFilter{
					Search: "wid", Status: &active, Page: 1, Limit: 10, Sort: "created_at", Order: "DESC",
				}).Return(items, int64(2), nil)
			},
			wantOutput: usecase.GetListProductOutput{Items: items, TotalItems: 2},
		},
		{
			name:    "error permission denied",
			input:   usecase.GetListProductInput{Permissions: []string{"other.permission"}},
			setup:   func(d deps) {},
			wantErr: apperr.Forbidden(errorcode.PermissionDenied),
		},
		{
			name:  "error repository find",
			input: usecase.GetListProductInput{Permissions: allowed},
			setup: func(d deps) {
				d.repo.EXPECT().Find(mock.Anything, mock.Anything, mock.Anything).Return(nil, int64(0), repoErr)
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setup(d)

			output, err := d.usecase().GetListProduct(ctx, tt.input)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantOutput, output)
		})
	}
}
