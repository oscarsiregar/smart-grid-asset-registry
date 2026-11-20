# Feature Specification: Asset Explorer

**Feature Branch**: `002-asset-explorer`

**Created**: 2026-11-19

**Status**: Draft

**Input**: User description: "Feature short name: asset-explorer. Browse the stored asset hierarchy. After
importing, an operator needs to explore the stored asset hierarchy: see which substations exist, drill down into
the equipment installed at each, and look at the details of any asset. Display the stored hierarchy as an
expandable tree with a selected-asset details area. Selecting a substation shows its details, its immediate
children grouped by type, and counts of the transformers, LV boards, switchboards and switchboard panels beneath
it. Selecting any other asset shows its details as well. Each asset has a stable address (for example
/assets/<asset_id>); opening that address directly or refreshing the page keeps the selected asset, with its
ancestors expanded in the tree so it is visible and highlighted. Provide appropriate loading, empty and error
states. An unknown asset ID shows a clear 'not found' state rather than an error. Priority order: correctness,
then readability, then optimisation. Assets are only created by imports (feature 001); this feature only reads
them. The import panel stays available on the same page, and after a committed import the tree shows the new
assets. Out of scope: search, editing or deleting assets, user accounts."

## Context

Feature 001 ([specs/001-csv-import/spec.md](../001-csv-import/spec.md)) defines the asset data this feature
displays:

- **Asset types and permitted parents**: SUBSTATION has no parent; TRANSFORMER, LV_BOARD and SWITCHBOARD sit
  under a SUBSTATION; SWITCHBOARD_PANEL sits under a SWITCHBOARD. The stored hierarchy is therefore at most three
  levels deep: substation → transformer / LV board / switchboard → switchboard panel.
- **Operational statuses**: IN_SERVICE, MAINTENANCE, OUT_OF_SERVICE, or none recorded.
- **Asset attributes**: asset ID, type, name, operational status, voltage (kV), rating (kVA), manufacturer,
  model, serial number, commissioned date, and parent asset. Optional attributes may be blank.

Imports only ever add assets that satisfy these rules, so every stored asset other than a substation has a
stored parent of the permitted type.

## Clarifications

### Session 2026-11-19

- Q: When an asset has a very large number of children, should the tree and the details list show them all at
  once, or in pages? → A: Show all at once, no paging; responsiveness at 50,000 assets is covered by SC-006.
- Q: When a switchboard is selected, should the details area also list the panels under it? → A: Yes — its
  panels with asset ID, name, status and their number, each selectable (FR-015).
- Q: Should each tree node also show the asset's operational status? → A: Only when it is not in service: a
  node shows a status label for Maintenance, Out of service or No status recorded; in-service nodes show none
  (FR-002).
- Q: In what order should substations and the children within each type group be listed? → A: By asset ID,
  comparing runs of digits as numbers, so SS-2 comes before SS-10 (FR-004).
- Q: Should the details area show the selected asset's full path from its substation, or only its direct
  parent? → A: The full path (substation › … › selected asset), each ancestor selectable (FR-011).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Explore the hierarchy and see an asset's details (Priority: P1)

An operator opens the asset page and sees every stored substation as the top level of a tree. They expand a
substation to see the transformers, LV boards and switchboards installed at it, expand a switchboard to see its
panels, and select any asset to see all of its details next to the tree.

**Why this priority**: This is the core of the feature: without the tree and the details area, nothing stored
by an import can be seen at all.

**Independent Test**: Import the sample file from feature 001, open the asset page, expand a substation and one
of its switchboards, select a panel, and check that every attribute shown matches the sample row.

**Acceptance Scenarios**:

1. **Given** the sample file has been imported, **When** the operator opens the asset page, **Then** the tree
   shows the 13 stored substations as its top level, each collapsed, and the details area invites the operator
   to select an asset.
2. **Given** the tree is shown, **When** the operator expands a substation, **Then** its transformers, LV boards
   and switchboards appear beneath it, and only nodes that have children can be expanded further.
3. **Given** a switchboard is expanded, **When** the operator selects one of its panels, **Then** the details
   area shows the panel's asset ID, type, name, operational status, voltage, rating, manufacturer, model, serial
   number, commissioned date and parent, and the panel is highlighted in the tree.
4. **Given** an asset with blank optional attributes, **When** it is selected, **Then** each blank attribute is
   shown as explicitly not recorded rather than left empty or shown as zero.
5. **Given** a switchboard panel is selected, **When** the operator looks at the details area, **Then** it
   shows the path from the panel's substation through its switchboard to the panel; **When** the operator
   activates the substation or the switchboard in that path (or the parent), **Then** that asset becomes the
   selected asset.

---

