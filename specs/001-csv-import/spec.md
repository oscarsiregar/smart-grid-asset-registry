# Feature Specification: CSV Import of Grid Assets

**Feature Branch**: `001-csv-import`

**Created**: 2026-11-19

**Status**: Draft

**Input**: User description: "Feature short name: csv-import. Import grid assets from a CSV file.
Upstream systems export master asset data (substations and their equipment, linked by parent
asset ID) as CSV files with data-quality problems. An operator uploads such a file, sees which
rows are accepted and why the others are rejected, stores the accepted rows, and knows for
certain whether anything was committed."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Validate a file and see every problem (Priority: P1)

An operator selects a CSV export and asks the system to check it. Nothing is stored. The
operator sees the total number of rows, how many would be imported, how many are unchanged
compared with what is already stored, and how many are rejected. Each rejected row shows its
original CSV row number, its asset ID when present, a stable error code and a readable message.
The result states plainly that nothing has been saved yet.

**Why this priority**: Explaining data-quality problems is the core value. Even without saving,
an operator can use this to fix the upstream export.

**Independent Test**: Upload the supplied sample file and check that every planted problem
(missing ID, missing name, unknown type, unknown status, invalid date, negative rating, missing
parent, substation with a parent, non-substation without a parent, wrong parent type, cycle)
is reported on the correct row with the correct code, and that the store is still empty.

**Acceptance Scenarios**:

1. **Given** an empty store, **When** the operator validates the sample file, **Then** the result
   shows total, to-import, unchanged and rejected counts that add up to the number of data rows,
   lists every rejected row with row number, asset ID (if any), error code and message, and
   states "Nothing has been saved".
2. **Given** a file where a panel (row 2) appears before its switchboard (row 5) and substation
   (row 9), **When** the file is validated, **Then** all three rows are accepted.
3. **Given** the same file with its data rows shuffled, **When** it is validated, **Then** the
   same asset IDs are accepted and the same asset IDs are rejected with the same error codes.
4. **Given** a file whose header is missing a required column, **When** it is validated,
   **Then** the whole file is rejected with a file-level error naming the missing column, and no
   row results are shown.

---

### User Story 2 - Confirm and save the valid rows (Priority: P1)

After reviewing a validation result, the operator confirms the import. All valid rows are saved
together in one atomic step; rejected rows are reported and not saved. The result clearly states
that data was committed and how many assets were added.

**Why this priority**: Storing accepted assets is the purpose of the import and is needed by
the later browsing and search features.

**Independent Test**: Validate then confirm the sample file on an empty store; check that
exactly the accepted rows are stored, with correct parent links, and the result says
"Committed".

**Acceptance Scenarios**:

1. **Given** a validated file with 10 valid and 2 rejected rows, **When** the operator confirms
   in partial mode, **Then** the 10 valid rows are stored, the 2 rejected rows are listed, and
   the result states that data was committed with 10 imported.
2. **Given** an unexpected failure while saving, **When** the operator confirms, **Then** no
   rows from that file are stored and the result states that nothing was committed.
3. **Given** a validated file, **When** the stored data changed before the operator confirmed
   (for example another import added a conflicting asset), **Then** the confirmed result reflects
   the data at the moment of saving, and any difference from the earlier validation is visible
   in the counts and rejected rows.

---

### User Story 3 - Strict (all-or-nothing) import (Priority: P2)

An operator who needs the file to load completely or not at all chooses strict mode. If any row
is rejected, nothing is saved and the result says so; if no row is rejected, all rows are saved
together.

**Why this priority**: Some upstream loads must stay consistent as a set. Partial import is the
default, so this option adds value but is not needed for the first release.

**Independent Test**: Confirm the sample file in strict mode and check that nothing is stored
and the result explains that strict mode blocked the commit because of rejected rows.

**Acceptance Scenarios**:

1. **Given** a file with at least one rejected row, **When** the operator confirms in strict
   mode, **Then** nothing is stored and the result states "Not committed: strict mode, N rows
   rejected", still listing every rejected row.
2. **Given** a file with no rejected rows, **When** the operator confirms in strict mode,
   **Then** every new row is stored and the result states that data was committed.

---

### User Story 4 - Re-import the same or an overlapping file safely (Priority: P2)

Operators often receive the same export again or a newer export that repeats earlier assets.
Rows identical to stored assets are reported as unchanged; rows that reuse a stored asset ID
with different values are rejected because editing assets is out of scope.

**Why this priority**: Without this, repeating an import would either duplicate data or fail
noisily, which makes the tool unsafe to use routinely.

**Independent Test**: Import the sample file, then import it again; check that the second
result shows 0 imported, every previously imported row as unchanged, and the same rejections.

**Acceptance Scenarios**:

1. **Given** the sample file was already imported, **When** it is imported again, **Then** the
   result shows 0 imported, all previously stored rows as unchanged, and states that there was
   nothing new to commit.
