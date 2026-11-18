package usecase

import (
	"strings"
	"testing"

	"backend-template/domain/valueobject"

	"github.com/stretchr/testify/assert"
)

func intPtr(v int) *int { return &v }

func TestGetListProductInput_SetDefault(t *testing.T) {
	tests := []struct {
		name      string
		input     GetListProductInput
		wantPage  int
		wantLimit int
		wantSort  string
		wantOrder string
	}{
		{"nil page/limit get defaults", GetListProductInput{}, DefaultPage, DefaultLimit, "created_at", "DESC"},
		{"zero/negative page/limit get defaults", GetListProductInput{Page: intPtr(0), Limit: intPtr(-1)}, DefaultPage, DefaultLimit, "created_at", "DESC"},
		{"limit above max is capped", GetListProductInput{Page: intPtr(3), Limit: intPtr(1000)}, 3, MaxLimit, "created_at", "DESC"},
		{"explicit values are kept", GetListProductInput{Page: intPtr(2), Limit: intPtr(25), Sort: "name", Order: "ASC"}, 2, 25, "name", "ASC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := tt.input
			in.SetDefault()
			assert.Equal(t, tt.wantPage, *in.Page)
			assert.Equal(t, tt.wantLimit, *in.Limit)
			assert.Equal(t, tt.wantSort, in.Sort)
			assert.Equal(t, tt.wantOrder, in.Order)
		})
	}
}

func TestCreateProductInput_Validate(t *testing.T) {
	valid := CreateProductInput{Code: "P-001", Name: "Widget", Price: 10, Status: valueobject.ProductStatusActive}

	tests := []struct {
		name    string
		mutate  func(i *CreateProductInput)
		wantErr string
	}{
		{"valid", func(i *CreateProductInput) {}, ""},
		{"blank code", func(i *CreateProductInput) { i.Code = "  " }, "code is required"},
		{"code too long", func(i *CreateProductInput) { i.Code = strings.Repeat("x", 51) }, "code must be at most 50 characters"},
		{"blank name", func(i *CreateProductInput) { i.Name = "" }, "name is required"},
		{"name too long", func(i *CreateProductInput) { i.Name = strings.Repeat("x", 256) }, "name must be at most 255 characters"},
		{"negative price", func(i *CreateProductInput) { i.Price = -1 }, "price must not be negative"},
		{"invalid status", func(i *CreateProductInput) { i.Status = "UNKNOWN" }, "invalid product status"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := valid
			tt.mutate(&in)
			err := in.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestUpdateProductInput_Validate(t *testing.T) {
	assert.NoError(t, UpdateProductInput{Code: "P-001", Name: "Widget", Status: valueobject.ProductStatusInactive}.Validate())
	assert.EqualError(t, UpdateProductInput{Name: "Widget", Status: valueobject.ProductStatusActive}.Validate(), "code is required")
}
