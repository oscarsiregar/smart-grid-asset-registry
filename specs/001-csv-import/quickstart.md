# Quickstart: validating the CSV import end to end

**Feature**: [spec.md](spec.md) · **API**: [contracts/imports-api.md](contracts/imports-api.md) ·
**UI**: [contracts/ui-import-panel.md](contracts/ui-import-panel.md)

## Prerequisites

- Docker with Compose v2.
- For running checks outside containers: Go 1.26, `golangci-lint` v2.13, Node 24 + pnpm 11.
- The sample file at `backend/domain/assetimport/testdata/grid_assets.csv`.

## 1. Start the stack

```sh
docker compose up --build
```

Expected: `db` healthy on `localhost:55432`, `api` logs that migrations applied and the server
listens on `8080`, `web` serves `http://localhost:3000`.
`curl -s localhost:8080/api/health` returns `200`.

## 2. Automated checks (Definition of Done)

```sh
# backend (from backend/), with the compose db running
TEST_DATABASE_URL=postgres://postgres:postgres@localhost:55432/asset_registry_test?sslmode=disable make check
make lint

# frontend (from frontend/)
pnpm lint && pnpm build && pnpm test

# or both in containers
docker compose --profile test run --rm backend-test
docker compose --profile test run --rm frontend-test
```

Expected: exit code 0 everywhere. Without `TEST_DATABASE_URL` the integration tests report as
skipped, not passed.

## 3. API scenarios

Start from an empty database (`docker compose down -v && docker compose up --build`).

Note: the template's rate limiter blocks a client after 20 non-2xx responses in 15 minutes. The
error scenarios below (F–H3) plus repeated UI tries can trigger it; a `429` then means "wait or
restart `api`", not a bug.

| # | Command (from repo root) | Expected |
|---|---|---|
| A | `curl -s -F file=@backend/domain/assetimport/testdata/grid_assets.csv localhost:8080/api/imports/validate` | `200`, `outcome VALIDATED_ONLY`, counts total 222 / imported 205 / unchanged 0 / rejected 17 (golden values), warnings on rows 89 and 90; `asset` table still empty |
| B | same file to `POST /api/imports` with `-F mode=STRICT` | `201`, `STRICT_BLOCKED`, `isCommitted false`; `asset` empty; one `asset_import` row |
| C | same file to `POST /api/imports` (partial) | `201`, `COMMITTED`, imported 205; `asset` has 205 rows, 13 substations |
| D | repeat C | `201`, `NOTHING_TO_COMMIT`, imported 0, unchanged 205, same 17 rejections; still 205 assets |
| E | a CSV with one stored ID and a changed status | that row rejected `ASSET_ID_CONFLICT`, message names `operational_status` |
| F | header-only CSV | `422`, `data.errorCode IMPORT_FILE_EMPTY`, no row results |
| G | CSV missing the `asset_name` column | `422`, `IMPORT_MISSING_COLUMNS` |
| H | a 5.5 MB file, then a 7 MB file | 5.5 MB → `422 IMPORT_FILE_TOO_LARGE`; 7 MB → `413` with `data.errorCode IMPORT_FILE_TOO_LARGE` |
| H2 | a CSV with an unclosed quote on line 40 | `422 IMPORT_FILE_NOT_CSV`, message names line 40 |
| H3 | a CSV whose header has two `asset_id` columns | `422 IMPORT_DUPLICATE_COLUMNS` |
| I | stop `db`, repeat C | `500`; after restarting `db`, no new `asset_import` row and no assets added; api log shows the failure with file name and mode |

Inspect the database: `docker compose exec db psql -U postgres -d asset_registry -c "select outcome, imported_rows, unchanged_rows, rejected_rows from asset_import order by created_at"`.

## 4. UI scenarios (http://localhost:3000)

1. `/` redirects to `/assets`; empty state is shown.
2. **Import CSV** → URL becomes `/assets?panel=import`, focus is inside the panel; Tab stays inside;
   Escape closes it and focus returns to the button; browser Back reopens/closes it consistently.
3. Refresh with the panel open → panel still open, empty form.
4. Choose the sample, **Check file** → "Validated only — nothing saved", counts, 17 rejected rows
   (rows 93–96 show two errors each), 2 warnings.
5. **Save 205 assets** → "Committed — 205 assets saved".
6. **Import another file**, same file again → **Save** shows "Nothing to commit — no new assets".
7. Strict mode with the sample → "Not committed — strict mode, 17 rows rejected".
8. Upload a header-only file → alert with the server message, no table.
9. Stop the `api` container and check a file → error toast plus failed outcome text; restart.
10. Check every outcome with a greyscale filter (DevTools → Rendering → emulate achromatopsia):
    the state is still readable from text.

## 5. Order independence spot check

```sh
(head -1 grid_assets.csv; tail -n +2 grid_assets.csv | shuf) > shuffled.csv
```

Validate `shuffled.csv`: same accepted/rejected asset IDs and codes as scenario A (row numbers
differ). The automated test `TestValidate_OrderIndependent` covers 20 permutations.
