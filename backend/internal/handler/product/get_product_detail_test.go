package product

import (
	"testing"

	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"
	mocksusecase "smart-grid-asset-registry/internal/mocks/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProductHandler_GetProductDetail(t *testing.T) {
	t.Run("returns the product", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		p := sampleProduct()
		uc.EXPECT().GetProductDetail(mock.Anything, usecase.GetProductDetailInput{
			ID: p.ID, Permissions: []string{valueobject.PermissionGetProductDetail},
		}).Return(usecase.GetProductDetailOutput{Product: p}, nil)

		app := newTestApp(NewProductHandler(uc), identityWith(valueobject.PermissionGetProductDetail))
		resp := doRequest(t, app, "GET", "/products/"+p.ID.String(), "")
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		data := decodeCommon(t, resp.Body).Data.(map[string]interface{})
		assert.Equal(t, "P-001", data["code"])
	})

	t.Run("non-UUID id is a 400", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		resp := doRequest(t, newTestApp(NewProductHandler(uc), nil), "GET", "/products/not-a-uuid", "")
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("not found maps to 404", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		uc.EXPECT().GetProductDetail(mock.Anything, mock.Anything).
			Return(usecase.GetProductDetailOutput{}, apperr.NotFound(errorcode.ProductNotFound))

		resp := doRequest(t, newTestApp(NewProductHandler(uc), nil), "GET", "/products/"+sampleProduct().ID.String(), "")
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		assert.Equal(t, "Product not found", decodeCommon(t, resp.Body).Message)
	})
}
