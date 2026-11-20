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
  `psql`.
- **Validate-only checks are not recorded.** Failed (rolled-back) attempts are not recorded either; they are only
  logged.

## Asset explorer

- **Complete lists, no paging or virtual scrolling** (Template Exception 13, 002-R11). A node's children are all
  loaded and rendered at once. The API side stays fast at 50 000 assets (a 5 000-child list in about 30 ms),
  but a node with many thousands of children makes a long page in the browser.
- **The details are not one snapshot** (002-R5). The asset, its path, its children and a substation's counts
  are read separately, without a transaction. Assets are only ever added, so the only possible mismatch is an
  import committing between those reads (counts that already include a new child the list does not yet show);
  the refetch after the commit corrects it.
- **Only the selected asset's ancestors are restored.** The address holds the selected asset; which other nodes
  were expanded is not kept across a refresh or a new tab.
- **Asset IDs `.` and `..` cannot be selected.** Imports accept them, but browsers and `fetch` rewrite those path
  segments, even encoded as `%2E`, so such an asset has no address and its details cannot be requested. It is
  still listed in the tree with its type and status and can be expanded to reach its children; its label is
  plain text rather than a link.
- **Numbers are shown by value** (FR-010). `NUMERIC` values reach the browser as JSON numbers, so trailing zeros
  typed in the CSV are not kept: `22.50` shows as `22.5 kV`.

## Asset search

- **Only the asset ID and name are searched.** Manufacturer, model, serial number and other attributes are not;
  the type is a filter, not a search term.
- **At most 50 results are shown.** A broader search shows the first 50 in order and the total, and asks for more
  of the ID or name or a type to narrow it; there is no paging in the page (clarification of FR-010). The API
  itself accepts `page` and `limit`.
- **Every search loads all its matches before ordering them** (Template Exception 14, 003-R0). The natural ID
  order is applied in Go, so a search matching most of the registry reads and sorts every row: about 770 ms for
  `-` at 55 000 assets, inside the 1 s budget but with little room. Beyond this scale the path is a `pg_trgm`
  index for matching and an ordering expressible in SQL, so the page can be cut by the database.
- **"Exact" can differ from "matches" for exotic characters** (003-R2). Matching uses PostgreSQL's case folding
  (`ILIKE`); the "exact match first" decision uses Go's `strings.EqualFold`. They agree for the IDs and names
  seen in practice, but for some non-ASCII characters an asset could match without being ranked as exact, or the
  other way round.
- **Contains, not fuzzy.** A typo finds nothing; there are no suggestions or partial-word scoring.

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
- **`backend/values.yml` (Helm) and `backend/.gitlab-ci.yml` are not maintained.** They are left from the
  template; the app runs with `docker compose` locally and is deployed by GitHub Actions to an Azure VM
  (Template Exception 15, [deploy/README.md](../deploy/README.md)).

## Deployment

- **Public without sign-in.** The deployed link is open to anyone (Template Exceptions 1 and 15): anyone who has
  it can browse, search and import into the shared database.
- **One VM, no high availability.** One `Standard_B2ats_v2` (1 GB memory) runs every container. A deploy restarts
  the changed containers, so the app is briefly unavailable, and a VM failure stops it until the VM is back.
- **No automatic backups.** The database lives in a Docker volume on the VM disk; backups are taken by hand
  (`pg_dump`, see deploy/README.md).
- **Development logs show SQL values.** In `APP_ENV=dev` the GORM logger prints statements with their values.
  `APP_ENV=prod` uses the quiet logger. Row contents are never written by the application's own logs.
