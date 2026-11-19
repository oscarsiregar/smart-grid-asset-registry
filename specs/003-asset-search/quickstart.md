# Quickstart: validating the asset search end to end

**Feature**: [spec.md](spec.md) · **API**: [contracts/assets-search-api.md](contracts/assets-search-api.md) ·
**UI**: [contracts/ui-asset-search.md](contracts/ui-asset-search.md)

## Prerequisites

As in features 001 and 002: Docker with Compose v2; the stack running (`docker compose up --build`) with the sample
imported (`curl -s -F file=@backend/domain/assetimport/testdata/grid_assets.csv localhost:8080/api/imports`).

The oracle for "every match, none extra" (SC-002) is a plain SQL count of the stored rows:

```sh
docker compose exec db psql -U postgres -d asset_registry -tAc \
  "select count(*) from asset where asset_id ilike '%<text>%' or asset_name ilike '%<text>%'"
```

## 1. Automated checks (Definition of Done)

```sh
docker compose --profile test run --rm backend-test     # make check incl. integration tests
docker compose --profile test run --rm frontend-test    # pnpm test
# plus, locally: make lint (backend/), pnpm lint && pnpm build (frontend/)
```

Expected: exit code 0 everywhere.

## 2. API scenarios

| # | Request | Expected |
|---|---|---|
| A | `GET /api/assets?q=PNL-001-1-01` | `200`; `items[0].assetId` is `PNL-001-1-01` with `ancestorAssetIds` `["SS-001","SWB-001-1"]` |
| B | `GET /api/assets?q=aurora%20panel` | `200`; `totalItems` equals the SQL count for `aurora panel`; every item's ID or name contains it, letter case ignored |
| C | `GET /api/assets?q=swb-001` | `200`; lower-case text matches `SWB-001-…` IDs; `totalItems` equals the SQL count |
| D | `GET /api/assets?q=001&assetType=SWITCHBOARD_PANEL` | `200`; every item is a `SWITCHBOARD_PANEL`; `totalItems` equals the SQL count with `and asset_type = 'SWITCHBOARD_PANEL'` |
| E | `GET /api/assets?q=-` | `200`; `totalItems` = number of stored assets, `items` has 50, substations first |
| F | `GET /api/assets?q=-&page=2&limit=50` | `200`; the next 50, no overlap with E |
| G | `GET /api/assets?q=NO-SUCH-ASSET` | `200`, `items: []`, `totalItems: 0` |
| H | `GET /api/assets?q=%25` and `?q=_` | `200`; only assets whose ID or name contains a literal `%` / `_` (none in the sample) |
| I | `GET /api/assets?assetType=TRANSFORMER`; `?q=x&parentAssetId=SS-001`; `?page=2`; `?q=x&assetType=FOO`; a 101-character `q` | each `400 INVALID_REQUEST` with the message from the contract |
| J | `GET /api/assets` and `?parentAssetId=SS-001` | unchanged from feature 002 (complete lists, no `ancestorAssetIds`) |

Note: the rate limiter counts non-2xx responses (20 per 15 minutes); scenario I adds five.

## 3. UI scenarios (http://localhost:3000/assets)

1. Search `PNL-001-1-01`: the results replace the tree; the panel is first, with "Switchboard panel",
   "Maintenance" and "SS-001 › SWB-001-1". The address is `/assets?q=PNL-001-1-01`.
2. Select it: the address becomes `/assets/PNL-001-1-01` (no `q`); the tree is back with `SS-001` and `SWB-001-1`
   expanded, the panel highlighted and in view, its details shown — the same as opening the address directly.
3. Press Back: the search and its results are shown again; Forward returns to the panel.
4. Search `aurora panel`, then choose "Switchboard panel" and search again: only panels are listed; refresh — the
   same text, type and results.
5. Search `-`: "Showing 50 of N assets that match "-"…" with the request to narrow the search.
6. Search `NO-SUCH-ASSET`: "No asset matches "NO-SUCH-ASSET"" with the hint; no toast.
7. Submit a blank search: "Enter an asset ID or name to search."; the address does not change. Submit a
   101-character text: the results area shows the server's "Invalid request: q must be at most 100 characters".
8. **Clear**: the tree is back, the address has no `q`.
9. With a search shown, **Import CSV**: the address keeps `q` and adds `panel=import`; close it — the same search.
   Import a file with a new matching asset: the search refreshes and lists it.
10. Stop the `api` container and search: "The search could not be completed." + **Try again**; restart `api`, retry.
11. Keyboard only: Tab to the search field, type, Enter; Tab through the results and Enter on one.
12. By hand (SC-003, SC-006): from `/assets`, reach the highlighted details of a known ID through the search in
    under 15 seconds; with a greyscale filter, every search state is readable from its text alone.

## 4. Scale check (SC-005)

With the 50 000-row file of feature 001 imported (or let the integration test measure the API side): a search
matching every asset, one matching a few, and one with a type each return within 1 s on the local compose stack.
