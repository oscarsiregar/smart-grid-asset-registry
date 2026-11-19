---

description: "Task list for the asset explorer"
---

# Tasks: Asset Explorer

**Input**: Design documents from `specs/002-asset-explorer/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md),
[contracts/assets-api.md](contracts/assets-api.md), [contracts/ui-asset-explorer.md](contracts/ui-asset-explorer.md),
[quickstart.md](quickstart.md); constitution v1.2.0 (Template Exception 13 covers the unpaged lists).

**Tests**: Requested (constitution principle IV, user input "tests before code"). Every test task is written
first and must FAIL before the task that makes it pass.

**Organization**: Tasks are grouped by user story. Reference modules to copy: backend
`internal/{handler,usecase}/asset_import`, `internal/repository/psql/asset`, `domain/assetimport`; frontend
`modules/imports`, `pages/assets/components/import-panel`. One file per endpoint / use case / repository method /
component / hook / helper / type / enum, each with its test next to it (`backend/AGENTS.md`,
`frontend/docs/code-conventions.md` §3).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on an unfinished task)
- **[Story]**: US1–US4 from [spec.md](spec.md)

## Path Conventions

Web app: `backend/` (Go module `smart-grid-asset-registry`) and `frontend/src/`. Backend checks: `make check`,
`make lint` from `backend/` (or `docker compose --profile test run --rm backend-test`); integration tests need
`TEST_DATABASE_URL=postgres://postgres:postgres@localhost:55432/asset_registry_test?sslmode=disable`. Frontend
checks: `pnpm lint`, `pnpm build`, `pnpm test` from `frontend/`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm the starting point; no new project, dependency or migration is needed (plan: no schema
change, no new dependency).

- [ ] T001 Baseline on branch `002-asset-explorer`: `docker compose up --build -d --wait`, then `docker compose --profile test run --rm backend-test`, `make lint` (golangci-lint container), `pnpm lint && pnpm build && pnpm test` in `frontend/` all exit 0; record the commands and exit codes as the baseline in the implementation report

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The shared ordering rule, error code, contracts, repository reads, DTOs, routes and frontend data
module that every story uses.

**⚠️ CRITICAL**: No user-story work can begin until this phase is complete.

### Tests first (write, confirm they FAIL)

- [ ] T002 [P] Table test for `AssetType.Rank()` in `backend/domain/valueobject/asset_type_test.go`: SUBSTATION 0, TRANSFORMER 1, LV_BOARD 2, SWITCHBOARD 3, SWITCHBOARD_PANEL 4 (research R3)
- [ ] T003 [P] Table tests for `CompareAssetIDs(a, b string) int` and `SortAssetNodes([]*AssetNode)` in `backend/domain/entity/asset_node_test.go` (research R3, FR-004): `SS-2` < `SS-10`; `SS-001` < `SS-012` < `SS-T01` (a digit run comes before a non-digit run); leading zeros ignored then shorter first (`SS-01` vs `SS-1` deterministic, never 0 for different strings); non-digit runs by Unicode code point; equal IDs → 0; `SortAssetNodes` orders by type rank, then ID (transformers, LV boards, switchboards, panels); stable and total for shuffled input
- [ ] T004 [P] Extend `backend/internal/apperr/apperr_test.go`: `Message(errorcode.AssetNotFound, "SS-9")` is `"No asset with ID SS-9 exists."` and `NotFound(errorcode.AssetNotFound, "SS-9")` maps to 404 with `ErrorCode ASSET_NOT_FOUND` (the catalog completeness test then also covers the new code)
- [ ] T005 [P] Mapper test in `backend/internal/repository/psql/mapper/asset_node_test.go` for `ToEntityAssetNodes`: every field copied, `operational_status` NULL → `nil` (= "no status recorded"), `child_count` → `ChildCount`, nil input → empty non-nil slice
- [ ] T006 [P] sqlmock test in `backend/internal/repository/psql/asset/find_by_asset_id_test.go` (shape of `find_by_asset_ids_test.go`): found → entity; `gorm.ErrRecordNotFound` → `nil, nil`; DB error returned as-is; trx given; wrong Tx type
- [ ] T007 [P] sqlmock test in `backend/internal/repository/psql/asset/find_nodes_test.go`: `ParentAssetID == nil` queries `parent_asset_id IS NULL`, otherwise `parent_asset_id = ?` with the ID as a placeholder argument; the `child_count` subquery is selected; rows mapped; empty result → empty non-nil slice; DB error; trx given; wrong Tx type
- [ ] T008 [P] sqlmock test in `backend/internal/repository/psql/asset/find_path_test.go`: recursive CTE upward from the asset with the ID as a `?` argument; rows returned substation first, the asset last; unknown ID → empty non-nil slice; DB error; trx given; wrong Tx type
- [ ] T009 [P] Frontend test `frontend/src/pages/assets/libs/helpers/get-asset-path.helper.test.ts`: `getAssetPath("SS-001")` → `/assets/SS-001`; `"SS/9 A"` → `/assets/SS%2F9%20A`; `"50%"` → `/assets/50%25`; `"x?y#z"` → `/assets/x%3Fy%23z` (research R2)
- [ ] T010 [P] Frontend test `frontend/src/modules/assets/assets-api.test.tsx` (shape of `modules/imports/imports-api.test.tsx`): `getAssetNodes(null)` requests `GET /api/assets`; `getAssetNodes("SS/1")` requests `/api/assets?parentAssetId=SS%2F1`; `getAssetDetail("SS/1")` requests `/api/assets/SS%2F1`; the envelope's `data` is returned; an error response produces no toast (`shouldSuppressToast`) and keeps `data.errorCode`

