package product

import (
	"errors"
	"testing"

	"backend-template/domain/entity"
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

func TestProductHandler_GetListProduct(t *testing.T) {
	t.Run("parses query, forwards permissions and maps the list", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		p := sampleProduct()
		page, limit := 2, 5
		active := valueobject.ProductStatusActive
		uc.EXPECT().GetListProduct(mock.Anything, usecase.GetListProductInput{
			Search: "wid", Status: &active, Page: &page, Limit: &limit, Sort: "name", Order: "asc",
			Permissions: []string{valueobject.PermissionGetListProduct},
		}).Return(usecase.GetListProductOutput{Items: []*entity.Product{p}, TotalItems: 6}, nil)

		app := newTestApp(NewProductHandler(uc), identityWith(valueobject.PermissionGetListProduct))
		resp := doRequest(t, app, "GET", "/products?search=wid&status=ACTIVE&page=2&limit=5&sort=name&order=asc", "")
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		common := decodeCommon(t, resp.Body)
		data := common.Data.(map[string]interface{})
		assert.Equal(t, float64(6), data["totalItems"])
		items := data["items"].([]interface{})
		require.Len(t, items, 1)
		item := items[0].(map[string]interface{})
		assert.Equal(t, p.ID.String(), item["id"])
		assert.Equal(t, "ACTIVE", item["status"])
		assert.Equal(t, "2026-01-02T03:04:05Z", item["createdAt"])
	})

	t.Run("empty result is an empty array, not null", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		uc.EXPECT().GetListProduct(mock.Anything, mock.Anything).Return(usecase.GetListProductOutput{}, nil)

		resp := doRequest(t, newTestApp(NewProductHandler(uc), nil), "GET", "/products", "")
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		data := decodeCommon(t, resp.Body).Data.(map[string]interface{})
		assert.Equal(t, []interface{}{}, data["items"])
	})

	t.Run("invalid status filter is a 400 and never reaches the usecase", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		resp := doRequest(t, newTestApp(NewProductHandler(uc), nil), "GET", "/products?status=bogus", "")
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("non-numeric page is a 400", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		resp := doRequest(t, newTestApp(NewProductHandler(uc), nil), "GET", "/products?page=abc", "")
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("unidentified caller passes an empty permission list", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		uc.EXPECT().GetListProduct(mock.Anything, usecase.GetListProductInput{Permissions: []string{}}).
			Return(usecase.GetListProductOutput{}, apperr.Forbidden(errorcode.PermissionDenied))

		resp := doRequest(t, newTestApp(NewProductHandler(uc), nil), "GET", "/products", "")
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		assert.Equal(t, "error", decodeCommon(t, resp.Body).Status)
	})

	t.Run("plain usecase error becomes a 500 without leaking details", func(t *testing.T) {
		uc := mocksusecase.NewProductUsecase(t)
		uc.EXPECT().GetListProduct(mock.Anything, mock.Anything).Return(usecase.GetListProductOutput{}, errors.New("pq: password=secret"))

		resp := doRequest(t, newTestApp(NewProductHandler(uc), nil), "GET", "/products", "")
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		assert.Equal(t, "Internal Server Error", decodeCommon(t, resp.Body).Message)
	})
}
