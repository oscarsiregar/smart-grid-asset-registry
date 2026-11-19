# Contract: Import panel (UI)

**Feature**: [../spec.md](../spec.md) · **Decisions**: [../research.md](../research.md) R10
**Constitution**: Principle V

## URL state

| URL | Shows |
|---|---|
| `/` | redirects to `/assets` |
| `/assets` | main page (empty state until the browse feature) with an **Import CSV** button |
| `/assets?panel=import` | main page with the import panel open |

- Opening the panel pushes `?panel=import` (back button closes it); closing removes it.
- Refresh or a deep link to `/assets?panel=import` opens the panel with an empty form (results are
  not stored, spec Assumptions).
- An unknown `panel` value is ignored (panel closed).
- Later features add further parameters (selected asset) without changing this one.

## Panel behaviour

- Native modal `<dialog>`: focus moves into the panel on open, the page behind is inert, **Escape**
  and the **Close** button close it, focus returns to the **Import CSV** button.
- Labelled by its heading ("Import assets from CSV").

## States

| State | What the operator sees |
|---|---|
| Empty form | file picker, mode choice (Partial — default / Strict, each with one-line explanation), **Check file** button (disabled until a file is chosen) |
| Checking / saving | button shows loading and is disabled; status text "Checking file…" / "Saving…" announced via `role="status"` |
| File-level error | `role="alert"` message from the server (`message`), no row results; form stays usable |
| Validated | outcome line **"Validated only — nothing saved"**, counts (total, to import, unchanged, rejected, rows with warnings), rejected-rows table, warnings table, **Save N assets** button, **Choose another file** |
| Committed / Nothing to commit / Strict-blocked | outcome line with the spec's exact text (FR-027), counts, tables, **Import another file** |
| Failed (500) | **"Not committed — the import failed and was rolled back"**, retry possible |
| Network failure | the global error toast plus the failed outcome line above |

Every outcome is written in text and paired with an icon that has a text label; colour is never the
only signal (Principle V, SC-005).

## Tables

- Rejected rows: columns *Row*, *Asset ID* ("—" with visually hidden "none" when absent), *Code*,
  *Message*; one line per error, rows sorted by row number; caption states the count.
- Warnings: *Row*, *Asset ID*, *Code*, *Message*.
- Empty tables are replaced by a sentence ("No rows were rejected.").

## Resetting

Choosing another file, switching mode, or **Import another file** clears the shown result
(resets both mutations) so a result is never shown next to a different file.
