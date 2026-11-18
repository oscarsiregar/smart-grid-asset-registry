package product_usecase

import (
	"context"
	"errors"
	"testing"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateProduct(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db down")
	allowed := []string{valueobject.PermissionCreateProduct}
	desc := "a widget"

	validInput := usecase.CreateProductInput{
		Code:        "  P-001 ",
		Name:        " Widget ",
		Description: &desc,
		Price:       10,
		Status:      valueobject.ProductStatusActive,
		ActorName:   "alice",
		Permissions: allowed,
	}

	tests := []struct {
		name    string
		input   usecase.CreateProductInput
		setup   func(d deps)
		wantErr error
	}{
		{
			name:  "success trims fields, stamps audit columns and commits",
			input: validInput,
			setup: func(d deps) {
				d.withinTx()
				d.repo.EXPECT().FindByCode(mock.Anything, d.trx, "P-001").Return(nil, nil)
				d.repo.EXPECT().Create(mock.Anything, d.trx, mock.MatchedBy(func(p *entity.Product) bool {
					return p.Code == "P-001" && p.Name == "Widget" &&
						p.CreatedAt.Equal(fixedNow) && p.CreatedByName == "alice" &&
						p.UpdatedAt.Equal(fixedNow) && p.UpdatedByName == "alice"
				})).Return(nil)
				d.trx.EXPECT().Commit().Return(nil)
			},
		},
		{
			name:    "error permission denied",
			input:   usecase.CreateProductInput{Code: "P-001"},
			setup:   func(d deps) {},
			wantErr: apperr.Forbidden(errorcode.PermissionDenied),
		},
		{
			name:    "error invalid input",
			input:   usecase.CreateProductInput{Name: "Widget", Status: valueobject.ProductStatusActive, Permissions: allowed},
			setup:   func(d deps) {},
			wantErr: apperr.BadRequest(errorcode.InvalidRequest, "code is required"),
		},
		{
			name:  "error code already exists rolls back",
			input: validInput,
			setup: func(d deps) {
				d.withinTx()
				d.repo.EXPECT().FindByCode(mock.Anything, d.trx, "P-001").Return(sampleProduct(), nil)
				d.trx.EXPECT().Rollback().Return(nil)
			},
			wantErr: apperr.Conflict(errorcode.ProductCodeAlreadyExists, "P-001"),
		},
		{
			name:  "error repository find by code rolls back",
			input: validInput,
			setup: func(d deps) {
				d.withinTx()
				d.repo.EXPECT().FindByCode(mock.Anything, d.trx, "P-001").Return(nil, repoErr)
				d.trx.EXPECT().Rollback().Return(nil)
			},
			wantErr: repoErr,
		},
		{
			name:  "error repository create rolls back",
			input: validInput,
			setup: func(d deps) {
				d.withinTx()
				d.repo.EXPECT().FindByCode(mock.Anything, d.trx, "P-001").Return(nil, nil)
				d.repo.EXPECT().Create(mock.Anything, d.trx, mock.Anything).Return(repoErr)
				d.trx.EXPECT().Rollback().Return(nil)
			},
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setup(d)

			output, err := d.usecase().CreateProduct(ctx, tt.input)

			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				assert.Nil(t, output.Product)
				return
			}
			require.NotNil(t, output.Product)
			assert.Equal(t, "P-001", output.Product.Code)
			assert.Equal(t, &desc, output.Product.Description)
		})
	}
}
