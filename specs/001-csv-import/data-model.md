# Data Model: CSV Import of Grid Assets

**Feature**: [spec.md](spec.md) · **Decisions**: [research.md](research.md) R2, R3, R7, R8

Database rules (Constitution II, exception 8): primary keys, unique keys, NOT NULL and foreign
keys only — no CHECK constraints, no triggers. Every business rule below is enforced in the backend.
Audit columns follow the template (`created_at`, `created_by_name`, `updated_at`,
`updated_by_name`), with `"system"` as the actor name.

## Tables

### `asset_import` — one row per completed confirm attempt (FR-029)

| Column | Type | Null | Notes |
|---|---|---|---|
| `id` | `UUID` | no | PK, generated in the usecase |
| `file_name` | `TEXT` | no | original upload name |
| `mode` | `VARCHAR(16)` | no | `ImportMode`: `PARTIAL`, `STRICT` |
| `outcome` | `VARCHAR(32)` | no | `ImportOutcome`: `COMMITTED`, `NOTHING_TO_COMMIT`, `STRICT_BLOCKED` |
| `total_rows` | `INTEGER` | no | data rows in the file |
| `imported_rows` | `INTEGER` | no | new valid rows; saved only when `outcome = COMMITTED` (for `STRICT_BLOCKED` they were not saved), so the count invariant always holds |
| `unchanged_rows` | `INTEGER` | no | |
| `rejected_rows` | `INTEGER` | no | |
| `warning_rows` | `INTEGER` | no | rows carrying at least one warning |
| `created_at` | `TIMESTAMPTZ` | no | `u.now()` |
| `created_by_name` | `VARCHAR(255)` | no | always `"system"` (`usecase.SystemActorName`); template audit column |
| `updated_at` | `TIMESTAMPTZ` | no | same as `created_at` (records are never updated) |
| `updated_by_name` | `VARCHAR(255)` | no | always `"system"` |

Indexes: `ix_asset_import_created_at (created_at DESC)`.
Invariant (checked in tests, not the DB): `total = imported + unchanged + rejected`.
Failed attempts and validate-only runs leave no row.

### `asset` — stored grid assets

| Column | Type | Null | Notes |
|---|---|---|---|
| `id` | `UUID` | no | PK, generated in the usecase |
| `asset_id` | `TEXT` | no | natural key from the CSV; unique |
| `parent_asset_id` | `TEXT` | yes | FK → `asset(asset_id)` `ON DELETE RESTRICT`; `NULL` only for substations |
| `asset_type` | `VARCHAR(32)` | no | `AssetType` |
| `asset_name` | `TEXT` | no | inner whitespace kept |
| `operational_status` | `VARCHAR(32)` | yes | `OperationalStatus`; `NULL` = "no status" |
| `voltage_kv` | `NUMERIC` | yes | unconstrained precision, value kept as typed |
| `rating_kva` | `NUMERIC` | yes | ≥ 0 (backend rule) |
| `manufacturer` | `TEXT` | yes | |
| `model` | `TEXT` | yes | |
| `serial_number` | `TEXT` | yes | not unique (duplicates only warn, FR-020) |
| `commissioned_date` | `DATE` | yes | |
| `import_id` | `UUID` | no | FK → `asset_import(id)`; the import that created the asset |
| `created_at` | `TIMESTAMPTZ` | no | |
| `created_by_name` | `VARCHAR(255)` | no | always `"system"` |
| `updated_at` | `TIMESTAMPTZ` | no | |
| `updated_by_name` | `VARCHAR(255)` | no | always `"system"` |

Audit columns follow the template (rule 05) with no defaults in the database: the usecase sets
them, as the template's `CreateProduct` does with `ActorName`. They are not CSV columns, so they
take no part in the re-import comparison (FR-018).

Indexes: `ux_asset_asset_id (asset_id)` unique; `ix_asset_parent_asset_id (parent_asset_id)`
(children lookups, later browse feature); `ix_asset_serial_number (serial_number)` (FR-020 lookup);
`ix_asset_import_id (import_id)`.

Blank CSV values are stored as `NULL` for every nullable column, so "blank" and "absent" compare
equal during re-import (FR-018).

### Migrations

| File | Creates |
|---|---|
| `backend/migrations/000001_create_asset_import.up.sql` / `.down.sql` | `asset_import` + index |
| `backend/migrations/000002_create_asset.up.sql` / `.down.sql` | `asset` + FKs + indexes |

