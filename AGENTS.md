# Agent guidelines — Smart Grid Asset Registry

Instructions for any AI agent (or human) working in this repository. Read this first; then read the
guide for the part you are changing.

## What this is

A full-stack app that imports a grid-asset CSV, validates it on the server, stores accepted rows in
PostgreSQL, and lets users explore and search the asset hierarchy.

## Where the rules live

| Scope | File | Notes |
|---|---|---|
| Whole project (highest authority) | [.specify/memory/constitution.md](.specify/memory/constitution.md) | Principles, template exceptions, definition of done |
| Backend (Go) | [backend/AGENTS.md](backend/AGENTS.md) + [backend/docs/rules/](backend/docs/rules/) | Reference module: `asset_import` (handler, usecase) with repositories `asset` and `asset_import`; pure rules in `domain/assetimport` |
| Frontend (React) | [frontend/AGENTS.md](frontend/AGENTS.md) + [frontend/docs/code-conventions.md](frontend/docs/code-conventions.md) | Reference module: `modules/imports` + `pages/assets` |
| Features | `specs/<NNN-feature>/` (spec, plan, tasks) | Created through Spec Kit |

Precedence: constitution → the part-specific `AGENTS.md` → the feature's spec and plan. Each
`CLAUDE.md` only imports the `AGENTS.md` next to it; edit the `AGENTS.md`, never the `CLAUDE.md`.

## Repository layout

```
backend/    Go API (Fiber, GORM, PostgreSQL), layered like the template
frontend/   React + TypeScript SPA (Vite, RTK Query), layered like the template
specs/      one folder per feature: spec.md, plan.md, tasks.md
.specify/   Spec Kit: constitution, templates, scripts
```

## Template exceptions (summary)

1. No authentication: routes are public, no permission checks.
2. The frontend has automated tests (Vitest + React Testing Library).
3. A root `docker-compose.yml` runs `db` (55432), `api` (8080) and `web` (3000); migrations run on start.
4. The private `go-common` library is re-implemented in `backend/` with the same names and signatures;
   its only addition is `data.errorCode` on error responses that carry an application error code.
5. No API version prefix: routes live under `/api` (for example `/api/imports`).
6. The error message catalog holds English text only.
7. Template example modules are removed before delivery.
8. No CHECK constraints: enum and range columns are plain columns validated by the backend.
9. Project documentation (root `docs/`, `backend/docs/api/*`, READMEs) is in English; the
   template's Indonesian rule files stay as they are.
10. Pure import rules live in `backend/domain/assetimport/` (standard library and `domain/*` only).
11. Real-PostgreSQL integration tests live in `backend/internal/integrationtest/` and
    `backend/internal/testsupport/pgtest/`, skipped without `TEST_DATABASE_URL`.
12. Uploaded CSV files are parsed in memory and never stored.
13. The asset hierarchy reads (top-level substations, the children of one asset) return every row,
    without `page`/`limit`; other list endpoints keep the template's paging.
14. The asset search loads every match, orders it with the domain's natural asset order (exact
    matches first) and pages in the usecase; the API keeps `page`/`limit`/`totalItems`. Other paged
    lists page in SQL.

The constitution holds the full text and the reasons.

## Working approach

- Work feature by feature through Spec Kit: specify → clarify → plan → tasks → analyze → implement.
- Change the spec or plan first when behaviour or design changes; never patch only the code.
- Follow the closest existing example before inventing a pattern.

## Verification

- Backend: `make check`, `make lint` (from `backend/`).
- Frontend: `pnpm lint`, `pnpm build`, `pnpm test` (from `frontend/`).
- Report the exact commands and their exit codes; a non-zero exit is a failure.
