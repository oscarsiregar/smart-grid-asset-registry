# Tasks: UI polish to the reference design, with loading placeholders

**Input**: Design documents from `specs/005-ui-polish/` — [plan.md](plan.md), [spec.md](spec.md),
[research.md](research.md), [data-model.md](data-model.md), [contracts/ui-asset-page.md](contracts/ui-asset-page.md),
[quickstart.md](quickstart.md)

**Tests**: Tests come before code, as in every feature of this project (Vitest + React Testing Library, `fetch`
mocked). Write each test task first and confirm it fails before the matching implementation task. Existing
tests of features 001–004 keep their behaviour; only assertions on copy or structure this feature changes are
updated, in the task that changes it.

**Organization**: by user story — US1 the reference look (P1), US2 loading placeholders (P1), US3 the import
drawer (P2). Frontend only; the backend is not touched.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: can run in parallel (different files, no dependency on an unfinished task)
- **[Story]**: US1, US2, US3; setup, foundational and polish tasks have none

## Path Conventions

All paths are under `frontend/src/` unless they start with `specs/`, `docs/` or `frontend/`. Conventions:
`frontend/docs/code-conventions.md` (one declaration per file, barrels, tokens only, no inline styles, no magic
numbers, handlers in `useCallback`).

---

## Phase 1: Setup

**Purpose**: Tokens and the bolt mark.

