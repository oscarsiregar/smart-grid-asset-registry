# Feature Specification: UI polish to the reference design, with loading placeholders

**Feature Branch**: `005-ui-polish`
**Created**: 2026-11-20
**Status**: Ready
**Input**: User description: "Make the design match the reference screenshots ('Grid Assets'); keep our
own process. I like the colours of the grid and so on. Add some loading, because without it the page
appears in jerks." Scope agreed with the user: presentation only (no new endpoint, no import history),
loading shown as placeholders shaped like the content.

## Clarifications

### Session 2026-11-20

- Q: Tree node content — keep feature 002's asset ID, name, type and status words, or follow the reference
  (status shape and name)? → A: Follow the reference: the status shape and the name are shown; the asset
  ID, type and status words stay for assistive technology and on hover. Amends feature 002 FR-002.
- Q: Path steps (breadcrumb and details intro) by asset ID, by name, or both? → A: By name; the asset ID
  stays for assistive technology and on hover. Amends feature 002 FR-011.
- Q: Blank values and dates — keep feature 002's "Not recorded" and 2010-03-12, or the reference's "—"
  and "Mar 12, 2010"? → A: Keep feature 002's texts.

## Context

Feature 004 (ui-design) already moved the asset page to the reference layout: a fixed sidebar with the
tree, a sticky header, details in cards, status badges with shapes, and the import panel as a drawer
from the right. Colours, fonts and design tokens are already identical to the reference.

A side-by-side comparison of the running application with the reference screenshots (2026-11-20,
1920 × 940) still shows these differences, which this feature closes:

| Area | Now | Reference |
|---|---|---|
| Brand | template "P" mark, "Smart Grid Asset Registry" | bolt mark in an accent tile |
| Import CSV | neutral filled button | outlined button in the accent colour |
| Search form | visible labels "Search assets" and "Asset type", a full-width accent **Search** button | one compact field "Search by ID or name" with an icon, the type select below it, no visible labels, no large button |
| Tree node | asset ID, name, a status line under nodes not in service, a "Selected" line | a status shape before every name, name only, the selected row marked by a raised row, an accent bar and bold text |
| Sidebar bottom | nothing | a "Register" card with the number of substations |
| Header | "Assets" + fixed description | "Asset explorer" + the selected asset's path as a breadcrumb |
| Details intro | path of asset IDs above the type | path of names (link per ancestor), then type, name, ID tag and status badge |
| Details card | 4-column grid that repeats ID, type, name and status | 5-column grid of the remaining attributes: parent first, then voltage, rating, commissioned, manufacturer, model, serial number |
| Substation | "Beneath this substation" counts first | details first, then the children; counts after |
| Children | rows with name and ID side by side, no column headings | a table with headings Name / Asset ID / Status, group headings with a count ("Transformers 2") |
| Leaf asset | no children card | an "Immediate children" card saying nothing is connected below |
| Import drawer | title only, text **Close** button | title with a one-line subtitle, an icon close button, a 1-2-3 step indicator (Choose file, Review, Done), and a footer that stays visible with the next action |
| Loading | text such as "Loading assets…" appears and is replaced, so the layout jumps | — (requested by the user) |

## User Scenarios & Testing *(mandatory)*

### User Story 1 - The page looks like the reference (Priority: P1)

An operator opens the asset page and sees the sidebar, header, tree and details laid out and styled like
the reference: the brand, the outlined Import CSV button, the compact search, the tree with a status
shape per node, the register card, the breadcrumb header, the details card and the children table.

**Why this priority**: It is the user's main request; every other story builds on the restyled page.

**Independent Test**: Open `/assets/SS-001`, `/assets/TX-001-1` and `/assets/SWB-001-2` at 1920 × 940 next
to the reference screenshots; each difference in the Context table is gone, and every behaviour test of
features 001–003 still passes.

**Acceptance Scenarios**:

1. **Given** assets are stored, **When** the operator opens `/assets/SS-001`, **Then** the header reads
   "Asset explorer" with the path "Aurora Substation (North)", the details show the type, name, ID tag and
   status badge, a details card without the repeated ID/type/name/status, the counts, and the children
   as a table with column headings and group headings with their counts.
2. **Given** a transformer without children is selected, **When** its details show, **Then** an
   "Immediate children" card says nothing is connected below it.
3. **Given** the tree is shown, **When** the operator looks at any node, **Then** it shows a status shape
   before the name, and the selected node is marked by a raised row, an accent bar and bold text.
4. **Given** 13 substations are stored, **When** the page loads, **Then** the register card at the bottom
   of the sidebar says "13 substations".

---

### User Story 2 - Nothing jumps while data loads (Priority: P1)

While the tree, a node's children, the details or the search results are being read, the operator sees
grey placeholders shaped like the content that is coming, and the content replaces them in place.

**Why this priority**: The user named it explicitly: without it the page appears "in jerks".

