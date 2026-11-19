# Contract: asset search UI

The search lives on the asset page of feature 002
([../../002-asset-explorer/contracts/ui-asset-explorer.md](../../002-asset-explorer/contracts/ui-asset-explorer.md)).
Page code: `frontend/src/pages/assets`; data: `frontend/src/modules/assets` ([assets-search-api.md](assets-search-api.md)).

## Addresses

| Address | Shows |
|---|---|
| `/assets?q=<text>` | the search results in place of the tree; the details area as without a search |
| `/assets?q=<text>&type=<TYPE>` | the same, limited to one type |
| `/assets/<id>?q=<text>` (typed by hand) | the results in place of the tree, the details of `<id>` |
| `…&panel=import` | the import panel over the page; the search is kept |

The address's `type` parameter is sent to the API as `assetType`; `q` keeps its name.

Submitting the form pushes a history entry with `q` (and `type`); clearing removes both. A result links to
`/assets/<id>` **without** the search, so following it leaves the search and Back returns to it.

## Layout

```
┌ page intro ─────────────────────────────────────── [Import CSV] ┐
├ search (form, role=search) ──────────────────────────────────────┤
│ [ Search by asset ID or name________ ] [All types ▾] [Search] [Clear] │
├ results (in place of the tree) ─┬ details area (feature 002) ───┤
│ 12 assets match "pnl-001-1"     │ Select an asset in the tree…  │
│ • PNL-001-1-01 Aurora Panel A1  │                               │
│   Switchboard panel · Maint.    │                               │
│   SS-001 › SWB-001-1            │                               │
│ • …                             │                               │
└─────────────────────────────────┴───────────────────────────────┘
```

Shown only when at least one asset is stored (the empty registry state of 002 has no search).

## Search form

- `role="search"` landmark; the text field is the restored template `Input` with a visible label "Search assets"
  (hint: "Asset ID or name, or part of it") and a leading search icon; the type field is a labelled `<select>`
  ("Asset type": "All types", "Substation", "Transformer", "LV board", "Switchboard", "Switchboard panel").
- **Search** submits (Enter in the field too); a blank text shows "Enter an asset ID or name to search." under the
  field and does not change the address. **Clear** (shown while a search is in the address) empties the form and
  removes the search.
- The form's values come from the address, so a refresh or link fills them.

## Results area (`role="region"`, labelled "Search results")

| State | Shows |
|---|---|
| Running | "Searching…" (`role="status"`) |
| All matches received | "N assets match “<text>”" ("1 asset matches …" for one) (and " of type <Type>" when chosen), then the list |
| More matches than received | "Showing <items received> of N assets that match “<text>”. Add more of the ID or name, or choose a type, to narrow the search." then the received items (the backend's first 50) |
| No match | "No asset matches “<text>”" (+ " of type <Type>"). "Check the spelling, or search all types." |
| Refused (`400 INVALID_REQUEST`) | the server's message as sent, for example "Invalid request: q must be at most 100 characters"; no retry (the same search would be refused again); the form stays editable |
| Failure | "The search could not be completed." + **Try again** |

Each result (a list item) shows: the asset ID and name as one link to `/assets/<id>`; the type in words; the
status in words ("No status recorded" for none); and the location — the ancestor IDs joined with " › ", or
"Top level" for a substation. A result whose ID is `.` or `..` has no link and says "This asset cannot be opened."

## After an import

A committed import refreshes a shown search (tag `ASSET`); an import that saved nothing changes nothing. The
summary and list update in place; the form keeps its values.

Quoted search text uses typographic quotes (“…”) in all copy.

Every state is readable from its text alone; nothing relies on colour (principle V, SC-006).
