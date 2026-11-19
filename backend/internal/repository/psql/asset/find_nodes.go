package asset_repository

import (
	"context"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/internal/repository/psql/gormtx"
	"smart-grid-asset-registry/internal/repository/psql/mapper"
	"smart-grid-asset-registry/internal/repository/psql/model"
	"smart-grid-asset-registry/internal/repository/psql/querysafety"
)

// nodeSelectColumns selects an asset as a node of alias a, with its number of
// immediate children (served by ix_asset_parent_asset_id).
const nodeSelectColumns = "a.asset_id, a.asset_name, a.asset_type, a.operational_status, " +
	"(SELECT count(*) FROM asset c WHERE c.parent_asset_id = a.asset_id) AS child_count"

func (r assetRepository) FindNodes(ctx context.Context, trx repository.Tx, filter repository.FindAssetNodesFilter) ([]*entity.AssetNode, error) {
	db, err := gormtx.Conn(ctx, r.db, trx)
	if err != nil {
		return nil, err
	}

	query := db.Table("asset AS a").Select(nodeSelectColumns)
	switch {
	case filter.Search != nil:
		// Every level; "contains", letter case ignored, characters literal.
		pattern := querysafety.ContainsPattern(*filter.Search)
		query = query.Where("(a.asset_id ILIKE ? OR a.asset_name ILIKE ?)", pattern, pattern)
		if filter.AssetType != nil {
			query = query.Where("a.asset_type = ?", *filter.AssetType)
		}
	case filter.ParentAssetID == nil:
		query = query.Where("a.parent_asset_id IS NULL")
	default:
		query = query.Where("a.parent_asset_id = ?", *filter.ParentAssetID)
	}

	var rows []*model.AssetNode
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapper.ToEntityAssetNodes(rows), nil
}
