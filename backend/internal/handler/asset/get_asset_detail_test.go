package asset

import (
	"testing"
	"time"

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

func withAssetID(id string) any {
	return mock.MatchedBy(func(in usecase.GetAssetDetailInput) bool { return in.AssetID == id })
}

func TestAssetHandler_GetAssetDetail(t *testing.T) {
	t.Run("answers the asset, its path and its children", func(t *testing.T) {
		uc := mocksusecase.NewAssetUsecase(t)
		uc.EXPECT().GetAssetDetail(mock.Anything, withAssetID("PNL-001-1-01")).Return(usecase.GetAssetDetailOutput{
			Asset: &entity.Asset{
				AssetID:          "PNL-001-1-01",
				ParentAssetID:    pointer("SWB-001-1"),
				AssetType:        valueobject.AssetTypeSwitchboardPanel,
				AssetName:        "Aurora Panel 1-1",
				VoltageKV:        pointer(22.5),
				Manufacturer:     pointer("Helios Gridworks"),
				SerialNumber:     pointer("PNL0011A"),
				CommissionedDate: pointer(time.Date(2010, 9, 11, 0, 0, 0, 0, time.UTC)),
			},
			Path: []*entity.AssetNode{
				{AssetID: "SS-001", AssetName: "Aurora", AssetType: valueobject.AssetTypeSubstation,
					OperationalStatus: pointer(valueobject.OperationalStatusInService), ChildCount: 5},
				{AssetID: "SWB-001-1", AssetName: "Switchboard 1", AssetType: valueobject.AssetTypeSwitchboard,
					OperationalStatus: pointer(valueobject.OperationalStatusOutOfService), ChildCount: 4},
				{AssetID: "PNL-001-1-01", AssetName: "Aurora Panel 1-1", AssetType: valueobject.AssetTypeSwitchboardPanel},
			},
		}, nil)

		resp := doGet(t, newTestApp(NewAssetHandler(uc)), "/api/assets/PNL-001-1-01")

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		body := decodeEnvelope(t, resp.Body)
		assert.Equal(t, "Successfully get asset detail", body.Message)
		assert.JSONEq(t, `{
			"asset": {
				"assetId": "PNL-001-1-01",
				"parentAssetId": "SWB-001-1",
				"assetType": "SWITCHBOARD_PANEL",
				"assetName": "Aurora Panel 1-1",
				"operationalStatus": null,
				"voltageKv": 22.5,
				"ratingKva": null,
				"manufacturer": "Helios Gridworks",
				"model": null,
				"serialNumber": "PNL0011A",
				"commissionedDate": "2010-09-11"
			},
			"path": [
				{"assetId":"SS-001","assetName":"Aurora","assetType":"SUBSTATION","operationalStatus":"IN_SERVICE","childCount":5},
				{"assetId":"SWB-001-1","assetName":"Switchboard 1","assetType":"SWITCHBOARD","operationalStatus":"OUT_OF_SERVICE","childCount":4},
				{"assetId":"PNL-001-1-01","assetName":"Aurora Panel 1-1","assetType":"SWITCHBOARD_PANEL","operationalStatus":null,"childCount":0}
			],
			"children": []
		}`, string(body.Data))
	})

	t.Run("a substation has a null parent and a null date when none is stored", func(t *testing.T) {
		uc := mocksusecase.NewAssetUsecase(t)
		uc.EXPECT().GetAssetDetail(mock.Anything, withAssetID("SS-001")).Return(usecase.GetAssetDetailOutput{
			Asset: &entity.Asset{AssetID: "SS-001", AssetType: valueobject.AssetTypeSubstation, AssetName: "Aurora",
				RatingKVA: pointer(1000.0)},
			Path: []*entity.AssetNode{{AssetID: "SS-001", AssetName: "Aurora", AssetType: valueobject.AssetTypeSubstation}},
		}, nil)

		resp := doGet(t, newTestApp(NewAssetHandler(uc)), "/api/assets/SS-001")

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		data := string(decodeEnvelope(t, resp.Body).Data)
		assert.Contains(t, data, `"parentAssetId":null`)
		assert.Contains(t, data, `"commissionedDate":null`)
		assert.Contains(t, data, `"ratingKva":1000`)
		assert.Contains(t, data, `"children":[]`)
	})

	t.Run("an unknown asset is 404 ASSET_NOT_FOUND", func(t *testing.T) {
		uc := mocksusecase.NewAssetUsecase(t)
		uc.EXPECT().GetAssetDetail(mock.Anything, withAssetID("NOPE")).
			Return(usecase.GetAssetDetailOutput{}, apperr.NotFound(errorcode.AssetNotFound, "NOPE"))

		resp := doGet(t, newTestApp(NewAssetHandler(uc)), "/api/assets/NOPE")

		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		body := decodeEnvelope(t, resp.Body)
		assert.Equal(t, "No asset with ID NOPE exists.", body.Message)
		assert.JSONEq(t, `{"errorCode":"ASSET_NOT_FOUND"}`, string(body.Data))
	})
}
