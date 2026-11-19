package asset_usecase

import (
	"context"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"
	"smart-grid-asset-registry/pkg/gocommon/pkg/logger"
)

// GetAssetDetail returns one asset with its path (substation first) and its
// ordered children. The reads run without a transaction: assets are only
// ever added, so they can at most straddle an import commit (research R5).
func (u assetUsecase) GetAssetDetail(ctx context.Context, input usecase.GetAssetDetailInput) (usecase.GetAssetDetailOutput, error) {
	log := logger.WithCtx(ctx, u.log).WithField("usecase", "GetAssetDetail")

	if err := input.Validate(); err != nil {
		return usecase.GetAssetDetailOutput{}, apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	var noTransaction repository.Tx
	asset, err := u.assetRepository.FindByAssetID(ctx, noTransaction, input.AssetID)
	if err != nil {
		log.WithError(err).Error("failed to find asset")
		return usecase.GetAssetDetailOutput{}, err
	}
	if asset == nil {
		return usecase.GetAssetDetailOutput{}, apperr.NotFound(errorcode.AssetNotFound, input.AssetID)
	}

	path, err := u.assetRepository.FindPath(ctx, noTransaction, input.AssetID)
	if err != nil {
		log.WithError(err).Error("failed to find asset path")
		return usecase.GetAssetDetailOutput{}, err
	}

	children, err := u.assetRepository.FindNodes(ctx, noTransaction, repository.FindAssetNodesFilter{ParentAssetID: &input.AssetID})
	if err != nil {
		log.WithError(err).Error("failed to find asset children")
		return usecase.GetAssetDetailOutput{}, err
	}
	entity.SortAssetNodes(children)

	// Which types are counted, and only for a substation, is this usecase's
	// decision; the repository only counts (research R4).
	var descendantCounts *usecase.DescendantCounts
	if asset.AssetType == valueobject.AssetTypeSubstation {
		counts, err := u.assetRepository.CountDescendantsByType(ctx, noTransaction, input.AssetID)
		if err != nil {
			log.WithError(err).Error("failed to count descendants")
			return usecase.GetAssetDetailOutput{}, err
		}
		descendantCounts = toDescendantCounts(counts)
	}

	log.WithField("assetId", input.AssetID).WithField("children", len(children)).Info("success")
	return usecase.GetAssetDetailOutput{
		Asset:            asset,
		Path:             path,
		Children:         children,
		DescendantCounts: descendantCounts,
	}, nil
}
