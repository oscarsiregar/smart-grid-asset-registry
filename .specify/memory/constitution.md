<!--
Sync Impact Report
- Version change: 1.0.0 → 1.1.0 (MINOR: Template Exceptions materially expanded)
- Modified principles: none
- Added: Template Exceptions 8–12 (no CHECK constraints, English project docs,
  domain/assetimport package, real-PostgreSQL test packages, in-memory CSV uploads)
- Removed sections: none
- Templates: none modified (dependent templates read this file at runtime)
- Follow-up: specs/001-csv-import plan deviation D2 (omit *_by_name audit columns) is NOT
  covered by this amendment; the plan must keep the template audit columns or a further
  amendment must add it.
-->

# Smart Grid Asset Registry Constitution

A web app that imports a CSV of grid assets (substations and their equipment, linked by
parent asset ID), validates it on the server, stores accepted rows in PostgreSQL, and lets
users browse and search the asset hierarchy.

## Core Principles

### I. Follow the Template

`backend/` and `frontend/` come from the team template. `backend/AGENTS.md`,
`backend/docs/rules/*`, `frontend/AGENTS.md` and `frontend/docs/*` are binding. New code
MUST copy the shape of existing modules rather than invent new patterns. Any deviation
MUST be listed under Template Exceptions before it is made.

### II. Validation Lives in the Backend

All business rules MUST be enforced by the backend. The database uses only primary keys,
unique keys, NOT NULL and foreign keys; CHECK constraints and triggers are forbidden.
The frontend MUST NOT re-implement validation.

### III. Predictable, Explainable Imports

CSV row order MUST NOT change the import result. Every rejected row MUST report its CSV
row number, its asset ID when present, a stable error code and a readable message. Every
import result MUST state clearly whether data was committed.

### IV. Test the Rules

Every import and hierarchy rule MUST have a table-driven unit test written before the
rule. The supplied sample CSV MUST be covered by a golden test. Repositories and usecases
MUST also be tested against a real PostgreSQL.

### V. Usable and Accessible

The selected asset and open panels MUST live in the URL so refresh, deep links and the
back button work. Loading, empty and error states MUST always be handled. Status MUST NOT
be conveyed by colour alone; panels MUST trap focus and close on Escape.

## Template Exceptions

1. **No authentication**: the product has no users, so routes are public and there are no
   permission checks.
2. **Frontend tests**: Vitest + React Testing Library are added (the template has no test
   runner).
3. **Local stack**: a root `docker-compose.yml` runs `db` (PostgreSQL 16, port 55432),
   `api` (port 8080) and `web` (nginx, port 3000, forwarding `/api`); database migrations
   run automatically on start.
4. **Local shared library**: the template's private go-common library is re-implemented
   inside `backend/` with the same names and signatures, so the project builds without
   company credentials. Its only addition: error responses with an application error code
   include it as `data.errorCode`.
5. **No API version prefix**: routes live under `/api` (for example `/api/imports`).
6. **English messages**: the error message catalog holds English text only.
7. **Template example modules** are removed before delivery.
8. **No CHECK constraints**: enum and range columns are plain columns validated by the
   backend, overriding the template's CHECK pattern (follows from Principle II).
9. **English documentation**: project documentation (root `docs/`, `backend/docs/api/*`,
   READMEs) is written in English; the template's existing Indonesian rule files stay as
   they are.
10. **Pure import rules package**: the import rules live in their own package
    `backend/domain/assetimport/`, importing only the standard library and other `domain`
    packages.
11. **Integration test packages**: real-PostgreSQL integration tests live in test-only
    packages (`backend/internal/integrationtest/` and `backend/internal/testsupport/pgtest/`)
    and are skipped when `TEST_DATABASE_URL` is not set.
12. **In-memory uploads**: uploaded CSV files are parsed in memory and never stored, so the
    template's file-storage rules do not apply.

## Definition of Done

- Backend `make check` and `make lint` pass.
- Frontend `pnpm lint`, `pnpm build` and `pnpm test` pass.
- `docker compose up --build` starts cleanly and the change was tried through the real app.
- `docs/ARCHITECTURE.md` and `docs/KNOWN_LIMITATIONS.md` are updated when relevant.
- No secrets are committed.

## Governance

This constitution overrides all other practices. Changes are made only by amending this
file with a semantic version bump (MAJOR: principle removed or redefined; MINOR: principle
or section added or materially expanded; PATCH: wording clarifications). Plans and reviews
MUST check compliance with these principles and the Definition of Done.

**Version**: 1.1.0 | **Ratified**: 2026-11-19 | **Last Amended**: 2026-11-19
