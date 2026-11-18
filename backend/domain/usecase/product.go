package usecase

import (
	"context"
	"errors"
	"strings"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/valueobject"

	"github.com/google/uuid"
)

// ProductUsecase: one method = one endpoint = one file in
// internal/usecase/product/. Every method takes (ctx, XxxInput) and returns
// (XxxOutput, error). Inputs carry the caller's Permissions (and actor name
// for writes) — authorization is decided in the usecase, not the handler.
type ProductUsecase interface {
	GetListProduct(ctx context.Context, input GetListProductInput) (GetListProductOutput, error)
	GetProductDetail(ctx context.Context, input GetProductDetailInput) (GetProductDetailOutput, error)
	CreateProduct(ctx context.Context, input CreateProductInput) (CreateProductOutput, error)
	UpdateProduct(ctx context.Context, input UpdateProductInput) (UpdateProductOutput, error)
	DeleteProduct(ctx context.Context, input DeleteProductInput) error
}

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

// ---- GetListProduct ---------------------------------------------------------

type GetListProductInput struct {
	Search string
	Status *valueobject.ProductStatus
	Page   *int
	Limit  *int
	Sort   string
	Order  string

	Permissions []string
}

type GetListProductOutput struct {
	Items      []*entity.Product
	TotalItems int64
}

// SetDefault fills paging/sorting defaults and caps Limit at MaxLimit.
func (i *GetListProductInput) SetDefault() {
	if i.Page == nil || *i.Page < 1 {
		page := DefaultPage
		i.Page = &page
	}
	if i.Limit == nil || *i.Limit < 1 {
		limit := DefaultLimit
		i.Limit = &limit
	}
	if *i.Limit > MaxLimit {
		limit := MaxLimit
		i.Limit = &limit
	}
	if i.Sort == "" {
		i.Sort = "created_at"
	}
	if i.Order == "" {
		i.Order = "DESC"
	}
}

// ---- GetProductDetail -------------------------------------------------------

type GetProductDetailInput struct {
	ID          uuid.UUID
	Permissions []string
}

type GetProductDetailOutput struct {
	Product *entity.Product
}

// ---- CreateProduct ----------------------------------------------------------

type CreateProductInput struct {
	Code        string
	Name        string
	Description *string
	Price       float64
	Status      valueobject.ProductStatus

	ActorName   string
	Permissions []string
}

type CreateProductOutput struct {
	Product *entity.Product
}

// Validate checks input shape only (required fields, ranges). Business rules
// that need the database (e.g. unique code) belong in the usecase.
// Returns a plain error; the usecase wraps it into apperr.BadRequest.
func (i CreateProductInput) Validate() error {
	return validateProductFields(i.Code, i.Name, i.Price, i.Status)
}

// ---- UpdateProduct ----------------------------------------------------------

type UpdateProductInput struct {
	ID          uuid.UUID
	Code        string
	Name        string
	Description *string
	Price       float64
	Status      valueobject.ProductStatus

	ActorName   string
	Permissions []string
}

type UpdateProductOutput struct {
	Product *entity.Product
}

func (i UpdateProductInput) Validate() error {
	return validateProductFields(i.Code, i.Name, i.Price, i.Status)
}

// ---- DeleteProduct ----------------------------------------------------------

type DeleteProductInput struct {
	ID          uuid.UUID
	Permissions []string
}

// ---- shared -----------------------------------------------------------------

func validateProductFields(code, name string, price float64, status valueobject.ProductStatus) error {
	if strings.TrimSpace(code) == "" {
		return errors.New("code is required")
	}
	if len(code) > 50 {
		return errors.New("code must be at most 50 characters")
	}
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}
	if len(name) > 255 {
		return errors.New("name must be at most 255 characters")
	}
	if price < 0 {
		return errors.New("price must not be negative")
	}
	if err := status.IsValid(); err != nil {
		return err
	}
	return nil
}
