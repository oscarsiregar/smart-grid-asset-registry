package asset_repository

import (
	"context"

	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/repository/psql/model"
)

// findAncestorIDsSQL walks up parent_asset_id from every given asset at once
// (research R4): one row per ancestor, depth 1 being the parent. The IDs are
// one text-array parameter. Imports guarantee there are no cycles; the depth
// limit only stops a corrupted store from looping forever.
const findAncestorIDsSQL = `WITH RECURSIVE ancestors AS (
	SELECT a.asset_id, a.parent_asset_id AS ancestor_id, 1 AS depth
	FROM asset a
	WHERE a.asset_id = ANY(?) AND a.parent_asset_id IS NOT NULL
	UNION ALL
	SELECT ancestors.asset_id, p.parent_asset_id, ancestors.depth + 1
	FROM ancestors
	JOIN asset p ON p.asset_id = ancestors.ancestor_id
	WHERE p.parent_asset_id IS NOT NULL AND ancestors.depth < 16
)
SELECT asset_id, ancestor_id, depth FROM ancestors ORDER BY asset_id, depth DESC`

func (r assetRepository) FindAncestorIDs(ctx context.Context, trx repository.Tx, assetIDs []string) (map[string][]string, error) {
	if len(assetIDs) == 0 {
		return map[string][]string{}, nil
	}

	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return nil, err
	}

	var rows []model.AssetAncestor
	if err := db.Raw(findAncestorIDsSQL, textArray(assetIDs)).Scan(&rows).Error; err != nil {
		return nil, err
	}

	ancestors := make(map[string][]string, len(assetIDs))
	for _, assetID := range assetIDs {
		ancestors[assetID] = []string{}
	}
	// Rows come deepest first per asset, so appending gives substation first.
	for _, row := range rows {
		ancestors[row.AssetID] = append(ancestors[row.AssetID], row.AncestorID)
	}
	return ancestors, nil
}