- [X] T001 Add tokens to `assets/css/app-tokens.css`: `--motion-delay-skeleton: 200ms`, `--motion-duration-skeleton-pulse: 1200ms`, `--surface-skeleton` (a tint of `--surface-raised`), skeleton heights `--size-skeleton-text` (14px), `--size-skeleton-title` (28px), `--size-skeleton-block` (64px), `--size-skeleton-dot` (10px), `--size-brand-tile` (32px), `--layout-attribute-min-width` (160px); no value outside tokens (spec FR-002)
- [X] T002 [P] Add `assets/img/zap.svg` (a 24 × 24 lightning bolt, `stroke="currentColor"`, no fill colour), add `ZAP` to the `IconName` enum and map it in `libs/components/icon/libs/maps.ts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The shared pieces every story uses. No story starts before this phase is done.

- [X] T003 [P] Test `libs/components/skeleton/skeleton.test.tsx`: renders one element with `aria-hidden="true"`; the shape (`SkeletonShape.TEXT | TITLE | BLOCK | DOT`) and width (`SkeletonWidth.FULL | THREE_QUARTERS | HALF | QUARTER`) select CSS-module classes; defaults are `TEXT` and `FULL`; `className` is merged last
- [X] T004 [P] Test `libs/components/button/button.test.tsx` (new file): with `isLabelHidden` the label is in a `.visually-hidden` span and the button's accessible name is still the label; without it nothing changes
- [X] T005 Implement `libs/components/skeleton/skeleton.tsx` + `styles.module.css` + `libs/enums/{skeleton-shape.enum.ts, skeleton-width.enum.ts, enums.ts}` (research R1, R2): a `<span aria-hidden="true">` block with `--surface-skeleton`, radius from tokens; animation: transparent for `--motion-delay-skeleton`, then fade in and pulse (opacity) every `--motion-duration-skeleton-pulse`; under `prefers-reduced-motion: reduce` no fade and no pulse, shown after the same delay (a `steps()` step); T003 passes
- [X] T006 Add `isLabelHidden?: boolean` to `libs/components/button/button.tsx` (label in a `visually-hidden` span when set; icon stays visible); T004 passes

**Checkpoint**: `pnpm lint`, `pnpm build`, `pnpm test` exit 0.

---

## Phase 3: User Story 1 - The page looks like the reference (Priority: P1) 🎯 MVP

**Goal**: Every difference in the spec's Context table, except loading and the drawer, is gone.

**Independent Test**: Open `/assets/SS-001`, `/assets/TX-001-1`, `/assets/SWB-001-2` and `/assets?q=aurora&type=LV_BOARD` at 1920 × 940 next to the reference screenshots (quickstart §2 steps 1–3, 5); all tests of features 001–003 pass.

### Tests for User Story 1 ⚠️

- [X] T007 [P] [US1] Test `libs/components/logo/logo.test.tsx` (new): the mark is the `ZAP` icon inside the tile (decorative, `aria-hidden`), the wordmark is `APP_NAME`; with `isIconOnly` the tile has the accessible name "`APP_NAME` logo"
- [X] T008 [P] [US1] Update `pages/assets/components/asset-tree/asset-tree.test.tsx` (spec FR-007, FR-008, research R5): each node shows its name and a status shape for every status, in service included; the node's link contains the asset ID, the type in words and the status in words as visually hidden text and has `title` "`<Name> · <ID> · <Type> · <Status>`"; no visible status badge; the selected node has the `selected` class and a visually hidden "Selected"; keep every behaviour test (order, expansion, keyboard, `.`/`..` as text, scroll into view)
- [X] T009 [P] [US1] Test `pages/assets/components/register-summary/register-summary.test.tsx` (new, FR-009, R4): with 13 top-level items it shows "Register" and "13 substations"; with 1 "1 substation"; with 0 it renders nothing; it makes no request beyond `GET /api/assets` (shares the cache)
- [X] T010 [P] [US1] Test `pages/assets/components/asset-header/asset-header.test.tsx` (new, FR-010, R6): title "Asset explorer"; with a selected asset whose details are cached the line reads its path names joined by " / " (plain text, no links); with nothing selected, while loading, not found or failed it reads "Substations and the equipment installed in them."
- [X] T011 [P] [US1] Update `pages/assets/components/asset-details/asset-details.test.tsx` (FR-011–FR-015, R6, R7): path steps show names with the ID as hidden text and `title`, ancestors are links to `/assets/<id>`; the details card lists parent, voltage, rating, commissioned date, manufacturer, model, serial number and not asset ID/type/name/status; blanks read "Not recorded", dates `2010-03-12`; for a substation the order is details card → counts → "Immediate children (5)"; a transformer shows "Immediate children" with "Nothing is connected below this asset."
- [X] T012 [P] [US1] Update `pages/assets/components/asset-details/components/child-list/child-list.test.tsx` and `child-groups/child-groups.test.tsx` (FR-014, R8): a column heading row "Name / Asset ID / Status" is `aria-hidden`; rows keep one link each to `/assets/<id>` with name and ID tag, and the status badge; a child with ID `..` is plain text, not a link (spec edge case)
- [X] T013 [P] [US1] Update `pages/assets/components/asset-search/components/search-form/search-form.test.tsx` (FR-006, R9): the text field's accessible name is still "Search assets" but its label is visually hidden; placeholder "Search by ID or name"; the type select's label "Asset type" is visually hidden; Search and Clear behave as before
- [X] T014 [US1] Update `pages/assets/assets.test.tsx` for the page shell: the register card in the sidebar, the header title "Asset explorer", Import CSV still opens the panel; adjust only assertions that read the old title, the old placeholder or path steps by ID

### Implementation for User Story 1

- [X] T015 [P] [US1] `libs/components/logo/logo.tsx` + `styles.module.css`: render `<Icon iconName={IconName.ZAP}>` in an accent tile (`--surface-accent` background, `--text-on-accent` icon, `--size-brand-tile`, `--radius-md`) instead of `logo.svg`; keep the `isIconOnly` and `size` props (FR-004, R10); T007 passes
- [X] T016 [P] [US1] Tree node in `pages/assets/components/asset-tree/components/tree-node/tree-node.tsx` + `styles.module.css`: `StatusShape` (coloured by status) + name as the visible label; ID, type and status in words as `visually-hidden` spans inside the link, and `title` "`<Name> · <ID> · <Type> · <Status>`"; "Selected" visually hidden; selected row: `--surface-raised` background, a `--border-width-thick` accent bar on the left, bold name; long names truncate with an ellipsis; remove the now-unused `components/status-label` (FR-007, FR-008, R5); T008 passes
- [X] T017 [P] [US1] Create `pages/assets/components/register-summary/register-summary.tsx` + `styles.module.css` and the copy `REGISTER_LABEL`, `REGISTER_COUNT` (":count substations"), `REGISTER_COUNT_ONE` ("1 substation") in `pages/assets/libs/enums/assets-page-message.enum.ts`, with a `getRegisterCountText` helper in `pages/assets/libs/helpers/` (FR-009, R4); T009 passes
- [X] T018 [P] [US1] Create `pages/assets/components/asset-header/asset-header.tsx` + `styles.module.css` (reads `useGetAssetDetailQuery(selectedAssetId ?? skipToken)`'s `currentData`), change `AssetsPageMessage.TITLE` to "Asset explorer" and add `PATH_SEPARATOR` (" / ") (FR-010, R6); T010 passes
- [X] T019 [US1] Page shell in `pages/assets/assets.tsx` + `styles.module.css`: use `AssetHeader`; Import CSV gets the page class `import-button` (accent border and text, transparent background, accent tint on hover); the sidebar becomes a column whose tree/results region is the only scrolling part and whose bottom holds `RegisterSummary` (not rendered for an empty registry); off-canvas behaviour below 960 px unchanged (FR-005, FR-009); T014 passes
- [X] T020 [P] [US1] Path in `pages/assets/components/asset-details/components/asset-path/asset-path.tsx`: each step shows `assetName`, with the ID as hidden text and `title`; ancestors stay links, the current step `aria-current="page"`; truncate long names (FR-011)
- [X] T021 [US1] Details in `pages/assets/components/asset-details/asset-details.tsx`, `components/asset-attributes/asset-attributes.tsx` and their styles: order intro → details card → counts → children; attribute grid `repeat(auto-fill, minmax(var(--layout-attribute-min-width), 1fr))` without asset ID/type/name/status; parent first (not for a substation); the children card titled `CHILDREN_HEADING` ("Immediate children (:count)") for a substation; a card "Immediate children" with `CHILDREN_NONE` ("Nothing is connected below this asset.") for transformer, LV board and panel (FR-012, FR-013, FR-015, R7); T011 passes
- [X] T022 [US1] Children table look in `pages/assets/components/asset-details/components/child-list/child-list.tsx`, `child-groups/child-groups.tsx` and styles: an `aria-hidden` heading row with `COLUMN_NAME`, `COLUMN_ASSET_ID`, `COLUMN_STATUS`; rows as a 4-column grid (name, ID tag, status badge, chevron) aligned under it, stacked below 600 px; `.`/`..` as plain text through `checkIsAddressableAssetId`; group heading count muted (FR-014, R8, spec edge case "Narrow screens"); T012 passes
- [X] T023 [P] [US1] Compact search in `pages/assets/components/asset-search/components/search-form/search-form.tsx`, `components/asset-type-field/asset-type-field.tsx` and styles: `Input` with `isLabelHidden`, `SEARCH_FIELD_PLACEHOLDER` "Search by ID or name"; the select's label `visually-hidden`; Search and Clear as `ControlSize.SM` secondary buttons on one row (FR-006, R9); T013 passes
- [X] T024 [US1] Search results in `pages/assets/components/asset-search/components/search-results/`: keep feature 004's look; only make sure they sit in the sidebar's scrolling region above the register card (FR-009); no behaviour change
- [ ] T025 [US1] **Checkpoint — real stack**: rebuild (`DB_HOST_PORT=55434 API_HOST_PORT=18080 WEB_HOST_PORT=13000 docker compose up --build -d --wait`); headless screenshots of quickstart §2 steps 1–3 and 5 at 1920 × 940 and of step 1 at 390 px (off-canvas sidebar, register card at its bottom, stacked children) compared with the reference; `pnpm lint`, `pnpm build`, `pnpm test` exit 0; list the side-by-side check left for the user

**Checkpoint**: the page matches the reference apart from loading and the drawer.

---

## Phase 4: User Story 2 - Nothing jumps while data loads (Priority: P1)

**Goal**: Placeholders shaped like the content in every first-load area; no flash, no motion with reduced motion.

**Independent Test**: quickstart §3 (Slow 4G, no throttling, reduced motion, api stopped).

### Tests for User Story 2 ⚠️

- [X] T026 [P] [US2] Test `pages/assets/components/asset-tree/components/tree-skeleton/tree-skeleton.test.tsx` (new): a wrapper with `role="status"` holding the given text as visually hidden text and `rows` placeholder rows (dot + name line) at the given `level`'s indent, all `aria-hidden`
- [X] T027 [P] [US2] Update `pages/assets/components/asset-tree/asset-tree.test.tsx`: while the top level loads, a status "Loading assets…" with 8 placeholder rows and no tree; while a node's children load, a status "Loading…" with 3 placeholder rows inside that node's group; after a committed import the visible rows stay while the lists refetch (no placeholders); errors still show Try again
- [X] T028 [P] [US2] Test `pages/assets/components/asset-details/components/details-skeleton/details-skeleton.test.tsx` (new) and update `asset-details.test.tsx`: while the details load, a status "Loading asset…" with placeholders for the path, kicker, title, tag and badge, a details card of 2 rows × 5 and a children card of 3 rows; switching to a cached asset shows its details at once
- [X] T029 [P] [US2] Update `pages/assets/components/asset-search/components/search-results/search-results.test.tsx`: while searching, a status "Searching…" with 4 placeholder result rows; the refused, failed and no-match states unchanged
- [X] T030 [P] [US2] Extend `pages/assets/components/register-summary/register-summary.test.tsx`: while the top level loads, the label "Register" shows with a placeholder in place of the number
- [X] T031 [P] [US2] Extend `pages/assets/assets.test.tsx` (spec edge case "Failed reads", SC-004): when the top level, a node's children, the details or a search fails (500), the area shows its error with **Try again** and no placeholder (`aria-hidden` skeleton block) remains in it; the delay and reduced-motion behaviour are CSS-only and checked by hand (quickstart §3)

### Implementation for User Story 2

- [X] T032 [P] [US2] Create `pages/assets/components/asset-tree/components/tree-skeleton/tree-skeleton.tsx` + `styles.module.css` (props `label`, `level`, `rows`; constants `TREE_SKELETON_ROWS = 8`, `CHILDREN_SKELETON_ROWS = 3` in `pages/assets/libs/constants/`) and use it in `asset-tree.tsx` and `components/tree-children/tree-children.tsx` instead of the loading text (FR-019, R1, R3); T026, T027 pass
- [X] T033 [P] [US2] Create `pages/assets/components/asset-details/components/details-skeleton/details-skeleton.tsx` + styles (sizes mirror the intro, the details grid and the children rows) and use it in `asset-details.tsx` for the `!currentData` state instead of `DetailsMessage isLoading` (FR-019, FR-020); T028 passes
- [X] T034 [P] [US2] Create `pages/assets/components/asset-search/components/search-results/components/search-results-skeleton/search-results-skeleton.tsx` + styles (4 rows shaped like a result: ID tag, name line, meta line) and use it in `search-results.tsx` for the searching state (FR-019); T029 passes
- [X] T035 [P] [US2] Register number placeholder in `pages/assets/components/register-summary/register-summary.tsx` (`Skeleton` `TITLE`/`QUARTER` while `isLoading`) (FR-019); T030 passes
- [ ] T036 [US2] **Checkpoint — real stack**: rebuild; quickstart §3 by hand in the developer tools (Slow 4G, no throttling, reduced motion, api stopped) — run what can be run headless, list the rest for the user; `pnpm lint`, `pnpm build`, `pnpm test` exit 0

**Checkpoint**: loading never jumps and never flashes.

---

## Phase 5: User Story 3 - The import drawer guides through its steps (Priority: P2)

**Goal**: Subtitle, icon close, steps, a check placeholder and a footer that never covers a row.

**Independent Test**: quickstart §2 step 4; all import tests of feature 001 pass.

### Tests for User Story 3 ⚠️

- [X] T037 [P] [US3] Update `libs/components/panel/panel.test.tsx`: with `description` a paragraph under the title is the dialog's `aria-describedby`; the close button shows only its icon and is still named "Close"; Escape, focus trap and focus return unchanged
- [X] T038 [P] [US3] Test `pages/assets/components/import-panel/components/import-steps/import-steps.test.tsx` (new): an `<ol>` labelled "Import steps" with "1 Choose file", "2 Review", "3 Done"; the given step has `aria-current="step"`; earlier steps show a check and a visually hidden "(done)"; later steps neither
- [X] T039 [P] [US3] Test `pages/assets/components/import-panel/libs/helpers/get-import-step.helper.test.ts` (new): no result → `ImportStep.CHOOSE_FILE`; validated → `REVIEW`; confirmed (committed, nothing to commit, strict-blocked) → `DONE`
- [X] T040 [P] [US3] Update `pages/assets/components/import-panel/import-panel.test.tsx` and `components/import-result/import-result.test.tsx`: the steps follow check → Review and save → Done; Choose another file returns to Choose file; once a result shows, Save / Choose another file / Import another file and "Saving…" are inside a footer region outside the scrolling content (exactly one of each button); the outcome banner stays above the counts; while a check runs the form's existing status "Checking file…" and `aria-hidden` placeholders for the banner, 5 count cards and 3 table rows (one announcement, not two); a failed check or save shows its banner as before

### Implementation for User Story 3

- [X] T041 [US3] `libs/components/panel/panel.tsx` + `styles.module.css`: optional `description` (paragraph, `aria-describedby`), close button `isLabelHidden` with `IconName.CLOSE`, body as a column with `min-height: 0` so a child can own the scrolling (R11); pass `ImportPanelMessage.DESCRIPTION` ("Validate a file, review what would be rejected, then save the valid rows.") from `pages/assets/assets.tsx` (FR-016); T037 passes
- [X] T042 [P] [US3] Create `pages/assets/components/import-panel/libs/enums/import-step.enum.ts` (`CHOOSE_FILE`, `REVIEW`, `DONE`), `libs/helpers/get-import-step.helper.ts`, and `components/import-steps/import-steps.tsx` + styles (number in a circle, word, current step outlined in the accent colour, done steps with a check; wraps on narrow screens), copy `STEPS_LABEL`, `STEP_CHOOSE_FILE`, `STEP_REVIEW`, `STEP_DONE`, `STEP_DONE_MARKER` in `import-panel-message.enum.ts` (FR-017); T038, T039 pass
- [X] T043 [P] [US3] Create `pages/assets/components/import-panel/components/import-result-skeleton/import-result-skeleton.tsx` + styles (banner block, 5 count blocks in the counts grid, 3 table rows; status text `CHECKING_STATUS`) (FR-019)
- [X] T044 [US3] `pages/assets/components/import-panel/import-panel.tsx`, `components/import-result/import-result.tsx` and styles: render `ImportSteps` at the top; a scrolling content area (form, alerts, result, the check skeleton while `isChecking`) and, when a result is shown, a footer outside it with the actions `ImportResult` rendered before (moved, not duplicated: `ImportResult` gets the actions through a render split such as an `ImportResultActions` component) (FR-018, R11); T040 passes
- [ ] T045 [US3] **Checkpoint — real stack**: rebuild; quickstart §2 step 4 with the sample (Review, footer never over a row while scrolling) and a file that saves nothing (Done); `pnpm lint`, `pnpm build`, `pnpm test` exit 0; list the hand checks left for the user

**Checkpoint**: the drawer matches the reference without its overlap defect.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T046 [P] Update the UI contracts the feature changes: `specs/002-asset-explorer/contracts/ui-asset-explorer.md` (tree node content, path by name, header, details order — pointing to 005), `specs/003-asset-search/contracts/ui-asset-search.md` (placeholder "Search by ID or name", hidden labels, compact buttons, searching placeholders)
- [X] T047 [P] Update `docs/ARCHITECTURE.md` (frontend section: the shared `Skeleton` and the CSS delay, the register card from the cached top-level list, the drawer's footer outside the scroll area; trade-off rows for research R2, R5, R11) and `docs/KNOWN_LIMITATIONS.md` (the register card counts substations only, no import history; placeholders appear after 200 ms, so a slower first paint shows briefly empty space)
- [X] T048 [P] Update `frontend/README.md` (the asset page description: register card, header path, loading placeholders) and the root `README.md` feature list if it describes the page's look
- [X] T051 [P] Replace the template's browser tab icon (FR-004, research R10a): `frontend/public/favicon.svg` (32 × 32 tile `#0a0d12`, radius 6, the `zap.svg` bolt in `#e1ad66`) and `frontend/public/favicon.ico` (16, 32, 48 px of the same picture); `frontend/index.html` links both and sets `<meta name="theme-color" content="#0a0d12">`; `pnpm lint`, `pnpm build` exit 0
- [ ] T049 Accessibility pass (quickstart §4): check the accessibility tree of a tree node, a path step, the close button, a step and each loading area in the running app; fix anything found in the owning task's files
- [ ] T050 Final Definition of Done run: `pnpm lint`, `pnpm build`, `pnpm test`, `make check`, `make lint` exit 0; the full quickstart (§2 side by side, §3 loading, §4 accessibility) on the rebuilt stack; regression of features 001–003 quickstart UI scenarios that touch the page; `UNIT_TEST_REPORT.md` (root) regenerated with the new test counts; no secrets committed. Record every command with its exit code and list the hand checks left for the user

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: none. T001 before T005 (tokens); T002 before T015.
- **Foundational (Phase 2)**: T003 → T005, T004 → T006. Blocks every story.
- **US1 (Phase 3)**: needs Phase 2. Tests T007–T014 before T015–T024; T017, T018 before T019; T020 before T021; T025 last.
- **US2 (Phase 4)**: needs Phase 2 and the US1 components it fills (tree node T016, details T021, results T024, register T017). T026–T031 before T032–T035; T036 last.
- **US3 (Phase 5)**: needs Phase 2 (`Button` `isLabelHidden`, `Skeleton`); independent of US1/US2 otherwise. T037–T040 before T041–T044; T041 before T044; T045 last.
- **Polish (Phase 6)**: needs all stories; T050 last.

