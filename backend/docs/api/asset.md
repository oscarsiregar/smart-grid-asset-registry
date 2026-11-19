# API: asset explorer and search

Reads the stored asset hierarchy: the top level, the children of one asset, one asset's details, and the search
by asset ID or name. Source of truth for the module's HTTP contract; it matches
`specs/002-asset-explorer/contracts/assets-api.md` and `specs/003-asset-search/contracts/assets-search-api.md`.
The module only reads; assets are created by imports ([asset-import.md](asset-import.md)).

- Base path: `/api` (no version prefix, constitution Template Exception 5)
- Authentication: none (Template Exception 1)
- Handler: `internal/handler/asset` · Usecase: `internal/usecase/asset` · Repository:
  `internal/repository/psql/asset` · Ordering: `domain/entity/asset_node.go`, `domain/entity/asset_search.go`

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

The hierarchy lists are **complete, not paged** (Template Exception 13): the explorer shows every child of a
node at once. The [search](#search-get-apiassetsq) is the exception: it is paged (Template Exception 14). There are no `page` / `limit` parameters, and `totalItems` always equals the number of `items`.
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

A non-blank `q` turns the request into a [search](#search-get-apiassetsq).

**200** — `"message": "Successfully get list asset"`

```json
{ "items": [ { "assetId": "SS-001", "assetName": "Aurora Substation (North)", "assetType": "SUBSTATION",
               "operationalStatus": "IN_SERVICE", "childCount": 5 } ],
  "totalItems": 13 }
```

An empty registry answers `200` with `{ "items": [], "totalItems": 0 }`; a stored asset without children
answers `200` with an empty list.

**404** `ASSET_NOT_FOUND` — `parentAssetId` names no stored asset.

## Search: `GET /api/assets?q=…`

Finds assets whose ID or name contains the text, at every level of the hierarchy (feature 003). It is a filter on
the list read above; without a non-blank `q` the request is the hierarchy read, unchanged.

| Query parameter | Required | Meaning |
|---|---|---|
| `q` | yes (for a search) | search text; trimmed; 1–100 characters after trimming; matched as "contains", letter case ignored, every character literal (`%` and `_` match only themselves) |
| `assetType` | no | `SUBSTATION` \| `TRANSFORMER` \| `LV_BOARD` \| `SWITCHBOARD` \| `SWITCHBOARD_PANEL`; only with `q` |
| `page` | no | ≥ 1, default 1; only with `q` |
| `limit` | no | 1–100, default 50; values above 100 are capped at 100; only with `q` |

A blank `q` (absent, empty or only spaces) is not a search.

**200** — `"message": "Successfully get list asset"`

```json
{
  "items": [
    { "assetId": "PNL-001-1-01", "assetName": "Aurora Panel A1", "assetType": "SWITCHBOARD_PANEL",
      "operationalStatus": "MAINTENANCE", "childCount": 0, "ancestorAssetIds": ["SS-001", "SWB-001-1"] }
  ],
  "totalItems": 10
}
```

| Field | Notes |
|---|---|
| `items` | the requested page of matches, never `null`; each is an [asset node](#asset-node) plus `ancestorAssetIds` |
| `ancestorAssetIds` | the asset's ancestors from its substation down to its parent; `[]` for a substation |
| `totalItems` | the number of **all** matches, not only this page |

**Order.** Assets whose ID or name equals `q` (letter case ignored) come first, then the rest; each part in the
order of the lists above (type, then natural asset ID). A search for a full asset ID therefore lists that asset
first. A search that matches nothing answers `200` with `{ "items": [], "totalItems": 0 }`; a `page` past the last
match answers `items: []` with the real `totalItems`.

**Paging (Template Exception 14).** The repository loads every match (`ILIKE` on `asset_id` and `asset_name` with
`querysafety.ContainsPattern`); the usecase orders them with `entity.SortSearchMatches`, cuts the page, and reads
the ancestors of that page only (`FindAncestorIDs`, one recursive CTE). The natural ID order is not expressible in
SQL, which is why the page is cut in Go.

**400** `INVALID_REQUEST` — no search is run:

| Request | Message |
|---|---|
| `q` longer than 100 characters | `Invalid request: q must be at most 100 characters` |
| `q` with `parentAssetId` | `Invalid request: q and parentAssetId cannot be combined` |
| `assetType` without `q` | `Invalid request: assetType requires q` |
| `page` or `limit` without `q` | `Invalid request: page and limit apply to a search only` |
| `assetType` not a known type | `Invalid request: assetType is not a valid asset type` |
| `page` / `limit` negative | `Invalid request: page and limit must not be negative` |
| `page` / `limit` not a number | `Invalid request: invalid query parameters` |

The search never answers `404`: an unknown text is an empty result.

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
| 400 | `INVALID_REQUEST` | bad percent-encoding in `:assetId`; a search parameter rule (see [Search](#search-get-apiassetsq)) |
| 404 | `ASSET_NOT_FOUND` | unknown `:assetId` or `parentAssetId` |
| 429 | — | rate limiter (20 non-2xx responses per 15 minutes per client) |
| 500 | — | database or other technical failure; logged, nothing leaked |

## Reads

The details are read without a transaction (assets are only ever added): the asset, its path (a
recursive CTE upwards), its children with their child counts, and — for a substation — a recursive CTE
downwards grouped by type. Every value from the request is a bound `?` parameter. With ~55 000 stored
assets the integration test measures the top level, a 5 000-child list and a substation's details in
well under 100 ms each (`internal/integrationtest/asset_explorer_scale_test.go`). The same test times three
searches against the 1 s budget of feature 003 (SC-005): one asset (`WIDE-4999`) about 70 ms, `PNL` limited to
switchboard panels about 270 ms, and `-` (every asset matches, so every row is loaded and sorted) about 770 ms.
The search logs the text's length and the counts, never the text itself.
