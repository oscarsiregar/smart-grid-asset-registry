# Data Model: CSV Import of Grid Assets

**Feature**: [spec.md](spec.md) · **Decisions**: [research.md](research.md) R2, R3, R7, R8

Database rules (Constitution II): primary keys, unique keys, NOT NULL and foreign keys only — no
CHECK constraints, no triggers. Every business rule below is enforced in the backend.

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
| `updated_at` | `TIMESTAMPTZ` | no | same as `created_at` (records are never updated) |

Indexes: `ix_asset_import_created_at (created_at DESC)`.
Invariant (checked in tests, not the DB): `total = imported + unchanged + rejected`.
No `*_by_name` audit columns (research D2). Failed attempts and validate-only runs leave no row.

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
| `updated_at` | `TIMESTAMPTZ` | no | |

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
  (date only, UTC midnight), `ImportID uuid.UUID`, `CreatedAt, UpdatedAt time.Time`.
  Method `DifferingFields(other *Asset) []string` compares the eleven CSV-backed fields by value
  (numbers numerically, dates by calendar day) and returns CSV column names in CSV order (FR-019).
- **`AssetImport`**: `ID`, `FileName`, `Mode valueobject.ImportMode`, `Outcome valueobject.ImportOutcome`,
  `TotalRows`, `ImportedRows`, `UnchangedRows`, `RejectedRows`, `WarningRows int`, `CreatedAt`, `UpdatedAt`.

### Pure import engine (`domain/assetimport`, research D4/R2)

| Type | Fields / role |
|---|---|
| `Row` | `Number int` (file line, header = 1), `Values map[Column]string` (raw), plus trimmed accessors |
| `ParsedFile` | `Rows []Row`; header resolved by name (order-free, extra columns ignored) |
| `FileError` | `Code string` (`errorcode.Import*`), `Args []any`; returned by `Parse` |
| `StoredAssets` | `ByAssetID map[string]*entity.Asset`; `BySerial map[string][]string` (serial → asset IDs) |
| `Issue` | `Code string`, `Args []any` — rendered to English by the handler via `apperr.Message` |
| `RejectedRow` | `Row int`, `AssetID *string`, `Errors []Issue` |
| `Warning` | `Row int`, `AssetID string`, `Issue` |
| `Report` | `Accepted []*entity.Asset` (parents first, then by asset ID), `UnchangedRows []int`, `Rejected []RejectedRow` (by row), `Warnings []Warning` (by row), `Counts` |
| `Counts` | `Total`, `ToImport`, `Unchanged`, `Rejected`, `WarningRows` |

Entry points: `Parse(r io.Reader, maxBytes int64) (ParsedFile, *FileError)` and
`Validate(file ParsedFile, stored StoredAssets, newID func() uuid.UUID) Report`
(IDs injected so tests are deterministic).

## Validation rules → codes

Row-level codes (row rejected; a row may carry several). Field rules run on every row; hierarchy
rules only on rows whose fields are valid; all applicable hierarchy errors are listed (FR-016a).

| Code | Rule | FR |
|---|---|---|
| `MALFORMED_ROW` | wrong number of fields | edge case |
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

File-level codes (whole file rejected, `422`, no row results): `IMPORT_FILE_TOO_LARGE`,
`IMPORT_FILE_EMPTY` (no data rows), `IMPORT_FILE_NOT_CSV` (not UTF-8 / unparseable),
`IMPORT_MISSING_COLUMNS` (arg: missing column names). Request-shape problems (no `file` part, bad
`mode`) use the existing `INVALID_REQUEST` (`400`).

## Commit state transitions (`CreateAssetImport`)

```text
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