### User Story Dependencies

```text
Setup → Foundational ─┬→ US1 (reference look, MVP) → US2 (placeholders in the restyled areas)
                      └→ US3 (import drawer)
                                   US1 + US2 + US3 → Polish
```

### Within Each User Story

Tests first and failing → components → page wiring → checkpoint on the real stack.

### Parallel Opportunities

- T001 ∥ T002; T003 ∥ T004.
- US1 tests T007–T013 in parallel; implementations T015, T016, T017, T018, T020, T023 in parallel (separate files).
- US2 tests T026–T031 in parallel; T032–T035 in parallel.
- US3 can run beside US1/US2 after Phase 2 (different files except `assets.tsx`, touched by T019 and T041 — run those in sequence).
- Polish T046–T048 in parallel.

---

## Parallel Example: User Story 1

```text
Tests together:   T007 logo · T008 tree node · T009 register · T010 header · T011 details · T012 children · T013 search form
Then together:    T015 logo · T016 tree node · T017 register · T018 header · T020 path · T023 search form
Then in order:    T019 page shell → T021 details → T022 children → T024 results → T025 checkpoint
```

## Parallel Example: User Story 2

```text
Tests together:   T026 · T027 · T028 · T029 · T030 · T031
Then together:    T032 tree skeleton · T033 details skeleton · T034 results skeleton · T035 register number
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

Phases 1–3: the page looks like the reference. Stop at T025 and compare with the screenshots.

### Incremental Delivery

1. Setup + Foundational → shared skeleton and button prop ready.
2. US1 → the look (MVP).
3. US2 → placeholders in the restyled areas.
4. US3 → the drawer.
5. Polish → contracts, docs, accessibility, Definition of Done.

### Parallel Team Strategy

After Phase 2: one person on US1 then US2, another on US3; `assets.tsx` changes (T019, T041) coordinated.

---

## Notes

- UI checkpoint steps that need a person (side-by-side look, throttled loading, greyscale, screen reader) stay
  unchecked for the user, as in features 002 and 003.
- Copy changes are limited to data-model.md's "New copy" table; every other text stays.
- Commit after each phase or logical group; the user commits.
