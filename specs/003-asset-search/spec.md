# Feature Specification: Asset Search

**Feature Branch**: `003-asset-search`

**Created**: 2026-11-19

**Status**: Draft

**Input**: User description: "Feature short name: asset-search. Find an asset by ID or name and jump to it. WHY: With
hundreds or thousands of assets, an operator who knows an asset's ID or part of its name needs to find it without
expanding the tree by hand. FROM THE BRIEF: Allow exact or partial search by asset ID or asset name. Search may
additionally filter by asset type. Selecting a result must: navigate to the asset's page; expand the tree enough to
reveal its ancestors; highlight the selected asset; display its details. Use the stable route from feature 002
(/assets/<asset_id>); direct navigation and refresh keep the selected asset. Provide appropriate loading, empty and
error states. Priority order: correctness, then readability, then optimisation. CONTEXT: Feature 002 already provides
the tree, the details area, deep links that expand ancestors and highlight the asset, and the not-found state;
selecting a search result should reuse that behaviour rather than duplicate it. Assets are read-only here. OUT OF
SCOPE: editing or deleting assets, full-text search engines, user accounts."

## Context

Feature 002 ([../002-asset-explorer/spec.md](../002-asset-explorer/spec.md)) turned the asset page into an explorer:
a tree of the stored hierarchy beside a details area, and a stable address `/assets/<asset_id>` that, when opened,
expands the asset's ancestors, highlights it, scrolls it into view and shows its details, or shows "not found" for
an unknown ID. This feature adds a search to the same page. A search only *finds* assets; showing a found asset is
done by going to its address, so the result looks exactly as if the operator had opened that address.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Find an asset by ID or name and open it (Priority: P1)

An operator types all or part of an asset ID or name into the search field on the asset page and starts the
search. A list of matching assets appears, each with its ID, name, type, status and where it sits in the hierarchy.
The operator picks one; the page goes to that asset's address, the tree expands to reveal it, the asset is
highlighted and its details are shown.

**Why this priority**: This is the whole purpose of the feature: reaching a known asset without walking the tree.
It is useful on its own even without the type filter.

**Independent Test**: With the sample imported, search for `PNL-001-1-01` and for part of a name such as
`aurora panel`, pick a result, and check that the address, the tree, the highlight and the details are the same
as when opening `/assets/PNL-001-1-01` directly.

**Acceptance Scenarios**:

1. **Given** the sample is imported, **When** the operator searches for the full ID `PNL-001-1-01`, **Then** that
   asset is the first result.
2. **Given** the sample is imported, **When** the operator searches for `aurora panel`, **Then** every asset whose
   name contains "aurora panel" (letter case ignored) is listed, and no other.
3. **Given** the sample is imported, **When** the operator searches for `swb-001`, **Then** every asset whose ID
   contains "SWB-001" is listed, letter case ignored.
4. **Given** a list of results, **When** the operator selects one, **Then** the address becomes that asset's
   address, its ancestors are expanded, it is highlighted and scrolled into view, and its details are shown, exactly
   as in feature 002 when opening that address.
5. **Given** the operator selected a result, **When** they use the browser's Back button, **Then** the search with
   its results is shown again; Forward returns to the selected asset.
6. **Given** a search with results is shown, **When** the operator refreshes the page or opens the same address in
   a new tab, **Then** the same search and results are shown.

---

### User Story 2 - Narrow a search by asset type (Priority: P2)

An operator who is looking for a switchboard panel, not a substation of the same name, chooses a type alongside
the search text. Only assets of that type are listed.

**Why this priority**: The brief allows it and it removes noise for common fragments (every asset at a site shares
its prefix), but a text search alone already lets the operator find any asset.

**Independent Test**: Search for `001` with and without the type "Switchboard panel" and compare the lists.

**Acceptance Scenarios**:

1. **Given** the sample is imported, **When** the operator searches for `001` with the type "Switchboard panel",
   **Then** only switchboard panels whose ID or name contains "001" are listed.
