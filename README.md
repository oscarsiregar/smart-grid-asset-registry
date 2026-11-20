# Smart Grid Asset Registry

Imports a grid-asset CSV file (substations, transformers, LV boards, switchboards and their panels), validates
every row on the server, stores the accepted assets in PostgreSQL, and lets operators browse and search the
stored hierarchy.

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
  grouped by type and the counts of each type beneath it; a switchboard lists its panels. The sidebar shows how
  many substations are stored, and every area shows placeholders shaped like its content while it loads.
- **Link to any asset**: every asset has its own address, `/assets/<asset_id>` (the ID percent-encoded), which
  reopens with its ancestors expanded and the asset highlighted; an unknown ID shows "not found". The tree
  refreshes after an import that saved data.
- **Search by asset ID or name**: any part of either, letter case ignored, optionally limited to one asset type.
  The results replace the tree and show each asset's type, status and location; an exact ID or name comes first.
  Selecting a result opens it in the tree, and Back returns to the search. The search is in the address
  (`/assets?q=<text>&type=<TYPE>`), so it survives a refresh and can be shared.

Stack: Go 1.26 (Fiber, GORM) · PostgreSQL 16 · React 19 + TypeScript (Vite, RTK Query) · Docker Compose.

## Prerequisites

- Docker with Compose v2. That is enough to run the app and the tests.
- Only for running checks outside containers: Go 1.26, golangci-lint v2.13, Node.js 24 and pnpm 11.

No private packages, credentials or `.env` files are needed to run the stack.

## Run

```sh
docker compose up --build
```

| Service | URL / port | What |
|---|---|---|
| `web` | http://localhost:3000 | the app: the asset explorer at `/assets` (an asset at `/assets/<asset_id>`, a search at `/assets?q=<text>`); **Import CSV** opens the import panel |
| `api` | http://localhost:8080/api | the API; `GET /api/health` answers when it is ready |
| `db` | `localhost:5432` | PostgreSQL (`postgres` / `postgres`, database `asset_registry`) |

Migrations run when the API starts. To start from an empty database: `docker compose down -v`, then up again. A
sample file is in `backend/domain/assetimport/testdata/grid_assets.csv` (222 rows: 205 accepted, 17 rejected).

If another local stack already uses one of these ports, override it for one run with `DB_HOST_PORT`,
`API_HOST_PORT` or `WEB_HOST_PORT`.

## Test

In containers:

```sh
docker compose --profile test run --rm backend-test    # go test ./... (unit tests)
docker compose --profile test run --rm frontend-test   # pnpm test
```

Or locally, including the integration tests on real PostgreSQL:

```sh
# backend/ (integration tests are skipped without TEST_DATABASE_URL; they need a database asset_registry_test)
TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/asset_registry_test?sslmode=disable' make check
make lint

# frontend/
pnpm lint && pnpm build && pnpm test
```

End-to-end API and UI scenarios: [specs/001-csv-import/quickstart.md](specs/001-csv-import/quickstart.md) (import),
[specs/002-asset-explorer/quickstart.md](specs/002-asset-explorer/quickstart.md) (explorer) and
[specs/003-asset-search/quickstart.md](specs/003-asset-search/quickstart.md) (search).

## Deploy

Every push to `main` is checked, built and deployed to an Azure VM by GitHub Actions
([.github/workflows/](.github/workflows/)); pull requests run the same checks. Setup, operations and backups:
[deploy/README.md](deploy/README.md).

## Documentation

| Document | Contents |
|---|---|
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | System overview, backend layers, the import engine, transaction and lock, the explorer reads and ordering, the search, schema, frontend, trade-offs |
| [docs/KNOWN_LIMITATIONS.md](docs/KNOWN_LIMITATIONS.md) | What is simplified or out of scope |
| [specs/001-csv-import/](specs/001-csv-import/), [specs/002-asset-explorer/](specs/002-asset-explorer/), [specs/003-asset-search/](specs/003-asset-search/) | Per feature: spec, plan, research decisions, data model, API and UI contracts, tasks |
| [backend/docs/api/asset-import.md](backend/docs/api/asset-import.md) | API reference for the two import endpoints |
| [backend/docs/api/asset.md](backend/docs/api/asset.md) | API reference for the two explorer endpoints and the search |
| [deploy/README.md](deploy/README.md) | Azure deployment and the GitHub Actions pipeline: one-time setup, operations, backups |
| [AGENTS.md](AGENTS.md) | Rules for anyone (human or AI) changing the code, and where the detailed rules live |
| [AI_JOURNAL.md](AI_JOURNAL.md) | How the AI agent was used: tools, workflow, prompts, corrections, review, time |
| [.specify/memory/constitution.md](.specify/memory/constitution.md) | Project principles, template exceptions, Definition of Done |
| [backend/README.md](backend/README.md), [frontend/README.md](frontend/README.md) | Working on each part on its own |

The project was built feature by feature with Spec Kit (specify → clarify → plan → tasks → analyze → implement)
on top of the company backend and frontend templates; the deliberate differences from those templates are the
constitution's Template Exceptions.
