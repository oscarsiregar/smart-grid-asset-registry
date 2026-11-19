---

description: "Task list for the asset search"
---

# Tasks: Asset Search

**Input**: Design documents from `specs/003-asset-search/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md),
[contracts/assets-search-api.md](contracts/assets-search-api.md), [contracts/ui-asset-search.md](contracts/ui-asset-search.md),
[quickstart.md](quickstart.md); constitution v1.3.0 (Template Exception 14 covers the search paging cut in the usecase).

**Tests**: Requested (constitution principle IV, user input "tests before code"). Every test task is written first
and must FAIL before the task that makes it pass.

**Organization**: Tasks are grouped by user story. Reference modules to extend: backend `internal/{handler,usecase}/asset`,
`internal/repository/psql/asset`, `domain/entity/asset_node.go`; frontend `modules/assets`, `pages/assets` (explorer,
tree, details from feature 002) and the import form's controlled fields. One file per endpoint / use case /
repository method / component / hook / helper / type / enum, each with its test next to it (`backend/AGENTS.md`,
`frontend/docs/code-conventions.md` §3).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on an unfinished task)
- **[Story]**: US1–US3 from [spec.md](spec.md)

## Path Conventions

Web app: `backend/` (Go module `smart-grid-asset-registry`) and `frontend/src/`. Backend checks: `make check`,
`make lint` from `backend/` (or `docker compose --profile test run --rm backend-test`); integration tests need
`TEST_DATABASE_URL=postgres://postgres:postgres@localhost:55432/asset_registry_test?sslmode=disable`. Frontend checks:
`pnpm lint`, `pnpm build`, `pnpm test` from `frontend/`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm the starting point; no new project, dependency or migration is needed (plan: no schema change,
no new dependency).

- [ ] T001 Baseline on branch `003-asset-search`: `docker compose up --build -d --wait`, then `docker compose --profile test run --rm backend-test`, `make lint` (golangci-lint container), `pnpm lint && pnpm build && pnpm test` in `frontend/` all exit 0; record the commands and exit codes as the baseline in the implementation report

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The search ordering rule, the extended list input and filter, the repository reads, the DTOs and the
frontend data module that every story uses.

**⚠️ CRITICAL**: No user-story work can begin until this phase is complete.

### Tests first (write, confirm they FAIL)

