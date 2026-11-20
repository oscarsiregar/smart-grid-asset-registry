# Quickstart: validating the UI polish

**Spec**: [spec.md](spec.md) · **UI contract**: [contracts/ui-asset-page.md](contracts/ui-asset-page.md)

## Prerequisites

The stack running with the sample imported (as in features 001–003). If ports 3000/8080/5432 are taken by
another local stack, start this one on others:

```sh
DB_HOST_PORT=55434 API_HOST_PORT=18080 WEB_HOST_PORT=13000 docker compose up --build -d --wait
```

The reference screenshots are the ones the user supplied on 2026-11-20.

## 1. Automated checks

```sh
cd frontend && pnpm lint && pnpm build && pnpm test     # each exit 0
cd backend && make check && make lint                   # unchanged backend, still exit 0
```

## 2. Side by side with the reference (SC-001), 1920 × 940

1. `/assets/SS-001`: brand with the bolt tile; outlined Import CSV; compact search without visible labels;
   tree rows with a status shape and the name only; "Register — 13 substations" at the bottom of the
   sidebar; header "Asset explorer" / "Aurora Substation (North)"; details card without ID/type/name/status;
   counts after it; "Immediate children (5)" with column headings.
2. `/assets/TX-001-1`: path "Aurora Substation (North) /"; the details card starts with the parent link;
   "Immediate children" says nothing is connected below.
3. `/assets/SWB-001-2`: its panels under the column headings.
4. `?panel=import`: title, subtitle, icon close button, steps with "Choose file" current. Check the sample:
   "Review" current, banner at the top, footer with **Save … assets** / **Choose another file** that never
   covers a row while the rejected rows scroll. Save (on a fresh database) or choose a file that saves
   nothing: "Done" current.
5. `/assets?q=aurora&type=LV_BOARD`: results in the sidebar, the register card still at the bottom.
6. At 390 px wide: the menu button opens the sidebar with the register card at its bottom; the children
   rows stack; the import steps wrap.

## 3. Loading (SC-002, SC-004)

In the browser's developer tools set the network to "Slow 4G" and disable the cache:

1. Reload `/assets/SS-001`: tree rows, the register number and the details show placeholders of their final
   size; when the data arrives nothing below them moves.
2. Expand an unvisited switchboard: placeholder rows at the child indent.
3. Search `pnl`: placeholder result rows.
4. Open the import drawer and check the sample: placeholders under the form until the result shows.
5. With the network back to "No throttling": no placeholder flashes on any of the above.
6. With "Emulate CSS prefers-reduced-motion: reduce": placeholders do not pulse.
7. Stop the `api` container and reload: each area shows its error with **Try again**, no placeholder stays.

## 4. Accessibility (SC-005)

1. Greyscale (rendering: "Emulate vision deficiency: achromatopsia"): status shapes, the selected row (bar +
   bold), the current step (outline + number) and every banner remain distinguishable.
2. Screen reader or the accessibility tree: a tree node reads its name, ID, type and status; a path step
   reads its name and ID; the close button reads "Close"; loading areas announce their loading text.
3. Keyboard only: the tree, the search, the drawer (Tab order, Escape closes, focus returns to Import CSV).
