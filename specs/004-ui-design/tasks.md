# Tasks: UI design refresh

**Input**: [spec.md](spec.md). Presentation only; every existing test must keep passing.

## Phase 1: Foundation

- [x] T001 Add `frontend/src/assets/css/app-tokens.css` (sidebar width, status colours, overlay shadow, z-indexes) and import it from `styles.css`
- [x] T002 Simplify `libs/components/app-shell` to a bare main area (the logo moves into the page sidebar) and remove its unused navigation constants, enums and types
- [x] T003 Add page-level presentational components `id-tag` and `status-badge` (shape + text) under `pages/assets/components/`

## Phase 2: Screens

- [x] T004 Asset page layout: sidebar (logo, Import CSV, search, "Substations" label, tree or results), sticky header, details area; off-canvas sidebar below 960 px (FR-001, FR-002)
- [x] T005 Details: intro (path, eyebrow, title, ID tag, status badge), cards for attributes, children and counts; child rows with ID tag, badge and chevron (FR-003, FR-004)
- [x] T006 Tree restyle with guide lines, selected treatment and shape-based status labels (FR-004, FR-005)
- [x] T007 Search form and results restyle in the sidebar (FR-006)
- [x] T008 Panel as a right-hand drawer; import form, outcome banner, counts and tables restyled (FR-007)
- [x] T009 Empty registry and not-found states (FR-008)

## Phase 3: Verification

- [x] T010 `pnpm lint`, `pnpm build`, `pnpm test` exit 0
- [ ] T011 Rebuild the stack and compare screenshots with the reference at 1440 px and 390 px; developer checks by hand in the browser
