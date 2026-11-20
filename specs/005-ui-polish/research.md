# Research: UI polish to the reference design, with loading placeholders

Decisions for [spec.md](spec.md). Each entry: decision, rationale, alternatives. The base is the code after
feature 004 (`frontend/src/pages/assets`, restyled), the design tokens in `frontend/src/assets/css`
(`tokens.css` from the template, `app-tokens.css` for this app) and `frontend/docs/code-conventions.md`.
The backend is not touched.

## R0. Template fit — no new exception

**Decision**: Everything is presentation inside the frontend template's structure. One new shared component
(`libs/components/skeleton`) and two small, backwards-compatible props on template components (`Button`
`isLabelHidden`, `Panel` `description`) are the only additions outside `pages/assets`. No new dependency, no
new pattern, no new template exception.

**Rationale**: Principle I allows extending the template's components in their own shape; a generic
placeholder with no feature knowledge belongs in `libs/components` (conventions §2), because three page areas
and the import panel use it. Optional props keep every existing call site valid.

**Alternatives**: a skeleton library (a dependency the stack does not need: CSS covers it, conventions §1);
copying the reference's own components (another codebase with another stack — TanStack Query, its own
`drawer`, `loading-state` — its shapes do not match this template).

## R1. One skeleton component, page-level compositions

**Decision**: `libs/components/skeleton/skeleton.tsx` renders one placeholder block
`<span aria-hidden="true">` with two props from enums: `shape` (`TEXT` — one line of body text, `TITLE` — a
title line, `BLOCK` — a card-sized area, `DOT` — a status shape) and `width` (`FULL`, `THREE_QUARTERS`,
`HALF`, `QUARTER`). Sizes come from tokens; no inline styles (conventions §7). Each loading area composes
these blocks in a small page component — `TreeSkeleton` (rows at a given level), `DetailsSkeleton` (intro,
details card, children card), `SearchResultsSkeleton` (result rows), `ImportResultSkeleton` (banner, count
cards, table rows) — inside a wrapper with `role="status"` that keeps the area's existing loading text as
visually hidden text (FR-022).

**Rationale**: One block with a few named shapes is enough for every area and keeps the shapes consistent;
the compositions live next to the content they imitate, so their size follows it (FR-020). Keeping the
existing loading texts means screen-reader users hear the same as before, and the existing tests that look
for "Loading assets…", "Loading asset…", "Loading…" and "Searching…" still find them.

**Alternatives**: a generic "skeleton list" with a row count (cannot imitate the details or the import
result); replacing the `Loader` spinner (a spinner still jumps; the template `Loader` stays for buttons).

## R2. No flash: a CSS delay, not a timer

**Decision**: A placeholder appears with a CSS animation that keeps it transparent for
`--motion-delay-skeleton` (200 ms, a new token) and then fades in; afterwards it pulses with an opacity
animation of 1.2 s. Under `prefers-reduced-motion: reduce` the fade and the pulse are off and the placeholder
shows, static, after the same delay (a step change, not motion).

**Rationale**: A response under 200 ms replaces the transparent placeholder before anyone sees it (FR-023,
SC-004) while the placeholder already holds the space, so nothing jumps either way (FR-020). It needs no
`setTimeout`, no state and no cleanup, and tests are unaffected by timing.

**Alternatives**: a `useDelayedFlag` hook with a timer (state, cleanup and fake timers in tests for the same
effect); no delay (flashes on a fast local API).

## R3. When an area shows placeholders