### Implementation

- [ ] T011 [P] Add `Rank() int` to `backend/domain/valueobject/asset_type.go` — T002 passes
- [ ] T012 [P] Create `backend/domain/entity/asset_node.go`: `AssetNode{AssetID string; AssetName string; AssetType valueobject.AssetType; OperationalStatus *valueobject.OperationalStatus /* nil = no status recorded */; ChildCount int /* > 0 makes the node expandable (FR-003) */}`, `CompareAssetIDs`, `SortAssetNodes` (in place: type rank, natural ID, plain string) — T003 passes
- [ ] T013 [P] Add `AssetNotFound = "ASSET_NOT_FOUND"` to `backend/domain/errorcode/errorcode.go` (+ `All`) and the message `"No asset with ID %s exists."` to the catalog in `backend/internal/apperr/apperr.go` — T004 passes
- [ ] T014 Add to `backend/domain/repository/asset.go`: `FindByAssetID(ctx, trx, assetID string) (*entity.Asset, error)` (nil, nil when not found), `FindNodes(ctx, trx, filter FindAssetNodesFilter) ([]*entity.AssetNode, error)` (unsorted, non-nil), `FindPath(ctx, trx, assetID string) ([]*entity.AssetNode, error)` (substation first, the asset last, empty when not found), and `FindAssetNodesFilter{ ParentAssetID *string }` (nil = top level) in `backend/domain/repository/find_asset_nodes_filter.go`
- [ ] T015 Create `backend/domain/usecase/asset.go`: `AssetUsecase` with `GetListAsset(ctx, GetListAssetInput) (GetListAssetOutput, error)` and `GetAssetDetail(ctx, GetAssetDetailInput) (GetAssetDetailOutput, error)`; `GetListAssetInput{ParentAssetID *string}`; `GetListAssetOutput{Assets []*entity.AssetNode}`; `GetAssetDetailInput{AssetID string}` with `Validate()` returning an error when `AssetID` is blank after trimming; `GetAssetDetailOutput{Asset *entity.Asset; Path []*entity.AssetNode; Children []*entity.AssetNode}` (US2 adds `DescendantCounts`); input `Validate` tests in `backend/domain/usecase/asset_test.go`
- [ ] T016 Run `make mocks` (mockery v2.53 in the `golang:1.26.9` container) to regenerate `backend/internal/mocks/repository/AssetRepository.go` and create `backend/internal/mocks/usecase/AssetUsecase.go`; do not edit generated files
- [ ] T017 [P] Create the projection `backend/internal/repository/psql/model/asset_node.go` (columns `asset_id, asset_name, asset_type, operational_status, child_count`; not a table) and `ToEntityAssetNodes` in `backend/internal/repository/psql/mapper/asset_node.go` — T005 passes
- [ ] T018 [P] Implement `backend/internal/repository/psql/asset/find_by_asset_id.go` with `gormtx.Conn` and `WHERE asset_id = ?` — T006 passes
- [ ] T019 [P] Implement `backend/internal/repository/psql/asset/find_nodes.go`: `SELECT a.asset_id, a.asset_name, a.asset_type, a.operational_status, (SELECT count(*) FROM asset c WHERE c.parent_asset_id = a.asset_id) AS child_count FROM asset a WHERE …` (research R4) — T007 passes
- [ ] T020 [P] Implement `backend/internal/repository/psql/asset/find_path.go`: `db.Raw` recursive CTE from the asset up through `parent_asset_id`, with depth, each step's `child_count`, ordered substation first; `?` placeholders only — T008 passes
- [ ] T021 [P] Create `backend/domain/dto/asset.go`: `GetListAssetRequest{ParentAssetID string \`query:"parentAssetId"\`}`, `GetAssetDetailRequest{AssetID string \`params:"assetId"\`}`, `AssetNodeResponse{assetId, assetName, assetType, operationalStatus *string, childCount}`, `AssetListResponse{Items []AssetNodeResponse; TotalItems int}` (`items` never null), `AssetResponse` (all eleven attributes, camelCase, blanks `null`, `commissionedDate` `YYYY-MM-DD`), `AssetDetailResponse{Asset, Path, Children}` per [contracts/assets-api.md](contracts/assets-api.md)
- [ ] T022 Create the module skeletons `backend/internal/usecase/asset/init.go` (`assetUsecase` with `assetRepository`, `log`; `NewAssetUsecase`) and `backend/internal/handler/asset/init.go` (`AssetHandler`, `NewAssetHandler`), wire them in `backend/internal/bootstrap/setup.go` (reuse the asset repository) and add `AssetHandler` to `RouteDependencies`; methods return `errors.New("not implemented yet")` until US1; `go build ./...` and `make check` pass
- [ ] T023 [P] Frontend enums and route: `ASSETS: "/assets"` in `frontend/src/libs/enums/api-path.enum.ts`; `ASSETS_$ASSET_ID: "/assets/:assetId"` in `frontend/src/libs/enums/app-route.enum.ts`; a second lazy route entry for `AppRoute.ASSETS_$ASSET_ID` rendering the same `Assets` page in `frontend/src/index.tsx`
- [ ] T024 [P] Frontend data module `frontend/src/modules/assets/`: enums `libs/enums/{asset-type.enum.ts, operational-status.enum.ts, assets-api-tag.enum.ts, enums.ts}` (`AssetsApiTag.ASSET`); types `libs/types/{asset-node-dto.type.ts, asset-list-dto.type.ts, asset-dto.type.ts, asset-detail-dto.type.ts, types.ts}` matching [contracts/assets-api.md](contracts/assets-api.md) (`operationalStatus: null` = no status; numbers `number | null`; `commissionedDate: string | null`)
- [ ] T025 Implement `frontend/src/modules/assets/assets-api.ts` (`baseApi.enhanceEndpoints({ addTagTypes: [AssetsApiTag.ASSET] }).injectEndpoints`) with queries `getAssetNodes(parentAssetId: string | null)` and `getAssetDetail(assetId: string)`, both `providesTags: [AssetsApiTag.ASSET]` and `extraOptions: { shouldSuppressToast: true }`, the ID URL-encoded with `encodeURIComponent`; public barrel `frontend/src/modules/assets/assets.ts` exporting the hooks, DTO types, `AssetType`, `OperationalStatus`, `AssetsApiTag` — T010 passes
- [ ] T026 [P] Implement `frontend/src/pages/assets/libs/helpers/get-asset-path.helper.ts` with `generatePath(AppRoute.ASSETS_$ASSET_ID, { assetId })`, exported from `libs/helpers/helpers.ts` — T009 passes
- [ ] T027 [P] Page copy and labels: extend `frontend/src/pages/assets/libs/enums/assets-page-message.enum.ts` with every text in [contracts/ui-asset-explorer.md](contracts/ui-asset-explorer.md) (loading, empty, error, select prompt, not found, "Not recorded", "No status recorded", "Back to all assets", "Try again", group and count labels, units kV / kVA); add `frontend/src/pages/assets/libs/maps.ts` with `assetTypeToLabel` and `operationalStatusToLabel` (`satisfies Record<…>`)