**Independent Test**: Slow the network in the browser and open `/assets/SS-001`, expand a switchboard, and
run a search; each area shows placeholders of the same size as its content, and nothing below it moves
when the data arrives.

**Acceptance Scenarios**:

1. **Given** the top level is being read, **When** the page opens, **Then** the tree area shows placeholder
   rows the height of tree rows, and the register card shows a placeholder for its number.
2. **Given** a node is expanded for the first time, **When** its children are being read, **Then**
   placeholder rows appear under it at the child indent.
3. **Given** an asset is selected, **When** its details are being read, **Then** the details area shows
   placeholders for the intro, the details card and the children card.
4. **Given** a search was submitted, **When** the results are being read, **Then** the results area shows
   placeholder result rows.
5. **Given** the operator's system asks for reduced motion, **When** placeholders are shown, **Then** they
   do not pulse.
6. **Given** any placeholder is shown, **When** a screen reader reads the area, **Then** it announces the
   existing loading text of that area ("Loading assets…", "Loading…", "Loading asset…", "Searching…").

---

### User Story 3 - The import drawer guides through its steps (Priority: P2)

The operator opens Import CSV and sees which step they are on, works through choosing, checking and saving
a file with the same buttons and rules as today, and always sees the next action at the bottom of the
drawer.

**Why this priority**: It completes the reference look of the drawer; the import itself already works.

**Independent Test**: Open `?panel=import`, check the sample file, save it, and check a file that saves
nothing; the step indicator and the footer follow each step, and the import tests of feature 001 still
pass.

**Acceptance Scenarios**:

1. **Given** the drawer opens, **Then** it shows the title, the subtitle, an icon close button and the
   steps "Choose file", "Review", "Done" with "Choose file" as the current step.
2. **Given** a file was checked, **When** the result shows, **Then** "Review" is the current step and
   "Choose file" is marked done, the outcome stays in its banner at the top of the result, and a footer
   stays visible at the bottom of the drawer with the next actions (**Save N assets** when assets can be
   saved, and **Choose another file**).
3. **Given** a file was saved, **Then** "Done" is the current step and the footer offers **Import another
   file**.
4. **Given** the rejected rows are long, **When** the operator scrolls the drawer body, **Then** the footer
   never covers a table row (the last row can always be scrolled above it).
5. **Given** a check is running, **Then** the result area below the form shows placeholders for the
   outcome banner, the count cards and the first table rows; **Given** a save is running, **Then** the
   result stays on screen and the buttons show they are busy, as today.

### Edge Cases

- **Empty registry**: the empty state of feature 002 stays as it is (no search, no tree); the register card
  is not shown, since there is nothing to count.
- **Fast responses**: a placeholder that would show for less than a moment must not flash; data that is
  already cached shows at once without placeholders.
- **Failed reads**: placeholders are replaced by the existing error state with **Try again**; they never
  stay on screen after a failure.
- **Refetch after an import**: lists that refetch after a committed import keep showing their current
  rows rather than switching back to placeholders.
- **Same name twice**: two assets may share a name; the tree and the path stay unambiguous through their
  position and the asset ID in the accessible name and hover text.
- **Long names**: tree rows and breadcrumb steps truncate with an ellipsis; the full name stays available
  to assistive technology and on hover.
- **Narrow screens**: the off-canvas sidebar of feature 004 keeps working; the register card stays at its
  bottom, the children table drops to stacked rows, the step indicator wraps.
- **IDs `.` and `..`**: still plain text, not links, in the tree, breadcrumb and children table.

## Requirements *(mandatory)*

### Functional Requirements

**General**

- **FR-001**: Every behaviour, address, URL parameter, API call and validation rule of features 001–003
  MUST stay as it is; this feature changes presentation and adds loading placeholders only.
- **FR-002**: Colours, spacing, radii and fonts MUST come from the existing design tokens; new values are
  added as tokens.
- **FR-003**: Every state MUST stay readable without colour (constitution principle V): status by shape and
  words (visible or announced), the selected node by more than colour, steps by number and words.

**Sidebar**

- **FR-004**: The brand MUST use a bolt mark in an accent tile with the application name "Smart Grid Asset
  Registry".
- **FR-005**: **Import CSV** MUST be an outlined button in the accent colour, full width.
- **FR-006**: The search form MUST show one text field with a search icon and the placeholder "Search by ID
  or name", and the type select below it; their labels ("Search assets", "Asset type") stay for assistive
  technology but are not shown. The search still runs on Enter or a **Search** button (feature 003); the
  button and **Clear** are compact and sit on one row below the select.
- **FR-007**: Each tree node MUST show a status shape before the asset name — filled dot in service, half
  dot maintenance, ring out of service, dashed ring no status — and the name; the asset ID, the type and
  the status in words are not shown but MUST be part of the node's accessible name and of its hover text
  (amends feature 002 FR-002).
