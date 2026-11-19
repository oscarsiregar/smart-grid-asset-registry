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

func withSearch(text string, page, limit int) any {
	return mock.MatchedBy(func(in usecase.GetListAssetInput) bool {
		return in.Search == text && in.Page == page && in.Limit == limit && in.ParentAssetID == nil
	})
}

func TestAssetHandler_GetListAsset_Search(t *testing.T) {
	t.Run("trims and decodes the text and answers one page of matches", func(t *testing.T) {
		uc := mocksusecase.NewAssetUsecase(t)
		uc.EXPECT().GetListAsset(mock.Anything, withSearch("PNL/1", 0, 0)).Return(usecase.GetListAssetOutput{
			Assets: []*entity.AssetNode{
				{AssetID: "PNL/1-A", AssetName: "Panel", AssetType: valueobject.AssetTypeSwitchboardPanel,
					OperationalStatus: pointer(valueobject.OperationalStatusMaintenance)},
				{AssetID: "SS-PNL/1", AssetName: "Substation", AssetType: valueobject.AssetTypeSubstation, ChildCount: 3},
			},
			TotalItems:  12,
			AncestorIDs: map[string][]string{"PNL/1-A": {"SS-001", "SWB-001-1"}, "SS-PNL/1": {}},
		}, nil)

		resp := doGet(t, newTestApp(NewAssetHandler(uc)), "/api/assets?q=%20PNL%2F1%20")

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		body := decodeEnvelope(t, resp.Body)
		assert.Equal(t, "Successfully get list asset", body.Message)
		assert.JSONEq(t, `{
			"items": [
				{"assetId":"PNL/1-A","assetName":"Panel","assetType":"SWITCHBOARD_PANEL","operationalStatus":"MAINTENANCE",
				 "childCount":0,"ancestorAssetIds":["SS-001","SWB-001-1"]},
				{"assetId":"SS-PNL/1","assetName":"Substation","assetType":"SUBSTATION","operationalStatus":null,
				 "childCount":3,"ancestorAssetIds":[]}
			],
			"totalItems": 12
		}`, string(body.Data))
	})

	t.Run("an asset missing from the ancestors map gets [] not null", func(t *testing.T) {
		uc := mocksusecase.NewAssetUsecase(t)
		uc.EXPECT().GetListAsset(mock.Anything, withSearch("SS", 0, 0)).Return(usecase.GetListAssetOutput{
			Assets:     []*entity.AssetNode{{AssetID: "SS-001", AssetName: "Aurora", AssetType: valueobject.AssetTypeSubstation}},
			TotalItems: 1,
		}, nil)

		resp := doGet(t, newTestApp(NewAssetHandler(uc)), "/api/assets?q=SS")

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Contains(t, string(decodeEnvelope(t, resp.Body).Data), `"ancestorAssetIds":[]`)
	})

	t.Run("no matches answers items [] and totalItems 0", func(t *testing.T) {
		uc := mocksusecase.NewAssetUsecase(t)
		uc.EXPECT().GetListAsset(mock.Anything, withSearch("NOPE", 0, 0)).Return(usecase.GetListAssetOutput{}, nil)

		resp := doGet(t, newTestApp(NewAssetHandler(uc)), "/api/assets?q=NOPE")

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.JSONEq(t, `{"items":[],"totalItems":0}`, string(decodeEnvelope(t, resp.Body).Data))
	})

	t.Run("page and limit are forwarded", func(t *testing.T) {
		uc := mocksusecase.NewAssetUsecase(t)
		uc.EXPECT().GetListAsset(mock.Anything, withSearch("-", 2, 25)).Return(usecase.GetListAssetOutput{}, nil)

		resp := doGet(t, newTestApp(NewAssetHandler(uc)), "/api/assets?q=-&page=2&limit=25")

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	for _, target := range []string{"/api/assets?q=", "/api/assets?q=%20%20"} {
		t.Run("a blank text is the hierarchy: "+target, func(t *testing.T) {
			uc := mocksusecase.NewAssetUsecase(t)
			uc.EXPECT().GetListAsset(mock.Anything, withParent(nil)).Return(usecase.GetListAssetOutput{}, nil)

			resp := doGet(t, newTestApp(NewAssetHandler(uc)), target)

			require.Equal(t, fiber.StatusOK, resp.StatusCode)
			assert.NotContains(t, string(decodeEnvelope(t, resp.Body).Data), "ancestorAssetIds")
		})
	}

	t.Run("page and limit without a text reach the usecase, which refuses them", func(t *testing.T) {
		uc := mocksusecase.NewAssetUsecase(t)
		uc.EXPECT().GetListAsset(mock.Anything, mock.MatchedBy(func(in usecase.GetListAssetInput) bool {
			return in.Search == "" && in.Page == 2
		})).Return(usecase.GetListAssetOutput{},
			apperr.BadRequest(errorcode.InvalidRequest, "page and limit apply to a search only"))

		resp := doGet(t, newTestApp(NewAssetHandler(uc)), "/api/assets?page=2")

		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		body := decodeEnvelope(t, resp.Body)
		assert.Equal(t, "Invalid request: page and limit apply to a search only", body.Message)
		assert.JSONEq(t, `{"errorCode":"INVALID_REQUEST"}`, string(body.Data))
	})

	t.Run("a page that is not a number is 400 INVALID_REQUEST", func(t *testing.T) {
		uc := mocksusecase.NewAssetUsecase(t)

		resp := doGet(t, newTestApp(NewAssetHandler(uc)), "/api/assets?q=x&page=two")

		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		assert.JSONEq(t, `{"errorCode":"INVALID_REQUEST"}`, string(decodeEnvelope(t, resp.Body).Data))
	})
}
