package asset_usecase

import (
	"context"

	"smart-grid-asset-registry/domain/usecase"
)

func (u assetUsecase) GetAssetDetail(_ context.Context, _ usecase.GetAssetDetailInput) (usecase.GetAssetDetailOutput, error) {
	return usecase.GetAssetDetailOutput{}, errNotImplemented
}
