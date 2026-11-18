package product

import (
	"testing"

	"backend-template/domain/errorcode"
	"backend-template/domain/usecase"
	"backend-template/domain/valueobject"
	"backend-template/internal/apperr"
	mocksusecase "backend-template/internal/mocks/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProductHandler_CreateProduct(t *testing.T) {
	body := `{"code":"P-001","name":"Widget","description":"a widget","price":10,"status":"ACTIVE"}`

	t.Run("creates and returns 201 with actor and permissions forwarded", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		p := sampleProduct()
		uc.EXPECT().CreateProduct(mock.Anything, mock.MatchedBy(func(in usecase.CreateProductInput) bool {
			return in.Code == "P-001" && in.Name == "Widget" && in.Description != nil && *in.Description == "a widget" &&
				in.Price == 10 && in.Status == valueobject.ProductStatusActive &&
				in.ActorName == "alice" && len(in.Permissions) == 1
		})).Return(usecase.CreateProductOutput{Product: p}, nil)

		app := newTestApp(NewProductHandler(uc), identityWith(valueobject.PermissionCreateProduct))
		resp := doRequest(t, app, "POST", "/products", body)
		require.Equal(t, fiber.StatusCreated, resp.StatusCode)

		data := decodeCommon(t, resp.Body).Data.(map[string]interface{})
		assert.Equal(t, p.ID.String(), data["id"])
	})

	t.Run("malformed JSON is a 400", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		resp := doRequest(t, newTestApp(NewProductHandler(uc), nil), "POST", "/products", `{"code":`)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid status is a 400", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		resp := doRequest(t, newTestApp(NewProductHandler(uc), nil), "POST", "/products", `{"code":"P","name":"N","status":"nope"}`)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("duplicate code maps to 409", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		uc.EXPECT().CreateProduct(mock.Anything, mock.Anything).
			Return(usecase.CreateProductOutput{}, apperr.Conflict(errorcode.ProductCodeAlreadyExists, "P-001"))

		resp := doRequest(t, newTestApp(NewProductHandler(uc), nil), "POST", "/products", body)
		assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
	})
}
