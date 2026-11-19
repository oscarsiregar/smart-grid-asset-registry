package asset_repository

import (
	"context"
	"errors"
	"testing"

	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/valueobject"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var nodeColumns = []string{"asset_id", "asset_name", "asset_type", "operational_status", "child_count"}

const nodeSelectSQL = `SELECT a\.asset_id, a\.asset_name, a\.asset_type, a\.operational_status, ` +
	`\(SELECT count\(\*\) FROM asset c WHERE c\.parent_asset_id = a\.asset_id\) AS child_count ` +
	`FROM asset AS a `

func TestAssetRepository_FindNodes(t *testing.T) {
	ctx := context.Background()
	topLevelSQL := nodeSelectSQL + `WHERE a\.parent_asset_id IS NULL$`
	childrenSQL := nodeSelectSQL + `WHERE a\.parent_asset_id = \$1$`
	parentID := "SS-1"
	children := repository.FindAssetNodesFilter{ParentAssetID: &parentID}

	t.Run("top level queries the assets without a parent", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		rows := sqlmock.NewRows(nodeColumns).
			AddRow("SS-2", "Beacon", "SUBSTATION", "IN_SERVICE", 3).
			AddRow("SS-1", "Aurora", "SUBSTATION", nil, 0)
		mock.ExpectQuery(topLevelSQL).WithoutArgs().WillReturnRows(rows)

		result, err := NewAssetRepository(gdb).FindNodes(ctx, nil, repository.FindAssetNodesFilter{})

		require.NoError(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, "SS-2", result[0].AssetID, "returned unsorted, as the database sends them")
		assert.Equal(t, "Beacon", result[0].AssetName)
		assert.Equal(t, valueobject.AssetTypeSubstation, result[0].AssetType)
		require.NotNil(t, result[0].OperationalStatus)
		assert.Equal(t, valueobject.OperationalStatusInService, *result[0].OperationalStatus)
		assert.Equal(t, 3, result[0].ChildCount)
		assert.Nil(t, result[1].OperationalStatus)
		assert.Zero(t, result[1].ChildCount)
	})

	t.Run("children pass the parent ID as a placeholder argument", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		rows := sqlmock.NewRows(nodeColumns).AddRow("TX-1", "T1", "TRANSFORMER", "MAINTENANCE", 0)
		mock.ExpectQuery(childrenSQL).WithArgs("SS-1").WillReturnRows(rows)

		result, err := NewAssetRepository(gdb).FindNodes(ctx, nil, children)

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "TX-1", result[0].AssetID)
	})

	t.Run("returns an empty, non-nil slice when nothing matches", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(childrenSQL).WithArgs("SS-1").WillReturnRows(sqlmock.NewRows(nodeColumns))

		result, err := NewAssetRepository(gdb).FindNodes(ctx, nil, children)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("propagates a database error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("connection refused")
		mock.ExpectQuery(childrenSQL).WillReturnError(dbErr)

		result, err := NewAssetRepository(gdb).FindNodes(ctx, nil, children)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("uses the given transaction", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(topLevelSQL).WillReturnRows(sqlmock.NewRows(nodeColumns))

		_, err := NewAssetRepository(gdb).FindNodes(ctx, wrapTx(gdb), repository.FindAssetNodesFilter{})

		assert.NoError(t, err)
	})

	t.Run("returns an error for an unsupported Tx type", func(t *testing.T) {
		gdb, _ := newMockDB(t)

		result, err := NewAssetRepository(gdb).FindNodes(ctx, unsupportedTx{}, children)

		assert.Nil(t, result)
		assert.Error(t, err)
	})
}

func TestAssetRepository_FindNodes_Search(t *testing.T) {
	ctx := context.Background()
	searchSQL := nodeSelectSQL + `WHERE \(a\.asset_id ILIKE \$1 OR a\.asset_name ILIKE \$2\)$`
	searchWithTypeSQL := nodeSelectSQL +
		`WHERE \(\(a\.asset_id ILIKE \$1 OR a\.asset_name ILIKE \$2\)\) AND a\.asset_type = \$3$`

	t.Run("matches every level by ID or name, without the top-level condition", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		rows := sqlmock.NewRows(nodeColumns).
			AddRow("PNL-1", "Panel 1", "SWITCHBOARD_PANEL", nil, 0).
			AddRow("SS-1", "Aurora", "SUBSTATION", "IN_SERVICE", 2)
		mock.ExpectQuery(searchSQL).WithArgs("%pnl%", "%pnl%").WillReturnRows(rows)

		result, err := NewAssetRepository(gdb).FindNodes(ctx, nil, repository.FindAssetNodesFilter{Search: pointer("pnl")})

		require.NoError(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, "PNL-1", result[0].AssetID, "returned unsorted")
		assert.Equal(t, 2, result[1].ChildCount, "child counts are still selected")
	})

	t.Run("wildcards and the escape character in the text match literally", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(searchSQL).WithArgs(`%a\%b\_c\\d%`, `%a\%b\_c\\d%`).
			WillReturnRows(sqlmock.NewRows(nodeColumns))

		result, err := NewAssetRepository(gdb).FindNodes(ctx, nil, repository.FindAssetNodesFilter{Search: pointer(`a%b_c\d`)})

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("a type limits the matches", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		panel := valueobject.AssetTypeSwitchboardPanel
		mock.ExpectQuery(searchWithTypeSQL).WithArgs("%001%", "%001%", "SWITCHBOARD_PANEL").
			WillReturnRows(sqlmock.NewRows(nodeColumns))

		_, err := NewAssetRepository(gdb).FindNodes(ctx, nil,
			repository.FindAssetNodesFilter{Search: pointer("001"), AssetType: &panel})

		assert.NoError(t, err)
	})

	t.Run("propagates a database error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("connection refused")
		mock.ExpectQuery(searchSQL).WillReturnError(dbErr)

		result, err := NewAssetRepository(gdb).FindNodes(ctx, nil, repository.FindAssetNodesFilter{Search: pointer("x")})

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})
}
