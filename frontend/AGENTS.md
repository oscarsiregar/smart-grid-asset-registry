# Agent guidelines

Instructions for an AI assistant working in this frontend.

> **Project-specific:** this app has no authentication (an approved exception to the template, see
> `.specify/memory/constitution.md` → *Template Exceptions* no. 1). There is no sign-in, no private routes and no auth state;
> template auth pieces exist only until they are removed.

## Working approach

- Read [docs/code-conventions.md](docs/code-conventions.md) before writing code. It is the source of truth; follow it
  over general habits.
- Follow the closest existing example before inventing a pattern. For a new feature, copy the shape of `modules/notes`
  and `pages/notes` as described in [docs/adding-a-feature.md](docs/adding-a-feature.md).
- Make the smallest coherent change that solves the request. No unrelated refactoring.
- Ask before adding a dependency, changing the API contract, changing authentication, or changing tooling.
- State any low-risk assumption you make.

## Architecture in one paragraph

Pages in `src/pages` compose screens. Components render UI; shared ones live in `src/libs/components`, page-specific
ones in the page's `components` folder. `src/modules/<feature>` holds a feature's RTK Query endpoints, DTO types, and
Zod schemas. `src/libs` holds app-wide infrastructure: the store, the base API with central error handling, config,
storage, notifications, and shared enums, types, helpers, and hooks.

## State and data

- RTK Query for every server request, with tags for caching and invalidation.
- No manual fetching in `useEffect`, and no copying query data into local state.
- Local UI state stays local. Redux slices only for client-only global state.
- React Hook Form through `useAppForm`, with the feature's Zod schema.

## User interface

- Reuse `libs/components` before writing a new control.
- Semantic HTML, labelled controls, working keyboard interaction.
- Handle loading, error, empty, and success states.

## Verification

Run `pnpm lint`, `pnpm build` and `pnpm test`. Report the exact commands and their exit codes. A non-zero exit is a
failure. Tests use Vitest + React Testing Library with `fetch` mocked, and sit next to the code they cover (template exception no. 2:
the template had no test runner). For UI changes, name the route and the states that should be
checked by hand.

## Conventions

@docs/code-conventions.md
