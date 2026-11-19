package asset_repository

import (
	"context"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/repository/psql/mapper"
	"smart-grid-asset-registry/internal/repository/psql/model"
)

// findPathSQL walks from the asset up through parent_asset_id (research R4).
// Imports guarantee there are no cycles; the depth limit only stops a
// corrupted store from looping forever.
const findPathSQL = `WITH RECURSIVE path AS (
	SELECT a.asset_id, a.asset_name, a.asset_type, a.operational_status, a.parent_asset_id, 0 AS depth
	FROM asset a
	WHERE a.asset_id = ?
	UNION ALL
	SELECT p.asset_id, p.asset_name, p.asset_type, p.operational_status, p.parent_asset_id, path.depth + 1
	FROM asset p
	JOIN path ON p.asset_id = path.parent_asset_id
	WHERE path.depth < 16
)
SELECT path.asset_id, path.asset_name, path.asset_type, path.operational_status,
	(SELECT count(*) FROM asset c WHERE c.parent_asset_id = path.asset_id) AS child_count
FROM path ORDER BY path.depth DESC`

func (r assetRepository) FindPath(ctx context.Context, trx repository.Tx, assetID string) ([]*entity.AssetNode, error) {
	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return nil, err
	}

	var rows []*model.AssetNode
	if err := db.Raw(findPathSQL, assetID).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return mapper.ToEntityAssetNodes(rows), nil
}
