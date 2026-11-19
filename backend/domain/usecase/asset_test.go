package usecase

import (
	"strings"
	"testing"

	"smart-grid-asset-registry/domain/valueobject"

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

func pointerTo[T any](value T) *T { return &value }

func TestGetListAssetInput_Validate(t *testing.T) {
	transformer := valueobject.AssetTypeTransformer

	tests := []struct {
		name    string
		input   GetListAssetInput
		wantErr string
	}{
		{"top level", GetListAssetInput{}, ""},
		{"children of one asset", GetListAssetInput{ParentAssetID: pointerTo("SS-001")}, ""},
		{"a search", GetListAssetInput{Search: "PNL"}, ""},
		{"a search with a type and a page", GetListAssetInput{Search: "PNL", AssetType: &transformer, Page: 2, Limit: 10}, ""},
		{"100 characters", GetListAssetInput{Search: strings.Repeat("é", MaxSearchLength)}, ""},
		{"101 characters", GetListAssetInput{Search: strings.Repeat("a", MaxSearchLength+1)}, "q must be at most 100 characters"},
		{"search and parent", GetListAssetInput{Search: "PNL", ParentAssetID: pointerTo("SS-001")}, "q and parentAssetId cannot be combined"},
		{"type without search", GetListAssetInput{AssetType: &transformer}, "assetType requires q"},
		{"page without search", GetListAssetInput{Page: 2}, "page and limit apply to a search only"},
		{"limit without search", GetListAssetInput{Limit: 10}, "page and limit apply to a search only"},
		{"negative page", GetListAssetInput{Search: "PNL", Page: -1}, "page and limit must not be negative"},
		{"negative limit", GetListAssetInput{Search: "PNL", Limit: -1}, "page and limit must not be negative"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestGetListAssetInput_SetDefault(t *testing.T) {
	tests := []struct {
		name      string
		input     GetListAssetInput
		wantPage  int
		wantLimit int
	}{
		{"search defaults", GetListAssetInput{Search: "PNL"}, 1, SearchDefaultLimit},
		{"search keeps a valid page and limit", GetListAssetInput{Search: "PNL", Page: 3, Limit: 20}, 3, 20},
		{"search caps the limit", GetListAssetInput{Search: "PNL", Limit: 500}, 1, MaxLimit},
		{"hierarchy is left untouched", GetListAssetInput{}, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.input
			input.SetDefault()
			assert.Equal(t, tt.wantPage, input.Page)
			assert.Equal(t, tt.wantLimit, input.Limit)
		})
	}
}

func TestSearchConstants(t *testing.T) {
	assert.Equal(t, 100, MaxLimit)
	assert.Equal(t, 50, SearchDefaultLimit)
	assert.Equal(t, 100, MaxSearchLength)
}
