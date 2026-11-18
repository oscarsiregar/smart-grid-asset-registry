package dto

// DTOs are the HTTP contract: request structs use `query`/`params`/`json`
// tags (parsed by Fiber), response structs use camelCase `json` tags.
// Naming: <UsecaseMethod>Request / <UsecaseMethod>Response.

type GetListProductRequest struct {
	Search string `query:"search"`
	Status string `query:"status"`
	Page   *int   `query:"page"`
	Limit  *int   `query:"limit"`
	Sort   string `query:"sort"`
	Order  string `query:"order"`
}

type GetListProductResponse struct {
	Items      []ProductResponse `json:"items"`
	TotalItems int64             `json:"totalItems"`
}

type GetProductDetailRequest struct {
	ID string `params:"id"`
}

type CreateProductRequest struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Price       float64 `json:"price"`
	Status      string  `json:"status"`
}

type UpdateProductRequest struct {
	ID          string  `params:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Price       float64 `json:"price"`
	Status      string  `json:"status"`
}

type DeleteProductRequest struct {
	ID string `params:"id"`
}

// ProductResponse is shared by list/detail/create/update so every endpoint
// returns the same shape for a product.
type ProductResponse struct {
	ID            string  `json:"id"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Description   *string `json:"description"`
	Price         float64 `json:"price"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"createdAt"`
	CreatedByName string  `json:"createdByName"`
	UpdatedAt     string  `json:"updatedAt"`
	UpdatedByName string  `json:"updatedByName"`
}