**Decision**: Placeholders replace the existing first-load conditions only: the tree and a node's children
on `isLoading` (RTK Query's first request), the details when `currentData` is undefined and there is no error
(as today), the search results when `currentData` is undefined and there is no error, the register number on
the top-level query's `isLoading`, and the import result while the check mutation runs. Refetches keep their
data on screen, as they do now.

**Rationale**: These are exactly the places that render a loading text today, so the behaviour of features
002 and 003 is unchanged (FR-001); cached data shows at once and a refetch after an import never blinks
(spec edge cases). A save keeps its result on screen because the result is already there (spec US3-5).

**Alternatives**: using `isFetching` (would blank lists on every refetch after an import).

## R4. Register card

**Decision**: A `RegisterSummary` page component at the bottom of the sidebar reads the top-level list with
`useGetAssetNodesQuery(null)` — the same cached query the tree and the empty-registry check already use, so
no extra request — and shows "Register" and ":count substations" ("1 substation" for one). It is not rendered
for an empty registry. Its number shows a placeholder while the query first loads.

**Rationale**: FR-009 without new data: the top-level list is every substation and `totalItems` is its
length (Exception 13).

**Alternatives**: a count endpoint or the "last import" line (out of scope, agreed with the user).

## R5. Tree node: shape and name, details for assistive technology

**Decision**: A node's visible label is the `StatusShape` (status colour) and the asset name. Inside the same
link, visually hidden text carries the asset ID, the type and the status in words, and the link's `title`
gives the name and the same details as a hover text ("Aurora Substation (North) · SS-001 · Substation ·
In service"), so a truncated name can still be read. The "Selected" marker becomes visually
hidden; the selected row keeps the raised background, an accent bar and bold text. The status-in-words badge
(`StatusLabel`) leaves the tree.

**Rationale**: Clarification Q1 (amends 002 FR-002). The shape is not colour alone (principle V); the words
stay in the accessible name, so the tree reads as before. Feature 002's tests that look for the ID, type or
status text still find it in the DOM.

**Alternatives**: `aria-label` on the link (would override and duplicate the visible name and is easy to let
drift); a tooltip component (no such component in the template; `title` is enough for hover).

## R6. Path by name; the header breadcrumb

**Decision**: `AssetPath` shows each step's asset name; the asset ID follows as visually hidden text and is
the step's `title`. The page header reads "Asset explorer" (`AssetsPageMessage.TITLE`) and, below it, the
selected asset's path names joined by " / " as plain text, read from the same cached
`useGetAssetDetailQuery(selectedAssetId)` that the details use (no extra request); with nothing selected, or
while the details load or fail, it shows the existing description.

**Rationale**: Clarification Q2 (amends 002 FR-011). The selectable path stays the one in the details
(feature 002); the header line is orientation only, so it is not a second navigation landmark.

**Alternatives**: links in the header too (two identical navigations, one of them hidden behind the sticky
header on scroll).

## R7. Details layout

**Decision**: Order intro → details card → descendant counts (substation) → children card. The details card
drops asset ID, type, name and status (the intro shows each, so 002 FR-008 still holds) and lists parent,
voltage, rating, commissioned date, manufacturer, model and serial number in a grid of up to five columns
(`repeat(auto-fill, minmax(…))` with a token width). Blank and number/date texts stay as 002 FR-009/FR-010
(clarification Q3). A transformer, LV board or panel gets an "Immediate children" card with "Nothing is
connected below this asset." (FR-015); a substation's card title becomes "Immediate children (N)".

**Rationale**: Matches the reference while every attribute stays visible once.

## R8. Children as a table look, list semantics

**Decision**: Keep the existing list (`ul` / `li`, one stretched link per row) and add a column-heading row
"Name / Asset ID / Status" marked `aria-hidden="true"`; rows become a three-column grid aligned under it.
Group headings keep their text (":label (:count)") with the count styled muted.

**Rationale**: Each row already reads completely ("Aurora Transformer T1, TX-001-1, In service"); a real
`table` would add little for assistive technology but break the one-link-per-row pattern and feature 002's
tests that use list items. The visual headings are decoration over that.

**Alternatives**: a `table` with a link in the name cell (row-wide click needs the same stretched link, and
more test churn for no user gain).

## R9. Search form

**Decision**: The text field uses the restored template `Input` with `isLabelHidden` and the placeholder
"Search by ID or name"; the type `<select>` keeps its `<label>` as `visually-hidden`. **Search** and **Clear**
become small secondary buttons on one row. The search still runs on Enter or **Search** (feature 003).

**Rationale**: FR-006. The label stays for assistive technology (conventions §6.3 allow `.visually-hidden`).
The new placeholder replaces feature 003's hint text, so the 003 UI contract's copy is updated with it.

## R10. Brand and Import CSV

**Decision**: The template `Logo` keeps its API; its mark becomes a bolt (`zap`) in an accent tile: a new
`zap.svg` is added to `assets/img`, `IconName.ZAP` maps it, and `Logo` renders the icon inside a tile styled
with tokens instead of the template's `logo.svg`. The wordmark stays `APP_NAME`. **Import CSV** gets a page
class (accent border and text, transparent background) rather than a new button variant.

**Rationale**: FR-004/FR-005 with the smallest change; one page uses the outlined look, so a page class is
enough (conventions: no new pattern for a single use).

## R11. Import drawer: description, icon close, steps, footer

**Decision**:
- `Panel` gets an optional `description` (a subtitle under the title, also its `aria-describedby`); its close
  button becomes icon-only through a new `Button` prop `isLabelHidden` (the label stays as visually hidden
  text, so the accessible name is still "Close").
- `ImportSteps` (import-panel component): an `<ol aria-label="Import steps">` of three items, number and
  word in each, `aria-current="step"` on the current one and a check icon plus visually hidden "(done)" on
  earlier ones. The step is derived from the panel's state: a confirmed result → Done, a validated result →
  Review, otherwise Choose file. It is a display, not navigation.
- The panel body stops scrolling itself: `Panel`'s body becomes a column with `min-height: 0`, and
  `ImportPanel` lays out a scrolling content area and, once a result is shown, a footer outside it with the
  actions that `ImportResult` renders today. The outcome banner stays at the top of the result.

**Rationale**: FR-016–FR-018. A footer outside the scrolling area can never cover a row (the defect seen in
the reference). The actions move, they are not duplicated, so feature 001's tests keep finding one Save and
one Choose-another-file button.

**Alternatives**: a sticky footer inside the scrolling body (covers rows while scrolling — the reference's
defect); a `footer` slot on `Panel` (the footer depends on the import panel's mutation state, which lives
inside `ImportPanel`).

## R12. Tests

**Decision**: Behaviour tests of features 001–003 stay; assertions that read changed visible copy or
structure are updated in the same task as the change (path steps by name, the header title, the search
placeholder, the details card's attribute set, the panel description). New tests cover: each skeleton
wrapper (status role, hidden shapes, existing text), the register count and its pluralisation, the step
derivation, the footer actions, the tree node's hidden ID/type/status, the path's hidden ID, and `Button`
`isLabelHidden`. Tests first, as in every feature.
