package asset

import (
	"errors"
	"testing"

	"smart-grid-asset-registry/domain/entity"
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

func withParent(id *string) any {
	return mock.MatchedBy(func(in usecase.GetListAssetInput) bool {
		if id == nil || in.ParentAssetID == nil {
			return id == nil && in.ParentAssetID == nil
		}
		return *in.ParentAssetID == *id
	})
}

func TestAssetHandler_GetListAsset(t *testing.T) {
	t.Run("no query lists the top level and answers the complete list", func(t *testing.T) {
		uc := mocksusecase.NewAssetUsecase(t)
		uc.EXPECT().GetListAsset(mock.Anything, withParent(nil)).Return(usecase.GetListAssetOutput{Assets: []*entity.AssetNode{
			{AssetID: "SS-001", AssetName: "Aurora", AssetType: valueobject.AssetTypeSubstation,
				OperationalStatus: pointer(valueobject.OperationalStatusInService), ChildCount: 5},
			{AssetID: "SS-002", AssetName: "Beacon", AssetType: valueobject.AssetTypeSubstation},
		}}, nil)

		resp := doGet(t, newTestApp(NewAssetHandler(uc)), "/api/assets")

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		body := decodeEnvelope(t, resp.Body)
		assert.Equal(t, "Successfully get list asset", body.Message)
		assert.JSONEq(t, `{
			"items": [
				{"assetId":"SS-001","assetName":"Aurora","assetType":"SUBSTATION","operationalStatus":"IN_SERVICE","childCount":5},
				{"assetId":"SS-002","assetName":"Beacon","assetType":"SUBSTATION","operationalStatus":null,"childCount":0}
			],
			"totalItems": 2
		}`, string(body.Data))
	})

	t.Run("an empty list is [] not null", func(t *testing.T) {
		uc := mocksusecase.NewAssetUsecase(t)
		uc.EXPECT().GetListAsset(mock.Anything, withParent(nil)).Return(usecase.GetListAssetOutput{}, nil)

		resp := doGet(t, newTestApp(NewAssetHandler(uc)), "/api/assets")

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.JSONEq(t, `{"items":[],"totalItems":0}`, string(decodeEnvelope(t, resp.Body).Data))
	})

	parentCases := []struct {
		name   string
		target string
		want   *string
	}{
		{"a plain parent ID", "/api/assets?parentAssetId=SS-001", pointer("SS-001")},
		{"an encoded parent ID is decoded", "/api/assets?parentAssetId=SS%2F1", pointer("SS/1")},
		{"an empty parent ID means the top level", "/api/assets?parentAssetId=", nil},
	}
	for _, tc := range parentCases {
		t.Run(tc.name, func(t *testing.T) {
			uc := mocksusecase.NewAssetUsecase(t)
			uc.EXPECT().GetListAsset(mock.Anything, withParent(tc.want)).Return(usecase.GetListAssetOutput{}, nil)

			resp := doGet(t, newTestApp(NewAssetHandler(uc)), tc.target)

			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		})
	}

	t.Run("an unknown parent passes through as 404 ASSET_NOT_FOUND", func(t *testing.T) {
		uc := mocksusecase.NewAssetUsecase(t)
		uc.EXPECT().GetListAsset(mock.Anything, mock.Anything).
			Return(usecase.GetListAssetOutput{}, apperr.NotFound(errorcode.AssetNotFound, "NOPE"))

		resp := doGet(t, newTestApp(NewAssetHandler(uc)), "/api/assets?parentAssetId=NOPE")

		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		body := decodeEnvelope(t, resp.Body)
		assert.Equal(t, "No asset with ID NOPE exists.", body.Message)
		assert.JSONEq(t, `{"errorCode":"ASSET_NOT_FOUND"}`, string(body.Data))
	})

	t.Run("a technical error is a 500 without detail", func(t *testing.T) {
		uc := mocksusecase.NewAssetUsecase(t)
		uc.EXPECT().GetListAsset(mock.Anything, mock.Anything).
			Return(usecase.GetListAssetOutput{}, errors.New("pq: connection refused"))

		resp := doGet(t, newTestApp(NewAssetHandler(uc)), "/api/assets")

		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		body := decodeEnvelope(t, resp.Body)
		assert.NotContains(t, body.Message, "connection refused")
	})
}
