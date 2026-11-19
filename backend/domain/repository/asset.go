package repository

import (
	"context"

	"smart-grid-asset-registry/domain/entity"
	"smart-grid-asset-registry/domain/valueobject"
)

// AssetRepository: every method takes ctx and an optional trx (nil = run
// outside a transaction). Lookups by a list never fail for an empty list:
// they return an empty, non-nil slice without querying.
type AssetRepository interface {
	// FindByAssetIDs returns the stored assets whose asset_id is in assetIDs.
	FindByAssetIDs(ctx context.Context, trx Tx, assetIDs []string) ([]*entity.Asset, error)
	// FindBySerialNumbers returns the stored assets whose serial_number is in
	// serialNumbers.
	FindBySerialNumbers(ctx context.Context, trx Tx, serialNumbers []string) ([]*entity.Asset, error)
	// CreateMany inserts assets in batches, in the given order; callers pass
	// parents before children so the parent foreign key holds.
	CreateMany(ctx context.Context, trx Tx, assets []*entity.Asset) error
	// FindByAssetID returns the asset with exactly this asset_id, or nil, nil
	// when none is stored.
	FindByAssetID(ctx context.Context, trx Tx, assetID string) (*entity.Asset, error)
	// FindNodes returns the assets matching filter as nodes with their child
	// counts, unsorted (the usecase orders them); never nil.
	FindNodes(ctx context.Context, trx Tx, filter FindAssetNodesFilter) ([]*entity.AssetNode, error)
	// FindPath returns the asset and its ancestors as nodes, substation first
	// and the asset last; empty, non-nil when the asset is not stored.
	FindPath(ctx context.Context, trx Tx, assetID string) ([]*entity.AssetNode, error)
	// CountDescendantsByType counts the assets on every level below the
	// asset, by type; a type without descendants is absent; never nil.
	CountDescendantsByType(ctx context.Context, trx Tx, assetID string) (map[valueobject.AssetType]int, error)
	// FindAncestorIDs returns, for each given asset, its ancestor IDs from the
	// substation down to the parent. Every given ID is a key; one without
	// ancestors maps to an empty slice. An empty list runs no query.
	FindAncestorIDs(ctx context.Context, trx Tx, assetIDs []string) (map[string][]string, error)
}
