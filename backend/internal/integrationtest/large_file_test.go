package integrationtest

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	asset_repository "smart-grid-asset-registry/internal/repository/psql/asset"
	"smart-grid-asset-registry/internal/testsupport/pgtest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Keep the generated file just under the 5 MB limit.
	largeFileTarget = usecase.MaxFileBytes - 64*1024
	// SC-004: the validation result appears within 3 seconds.
	validateBudget = 3 * time.Second
	panelsPerBoard = 8
)

// largeFile builds a valid CSV just under 5 MB: substations, each with a
// transformer, an LV board and a switchboard carrying several panels.
func largeFile(t *testing.T) (body []byte, rows int) {
	t.Helper()

	var b bytes.Buffer
	b.WriteString(csvHeader + "\n")
	write := func(format string, args ...any) {
		fmt.Fprintf(&b, format+"\n", args...)
		rows++
	}

	for site := 0; b.Len() < largeFileTarget; site++ {
		ss := fmt.Sprintf("SS-%05d", site)
		write("%s,,SUBSTATION,Substation %d,IN_SERVICE,22,,Arcwell,SUB-1,SN-%s,15/2/09", ss, site, ss)
		write("TX-%05d,%s,TRANSFORMER,Transformer %d,IN_SERVICE,22,2000,Arcwell,TXR-2000,SN-TX-%05d,12/3/10", site, ss, site, site)
		write("LVB-%05d,%s,LV_BOARD,LV Board %d,MAINTENANCE,0.4,,Helios,LVB-400,SN-LVB-%05d,20/6/11", site, ss, site, site)
		swb := fmt.Sprintf("SWB-%05d", site)
		write("%s,%s,SWITCHBOARD,Switchboard %d,IN_SERVICE,22,,Helios,SWB-22,SN-%s,2011-06-20", swb, ss, site, swb)
		for panel := range panelsPerBoard {
			write("PNL-%05d-%d,%s,SWITCHBOARD_PANEL,Panel %d-%d,IN_SERVICE,22,,Helios,PNL-1,SN-PNL-%05d-%d,2011-06-20",
				site, panel, swb, site, panel, site, panel)
		}
	}
	require.Less(t, b.Len(), usecase.MaxFileBytes, "the generated file must stay under the limit")
	return b.Bytes(), rows
}

// TestLargeFile proves a near-limit file works end to end: the single-array
// "= ANY(?)" lookups and 1 000-row insert batches stay far below
// PostgreSQL's 65 535 bind-parameter limit (research R6, SC-004).
func TestLargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("large-file test skipped with -short")
	}
	db := pgtest.Open(t)
	uc := newImportUsecase(db, asset_repository.NewAssetRepository(db))
	body, rows := largeFile(t)
	t.Logf("generated %d rows, %d bytes", rows, len(body))

	start := time.Now()
	checked, err := uc.ValidateAssetImport(context.Background(), usecase.ValidateAssetImportInput{
		FileName: "large.csv", File: bytes.NewReader(body), Mode: valueobject.ImportModePartial,
	})
	elapsed := time.Since(start)
	t.Logf("validate took %s", elapsed)

	require.NoError(t, err)
	assert.Equal(t, rows, checked.Report.Counts.ToImport)
	assert.Empty(t, checked.Report.Rejected)
	assert.Less(t, elapsed, validateBudget, "SC-004")

	start = time.Now()
	committed, err := uc.CreateAssetImport(context.Background(), usecase.CreateAssetImportInput{
		FileName: "large.csv", File: bytes.NewReader(body), Mode: valueobject.ImportModePartial,
	})
	t.Logf("commit took %s", time.Since(start))

	require.NoError(t, err)
	assert.Equal(t, valueobject.ImportOutcomeCommitted, committed.Import.Outcome)
	assert.EqualValues(t, rows, countRows(t, db, "asset"))
}
