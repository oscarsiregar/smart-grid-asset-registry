# Research: Asset Search

**Feature**: [spec.md](spec.md) · **Plan**: [plan.md](plan.md)

Decisions for finding an asset by ID or name. Each entry: decision, rationale, alternatives. The base is feature
002 ([../002-asset-explorer/research.md](../002-asset-explorer/research.md), cited as 002-R*n*).

## R0. Template fit and the one deviation

**Decision**: The search follows the template wherever it can: it is a filter on the existing list read
(`GetListAsset`, `FindNodes`, `FindAssetNodesFilter`), it keeps the template's list paging contract (`page`,
`limit`, `SetDefault`, `MaxLimit` 100, `{ items, totalItems }`), and the frontend form uses `useAppForm` with a Zod
schema and the template's restored `Input` component. One rule cannot be followed and is flagged, not worked
around:

- **D1 — search paging in the usecase.** The template pages in SQL (`ORDER BY … LIMIT … OFFSET`, with `Count`,
  `05-repository-database.md`) and the usecase only passes `Page`/`Limit` to the repository filter
  (`04-usecase.md`). Search results must be in the explorer's order (spec FR-008), which is the natural asset-ID
  order of 002-R3, implemented as a pure Go function and not expressible in SQL, with exact matches first. The
  page can therefore only be cut after sorting: the repository returns every match, the usecase orders them and
  takes the page. **Proposed Template Exception 14**:

  > *Search ordering and paging in the usecase*: the asset search loads every match, orders it with the domain's
  > natural asset order (exact matches first) and takes the requested page in the usecase, because that order
  > cannot be expressed in SQL. The API keeps the template's `page`/`limit`/`totalItems` contract. Other paged
  > lists page in SQL.

**Rationale**: Principle I requires any deviation to be listed before it is made. Exception 13 (complete
hierarchy lists) does not cover this: search is a list that *is* paged, as Exception 13 itself says other lists
must be.

**Alternatives**: ordering in SQL with an ICU numeric collation (rejected in 002-R3: a migration, a dependency on
the image's ICU build, and its rules differ from `CompareAssetIDs`, so search and tree would order differently);
SQL order by plain `asset_id` (`SS-10` before `SS-2`, contradicts FR-008); returning every match to the browser
and sorting there (business rule in the frontend, principle II, and up to 50 000 rows on the wire).

## R1. API shape: a search filter on `GET /api/assets`

**Decision**: `GET /api/assets?q=<text>[&assetType=<TYPE>][&page=1&limit=50]`. With `q` the endpoint searches every
level; without `q` it is the hierarchy read of 002 unchanged (top level, or `parentAssetId` children, complete,
Exception 13). Rules, enforced by `GetListAssetInput.Validate` (`400 INVALID_REQUEST`):

- `q` is trimmed; blank after trimming means "no search"; at most 100 characters (rule 07: lengths validated in
  `Validate`).
- `assetType` is one of the five types and is allowed only with `q` (spec FR-005: a type only narrows a text).
- `parentAssetId` together with `q` is refused: a search spans every level.
- `page`/`limit` are allowed only with `q` (the hierarchy lists are not paged); `SetDefault` gives page 1 and
  limit 50 (`usecase.SearchDefaultLimit`), and `limit` is capped at `usecase.MaxLimit` (100).

Response: `{ items, totalItems }` where each item is the 002 asset node plus `ancestorAssetIds` (substation first,
empty for a substation), and `totalItems` is the number of all matches.

**Rationale**: The user asked to extend the existing list read and its filter, and it is the template's way
(`GetListXxx` with filter parameters). The frontend's first page of 50 is exactly spec FR-010 (clarification:
first 50 plus the total, no paging UI); the frontend sends neither `page` nor `limit`, so the 50 lives only in
the backend's default (principle II). The API still accepts `page`/`limit` so it stays a normal template list.

**Alternatives**: a separate `GET /api/assets/search` (a parallel read; its path would also collide with an asset
whose ID is `search` under `/api/assets/:assetId`); a POST search body (not a read in REST terms, not cacheable);
silently ignoring `assetType`/`page` in hierarchy mode (hides client mistakes).

## R2. Matching

