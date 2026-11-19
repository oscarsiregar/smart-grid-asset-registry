package integrationtest

import (
	"context"
	"slices"
	"strings"
	"testing"

	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	asset_repository "smart-grid-asset-registry/internal/repository/psql/asset"
	"smart-grid-asset-registry/internal/testsupport/pgtest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type storedName struct {
	AssetID   string
	AssetName string
}

// expectedMatches is the oracle for SC-002: every stored asset whose ID or
// name contains the text, letter case ignored, computed in Go from the stored
// rows rather than by the query under test.
func expectedMatches(t *testing.T, db *gorm.DB, text string) []string {
	t.Helper()
	var rows []storedName
	require.NoError(t, db.Table("asset").Select("asset_id, asset_name").Scan(&rows).Error)

	needle := strings.ToLower(text)
	matches := []string{}
	for _, row := range rows {
		if strings.Contains(strings.ToLower(row.AssetID), needle) || strings.Contains(strings.ToLower(row.AssetName), needle) {
			matches = append(matches, row.AssetID)
		}
	}
	slices.Sort(matches)
	return matches
}

func search(t *testing.T, uc usecase.AssetUsecase, text string, page int) usecase.GetListAssetOutput {
	t.Helper()
	output, err := uc.GetListAsset(context.Background(), usecase.GetListAssetInput{Search: text, Page: page, Limit: usecase.MaxLimit})
	require.NoError(t, err, text)
	return output
}

// maxSearchPages bounds searchAll, so a search that never runs out of pages
// fails instead of looping (the sample needs at most 3 pages of 100).
const maxSearchPages = 20

// searchAll walks every page of a search and returns the IDs in order.
func searchAll(t *testing.T, uc usecase.AssetUsecase, text string) (ids []string, total int) {
	t.Helper()
	for page := 1; page <= maxSearchPages; page++ {
		output := search(t, uc, text, page)
		total = output.TotalItems
		if len(output.Assets) == 0 {
			return ids, total
		}
		ids = append(ids, nodeIDs(output.Assets)...)
	}
	require.FailNowf(t, "search never ran out of pages", "search for %q", text)
	return nil, 0
}

func sortedCopy(ids []string) []string {
	sorted := slices.Clone(ids)
	slices.Sort(sorted)
	return sorted
}

// TestAssetSearch_Sample searches the imported sample (US1, SC-001, SC-002).
func TestAssetSearch_Sample(t *testing.T) {
	db := pgtest.Open(t)
	accepted := importSample(t, newImportUsecase(db, asset_repository.NewAssetRepository(db))).Report.Accepted
	uc := newExplorerUsecase(db)

	t.Run("a full ID lists that asset first, with its ancestors", func(t *testing.T) {
		output := search(t, uc, "PNL-001-1-01", 1)

		require.NotEmpty(t, output.Assets)
		assert.Equal(t, "PNL-001-1-01", output.Assets[0].AssetID)
		assert.Equal(t, []string{"SS-001", "SWB-001-1"}, output.AncestorIDs["PNL-001-1-01"])
	})

	t.Run("every stored asset is first when searched by its full ID (SC-001)", func(t *testing.T) {
		for _, asset := range accepted {
			output := search(t, uc, asset.AssetID, 1)
			require.NotEmpty(t, output.Assets, asset.AssetID)
			assert.Equalf(t, asset.AssetID, output.Assets[0].AssetID, "search for %s", asset.AssetID)
		}
	})

	t.Run("an asset matching by ID and by name is listed once (FR-009)", func(t *testing.T) {
		ids, total := searchAll(t, uc, "Aurora")

		assert.Len(t, ids, total)
		assert.Equal(t, len(ids), len(slices.Compact(sortedCopy(ids))), "no asset twice")
	})

	for _, text := range []string{"aurora panel", "swb-001", "001", "-", "T1", "PNL-001-1-01"} {
		t.Run("results equal the oracle for "+text, func(t *testing.T) {
			ids, total := searchAll(t, uc, text)
			want := expectedMatches(t, db, text)

			assert.Equal(t, want, sortedCopy(ids), "none missing, none extra (SC-002)")
			assert.Equal(t, len(want), total)
		})
	}

	t.Run("letter case is ignored", func(t *testing.T) {
		lower, _ := searchAll(t, uc, "swb-001")
		upper, _ := searchAll(t, uc, "SWB-001")

		assert.Equal(t, upper, lower)
		assert.NotEmpty(t, lower)
	})

	t.Run("the pages of a broad search do not overlap", func(t *testing.T) {
		first, err := uc.GetListAsset(context.Background(), usecase.GetListAssetInput{Search: "-", Page: 1})
		require.NoError(t, err)
		second, err := uc.GetListAsset(context.Background(), usecase.GetListAssetInput{Search: "-", Page: 2})
		require.NoError(t, err)

		require.Len(t, first.Assets, usecase.SearchDefaultLimit)
		assert.Empty(t, slices.DeleteFunc(nodeIDs(second.Assets), func(id string) bool {
			return !slices.Contains(nodeIDs(first.Assets), id)
		}), "no asset on both pages")
		assert.Equal(t, first.TotalItems, second.TotalItems)
	})

	t.Run("the hierarchy reads of feature 002 are unchanged", func(t *testing.T) {
		assert.Len(t, listIDs(t, uc, nil), sampleSubstationCount)
		output, err := uc.GetListAsset(context.Background(), usecase.GetListAssetInput{})
		require.NoError(t, err)
		assert.Nil(t, output.AncestorIDs)
	})
}

// TestAssetSearch_WildcardsAreLiteral: % and _ in the text match only those
// characters (FR-003).
func TestAssetSearch_WildcardsAreLiteral(t *testing.T) {
	db := pgtest.Open(t)
	commit(t, newImportUsecase(db, asset_repository.NewAssetRepository(db)), valueobject.ImportModeStrict,
		"A%B,,SUBSTATION,Percent Substation,IN_SERVICE,,,,,,",
		"A_B,,SUBSTATION,Underscore Substation,IN_SERVICE,,,,,,",
		"AXB,,SUBSTATION,Plain Substation,IN_SERVICE,,,,,,",
	)
	uc := newExplorerUsecase(db)

	percent, _ := searchAll(t, uc, "%")
	underscore, _ := searchAll(t, uc, "_")

	assert.Equal(t, []string{"A%B"}, percent)
	assert.Equal(t, []string{"A_B"}, underscore)
}
