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

func TestProductHandler_UpdateProduct(t *testing.T) {
	body := `{"code":"P-001","name":"Widget v2","price":20,"status":"INACTIVE"}`

	t.Run("updates and returns 200", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		p := sampleProduct()
		uc.EXPECT().UpdateProduct(mock.Anything, mock.MatchedBy(func(in usecase.UpdateProductInput) bool {
			return in.ID == p.ID && in.Name == "Widget v2" && in.Status == valueobject.ProductStatusInactive && in.ActorName == "alice"
		})).Return(usecase.UpdateProductOutput{Product: p}, nil)

		app := newTestApp(NewProductHandler(uc), identityWith(valueobject.PermissionUpdateProduct))
		resp := doRequest(t, app, "PUT", "/products/"+p.ID.String(), body)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("non-UUID id is a 400", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		resp := doRequest(t, newTestApp(NewProductHandler(uc), nil), "PUT", "/products/xyz", body)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("malformed JSON is a 400", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		resp := doRequest(t, newTestApp(NewProductHandler(uc), nil), "PUT", "/products/"+sampleProduct().ID.String(), `{`)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid status is a 400", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		resp := doRequest(t, newTestApp(NewProductHandler(uc), nil), "PUT", "/products/"+sampleProduct().ID.String(), `{"status":"x"}`)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("not found maps to 404", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		uc.EXPECT().UpdateProduct(mock.Anything, mock.Anything).
			Return(usecase.UpdateProductOutput{}, apperr.NotFound(errorcode.ProductNotFound))

		resp := doRequest(t, newTestApp(NewProductHandler(uc), nil), "PUT", "/products/"+sampleProduct().ID.String(), body)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}
