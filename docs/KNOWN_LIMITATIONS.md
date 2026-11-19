# Known limitations

What this version deliberately does not do, where it simplifies, and what a reviewer may notice. The design
itself is described in [ARCHITECTURE.md](ARCHITECTURE.md).

## Data and rules

- **The brief's "220 valid assets" is not what the sample yields.** The brief calls the sample "a valid dataset
  of 220 assets across 13 substations". The file holds 222 data rows, including deliberately invalid test rows.
  Applying the spec's rules gives 205 accepted, 17 rejected and 2 rows with warnings. The 13 substations match.
  The golden test (`backend/domain/assetimport/testdata/grid_assets.golden.json`) pins the values derived from
  the rules, not the brief's number.
- **No extra sanity rules beyond the brief** (clarification Q2). A commissioned date in the future and a
  negative `voltage_kv` are accepted; only `rating_kva < 0` is rejected. Adding such rules is a possible next
  step.
- **Stored assets are never edited.** A row that reuses a stored asset ID with different values is rejected
  (`ASSET_ID_CONFLICT`). Updating or deleting assets is out of scope.
- **Duplicate serial numbers are only a warning**, also for rows that are unchanged on re-import. Re-importing
  the sample therefore reports the same 2 warning rows again.
- **Numbers are compared as `float64`.** `voltage_kv` and `rating_kva` are stored as unconstrained `NUMERIC`
  (exactly as typed) but read back into `float64` when deciding whether a re-imported row is unchanged. Values
  that differ only beyond about 15 significant digits compare as equal.

## Results and history

- **Detailed results are not stored.** Rejected rows and warnings are shown once, in the response. Refreshing
  the page returns to an empty form, and only the import record's summary (counts and outcome) is kept
  (FR-029, FR-031).
- **No screen lists past imports.** Import records are in the `asset_import` table and can be inspected with
  `psql`. Browsing and searching the asset hierarchy are also not part of this feature; `/assets` shows an empty
  state.
- **Validate-only checks are not recorded.** Failed (rolled-back) attempts are not recorded either; they are only
  logged.

## Size and concurrency

- **5 MB per file (about 50 000 rows).** A larger file is refused with `IMPORT_FILE_TOO_LARGE`; a request body
  above 6 MB is cut off by the server with a `413`. The file is parsed in memory and never stored (Template
  Exception 12).
- **Imports are serialised per database.** A transaction-scoped advisory lock lets one commit run at a time.
  This is simple and safe, but a second import waits for the first; on the 50 000-row file a commit takes about
  9 seconds on a developer machine.

## Platform and template

- **No users.** There is no authentication or authorisation (Template Exception 1). The template's audit
  columns `created_by_name` and `updated_by_name` always hold `"system"`.
- **Local go-common is a subset.** `backend/pkg/gocommon` re-implements only what this service uses (response
  envelope, AppError, logger, database, common middleware, error and health handlers). There is no audit plugin
  and no telemetry.
- **Rate limiter during error testing.** The template's limiter counts only non-2xx responses: 20 in 15 minutes
  per client, then `429`. Repeating the error scenarios (empty file, missing columns, too large…) can trigger
  it. Wait, or restart the `api` container.
- **`backend/values.yml` (Helm) is not maintained.** It is left from the template; the supported way to run the
  app is `docker compose`.
- **Development logs show SQL values.** In `APP_ENV=dev` the GORM logger prints statements with their values.
  `APP_ENV=prod` uses the quiet logger. Row contents are never written by the application's own logs.