2. **Given** a stored asset TX-001-1 with status IN_SERVICE, **When** a file contains TX-001-1
   with status MAINTENANCE, **Then** that row is rejected with code `ASSET_ID_CONFLICT` and a
   message naming the differing fields.
3. **Given** a stored substation SS-001, **When** a new file contains only a transformer whose
   parent is SS-001, **Then** the transformer is accepted because its parent already exists.

---

### Edge Cases

- **Empty file or header only**: file-level error "file contains no data rows"; nothing saved.
- **Not a CSV / unreadable encoding / larger than the size limit**: file-level error; nothing saved.
- **Row with the wrong number of fields**: that row is rejected with `MALFORMED_ROW`; other rows
  are processed normally.
- **Duplicate asset ID within the file**: every row sharing that ID is rejected with
  `DUPLICATE_ASSET_ID`, so the result does not depend on which copy comes first.
- **Missing asset ID**: the row is rejected with `ASSET_ID_REQUIRED` and its asset ID is shown as
  blank; any rows naming it as parent cannot reference it.
- **Child of a rejected row**: a row whose parent appears in the file but is itself rejected is
  rejected with `PARENT_REJECTED`, because its chain no longer reaches a stored or accepted
  substation. This applies down the whole subtree.
- **Cycle**: every row in the cycle (e.g. A→B→A, or C→D→E→C) is rejected with `CYCLE_DETECTED`;
  rows hanging below the cycle are rejected with `PARENT_REJECTED`.
- **Parent ID that looks like a self-reference but names a different, missing asset**
  (e.g. SWB-S with parent SWB-SELF): rejected with `PARENT_NOT_FOUND`, not `SELF_PARENT`.
- **Row with several problems**: all of its field-level errors are listed; hierarchy checks are
  only reported for rows whose own fields are valid.
- **Duplicate serial number** (different asset IDs, same non-blank serial number, within the
  file or against a stored asset): rows are not rejected; each affected row is accepted and
  carries a `DUPLICATE_SERIAL_NUMBER` warning.
- **Leading/trailing whitespace** in any field is ignored; a field containing only whitespace
  counts as empty.
- **Two imports confirmed at the same time**: they are processed one after the other; the second
  sees the first one's stored assets.

## Requirements *(mandatory)*

### Functional Requirements

**File handling**

- **FR-001**: Operators MUST be able to select a CSV file and submit it for validation without
  anything being stored.
- **FR-002**: The system MUST require a header row containing all eleven columns: asset_id,
  parent_asset_id, asset_type, asset_name, operational_status, voltage_kv, rating_kva,
  manufacturer, model, serial_number, commissioned_date. Column order MUST NOT matter;
  unknown extra columns MUST be ignored.
- **FR-003**: The system MUST reject the whole file with a file-level error (stable code and
  message, no row results) when it is empty, has no data rows, is missing a required column,
  cannot be read as CSV text, or exceeds the size limit.
- **FR-004**: Row numbers MUST be reported as the line number in the original file, with the
  header as row 1.

**Row validation**

- **FR-005**: asset_id, asset_type and asset_name MUST be present (`ASSET_ID_REQUIRED`,
  `ASSET_TYPE_REQUIRED`, `ASSET_NAME_REQUIRED`).
- **FR-006**: asset_id MUST be unique within the file; all rows sharing an ID MUST be rejected
  (`DUPLICATE_ASSET_ID`).
- **FR-007**: asset_type MUST be one of SUBSTATION, TRANSFORMER, LV_BOARD, SWITCHBOARD,
  SWITCHBOARD_PANEL (`INVALID_ASSET_TYPE`).
- **FR-008**: operational_status, when present, MUST be one of IN_SERVICE, MAINTENANCE,
  OUT_OF_SERVICE (`INVALID_OPERATIONAL_STATUS`).
- **FR-009**: voltage_kv and rating_kva, when present, MUST be valid decimal numbers
  (`INVALID_NUMBER`); rating_kva MUST NOT be negative (`NEGATIVE_RATING`).
- **FR-010**: commissioned_date, when present, MUST be a real calendar date in either
  day/month/two-digit-year form (e.g. 15/2/09) or year-month-day form (e.g. 2009-02-15)
  (`INVALID_DATE`).

**Hierarchy validation**

- **FR-011**: A SUBSTATION MUST NOT have a parent (`SUBSTATION_HAS_PARENT`); every other type
  MUST have one (`PARENT_REQUIRED`).
- **FR-012**: An asset MUST NOT name itself as its parent (`SELF_PARENT`).
- **FR-013**: The parent MUST be either a stored asset or an asset in the same file
  (`PARENT_NOT_FOUND`).
- **FR-014**: The parent's type MUST be the permitted parent for the child's type: SUBSTATION
  for TRANSFORMER, LV_BOARD and SWITCHBOARD; SWITCHBOARD for SWITCHBOARD_PANEL
  (`INVALID_PARENT_TYPE`).
