package valueobject

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAssetType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    AssetType
		wantErr bool
	}{
		{"substation", "SUBSTATION", AssetTypeSubstation, false},
		{"transformer", "TRANSFORMER", AssetTypeTransformer, false},
		{"lv_board", "LV_BOARD", AssetTypeLVBoard, false},
		{"switchboard", "SWITCHBOARD", AssetTypeSwitchboard, false},
		{"switchboard_panel", "SWITCHBOARD_PANEL", AssetTypeSwitchboardPanel, false},
		{"lowercase is rejected", "substation", "", true},
		{"surrounding whitespace is rejected", " SUBSTATION ", "", true},
		{"empty", "", "", true},
		{"unknown", "UNKNOWN", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAssetType(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAssetType_IsValid(t *testing.T) {
	assert.NoError(t, AssetTypeSubstation.IsValid())
	assert.NoError(t, AssetTypeTransformer.IsValid())
	assert.NoError(t, AssetTypeLVBoard.IsValid())
	assert.NoError(t, AssetTypeSwitchboard.IsValid())
	assert.NoError(t, AssetTypeSwitchboardPanel.IsValid())
	assert.Error(t, AssetType("UNKNOWN").IsValid())
}

func TestAssetType_String(t *testing.T) {
	assert.Equal(t, "SUBSTATION", AssetTypeSubstation.String())
}

func TestAssetType_Scan(t *testing.T) {
	var s AssetType
	assert.NoError(t, s.Scan("SUBSTATION"))
	assert.Equal(t, AssetTypeSubstation, s)

	var s2 AssetType
	assert.NoError(t, s2.Scan([]byte("SWITCHBOARD_PANEL")))
	assert.Equal(t, AssetTypeSwitchboardPanel, s2)

	var s3 AssetType
	assert.Error(t, s3.Scan(123))

	var s4 AssetType
	assert.Error(t, s4.Scan("UNKNOWN"))
}

func TestAssetType_Value(t *testing.T) {
	v, err := AssetTypeSubstation.Value()
	assert.NoError(t, err)
	assert.Equal(t, driver.Value("SUBSTATION"), v)

	_, err = AssetType("UNKNOWN").Value()
	assert.Error(t, err)
}

func TestAssetType_PermittedParent(t *testing.T) {
	tests := []struct {
		assetType  AssetType
		wantParent AssetType
		wantOK     bool
	}{
		{AssetTypeSubstation, "", false},
		{AssetTypeTransformer, AssetTypeSubstation, true},
		{AssetTypeLVBoard, AssetTypeSubstation, true},
		{AssetTypeSwitchboard, AssetTypeSubstation, true},
		{AssetTypeSwitchboardPanel, AssetTypeSwitchboard, true},
		{AssetType("GENERATOR"), "", false},
	}
	for _, tt := range tests {
		t.Run(tt.assetType.String(), func(t *testing.T) {
			parent, ok := tt.assetType.PermittedParent()
			assert.Equal(t, tt.wantParent, parent)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}
