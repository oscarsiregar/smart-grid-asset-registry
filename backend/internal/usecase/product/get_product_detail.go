package product_usecase

import (
	"context"

	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"

	"smart-grid-asset-registry/pkg/gocommon/pkg/logger"
)

func (pu productUsecase) GetProductDetail(ctx context.Context, input usecase.GetProductDetailInput) (usecase.GetProductDetailOutput, error) {
	log := logger.WithCtx(ctx, pu.log).WithField("usecase", "GetProductDetail")

	if !hasPermission(input.Permissions, valueobject.PermissionGetProductDetail) {
		log.Warn("permission denied")
		return usecase.GetProductDetailOutput{}, apperr.Forbidden(errorcode.PermissionDenied)
	}

	product, err := pu.productRepository.FindByID(ctx, nil, input.ID)
	if err != nil {
		log.WithError(err).Error("failed to find product by id")
		return usecase.GetProductDetailOutput{}, err
	}
	if product == nil {
		return usecase.GetProductDetailOutput{}, apperr.NotFound(errorcode.ProductNotFound)
	}

	log.WithField("productID", input.ID).Info("success")
	return usecase.GetProductDetailOutput{Product: product}, nil
}