### User Story 2 - Substation overview (Priority: P2)

An operator selects a substation and sees, besides its own details, what is installed at it: its immediate
children grouped by type, and how many transformers, LV boards, switchboards and switchboard panels sit beneath
it in total.

**Why this priority**: It answers the operator's main question about a substation ("what is installed here?")
at a glance, without expanding and counting by hand. It builds on the details area of story 1.

**Independent Test**: Select a substation from the sample whose switchboards have panels, and check the grouped
children and the four counts against the stored data.

**Acceptance Scenarios**:

1. **Given** a substation with transformers, LV boards and switchboards, **When** it is selected, **Then** the
   details area lists its immediate children in three groups — transformers, LV boards, switchboards — each
   child showing its asset ID, name and operational status, and each group showing how many children it holds.
2. **Given** a substation whose switchboards have panels, **When** it is selected, **Then** the details area
   shows the number of transformers, LV boards, switchboards and switchboard panels beneath the substation,
   where the panel count includes every panel under each of its switchboards.
3. **Given** a substation with no children of a type, **When** it is selected, **Then** that type's group says
   there are none and its count is 0.
4. **Given** the grouped children are shown, **When** the operator activates one of them, **Then** that child
   becomes the selected asset.

---

### User Story 3 - Every asset has a stable address (Priority: P2)

Each asset can be reached at its own address. Opening that address directly, refreshing the page or using the
browser's back and forward buttons always shows the same selected asset, with the tree expanded down to it so it
is visible and highlighted. An address for an asset that does not exist shows a clear "not found" state.

**Why this priority**: Operators share and bookmark assets, and the constitution requires the selected asset to
live in the address so refresh, deep links and the back button work (principle V).

**Independent Test**: Open the address of a switchboard panel in a new browser tab: the tree opens with its
substation and switchboard expanded and the panel highlighted, and its details are shown. Then open the address
of an ID that does not exist.

**Acceptance Scenarios**:

1. **Given** a stored panel, **When** the operator opens its address directly, **Then** the panel's details are
   shown, its substation and switchboard are expanded in the tree, and the panel is highlighted and scrolled
   into view.
2. **Given** an asset is selected, **When** the operator refreshes the page, **Then** the same asset is still
   selected, shown and highlighted, with its ancestors expanded.
3. **Given** the operator selected asset A, then asset B, **When** they press the browser's back button,
   **Then** asset A is selected again; forward selects B again.
4. **Given** no stored asset has the ID in the address, **When** the operator opens that address, **Then** the
   details area states that no asset with that ID exists, shows the ID, and offers a way back to the overview;
   the tree is still shown and usable, and no error message is shown.
5. **Given** an asset ID containing characters that are special in addresses (for example a space or a
   slash), **When** the operator selects that asset and later opens its address, **Then** the same asset is
   selected.

---

### User Story 4 - The tree reflects new imports (Priority: P3)

The import panel from feature 001 is still available from the asset page. After an import that saves new assets,
the operator sees them in the tree without reloading the page.

**Why this priority**: It keeps the two features working together, but the operator could also simply refresh
the page, so it ranks below the core browsing stories.

**Independent Test**: With some assets stored and one selected, open the import panel, commit a file that adds a
new substation, close the panel, and check that the new substation is in the tree and the selected asset is
unchanged.

**Acceptance Scenarios**:

1. **Given** an asset is selected, **When** the operator opens the import panel, **Then** the panel opens over
   the asset page and the selected asset stays in the address; closing the panel returns to the same selected
   asset.
2. **Given** an import that commits new assets, **When** it completes, **Then** the tree and the details area
   show the new assets (for example a new substation at the top level, or new children under an expanded
   substation, with updated counts) without a page reload.
3. **Given** an import that stores nothing (validated only, nothing to commit, strict-blocked or failed),
   **When** it completes, **Then** the tree is unchanged.
4. **Given** the registry was empty, **When** the first import commits assets, **Then** the empty state is
   replaced by the tree.

---

### Edge Cases

- **Empty registry**: no assets are stored. The page shows an empty state that says no assets have been
  imported yet and offers to open the import panel; no tree and no error.
- **Loading**: while the top level, a node's children or an asset's details are loading, the affected area
  says so; the rest of the page stays usable.
