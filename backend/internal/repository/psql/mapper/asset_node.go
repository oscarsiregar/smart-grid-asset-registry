package mapper

import (
	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/internal/repository/psql/model"
)

func ToEntityAssetNode(m *model.AssetNode) *entity.AssetNode {
	if m == nil {
		return nil
	}
	return &entity.AssetNode{
		AssetID:           m.AssetID,
		AssetName:         m.AssetName,
		AssetType:         m.AssetType,
		OperationalStatus: m.OperationalStatus,
		ChildCount:        m.ChildCount,
	}
}

// ToEntityAssetNodes always returns a non-nil slice.
func ToEntityAssetNodes(models []*model.AssetNode) []*entity.AssetNode {
	nodes := make([]*entity.AssetNode, 0, len(models))
	for _, m := range models {
		nodes = append(nodes, ToEntityAssetNode(m))
	}
	return nodes
}