2. **Given** a type is chosen, **When** the operator sets it back to "All types", **Then** matches of every type are
   listed.
3. **Given** a search with a type, **When** the page is refreshed or the address shared, **Then** the same text and
   type are applied.

---

### User Story 3 - Understand every state of a search (Priority: P2)

While a search is running, when it finds nothing, when it finds more than can be shown, and when it fails, the
operator is told so in words and knows what to do next.

**Why this priority**: The brief requires loading, empty and error states; without them a search that finds
nothing looks broken.

**Independent Test**: Search for a text that matches nothing, a text that matches thousands of assets, and search
while the server is stopped.

**Acceptance Scenarios**:

1. **Given** a search was started, **When** the results are not there yet, **Then** the results area says the
   search is running.
2. **Given** the operator searches for a text no asset matches, **When** the search completes, **Then** the
   results area says that no asset matches that text (and that type, if one is chosen) and suggests checking the
   spelling or removing the type.
3. **Given** a search matches more than 50 assets, **When** the results appear, **Then** the first 50 are listed,
   the operator is told how many assets match in total, and is asked to refine the search with more text or a
   type.
4. **Given** the server cannot be reached, **When** a search is started, **Then** the results area says the search
   failed and offers to try again; the tree and the details area keep working.

---

### Edge Cases

- **Blank search**: a search text made only of spaces is not run; the page shows the tree as without a search.
  Leading and trailing spaces are ignored.
- **Letter case**: matching ignores letter case for both ID and name (`pnl` finds `PNL-001-1-01`). This differs on
  purpose from addresses, which match an ID exactly (feature 002); a search is a way to find the exact ID.
