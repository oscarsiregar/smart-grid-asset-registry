package asset_repository

import (
	"context"

	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/repository/psql/model"
)

// countDescendantsSQL walks down from the asset through parent_asset_id
// (served by ix_asset_parent_asset_id) and counts every level by type
// (research R4). Imports guarantee there are no cycles; the depth limit only
// stops a corrupted store from looping forever.
const countDescendantsSQL = `WITH RECURSIVE descendants AS (
	SELECT c.asset_id, c.asset_type, 1 AS depth
	FROM asset c
	WHERE c.parent_asset_id = ?
	UNION ALL
	SELECT c.asset_id, c.asset_type, d.depth + 1
	FROM asset c
	JOIN descendants d ON c.parent_asset_id = d.asset_id
	WHERE d.depth < 16
)
SELECT asset_type, count(*) AS asset_count FROM descendants GROUP BY asset_type`

func (r assetRepository) CountDescendantsByType(ctx context.Context, trx repository.Tx, assetID string) (map[valueobject.AssetType]int, error) {
	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return nil, err
	}

	var rows []model.AssetTypeCount
	if err := db.Raw(countDescendantsSQL, assetID).Scan(&rows).Error; err != nil {
		return nil, err
	}

	counts := make(map[valueobject.AssetType]int, len(rows))
	for _, row := range rows {
		counts[row.AssetType] = row.AssetCount
	}
	return counts, nil
}
