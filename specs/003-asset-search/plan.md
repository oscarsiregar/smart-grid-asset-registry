# Implementation Plan: Asset Search

**Branch**: `003-asset-search` | **Date**: 2026-11-19 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/003-asset-search/spec.md`

## Summary

Add a search to the asset page: the operator types all or part of an asset ID or name, optionally picks a type,
and gets the matching assets (first 50 plus the total) in place of the tree, each with its location in the
hierarchy. Selecting a result is following a link to the asset's address from feature 002, so the existing
deep-link behaviour expands the ancestors, highlights the asset and shows its details.

Approach: the existing list read `GET /api/assets` gets a search filter (`q`, `assetType`, `page`, `limit`) on the
same usecase (`GetListAsset`), repository method (`FindNodes`) and filter struct (`FindAssetNodesFilter`), plus one
new repository read for the results' ancestors (`FindAncestorIDs`). Matching is `ILIKE` with the template's
escaping helper; ordering reuses 002's natural order with exact matches first, in a new pure function. The
frontend adds a `searchAssets` query, a search form built with `useAppForm`, Zod and the template's `Input`
restored from git history, a results component, and the search in the address (`?q=&type=`). No schema change.
One template rule conflict — paging after an in-memory sort — is flagged for a new Template Exception 14
(research R0).

## Technical Context

**Language/Version**: Go 1.26 (backend); TypeScript 6 + React 19 (frontend)

**Primary Dependencies**: Fiber v2.52, GORM + pgx v5, local go-common (backend); React Router 7.18, Redux Toolkit /
RTK Query, React Hook Form + Zod, CSS Modules (frontend). No new dependency.

**Storage**: PostgreSQL 16 — the existing `asset` table and indexes; read-only; no migration

**Testing**: testify + mockery + sqlmock, real-PostgreSQL integration tests with `TEST_DATABASE_URL` (backend);
Vitest + React Testing Library with `fetch` mocked (frontend)

**Target Platform**: Docker Compose stack (`db`, `api`, `web`); evergreen desktop browsers

**Project Type**: web application (Go API + React SPA)

**Performance Goals**: with 50 000 stored assets, search results within 1 s on the local stack (SC-005)

**Constraints**: results in the explorer's order with exact matches first (FR-008); first 50 plus the total, no
paging UI (FR-010); search state in the URL (principle V, FR-006); opening a result reuses 002's deep link
(FR-011); no frontend business rules (principle II)

**Scale/Scope**: up to ~50 000 assets; 1 endpoint extended, 1 repository read added; 1 page extended

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle / rule | How the plan complies | Status |
|---|---|---|
| I. Follow the Template | Extends the `asset` module of 002 in place (one more filter on `GetListAsset` / `FindNodes` / `FindAssetNodesFilter`, one repository method in its own file, `Validate()` + `SetDefault()` on the input, `MaxLimit` 100, `{ items, totalItems }`, `querysafety.ContainsPattern` for `ILIKE`); frontend form via `useAppForm` + Zod with the template's `Input` restored from git history, a query with `providesTags`, the search in `SearchParameter`. **Deviation D1**: the page is cut in the usecase after the natural-order sort instead of SQL `LIMIT/OFFSET` (`04-usecase.md`, `05-repository-database.md`) | **Needs amendment** (D1 → proposed Exception 14) |
| II. Validation lives in the backend | Matching, ordering, the 50-result page, parameter rules and the total are the backend's; the Zod schema only checks the text is not blank (like feature 001's "choose a file"); no schema change, no CHECK/trigger | Pass |
| III. Predictable, explainable imports | Not changed; a committed import refreshes a shown search through the existing `ASSET` tag | Pass |
| IV. Test the rules | `SortSearchMatches` / `CheckIsExactMatch` and `GetListAssetInput.Validate/SetDefault` table-tested first; repository sqlmock tests; usecase and handler tests; integration tests on real PostgreSQL with the sample (SC-001, SC-002 against an SQL oracle) and the scale test (SC-005) | Pass |
| V. Usable and accessible | Search text and type in the URL; running / no-match / too-many / failed states in words; `role="search"` form with labelled controls; results are links (keyboard and new-tab work); the import panel keeps its behaviour | Pass |
| Exceptions 1, 4, 5, 6, 9, 13 | Public `/api`, envelope with `data.errorCode`, English messages, English docs; Exception 13 untouched — hierarchy reads stay unpaged, the search is a paged list as that exception requires | Pass |
| Definition of Done | `make check`, `make lint`, `pnpm lint/build/test`, stack tried by hand; `docs/ARCHITECTURE.md`, `docs/KNOWN_LIMITATIONS.md`, API docs and READMEs updated | Planned |

**Gate result**: one violation, not worked around: **D1 search paging in the usecase**. It needs Template
Exception 14 (constitution 1.2.0 → 1.3.0) before `/speckit-implement`; wording proposed in
[research.md](research.md) R0. Everything else passes.

**Post-design re-check** (after Phase 1): data-model, contracts and quickstart add no further deviation — no
migration, no new dependency, no new package; `SortSearchMatches` lives in `domain/entity` beside 002's ordering,
`FindAncestorIDs` in the existing `psql/asset` repository, and the restored `Input` is the template's own file.
Status unchanged: pass, pending Exception 14.

## Project Structure

### Documentation (this feature)

```text
specs/003-asset-search/
├── spec.md
├── plan.md              # this file
├── research.md          # R0–R8 (R0: template fit, proposed Exception 14)
├── data-model.md        # no schema change; filter, repository, usecase, DTO and frontend additions
├── quickstart.md        # API scenarios A–J, UI scenarios 1–11, scale check
├── contracts/
│   ├── assets-search-api.md
│   └── ui-asset-search.md
├── checklists/requirements.md
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root)

