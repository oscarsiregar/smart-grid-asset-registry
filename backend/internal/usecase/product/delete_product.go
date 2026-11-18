package product_usecase

import (
	"context"

	"backend-template/domain/errorcode"
	"backend-template/domain/usecase"
	"backend-template/domain/valueobject"
	"backend-template/internal/apperr"

	"repo.smf-indonesia.co.id/package-helper/go-common/pkg/logger"
)

// DeleteProduct is a single-statement write, so no transaction is needed.
// Wrap in txManager.WithinTx as soon as a second write joins it.
func (pu productUsecase) DeleteProduct(ctx context.Context, input usecase.DeleteProductInput) error {
	log := logger.WithCtx(ctx, pu.log).WithField("usecase", "DeleteProduct")

	if !hasPermission(input.Permissions, valueobject.PermissionDeleteProduct) {
		log.Warn("permission denied")
		return apperr.Forbidden(errorcode.PermissionDenied)
	}

	product, err := pu.productRepository.FindByID(ctx, nil, input.ID)
	if err != nil {
		log.WithError(err).Error("failed to find product by id")
		return err
	}
	if product == nil {
		return apperr.NotFound(errorcode.ProductNotFound)
	}

	if err := pu.productRepository.Delete(ctx, nil, input.ID); err != nil {
		log.WithError(err).Error("failed to delete product")
		return err
	}

	log.WithField("productID", input.ID).Info("success")
	return nil
}
