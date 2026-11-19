package asset

import (
	"encoding/json"
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
			"children": [],
			"descendantCounts": null
		}`, string(body.Data))
	})

	t.Run("a substation carries the counts of every type beneath it", func(t *testing.T) {
		uc := mocksusecase.NewAssetUsecase(t)
		uc.EXPECT().GetAssetDetail(mock.Anything, withAssetID("SS-001")).Return(usecase.GetAssetDetailOutput{
			Asset: &entity.Asset{AssetID: "SS-001", AssetType: valueobject.AssetTypeSubstation, AssetName: "Aurora"},
			Path:  []*entity.AssetNode{{AssetID: "SS-001", AssetName: "Aurora", AssetType: valueobject.AssetTypeSubstation}},
			DescendantCounts: &usecase.DescendantCounts{
				Transformers: 2, LVBoards: 1, Switchboards: 2, SwitchboardPanels: 10,
			},
		}, nil)

		resp := doGet(t, newTestApp(NewAssetHandler(uc)), "/api/assets/SS-001")

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		var data struct {
			DescendantCounts json.RawMessage `json:"descendantCounts"`
		}
		require.NoError(t, json.Unmarshal(decodeEnvelope(t, resp.Body).Data, &data))
		assert.JSONEq(t, `{"transformers":2,"lvBoards":1,"switchboards":2,"switchboardPanels":10}`,
			string(data.DescendantCounts))
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

func TestAssetHandler_GetAssetDetail_DecodesTheAssetID(t *testing.T) {
	cases := []struct {
		name   string
		target string
		want   string
	}{
		{"an encoded slash", "/api/assets/SS%2F1", "SS/1"},
		{"an encoded space", "/api/assets/SS%209%20A", "SS 9 A"},
		{"an encoded percent sign", "/api/assets/50%25", "50%"},
		{"an encoded question mark and hash", "/api/assets/x%3Fy%23z", "x?y#z"},
		{"a plain ID is unchanged", "/api/assets/SS-001", "SS-001"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uc := mocksusecase.NewAssetUsecase(t)
			uc.EXPECT().GetAssetDetail(mock.Anything, withAssetID(tc.want)).Return(usecase.GetAssetDetailOutput{
				Asset: &entity.Asset{AssetID: tc.want, AssetType: valueobject.AssetTypeSubstation},
			}, nil)

			resp := doGetRaw(t, newTestApp(NewAssetHandler(uc)), tc.target)

			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		})
	}

	t.Run("a malformed escape is 400 INVALID_REQUEST and the usecase is not called", func(t *testing.T) {
		uc := mocksusecase.NewAssetUsecase(t)

		resp := doGetRaw(t, newTestApp(NewAssetHandler(uc)), "/api/assets/%zz")

		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		body := decodeEnvelope(t, resp.Body)
		assert.Equal(t, "Invalid request: assetId must be valid percent-encoding", body.Message)
		assert.JSONEq(t, `{"errorCode":"INVALID_REQUEST"}`, string(body.Data))
	})
}
