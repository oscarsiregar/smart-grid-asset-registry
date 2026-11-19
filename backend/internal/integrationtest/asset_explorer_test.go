package integrationtest

import (
	"context"
	"testing"
	"time"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	asset_repository "smart-grid-asset-registry/internal/repository/psql/asset"
	"smart-grid-asset-registry/internal/testsupport/pgtest"
	asset_usecase "smart-grid-asset-registry/internal/usecase/asset"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// sampleAssetCount is the number of rows of the sample the import accepts.
const sampleAssetCount = 205

func newExplorerUsecase(db *gorm.DB) usecase.AssetUsecase {
	return asset_usecase.NewAssetUsecase(asset_repository.NewAssetRepository(db), quietLogger())
}

func listIDs(t *testing.T, uc usecase.AssetUsecase, parentAssetID *string) []string {
	t.Helper()
	output, err := uc.GetListAsset(context.Background(), usecase.GetListAssetInput{ParentAssetID: parentAssetID})
	require.NoError(t, err)
	ids := make([]string, 0, len(output.Assets))
	for _, node := range output.Assets {
		ids = append(ids, node.AssetID)
	}
	return ids
}

// walk lists every asset reachable from the top level, expanding each node
// the way the tree does, and returns how often each asset ID was reached.
func walk(t *testing.T, uc usecase.AssetUsecase) map[string]int {
	t.Helper()
	reached := map[string]int{}
	var visit func(parentAssetID *string)
	visit = func(parentAssetID *string) {
		output, err := uc.GetListAsset(context.Background(), usecase.GetListAssetInput{ParentAssetID: parentAssetID})
		require.NoError(t, err)
		for _, node := range output.Assets {
			reached[node.AssetID]++
			if node.ChildCount > 0 {
				visit(pointer(node.AssetID))
			}
		}
	}
	visit(nil)
	return reached
}

// TestAssetExplorer_Sample browses the imported sample (US1, SC-001).
func TestAssetExplorer_Sample(t *testing.T) {
	db := pgtest.Open(t)
	accepted := importSample(t, newImportUsecase(db, asset_repository.NewAssetRepository(db))).Report.Accepted
	require.Len(t, accepted, sampleAssetCount)
	uc := newExplorerUsecase(db)
	ctx := context.Background()

	t.Run("top level: the 13 substations, SS-T01 last", func(t *testing.T) {
		assert.Equal(t, []string{
			"SS-001", "SS-002", "SS-003", "SS-004", "SS-005", "SS-006", "SS-007",
			"SS-008", "SS-009", "SS-010", "SS-011", "SS-012", "SS-T01",
		}, listIDs(t, uc, nil))
	})

	t.Run("children of SS-001: transformers, then the LV board, then switchboards", func(t *testing.T) {
		assert.Equal(t, []string{"TX-001-1", "TX-001-2", "LVB-001-1", "SWB-001-1", "SWB-001-2"},
			listIDs(t, uc, pointer("SS-001")))
	})

	t.Run("walking the tree reaches every asset exactly once", func(t *testing.T) {
		reached := walk(t, uc)
		assert.Len(t, reached, sampleAssetCount)
		for assetID, times := range reached {
			assert.Equalf(t, 1, times, "%s reached %d times", assetID, times)
		}
	})

	t.Run("every asset's details equal its CSV row", func(t *testing.T) {
		for _, want := range accepted {
			output, err := uc.GetAssetDetail(ctx, usecase.GetAssetDetailInput{AssetID: want.AssetID})
			require.NoError(t, err, want.AssetID)
			assert.Emptyf(t, want.DifferingFields(output.Asset), "%s differs", want.AssetID)

			require.NotEmpty(t, output.Path, want.AssetID)
			assert.Equal(t, valueobject.AssetTypeSubstation, output.Path[0].AssetType, want.AssetID)
			assert.Equal(t, want.AssetID, output.Path[len(output.Path)-1].AssetID)
		}
	})

	t.Run("PNL-001-1-01: blank rating, date and path", func(t *testing.T) {
		output, err := uc.GetAssetDetail(ctx, usecase.GetAssetDetailInput{AssetID: "PNL-001-1-01"})
		require.NoError(t, err)

		asset := output.Asset
		assert.Nil(t, asset.RatingKVA)
		require.NotNil(t, asset.CommissionedDate)
		assert.Equal(t, "2010-09-11", asset.CommissionedDate.Format(time.DateOnly))
		require.NotNil(t, asset.OperationalStatus)
		assert.Equal(t, valueobject.OperationalStatusMaintenance, *asset.OperationalStatus)
		assert.Equal(t, []string{"SS-001", "SWB-001-1", "PNL-001-1-01"}, nodeIDs(output.Path))
		assert.Empty(t, output.Children)
	})

	t.Run("SWB-001-1 lists its panels in natural ID order", func(t *testing.T) {
		output, err := uc.GetAssetDetail(ctx, usecase.GetAssetDetailInput{AssetID: "SWB-001-1"})
		require.NoError(t, err)

		ids := nodeIDs(output.Children)
		assert.Equal(t, listIDs(t, uc, pointer("SWB-001-1")), ids)
		require.NotEmpty(t, ids)
		assert.Equal(t, "PNL-001-1-01", ids[0])
	})
}

// TestAssetExplorer_EmptyRegistry: the top level is an empty list, not an error.
func TestAssetExplorer_EmptyRegistry(t *testing.T) {
	db := pgtest.Open(t)

	assert.Empty(t, listIDs(t, newExplorerUsecase(db), nil))
}

func nodeIDs(nodes []*entity.AssetNode) []string {
	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ids = append(ids, node.AssetID)
	}
	return ids
}
