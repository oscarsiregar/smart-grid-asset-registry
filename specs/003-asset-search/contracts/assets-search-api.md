# Contract: asset search API

The search is a filter on the existing list endpoint of feature 002
([../../002-asset-explorer/contracts/assets-api.md](../../002-asset-explorer/contracts/assets-api.md)); without `q`
that endpoint behaves exactly as before. Envelope, error format, base path and rate limiter are unchanged.

## `GET /api/assets?q=…`

| Query parameter | Required | Meaning |
|---|---|---|
| `q` | yes (for a search) | search text; trimmed; 1–100 characters after trimming; matched as "contains", letter case ignored, every character literal |
| `assetType` | no | `SUBSTATION` \| `TRANSFORMER` \| `LV_BOARD` \| `SWITCHBOARD` \| `SWITCHBOARD_PANEL`; only with `q` |
| `page` | no | ≥ 1, default 1; only with `q` |
| `limit` | no | 1–100, default 50; only with `q` |

A blank `q` (absent, empty or only spaces) is not a search: the request is the hierarchy read of feature 002.

**200** — `"message": "Successfully get list asset"`

```json
{
  "items": [
    { "assetId": "PNL-001-1-01", "assetName": "Aurora Panel A1", "assetType": "SWITCHBOARD_PANEL",
      "operationalStatus": "MAINTENANCE", "childCount": 0, "ancestorAssetIds": ["SS-001", "SWB-001-1"] },
    { "assetId": "PNL-001-1-02", "assetName": "Aurora Panel A2", "assetType": "SWITCHBOARD_PANEL",
      "operationalStatus": "IN_SERVICE", "childCount": 0, "ancestorAssetIds": ["SS-001", "SWB-001-1"] }
  ],
  "totalItems": 10
}
```

| Field | Notes |
|---|---|
| `items` | the requested page of matches, never `null`; each is a 002 asset node plus `ancestorAssetIds` |
| `ancestorAssetIds` | the asset's ancestors from its substation down to its parent; `[]` for a substation |
| `totalItems` | the number of **all** matches, not only this page |

**Order**: assets whose ID or name equals `q` (letter case ignored) first; then the rest; each part in the
explorer's order (type: `SUBSTATION`, `TRANSFORMER`, `LV_BOARD`, `SWITCHBOARD`, `SWITCHBOARD_PANEL`; then asset ID
with digit runs compared as numbers). A search for a full asset ID therefore lists that asset first.

A search that matches nothing answers `200` with `{ "items": [], "totalItems": 0 }`.

**400** `INVALID_REQUEST` (no search is run):

| Request | Message |
|---|---|
| `q` longer than 100 characters | `Invalid request: q must be at most 100 characters` |
| `q` with `parentAssetId` | `Invalid request: q and parentAssetId cannot be combined` |
| `assetType` without `q` | `Invalid request: assetType requires q` |
| `page` or `limit` without `q` | `Invalid request: page and limit apply to a search only` |
| `assetType` not a known type | `Invalid request: assetType is not a valid asset type` |
| `page`/`limit` negative or not a number | `Invalid request: page and limit must not be negative` / `invalid query parameters` |

`limit` above 100 is capped at 100 (template `MaxLimit`); a `page` past the last match returns `items: []` with the
real `totalItems`.

## Errors summary (additions)

| Status | `data.errorCode` | When |
|---|---|---|
| 400 | `INVALID_REQUEST` | the parameter rules above |
| 500 | — | database failure; logged, nothing leaked |

The search never answers `404`: an unknown text is an empty result.
