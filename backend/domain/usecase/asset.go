package usecase

import (
	"context"
	"errors"
	"strings"

	"smart-grid-asset-registry/domain/entity"
)

// AssetUsecase reads the stored asset hierarchy (feature 002). One method =
// one endpoint = one file in internal/usecase/asset/. It never writes.
type AssetUsecase interface {
	// GetListAsset returns the top level (every substation) or the immediate
	// children of one asset, complete and ordered (Template Exception 13).
	GetListAsset(ctx context.Context, input GetListAssetInput) (GetListAssetOutput, error)
	// GetAssetDetail returns one asset with its path and its children.
	GetAssetDetail(ctx context.Context, input GetAssetDetailInput) (GetAssetDetailOutput, error)
}

// ---- GetListAsset -----------------------------------------------------------

type GetListAssetInput struct {
	// ParentAssetID: nil = the top level; otherwise that asset's children.
	ParentAssetID *string
}

type GetListAssetOutput struct {
	Assets []*entity.AssetNode // ordered by entity.SortAssetNodes
}

// ---- GetAssetDetail ---------------------------------------------------------

type GetAssetDetailInput struct {
	AssetID string
}

type GetAssetDetailOutput struct {
	Asset    *entity.Asset
	Path     []*entity.AssetNode // substation first, the asset last
	Children []*entity.AssetNode // ordered by entity.SortAssetNodes
}

// Validate checks the request shape only; whether the asset exists is the
// usecase's decision.
func (i GetAssetDetailInput) Validate() error {
	if strings.TrimSpace(i.AssetID) == "" {
		return errors.New("assetId is required")
	}
	return nil
}