The template's `000001_create_product.*` is deleted first (never deployed; research R8).

## Domain types (`backend/domain`)

### Value objects (`domain/valueobject`, one file each, template enum shape: `Parse*`, `Scan`, `Value`, `String`, `IsValid`)

| Type | Values | Notes |
|---|---|---|
| `AssetType` | `SUBSTATION`, `TRANSFORMER`, `LV_BOARD`, `SWITCHBOARD`, `SWITCHBOARD_PANEL` | plus `PermittedParent() (AssetType, bool)` — `false` for `SUBSTATION` |
| `OperationalStatus` | `IN_SERVICE`, `MAINTENANCE`, `OUT_OF_SERVICE` | entity field is a pointer (nullable) |
| `ImportMode` | `PARTIAL`, `STRICT` | default `PARTIAL` |
| `ImportOutcome` | `VALIDATED_ONLY`, `COMMITTED`, `NOTHING_TO_COMMIT`, `STRICT_BLOCKED` | `VALIDATED_ONLY` is never stored |

Permitted-parent table (FR-014): `TRANSFORMER`, `LV_BOARD`, `SWITCHBOARD` → `SUBSTATION`;
`SWITCHBOARD_PANEL` → `SWITCHBOARD`; `SUBSTATION` → none.

### Entities (`domain/entity`)

- **`Asset`**: `ID uuid.UUID`, `AssetID string`, `ParentAssetID *string`, `AssetType valueobject.AssetType`,
  `AssetName string`, `OperationalStatus *valueobject.OperationalStatus`, `VoltageKV *float64`,
  `RatingKVA *float64`, `Manufacturer, Model, SerialNumber *string`, `CommissionedDate *time.Time`
  (date only, UTC midnight), `ImportID uuid.UUID`, `CreatedAt time.Time`, `CreatedByName string`,
  `UpdatedAt time.Time`, `UpdatedByName string`.
  Method `DifferingFields(other *Asset) []string` compares the eleven CSV-backed fields by value
  (numbers numerically, dates by calendar day) and returns CSV column names in CSV order (FR-019).
- **`AssetImport`**: `ID`, `FileName`, `Mode valueobject.ImportMode`, `Outcome valueobject.ImportOutcome`,
  `TotalRows`, `ImportedRows`, `UnchangedRows`, `RejectedRows`, `WarningRows int`, `CreatedAt`,
  `CreatedByName`, `UpdatedAt`, `UpdatedByName`.
- **`usecase.SystemActorName = "system"`** (`domain/usecase/asset_import.go`): the only value
  written to `*_by_name` columns while the product has no users (exception 1).

### Pure import engine (`domain/assetimport`, exception 10, research R2)

| Type | Fields / role |
|---|---|
| `Row` | `Number int` (file line, header = 1), `Values map[Column]string` (raw), plus trimmed accessors |
| `ParsedFile` | `Rows []Row`; header resolved by name (order-free, extra columns ignored) |
| `FileError` | `Code string` (`errorcode.Import*`), `Args []any`; returned by `Parse` |
| `StoredAssets` | `ByAssetID map[string]*entity.Asset`; `BySerial map[string][]string` (serial → asset IDs) |
| `Issue` | `Code string`, `Args []any` — rendered to English by the handler via `apperr.Message` |
| `RejectedRow` | `Row int`, `AssetID *string`, `Errors []Issue` |
| `Warning` | `Row int`, `AssetID string`, `Issue` |
| `Report` | `Accepted []*entity.Asset` (parents first, then by asset ID; **`ID`, `ImportID` and audit fields left zero**), `UnchangedRows []int`, `Rejected []RejectedRow` (by row), `Warnings []Warning` (by row), `Counts` |
| `Counts` | `Total`, `ToImport`, `Unchanged`, `Rejected`, `WarningRows` |

Entry points: `Parse(r io.Reader, maxBytes int64) (ParsedFile, *FileError)` and
`Validate(file ParsedFile, stored StoredAssets) Report`. The package imports only the standard
library and other `domain` packages (exception 10) — no `github.com/google/uuid`. The usecase
assigns `ID` (`uuid.New` via its injectable `newID`), `ImportID` and the audit fields to each
accepted asset before inserting.

