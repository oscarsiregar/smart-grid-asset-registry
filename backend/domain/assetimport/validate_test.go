package assetimport

import (
	"math/rand/v2"
	"os"
	"testing"

	"smart-grid-asset-registry/domain/valueobject"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sampleFile       = "testdata/grid_assets.csv"
	permutationCount = 20
	permutationSeed  = 20261119
)

func parseSample(t *testing.T) ParsedFile {
	t.Helper()

	f, err := os.Open(sampleFile)
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	file, fileErr := Parse(f, testMaxBytes)
	require.Nil(t, fileErr)
	return file
}

func TestValidate_CountsAndRowOrder(t *testing.T) {
	report := validateLines(t, emptyStore(),
		line("TX-NP", "", "TRANSFORMER"),
		line("SS-1", "", "SUBSTATION"),
		"GEN-1,SS-1,GENERATOR,Gen,,,,,,,",
		line("TX-1", "SS-1", "TRANSFORMER"),
	)

	assert.Equal(t, Counts{Total: 4, ToImport: 2, Unchanged: 0, Rejected: 2, WarningRows: 0}, report.Counts)
	assert.Equal(t, report.Counts.Total, report.Counts.ToImport+report.Counts.Unchanged+report.Counts.Rejected)
	require.Len(t, report.Rejected, 2)
	assert.Equal(t, 2, report.Rejected[0].Row, "rejected rows are sorted by row")
	assert.Equal(t, 4, report.Rejected[1].Row)
	assert.Empty(t, report.UnchangedRows)
}

func TestValidate_RejectedRowWithoutAssetID(t *testing.T) {
	report := validateLines(t, emptyStore(), ",SS-1,TRANSFORMER,MIA ID,IN_SERVICE,22,1000,,,,")

	require.Len(t, report.Rejected, 1)
	assert.Nil(t, report.Rejected[0].AssetID)
	assert.Equal(t, 2, report.Rejected[0].Row)
}

func TestValidate_AcceptedAreOrderedParentsFirst(t *testing.T) {
	store := storeWith(storedAsset("SS-STORED", "", valueobject.AssetTypeSubstation))

	report := validateLines(t, store,
		line("PNL-1", "SWB-1", "SWITCHBOARD_PANEL"),
		line("TX-9", "SS-STORED", "TRANSFORMER"),
		line("SWB-1", "SS-1", "SWITCHBOARD"),
		line("SS-2", "", "SUBSTATION"),
		line("SS-1", "", "SUBSTATION"),
	)

	// Depth first (a stored parent counts as depth 0), then asset ID.
	assert.Equal(t, []string{"SS-1", "SS-2", "TX-9", "SWB-1", "PNL-1"}, acceptedIDs(report))
}

func TestValidate_AcceptedAssetsHaveNoIdentityYet(t *testing.T) {
	report := validateLines(t, emptyStore(), line("SS-1", "", "SUBSTATION"))

	require.Len(t, report.Accepted, 1)
	asset := report.Accepted[0]
	assert.Equal(t, uuid.Nil, asset.ID)
	assert.Equal(t, uuid.Nil, asset.ImportID)
	assert.True(t, asset.CreatedAt.IsZero())
	assert.Empty(t, asset.CreatedByName)
	assert.True(t, asset.UpdatedAt.IsZero())
	assert.Empty(t, asset.UpdatedByName)
}

func TestValidate_EmptyStoreMaps(t *testing.T) {
	report := Validate(mustParse(t, line("TX-1", "SS-1", "TRANSFORMER")), StoredAssets{})

	assert.Len(t, report.Rejected, 1, "nil store maps behave like an empty store")
}

// TestValidate_OrderIndependent shuffles the sample's data rows (each row
// keeps its line number) and expects an identical report (spec FR-017, SC-002).
func TestValidate_OrderIndependent(t *testing.T) {
	file := parseSample(t)
	want := Validate(file, emptyStore())

	random := rand.New(rand.NewPCG(permutationSeed, permutationSeed))
	for permutation := range permutationCount {
		shuffled := ParsedFile{HeaderFieldCount: file.HeaderFieldCount, Rows: append([]Row(nil), file.Rows...)}
		random.Shuffle(len(shuffled.Rows), func(i, j int) {
			shuffled.Rows[i], shuffled.Rows[j] = shuffled.Rows[j], shuffled.Rows[i]
		})

		got := Validate(shuffled, emptyStore())

		require.Equalf(t, want, got, "permutation %d changed the result", permutation)
	}
}