- **Server unavailable or failing**: the affected area (tree, a node's children, or the details) shows an error
  state that says the data could not be loaded and offers to try again; nothing else is cleared.
- **Unknown asset ID in the address**: "not found" state in the details area (story 3, scenario 4), not an
  error. The same applies when the ID has different letter case from a stored ID: asset IDs match exactly.
- **Blank address segment** (for example `/assets/`): treated as the overview with no asset selected.
- **IDs `.` and `..`**: feature 001 accepts them as asset IDs, but browsers rewrite these path segments
  (even when encoded, in the page address and in requests to the server), so such an asset can be neither
  opened by its own address nor have its details loaded. It is still listed in the tree, with its type and
  status, and can be expanded to reach its children; its label is plain text rather than a link, so it cannot
  be selected. This is a known limitation; FR-017 holds for every other ID.
- **Asset with no children**: shown in the tree without an expand control; a substation with no children shows
  all three groups as empty and all four counts as 0.
- **Many children**: a substation or switchboard with hundreds or thousands of children lists every child at
  once in the tree and in the details lists, in a stable order; there is no paging or "show more".
- **Asset whose ancestors are collapsed**: selecting it through its address or through a link in the details
  area expands its ancestors so it is visible and highlighted.
- **Selected asset is collapsed out of view**: if the operator collapses an ancestor of the selected asset, the
  selection, the address and the details area stay unchanged.
- **Status values**: the operational status is always shown in words (In service, Maintenance, Out of service,
  or No status recorded), never by colour alone.
- **Import panel open on an asset address**: the panel opens over the page and the selected asset stays in the
  address (story 4, scenario 1).

## Requirements *(mandatory)*

### Functional Requirements

**Tree**

- **FR-001**: The asset page MUST show the stored hierarchy as a tree whose top level is every stored
  substation.
- **FR-002**: Each tree node MUST show the asset's ID, name and type in words. A node whose asset is not in
  service MUST also show its status in words — "Maintenance", "Out of service" or "No status recorded"; an
  in-service node shows no status label. *Amended by feature 005 (FR-007): a node shows a status shape and
  the name; the ID, type and status words are in its accessible name and hover text.*
- **FR-003**: A node MUST be expandable if and only if the asset has at least one stored child; expanding it
  MUST show all of its immediate children at once, without paging.
- **FR-004**: The order of nodes MUST be stable: top-level substations by asset ID; children grouped by type in
  the order transformers, LV boards, switchboards, and by asset ID within a type; panels by asset ID. Asset IDs
  are compared with runs of digits taken as numbers, so `SS-2` comes before `SS-10`. The details lists
  (FR-012, FR-015) use the same order.
- **FR-005**: Only the selected asset's ancestors are expanded automatically; all other nodes start collapsed,
  and the operator can expand and collapse any node.
- **FR-006**: The selected asset MUST be highlighted in the tree in a way that does not rely on colour alone,
  and MUST be scrolled into view when it is selected through its address or a link.
- **FR-007**: The tree MUST be operable with the keyboard alone: moving between nodes, expanding, collapsing
  and selecting.

**Details**

- **FR-008**: Selecting an asset MUST show its details: asset ID, type, name, operational status, voltage (kV),
  rating (kVA), manufacturer, model, serial number, commissioned date, and parent asset (ID and name) for every
  type except substation.
- **FR-009**: A blank optional attribute MUST be shown as "Not recorded" (or "No status recorded" for the
  operational status), never as an empty space or zero.
- **FR-010**: Numbers MUST be shown by value with their unit (kV, kVA), so trailing zeros typed in the CSV are
  not kept (`22.50` shows as `22.5 kV`); dates MUST be shown as calendar dates
  in an unambiguous format (for example 2010-03-12).
- **FR-011**: The details MUST show the selected asset's path from its substation down to the asset itself (for
  example `SS-001 › SWB-001-1 › PNL-001-1-3`), each step with its asset ID. Every ancestor in the path, like
  the parent in FR-008, MUST be selectable, making it the selected asset. For a substation the path is the
  substation alone. *Amended by feature 005 (FR-011): each step shows the asset name; its ID is in the
  step's accessible name and hover text.*
- **FR-012**: When a substation is selected, the details MUST also list its immediate children in three
  groups — transformers, LV boards, switchboards — each child with asset ID, name and operational status,
  each group with its number of children, and an empty group stated as having none.
- **FR-013**: When a substation is selected, the details MUST show the number of transformers, LV boards,
  switchboards and switchboard panels beneath it. The panel count MUST include the panels under all of its
  switchboards. The counts MUST equal what the tree shows when every node under the substation is expanded.
- **FR-014**: Every child listed in a substation's groups MUST be selectable, making it the selected asset.
- **FR-015**: When a non-substation asset with children is selected (a switchboard with panels), the details
  MUST list its immediate children with asset ID, name and operational status, and their number; each listed
  child MUST be selectable, making it the selected asset.
- **FR-016**: With no asset selected, the details area MUST invite the operator to select an asset.

**Addresses and navigation**

- **FR-017**: Each asset MUST have a stable address that contains its asset ID (for example
  `/assets/<asset_id>`), with characters that are special in addresses encoded so every stored ID round-trips,
  except the IDs `.` and `..` (see Edge Cases).
