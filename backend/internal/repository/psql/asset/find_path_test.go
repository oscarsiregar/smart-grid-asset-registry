package asset_repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetRepository_FindPath(t *testing.T) {
	ctx := context.Background()
	// (?s): the CTE spans several lines.
	pathSQL := `(?s)^WITH RECURSIVE path AS \(.*WHERE a\.asset_id = \$1.*UNION ALL.*` +
		`JOIN path ON p\.asset_id = path\.parent_asset_id.*\).*child_count.*FROM path ORDER BY path\.depth DESC$`

	t.Run("returns the asset and its ancestors, substation first", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		rows := sqlmock.NewRows(nodeColumns).
			AddRow("SS-1", "Aurora", "SUBSTATION", "IN_SERVICE", 2).
			AddRow("SWB-1", "Switchboard 1", "SWITCHBOARD", "OUT_OF_SERVICE", 4).
			AddRow("PNL-1", "Panel 1", "SWITCHBOARD_PANEL", nil, 0)
		mock.ExpectQuery(pathSQL).WithArgs("PNL-1").WillReturnRows(rows)

		result, err := NewAssetRepository(gdb).FindPath(ctx, nil, "PNL-1")

		require.NoError(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, "SS-1", result[0].AssetID)
		assert.Equal(t, "SWB-1", result[1].AssetID)
		assert.Equal(t, 4, result[1].ChildCount)
		assert.Equal(t, "PNL-1", result[2].AssetID)
		assert.Nil(t, result[2].OperationalStatus)
	})

	t.Run("returns an empty, non-nil slice for an unknown ID", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(pathSQL).WithArgs("NOPE").WillReturnRows(sqlmock.NewRows(nodeColumns))

		result, err := NewAssetRepository(gdb).FindPath(ctx, nil, "NOPE")

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("propagates a database error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("connection refused")
		mock.ExpectQuery(pathSQL).WillReturnError(dbErr)

		result, err := NewAssetRepository(gdb).FindPath(ctx, nil, "PNL-1")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("uses the given transaction", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(pathSQL).WillReturnRows(sqlmock.NewRows(nodeColumns))

		_, err := NewAssetRepository(gdb).FindPath(ctx, wrapTx(gdb), "PNL-1")

		assert.NoError(t, err)
	})

	t.Run("returns an error for an unsupported Tx type", func(t *testing.T) {
		gdb, _ := newMockDB(t)

		result, err := NewAssetRepository(gdb).FindPath(ctx, unsupportedTx{}, "PNL-1")

		assert.Nil(t, result)
		assert.Error(t, err)
	})
}