- [ ] T002 [P] Table tests for `CheckIsExactMatch(node *AssetNode, text string) bool` and `SortSearchMatches(nodes []*AssetNode, text string)` in `backend/domain/entity/asset_search_test.go` (research R3): equal ID or equal name is exact, letter case ignored (`pnl-001-1-01` vs `PNL-001-1-01`); a contained text is not exact; exact matches come first, then the rest; each part in `SortAssetNodes` order (type rank, then natural ID: `SS-2` before `SS-10`); stable and total for shuffled input; nil and empty slices do not panic
- [ ] T003 [P] Table tests for `GetListAssetInput.Validate()` and `SetDefault()` in `backend/domain/usecase/asset_test.go` (data-model "Usecase contract"): `Search` longer than 100 characters → `q must be at most 100 characters`; `Search` with `ParentAssetID` → `q and parentAssetId cannot be combined`; `AssetType` without `Search` → `assetType requires q`; `Page` or `Limit` without `Search` → `page and limit apply to a search only`; `Page < 0` or `Limit < 0` → `page and limit must not be negative`; hierarchy inputs (nil parent, a parent) valid; `SetDefault` in search mode gives page 1 and limit `SearchDefaultLimit` (50) when 0, caps `Limit` at `MaxLimit` (100), and leaves hierarchy inputs untouched; constants `MaxLimit = 100`, `SearchDefaultLimit = 50`, `MaxSearchLength = 100`
- [ ] T004 [P] Extend the sqlmock test `backend/internal/repository/psql/asset/find_nodes_test.go`: with `Search` the query has no `parent_asset_id IS NULL` condition and selects `WHERE (a.asset_id ILIKE $1 OR a.asset_name ILIKE $2)` with both arguments `%<text>%`; `%`, `_` and `\` in the text are escaped (`querysafety.ContainsPattern`); with `AssetType` it adds `AND a.asset_type = $3`; the `child_count` subquery is still selected; the hierarchy cases of feature 002 stay unchanged
- [ ] T005 [P] sqlmock test `backend/internal/repository/psql/asset/find_ancestor_ids_test.go` (research R4): recursive CTE upward from `asset_id = ANY($1)` with the IDs as one text-array argument; rows `(asset_id, ancestor_id, depth)` map to `map[assetID][]ancestorID` substation first; a substation maps to an empty slice; empty input → empty non-nil map and no query; DB error; trx given; wrong Tx type
- [ ] T006 [P] Extend `frontend/src/modules/assets/assets-api.test.tsx`: `searchAssets({ assetType: null, q: "PNL 1/2" })` requests `GET /api/assets?q=PNL%201%2F2` with no `page` or `limit` (the backend's default page of 50 applies; principle II); with `assetType: "SWITCHBOARD_PANEL"` it adds `assetType=SWITCHBOARD_PANEL`; the envelope's `data` (`items` with `ancestorAssetIds`, `totalItems`) is returned; an error produces no toast and keeps its code
- [ ] T007 [P] Test `frontend/src/modules/assets/libs/validation-schemas/asset-search.validation-schema.test.ts`: `q` is trimmed; blank or only spaces → "Enter an asset ID or name to search."; more than 100 characters → "Use at most 100 characters."; `assetType` accepts each `AssetType` value and `null` (all types); nothing else is checked (principle II)

### Implementation

- [ ] T008 [P] Implement `backend/domain/entity/asset_search.go` (`CheckIsExactMatch`, `SortSearchMatches` with `slices.SortStableFunc`, reusing `SortAssetNodes`' comparison) — T002 passes
- [ ] T009 [P] Extend `backend/domain/usecase/asset.go`: `GetListAssetInput` + `Search string`, `AssetType *valueobject.AssetType`, `Page int`, `Limit int`; `Validate()` and `SetDefault()` per T003; `GetListAssetOutput` + `TotalItems int` and `AncestorIDs map[string][]string`; constants `MaxLimit`, `SearchDefaultLimit`, `MaxSearchLength` — T003 passes
- [ ] T010 Extend `backend/domain/repository/find_asset_nodes_filter.go` with `Search *string` ("when set, matches every level by ID or name, letter case ignored, characters literal; ParentAssetID must be nil") and `AssetType *valueobject.AssetType` ("only with Search"); add `FindAncestorIDs(ctx, trx, assetIDs []string) (map[string][]string, error)` to `backend/domain/repository/asset.go`
- [ ] T011 Run `make mocks` (mockery v2.53 in the `golang:1.26.9` container) to regenerate `backend/internal/mocks/repository/AssetRepository.go`; do not edit generated files
- [ ] T012 [P] Extend `backend/internal/repository/psql/asset/find_nodes.go`: when `filter.Search != nil`, `Where("(a.asset_id ILIKE ? OR a.asset_name ILIKE ?)", pattern, pattern)` with `pattern := querysafety.ContainsPattern(*filter.Search)`, plus `Where("a.asset_type = ?", *filter.AssetType)` when set; otherwise the 002 hierarchy conditions — T004 passes
- [ ] T013 [P] Implement `backend/internal/repository/psql/asset/find_ancestor_ids.go` (`db.Raw` recursive CTE with `textArray(assetIDs)`, `?` placeholders only, depth limit as in `find_path.go`) and its projection `backend/internal/repository/psql/model/asset_ancestor.go` (`asset_id`, `ancestor_id`, `depth`; not a table) — T005 passes
- [ ] T014 [P] Extend `backend/domain/dto/asset.go`: `GetListAssetRequest` + `Q string \`query:"q"\``, `AssetType string \`query:"assetType"\``, `Page int \`query:"page"\``, `Limit int \`query:"limit"\``; `AssetSearchItemResponse` (the `AssetNodeResponse` fields + `AncestorAssetIDs []string \`json:"ancestorAssetIds"\``, never null) and `AssetSearchResponse{Items []AssetSearchItemResponse; TotalItems int}` (`items` never null) per [contracts/assets-search-api.md](contracts/assets-search-api.md)
- [ ] T015 [P] Add `QUERY: "q"` and `ASSET_TYPE: "type"` to `frontend/src/libs/enums/search-parameter.enum.ts`
- [ ] T016 [P] Frontend search data in `frontend/src/modules/assets/`: types `libs/types/{asset-search-item-dto.type.ts, asset-search-result-dto.type.ts, asset-search-request-dto.type.ts}` (+ `types.ts`; `assetType: null | ValueOf<typeof AssetType>`), query `searchAssets` in `assets-api.ts` (`params: { q }` plus `assetType` when not null, no `page` or `limit` and no frontend copy of the 50, `providesTags: [AssetsApiTag.ASSET]`, `extraOptions: { shouldSuppressToast: true }`), exports in `assets.ts` — T006 passes
- [ ] T017 [P] Implement `frontend/src/modules/assets/libs/validation-schemas/asset-search.validation-schema.ts` (+ `validation-schemas.ts` barrel, `AssetSearchValidationMessage` enum in `libs/enums/`; `assetType` is an `AssetType` value or `null`) and export `assetSearchValidationSchema` from `assets.ts` — T007 passes

