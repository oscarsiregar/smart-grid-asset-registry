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
	// SC-006: the top level within 2 s; a node's children or an asset's
	// details within 1 s. Measured on the API side (usecase + PostgreSQL).
	topLevelBudget = 2 * time.Second
	nodeBudget     = time.Second
	// One switchboard far wider than any in the sample (complete lists,
	// Template Exception 13).
	widePanelCount = 5000
)

// wideSwitchboardFile is a substation with one switchboard holding
// widePanelCount panels.
func wideSwitchboardFile() []byte {
	var b bytes.Buffer
	b.WriteString(csvHeader + "\n")
	b.WriteString("SS-WIDE,,SUBSTATION,Wide Substation,IN_SERVICE,22,,,,,\n")
	b.WriteString("SWB-WIDE,SS-WIDE,SWITCHBOARD,Wide Switchboard,IN_SERVICE,22,,,,,\n")
	for panel := range widePanelCount {
		fmt.Fprintf(&b, "PNL-WIDE-%d,SWB-WIDE,SWITCHBOARD_PANEL,Wide Panel %d,IN_SERVICE,0.4,,,,,\n", panel, panel)
	}
	return b.Bytes()
}

func importBody(t *testing.T, uc usecase.AssetImportUsecase, fileName string, body []byte) {
	t.Helper()
	output, err := uc.CreateAssetImport(context.Background(), usecase.CreateAssetImportInput{
		FileName: fileName, File: bytes.NewReader(body), Mode: valueobject.ImportModeStrict,
	})
	require.NoError(t, err)
	require.Equal(t, valueobject.ImportOutcomeCommitted, output.Import.Outcome)
}

// timed runs read and returns how long it took.
func timed(t *testing.T, label string, read func() error) time.Duration {
	t.Helper()
	start := time.Now()
	require.NoError(t, read())
	elapsed := time.Since(start)
	t.Logf("%s took %s", label, elapsed)
	return elapsed
}

// TestAssetExplorer_Scale measures the explorer reads with ~50 000 stored
// assets (SC-006, research R11).
func TestAssetExplorer_Scale(t *testing.T) {
	if testing.Short() {
		t.Skip("scale test skipped with -short")
	}
	db := pgtest.Open(t)
	importUsecase := newImportUsecase(db, asset_repository.NewAssetRepository(db))
	body, rows := largeFile(t)
	importBody(t, importUsecase, "large.csv", body)
	importBody(t, importUsecase, "wide.csv", wideSwitchboardFile())
	t.Logf("stored %d assets", countRows(t, db, "asset"))
	require.Greater(t, rows+widePanelCount, 40000, "the scale test needs tens of thousands of assets")

	uc := newExplorerUsecase(db)
	ctx := context.Background()

	var substations []string
	elapsed := timed(t, "top level", func() error {
		output, err := uc.GetListAsset(ctx, usecase.GetListAssetInput{})
		for _, node := range output.Assets {
			substations = append(substations, node.AssetID)
		}
		return err
	})
	assert.Less(t, elapsed, topLevelBudget, "SC-006 top level")
	assert.Contains(t, substations, "SS-WIDE")

	var panels int
	elapsed = timed(t, "children of the widest switchboard", func() error {
		output, err := uc.GetListAsset(ctx, usecase.GetListAssetInput{ParentAssetID: pointer("SWB-WIDE")})
		panels = len(output.Assets)
		return err
	})
	assert.Less(t, elapsed, nodeBudget, "SC-006 expanding a node")
	assert.Equal(t, widePanelCount, panels, "every child at once, no paging")

	var counts *usecase.DescendantCounts
	elapsed = timed(t, "details of the wide substation", func() error {
		output, err := uc.GetAssetDetail(ctx, usecase.GetAssetDetailInput{AssetID: "SS-WIDE"})
		counts = output.DescendantCounts
		return err
	})
	assert.Less(t, elapsed, nodeBudget, "SC-006 selecting an asset")
	require.NotNil(t, counts)
	assert.Equal(t, widePanelCount, counts.SwitchboardPanels)

	elapsed = timed(t, "details of a panel", func() error {
		_, err := uc.GetAssetDetail(ctx, usecase.GetAssetDetailInput{AssetID: "PNL-WIDE-4999"})
		return err
	})
	assert.Less(t, elapsed, nodeBudget, "SC-006 selecting an asset")
}
