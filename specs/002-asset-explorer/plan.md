# Implementation Plan: Asset Explorer

**Branch**: `002-asset-explorer` | **Date**: 2026-11-19 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/002-asset-explorer/spec.md`

## Summary

Turn the `/assets` page into an explorer of the stored hierarchy: an expandable tree (substations at the top,
children loaded only when a node is expanded) next to a details area; every asset at its own address
`/assets/<asset_id>`, which reopens with its ancestors expanded and the asset highlighted. A substation shows
its children grouped by type and counts of each type beneath it; a switchboard lists its panels; an unknown ID
shows a "not found" state.

Approach: a read-only backend module `asset` (two endpoints, `GET /api/assets[?parentAssetId=]` and
`GET /api/assets/:assetId`) built like `asset_import`, with new read methods on the existing asset repository
(recursive CTEs for the path and the descendant counts), ordering done by a pure, tested function in
`domain/entity`; and a frontend `modules/assets` (RTK Query queries, tag invalidated by a committed import)
plus a hand-written accessible tree and details components in `pages/assets`. No schema change. One template
rule conflict — complete, unpaged lists — is flagged for a new Template Exception (research R0).

## Technical Context

**Language/Version**: Go 1.26 (backend); TypeScript 6 + React 19 (frontend)

**Primary Dependencies**: Fiber v2.52, GORM + pgx v5, local go-common (backend); React Router 7.18, Redux
Toolkit / RTK Query, CSS Modules (frontend). No new dependency.

**Storage**: PostgreSQL 16 — the existing `asset` table and indexes from feature 001; read-only; no migration

**Testing**: testify + mockery + sqlmock, real-PostgreSQL integration tests with `TEST_DATABASE_URL`
(backend); Vitest + React Testing Library with `fetch` mocked (frontend)

**Target Platform**: Docker Compose stack (`db`, `api`, `web`); evergreen desktop browsers

**Project Type**: web application (Go API + React SPA)

**Performance Goals**: with 50 000 stored assets, top level within 2 s, expanding a node or selecting an asset
within 1 s (SC-006)

**Constraints**: complete lists, no paging (clarification Q1); every asset ID except `.` and `..` round-trips
through the address (FR-017); state in the URL (principle V); no frontend business rules (principle II)

**Scale/Scope**: up to ~50 000 assets (5 MB import limit); hierarchy at most 3 levels; 2 endpoints, 1 page

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle / rule | How the plan complies | Status |
|---|---|---|
| I. Follow the Template | New `asset` module copies `asset_import` (handler → usecase → repository, one file per endpoint / use case / method, `init.go` + `helpers.go`, `GetListXxx` / `GetXxxDetail`, `{ items, totalItems }`, `response.Success`, errors via `apperr`); frontend `modules/assets` + `pages/assets` follow `modules/imports` + `pages/assets` and the conventions (§8 tags, §6 components, §12 routes). **Exception: complete lists without paging** conflicts with the template's list paging (`04-usecase.md`, `05-repository-database.md`) | **Needs amendment** (D1 → proposed Exception 13) |
| II. Validation lives in the backend | Ordering, grouping order, counts and "not found" are decided by the backend; the frontend displays lists in the order received; no schema change, so no CHECK/trigger question | Pass |
| III. Predictable, explainable imports | Not changed; the explorer only reads; the tree refreshes only after a committed import | Pass |
| IV. Test the rules | Ordering (`CompareAssetIDs`, `SortAssetNodes`, `AssetType.Rank`) table tests first; usecases and repositories also tested against real PostgreSQL with the sample (SC-001, SC-002) | Pass |
| V. Usable and accessible | Selected asset in the path, panel in `?panel=`; loading / empty / error / not-found states per area; status and selection in words and markers, never colour alone; WAI-ARIA tree with full keyboard support; the import panel keeps trapping focus and closing on Escape | Pass |
| Exceptions 1, 4, 5, 6, 9 | Public `/api` routes, go-common envelope with `data.errorCode`, English message for `ASSET_NOT_FOUND`, English API doc | Pass |
| Definition of Done | `make check`, `make lint`, `pnpm lint/build/test`, compose stack tried by hand; `docs/ARCHITECTURE.md` and `docs/KNOWN_LIMITATIONS.md` updated (R5, R11) | Planned |

**Gate result**: one violation, not worked around: **D1 complete hierarchy lists**. It needs Template Exception
13 (constitution 1.1.0 → 1.2.0) before `/speckit-implement`; wording proposed in [research.md](research.md) R0.
Everything else passes.

**Post-design re-check** (after Phase 1): data-model, contracts and quickstart introduce no further deviation —
no migration, no new dependency, no new package outside the template layout (`entity.AssetNode` and
`SortAssetNodes` live in `domain/entity`, the CTE queries in the existing `psql/asset` repository). Status
unchanged: pass, pending Exception 13.

## Project Structure

### Documentation (this feature)

```text
specs/002-asset-explorer/
├── spec.md
├── plan.md              # this file
├── research.md          # R0–R11 (R0: template fit, proposed Exception 13)
├── data-model.md        # no schema change; domain, repository, usecase and DTO additions
├── quickstart.md        # API scenarios A–J, UI scenarios 1–12, scale check
├── contracts/
│   ├── assets-api.md
│   └── ui-asset-explorer.md
├── checklists/requirements.md
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root)