**Decision**: `asset_id ILIKE ? OR asset_name ILIKE ?` with the pattern from `querysafety.ContainsPattern(text)`
(escapes `%`, `_`, `\`, so every character matches literally, FR-003), plus `AND asset_type = ?` when a type is
given. Each asset is one row, so it appears once (FR-009). "Exact" (FR-008) is decided in Go:
`strings.EqualFold(text, assetID) || strings.EqualFold(text, assetName)`.

**Rationale**: `ILIKE` gives "contains, letter case ignored" (FR-002) in one statement with bound parameters; the
template already provides the escaping helper for exactly this. At 50 000 rows a sequential scan of two text
columns takes milliseconds (measured by the scale test, R8), so no index is added.

**Alternatives**: a `pg_trgm` GIN index (fast `ILIKE` at millions of rows, but a migration and an extension for no
current need; noted as the scale path in KNOWN_LIMITATIONS); `lower(col) LIKE lower(?)` (same result, more text);
full-text search (`tsvector`) — out of scope per the spec and wrong for IDs like `PNL-001-1-01`.

**Note**: PostgreSQL's case folding (`ILIKE`, database collation) and Go's `EqualFold` (Unicode simple folding)
agree for the ASCII IDs and names in practice; for exotic characters an "exact" decision could differ from the
match. Recorded in KNOWN_LIMITATIONS.

## R3. Ordering and the page cut

**Decision**: A new pure function `entity.SortSearchMatches(nodes []*AssetNode, text string)`: a stable partition
into exact matches and the rest, each part ordered by `SortAssetNodes` (type rank, then natural ID — 002-R3). The
usecase then sets `TotalItems = len(matches)` and returns the page `matches[(page-1)*limit : …]`.

**Rationale**: Reuses the tree's order exactly (FR-008) and is unit-tested table-driven like 002's ordering. A
full ID is unique, so a search for it always puts that asset first (SC-001).

**Alternatives**: rank by match position or length (a relevance model the spec does not ask for); exact matches
only by ID (names typed in full would not come first).

## R4. Ancestors for each result

**Decision**: A new repository method `FindAncestorIDs(ctx, trx, assetIDs []string) (map[string][]string, error)`:
one recursive CTE walking up `parent_asset_id` from all the given assets at once (`= ANY(?)` with the existing
`textArray` parameter), returning, per asset, its ancestor IDs substation first (empty for a substation). Called
only for the page being returned (at most 100 IDs).

**Rationale**: FR-007 shows where a result sits; one query for the page instead of one `FindPath` per result
(N+1). Reuses the array-parameter technique of feature 001 (R6) and the upward CTE of 002-R4.

**Alternatives**: calling `FindPath` per result (50 round trips); a materialised path column (schema and import
change).

## R5. Opening a result reuses feature 002

**Decision**: Each selectable result is a link to `getAssetPath(assetId)` (`/assets/<id>`, no search parameters).
Following it is a normal navigation: the explorer of 002 reads `:assetId`, expands the ancestors
(`useExpandedNodes`), highlights and scrolls (`TreeNode`), and loads the details. Results for `.` / `..`
(`checkIsAddressableAssetId`) are plain text with "cannot be opened".

**Rationale**: FR-011 and the brief: reuse, not duplicate. Leaving the search on selection (clarification) means
the target address is exactly the deep link, so SC-004 holds by construction; Back returns to the search because
the search lives in the address (R6).

**Alternatives**: keeping the results and selecting in place (rejected in clarification); expanding the tree
from the result's `ancestorAssetIds` directly (a second reveal path next to 002's).

## R6. Search state in the address

**Decision**: The search is `?q=<text>` and optionally `&type=<TYPE>` on the asset page (new `SearchParameter.QUERY`
and `SearchParameter.ASSET_TYPE`), read with `useSearchParams` in a page hook `use-asset-search`. Submitting the
form pushes a history entry with the new parameters (other parameters such as `panel` kept); clearing removes
both. While `q` is present the results take the tree's place (clarification); the details area is unchanged.

**Rationale**: Principle V (state in the URL) and FR-006: refresh, shared links and Back show the same search.
Using the same mechanism as the import panel (`usePanel`) keeps the parameters independent.

**Alternatives**: local state (lost on refresh, breaks Back); a separate `/assets/search` route (the brief keeps
the search on the asset page).

## R7. Frontend data and form

**Decision**: `modules/assets` gets a query `searchAssets({ q, assetType })` → `GET /api/assets?q=…` (no `page` or
`limit`: the backend's default page of 50 applies, and the frontend shows "Showing <items received> of
<totalItems>" from the response, so it never holds its own copy of the limit),
`providesTags: [AssetsApiTag.ASSET]` (so a committed import refreshes a shown search, FR-017, through 002's
invalidation), `shouldSuppressToast` (the results area renders failures). DTOs `AssetSearchItemDto` and
`AssetSearchResultDto`; a Zod schema `assetSearchValidationSchema` (text trimmed, 1–100 characters; type one of
the five or `null` for all types). The select's "All types" option is a page value (`AssetTypeFilter.ALL` in
`pages/assets/libs/enums`) that the form turns into `null`. The page's search form uses `useAppForm` with that schema, the template `Input` restored from
git history (below), and a page-specific type field (a native `<select>` through `useController`, like the import
form's mode field). The form's defaults come from the address, so a refresh fills the field.

**Restoring `Input`**: `frontend/src/libs/components/input/{input.tsx,styles.module.css}` and
`frontend/src/libs/enums/input-type.enum.ts` (+ its `enums.ts` export) are restored byte for byte from
`40d511b^` (removed in feature 001's polish as unused), not rewritten; it is a controlled React Hook Form field
with a label, an optional leading icon and an error message.

**Rationale**: The conventions require forms through `useAppForm` + Zod and reuse of `libs/components` before
writing a control (§6, §10); the user asked for the template's component, not a new one. Blank-text handling in
the schema mirrors the backend's `Validate` (the server stays the authority, principle II).

**Alternatives**: a raw `<input>` in the page (forbidden by §10); a new search input component (duplicates the
template's).

## R8. Scale

**Decision**: Extend 002's scale test: with ~55 000 stored assets, a search matching every asset (`q=-`), one
matching a few (`q=WIDE-4999`) and one with a type, each ≤ 1 s on the API side (SC-005), times logged.

**Rationale**: The worst case of R0/R3 is loading and sorting every row; 002 measured a 5 000-row list at ~30 ms,
so ~55 000 matches with child counts and sorting is expected well inside 1 s, but it is measured, not assumed.

**Alternatives**: a database-side limit for very broad searches (would break FR-008's order and the exact
`totalItems`).
