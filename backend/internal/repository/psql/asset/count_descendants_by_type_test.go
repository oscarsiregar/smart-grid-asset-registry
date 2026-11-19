package asset_repository

import (
	"context"
	"errors"
	"testing"

	"smart-grid-asset-registry/domain/valueobject"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetRepository_CountDescendantsByType(t *testing.T) {
	ctx := context.Background()
	countColumns := []string{"asset_type", "asset_count"}
	// (?s): the CTE spans several lines.
	countSQL := `(?s)^WITH RECURSIVE descendants AS \(.*WHERE c\.parent_asset_id = \$1.*UNION ALL.*` +
		`JOIN descendants d ON c\.parent_asset_id = d\.asset_id.*\).*` +
		`SELECT asset_type, count\(\*\) AS asset_count FROM descendants GROUP BY asset_type$`

	t.Run("counts every level below the asset by type", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		rows := sqlmock.NewRows(countColumns).
			AddRow("TRANSFORMER", 2).
			AddRow("LV_BOARD", 1).
			AddRow("SWITCHBOARD", 2).
			AddRow("SWITCHBOARD_PANEL", 10)
		mock.ExpectQuery(countSQL).WithArgs("SS-001").WillReturnRows(rows)

		result, err := NewAssetRepository(gdb).CountDescendantsByType(ctx, nil, "SS-001")

		require.NoError(t, err)
		assert.Equal(t, map[valueobject.AssetType]int{
			valueobject.AssetTypeTransformer:      2,
			valueobject.AssetTypeLVBoard:          1,
			valueobject.AssetTypeSwitchboard:      2,
			valueobject.AssetTypeSwitchboardPanel: 10,
		}, result)
	})

	t.Run("no descendants gives an empty, non-nil map", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(countSQL).WithArgs("PNL-1").WillReturnRows(sqlmock.NewRows(countColumns))

		result, err := NewAssetRepository(gdb).CountDescendantsByType(ctx, nil, "PNL-1")

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("propagates a database error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("connection refused")
		mock.ExpectQuery(countSQL).WillReturnError(dbErr)

		result, err := NewAssetRepository(gdb).CountDescendantsByType(ctx, nil, "SS-001")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("uses the given transaction", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(countSQL).WillReturnRows(sqlmock.NewRows(countColumns))

		_, err := NewAssetRepository(gdb).CountDescendantsByType(ctx, wrapTx(gdb), "SS-001")

		assert.NoError(t, err)
	})

	t.Run("returns an error for an unsupported Tx type", func(t *testing.T) {
		gdb, _ := newMockDB(t)

		result, err := NewAssetRepository(gdb).CountDescendantsByType(ctx, unsupportedTx{}, "SS-001")

		assert.Nil(t, result)
		assert.Error(t, err)
	})
}
