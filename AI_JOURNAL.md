# AI Journal

A short, honest record of how I used an AI coding agent on this project.

## Tools and models

- **Claude Code** (VS Code extension) with **Claude Opus** as the model.
- **Spec Kit** for the workflow: constitution, then specify → clarify → plan → tasks → analyze → implement for each feature.
- A frontend design plugin for Claude Code during the UI restyle.

I usually wrote to the agent in Indonesian and asked it to keep every file in the repository in English.

## How I split the work

I started with the constitution: the rules from the company templates, the few deliberate exceptions
(no auth, frontend tests, a local Docker stack, no CHECK constraints) and a Definition of Done.

After that I built one feature at a time, each with its own spec and task list in `specs/`:

1. **001 CSV import**: parsing, validation, partial and strict mode, one transaction.
2. **002 Asset explorer**: the tree, asset details and a stable URL per asset.
3. **003 Asset search**: search by ID or name, with an optional type filter.
4. **004 UI design**: restyling to a design reference, with no change in behaviour.
5. **005 UI polish**: a closer match to the reference screenshots, and loading placeholders so the page no
   longer jumps while data arrives.

After the first review I also handled the reviewers' comments on a separate branch (see "After the review").

Each feature ended with checkpoints that I ran myself in the browser and against the API before moving on.

## Decisions I made along the way

During each `clarify` step the agent asked me questions and usually suggested an answer. Most of these were
mine to decide, and I didn't always take the suggestion. The full list is in the "Clarifications" section of
each spec. The ones that shaped the app most:

**Import and validation**

- **All rules live in the backend.** The database has only primary keys, NOT NULL and the foreign keys,
  including the self-reference from parent to asset. It has no CHECK constraints or triggers. I wanted one
  place where the rules live and one place that explains a rejection.
- **The validator is pure.** It gets the parsed rows and a plain map of the assets already stored, and never
  touches the database or HTTP. This keeps it easy to test, and lets the result not depend on row order:
  hierarchy checks only run after the whole file is indexed.
- **Check first, then commit.** There are two endpoints: one only validates, and the other validates again
  and saves in a single transaction. The server keeps no preview state between the two calls.
- **Report every broken hierarchy rule on a row**, not just the first one, so the user can fix a row in one go.
- **A blank operational status is allowed** and stored as "no status". Only an unknown value is rejected.
- **A broken CSV is a file error, not a row error.** A broken quote or a repeated column name refuses the whole
  file. A row with the wrong number of fields is reported only as `MALFORMED_ROW`, because its other values no
  longer line up with the columns.
- **Store a short summary of each import** (time, file, mode, counts, outcome), not the full list of rejected
  rows. That is enough for an audit trail and keeps the schema small.

**Explorer and search**

- **Show all children at once, with no paging.** I checked that the API stays fast at 50 000 assets, and
  wrote down the browser-side limit in `KNOWN_LIMITATIONS.md`.
- **Sort IDs naturally**, so `SS-2` comes before `SS-10`.
- **Only show a status in the tree when it is not "in service"**, to keep the tree quiet. I changed this in
  005: every node now shows a small status shape and its name, as in the reference. The asset ID, type and
  status in words stay available to screen readers and on hover.
- **Search runs on Enter or on the button**, not while typing. It shows the first 50 matches with the total
  count, and the results replace the tree. Selecting a result goes to the asset's own URL, and Back returns
  to the results.

**UI polish (005)**

- **Appearance only.** No new endpoint and no import history, even though the reference shows "recent
  imports". The register card counts substations from data the page already loads.
- **Placeholders, not spinners.** Each area that waits for data shows grey shapes the size of its content. They
  appear only on a first load and only after 200 ms, so a fast response never flashes. With reduced motion
  turned on they don't pulse.
- **Keep our own texts.** I kept "Not recorded" for blank values and `2010-03-12` dates instead of the
  reference's "—" and "Mar 12, 2010".
- **The path shows names**, with the asset ID for screen readers and on hover.
- **The tab icon matches the app**: the amber bolt on a dark tile, replacing the template's icon.

**Assumptions about the brief**

