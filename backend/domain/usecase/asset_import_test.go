package usecase

import (
	"io"
	"strings"
	"testing"

	"smart-grid-asset-registry/domain/valueobject"

	"github.com/stretchr/testify/assert"
)

func TestAssetImportInputs_Validate(t *testing.T) {
	file := strings.NewReader("asset_id\n")

	tests := []struct {
		name     string
		fileName string
		file     io.Reader
		mode     valueobject.ImportMode
		wantErr  string
	}{
		{"valid partial", "grid_assets.csv", file, valueobject.ImportModePartial, ""},
		{"valid strict", "grid_assets.csv", file, valueobject.ImportModeStrict, ""},
		{"missing file", "grid_assets.csv", nil, valueobject.ImportModePartial, "file is required"},
		{"missing file name", "  ", file, valueobject.ImportModePartial, "file name is required"},
		{"invalid mode", "grid_assets.csv", file, valueobject.ImportMode("ALL"), "invalid import mode"},
		{"empty mode", "grid_assets.csv", file, "", "invalid import mode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validateErr := ValidateAssetImportInput{FileName: tt.fileName, File: tt.file, Mode: tt.mode}.Validate()
			createErr := CreateAssetImportInput{FileName: tt.fileName, File: tt.file, Mode: tt.mode}.Validate()

			for _, err := range []error{validateErr, createErr} {
				if tt.wantErr == "" {
					assert.NoError(t, err)
					continue
				}
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestImportConstants(t *testing.T) {
	assert.Equal(t, "system", SystemActorName)
	assert.Equal(t, 5*1024*1024, MaxFileBytes)
}