**Checkpoint**: `make check`, `make lint`, `pnpm lint`, `pnpm build`, `pnpm test` exit 0; the stack still serves the import page.

---

## Phase 3: User Story 1 - Explore the hierarchy and see an asset's details (Priority: P1) 🎯 MVP

**Goal**: The `/assets` page shows the substations as a tree; expanding a node loads its children; selecting an
asset shows all its attributes and its path.

**Independent Test**: Import the sample, open `/assets`, expand `SS-001` and `SWB-001-1`, select `PNL-001-1-01`,
and check every attribute against the CSV row (quickstart UI 3, 5; API B, C, E).

### Tests for User Story 1 (write first, confirm they FAIL) ⚠️

- [ ] T028 [P] [US1] Usecase tests in `backend/internal/usecase/asset/get_list_asset_test.go` (mocks per case, shape of `asset_import` tests): top level (`ParentAssetID nil`) → `FindNodes` with nil filter, result sorted by `SortAssetNodes`; children of a stored asset → sorted; children list empty and the parent exists (`FindByAssetID` returns it) → empty list; children list empty and the parent does not exist → `apperr.NotFound(errorcode.AssetNotFound, id)`; empty registry → empty list, no error; repository error returned as-is and logged
- [ ] T029 [P] [US1] Usecase tests in `backend/internal/usecase/asset/get_asset_detail_test.go`: blank `AssetID` → `400 INVALID_REQUEST` and no repository call; unknown → `404 ASSET_NOT_FOUND`; found panel → `Asset`, `Path` from `FindPath` (substation first, asset last), `Children` sorted; found switchboard → its panels as `Children`; each repository error returned as-is; success logs the asset ID only
- [ ] T030 [P] [US1] Handler tests in `backend/internal/handler/asset/get_list_asset_test.go` (`app.Test` + `utils.NewErrorHandler`, shape of `asset_import` handler tests): no query → usecase called with nil parent; `?parentAssetId=SS-001` and `?parentAssetId=SS%2F1` → decoded ID; empty `parentAssetId` → nil (top level); 200 envelope `"Successfully get list asset"` with `items`/`totalItems`, `items` `[]` not `null`; status `null` when absent; usecase 404 passes through with `data.errorCode`
- [ ] T031 [P] [US1] Handler tests in `backend/internal/handler/asset/get_asset_detail_test.go`: 200 envelope `"Successfully get asset detail"` with `asset`, `path`, `children`; blanks as `null`; numbers as JSON numbers; `commissionedDate` `"2010-09-11"`; usecase 404 → 404 with `ASSET_NOT_FOUND`
- [ ] T032 [P] [US1] Integration test `backend/internal/integrationtest/asset_explorer_test.go` (real PostgreSQL, `pgtest.Open`): import the sample through the feature 001 usecase, then the top level has 13 substations ordered `SS-001` … `SS-012`, `SS-T01`; `?parentAssetId=SS-001` has 5 children ordered transformers, LV board, switchboards; walking `GetListAsset` from the top reaches all 205 assets exactly once (SC-001); the detail of `PNL-001-1-01` matches its CSV row attribute by attribute (rating `nil`, date 2010-09-11) and its path is `SS-001 → SWB-001-1 → PNL-001-1-01`
- [ ] T033 [P] [US1] Frontend tests `frontend/src/pages/assets/components/asset-tree/asset-tree.test.tsx` (`fetch` mocked, real store): renders the top level from `GET /api/assets`; only nodes with `childCount > 0` have `aria-expanded`; children are requested only when a node is expanded and come from the cache on re-expand; a node shows ID, name, type label, and a status label only when not in service (FR-002); a failed child load shows "Could not load children" + **Try again** that refetches
- [ ] T034 [P] [US1] Frontend test `frontend/src/pages/assets/libs/hooks/use-tree-keyboard/use-tree-keyboard.hook.test.tsx`: ↑/↓ move between visible nodes, → expands or moves to the first child, ← collapses or moves to the parent, Home/End, Enter follows the node's link; roving `tabIndex` (exactly one node has `tabIndex=0`) (FR-007)
- [ ] T035 [P] [US1] Frontend tests `frontend/src/pages/assets/components/asset-details/asset-details.test.tsx`: no selection → "Select an asset in the tree to see its details."; loading → status text; loaded panel → path links, heading, all attributes, blanks "Not recorded", status null "No status recorded", `22 kV`, `1000 kVA`, date `2010-09-11`, parent link to `/assets/SWB-001-1` (FR-008–FR-011, FR-016); error (500) → "The asset could not be loaded." + **Try again**
- [ ] T036 [P] [US1] Page tests `frontend/src/pages/assets/assets.test.tsx`: empty registry → "No assets have been imported yet." + **Import CSV** that opens the panel, no tree (FR-023); top-level failure → "The assets could not be loaded." + **Try again** while **Import CSV** still works (FR-024); with assets → tree and details side by side; clicking a node changes the address to `/assets/<id>` (FR-018)

