package usecase

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/valueobject"
)

// AssetUsecase reads the stored asset hierarchy (feature 002). One method =
// one endpoint = one file in internal/usecase/asset/. It never writes.
type AssetUsecase interface {
	// GetListAsset returns the top level (every substation) or the immediate
	// children of one asset, complete and ordered (Template Exception 13);
	// with a search text it returns one page of matching assets instead
	// (feature 003, Template Exception 14).
	GetListAsset(ctx context.Context, input GetListAssetInput) (GetListAssetOutput, error)
	// GetAssetDetail returns one asset with its path and its children.
	GetAssetDetail(ctx context.Context, input GetAssetDetailInput) (GetAssetDetailOutput, error)
}

// ---- GetListAsset -----------------------------------------------------------

const (
	// MaxLimit is the largest page a list accepts (template rule 04).
	MaxLimit = 100
	// SearchDefaultLimit is the search page size: the first 50 matches
	// (spec 003 FR-010).
	SearchDefaultLimit = 50
	// MaxSearchLength is the longest search text, in characters.
	MaxSearchLength = 100
)

type GetListAssetInput struct {
	// ParentAssetID: nil = the top level; otherwise that asset's children.
	// Hierarchy mode only.
	ParentAssetID *string
	// Search, trimmed by the handler: empty = hierarchy mode; otherwise a
	// search of every level by ID or name.
	Search string
	// AssetType limits a search to one type (search mode only).
	AssetType *valueobject.AssetType
	// Page and Limit select the page of matches (search mode only).
	Page  int
	Limit int
}

type GetListAssetOutput struct {
	// Assets: hierarchy mode, every node ordered by entity.SortAssetNodes;
	// search mode, the requested page ordered by entity.SortSearchMatches.
	Assets []*entity.AssetNode
	// TotalItems: hierarchy mode, len(Assets); search mode, all matches.
	TotalItems int
	// AncestorIDs: search mode, the ancestors of each returned asset,
	// substation first; nil in hierarchy mode.
	AncestorIDs map[string][]string
}

// Validate checks the request shape: which parameters go together. A blank
// search is hierarchy mode, so the search parameters need a search text.
func (i GetListAssetInput) Validate() error {
	if i.Search == "" {
		switch {
		case i.AssetType != nil:
			return errors.New("assetType requires q")
		case i.Page != 0 || i.Limit != 0:
			return errors.New("page and limit apply to a search only")
		}
		return nil
	}

	switch {
	case utf8.RuneCountInString(i.Search) > MaxSearchLength:
		return errors.New("q must be at most 100 characters")
	case i.ParentAssetID != nil:
		return errors.New("q and parentAssetId cannot be combined")
	case i.Page < 0 || i.Limit < 0:
		return errors.New("page and limit must not be negative")
	}
	return nil
}

// SetDefault fills a search's page (1) and page size (SearchDefaultLimit) and
// caps the page size at MaxLimit. Hierarchy inputs are left untouched, so
// Validate still sees page or limit given without a search.
func (i *GetListAssetInput) SetDefault() {
	if i.Search == "" {
		return
	}
	if i.Page == 0 {
		i.Page = 1
	}
	if i.Limit == 0 {
		i.Limit = SearchDefaultLimit
	}
	if i.Limit > MaxLimit {
		i.Limit = MaxLimit
	}
}

// ---- GetAssetDetail ---------------------------------------------------------

type GetAssetDetailInput struct {
	AssetID string
}

type GetAssetDetailOutput struct {
	Asset    *entity.Asset
	Path     []*entity.AssetNode // substation first, the asset last
	Children []*entity.AssetNode // ordered by entity.SortAssetNodes
	// DescendantCounts is nil unless the asset is a substation (spec FR-013).
	DescendantCounts *DescendantCounts
}

// DescendantCounts are the assets beneath a substation, on every level, by
// type; a type with none is 0.
type DescendantCounts struct {
	Transformers      int
	LVBoards          int
	Switchboards      int
	SwitchboardPanels int
}

// Validate checks the request shape only; whether the asset exists is the
// usecase's decision.
func (i GetAssetDetailInput) Validate() error {
	if strings.TrimSpace(i.AssetID) == "" {
		return errors.New("assetId is required")
	}
	return nil
}
