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