- **FR-015**: Rows forming a cycle MUST be rejected (`CYCLE_DETECTED`).
- **FR-016**: Every accepted asset's chain of parents MUST end at a stored or accepted
  SUBSTATION; a row whose in-file parent is rejected MUST be rejected (`PARENT_REJECTED`).
- **FR-017**: The accepted and rejected sets and their error codes MUST be identical for any
  ordering of the same data rows.

**Re-import**

- **FR-018**: A row whose asset ID is already stored with identical values (after trimming
  whitespace and comparing dates and numbers by value) MUST be counted as unchanged, not as
  imported and not as rejected.
- **FR-019**: A row whose asset ID is already stored with different values MUST be rejected
  (`ASSET_ID_CONFLICT`), and the message MUST name the differing fields.

**Warnings**

- **FR-020**: When two or more accepted or stored assets with different asset IDs share the same
  non-blank serial number, each such row in the file MUST carry a `DUPLICATE_SERIAL_NUMBER`
  warning naming the other asset IDs. Warnings MUST NOT change whether a row is accepted.

**Commit**

- **FR-021**: Operators MUST be able to confirm a validated file to save it, choosing partial
  mode (default) or strict mode.
- **FR-022**: On confirm, the system MUST re-validate the file against the store at that moment
  and save all rows to import in a single atomic step: either all of them are stored or none.
- **FR-023**: In strict mode, if any row is rejected, the system MUST store nothing.
- **FR-024**: Imports MUST be applied one at a time so that two confirmations cannot together
  break uniqueness or hierarchy rules.

**Results**

- **FR-025**: Every result MUST show total rows, imported (or "to import" when only validated),
  unchanged and rejected counts, where total = imported + unchanged + rejected.
- **FR-026**: Every rejected row MUST show its CSV row number, its asset ID when present, and one
  or more errors, each with a stable error code and a readable English message.
- **FR-027**: Every result MUST state its commit outcome as exactly one of: *Validated only —
  nothing saved*; *Committed — N assets saved*; *Nothing to commit — no new assets*;
  *Not committed — strict mode, N rows rejected*; *Not committed — the import failed and
  was rolled back*. The outcome MUST be conveyed in text, not by colour alone.
- **FR-028**: Results MUST list warnings separately from rejections, with row number, asset ID,
  code and message.

### Key Entities

- **Asset**: a piece of grid equipment identified by asset ID. Attributes: asset type, name,
  operational status, voltage (kV), rating (kVA), manufacturer, model, serial number,
  commissioned date, and an optional parent asset. Substations are roots; every other asset
  has exactly one parent of the permitted type.
- **Import file**: the uploaded CSV; a header plus data rows, each with an original row number.
- **Import result**: the outcome of validating or committing one file: mode (partial/strict),
  counts, commit outcome, rejected rows with their errors, and warnings.
- **Row error / warning**: a row number, asset ID if present, stable code and English message.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Every planted problem in the supplied sample file is reported on the correct row
  with the correct error code, and no valid row is rejected (checked by an automated golden
  test).
- **SC-002**: Shuffling the data rows of the sample file into 20 random orders produces the same
  accepted asset IDs, rejected asset IDs and error codes every time.
- **SC-003**: Importing the sample file twice results in 0 new assets and 0 duplicated assets
  on the second run.
- **SC-004**: An operator can upload, review and commit the sample file in under 2 minutes, and
  the validation result appears within 3 seconds of submitting the file.
- **SC-005**: In 100% of results, an operator can tell from the text alone whether data was
  saved, without reading colours or icons.
- **SC-006**: After any failed or strict-mode-blocked import, the stored assets are exactly the
  same as before the attempt.

## Assumptions

- Operators are trusted; there are no user accounts or permissions (constitution, Template
  Exception 1).
- Files are UTF-8 (a leading byte-order mark is accepted), comma-separated, with standard
  double-quote escaping, and at most 5 MB.
- Two-digit years in day/month/year dates mean 2000–2099 (the sample uses 08–20).
- Asset type and status values are matched exactly, case-sensitively, after trimming whitespace.
- operational_status, voltage_kv, rating_kva, manufacturer, model, serial_number and
  commissioned_date are optional; only rating_kva has a sign rule, as in the brief.
- Inner whitespace in names is kept as supplied (e.g. "Dual Node  B").
- Import results are shown to the operator once and are not stored; refreshing the import page
  returns to an empty upload form.
- The brief calls the sample "a valid dataset of 220 assets across 13 substations", but the file
  holds 222 data rows including deliberately invalid test rows; the golden test's expected
  counts are derived from these rules, not from the brief's numbers.
- Browsing the hierarchy, search, editing and deleting assets, and user accounts are out of
  scope for this feature.