### Implementation for User Story 1

- [ ] T037 [US1] Implement `backend/internal/usecase/asset/get_list_asset.go` (log setup → reads → `SortAssetNodes` → log success with count) — T028 passes
- [ ] T038 [US1] Implement `backend/internal/usecase/asset/get_asset_detail.go` (validate → `FindByAssetID` → nil ⇒ `apperr.NotFound(errorcode.AssetNotFound, id)` → `FindPath` → `FindNodes(parent = id)` → sort) — T029 passes
- [ ] T039 [US1] Implement `backend/internal/handler/asset/helpers.go` (`toAssetNodeResponse`, `toAssetListResponse`, `toAssetResponse`, `toAssetDetailResponse`) and `backend/internal/handler/asset/get_list_asset.go` (`QueryParser`, empty ⇒ nil, 200) — T030 passes
- [ ] T040 [US1] Implement `backend/internal/handler/asset/get_asset_detail.go` (`ParamsParser`, `ctx.UserContext()`, 200) — T031 passes except the encoding cases added in US3
- [ ] T041 [US1] Register `assetGroup := api.Group("/assets")`, `assetGroup.Get("/", …GetListAsset)`, `assetGroup.Get("/:assetId", …GetAssetDetail)` in `backend/internal/routes/router.go` — T032 passes against `TEST_DATABASE_URL`
- [ ] T042 [P] [US1] Implement `frontend/src/pages/assets/libs/hooks/use-tree-keyboard/use-tree-keyboard.hook.ts` using `KeyboardKey` from `libs/enums` — T034 passes
- [ ] T043 [P] [US1] Implement the tree in `frontend/src/pages/assets/components/asset-tree/`: `asset-tree.tsx` (`role="tree"`, top-level query, loading/error rows), `components/tree-node/tree-node.tsx` (`role="treeitem"`, `aria-level`, `aria-expanded` only when `childCount > 0`, label is a `Link` to `getAssetPath`, separate expand toggle), `components/tree-children/tree-children.tsx` (`role="group"`, mounted only while expanded, its own `getAssetNodes(id)` query, loading / error + retry rows), `components/status-label/status-label.tsx` (words + icon, nothing for IN_SERVICE), each with `styles.module.css`; expansion state local in `asset-tree.tsx` — T033 passes
- [ ] T044 [P] [US1] Implement the details in `frontend/src/pages/assets/components/asset-details/`: `asset-details.tsx` (`getAssetDetail` with `skipToken` when nothing is selected), `components/asset-path/asset-path.tsx` (ordered links, current asset not a link), `components/asset-attributes/asset-attributes.tsx` (`<dl>`, "Not recorded", units, dates as received), `components/details-message/details-message.tsx` (prompt / loading / error + retry), each with `styles.module.css` — T035 passes
- [ ] T045 [US1] Implement `frontend/src/pages/assets/components/asset-explorer/asset-explorer.tsx` (two columns, stacked under the narrow breakpoint, tokens only) and `frontend/src/pages/assets/components/empty-registry/empty-registry.tsx`; update `frontend/src/pages/assets/assets.tsx` to read `assetId` from `useParams`, render the empty state or the explorer, keep the import panel — T036 passes
- [ ] T046 [US1] **Checkpoint — real stack**: `docker compose up --build -d --wait`; import the sample; quickstart API B, C, E and UI 3 and 5 (tree, lazy expansion seen in the browser's network tab, details of `PNL-001-1-01`); `make check`, `make lint`, `pnpm lint`, `pnpm build`, `pnpm test` exit 0; list the UI steps left for the user

**Checkpoint**: The explorer works for browsing and details (MVP).

---

## Phase 4: User Story 2 - Substation overview (Priority: P2)

**Goal**: A selected substation shows its children grouped by type and the counts of each type beneath it; a
selected switchboard lists its panels.

**Independent Test**: Select `SS-001` and check groups 2 / 1 / 2 and counts 2 / 1 / 2 / 10; select `SWB-001-1` and
check its panels (quickstart API D, J; UI 4).

### Tests for User Story 2 (write first, confirm they FAIL) ⚠️

- [ ] T047 [P] [US2] sqlmock test in `backend/internal/repository/psql/asset/count_descendants_by_type_test.go`: recursive CTE downward from the asset with the ID as a `?` argument, `GROUP BY asset_type`; rows → `map[valueobject.AssetType]int`; no descendants → empty non-nil map; DB error; trx given; wrong Tx type
- [ ] T048 [P] [US2] Extend `backend/internal/usecase/asset/get_asset_detail_test.go`: substation → `CountDescendantsByType` called once and `DescendantCounts{Transformers, LVBoards, Switchboards, SwitchboardPanels}` filled, a type without descendants = 0 (FR-013); any other type → `DescendantCounts == nil` and no count query; count error returned as-is
- [ ] T049 [P] [US2] Extend `backend/internal/handler/asset/get_asset_detail_test.go`: `descendantCounts` `{ "transformers", "lvBoards", "switchboards", "switchboardPanels" }` for a substation, `null` otherwise
- [ ] T050 [P] [US2] Extend `backend/internal/integrationtest/asset_explorer_test.go`: `SS-001` counts 2 / 1 / 2 / 10; for every one of the 13 substations the counts equal the numbers obtained by walking `GetListAsset` through its subtree (SC-002); `SWB-001-1` detail lists its panels in ID order
- [ ] T051 [P] [US2] Frontend tests `frontend/src/pages/assets/components/asset-details/components/child-groups/child-groups.test.tsx` and `descendant-counts/descendant-counts.test.tsx`: three groups in the order Transformers, LV boards, Switchboards with "(n)", each child a link with ID, name and status, an empty group "None" (FR-012, FR-014); the four counts with labels (FR-013)
- [ ] T052 [P] [US2] Frontend test `frontend/src/pages/assets/components/asset-details/components/child-list/child-list.test.tsx`: a switchboard's "Switchboard panels (n)" list of links with ID, name, status (FR-015); extend `asset-details.test.tsx` so a substation renders groups + counts, a switchboard renders the panel list, other types neither

### Implementation for User Story 2

- [ ] T053 [US2] Add `CountDescendantsByType(ctx, trx, assetID string) (map[valueobject.AssetType]int, error)` to `backend/domain/repository/asset.go`, `DescendantCounts{Transformers, LVBoards, Switchboards, SwitchboardPanels int}` and the `DescendantCounts *DescendantCounts` output field (nil unless a substation) to `backend/domain/usecase/asset.go`; `make mocks`
- [ ] T054 [US2] Implement `backend/internal/repository/psql/asset/count_descendants_by_type.go` (`db.Raw` recursive CTE, `?` placeholders) — T047 passes
- [ ] T055 [US2] Extend `backend/internal/usecase/asset/get_asset_detail.go` and add `toDescendantCounts` in `backend/internal/usecase/asset/helpers.go` — T048 passes
- [ ] T056 [US2] Add `DescendantCountsResponse` to `backend/domain/dto/asset.go` and map it in `backend/internal/handler/asset/helpers.go` — T049 and T050 pass
- [ ] T057 [P] [US2] Add `descendant-counts-dto.type.ts` and the `descendantCounts` field to `frontend/src/modules/assets/libs/types/`; implement `child-groups/child-groups.tsx`, `descendant-counts/descendant-counts.tsx`, `child-list/child-list.tsx` (each with `styles.module.css`) under `frontend/src/pages/assets/components/asset-details/components/`, grouping consecutive children by `assetType` in the order received (no re-sorting, principle II), and render them from `asset-details.tsx` — T051 and T052 pass
- [ ] T058 [US2] **Checkpoint — real stack**: rebuild; quickstart API D and J and UI 4 (select `SS-001`, then `SWB-001-1` from the Switchboards group); all five check commands exit 0; list the UI steps left for the user

**Checkpoint**: Stories 1 and 2 work.

---

## Phase 5: User Story 3 - Every asset has a stable address (Priority: P2)

**Goal**: Opening or refreshing `/assets/<id>` selects the asset with its ancestors expanded, highlighted and in
view; back/forward walk the selections; unknown IDs show "not found"; every stored ID round-trips.

**Independent Test**: Open `/assets/PNL-001-1-01` in a new tab, refresh, use back/forward, open `/assets/NOPE`, and
round-trip an ID with `/` and a space (quickstart API F–I; UI 6–8).

### Tests for User Story 3 (write first, confirm they FAIL) ⚠️

- [ ] T059 [P] [US3] Extend `backend/internal/handler/asset/get_asset_detail_test.go` with the encoding cases from research R2: `/api/assets/SS%2F1` → usecase receives `SS/1`; `%20`, `%25`, `%3F%23` decoded; `/api/assets/%zz` → `400 INVALID_REQUEST` without calling the usecase
- [ ] T060 [P] [US3] Extend `backend/internal/integrationtest/asset_explorer_test.go`: import a file with substation `SS/9 A` and a transformer under it; the detail of `SS/9 A` (requested through the router with `SS%2F9%20A`) and its children list round-trip; `ss-001` (letter case) and `NOPE` → `ASSET_NOT_FOUND`
- [ ] T061 [P] [US3] Frontend test `frontend/src/pages/assets/libs/hooks/use-expanded-nodes/use-expanded-nodes.hook.test.tsx`: when the selected asset's path arrives, its ancestors are added to the expanded set; the user can still collapse them afterwards and the selection stays (spec edge case); a new selection adds its own ancestors; no effect copies query data into state
- [ ] T062 [P] [US3] Frontend routing tests in `frontend/src/pages/assets/assets.test.tsx` (memory router): starting at `/assets/PNL-001-1-01` → `SS-001` and `SWB-001-1` expanded, the panel `aria-selected="true"` with the visible marker, `scrollIntoView` called on it (FR-006, FR-019); select A then B, history back → A selected, forward → B (FR-018); `/assets/SS%2F9%20A` → detail requested for `SS/9 A`; `/assets` → no selection (FR-021)
- [ ] T063 [P] [US3] Frontend test for the not-found state in `frontend/src/pages/assets/components/asset-details/asset-details.test.tsx`: a 404 with `ASSET_NOT_FOUND` → "No asset with ID NOPE exists." and a "Back to all assets" link to `/assets`, no toast, no error text, the tree still rendered (FR-020); a 404 without that code stays an error

### Implementation for User Story 3

- [ ] T064 [US3] Add `decodeAssetID` (`url.PathUnescape`, failure → `apperr.BadRequest(errorcode.InvalidRequest, "assetId must be valid percent-encoding")`) to `backend/internal/handler/asset/helpers.go` and use it in `get_asset_detail.go` — T059 and T060 pass
- [ ] T065 [P] [US3] Implement `frontend/src/pages/assets/libs/hooks/use-expanded-nodes/use-expanded-nodes.hook.ts` (expanded set + "adjust state when the selection changes" during render, research R8) and use it in `asset-tree.tsx` instead of the plain local state — T061 passes
- [ ] T066 [P] [US3] Implement `frontend/src/pages/assets/libs/helpers/check-is-asset-not-found.helper.ts` (+ test) and the not-found variant of `details-message.tsx` — T063 passes
- [ ] T067 [US3] Selection highlight and scroll in `frontend/src/pages/assets/components/asset-tree/components/tree-node/tree-node.tsx` (`aria-selected`, marker + bold, not colour alone; `scrollIntoView({ block: "nearest" })` in an effect on the selected element only when the selection changes) — T062 passes
- [ ] T068 [US3] **Checkpoint — real stack**: rebuild; quickstart API F, G, H, I and UI 6, 7, 8 (new tab, refresh, back/forward, not found, the `SS/9 A` address); all five check commands exit 0; list the UI steps left for the user

**Checkpoint**: Stories 1–3 work.

---

## Phase 6: User Story 4 - The tree reflects new imports (Priority: P3)

**Goal**: The import panel stays on the asset page without losing the selection, and a committed import shows
the new assets without a reload.

**Independent Test**: With an asset selected, open the panel, commit a file adding a substation, close the panel:
the new substation is in the tree, the selection is unchanged (quickstart UI 2, 10).

### Tests for User Story 4 (write first, confirm they FAIL) ⚠️

- [ ] T069 [P] [US4] Extend `frontend/src/modules/imports/imports-api.test.tsx`: after a `createImport` answer with `isCommitted: true`, a mounted `getAssetNodes(null)` query refetches; after `isCommitted: false` (validated only, nothing to commit, strict-blocked) or a failure, it does not (FR-026)
- [ ] T070 [P] [US4] Extend `frontend/src/pages/assets/assets.test.tsx`: on `/assets/SS-001`, **Import CSV** → address `/assets/SS-001?panel=import`, closing → `/assets/SS-001` with the same selection (FR-025); an empty registry followed by a committed import replaces the empty state with the tree (US4 scenario 4)

### Implementation for User Story 4

- [ ] T071 [US4] In `frontend/src/modules/imports/imports-api.ts`: `enhanceEndpoints({ addTagTypes: [AssetsApiTag.ASSET] })` and `invalidatesTags: (result) => (result?.isCommitted ? [AssetsApiTag.ASSET] : [])` on `createImport`, importing `AssetsApiTag` from `~/modules/assets/assets.js` — T069 passes
- [ ] T072 [US4] Make `frontend/src/pages/assets/assets.tsx` render the import panel in both the empty state and the explorer and keep the path when the panel opens/closes (`usePanel` only touches search parameters) — T070 passes
- [ ] T073 [US4] **Checkpoint — real stack**: `docker compose down -v && docker compose up --build -d --wait`; quickstart UI 1, 2 and 10 (empty state, first import fills the tree without reload, panel over a selected asset, re-import "Nothing to commit" leaves the tree unchanged); all five check commands exit 0; list the UI steps left for the user

**Checkpoint**: All four stories work.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Scale check, API and project documentation, security and coverage gates, and the final Definition
of Done run.

- [ ] T074 Scale test in `backend/internal/integrationtest/asset_explorer_scale_test.go`: import the 50 000-row file from `large_file_test.go` (reuse its generator) plus a file with one switchboard holding 5 000 panels; `GetListAsset` top level ≤ 2 s and that switchboard's children ≤ 1 s, `GetAssetDetail` of a substation ≤ 1 s (SC-006); log the measured times
- [ ] T075 [P] Write `backend/docs/api/asset.md` (English, exception 9) from [contracts/assets-api.md](contracts/assets-api.md): both endpoints, the node and detail shapes, ordering, encoding, errors, "complete lists (Template Exception 13)"; link it from `backend/README.md`
- [ ] T076 [P] Update `docs/ARCHITECTURE.md`: explorer endpoints and the `asset` module in the layers section, the recursive CTE reads and natural ordering (research R3, R4), frontend tree / details / lazy children / tag invalidation (R7, R8), the address scheme and encoding (R2), and add the research R1–R4, R8 trade-offs to the table
- [ ] T077 [P] Update `docs/KNOWN_LIMITATIONS.md`: complete lists without paging or virtualised rendering (Exception 13, R11); detail reads are not one snapshot and can straddle an import commit until the refetch (R5); expanded nodes other than the selection's ancestors are not kept across refresh; no search yet (next feature)
- [ ] T078 [P] Update the READMEs: root `README.md` (what the app does now includes browsing; the `/assets/<id>` address; API doc link), `backend/README.md` (the two new routes in the route table), `frontend/README.md` (backend contract table with the two asset endpoints; reference feature mentions `pages/assets` explorer)
- [ ] T079 [P] Security pass per `backend/docs/rules/07-keamanan.md` (minus auth): the new SQL uses `?` placeholders only (no concatenation, no `Sprintf`); logs carry asset IDs and counts, no row contents; `ASSET_NOT_FOUND` and `500` leak no detail; `git ls-files | grep -E '(^|/)\.env$'` is empty
- [ ] T080 Coverage gate: `make cover` in `backend/` (Go container) reports `internal/usecase` ≥ 80 %; add missing branch tests if not
- [ ] T081 [P] Remove the "Sync Impact Report" HTML comment from the top of `.specify/memory/constitution.md` (review scaffolding left by the v1.2.0 amendment)
- [ ] T082 Final Definition of Done run (constitution): `docker compose down -v && docker compose up --build` starts cleanly; `docker compose --profile test run --rm backend-test` and `frontend-test` exit 0; `make lint`, `pnpm lint`, `pnpm build`, `pnpm test` exit 0; full [quickstart.md](quickstart.md) (API A–J, UI 1–12, scale check) in the real app, plus a regression run of feature 001's API scenarios A–I; `docs/ARCHITECTURE.md` and `docs/KNOWN_LIMITATIONS.md` reflect the final code; no secrets committed. Record every command with its exit code and list the UI steps left for the user

---

## Dependencies & Execution Order

### Phase dependencies

- **Setup (Phase 1)**: none.
- **Foundational (Phase 2)**: needs T001. Tests T002–T010 before their implementations: T002 → T011, T003 → T012,
  T004 → T013, T005 → T017, T006 → T018, T007 → T019, T008 → T020, T009 → T026, T010 → T025. T014 needs T012;
  T015 needs T012; T016 needs T014 + T015; T018–T020 need T014 + T017; T022 needs T015 + T016; T025 needs T023 +
  T024; T027 needs T024.
- **US1 (Phase 3)**: needs Phase 2. T028–T036 before T037–T045; T037/T038 → T039/T040 → T041; T042 before T043;
  T043 + T044 → T045; T046 last.
- **US2 (Phase 4)**: needs US1 (extends the detail usecase, handler and details component). T047–T052 before
  T053–T057; T053 → T054 → T055 → T056; T058 last.
- **US3 (Phase 5)**: needs US1 (tree and details). Independent of US2. T059–T063 before T064–T067; T068 last.
- **US4 (Phase 6)**: needs US1 (tree and page). Independent of US2 and US3. T069–T070 before T071–T072; T073 last.
- **Polish (Phase 7)**: needs all stories; T082 is last.

### User story dependencies

```text
Setup → Foundational → US1 (browse + details, MVP) ─┬→ US2 (substation overview)
                                                     ├→ US3 (stable addresses)
                                                     └→ US4 (refresh after import)
                                         US2 + US3 + US4 → Polish
```

US2, US3 and US4 touch different files except `asset-details.test.tsx` (US2, US3), `get_asset_detail_test.go` and
`helpers.go` in `handler/asset` (US2, US3), `asset_explorer_test.go` (US1–US3) and `assets.test.tsx` (US1, US3,
US4): serialise edits to those files.

### Within each story

Tests (fail) → domain contract → repository → usecase → handler → route → integration test → frontend data →
frontend components → real-stack checkpoint.

## Parallel Examples

```text
# Foundational: all tests at once, then the independent implementations
T002 · T003 · T004 · T005 · T006 · T007 · T008 · T009 · T010
T011 · T012 · T013 · T017 · T021 · T023 · T024 · T026, then T018 · T019 · T020

# US1: backend and frontend tests together
T028 · T029 · T030 · T031 · T032 · T033 · T034 · T035 · T036
T042 · T043 · T044 (frontend) while T037–T041 (backend) proceed

# US2: repository, usecase, handler, integration and UI tests
T047 · T048 · T049 · T050 · T051 · T052

# US3 and US4 can run side by side after US1 (mind the shared test files above)
T059 · T060 · T061 · T062 · T063   |   T069 · T070

# Polish: docs and checks
T075 · T076 · T077 · T078 · T079 · T081
```

## Implementation Strategy

### MVP first (User Story 1 only)

1. Phase 1 baseline, Phase 2 foundation.
2. Phase 3 (US1): the tree with lazy children and the details area.
3. **Stop and validate** at T046 in the real stack.

### Incremental delivery

1. US1 → browse and inspect (MVP).
2. US2 → substation overview and switchboard panels.
3. US3 → addresses, deep links, not found, encoded IDs.
4. US4 → tree refresh after a committed import.
5. Polish → scale, docs, DoD.

Each story ends with a real-stack checkpoint; stop there and report before moving on.

## Notes

- [P] = different files, no dependency on an unfinished task; [USn] maps the task to its story.
- Write each test first and see it fail; then the smallest code that makes it pass.
- Backend: one file per endpoint / use case / repository method, each with its `_test.go`; `make mocks` after any
  change to `domain/` interfaces; never edit generated mocks.
- Frontend: one declaration per file, barrels only where the conventions allow, no inline handlers in JSX,
  tokens for every style value, every state rendered (loading, empty, error, not found, success).
- Ordering, grouping order and counts come from the backend; the frontend never re-sorts (principle II).
- Commit only when the user asks.
