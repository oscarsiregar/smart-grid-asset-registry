package asset_usecase

import (
	"context"

	"smart-grid-asset-registry/domain/usecase"
)

func (u assetUsecase) GetListAsset(_ context.Context, _ usecase.GetListAssetInput) (usecase.GetListAssetOutput, error) {
	return usecase.GetListAssetOutput{}, errNotImplemented
}