- **FR-018**: Selecting an asset (in the tree, through its path or parent, or through a listed child) MUST change the
  address to that asset's address and add a browser history entry, so back and forward move between selections.
- **FR-019**: Opening an asset's address directly or refreshing it MUST select that asset, show its details,
  expand its ancestors and highlight it (FR-005, FR-006).
- **FR-020**: An address with an asset ID that matches no stored asset (exact match, letter case included) MUST
  show a "not found" state in the details area that names the ID and offers a way back to the overview. It MUST
  NOT be shown as an error, and the tree MUST remain usable.
- **FR-021**: The overview address (`/assets`) MUST show the tree with no asset selected.

**States**

- **FR-022**: The tree's top level, each node's children and the details MUST each show a loading state while
  their data is being fetched.
- **FR-023**: When no assets are stored, the page MUST show an empty state saying no assets have been imported
  yet, with a control that opens the import panel.
- **FR-024**: When the tree's top level, a node's children or the details fail to load, the affected area MUST
  say so in words and offer a retry; the other areas MUST keep working.

**Import panel**

- **FR-025**: The import panel from feature 001 MUST remain available from the asset page, whether or not an
  asset is selected, and opening or closing it MUST keep the selected asset in the address.
- **FR-026**: After an import that commits assets, the tree, any expanded node's children, and the details
  area (including a substation's groups and counts) MUST reflect the new assets without a page reload. An
  import that commits nothing MUST NOT change what is shown.

**Read-only**

- **FR-027**: This feature MUST NOT create, change or delete assets; the stored assets are exactly those
  created by imports (feature 001).

### Key Entities

- **Asset**: as defined in feature 001 — a piece of grid equipment identified by its asset ID, with type, name,
  operational status, voltage, rating, manufacturer, model, serial number, commissioned date and an optional
  parent. Substations are the roots of the hierarchy.
- **Asset node**: an asset as it appears in the tree: asset ID, name, type, operational status, and whether it
  has children.
- **Substation summary**: for one substation, its immediate children grouped by type and the counts of
  transformers, LV boards, switchboards and switchboard panels beneath it.
- **Ancestor path**: the chain from a substation down to a given asset (at most three assets), used to expand
  the tree to a selected asset and shown as the path in the details area.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: With the sample file imported, every one of the 205 stored assets can be reached in the tree, and
  for 100% of them the details shown match the stored values attribute by attribute.
- **SC-002**: For every one of the 13 substations in the sample, the four counts shown equal the numbers
  obtained by expanding the whole subtree and counting by hand.
- **SC-003**: Opening the address of any stored asset in a new tab shows that asset selected, highlighted and
  visible in the tree in 100% of attempts; refreshing and back/forward keep the selection consistent in 100% of
  attempts.
- **SC-004**: Opening an address with an unknown asset ID shows the "not found" state, never an error
  message, in 100% of attempts.
- **SC-005**: An operator can go from the overview to the details of a specific switchboard panel, knowing its
  substation and switchboard, in under 30 seconds. This is checked qualitatively, by a hand walk-through
  (expand the substation, expand the switchboard, select the panel), not by a timed test.
- **SC-006**: With 50,000 stored assets, the tree's top level appears within 2 seconds, and expanding a node or
  selecting an asset shows its content within 1 second, measured against the local Docker Compose stack.
- **SC-007**: After an import that adds assets, the new assets are visible in the tree without reloading the
  page in 100% of attempts.
- **SC-008**: Every loading, empty, error and not-found state can be understood from its text alone, without
  relying on colour.

## Assumptions

- Feature 001 is in place: assets are stored by imports and satisfy its type, parent and status rules, so the
  hierarchy is at most three levels deep and every non-substation asset has a stored parent.
- The asset page at `/assets` from feature 001 becomes the explorer; asset addresses are `/assets/<asset_id>`.
  The import panel keeps its existing address parameter, which is combined with the asset address.
- Asset IDs are matched exactly, letter case included, consistent with feature 001 treating IDs as exact
  values.
- Ordering is by asset ID with numeric digit runs, as described in FR-004; other orderings (by name, by status)
  are not needed.
- Only the selected asset is part of the address; which other nodes the operator expanded is not remembered
  across refreshes.
- Details show the stored attributes only. Information about the import that created an asset, and history of
  changes, are not shown (assets are never changed).
- A switchboard panel has no children; a transformer and an LV board have no children under feature 001's
  rules.
- No authentication: every operator sees every asset (constitution, Template Exception 1).
- Out of scope: search (the next feature), editing or deleting assets, user accounts, filtering, exporting, and
  a list of past imports.
