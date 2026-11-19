# Smart Grid Asset Registry — frontend

React + TypeScript single-page app for importing grid-asset CSV files: choose a file and a mode, see the server's
validation result (counts, rejected rows, warnings), then save it. Built on the company frontend template (store,
API layer, error handling, notifications, forms, routing) with strict conventions.

- Rules: [docs/code-conventions.md](docs/code-conventions.md)
- Recipe for a new feature: [docs/adding-a-feature.md](docs/adding-a-feature.md)
- Review rubric: [docs/quality-criteria.md](docs/quality-criteria.md)
- Architecture: [../docs/ARCHITECTURE.md](../docs/ARCHITECTURE.md)

The reference feature is the CSV import: `src/modules/imports` (endpoints, DTOs, schema) and `src/pages/assets`
(the `/assets` page and its import panel, opened with `?panel=import`). There is no sign-in: the app has no users
(constitution, Template Exception 1).

## Prerequisites

- Node.js 24 (`.nvmrc`)
- pnpm 11 (`corepack enable`)
- The backend on `http://localhost:8080` for the dev server, for example `docker compose up -d db api` from the
  repository root

No private packages and no credentials are needed.

## Run

```
cp .env.example .env
pnpm install
pnpm start:dev
```

The dev server runs on `VITE_APP_DEVELOPMENT_PORT` (3000) and proxies `VITE_APP_API_ORIGIN_URL` (`/api`) to
`VITE_APP_PROXY_SERVER_URL` (`http://localhost:8080`), so the browser talks to one origin and needs no CORS setup.
Stop the compose `web` service first if it is running, since it also uses port 3000.

To install the git hooks (pre-commit lint, commit message check), run `pnpm ci:prepare` once.

## Test

```
pnpm lint
pnpm build
pnpm test
```

Tests use Vitest + React Testing Library with `fetch` mocked (Template Exception 2) and sit next to the code they
cover. The same tests run in a container from the repository root:
`docker compose --profile test run --rm frontend-test`.

## Scripts

| Script           | Purpose                                                                       |
| ---------------- | ----------------------------------------------------------------------------- |
| `pnpm start:dev` | Vite dev server                                                               |
| `pnpm build`     | Type check + production build into `build/`                                   |
| `pnpm preview`   | Serve the production build                                                    |
| `pnpm test`      | Vitest, once                                                                  |
| `pnpm lint`      | Every check (types, ESLint, Stylelint, Prettier, ls-lint, editorconfig, knip) |
| `pnpm format`    | Apply Prettier                                                                |

## Environment

| Variable                    | Used by        | Meaning                                 |
| --------------------------- | -------------- | --------------------------------------- |
| `VITE_APP_NODE_ENV`         | app            | `local`, `development`, or `production` |
| `VITE_APP_API_ORIGIN_URL`   | app, dev proxy | Base path of the API: `/api`            |
| `VITE_APP_DEVELOPMENT_PORT` | dev server     | Port of the Vite dev server             |
| `VITE_APP_PROXY_SERVER_URL` | dev server     | Backend the dev proxy forwards to       |

Everything prefixed `VITE_` ships to the browser; never put a secret there.

## Backend contract

| Method | Path                    | Body (multipart)                    | Response                             |
| ------ | ----------------------- | ----------------------------------- | ------------------------------------ |
| POST   | `/api/imports/validate` | `file`, `mode` (`PARTIAL`/`STRICT`) | `200`, import result, nothing stored |
| POST   | `/api/imports`          | `file`, `mode`                      | `201`, import result with the record |

Responses use the go-common envelope `{ code, status, message, data }`; `baseQuery` unwraps `data`. Errors carry
their application code in `data.errorCode` (for example `IMPORT_MISSING_COLUMNS`). The full contract is in
[../specs/001-csv-import/contracts/imports-api.md](../specs/001-csv-import/contracts/imports-api.md).

## Docker

The root `docker-compose.yml` builds this folder as the `web` service: the image builds the app with
`VITE_APP_API_ORIGIN_URL=/api` and serves `build/` with nginx on port 80 (host port 3000), proxying `/api` to the
`api` service (`nginx/nginx.conf`).
