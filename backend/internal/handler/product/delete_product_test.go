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
)

func TestProductHandler_DeleteProduct(t *testing.T) {
	t.Run("deletes and returns 204", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		p := sampleProduct()
		uc.EXPECT().DeleteProduct(mock.Anything, usecase.DeleteProductInput{
			ID: p.ID, Permissions: []string{valueobject.PermissionDeleteProduct},
		}).Return(nil)

		app := newTestApp(NewProductHandler(uc), identityWith(valueobject.PermissionDeleteProduct))
		resp := doRequest(t, app, "DELETE", "/products/"+p.ID.String(), "")
		assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
	})

	t.Run("non-UUID id is a 400", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		resp := doRequest(t, newTestApp(NewProductHandler(uc), nil), "DELETE", "/products/xyz", "")
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("forbidden maps to 403", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		uc.EXPECT().DeleteProduct(mock.Anything, mock.Anything).Return(apperr.Forbidden(errorcode.PermissionDenied))

		resp := doRequest(t, newTestApp(NewProductHandler(uc), nil), "DELETE", "/products/"+sampleProduct().ID.String(), "")
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})
}
