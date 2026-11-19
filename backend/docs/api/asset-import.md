# API: asset import

Imports grid assets from a CSV file. Source of truth for the module's HTTP contract; it matches
`specs/001-csv-import/contracts/imports-api.md`.

- Base path: `/api` (no version prefix, constitution Template Exception 5)
- Authentication: none (Template Exception 1)
- Handler: `internal/handler/asset_import` · Usecase: `internal/usecase/asset_import` ·
  Rules: `domain/assetimport`

## Response envelope

Every response uses the go-common envelope:

```json
{ "code": 200, "status": "success", "message": "Successfully validate asset import", "data": { } }
```

Errors that carry an application error code put it in `data.errorCode` (Template Exception 4):

```json
{ "code": 422, "status": "error", "message": "The file contains no data rows.", "data": { "errorCode": "IMPORT_FILE_EMPTY" } }
```

An unexpected failure is `500` with `"message": "Internal Server Error"` and `"data": null`; the
details are only logged.

## Request

`Content-Type: multipart/form-data`

| Part | Required | Value |
|---|---|---|
| `file` | yes | the CSV file, UTF-8, at most 5 MB |
| `mode` | no | `PARTIAL` (default) or `STRICT` |

The file must have a header naming the eleven columns `asset_id`, `parent_asset_id`, `asset_type`,
`asset_name`, `operational_status`, `voltage_kv`, `rating_kva`, `manufacturer`, `model`,
`serial_number`, `commissioned_date` (any order; extra columns are ignored). The file is read in
memory and never stored (Template Exception 12).

## `POST /api/imports/validate`

Checks the file against the current store. Stores nothing, takes no lock, leaves no import record.

**200** — `data` is an [ImportResult](#importresult) with `outcome` `VALIDATED_ONLY`,
`isCommitted` `false` and `import` `null`.

```sh
curl -s -F file=@backend/domain/assetimport/testdata/grid_assets.csv localhost:8080/api/imports/validate
```

## `POST /api/imports`

Re-validates the file against the store **at this moment** and applies it in one transaction,
serialised with every other import by a PostgreSQL advisory lock: lock → read the stored assets the
file mentions → validate → store the import record → insert the accepted assets (parents first).

**201** for every completed attempt — each one stores an import record (time, file name, mode,
counts, outcome) — with `data` an [ImportResult](#importresult) whose `import` is `{ id, createdAt }`:

| `outcome` | What was stored | `isCommitted` |
|---|---|---|
| `COMMITTED` | the record and every new valid row | `true` |
| `NOTHING_TO_COMMIT` | the record only (no new valid rows) | `false` |
| `STRICT_BLOCKED` | the record only (`mode` `STRICT` and at least one row rejected) | `false` |

**500** — an unexpected failure rolls the transaction back: no assets and no import record are
stored. The failure is logged with the file name and mode.

Rejected-row and warning details are returned but never stored.

```sh
curl -s -F file=@backend/domain/assetimport/testdata/grid_assets.csv -F mode=PARTIAL localhost:8080/api/imports
```

## Errors

| HTTP | `data.errorCode` | When |
|---|---|---|
| 400 | `INVALID_REQUEST` | no `file` part, or `mode` is not `PARTIAL`/`STRICT` |
| 413 | `IMPORT_FILE_TOO_LARGE` | request body above the 6 MB server limit (answered before the handler) |
| 422 | `IMPORT_FILE_TOO_LARGE` | file larger than 5 MB |
| 422 | `IMPORT_FILE_EMPTY` | empty file or header only |
| 422 | `IMPORT_FILE_NOT_CSV` | not UTF-8, or a CSV syntax error such as a broken quote; the message names the line |
| 422 | `IMPORT_MISSING_COLUMNS` | header lacks required columns (named in the message) |
| 422 | `IMPORT_DUPLICATE_COLUMNS` | header repeats a column name |
| 500 | — | unexpected failure (for `POST /api/imports`: rolled back, nothing stored) |

These apply to both endpoints. A file-level error never includes row results and, for
`POST /api/imports`, leaves no import record.

## ImportResult

| Field | Type | Notes |
|---|---|---|
| `outcome` | string | `VALIDATED_ONLY` \| `COMMITTED` \| `NOTHING_TO_COMMIT` \| `STRICT_BLOCKED` |
| `isCommitted` | boolean | `true` only when assets were saved |
| `mode` | string | `PARTIAL` \| `STRICT` |
| `fileName` | string | as uploaded |
| `counts.total` | int | data rows; `total = imported + unchanged + rejected` |
| `counts.imported` | int | new valid rows (saved only when `outcome` is `COMMITTED`) |
| `counts.unchanged` | int | rows identical to stored assets |
| `counts.rejected` | int | rows with at least one error |
| `counts.warningRows` | int | rows with at least one warning |
| `rejectedRows[]` | array | sorted by `row`; never `null` |
| `rejectedRows[].row` | int | file line the record starts on; the header is line 1 |
| `rejectedRows[].assetId` | string \| null | `null` when the row has no asset ID |
| `rejectedRows[].errors[]` | array | `{ code, message }`, at least one |
| `warnings[]` | array | `{ row, assetId, code, message }`, sorted by `row`; never `null` |
| `import` | object \| null | `null` for `VALIDATED_ONLY`; otherwise `{ id, createdAt }` (RFC 3339, UTC) |

### Row error codes

| Code | Meaning |
|---|---|
| `MALFORMED_ROW` | the record has a different number of fields than the header (no other field rule is reported for it) |
| `ASSET_ID_REQUIRED` / `ASSET_TYPE_REQUIRED` / `ASSET_NAME_REQUIRED` | the value is blank |
| `DUPLICATE_ASSET_ID` | the asset ID is on more than one row; every such row is rejected |
| `INVALID_ASSET_TYPE` | not `SUBSTATION`, `TRANSFORMER`, `LV_BOARD`, `SWITCHBOARD` or `SWITCHBOARD_PANEL` (case-sensitive) |
| `INVALID_OPERATIONAL_STATUS` | not blank and not `IN_SERVICE`, `MAINTENANCE` or `OUT_OF_SERVICE` |
| `INVALID_NUMBER` | `voltage_kv` or `rating_kva` is not a plain decimal |
| `NEGATIVE_RATING` | `rating_kva` is below zero |
| `INVALID_DATE` | `commissioned_date` is not a real date as `d/m/yy` (year 20yy) or `yyyy-mm-dd` |
| `SUBSTATION_HAS_PARENT` / `PARENT_REQUIRED` | a substation has a parent / another type has none |
| `SELF_PARENT` | the parent ID is the row's own ID |
| `PARENT_NOT_FOUND` | the parent is neither stored nor any row's asset ID |
| `INVALID_PARENT_TYPE` | the parent's type is not the permitted one (`SUBSTATION` for transformers, LV boards and switchboards; `SWITCHBOARD` for panels) |
| `CYCLE_DETECTED` | the row is on a parent cycle; the message lists the cycle |
| `PARENT_REJECTED` | the parent row in the file is itself rejected |

Hierarchy codes are only reported for rows whose own fields are valid, and a row lists every
hierarchy code that applies. Warning: `DUPLICATE_SERIAL_NUMBER` — the serial number is also used by
another accepted or stored asset; the row is still accepted.