- **The sample does not match the brief's numbers.** The brief calls it a valid dataset of 220 assets, but the
  file has 222 rows, including some that are clearly meant to be invalid. My rules give 205 accepted and 17
  rejected, and the 13 substations match. I kept the result that follows from the rules, pinned it in a
  golden test, and documented the difference instead of bending the rules to hit 220.

## After the review

The reviewers left three comments, which I handled on a separate branch:

- **A unit test report at the root.** `UNIT_TEST_REPORT.md` is generated from real test runs: the counts and
  the coverage per package, the integration tests and the frontend tests, plus the commands to reproduce it.
- **Remove an unneeded file.** `backend/domain/service/doc.go` was an empty package left from the template.
- **The backend rule files.** `backend/docs/rules/` came from my own company template, which is written in
  Indonesian. At first I wanted to keep them as they were, because the template is a separate project. Our
  Technical Manager asked for English so the client can review them, so I had the agent translate every rule
  file and remove the remaining Indonesian examples. Wherever this project deliberately differs from the
  template, the translation adds a short "This project:" note pointing to the exception in the constitution.

## Three representative prompts

These are translated from Indonesian and slightly shortened.

1. _"Fix the design so it looks like our reference. Layout and styling only: same features, same behaviour.
   Write a short spec and tasks first, then implement it, and I'll check it in the browser."_
2. _"The docker-compose file should only run the frontend and the backend. The test database and the other
   things we don't need can go."_
3. _"Check whether there is any seeder or migration that inserts data. There must not be one, because I want
   to test locally from an empty database."_

## Suggestions I rejected or corrected

- **Extra validation rules.** The agent suggested also rejecting future commissioned dates and negative
  voltages. I said no: they are not in the brief, and inventing rules would reject rows the client may
  consider valid. They are listed as a possible next step in `KNOWN_LIMITATIONS.md`.
- **"The brief requires Docker Compose."** The agent said this when I asked it to remove the compose file. I
  checked the brief: it asks for a _containerized setup, run and test instructions_, not for Compose in
  particular. I kept Compose but slimmed it down. The test profile now runs only the unit tests, and the
  separate test database and its init script are gone.
- **AI attribution in commits.** The agent added `Co-Authored-By` lines to its commits. I had them removed.
  This journal is where the AI involvement is described.
- **A misleading code comment.** A comment in the golden test said the expected values were "derived by
  hand". I had it reworded to say what is actually true: the values follow the spec's rules for each row.
- **Placeholders that stayed after an error.** A test written first showed that the register card kept its
  placeholder when loading failed. The card now hides on an error.
- **Too much at once.** An earlier attempt had very long specs and a lot of generated test code. I threw it
  away and restarted with smaller specs, and tests only for the core import and hierarchy behaviour.

## How I reviewed and verified the code

- I read the diff before every commit, and only committed when I was happy with it.
- I ran the checks myself: `make check`, `make lint`, `pnpm lint`, `pnpm build` and `pnpm test`, and the
  same tests through `docker compose --profile test`.
- I tested every feature by hand on a fresh database:
  - imported `grid_assets.csv` myself and checked the counts (205 accepted, 17 rejected);
  - tried strict mode and re-imported the same file;
  - browsed the tree, used direct URLs and refresh;
  - searched by ID, name and type.
- For the restyle, I compared screenshots side by side with the reference at desktop and phone widths.

## Where the agent struggled

- **It stated assumptions as facts.** The Docker Compose claim above is one example. I learned to check the
  brief or the code myself before accepting this kind of statement.
- **It tended to over-build.** It wrote more tests, documentation and process than the task needed, unless I
  kept the scope tight.
- **It needed context about our templates.** It didn't know the company template conventions (folder layout,
  one component per file, design tokens). I had to point it at the template and the rules before it followed
  them consistently.
- **It missed visual bugs in the code.** Two layout bugs only became visible in screenshots: the import drawer
  was clipped by the scrollbar gutter, and status badges were squeezing the asset names in the tree. The
  browser tab also still showed the template's icon until I noticed it myself.
- **It lost track of earlier work.** At the start of 005 it said nothing from the reference existed yet, while
  much of it had already been built in 004. It corrected itself after reading the code.

## Time

About **20 hours** of my own working time, not counting the time the agent ran unattended.

I could only give this about 4 hours a day, outside working hours, because I couldn't work on it on my
office laptop.
