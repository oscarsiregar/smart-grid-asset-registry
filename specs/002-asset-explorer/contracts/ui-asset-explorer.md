# Contract: asset explorer UI

The `/assets` page from feature 001 becomes the explorer. Page code: `frontend/src/pages/assets`; data:
`frontend/src/modules/assets` ([assets-api.md](assets-api.md)).

## Addresses

| Address | Shows |
|---|---|
| `/` | redirects to `/assets` (unchanged) |
| `/assets` | tree, no asset selected, details area invites a selection (FR-016, FR-021) |
| `/assets/<assetId>` | tree expanded down to the asset, the asset highlighted, its details (FR-017–FR-019) |
| `…?panel=import` | the import panel from feature 001 over the page; the path is kept (FR-025) |

`<assetId>` is percent-encoded (`/assets/SS%2F1`). Selecting an asset (tree, path, listed child) pushes a history
entry; opening/closing the panel behaves as in feature 001.

## Layout

```
┌ page intro ─────────────────────────────────────────────┐
│ Asset registry · Smart Grid Asset Registry  [Import CSV] │
├ tree (role=tree) ──────────┬ details area ──────────────┤
│ ▾ SS-001 Aurora Substation │ SS-001 › SWB-001-1 › PNL-…  │  ← path, each step a link
│   ▸ TX-001-1 Transformer…  │ Aurora Panel 1-3            │
│   ▾ SWB-001-1 Switchboard  │ Switchboard panel           │
│      • PNL-001-1-3  ◀ sel. │ attributes (dl)             │
│ ▸ SS-002 Beacon …          │ [children / groups / counts]│
└────────────────────────────┴─────────────────────────────┘
```

Narrow screens stack the tree above the details.

## Tree

- Top level: every substation, ordered by the API. Each node: expand toggle (only if `childCount > 0`), asset ID,
  name, type in words, and a status label only when not in service ("Maintenance", "Out of service", "No status
  recorded").
- Expanding a node loads its children the first time (loading row "Loading…"; failure row "Could not load
  children" + **Try again**); later expansions use the cache.
- Selected node: `aria-selected="true"`, a visible marker and bold text, scrolled into view.
- Keyboard: Tab enters the tree on the selected (or first) node; ↑/↓ move; → expands or moves to the first
  child; ← collapses or moves to the parent; Home/End; Enter follows the node's link.

## Details area

| State | Shows |
|---|---|
| No selection | "Select an asset in the tree to see its details." |
| Loading | "Loading asset…" (`role="status"`) |
| Loaded | path (links), name and type heading, attribute list; blanks "Not recorded" / "No status recorded"; numbers with kV / kVA; date `YYYY-MM-DD`; parent link (non-substations) |
| Loaded, substation | plus three groups — Transformers (n), LV boards (n), Switchboards (n) — each child a link with ID, name, status; empty group "None"; plus counts beneath: transformers, LV boards, switchboards, switchboard panels |
| Loaded, switchboard | plus "Switchboard panels (n)" list of links |
| Not found (`ASSET_NOT_FOUND`) | "No asset with ID <id> exists." and a link "Back to all assets" (`/assets`); no error styling or toast |
| Error | "The asset could not be loaded." + **Try again** |

## Page-level states

| State | Shows |
|---|---|
| Top level loading | "Loading assets…" in the tree area |
| Empty registry (`totalItems = 0`) | "No assets have been imported yet." + **Import CSV** (opens the panel); no tree, no details area |
| Top level failed | "The assets could not be loaded." + **Try again**; the Import CSV button still works |

Every state is readable from its text alone; status and selection never rely on colour (principle V, SC-008).

## After an import

A committed import (`isCommitted: true`) refetches every visible asset list and the details; a result that saved
nothing changes nothing (FR-026). The selection and the expanded nodes stay.
