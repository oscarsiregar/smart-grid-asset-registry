# Data model: UI polish

No stored data, API field or request changes. This feature only adds view states, all derived from data the
page already reads.

## View states

| State | Derived from | Values |
|---|---|---|
| Area loading (tree, a node's children) | `useGetAssetNodesQuery(parent).isLoading` | placeholders / rows / error |
| Details loading | `useGetAssetDetailQuery(id)`: `currentData === undefined` and no error | placeholders / details / not found / error |
| Search results loading | `useSearchAssetsQuery(search)`: `currentData === undefined` and no error | placeholders / results / refused / error |
| Register count | `useGetAssetNodesQuery(null)`: `data.totalItems`; `isLoading` | placeholder / ":count substations" / hidden when 0 |
| Header path | `useGetAssetDetailQuery(selectedId).currentData.path` (cached) | names joined by " / ", or the page description |
| Import step | the import panel's mutation results | `CHOOSE_FILE` (no result) → `REVIEW` (validated) → `DONE` (confirmed) |
| Import result loading | the validate mutation's `isLoading` | placeholders below the form / result |

## Import step transitions

```text
CHOOSE_FILE ──Check file succeeds──► REVIEW ──Save succeeds──► DONE
     ▲                                  │                         │
     └──── Choose another file ─────────┘                         │
     └──── Import another file ───────────────────────────────────┘
```

A failed check stays on `CHOOSE_FILE` (its banner shows); a failed save stays on `REVIEW`. A save that commits
nothing (outcome `NOTHING_TO_COMMIT` or `STRICT_BLOCKED`) is a confirmed result, so it is `DONE`, as feature
001 shows it today.

## New copy

| Key | Text |
|---|---|
| `AssetsPageMessage.TITLE` | Asset explorer (was "Assets") |
| `AssetsPageMessage.SEARCH_FIELD_PLACEHOLDER` | Search by ID or name |
| `AssetsPageMessage.REGISTER_LABEL` | Register |
| `AssetsPageMessage.REGISTER_UNIT` / `_ONE` | substations / substation (after the count, shown large) |
| `AssetsPageMessage.CHILDREN_NONE` | Nothing is connected below this asset. |
| `AssetsPageMessage.CHILDREN_HEADING` | Immediate children (:count) |
| `AssetsPageMessage.COLUMN_NAME` / `_ASSET_ID` / `_STATUS` | Name / Asset ID / Status |
| `AssetsPageMessage.PATH_SEPARATOR` | " / " (header line) |
| `ImportPanelMessage.DESCRIPTION` | Validate a file, review what would be rejected, then save the valid rows. |
| `ImportPanelMessage.STEPS_LABEL` | Import steps |
| `ImportPanelMessage.STEP_CHOOSE_FILE` / `_REVIEW` / `_DONE` | Choose file / Review / Done |
| `ImportPanelMessage.STEP_DONE_MARKER` | (done) |
