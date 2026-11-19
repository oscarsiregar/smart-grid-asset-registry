# Contract: assets API

Read-only endpoints of the new backend module `asset` (handler `internal/handler/asset`, usecase
`internal/usecase/asset`). After implementation, `backend/docs/api/asset.md` holds the same contract.

- Base path `/api`, no version prefix (Template Exception 5); no authentication (Template Exception 1).
- go-common envelope `{ code, status, message, data }`; errors with an application code carry it in
  `data.errorCode` (Template Exception 4). A `500` is `"Internal Server Error"` with `data: null`.
- JSON in camelCase. Lists are `{ "items": [...], "totalItems": N }`, `items` never `null`.
- Lists are **complete, not paged** (proposed Template Exception 13, research R0).
- Every list is ordered by the backend (research R3): type order SUBSTATION, TRANSFORMER, LV_BOARD,
  SWITCHBOARD, SWITCHBOARD_PANEL, then asset ID with digit runs compared as numbers (`SS-2` before `SS-10`).

## Asset node

The shape of every list item and every path step.

```json
{
  "assetId": "SWB-001-1",
  "assetName": "Aurora Switchboard 1",
  "assetType": "SWITCHBOARD",
  "operationalStatus": "IN_SERVICE",
  "childCount": 4
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

The top level of the hierarchy (every substation), or the immediate children of one asset.

| Query parameter | Required | Meaning |
|---|---|---|
| `parentAssetId` | no | asset ID whose immediate children to list; absent or empty = top level |

**200** — `"message": "Successfully get list asset"`

```json
{ "items": [ { "assetId": "SS-001", "assetName": "Aurora Substation (North)", "assetType": "SUBSTATION",
               "operationalStatus": "IN_SERVICE", "childCount": 6 } ],
  "totalItems": 13 }
```

An empty registry answers `200` with `{ "items": [], "totalItems": 0 }`.

**404** `ASSET_NOT_FOUND` — `parentAssetId` names no stored asset: `"No asset with ID SS-999 exists."`

A stored asset without children answers `200` with an empty list.

## `GET /api/assets/:assetId`

One asset's details. `:assetId` is percent-encoded (`SS%2F1` for `SS/1`, `A%20B` for `A B`); the handler decodes
it (research R2).

**200** — `"message": "Successfully get asset detail"`

```json
{
  "asset": {
    "assetId": "PNL-001-1-3",
    "parentAssetId": "SWB-001-1",
    "assetType": "SWITCHBOARD_PANEL",
    "assetName": "Aurora Panel 1-3",
    "operationalStatus": null,
    "voltageKv": 0.4,
    "ratingKva": null,
    "manufacturer": "Helios Gridworks",
    "model": null,
    "serialNumber": "PNL0011C",
    "commissionedDate": "2011-06-20"
  },
  "path": [
    { "assetId": "SS-001", "assetName": "Aurora Substation (North)", "assetType": "SUBSTATION",
      "operationalStatus": "IN_SERVICE", "childCount": 6 },
    { "assetId": "SWB-001-1", "assetName": "Aurora Switchboard 1", "assetType": "SWITCHBOARD",
      "operationalStatus": "IN_SERVICE", "childCount": 4 },
    { "assetId": "PNL-001-1-3", "assetName": "Aurora Panel 1-3", "assetType": "SWITCHBOARD_PANEL",
      "operationalStatus": null, "childCount": 0 }
  ],
  "children": [],
  "descendantCounts": null
}
```

(Values are illustrative.)

| Field | Notes |
|---|---|
| `asset` | every stored attribute; blank CSV values are `null`; `parentAssetId` is `null` only for a substation; numbers are JSON numbers; `commissionedDate` is `YYYY-MM-DD` |
| `path` | asset nodes from the substation down to the asset itself (1–3 entries); used for the path in the details (FR-011) and to expand the tree (FR-019) |
| `children` | the asset's immediate children as asset nodes, ordered; empty for assets without children |
| `descendantCounts` | only for a substation, otherwise `null`: `{ "transformers": n, "lvBoards": n, "switchboards": n, "switchboardPanels": n }`, counting every level below the substation; a type with none is `0` (FR-013) |

**404** `ASSET_NOT_FOUND` — no stored asset has exactly this ID (letter case included):
`"No asset with ID ss-001 exists."` The frontend shows this as the *not found* state (FR-020).

**400** `INVALID_REQUEST` — the path segment is not valid percent-encoding (for example `%zz`).

## Errors summary

| Status | `data.errorCode` | When |
|---|---|---|
| 400 | `INVALID_REQUEST` | bad percent-encoding in `:assetId` |
| 404 | `ASSET_NOT_FOUND` | unknown `:assetId` or `parentAssetId` |
| 429 | — | template rate limiter (20 non-2xx responses per 15 minutes) |
| 500 | — | database or other technical failure; logged, nothing leaked |
