# Smart Grid Asset Registry

Imports a grid-asset CSV file (substations, transformers, LV boards, switchboards and their panels), validates
every row on the server, stores the accepted assets in PostgreSQL, and lets operators browse the stored
hierarchy.

- **Check a file** without saving anything: every rejected row is listed with every reason (missing or invalid
  fields, unknown or wrongly typed parent, cycles, conflicts with stored assets), plus non-blocking warnings
  such as duplicate serial numbers.
- **Save it** in *partial* mode (store the valid rows) or *strict* mode (store nothing if any row is rejected).
  The file is re-validated at save time inside one transaction, and every result says in words whether data was
  saved.
- **Re-import safely**: rows identical to stored assets count as unchanged, and rows that reuse a stored ID with
  different values are rejected, so importing the same file twice adds nothing.
- **Browse the hierarchy**: an expandable tree (substations at the top, children loaded when a node is opened)
  beside a details area with every attribute and the path from the substation. A substation shows its children
  grouped by type and the counts of each type beneath it; a switchboard lists its panels.
- **Link to any asset**: every asset has its own address, `/assets/<asset_id>` (the ID percent-encoded), which
  reopens with its ancestors expanded and the asset highlighted; an unknown ID shows "not found". The tree
  refreshes after an import that saved data.

Stack: Go 1.26 (Fiber, GORM) · PostgreSQL 16 · React 19 + TypeScript (Vite, RTK Query) · Docker Compose.

## Prerequisites

- Docker with Compose v2. That is enough to run the app and every check.
- Only for running checks outside containers: Go 1.26, golangci-lint v2.13, Node.js 24 and pnpm 11.

No private packages, credentials or `.env` files are needed to run the stack.

## Run

```sh
docker compose up --build
```

| Service | URL / port | What |
|---|---|---|
| `web` | http://localhost:3000 | the app: the asset explorer at `/assets` (an asset at `/assets/<asset_id>`); **Import CSV** opens the import panel |
| `api` | http://localhost:8080/api | the API; `GET /api/health` answers when it is ready |
| `db` | `localhost:55432` | PostgreSQL (`postgres` / `postgres`, database `asset_registry`) |

Migrations run when the API starts. To start from an empty database: `docker compose down -v`, then up again. A
sample file is in `backend/domain/assetimport/testdata/grid_assets.csv` (222 rows: 205 accepted, 17 rejected).

If another local stack already uses one of these ports, override it for one run with `DB_HOST_PORT`,
`API_HOST_PORT` or `WEB_HOST_PORT`.

## Test

In containers, with the stack's database running:

```sh
docker compose --profile test run --rm backend-test    # make check, incl. integration tests on real PostgreSQL
docker compose --profile test run --rm frontend-test   # pnpm test
```

Or locally:

```sh
# backend/ (integration tests are skipped without TEST_DATABASE_URL)
TEST_DATABASE_URL='postgres://postgres:postgres@localhost:55432/asset_registry_test?sslmode=disable' make check
make lint

# frontend/
pnpm lint && pnpm build && pnpm test
```

End-to-end API and UI scenarios: [specs/001-csv-import/quickstart.md](specs/001-csv-import/quickstart.md) (import)
and [specs/002-asset-explorer/quickstart.md](specs/002-asset-explorer/quickstart.md) (explorer).

## Documentation

| Document | Contents |
|---|---|
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | System overview, backend layers, the import engine, transaction and lock, the explorer reads and ordering, schema, frontend, trade-offs |
| [docs/KNOWN_LIMITATIONS.md](docs/KNOWN_LIMITATIONS.md) | What is simplified or out of scope |
| [specs/001-csv-import/](specs/001-csv-import/), [specs/002-asset-explorer/](specs/002-asset-explorer/) | Per feature: spec, plan, research decisions, data model, API and UI contracts, tasks |
| [backend/docs/api/asset-import.md](backend/docs/api/asset-import.md) | API reference for the two import endpoints |
| [backend/docs/api/asset.md](backend/docs/api/asset.md) | API reference for the two explorer endpoints |
| [AGENTS.md](AGENTS.md) | Rules for anyone (human or AI) changing the code, and where the detailed rules live |
| [.specify/memory/constitution.md](.specify/memory/constitution.md) | Project principles, template exceptions, Definition of Done |
| [backend/README.md](backend/README.md), [frontend/README.md](frontend/README.md) | Working on each part on its own |

The project was built feature by feature with Spec Kit (specify → clarify → plan → tasks → analyze → implement)
on top of the company backend and frontend templates; the deliberate differences from those templates are the
constitution's Template Exceptions.
