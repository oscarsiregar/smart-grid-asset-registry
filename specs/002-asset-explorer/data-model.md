# Data Model: Asset Explorer

**Feature**: [spec.md](spec.md) · **Research**: [research.md](research.md)

This feature only reads. It adds **no table, column, index or migration**: it reads the `asset` table from
feature 001 ([../001-csv-import/data-model.md](../001-csv-import/data-model.md)).

## Stored data used

| Column | Used for |
|---|---|
| `asset_id` (unique, `ux_asset_asset_id`) | identity, addresses, ordering, path walk |
| `parent_asset_id` (FK → `asset.asset_id`, `ix_asset_parent_asset_id`) | top level (`IS NULL`), children, child counts, path and descendant walks |
| `asset_type`, `asset_name`, `operational_status` | tree nodes |
| `voltage_kv`, `rating_kva`, `manufacturer`, `model`, `serial_number`, `commissioned_date` | details |

Invariants guaranteed by feature 001 and relied on here: every non-substation asset has a stored parent of the
permitted type; substations have no parent; there are no cycles; assets are never updated or deleted.

## Domain additions (`backend/domain`)

### `entity.AssetNode` (new)

An asset as it appears in a list or the tree.

| Field | Type | Notes |
|---|---|---|
| `AssetID` | `string` | |
| `AssetName` | `string` | |
| `AssetType` | `valueobject.AssetType` | |
| `OperationalStatus` | `*valueobject.OperationalStatus` | `nil` = no status recorded |
| `ChildCount` | `int` | number of immediate children; `> 0` makes the node expandable (FR-003) |

### `entity.SortAssetNodes(nodes []*AssetNode)` (new, pure)

Orders in place by `AssetType.Rank()`, then by natural asset-ID comparison, then by plain string comparison
(R3). Used for every list the API returns. Exported comparison `entity.CompareAssetIDs(a, b string) int` is
tested on its own.

### `valueobject.AssetType.Rank() int` (new)

SUBSTATION 0, TRANSFORMER 1, LV_BOARD 2, SWITCHBOARD 3, SWITCHBOARD_PANEL 4 — the type order of FR-004.

### `errorcode.AssetNotFound` (new)

`ASSET_NOT_FOUND`, `404`, English message "No asset with ID %s exists." (argument: the asset ID as requested).

## Repository additions (`domain/repository/asset.go`)

| Method | Returns | Not found |
|---|---|---|
| `FindByAssetID(ctx, trx, assetID string)` | `*entity.Asset` | `nil, nil` |
| `FindNodes(ctx, trx, filter FindAssetNodesFilter)` | `[]*entity.AssetNode`, unsorted, non-nil | empty slice |
| `FindPath(ctx, trx, assetID string)` | `[]*entity.AssetNode`, substation first, the asset last | empty slice |
| `CountDescendantsByType(ctx, trx, assetID string)` | `map[valueobject.AssetType]int`, every level below the asset | empty map |

`FindAssetNodesFilter{ ParentAssetID *string }` — `nil` = top level (substations).

A projection struct `model.AssetNode` (not a table) carries the query columns plus `child_count`; the mapper
`ToEntityAssetNodes` converts it.

## Usecase contract (`domain/usecase/asset.go`)

| Method | Input | Output |
|---|---|---|
| `GetListAsset` | `ParentAssetID *string` | `Assets []*entity.AssetNode` (sorted) |
| `GetAssetDetail` | `AssetID string` (`Validate`: not blank) | `Asset *entity.Asset`, `Path []*entity.AssetNode`, `Children []*entity.AssetNode` (sorted), `DescendantCounts *DescendantCounts` (`nil` unless a substation) |

`usecase.DescendantCounts{ Transformers, LVBoards, Switchboards, SwitchboardPanels int }` — built from
`CountDescendantsByType`; a type with no descendants is `0` (FR-013).

Errors: unknown `AssetID` or `ParentAssetID` → `apperr.NotFound(errorcode.AssetNotFound, id)`; blank `AssetID` →
`apperr.BadRequest(errorcode.InvalidRequest, …)`; repository errors returned as-is (→ `500`, logged).

## Frontend types (`frontend/src/modules/assets`)

DTO types mirror [contracts/assets-api.md](contracts/assets-api.md): `AssetNodeDto`, `AssetListDto`
(`items`, `totalItems`), `AssetDto`, `AssetDetailDto` (`asset`, `path`, `children`, `descendantCounts`),
`DescendantCountsDto`. Enums `AssetType` and `OperationalStatus` (values as sent by the API) and
`AssetsApiTag.ASSET`. Display labels for types and statuses live in `pages/assets` (page copy).

## State (frontend, not persisted)

| State | Where | Notes |
|---|---|---|
| Selected asset | the URL path `/assets/:assetId` | survives refresh, deep links, back/forward (FR-017–FR-019) |
| Import panel open | the URL search parameter `?panel=import` | from feature 001 |
| Expanded nodes | local state in the tree | ancestors of the selected asset are added when the selection changes; not kept across refresh (spec assumption) |
| Server data | RTK Query cache, tag `ASSET` | invalidated by a committed import |