```text
backend/
├── domain/
│   ├── entity/asset_search.go           # SortSearchMatches, CheckIsExactMatch (+ test)
│   ├── repository/asset.go              # + FindAncestorIDs
│   ├── repository/find_asset_nodes_filter.go   # + Search, AssetType
│   ├── usecase/asset.go                 # GetListAssetInput + Search, AssetType, Page, Limit, Validate, SetDefault;
│   │                                    #   output + TotalItems, AncestorIDs; MaxLimit, SearchDefaultLimit
│   └── dto/asset.go                     # request q/assetType/page/limit; AssetSearchItemResponse, AssetSearchResponse
├── internal/
│   ├── repository/psql/asset/           # find_nodes.go (+ search WHERE), find_ancestor_ids.go (+ sqlmock tests)
│   ├── usecase/asset/get_list_asset.go  # hierarchy or search branch (+ tests); helpers.go (page cut)
│   ├── handler/asset/                   # get_list_asset.go (parse, type, mode), helpers.go (search response) (+ tests)
│   ├── mocks/                           # make mocks
│   └── integrationtest/                 # asset_search_test.go (sample vs SQL oracle), scale test extended
└── docs/api/asset.md                    # search section

frontend/src/
├── libs/components/input/               # input.tsx, styles.module.css — restored from 40d511b^
├── libs/enums/input-type.enum.ts        # restored; + export in enums.ts
├── libs/enums/search-parameter.enum.ts  # + QUERY "q", ASSET_TYPE "type"
├── modules/assets/                      # assets-api.ts + searchAssets; libs/types (search DTOs);
│                                        #   libs/validation-schemas/asset-search.validation-schema.ts; assets.ts barrel
└── pages/assets/
    ├── assets.tsx                       # passes the search to the explorer
    ├── components/asset-explorer/       # results in place of the tree while a search is shown
    ├── components/asset-search/         # search form (Input + type field) and the results list with its states
    └── libs/{enums,hooks,constants,helpers}/  # copy; use-asset-search (URL); result location text

docs/ARCHITECTURE.md, docs/KNOWN_LIMITATIONS.md, READMEs, .specify/memory/constitution.md (Exception 14)
```

**Structure Decision**: The existing `backend/` and `frontend/` from the templates. No new backend module: the
search extends the `asset` module's list read. The frontend gains one page component folder `asset-search` and
reuses the explorer, the tree and the details of feature 002 unchanged; the shared `Input` returns to
`libs/components` as the template had it.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| D1: search paging cut in the usecase after an in-memory sort (template pages in SQL) | FR-008 requires the explorer's natural asset order with exact matches first; that order is a Go function (002-R3) and cannot be expressed in SQL, so the page can only be cut after sorting every match. Requires Template Exception 14 | SQL `ORDER BY asset_id` gives `SS-10` before `SS-2`; an ICU numeric collation needs a migration, depends on the image's ICU build and orders differently from the tree; sorting in the browser moves a business rule to the frontend and sends every match |