- **FR-008**: The selected node MUST be marked by a raised row, an accent bar at its left and bold text;
  the word "Selected" stays for screen readers.
- **FR-009**: The tree (or the search results) MUST scroll inside the sidebar, and a "Register" card MUST
  stay at the bottom of the sidebar with the number of stored substations ("13 substations", "1
  substation"), taken from the top-level list the tree already reads.

**Main area**

- **FR-010**: The header MUST read "Asset explorer" and, below it, the selected asset's path names as one
  line of text (orientation only; the selectable path is the one in the details), or the description of
  feature 002 when nothing is selected.
- **FR-011**: The details intro MUST show the path (ancestors selectable, as feature 002 FR-011), the type
  as a small accent kicker, the name as a large title, the asset ID as a tag and the status as a badge.
  Path steps in the details MUST show the asset name; each step's asset ID MUST be part of its accessible
  name and of its hover text (amends feature 002 FR-011). The header line of FR-010 shows the same names.
- **FR-012**: The details card MUST list, in this order, parent asset (link, for every type but
  substation), voltage, rating, commissioned date, manufacturer, model and serial number, in up to five
  columns; asset ID, type, name and status are not repeated there because the intro shows them. Blank
  values and number and date formats stay as feature 002 FR-009 and FR-010 require ("Not recorded",
  "22.5 kV", 2010-03-12).
- **FR-013**: For a substation the order MUST be details card, descendant counts, children; for a switchboard
  details card, then its panels.
- **FR-014**: Children MUST be laid out as a table under the column headings Name, Asset ID and Status,
  grouped by type with the group's name and count as a group heading; each row opens that asset (one link
  per row; IDs `.` and `..` plain text).
- **FR-015**: An asset without children MUST show an "Immediate children" card that says nothing is
  connected below it.

**Import drawer**

- **FR-016**: The drawer header MUST show the title, the subtitle "Validate a file, review what would be
  rejected, then save the valid rows." and a close button with an icon and an accessible name "Close".
- **FR-017**: A step indicator MUST show "1 Choose file", "2 Review", "3 Done": Choose file while no result
  is shown, Review after a check, Done after a save that committed or that saved nothing; the current step
  is marked by more than colour and announced as current, earlier steps are marked done. The steps are a
  display of the existing flow, not a new navigation.
- **FR-018**: Once a result is shown, the actions feature 001 already offers (**Save N assets**, **Choose
  another file** / **Import another file**, and the "Saving…" status) MUST sit in a footer that stays
  visible at the bottom of the drawer, outside the scrolling content, so it never covers a row. The outcome
  stays stated once, in its banner (feature 001 FR-027).

**Loading**

- **FR-019**: While its data is read for the first time, each of these areas MUST show placeholders shaped
  like its content instead of a single loading line: the tree (rows), a node's children (rows at the child
  indent), the details (intro, details card, children card), the search results (result rows), the register
  card (its number), and the import result while a check runs (banner, count cards, first table rows).
- **FR-020**: Placeholders MUST keep the size of the content they stand for, so nothing below them moves
  when the content arrives.
- **FR-021**: Placeholders MUST pulse gently, and MUST NOT move when the system asks for reduced motion.
- **FR-022**: Each placeholder area MUST keep the existing loading text for assistive technology (as a
  status message) and hide the placeholder shapes from it.
- **FR-023**: Placeholders MUST NOT appear for data already in the cache or for a refetch of data already
  shown, and MUST NOT appear for a response faster than a short delay, so they never flash.

### Key Entities

No new data. The register card counts the top-level substations already read for the tree.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Side by side with the reference screenshots at 1920 × 940, every difference listed in the
  Context table is gone (checked by hand on the substation, transformer, switchboard, search and import
  screens).
- **SC-002**: With the network slowed to "Slow 4G", opening an asset, expanding a node and searching show no
  visible jump of content below the loading area (checked by hand).
- **SC-003**: Every existing automated test of features 001–003 passes with its behaviour unchanged;
  `pnpm lint`, `pnpm build` and `pnpm test` exit 0.
- **SC-004**: No placeholder is visible for a response under the flash delay, and none stays after a
  failure.
- **SC-005**: Every state can be told apart with colour removed (greyscale check by hand), as for features
  002 and 003.

## Assumptions

- The application keeps its own name; only the mark changes to the reference's bolt.
- The reference's "Recent imports", "last import" line, error-code filter chips and single-line diagram stay
  out of scope (agreed with the user: presentation only, no new endpoint).
- The **Search** button stays (feature 003 runs a search on Enter or the button) but becomes compact; the
  reference's lack of a button is not copied.
- The flash delay for placeholders is short (in the order of a few hundred milliseconds); the plan fixes it.
- The reference's visual artefacts the user's review noticed (a footer over the table, a native file-input
  tooltip over the hint) are not copied: FR-018 forbids the first, and the file field keeps its hint visible.
