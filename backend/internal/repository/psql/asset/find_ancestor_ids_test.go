package asset_repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetRepository_FindAncestorIDs(t *testing.T) {
	ctx := context.Background()
	ancestorColumns := []string{"asset_id", "ancestor_id", "depth"}
	// (?s): the CTE spans several lines.
	ancestorsSQL := `(?s)^WITH RECURSIVE ancestors AS \(.*WHERE a\.asset_id = ANY\(\$1\).*UNION ALL.*` +
		`JOIN asset p ON p\.asset_id = ancestors\.ancestor_id.*\).*ORDER BY asset_id, depth DESC$`
	ids := []string{"PNL-1", "SS-1"}

	t.Run("maps each asset to its ancestors, substation first", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		rows := sqlmock.NewRows(ancestorColumns).
			AddRow("PNL-1", "SS-1", 2).
			AddRow("PNL-1", "SWB-1", 1)
		mock.ExpectQuery(ancestorsSQL).WithArgs(`{"PNL-1","SS-1"}`).WillReturnRows(rows)

		result, err := NewAssetRepository(gdb).FindAncestorIDs(ctx, nil, ids)

		require.NoError(t, err)
		assert.Equal(t, map[string][]string{
			"PNL-1": {"SS-1", "SWB-1"},
			"SS-1":  {},
		}, result)
	})

	t.Run("runs no query for an empty list", func(t *testing.T) {
		gdb, _ := newMockDB(t)

		result, err := NewAssetRepository(gdb).FindAncestorIDs(ctx, nil, nil)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("propagates a database error", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		dbErr := errors.New("connection refused")
		mock.ExpectQuery(ancestorsSQL).WillReturnError(dbErr)

		result, err := NewAssetRepository(gdb).FindAncestorIDs(ctx, nil, ids)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("uses the given transaction", func(t *testing.T) {
		gdb, mock := newMockDB(t)
		mock.ExpectQuery(ancestorsSQL).WillReturnRows(sqlmock.NewRows(ancestorColumns))

		_, err := NewAssetRepository(gdb).FindAncestorIDs(ctx, wrapTx(gdb), ids)

		assert.NoError(t, err)
	})

	t.Run("returns an error for an unsupported Tx type", func(t *testing.T) {
		gdb, _ := newMockDB(t)

		result, err := NewAssetRepository(gdb).FindAncestorIDs(ctx, unsupportedTx{}, ids)

		assert.Nil(t, result)
		assert.Error(t, err)
	})
}
