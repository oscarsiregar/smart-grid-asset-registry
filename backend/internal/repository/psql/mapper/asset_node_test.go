package mapper

import (
	"testing"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/repository/psql/model"

	"github.com/stretchr/testify/assert"
)

func TestToEntityAssetNodes_CopiesEveryField(t *testing.T) {
	status := valueobject.OperationalStatusMaintenance

	nodes := ToEntityAssetNodes([]*model.AssetNode{
		{AssetID: "SWB-001-1", AssetName: "Aurora Switchboard 1", AssetType: valueobject.AssetTypeSwitchboard,
			OperationalStatus: &status, ChildCount: 4},
		{AssetID: "PNL-001-1-1", AssetName: "Aurora Panel 1-1", AssetType: valueobject.AssetTypeSwitchboardPanel},
	})

	assert.Equal(t, []*entity.AssetNode{
		{AssetID: "SWB-001-1", AssetName: "Aurora Switchboard 1", AssetType: valueobject.AssetTypeSwitchboard,
			OperationalStatus: &status, ChildCount: 4},
		{AssetID: "PNL-001-1-1", AssetName: "Aurora Panel 1-1", AssetType: valueobject.AssetTypeSwitchboardPanel},
	}, nodes)
	assert.Nil(t, nodes[1].OperationalStatus, "NULL status stays nil (no status recorded)")
	assert.Zero(t, nodes[1].ChildCount)
}

func TestToEntityAssetNodes_NilInputGivesEmptySlice(t *testing.T) {
	nodes := ToEntityAssetNodes(nil)
	assert.NotNil(t, nodes)
	assert.Empty(t, nodes)
}

func TestToEntityAssetNode_Nil(t *testing.T) {
	assert.Nil(t, ToEntityAssetNode(nil))
}
