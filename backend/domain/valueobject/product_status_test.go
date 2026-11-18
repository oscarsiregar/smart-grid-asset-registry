package valueobject

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseProductStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ProductStatus
		wantErr bool
	}{
		{"active", "ACTIVE", ProductStatusActive, false},
		{"inactive", "INACTIVE", ProductStatusInactive, false},
		{"lowercase is rejected", "active", "", true},
		{"unknown", "UNKNOWN", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseProductStatus(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProductStatus_IsValid(t *testing.T) {
	assert.NoError(t, ProductStatusActive.IsValid())
	assert.NoError(t, ProductStatusInactive.IsValid())
	assert.Error(t, ProductStatus("UNKNOWN").IsValid())
}

func TestProductStatus_String(t *testing.T) {
	assert.Equal(t, "ACTIVE", ProductStatusActive.String())
}

func TestProductStatus_Scan(t *testing.T) {
	var s ProductStatus
	assert.NoError(t, s.Scan("ACTIVE"))
	assert.Equal(t, ProductStatusActive, s)

	var s2 ProductStatus
	assert.NoError(t, s2.Scan([]byte("INACTIVE")))
	assert.Equal(t, ProductStatusInactive, s2)

	var s3 ProductStatus
	assert.Error(t, s3.Scan(123))

	var s4 ProductStatus
	assert.Error(t, s4.Scan("UNKNOWN"))
}

func TestProductStatus_Value(t *testing.T) {
	v, err := ProductStatusActive.Value()
	assert.NoError(t, err)
	assert.Equal(t, driver.Value("ACTIVE"), v)

	_, err = ProductStatus("UNKNOWN").Value()
	assert.Error(t, err)
}