**Checkpoint**: `make check`, `make lint`, `pnpm lint`, `pnpm build`, `pnpm test` exit 0; the explorer of 002 is unchanged.

---

## Phase 3: User Story 1 - Find an asset by ID or name and open it (Priority: P1) 🎯 MVP

**Goal**: The operator searches by all or part of an ID or name; the results (first 50 plus the total) replace the
tree; selecting one goes to `/assets/<id>`, where feature 002 reveals and shows it.

**Independent Test**: With the sample imported, search `PNL-001-1-01` and `aurora panel`, select a result, and
compare with opening `/assets/PNL-001-1-01` directly (quickstart API A, B, C, E, F, G, H, J; UI 1, 2, 3, 8, 11).

### Tests for User Story 1 (write first, confirm they FAIL) ⚠️

- [ ] T018 [P] [US1] Extend `backend/internal/usecase/asset/get_list_asset_test.go` with search cases (mocks per case): invalid input → `400 INVALID_REQUEST` and no repository call; `Search` → `FindNodes` with filter `{Search: &text}`, matches ordered by `SortSearchMatches`, `TotalItems` = all matches, `Assets` = the page (page 1 limit 2 of 5, page 3 of 5, a page past the end → empty with the real total); `FindAncestorIDs` called once with only the page's IDs and its map returned as `AncestorIDs`; no matches → no `FindAncestorIDs` call; each repository error returned as-is and logged; the success log carries the search length and the total, never the search text; hierarchy cases of 002 unchanged with `TotalItems = len(Assets)` and `AncestorIDs == nil`
- [ ] T019 [P] [US1] Extend `backend/internal/handler/asset/get_list_asset_test.go`: `?q=%20PNL%2F1%20` → usecase receives `Search "PNL/1"` (trimmed, decoded); `?q=` and `?q=%20%20` → hierarchy mode; `page`/`limit` forwarded; a search answers 200 `"Successfully get list asset"` with `items` (node fields + `ancestorAssetIds`, `[]` for a substation, never null) and `totalItems`; usecase 400 passes through with `data.errorCode INVALID_REQUEST`; hierarchy responses keep the 002 shape without `ancestorAssetIds`
- [ ] T020 [P] [US1] Integration test `backend/internal/integrationtest/asset_search_test.go` (real PostgreSQL, `pgtest.Open`): import the sample, then for `PNL-001-1-01` the first item is that panel with ancestors `[SS-001, SWB-001-1]`; for every accepted sample row, `q=<its asset ID>` lists that asset first (SC-001); a text that matches some assets by ID and others by name, and one asset by both (for example `Aurora`), lists each asset exactly once (FR-009); for `aurora panel`, `swb-001`, `001`, `-` and `T1` the set of returned IDs over all pages equals a SQL oracle `asset_id ILIKE '%x%' OR asset_name ILIKE '%x%'` (SC-002); letter case ignored; `%` and `_` match literally (import a file with assets `A%B` and `A_B` and check `q=%` finds only `A%B`); page 2 of `-` has no overlap with page 1; `totalItems` = oracle count; the hierarchy reads of 002 are unchanged
- [ ] T021 [P] [US1] Frontend test `frontend/src/pages/assets/libs/hooks/use-asset-search/use-asset-search.hook.test.tsx` (memory router): reads `q` and `type` from the address (`null` when absent or blank); `submitSearch({ q, assetType })` pushes a history entry with `q` (and `type` unless `assetType` is `null`, all types), keeping other parameters such as `panel`; `clearSearch()` removes both; Back after a submit restores the previous search
- [ ] T022 [P] [US1] Frontend test `frontend/src/pages/assets/components/asset-search/components/search-form/search-form.test.tsx`: a `role="search"` form with the labelled field "Search assets" (restored `Input`); Enter and **Search** submit the trimmed text; a blank text shows "Enter an asset ID or name to search." and does not submit; the field starts with the search from the address; **Clear** shows only while a search is in the address and clears it
- [ ] T023 [P] [US1] Frontend test `frontend/src/pages/assets/components/asset-search/components/search-results/search-results.test.tsx` (`fetch` mocked, real store): a summary "N assets match "<text>""; each result is one link to `/assets/<id>` with ID and name, the type and status in words ("No status recorded" for none), and the location `SS-001 › SWB-001-1` or "Top level"; a result with ID `..` has no link and says "This asset cannot be opened."; results appear in the order received (no re-sorting)
- [ ] T024 [P] [US1] Frontend test `frontend/src/pages/assets/libs/helpers/get-asset-location-text.helper.test.ts`: `["SS-001","SWB-001-1"]` → `SS-001 › SWB-001-1`; `[]` → `Top level`
- [ ] T025 [P] [US1] Extend `frontend/src/pages/assets/assets.test.tsx`: at `/assets?q=PNL-001-1-01` the results replace the tree (no `role="tree"`) and the details area still invites a selection; clicking the first result → address `/assets/PNL-001-1-01` with no `q`, the tree back with `SS-001` and `SWB-001-1` expanded and the panel `aria-selected="true"` (feature 002's reveal); history back → the search and its results again; submitting from the form at `/assets` → `?q=`; **Clear** → the tree, no `q`; with an empty registry there is no search form, only feature 002's empty state (FR-001, spec edge case)

### Implementation for User Story 1

- [ ] T026 [US1] Restore the template `Input` byte for byte from git: `git show 40d511b^:frontend/src/libs/components/input/input.tsx`, `…/input/styles.module.css` and `…/libs/enums/input-type.enum.ts` into the same paths, and re-add `export { InputType } from "./input-type.enum.js";` to `frontend/src/libs/enums/enums.ts` (research R7); do not rewrite it
- [ ] T027 [US1] Implement the search branch of `backend/internal/usecase/asset/get_list_asset.go` (log setup → `SetDefault` → `Validate` → `FindNodes` with the search filter → `SortSearchMatches` → total → page cut via a `pageOf` helper in `helpers.go` → `FindAncestorIDs` for the page → log success with search length and total; Template Exception 14) — T018 passes
- [ ] T028 [US1] Extend `backend/internal/handler/asset/get_list_asset.go` (trim `q`; search vs hierarchy input; `page`/`limit`) and `backend/internal/handler/asset/helpers.go` (`toAssetSearchResponse`, ancestors `[]` when missing) — T019 passes
- [ ] T029 [US1] Run `backend/internal/integrationtest/asset_search_test.go` (T020) against `TEST_DATABASE_URL` and fix any gap until it passes (`go test -count=1 -run TestAssetSearch -v ./internal/integrationtest/`)
- [ ] T030 [P] [US1] Implement `frontend/src/pages/assets/libs/hooks/use-asset-search/use-asset-search.hook.ts` (`useSearchParams`, `SearchParameter.QUERY` / `ASSET_TYPE`, memoised `submitSearch` / `clearSearch`) — T021 passes
- [ ] T031 [P] [US1] Implement `frontend/src/pages/assets/libs/helpers/get-asset-location-text.helper.ts` (+ `helpers.ts` export) and the page copy in `frontend/src/pages/assets/libs/enums/assets-page-message.enum.ts` (every text in [contracts/ui-asset-search.md](contracts/ui-asset-search.md)) — T024 passes
- [ ] T032 [P] [US1] Implement `frontend/src/pages/assets/components/asset-search/components/search-form/search-form.tsx` (+ `styles.module.css`): `useAppForm` with `assetSearchValidationSchema`, defaults from the address, the restored `Input` with a search icon, **Search** and **Clear** buttons — T022 passes
- [ ] T033 [P] [US1] Implement `frontend/src/pages/assets/components/asset-search/components/search-results/search-results.tsx` and `components/search-result/search-result.tsx` (+ `styles.module.css` each): `searchAssets` query, `role="region"` labelled "Search results", the summary, one list item per result with the shared `Link` to `getAssetPath`, plain text for `checkIsAddressableAssetId(id) === false` — T023 passes
- [ ] T034 [US1] Implement `frontend/src/pages/assets/components/asset-search/asset-search.tsx` (form + results) and update `frontend/src/pages/assets/components/asset-explorer/asset-explorer.tsx` to render the search form above the tree area and the results in place of the tree while `use-asset-search` reports a search; `frontend/src/pages/assets/assets.tsx` unchanged apart from passing what the explorer needs — T025 passes
- [ ] T035 [US1] **Checkpoint — real stack**: `docker compose up --build -d --wait`; quickstart API A, B, C, E, F, G, H, J and UI 1, 2, 3, 8, 11; `make check`, `make lint`, `pnpm lint`, `pnpm build`, `pnpm test` exit 0; list the UI steps left for the user

**Checkpoint**: The search finds assets and opens them through feature 002 (MVP).

---

## Phase 4: User Story 2 - Narrow a search by asset type (Priority: P2)

**Goal**: A type chosen next to the text limits the results; the type is part of the address.

**Independent Test**: Search `001` with and without "Switchboard panel" (quickstart API D; UI 4).

### Tests for User Story 2 (write first, confirm they FAIL) ⚠️

- [ ] T036 [P] [US2] Extend `backend/internal/handler/asset/get_list_asset_test.go`: `?q=001&assetType=SWITCHBOARD_PANEL` → usecase receives that type; `?q=x&assetType=FOO` → `400 INVALID_REQUEST "Invalid request: assetType is not a valid asset type"` without calling the usecase; `?assetType=TRANSFORMER` → usecase error `assetType requires q` as 400
- [ ] T037 [P] [US2] Extend `backend/internal/usecase/asset/get_list_asset_test.go`: `AssetType` is passed to `FindNodes` in the filter
- [ ] T038 [P] [US2] Extend `backend/internal/integrationtest/asset_search_test.go`: `q=001` with `SWITCHBOARD_PANEL` returns only panels and equals the SQL oracle with `AND asset_type = 'SWITCHBOARD_PANEL'`; every other type likewise
- [ ] T039 [P] [US2] Frontend test `frontend/src/pages/assets/components/asset-search/components/search-form/components/asset-type-field/asset-type-field.test.tsx`: a labelled `<select>` "Asset type" with "All types" (`AssetTypeFilter.ALL` from `frontend/src/pages/assets/libs/enums/asset-type-filter.enum.ts`, submitted as `null`) and the five types in words; extend `frontend/src/pages/assets/components/asset-search/components/search-form/search-form.test.tsx`: the chosen type is submitted and written to `type`, "All types" removes it, and the field starts with the type from the address
- [ ] T040 [P] [US2] Extend `frontend/src/pages/assets/components/asset-search/components/search-results/search-results.test.tsx`: with a type, the summary reads "N assets match "<text>" of type <Type>" and the request carries `assetType`

### Implementation for User Story 2

- [ ] T041 [US2] Parse `assetType` in `backend/internal/handler/asset/get_list_asset.go` with `valueobject.ParseAssetType` (failure → `apperr.BadRequest(errorcode.InvalidRequest, "assetType is not a valid asset type")`) and pass it in the usecase input; confirm the usecase forwards it — T036 and T037 pass, then T038 against `TEST_DATABASE_URL`
- [ ] T042 [P] [US2] Implement `frontend/src/pages/assets/components/asset-search/components/search-form/components/asset-type-field/asset-type-field.tsx` (+ `styles.module.css`; `useController`, labels from `assetTypeToLabel`, like the import form's mode field), the page enum `frontend/src/pages/assets/libs/enums/asset-type-filter.enum.ts` (`AssetTypeFilter.ALL`, + `enums.ts` export) and its mapping to `null` on submit and add it to `frontend/src/pages/assets/components/asset-search/components/search-form/search-form.tsx` — T039 passes
- [ ] T043 [P] [US2] Pass the address's type to the query and the summary in `frontend/src/pages/assets/components/asset-search/components/search-results/search-results.tsx` — T040 passes
- [ ] T044 [US2] **Checkpoint — real stack**: rebuild; quickstart API D and UI 4 (type narrows, refresh keeps text and type); all five check commands exit 0; list the UI steps left for the user

**Checkpoint**: Stories 1 and 2 work.

---

## Phase 5: User Story 3 - Understand every state of a search (Priority: P2)

**Goal**: Running, no-match, more-than-50 and failed searches each say so in words; a committed import refreshes a
shown search; the import panel and the search coexist.

**Independent Test**: Search `NO-SUCH-ASSET`, `-`, and with the `api` stopped (quickstart API G, I; UI 5, 6, 7, 9, 10).

### Tests for User Story 3 (write first, confirm they FAIL) ⚠️

- [ ] T045 [P] [US3] Extend `frontend/src/pages/assets/components/asset-search/components/search-results/search-results.test.tsx`: while loading → "Searching…" (`role="status"`); 0 matches → "No asset matches "<text>"" (+ type) and "Check the spelling, or search all types."; `totalItems` 1 234 with 50 items → "Showing 50 of 1234 assets that match "<text>". Add more of the ID or name, or choose a type, to narrow the search."; the count shown is the number of items received, never a frontend constant; a 500 → "The search could not be completed." + **Try again** that refetches; no toast in any state
- [ ] T046 [P] [US3] Extend `frontend/src/modules/imports/imports-api.test.tsx`: a mounted `searchAssets` query refetches after an import with `isCommitted: true` and not after one that saved nothing (FR-017)
- [ ] T047 [P] [US3] Extend `frontend/src/pages/assets/assets.test.tsx`: at `/assets?q=x` a failed search keeps the details area working and **Clear** brings the tree back (FR-016); **Import CSV** keeps `q` and adds `panel=import`, closing keeps the search (spec edge case); while a search is still running, the search field can be edited and **Import CSV** opens the panel (FR-014)

### Implementation for User Story 3

- [ ] T048 [US3] Implement the states in `frontend/src/pages/assets/components/asset-search/components/search-results/search-results.tsx` with `frontend/src/pages/assets/components/asset-search/components/search-results/components/search-summary/search-summary.tsx` (+ `styles.module.css`) for the count / too-many / no-match texts and the retry row — T045 passes
- [ ] T049 [US3] Confirm FR-017 and the panel coexistence need no new code in `frontend/src/modules/imports/imports-api.ts` or `frontend/src/libs/hooks/use-panel/use-panel.hook.ts` (the `ASSET` tag of 002 and `usePanel` keeping other parameters); fix anything T046/T047 reveal — T046 and T047 pass
- [ ] T050 [US3] **Checkpoint — real stack**: rebuild; quickstart API G, I and UI 5, 6, 7, 9, 10; all five check commands exit 0; list the UI steps left for the user

**Checkpoint**: All three stories work.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Scale check, API and project documentation, constitution sync, security and coverage gates, and the
final Definition of Done run.

- [ ] T051 Extend `backend/internal/integrationtest/asset_explorer_scale_test.go`: with the ~55 000 stored assets, `GetListAsset` searches for `-` (every asset), `WIDE-4999` (one) and `PNL` with type `SWITCHBOARD_PANEL` each ≤ 1 s (SC-005, research R8); log the measured times
- [ ] T052 [P] Add a "Search" section to `backend/docs/api/asset.md` from [contracts/assets-search-api.md](contracts/assets-search-api.md): parameters and rules, the item shape with `ancestorAssetIds`, ordering (exact first), paging (Template Exception 14), errors
- [ ] T053 [P] Update `docs/ARCHITECTURE.md`: the search filter on the list read, `ILIKE` with `ContainsPattern`, `SortSearchMatches` and the page cut in the usecase (Exception 14), `FindAncestorIDs`, the frontend search form / results / URL state and opening a result through 002's deep link; add the research R0–R5 trade-offs to the table
- [ ] T054 [P] Update `docs/KNOWN_LIMITATIONS.md`: remove "No search yet"; add: only asset ID and name are searched; at most 50 results shown, refine to reach others; every search loads all matches before ordering (fine at 50 000; a `pg_trgm` index and SQL ordering would be the path beyond that); case folding of exotic characters can differ between matching and the "exact" decision (research R2)
- [ ] T055 [P] Update the READMEs: root `README.md` (search in the feature list; quickstart link), `backend/README.md` (the search parameters on `GET /api/assets`), `frontend/README.md` (backend contract row; the restored `Input`; `?q=`/`?type=`)
- [ ] T056 [P] Sync `AGENTS.md`'s Template Exceptions summary with constitution v1.3.0 (add Exception 14) and remove the "Sync Impact Report" HTML comment from the top of `.specify/memory/constitution.md`
- [ ] T057 [P] Security pass per `backend/docs/rules/07-keamanan.md` (minus auth): the search SQL uses `?` placeholders only and `ContainsPattern` escaping; `q` length validated (100); `limit` capped (100); logs carry the search length and counts, not the text or row contents; `git ls-files | grep -E '(^|/)\.env$'` is empty
- [ ] T058 Coverage gate: `make cover` in `backend/` (Go container) reports `internal/usecase` ≥ 80 %; add missing branch tests if not
- [ ] T059 Final Definition of Done run (constitution): `docker compose up --build` starts cleanly (on a fresh volume too: a separate compose project such as `docker compose -p asset-registry-dod` with other host ports, removed afterwards, if the main volume must be kept); `docker compose --profile test run --rm backend-test` and `frontend-test` exit 0; `make lint`, `pnpm lint`, `pnpm build`, `pnpm test` exit 0; full [quickstart.md](quickstart.md) (API A–J, UI 1–12 including the hand checks for SC-003 and SC-006, scale check) plus a regression run of features 001 and 002's API scenarios; `docs/ARCHITECTURE.md` and `docs/KNOWN_LIMITATIONS.md` reflect the final code; no secrets committed. Record every command with its exit code and list the UI steps left for the user

---

## Dependencies & Execution Order

### Phase dependencies

- **Setup (Phase 1)**: none.
- **Foundational (Phase 2)**: needs T001. Tests T002–T007 before their implementations: T002 → T008, T003 → T009,
  T004 → T012, T005 → T013, T006 → T016, T007 → T017. T010 needs T009 (types); T011 needs T010; T012 and T013 need
  T010 + T011; T016 needs T015; T017 needs T016 (barrel).
- **US1 (Phase 3)**: needs Phase 2. T018–T025 before T026–T034; T026 before T032; T027 → T028 → T029; T030 and T031
  before T032–T034; T033 needs T031; T034 needs T030, T032, T033; T035 last.
- **US2 (Phase 4)**: needs US1 (extends the form, the results and the handler). T036–T040 before T041–T043; T044 last.
- **US3 (Phase 5)**: needs US1 (the results component). Independent of US2. T045–T047 before T048–T049; T050 last.
- **Polish (Phase 6)**: needs all stories; T059 is last.

### User story dependencies

```text
Setup → Foundational → US1 (search and open, MVP) ─┬→ US2 (type filter)
                                                    └→ US3 (states)
                                         US2 + US3 → Polish
```

US2 and US3 touch different files except `search-results.test.tsx` / `search-results.tsx` (US2, US3),
`get_list_asset_test.go` (US1, US2), `asset_search_test.go` (US1, US2) and `assets.test.tsx` (US1, US3): serialise
edits to those files.

### Within each story

Tests (fail) → domain contract → repository → usecase → handler → integration test → frontend data → frontend
components → real-stack checkpoint.

## Parallel Examples

```text
# Foundational: all tests at once, then the independent implementations
T002 · T003 · T004 · T005 · T006 · T007
T008 · T009 · T014 · T015, then T010 → T011 → T012 · T013, and T016 → T017

# US1: backend and frontend tests together
T018 · T019 · T020 · T021 · T022 · T023 · T024 · T025
T030 · T031 · T032 · T033 (frontend) while T027–T029 (backend) proceed

# US2 and US3 can run side by side after US1 (mind the shared files above)
T036 · T037 · T038 · T039 · T040   |   T045 · T046 · T047

# Polish: docs and checks
T052 · T053 · T054 · T055 · T056 · T057
```

## Implementation Strategy

### MVP first (User Story 1 only)

1. Phase 1 baseline, Phase 2 foundation.
2. Phase 3 (US1): search by text, results in place of the tree, open through feature 002.
3. **Stop and validate** at T035 in the real stack.

### Incremental delivery

1. US1 → find and open (MVP).
2. US2 → type filter.
3. US3 → every state, refresh after import, coexistence with the import panel.
4. Polish → scale, docs, DoD.

Each story ends with a real-stack checkpoint; stop there and report before moving on.

## Notes

- [P] = different files, no dependency on an unfinished task; [USn] maps the task to its story.
- Write each test first and see it fail; then the smallest code that makes it pass.
- Backend: one file per endpoint / use case / repository method, each with its `_test.go`; `make mocks` after any
  change to `domain/` interfaces; never edit generated mocks.
- Frontend: one declaration per file, barrels only where the conventions allow, no inline handlers in JSX, tokens
  for every style value, every state rendered (running, no match, too many, failure, success).
- Matching, ordering, the page and the total come from the backend; the frontend never filters or re-sorts
  (principle II).
- Opening a result is a link to `/assets/<id>`; the reveal, highlight and details are feature 002's (FR-011).
- Commit only when the user asks.
