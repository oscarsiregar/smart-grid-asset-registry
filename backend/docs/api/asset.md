# API: asset explorer

Reads the stored asset hierarchy: the top level, the children of one asset, and one asset's details. Source
of truth for the module's HTTP contract; it matches `specs/002-asset-explorer/contracts/assets-api.md`.
The module only reads; assets are created by imports ([asset-import.md](asset-import.md)).

- Base path: `/api` (no version prefix, constitution Template Exception 5)
- Authentication: none (Template Exception 1)
- Handler: `internal/handler/asset` · Usecase: `internal/usecase/asset` · Repository:
  `internal/repository/psql/asset` · Ordering: `domain/entity/asset_node.go`

## Response envelope

Every response uses the go-common envelope:

```json
{ "code": 200, "status": "success", "message": "Successfully get list asset", "data": { } }
```

Errors that carry an application error code put it in `data.errorCode` (Template Exception 4):

```json
{ "code": 404, "status": "error", "message": "No asset with ID SS-999 exists.", "data": { "errorCode": "ASSET_NOT_FOUND" } }
```

An unexpected failure is `500` with `"message": "Internal Server Error"` and `"data": null`; the
details are only logged.

## Complete lists, fixed order

Lists are **complete, not paged** (Template Exception 13): the explorer shows every child of a node at
once. There are no `page` / `limit` parameters, and `totalItems` always equals the number of `items`.
`items` is never `null`.

Every list is ordered by the backend; clients show it as received:

1. by type: `SUBSTATION`, `TRANSFORMER`, `LV_BOARD`, `SWITCHBOARD`, `SWITCHBOARD_PANEL`;
2. then by asset ID in natural order: runs of digits compare as numbers (`SS-2` before `SS-10`), other
   characters by Unicode code point, and a digit run comes before a non-digit character (`SS-012` before
   `SS-T01`). When leading zeros are the only difference, fewer zeros come first (`SS-1` before `SS-01`).

## Asset node

The shape of every list item and every path step.

```json
{
  "assetId": "SWB-001-1",
  "assetName": "Aurora 22kV Switchboard A",
  "assetType": "SWITCHBOARD",
  "operationalStatus": "OUT_OF_SERVICE",
  "childCount": 5
}
```

| Field | Type | Notes |
|---|---|---|
| `assetId` | string | |
| `assetName` | string | |
| `assetType` | `SUBSTATION` \| `TRANSFORMER` \| `LV_BOARD` \| `SWITCHBOARD` \| `SWITCHBOARD_PANEL` | |
| `operationalStatus` | `IN_SERVICE` \| `MAINTENANCE` \| `OUT_OF_SERVICE` \| `null` | `null` = no status recorded |
| `childCount` | integer ≥ 0 | immediate children; `0` = not expandable |

## `GET /api/assets`

The top level (every substation), or the immediate children of one asset.

| Query parameter | Required | Meaning |
|---|---|---|
| `parentAssetId` | no | asset whose immediate children to list; absent or empty = top level |

**200** — `"message": "Successfully get list asset"`

```json
{ "items": [ { "assetId": "SS-001", "assetName": "Aurora Substation (North)", "assetType": "SUBSTATION",
               "operationalStatus": "IN_SERVICE", "childCount": 5 } ],
  "totalItems": 13 }
```

An empty registry answers `200` with `{ "items": [], "totalItems": 0 }`; a stored asset without children
answers `200` with an empty list.

**404** `ASSET_NOT_FOUND` — `parentAssetId` names no stored asset.

## `GET /api/assets/:assetId`

One asset's details. `:assetId` is percent-encoded (`SS%2F9%20A` for `SS/9 A`); the handler decodes it.
IDs match exactly, letter case included.

**200** — `"message": "Successfully get asset detail"`

```json
{
  "asset": {
    "assetId": "PNL-001-1-01",
    "parentAssetId": "SWB-001-1",
    "assetType": "SWITCHBOARD_PANEL",
    "assetName": "Aurora Panel A1",
    "operationalStatus": "MAINTENANCE",
    "voltageKv": 22,
    "ratingKva": null,
    "manufacturer": "Helios Gridworks",
    "model": "PNL-22",
    "serialNumber": "PN0010101",
    "commissionedDate": "2010-09-11"
  },
  "path": [
    { "assetId": "SS-001", "assetName": "Aurora Substation (North)", "assetType": "SUBSTATION",
      "operationalStatus": "IN_SERVICE", "childCount": 5 },
    { "assetId": "SWB-001-1", "assetName": "Aurora 22kV Switchboard A", "assetType": "SWITCHBOARD",
      "operationalStatus": "OUT_OF_SERVICE", "childCount": 5 },
    { "assetId": "PNL-001-1-01", "assetName": "Aurora Panel A1", "assetType": "SWITCHBOARD_PANEL",
      "operationalStatus": "MAINTENANCE", "childCount": 0 }
  ],
  "children": [],
  "descendantCounts": null
}
```

| Field | Notes |
|---|---|
| `asset` | every stored attribute; blank CSV values are `null`; `parentAssetId` is `null` only for a substation; numbers are JSON numbers (so `22.50` is sent as `22.5`); `commissionedDate` is `YYYY-MM-DD` |
| `path` | asset nodes from the substation down to the asset itself (1–3 entries) |
| `children` | the asset's immediate children, ordered as above; empty when it has none |
| `descendantCounts` | only for a substation, otherwise `null`: `{ "transformers": n, "lvBoards": n, "switchboards": n, "switchboardPanels": n }`, counting every level below it; a type with none is `0` |

**404** `ASSET_NOT_FOUND` — no stored asset has exactly this ID: `"No asset with ID ss-001 exists."`

**400** `INVALID_REQUEST` — the path segment is not valid percent-encoding (for example `%zz`):
`"Invalid request: assetId must be valid percent-encoding"`.

### IDs `.` and `..`

Imports accept them, but the URL standard treats them (even encoded as `%2E`) as dot segments, so
clients rewrite `/api/assets/..` before it is sent. Such an asset can be listed and its children read
(`?parentAssetId=..` is a query value), but its details cannot be requested by path.

## Errors

| Status | `data.errorCode` | When |
|---|---|---|
| 400 | `INVALID_REQUEST` | bad percent-encoding in `:assetId` |
| 404 | `ASSET_NOT_FOUND` | unknown `:assetId` or `parentAssetId` |
| 429 | — | rate limiter (20 non-2xx responses per 15 minutes per client) |
| 500 | — | database or other technical failure; logged, nothing leaked |

## Reads

The details are read without a transaction (assets are only ever added): the asset, its path (a
recursive CTE upwards), its children with their child counts, and — for a substation — a recursive CTE
downwards grouped by type. Every value from the request is a bound `?` parameter. With ~55 000 stored
assets the integration test measures the top level, a 5 000-child list and a substation's details in
well under 100 ms each (`internal/integrationtest/asset_explorer_scale_test.go`).
