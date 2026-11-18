package product_usecase

import (
	"context"

	"smart-grid-asset-registry/domain/errorcode"
	"smart-grid-asset-registry/domain/repository"
	"smart-grid-asset-registry/domain/usecase"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/apperr"

	"smart-grid-asset-registry/pkg/gocommon/pkg/logger"
)

func (pu productUsecase) GetListProduct(ctx context.Context, input usecase.GetListProductInput) (usecase.GetListProductOutput, error) {
	log := logger.WithCtx(ctx, pu.log).WithField("usecase", "GetListProduct")

	if !hasPermission(input.Permissions, valueobject.PermissionGetListProduct) {
		log.Warn("permission denied")
		return usecase.GetListProductOutput{}, apperr.Forbidden(errorcode.PermissionDenied)
	}

	input.SetDefault()

	items, total, err := pu.productRepository.Find(ctx, nil, repository.FindProductFilter{
		Search: input.Search,
		Status: input.Status,
		Page:   *input.Page,
		Limit:  *input.Limit,
		Sort:   input.Sort,
		Order:  input.Order,
	})
	if err != nil {
		log.WithError(err).Error("failed to find products")
		return usecase.GetListProductOutput{}, err
	}

	log.WithField("count", len(items)).Info("success")
	return usecase.GetListProductOutput{Items: items, TotalItems: total}, nil
}
