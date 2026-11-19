package assetimport

import (
	"strconv"
	"strings"
	"testing"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/valueobject"

	"github.com/stretchr/testify/require"
)

// testHeader is the eleven required columns in CSV order.
const testHeader = "asset_id,parent_asset_id,asset_type,asset_name,operational_status," +
	"voltage_kv,rating_kva,manufacturer,model,serial_number,commissioned_date"

const testMaxBytes = 1 << 20

// csvText joins the header and the given data lines into a file body.
func csvText(lines ...string) string {
	return testHeader + "\n" + strings.Join(lines, "\n") + "\n"
}

func mustParse(t *testing.T, lines ...string) ParsedFile {
	t.Helper()

	file, fileErr := Parse(strings.NewReader(csvText(lines...)), testMaxBytes)
	require.Nil(t, fileErr, "unexpected file error")
	return file
}

func validateLines(t *testing.T, stored StoredAssets, lines ...string) Report {
	t.Helper()
	return Validate(mustParse(t, lines...), stored)
}

func emptyStore() StoredAssets {
	return StoredAssets{ByAssetID: map[string]*entity.Asset{}, BySerial: map[string][]string{}}
}

// storeWith returns a store holding the given assets, indexed by ID and serial.
func storeWith(assets ...*entity.Asset) StoredAssets {
	store := emptyStore()
	for _, asset := range assets {
		store.ByAssetID[asset.AssetID] = asset
		if asset.SerialNumber != nil {
			store.BySerial[*asset.SerialNumber] = append(store.BySerial[*asset.SerialNumber], asset.AssetID)
		}
	}
	return store
}

func storedAsset(assetID, parentAssetID string, assetType valueobject.AssetType) *entity.Asset {
	asset := &entity.Asset{AssetID: assetID, AssetType: assetType, AssetName: assetID}
	if parentAssetID != "" {
		asset.ParentAssetID = &parentAssetID
	}
	return asset
}

// codesOf returns the error codes per rejected row, keyed by asset ID
// ("#<row>" when the row has no asset ID).
func codesOf(report Report) map[string][]string {
	codes := map[string][]string{}
	for _, rejected := range report.Rejected {
		key := "#" + strconv.Itoa(rejected.Row)
		if rejected.AssetID != nil {
			key = *rejected.AssetID
		}
		for _, issue := range rejected.Errors {
			codes[key] = append(codes[key], issue.Code)
		}
	}
	return codes
}

// issuesOf returns the full issues of the rejected row with the given asset ID.
func issuesOf(t *testing.T, report Report, assetID string) []Issue {
	t.Helper()
	for _, rejected := range report.Rejected {
		if rejected.AssetID != nil && *rejected.AssetID == assetID {
			return rejected.Errors
		}
	}
	t.Fatalf("asset %s is not rejected", assetID)
	return nil
}

func acceptedIDs(report Report) []string {
	ids := make([]string, 0, len(report.Accepted))
	for _, asset := range report.Accepted {
		ids = append(ids, asset.AssetID)
	}
	return ids
}
