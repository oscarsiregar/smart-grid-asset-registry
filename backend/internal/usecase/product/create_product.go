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

	"github.com/google/uuid"
	"repo.smf-indonesia.co.id/package-helper/go-common/pkg/logger"
)

func (pu productUsecase) CreateProduct(ctx context.Context, input usecase.CreateProductInput) (usecase.CreateProductOutput, error) {
	log := logger.WithCtx(ctx, pu.log).WithField("usecase", "CreateProduct")

	// 1. authorization first — never leak validation details to a caller
	//    who is not allowed to call this endpoint at all.
	if !hasPermission(input.Permissions, valueobject.PermissionCreateProduct) {
		log.Warn("permission denied")
		return usecase.CreateProductOutput{}, apperr.Forbidden(errorcode.PermissionDenied)
	}

	// 2. input shape
	if err := input.Validate(); err != nil {
		return usecase.CreateProductOutput{}, apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	now := pu.now()
	product := &entity.Product{
		ID:            uuid.New(),
		Code:          strings.TrimSpace(input.Code),
		Name:          strings.TrimSpace(input.Name),
		Description:   input.Description,
		Price:         input.Price,
		Status:        input.Status,
		CreatedAt:     now,
		CreatedByName: input.ActorName,
		UpdatedAt:     now,
		UpdatedByName: input.ActorName,
	}

	// 3. business rules + writes, atomically
	err := pu.txManager.WithinTx(ctx, func(ctx context.Context, trx repository.Tx) error {
		existing, err := pu.productRepository.FindByCode(ctx, trx, product.Code)
		if err != nil {
			log.WithError(err).Error("failed to find product by code")
			return err
		}
		if existing != nil {
			return apperr.Conflict(errorcode.ProductCodeAlreadyExists, product.Code)
		}

		if err := pu.productRepository.Create(ctx, trx, product); err != nil {
			log.WithError(err).Error("failed to create product")
			return err
		}
		return nil
	})
	if err != nil {
		return usecase.CreateProductOutput{}, err
	}

	log.WithField("productID", product.ID).Info("success")
	return usecase.CreateProductOutput{Product: product}, nil
}