- **Characters with a special meaning elsewhere**: `%`, `_`, `\`, `/`, quotes and spaces in the search text are
  matched literally, as ordinary characters.
- **Many matches**: a short text such as `SS` can match thousands of assets; only the first 50 are listed, never
  silently — the total count is always shown with a request to refine the search.
- **An asset matches through both ID and name**: it is listed once.
- **Empty registry**: when no assets are stored, the page shows the empty state from feature 002 and no search.
- **IDs `.` and `..`**: such an asset can be found and is listed, but, as in feature 002, it cannot be selected;
  the result says so instead of being a link.
- **Results change after an import**: a search shown while an import commits new assets is refreshed, so the
  results include the new assets, as the tree is (feature 002).
- **A result is selected while its details cannot be loaded**: the details area shows its error state from
  feature 002; the search itself is not affected.
- **Search together with the import panel**: opening the import panel keeps the search in the address, and closing
  it shows the same search again.

## Requirements *(mandatory)*

### Functional Requirements

**Searching**

- **FR-001**: The asset page MUST offer a search field whenever at least one asset is stored.
- **FR-002**: A search MUST match every stored asset whose asset ID or asset name contains the search text,
  ignoring letter case and ignoring leading and trailing spaces in the text; a text that equals an ID or a name is
  the exact case of the same rule.
- **FR-003**: Every character of the search text MUST be matched literally; no character acts as a wildcard or
  operator.
- **FR-004**: A search MUST be started explicitly (by pressing Enter in the field or a search button); a blank text
  MUST NOT start a search.
- **FR-005**: The operator MUST be able to limit a search to one asset type (substation, transformer, LV board,
  switchboard, switchboard panel) or search all types; the default is all types.
- **FR-006**: The search text and the chosen type MUST be part of the page address, so refresh, a shared link and
  Back/Forward show the same search.

**Results**

- **FR-007**: Each result MUST show the asset's ID, name, type in words, operational status in words, and its
  position in the hierarchy as the chain of ancestor IDs from its substation (for example
  `SS-001 › SWB-001-1`); a substation shows that it is at the top level.
- **FR-008**: Results MUST be ordered so that assets whose ID or name equals the search text (letter case ignored)
  come first, followed by the others; within each of those two groups the order MUST be the explorer's order
  (by type, then by asset ID with digit runs compared as numbers, feature 002 FR-004).
- **FR-009**: An asset MUST appear at most once in the results.
- **FR-010**: The results MUST show at most the first 50 matches, in the order of FR-008, and MUST state how many
  assets match in total. When more than 50 match, the results MUST say so (for example "Showing 50 of 1 234
  matches") and ask the operator to refine the search with more text or a type; there is no paging or "show more".
- **FR-011**: Selecting a result MUST go to that asset's address from feature 002 (`/assets/<asset_id>`), adding a
  history entry; what the page then shows (ancestors expanded, asset highlighted and in view, details) MUST be the
  same as opening that address directly. The search MUST NOT implement its own way of revealing or showing an
  asset.
- **FR-012**: A result whose asset cannot be selected (IDs `.` and `..`, feature 002 edge case) MUST be listed with
  its details as text and MUST say that it cannot be opened.
- **FR-013**: Clearing the search MUST return the page to the tree without a search, and remove the search from
  the address.

**States**

- **FR-014**: While a search is running, the results area MUST say so; the rest of the page MUST stay usable.
- **FR-015**: A search that matches nothing MUST say so in words, naming the search text and the chosen type, and
  MUST suggest how to widen it.
- **FR-016**: A search that fails MUST say so in words and offer a retry; the tree and the details area MUST keep
  working.
- **FR-017**: After an import that commits assets, a shown search MUST be refreshed without a page reload; an
  import that commits nothing MUST NOT change it.

**Read-only**

- **FR-018**: This feature MUST NOT create, change or delete assets.

### Key Entities

- **Asset**: as defined in feature 001 — identified by its asset ID, with a name, a type, an operational status and
  an optional parent. Only the ID and the name are searched.
- **Search**: the search text (trimmed, at least one non-space character) and an optional asset type. It is part of
  the page address.
- **Search result**: one matching asset — its ID, name, type, operational status, and the chain of ancestor IDs from
  its substation — plus whether it is an exact match.
- **Result summary**: the total number of matching assets and how many are shown (at most 50).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: With the sample imported, searching for the full ID of any one of the stored assets lists that asset
  first in 100% of cases.
- **SC-002**: With the sample imported, for every search text tried in the acceptance tests, the results are
  exactly the assets whose ID or name contains that text (letter case ignored): none missing, none extra.
- **SC-003**: An operator who knows an asset's ID reaches its highlighted details from the asset page in under
  15 seconds.
- **SC-004**: Selecting a result produces the same address, expanded ancestors, highlight and details as opening
  that asset's address directly, in 100% of attempts.
- **SC-005**: With 50 000 stored assets, the results of a search appear within 1 second, measured against the
  local Docker Compose stack.
- **SC-006**: Every running, no-match, too-many-matches and failed state can be understood from its text alone,
  without relying on colour.

## Assumptions

- Features 001 and 002 are in place: assets are created by imports, and the explorer, its addresses, its reveal and
  highlight behaviour and its not-found state exist and are reused unchanged.
- The search is part of the asset page, next to the tree; there is no separate search page.
- "Partial" means "contains" (anywhere in the ID or name), not only "starts with"; "exact" is the case where the
  text equals the whole ID or name, and such matches are listed first.
- Letter case is ignored in searches, unlike addresses; accents and other characters are compared as typed (no
  folding of `é` to `e`).
- Only asset ID and asset name are searched; manufacturer, model, serial number and other attributes are not.
- A search is started on request (Enter or button), not on every keystroke; this keeps results predictable and
  avoids a request per character.
- Selecting a result leaves the search: the address becomes the asset's address without the search, and Back
  returns to the results.
- At most 50 results are shown, with the total count; reaching a match beyond the first 50 means refining the
  search. Because exact matches come first and asset IDs are unique, a search for a full asset ID always lists
  that asset first, however many other assets contain the text.
- The type filter narrows a text search; choosing a type without any text does not start a search (browsing all
  assets of a type is not part of this feature).
