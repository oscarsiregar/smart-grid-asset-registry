# Feature Specification: UI design refresh

**Feature Branch**: `004-ui-design`
**Created**: 2026-11-20
**Status**: Ready
**Input**: "Make the application look like the agreed reference design. Look only: no feature or
behaviour changes."

## Summary

Features 001–003 work, but the screens use a plain stacked layout. This feature restyles them to
the agreed reference design: a fixed sidebar for navigation, a sticky page header, card-based
details, status badges with shapes, and the import panel as a slide-over drawer. Every behaviour,
text, route, URL parameter and API call stays exactly as specified in features 001–003.

## Scope

### In scope (presentation only)

- **FR-001 Layout**: the asset page is a two-column layout. A fixed left sidebar holds the
  application logo, the "Import CSV" action, the search form and, below a "Substations" label, the
  tree (or the search results while a search is active, feature 003 FR-001). The main area has a
  sticky header with the page title and description, and the details area below it.
- **FR-002 Narrow screens**: below 960 px the sidebar becomes an off-canvas drawer opened from a
  menu button in the header and closed by a scrim, Escape, or selecting an asset. The details area
  takes the full width.
- **FR-003 Details**: the selected asset starts with its path, its type as a small eyebrow, its
  name as a large title, and its asset ID and status as a tag and a badge. Attributes, children
  and descendant counts sit in separate cards. Attributes form a responsive grid of label/value
  pairs. Children are table-like rows: name, asset ID tag, status badge and a chevron; the whole
  row is clickable through its single link. Descendant counts are stat cards.
- **FR-004 Status badges**: wherever a status is shown in words (details, child rows, search
  results), it is a badge whose colour is always paired with a shape (filled, half-filled, open
  ring, dashed ring for "No status recorded"), so status is never conveyed by colour alone
  (constitution principle V). The tree keeps feature 002's rule (status only when not in service)
  and uses the same shapes.
- **FR-005 Tree**: compact rows with a chevron toggle, child levels hanging off a guide line, and
  the selected row marked by a raised background, an accent bar, bold text and the existing
  "Selected" marker.
- **FR-006 Search**: the search form is stacked in the sidebar (text field with a search icon, the
  type filter, Search and Clear); results are compact rows with an ID tag, name, type, status
  badge and location.
- **FR-007 Import panel**: the import panel is a drawer that slides in from the right edge, full
  height, with a header (title and close button) and a scrollable body. The file field, mode
  options, outcome banner, count cards and the rejected-row and warning tables use the reference
  styles (form-control styling, option cards, toned alerts, stat cards, tables with sticky
  headers). Focus trapping, Escape and focus return are unchanged.
- **FR-008 Empty and not-found states**: the empty registry and the not-found message are centred
  states inside a card.

### Out of scope

New elements from the reference that would add features: a single-line diagram, a register
summary in the sidebar, numbered import steps, import history in the panel, error-code filter
chips, and a dropdown search. All behaviour of features 001–003.

## Success Criteria

- **SC-001**: Side by side with the reference screenshots, each screen has the same layout and
  visual hierarchy (desktop 1440 px and mobile 390 px).
- **SC-002**: Every existing automated test passes unchanged in behaviour; `pnpm lint`,
  `pnpm build` and `pnpm test` exit 0.
- **SC-003**: No colour, spacing or font is hard-coded outside the design tokens; new tokens live in
  `frontend/src/assets/css/app-tokens.css`.

## Constitution Check

- Principle I: template components (Button, Icon, Link, Panel, Logo) are reused and restyled; no
  new pattern beyond page-level presentational components (ID tag, status badge).
- Principle V: status never by colour alone; panels keep focus trap and Escape.
- No new template exception.
