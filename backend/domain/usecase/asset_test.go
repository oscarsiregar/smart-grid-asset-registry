package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAssetDetailInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		assetID string
		wantErr bool
	}{
		{"empty", "", true},
		{"only spaces", "  ", true},
		{"plain ID", "SS-001", false},
		{"ID with a slash and a space", "SS/9 A", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GetAssetDetailInput{AssetID: tt.assetID}.Validate()
			if tt.wantErr {
				assert.EqualError(t, err, "assetId is required")
				return
			}
			assert.NoError(t, err)
		})
	}
}