`Parse` is the single place the 5 MB limit is enforced (it reads at most `maxBytes + 1` bytes);
the usecase only translates its `FileError`.

One count, three names: `Counts.ToImport` (engine) = `counts.imported` (API) = `imported_rows`
(table) — new valid rows, saved only when the outcome is `COMMITTED`.

## Validation rules → codes

Row-level codes (row rejected; a row may carry several). Field rules run on every row; hierarchy
rules only on rows whose fields are valid; all applicable hierarchy errors are listed (FR-016a).

| Code | Rule | FR |
|---|---|---|
| `MALFORMED_ROW` | wrong number of fields on a record with valid CSV syntax; reported alone (values are misaligned, so no other rule is applied) | edge case |
| `ASSET_ID_REQUIRED` / `ASSET_TYPE_REQUIRED` / `ASSET_NAME_REQUIRED` | blank after trim | FR-005 |
| `DUPLICATE_ASSET_ID` | ID appears on >1 row; every such row | FR-006 |
| `INVALID_ASSET_TYPE` | not one of five types | FR-007 |
| `INVALID_OPERATIONAL_STATUS` | non-blank and not one of three | FR-008 |
| `INVALID_NUMBER` | `voltage_kv` / `rating_kva` not a decimal (arg: column) | FR-009 |
| `NEGATIVE_RATING` | `rating_kva < 0` | FR-009 |
| `INVALID_DATE` | not a real date in `d/m/yy` or `yyyy-mm-dd` (`yy` → 20yy) | FR-010 |
| `SUBSTATION_HAS_PARENT` / `PARENT_REQUIRED` | parent presence vs type | FR-011 |
| `SELF_PARENT` | parent ID = own ID | FR-012 |
| `PARENT_NOT_FOUND` | parent neither stored nor any row's ID | FR-013 |
| `INVALID_PARENT_TYPE` | parent (row or stored) has a supported type ≠ permitted parent | FR-014 |
| `CYCLE_DETECTED` | row is on a parent-pointer cycle | FR-015 |
| `PARENT_REJECTED` | in-file parent is rejected (fixpoint; not added to cycle members) | FR-016 |
| `ASSET_ID_CONFLICT` | ID stored with different values (arg: differing columns) | FR-019 |

Outcome per accepted row: **to import** (ID not stored) or **unchanged** (ID stored, `DifferingFields`
empty, FR-018).

Warning: `DUPLICATE_SERIAL_NUMBER` — an accepted row whose serial is shared with another accepted
row or a stored asset with a different ID (arg: the other asset IDs, sorted) (FR-020).

File-level codes (whole file rejected, no row results):

| Code | HTTP | Rule |
|---|---|---|
| `IMPORT_FILE_TOO_LARGE` | 422 | file larger than 5 MB (`Parse`) |
| `IMPORT_FILE_TOO_LARGE` | 413 | request body above the 6 MB server limit (Fiber's 413 mapped in the go-common error handler, `pkg/gocommon/pkg/utils`) |
| `IMPORT_FILE_EMPTY` | 422 | no data rows |
| `IMPORT_FILE_NOT_CSV` | 422 | not UTF-8, or **any CSV syntax error** (unclosed/stray quote); arg: the line number |
| `IMPORT_MISSING_COLUMNS` | 422 | header lacks required columns; arg: missing names in CSV order |
| `IMPORT_DUPLICATE_COLUMNS` | 422 | header repeats a column name (case-sensitive, after trim); arg: repeated names |

Request-shape problems (no `file` part, bad `mode`) use the existing `INVALID_REQUEST` (`400`).

## Commit state transitions (`CreateAssetImport`)

```text
body > 6 MB ──► 413 IMPORT_FILE_TOO_LARGE (before the handler)
parse ──file error──► 422, nothing stored, no record
  │
  ▼  (transaction + advisory lock)
validate against stored
  ├─ STRICT and rejected > 0 ───────────► record STRICT_BLOCKED, no assets     ─► commit ─► 201
  ├─ to-import = 0 ─────────────────────► record NOTHING_TO_COMMIT, no assets  ─► commit ─► 201
  └─ otherwise ─────────────────────────► record COMMITTED + insert accepted   ─► commit ─► 201
any technical error ─► rollback, log error, no record, no assets ─► 500
```

`ValidateAssetImport` runs the same parse + validate without transaction or lock and returns
outcome `VALIDATED_ONLY` (`200`).
