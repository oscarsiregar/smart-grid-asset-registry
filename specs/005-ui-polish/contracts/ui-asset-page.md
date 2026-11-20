# Contract: the asset page after UI polish

Delta to [../../002-asset-explorer/contracts/ui-asset-explorer.md](../../002-asset-explorer/contracts/ui-asset-explorer.md),
[../../003-asset-search/contracts/ui-asset-search.md](../../003-asset-search/contracts/ui-asset-search.md) and
feature 004. Addresses, requests, keyboard behaviour and every text not listed here are unchanged.

## Layout (1920 × 940)

```
┌ sidebar ─────────────────────┬ header ───────────────────────────────────────────────┐
│ [⚡] Smart Grid Asset Registry │ Asset explorer                                        │
│ [ + Import CSV ]  (outlined)  │ Aurora Substation (North) / Aurora Transformer T1     │
│ [🔍 Search by ID or name   ]  ├ details ──────────────────────────────────────────────┤
│ [ All types              ▾ ]  │ Aurora Substation (North) /            (path, links)  │
│ [Search] [Clear]  (small)     │ TRANSFORMER                                           │
│ SUBSTATIONS                   │ Aurora Transformer T1                                 │
│ ▾ ● Aurora Substation (North) │ [TX-001-1] (● In service)                             │
│   ┃● Aurora Transformer T1    │ ┌ Details ─────────────────────────────────────────┐ │
│    ◑ Aurora Transformer T2    │ │ PARENT  VOLTAGE  RATING  COMMISSIONED  MANUFACT.  │ │
│ › ● Beacon Substation (North) │ │ MODEL   SERIAL NUMBER                             │ │
│   … (scrolls)                 │ └───────────────────────────────────────────────────┘ │
│ ┌ REGISTER ────────────────┐  │ ┌ Immediate children ──────────────────────────────┐ │
│ │ 13 substations           │  │ │ Nothing is connected below this asset.            │ │
│ └──────────────────────────┘  │ └───────────────────────────────────────────────────┘ │
└───────────────────────────────┴────────────────────────────────────────────────────────┘
```

- The sidebar is a column: brand, Import CSV, search form, then the tree or the search results as the only
  scrolling region, then the register card fixed at the bottom. Below 960 px it stays the off-canvas drawer
  of feature 004.
- Search **Search** and **Clear** are small secondary buttons on one row; **Clear** only while a search is in
  the address (feature 003).

## Tree node

| Shown | For assistive technology and hover (`title`) |
|---|---|
| chevron toggle (when expandable), status shape, asset name (truncated with an ellipsis) | asset ID, type in words, status in words; `title` "Aurora Substation (North) · SS-001 · Substation · In service"; "Selected" for the selected node |

The selected row: raised background, accent bar on the left, bold name. IDs `.` and `..`: plain text, no link.

## Header

"Asset explorer"; below it, with an asset selected and its details loaded, its path names joined by " / " as
plain text; otherwise "Substations and the equipment installed in them."

## Details

| Part | Content |
|---|---|
| Path | each step's asset name; ancestors are links; each step's asset ID as hidden text and `title` |
| Intro | type kicker, name title, ID tag, status badge |
| Details card | parent (link; not for a substation), voltage, rating, commissioned date, manufacturer, model, serial number; "Not recorded" for blanks; up to 5 columns |
| Counts | substation only, after the details card |
| Children | substation: "Immediate children (N)" with groups; switchboard: its panels; others: "Immediate children" with "Nothing is connected below this asset." Rows: name, ID tag, status badge, chevron, under the column headings Name / Asset ID / Status (decorative) |

## Import drawer

```
┌ Import assets from CSV ─────────────────────────────────────── [×] ┐
│ Validate a file, review what would be rejected, then save the valid rows. │
├────────────────────────────────────────────────────────────────────┤
│ (1 Choose file) (2 Review) (3 Done)                                │
│ CSV FILE …  IMPORT MODE …  [Check file]                            │  ← scrolls
│ banner · counts · rejected rows · warnings                         │
├────────────────────────────────────────────────────────────────────┤
│ [Save 12 assets] [Choose another file]  Saving…                    │  ← fixed footer, only with a result
└────────────────────────────────────────────────────────────────────┘
```

- The close button shows an icon; its accessible name stays "Close". Escape and focus return are unchanged.
- Steps: the current one has `aria-current="step"`; done ones show a check and "(done)" for screen readers.

## Loading placeholders

| Area | Shape while loading | Announced (hidden text, `role="status"`) |
|---|---|---|
| Tree | 8 rows: dot + name line | Loading assets… |
| A node's children | 3 rows at the child indent | Loading… |
| Details | path line, kicker, title, tag + badge, details card (2 rows × 5), children card (3 rows) | Loading asset… |
| Search results | 4 result rows | Searching… |
| Register | the number | — (the card label stays visible) |
| Import check | banner, 5 count cards, 3 table rows | Checking file… (the form's existing status; the placeholders are `aria-hidden`) |

Placeholders are invisible for the first 200 ms, then fade in and pulse; with reduced motion they appear
static after the same delay. They never replace data already on screen.
