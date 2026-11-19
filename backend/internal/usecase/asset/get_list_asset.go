package asset_usecase

import (
	"context"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/internal/apperr"
	"smart-grid-asset-registry/pkg/gocommon/pkg/logger"
)

// GetListAsset returns the complete, ordered top level (ParentAssetID nil)
// or the children of one asset (Template Exception 13). An unknown parent is
// 404 ASSET_NOT_FOUND; a stored asset without children is an empty list.
func (u assetUsecase) GetListAsset(ctx context.Context, input usecase.GetListAssetInput) (usecase.GetListAssetOutput, error) {
	log := logger.WithCtx(ctx, u.log).WithField("usecase", "GetListAsset")

	var noTransaction repository.Tx
	nodes, err := u.assetRepository.FindNodes(ctx, noTransaction, repository.FindAssetNodesFilter{ParentAssetID: input.ParentAssetID})
	if err != nil {
		log.WithError(err).Error("failed to find asset nodes")
		return usecase.GetListAssetOutput{}, err
	}

	// An empty child list needs one more read to tell "no children" from
	// "no such parent".
	if len(nodes) == 0 && input.ParentAssetID != nil {
		parent, err := u.assetRepository.FindByAssetID(ctx, noTransaction, *input.ParentAssetID)
		if err != nil {
			log.WithError(err).Error("failed to find parent asset")
			return usecase.GetListAssetOutput{}, err
		}
		if parent == nil {
			return usecase.GetListAssetOutput{}, apperr.NotFound(errorcode.AssetNotFound, *input.ParentAssetID)
		}
	}

	entity.SortAssetNodes(nodes)

	log.WithField("parentAssetId", parentLogValue(input.ParentAssetID)).WithField("count", len(nodes)).Info("success")
	return usecase.GetListAssetOutput{Assets: nodes}, nil
}
