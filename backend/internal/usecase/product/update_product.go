package product_usecase

import (
	"context"
	"strings"

	"backend-template/domain/entity"
	"backend-template/domain/errorcode"
	"backend-template/domain/repository"
	"backend-template/domain/usecase"
	"backend-template/domain/valueobject"
	"backend-template/internal/apperr"

	"repo.smf-indonesia.co.id/package-helper/go-common/pkg/logger"
)

func (pu productUsecase) UpdateProduct(ctx context.Context, input usecase.UpdateProductInput) (usecase.UpdateProductOutput, error) {
	log := logger.WithCtx(ctx, pu.log).WithField("usecase", "UpdateProduct")

	if !hasPermission(input.Permissions, valueobject.PermissionUpdateProduct) {
		log.Warn("permission denied")
		return usecase.UpdateProductOutput{}, apperr.Forbidden(errorcode.PermissionDenied)
	}

	if err := input.Validate(); err != nil {
		return usecase.UpdateProductOutput{}, apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	var product *entity.Product

	err := pu.txManager.WithinTx(ctx, func(ctx context.Context, trx repository.Tx) error {
		var err error
		product, err = pu.productRepository.FindByID(ctx, trx, input.ID)
		if err != nil {
			log.WithError(err).Error("failed to find product by id")
			return err
		}
		if product == nil {
			return apperr.NotFound(errorcode.ProductNotFound)
		}

		code := strings.TrimSpace(input.Code)
		if code != product.Code {
			existing, err := pu.productRepository.FindByCode(ctx, trx, code)
			if err != nil {
				log.WithError(err).Error("failed to find product by code")
				return err
			}
			if existing != nil && existing.ID != product.ID {
				return apperr.Conflict(errorcode.ProductCodeAlreadyExists, code)
			}
		}

		product.Code = code
		product.Name = strings.TrimSpace(input.Name)
		product.Description = input.Description
		product.Price = input.Price
		product.Status = input.Status
		product.UpdatedAt = pu.now()
		product.UpdatedByName = input.ActorName

		if err := pu.productRepository.Update(ctx, trx, product); err != nil {
			log.WithError(err).Error("failed to update product")
			return err
		}
		return nil
	})
	if err != nil {
		return usecase.UpdateProductOutput{}, err
	}

	log.WithField("productID", input.ID).Info("success")
	return usecase.UpdateProductOutput{Product: product}, nil
}