```text
backend/
├── domain/
│   ├── entity/asset_node.go            # AssetNode; SortAssetNodes, CompareAssetIDs (+ tests)
│   ├── valueobject/asset_type.go       # + Rank() (+ test)
│   ├── errorcode/errorcode.go          # + ASSET_NOT_FOUND
│   ├── repository/asset.go             # + FindByAssetID, FindNodes, FindPath, CountDescendantsByType
│   ├── usecase/asset.go                # AssetUsecase: GetListAsset, GetAssetDetail; inputs, outputs, DescendantCounts
│   └── dto/asset.go                    # request (params/query) and response DTOs
├── internal/
│   ├── apperr/apperr.go                # + English message for ASSET_NOT_FOUND
│   ├── repository/psql/
│   │   ├── asset/                      # + find_by_asset_id.go, find_nodes.go, find_path.go,
│   │   │                               #   count_descendants_by_type.go (+ sqlmock tests)
│   │   ├── model/asset_node.go         # projection with child_count
│   │   └── mapper/asset.go             # + ToEntityAssetNodes
│   ├── usecase/asset/                  # init.go, helpers.go, get_list_asset.go, get_asset_detail.go (+ tests)
│   ├── handler/asset/                  # init.go, helpers.go, get_list_asset.go, get_asset_detail.go (+ tests)
│   ├── mocks/                          # make mocks (AssetRepository, AssetUsecase)
│   ├── bootstrap/setup.go              # wire AssetUsecase + AssetHandler
│   ├── routes/router.go                # GET /api/assets, GET /api/assets/:assetId
│   └── integrationtest/                # asset_explorer_test.go (sample, SC-001/SC-002, encoding, scale)
└── docs/api/asset.md

frontend/src/
├── libs/enums/app-route.enum.ts        # + ASSETS_$ASSET_ID
├── libs/enums/api-path.enum.ts         # + ASSETS
├── index.tsx                           # + /assets/:assetId route (same lazy page)
├── modules/assets/                     # assets-api.ts, assets.ts, libs/{enums,types}
├── modules/imports/imports-api.ts      # createImport invalidates ASSET when committed
└── pages/assets/
    ├── assets.tsx                      # reads :assetId; explorer or empty state; import panel
    ├── components/
    │   ├── asset-explorer/             # two-column layout
    │   ├── asset-tree/                 # tree, tree-node, tree-children (lazy query), status label
    │   ├── asset-details/              # path, attributes, child groups, descendant counts, child list,
    │   │                               #   select prompt, not-found, error
    │   ├── empty-registry/
    │   └── import-panel/               # unchanged (feature 001)
    └── libs/{enums,helpers,hooks,constants,maps}/   # labels, getAssetPath, use-tree-keyboard, use-expanded-nodes

docs/ARCHITECTURE.md, docs/KNOWN_LIMITATIONS.md      # explorer sections
```

**Structure Decision**: Web application with the existing `backend/` and `frontend/` from the templates. The
backend gets one new module `asset` beside `asset_import`, reusing the `psql/asset` repository; the frontend
gets one new data module `modules/assets` and new page components under the existing `pages/assets`.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| D1: hierarchy lists without `page`/`limit` (template paging rule) | Clarification Q1: every child is shown at once; the tree needs the whole top level and the whole child list of a node | Frontend looping over pages works around the rule and multiplies requests; a huge default limit hides the same behaviour behind a magic number; "show more" contradicts Q1. Requires Template Exception 13 |
