# Contract: Imports API

**Feature**: [../spec.md](../spec.md) · **Decisions**: [../research.md](../research.md) R4, R9
**Base path**: `/api` (exception 5) · **Auth**: none (exception 1)

The source of truth after implementation is `backend/docs/api/asset-import.md`; it must match this
file.

## Envelope (go-common, local re-implementation)

Every response:

```json
{ "code": 201, "status": "success", "message": "Successfully create asset import", "data": { } }
```

Errors carrying an application error code (exception 4):

```json
{ "code": 422, "status": "error", "message": "The file is missing required columns: asset_name, model", "data": { "errorCode": "IMPORT_MISSING_COLUMNS" } }
```

Errors without one (unhandled `500`, unknown route): `"data": null`.

## Request (both endpoints)

`Content-Type: multipart/form-data`

| Part | Required | Value |
|---|---|---|
| `file` | yes | the CSV file; at most 5 MB (5 242 880 bytes) |
| `mode` | no | `PARTIAL` (default) or `STRICT` |

## `POST /api/imports/validate`

Checks the file against the current store. Stores nothing, takes no lock.

- `200` → `data` = **ImportResult** with `outcome: "VALIDATED_ONLY"`, `isCommitted: false`,
  `import: null`. Message `"Successfully validate asset import"`.

## `POST /api/imports`

Re-validates the file inside one transaction (serialised by an advisory lock) and applies it.

- `201` → `data` = **ImportResult** with `outcome` one of `COMMITTED`, `NOTHING_TO_COMMIT`,
  `STRICT_BLOCKED` and `import` = the stored import record. Message
  `"Successfully create asset import"`. `isCommitted` is `true` only for `COMMITTED`.
- `500` → transaction rolled back; nothing stored; no import record. The client shows
  *"Not committed — the import failed and was rolled back"*.

## Errors (both endpoints)

| HTTP | `data.errorCode` | When |
|---|---|---|
| 400 | `INVALID_REQUEST` | no `file` part, or `mode` not `PARTIAL`/`STRICT` |
| 413 | `IMPORT_FILE_TOO_LARGE` | body above the server limit (6 MB) — rejected before the handler; Fiber's 413 is mapped to this code |
| 422 | `IMPORT_FILE_TOO_LARGE` | file larger than 5 MB |
| 422 | `IMPORT_FILE_EMPTY` | no data rows (empty file or header only) |
| 422 | `IMPORT_FILE_NOT_CSV` | not UTF-8 text, or a CSV syntax error such as a broken quote; the message names the line |
| 422 | `IMPORT_MISSING_COLUMNS` | header lacks one or more of the eleven columns |
| 422 | `IMPORT_DUPLICATE_COLUMNS` | header repeats a column name |
| 500 | — | unexpected failure |

A file-level error never includes row results.

## ImportResult

```json
{
  "outcome": "COMMITTED",
  "isCommitted": true,
  "mode": "PARTIAL",
  "fileName": "grid_assets.csv",
  "counts": { "total": 222, "imported": 205, "unchanged": 0, "rejected": 17, "warningRows": 2 },
  "rejectedRows": [
    {
      "row": 93,
      "assetId": "SWB-CY-A",
      "errors": [
        { "code": "INVALID_PARENT_TYPE", "message": "A SWITCHBOARD must have a SUBSTATION as parent, but SWB-CY-B is a SWITCHBOARD" },
        { "code": "CYCLE_DETECTED", "message": "SWB-CY-A is part of a parent cycle: SWB-CY-A, SWB-CY-B" }
      ]
    },
    {
      "row": 182,
      "assetId": null,
      "errors": [ { "code": "ASSET_ID_REQUIRED", "message": "asset_id is required" } ]
    }
  ],
  "warnings": [
    { "row": 89, "assetId": "TX-D1", "code": "DUPLICATE_SERIAL_NUMBER", "message": "Serial number SER-TX-DUP is also used by TX-D2" }
  ],
  "import": {
    "id": "6f1c2c1e-6a4e-4b8e-9a52-3d1f2b9c0a11",
    "createdAt": "2026-11-19T10:15:00Z"
  }
}
```

| Field | Type | Notes |
|---|---|---|
| `outcome` | string | `VALIDATED_ONLY` \| `COMMITTED` \| `NOTHING_TO_COMMIT` \| `STRICT_BLOCKED` |
| `isCommitted` | boolean | `true` only when assets were saved |
| `mode` | string | `PARTIAL` \| `STRICT` |
| `fileName` | string | as uploaded |
| `counts.total` | int | data rows; `total = imported + unchanged + rejected` |
| `counts.imported` | int | new valid rows. Saved only when `outcome` is `COMMITTED`; for `VALIDATED_ONLY` and `STRICT_BLOCKED` they are the rows that *would* be imported (the UI labels them "to import"); `NOTHING_TO_COMMIT` ⇒ 0 |
| `counts.unchanged` | int | rows identical to stored assets |
| `counts.rejected` | int | rows with at least one error |
| `counts.warningRows` | int | rows with at least one warning |
| `rejectedRows[]` | array | sorted by `row`; never `null` (empty array) |
| `rejectedRows[].row` | int | file line number, header = 1 |
| `rejectedRows[].assetId` | string \| null | `null` when the row has no asset ID |
| `rejectedRows[].errors[]` | array | ≥ 1; `{ code, message }`; codes listed in [../data-model.md](../data-model.md#validation-rules--codes) |
| `warnings[]` | array | sorted by `row`; never `null` |
| `import` | object \| null | `null` for `VALIDATED_ONLY`; otherwise `{ id, createdAt }` (RFC 3339) |

Message texts above are illustrative; the English catalog in `backend/internal/apperr` is the
source. Codes are stable; messages may be reworded.

## Health

`GET /api/health` → `200` when the database answers (go-common `RegisterHealthCheck`, moved from
`/v1` per exception 5).
