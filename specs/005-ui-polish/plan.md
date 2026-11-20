# Implementation Plan: UI polish to the reference design, with loading placeholders

**Branch**: `005-ui-polish` | **Date**: 2026-11-20 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/005-ui-polish/spec.md`

## Summary

Close the remaining visual differences between the asset page and the reference screenshots (brand, Import
CSV, compact search, tree with a status shape per node, register card, "Asset explorer" header with the path,
details card, children table, import drawer with steps and a fixed footer) and add placeholders shaped like
the content wherever data loads for the first time. Presentation only: no backend change, no new request,
no change to addresses, rules or texts beyond the copy the spec names. Approach in [research.md](research.md):
one shared `Skeleton` block with page-level compositions and a CSS delay against flashing (R1–R3), the
register count from the cached top-level list (R4), visually hidden ID/type/status for the shortened tree and
path (R5–R6), and a non-scrolling footer in the import drawer (R11).

## Technical Context

**Language/Version**: TypeScript 6, React 19 (frontend only)

**Primary Dependencies**: Vite 8, RTK Query, React Router 7, React Hook Form + Zod, CSS Modules with the
template's design tokens; no new dependency

**Storage**: N/A (no data change)

**Testing**: Vitest + React Testing Library, `fetch` mocked (Template Exception 2); `pnpm lint`, `pnpm build`,
`pnpm test`

**Target Platform**: Current desktop and mobile browsers; reference viewport 1920 × 940, narrow 390 px

**Project Type**: Web application (frontend part only)

**Performance Goals**: No extra request; placeholders invisible for responses under 200 ms

**Constraints**: Code conventions (one declaration per file, tokens only, no inline styles, no magic
numbers, `px` units, no nesting); principle V (status and selection never by colour alone; loading, empty
and error states handled); behaviour and tests of features 001–003 unchanged

**Scale/Scope**: One page (`pages/assets`), the import panel, three template components touched
(`Logo`, `Button`, `Panel`) plus one icon added to `Icon`'s map, one new shared component (`Skeleton`)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Result |
|---|---|---|
| I. Follow the Template | New code copies existing shapes; `Skeleton` is a generic `libs/components` unit; `Button`/`Panel` get optional props only; no dependency (R0) | Pass |
| II. Validation in the backend | No validation touched; the search keeps sending text to the backend | Pass |
| III. Predictable imports | Outcome still stated in words in its banner; actions only move to a footer (R11) | Pass |
| IV. Test the rules | No rule changes; frontend tests first for every new unit (R12) | Pass |
| V. Usable and accessible | Addresses unchanged; loading states improved and still announced (R1); status by shape, selection by bar + bold (R5); panel keeps focus trap and Escape | Pass |
| Template Exceptions | None needed (R0) | Pass |

**Post-design re-check (after Phase 1)**: unchanged — Pass. The UI contract adds no route, parameter or
request.

## Project Structure

### Documentation (this feature)

```text
specs/005-ui-polish/
├── plan.md              # This file
├── research.md          # Decisions R0–R12
├── data-model.md        # No data; the view states and their sources
├── quickstart.md        # Hand checks and commands
├── contracts/
│   └── ui-asset-page.md # The page and drawer after this feature (delta to 002/003/004)
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root)

```text
frontend/index.html                        # favicon links + theme-color (R10a)
frontend/public/favicon.svg, favicon.ico   # dark tile with the amber bolt (R10a)
frontend/src/
├── assets/
│   ├── css/app-tokens.css                 # + skeleton, register and drawer tokens
│   └── img/zap.svg                        # new bolt mark
├── libs/
│   ├── components/
│   │   ├── skeleton/                      # NEW: skeleton.tsx, styles.module.css, libs/enums (shape, width), test
│   │   ├── button/button.tsx              # + isLabelHidden
│   │   ├── panel/panel.tsx                # + description; icon-only close; body as a column
│   │   ├── logo/logo.tsx                  # bolt mark in an accent tile
│   │   └── icon/libs/…                    # + IconName.ZAP
│   └── enums/                             # IconName export
└── pages/assets/
    ├── assets.tsx, styles.module.css      # header title + breadcrumb, sidebar regions, register card
    ├── components/
    │   ├── register-summary/              # NEW (R4)
    │   ├── asset-tree/                    # tree-node: shape + name; tree/tree-children: TreeSkeleton
    │   │   └── components/tree-skeleton/  # NEW
    │   ├── asset-search/                  # search-form compact; search-results: SearchResultsSkeleton
    │   │   └── components/search-results/components/search-results-skeleton/   # NEW
    │   ├── asset-details/                 # order, details card set, children table look, leaf card
    │   │   └── components/details-skeleton/                                    # NEW
    │   ├── asset-header/                  # NEW: title + path line (R6)
    │   └── import-panel/
    │       ├── import-panel.tsx           # steps, scrolling content + footer
    │       └── components/{import-steps,import-result-skeleton}/              # NEW
    └── libs/{enums,constants,helpers}/    # copy, step enum, register/step helpers
```

**Structure Decision**: Frontend only, inside the existing `pages/assets` page and `libs/components`, as
feature 004 did. No backend, no migration, no API document change.

## Complexity Tracking

No constitution violations; nothing to justify.
