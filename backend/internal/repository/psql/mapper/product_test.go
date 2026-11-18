package mapper

import (
	"testing"
	"time"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/valueobject"
	"smart-grid-asset-registry/internal/repository/psql/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProductMapper_RoundTrip(t *testing.T) {
	desc := "a widget"
	now := time.Now()
	e := &entity.Product{
		ID:            uuid.New(),
		Code:          "P-001",
		Name:          "Widget",
		Description:   &desc,
		Price:         12.5,
		Status:        valueobject.ProductStatusActive,
		CreatedAt:     now,
		CreatedByName: "alice",
		UpdatedAt:     now,
		UpdatedByName: "bob",
	}

	m := ToModelProduct(e)
	assert.Equal(t, &model.Product{
		ID: e.ID, Code: "P-001", Name: "Widget", Description: &desc, Price: 12.5,
		Status: valueobject.ProductStatusActive, CreatedAt: now, CreatedByName: "alice", UpdatedAt: now, UpdatedByName: "bob",
	}, m)

	assert.Equal(t, e, ToEntityProduct(m))
}

func TestProductMapper_Nil(t *testing.T) {
	assert.Nil(t, ToModelProduct(nil))
	assert.Nil(t, ToEntityProduct(nil))
}

func TestToEntityProducts(t *testing.T) {
	result := ToEntityProducts([]*model.Product{{Name: "A"}, {Name: "B"}})
	assert.Len(t, result, 2)
	assert.Equal(t, "B", result[1].Name)

	empty := ToEntityProducts(nil)
	assert.NotNil(t, empty)
	assert.Empty(t, empty)
}
