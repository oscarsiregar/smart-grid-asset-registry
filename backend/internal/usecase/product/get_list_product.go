package product_usecase

import (
	"context"

	"backend-template/domain/errorcode"
	"backend-template/domain/repository"
	"backend-template/domain/usecase"
	"backend-template/domain/valueobject"
	"backend-template/internal/apperr"

	"repo.smf-indonesia.co.id/package-helper/go-common/pkg/logger"
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
