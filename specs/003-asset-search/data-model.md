# Data Model: Asset Search

**Feature**: [spec.md](spec.md) · **Research**: [research.md](research.md)

This feature only reads. It adds **no table, column, index or migration**: it searches the `asset` table from
feature 001 (`asset_id`, `asset_name`, `asset_type`) and walks `parent_asset_id` for ancestors.

## Domain additions (`backend/domain`)

### `repository.FindAssetNodesFilter` (extended, feature 002)

| Field | Type | Meaning |
|---|---|---|
| `ParentAssetID` | `*string` | unchanged: with no `Search`, `nil` = top level, otherwise that asset's children |
| `Search` | `*string` (new) | trimmed search text; when set, matches every level by ID or name, letter case ignored, characters literal, and `ParentAssetID` must be `nil` |
| `AssetType` | `*valueobject.AssetType` (new) | only with `Search`: limits matches to one type |

`FindNodes(ctx, trx, filter)` keeps its signature and return type (`[]*entity.AssetNode`, unsorted, never nil,
each with `ChildCount`).

### `repository.AssetRepository.FindAncestorIDs` (new)

`FindAncestorIDs(ctx, trx, assetIDs []string) (map[string][]string, error)` — for each given asset, its ancestor
IDs from the substation down to the parent; a substation maps to an empty slice; unknown IDs are absent; an empty
input returns an empty map without querying.

### `entity.SortSearchMatches(nodes []*AssetNode, text string)` (new, pure)

Stable: assets whose `AssetID` or `AssetName` equals `text` (letter case ignored, `strings.EqualFold`) first, the
rest after; each part in `SortAssetNodes` order (type rank, natural asset ID). `CheckIsExactMatch(node, text)` is
exported for its own tests.

## Usecase contract (`domain/usecase/asset.go`, extended)

| Field | Type | Notes |
|---|---|---|
| `GetListAssetInput.ParentAssetID` | `*string` | unchanged |
| `GetListAssetInput.Search` | `string` (new) | trimmed by the handler; empty = hierarchy mode |
| `GetListAssetInput.AssetType` | `*valueobject.AssetType` (new) | search mode only |
| `GetListAssetInput.Page`, `.Limit` | `int` (new) | search mode only; `SetDefault`: page 1, limit `SearchDefaultLimit` (50); `Limit` ≤ `MaxLimit` (100) |
| `GetListAssetOutput.Assets` | `[]*entity.AssetNode` | hierarchy: every node, sorted; search: the page, in search order |
| `GetListAssetOutput.TotalItems` | `int` (new) | hierarchy: `len(Assets)`; search: number of all matches |
| `GetListAssetOutput.AncestorIDs` | `map[string][]string` (new) | search mode: ancestors of each returned asset; `nil` in hierarchy mode |

`GetListAssetInput.Validate()` (plain errors → `400 INVALID_REQUEST`):

| Rule | Error |
|---|---|
| `Search` longer than 100 characters | `q must be at most 100 characters` |
| `Search` set and `ParentAssetID` set | `q and parentAssetId cannot be combined` |
| `AssetType` set without `Search` | `assetType requires q` |
| `Page` or `Limit` set without `Search` | `page and limit apply to a search only` |
| `Page < 0` or `Limit < 0` | `page and limit must not be negative` |
| `AssetType` not one of the five types | handled by the handler's `valueobject.ParseAssetType` → `assetType is not a valid asset type` |

Constants: `usecase.MaxLimit = 100` (template rule 04), `usecase.SearchDefaultLimit = 50` (spec FR-010),
`usecase.MaxSearchLength = 100`.

## DTOs (`domain/dto/asset.go`, extended)

- `GetListAssetRequest` gains `Q string \`query:"q"\``, `AssetType string \`query:"assetType"\``,
  `Page int \`query:"page"\``, `Limit int \`query:"limit"\``.
- `AssetSearchItemResponse` = the `AssetNodeResponse` fields plus `AncestorAssetIDs []string \`json:"ancestorAssetIds"\``
  (never `null`).
- `AssetSearchResponse { Items []AssetSearchItemResponse; TotalItems int }` (`items` never `null`).

## Frontend (`frontend/src/modules/assets`)

- Query `searchAssets({ q, assetType })` → `GET /api/assets?q=…[&assetType=…]` without `page` or `limit` (the
  backend's default page of 50 applies; the frontend keeps no copy of it); types `AssetSearchItemDto` (`AssetNodeDto` + `ancestorAssetIds:
  string[]`), `AssetSearchResultDto` (`items`, `totalItems`), `AssetSearchRequestDto` (`q`, `assetType | null`).
- Validation schema `assetSearchValidationSchema`: `q` trimmed, 1–100 characters; `assetType` one of `AssetType`
  or `null` (all types). The select's "All types" option is `AssetTypeFilter.ALL` in `pages/assets/libs/enums`,
  mapped to `null` by the form.

## State (frontend, not persisted)

| State | Where | Notes |
|---|---|---|
| Search text and type | URL `?q=` and `?type=` on the asset page | survive refresh, links and Back/Forward (FR-006) |
| Search form values | React Hook Form, defaults from the URL | submitting writes the URL |
| Results | RTK Query cache, tag `ASSET` | refreshed after a committed import (FR-017) |
| Selected asset | the path `/assets/:assetId` (feature 002) | a result is a link to it; the search parameters are dropped |
