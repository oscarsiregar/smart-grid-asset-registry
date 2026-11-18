# Code conventions

The rules every change in this frontend has to satisfy. Written for anyone working in the code: a developer, an AI
assistant implementing a change, or a reviewer. Each section states the rule, then lists what is allowed (**Do**) and
what is not (**Don't**).

Rules marked 🔒 are enforced by tooling, so breaking them fails `pnpm lint`. Unmarked rules are not checked by any tool,
so they depend on review. Codes such as `B10` or `A9` refer to the rubric in [quality-criteria.md](quality-criteria.md).

---

## 1. Stack

| Concern      | Choice                                        |
| ------------ | --------------------------------------------- |
| Runtime      | Node.js 24, pnpm 11                           |
| Build        | Vite 8, TypeScript 6 (`strict`)               |
| UI           | React 19                                      |
| Routing      | React Router 7 (data router, lazy routes)     |
| Server state | RTK Query (`@reduxjs/toolkit`)                |
| Client state | Redux Toolkit slices, only when truly global  |
| Forms        | React Hook Form + Zod (`@hookform/resolvers`) |
| Styling      | CSS Modules + design tokens in `tokens.css`   |
| Class names  | `clsx`, through `getValidClasses`             |
| Icons        | SVG files imported as components (`?react`)   |

**Don't** add a library that duplicates one above, such as axios next to RTK Query, Formik next to React Hook Form,
styled-components next to CSS Modules, or moment/dayjs where `Intl` is enough. A new dependency needs a reason that
the current stack cannot cover, and it is added in the change that first uses it.

---

## 2. Folder structure

```
src/
├── index.tsx                 Entry point: store provider + route table
├── vite-env.d.ts
├── assets/
│   ├── css/                  styles.css → fonts.css, tokens.css, scaffolding.css (the ONLY global CSS)
│   ├── fonts/                Local woff2 fonts
│   └── img/                  SVG icons and illustrations
├── libs/                     Reusable application infrastructure (no feature knowledge)
│   ├── components/           Shared UI components
│   ├── constants/            App-wide constants + constants.ts barrel
│   ├── enums/                App-wide constant objects + enums.ts barrel
│   ├── helpers/              App-wide pure functions + helpers.ts barrel
│   ├── hooks/                App-wide hooks (imported directly, no barrel)
│   ├── modules/              Infrastructure singletons: api, config, navigation, notification, storage, store
│   └── types/                App-wide types + types.ts barrel
├── modules/                  Feature data layer, one folder per backend resource
│   └── <feature>/
│       ├── <feature>-api.ts  RTK Query endpoints (injectEndpoints)
│       ├── <feature>.ts      Public barrel: types, schemas, hooks other code may use
│       └── libs/{enums,hooks,types,validation-schemas,helpers,constants}/
└── pages/                    Screens, one folder per route
    └── <page>/
        ├── <page>.tsx        Page component (composes, coordinates)
        ├── styles.module.css
        ├── components/       Components used only by this page (may nest further)
        └── libs/{enums,constants,helpers,hooks,types}/
```

### Who owns what

| Layer                     | Owns                                                                       | Must not                                                                                     |
| ------------------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| `pages/<page>`            | Screen layout, coordinating feature behavior, page copy                    | Define API endpoints or DTOs                                                                 |
| `pages/<page>/components` | UI used by one page                                                        | Be imported by another page (promote it to `libs/components`)                                |
| `libs/components`         | Generic, reusable UI with no feature knowledge                             | Import from `modules/*` or `pages/*` (exceptions: app shell, private route, which wire auth) |
| `modules/<feature>`       | Endpoints, DTO types, validation schemas, feature hooks, API tags          | Render JSX or own styles                                                                     |
| `libs/modules`            | Infrastructure singletons (store, base API, config, storage, notification) | Know about any feature                                                                       |

**Do**

- Put a component in `pages/<page>/components` first. Promote it to `libs/components` only when a second page needs it.
- Nest components inside the component that owns them (`notes-list/components/note-item`) when nothing else uses them.

**Don't**

- Create `src/components`, `src/utils`, `src/services`, `src/api`, or any folder outside the layout above.
- Import one page from another page.
- Put feature logic (endpoints, DTOs) in `libs/`.

---

## 3. Files and naming

### 3.1 File names 🔒

File and folder names are kebab-case (`ls-lint`). A file carries the suffix of its role:

| Suffix                   | Contains                                   | Example                          |
| ------------------------ | ------------------------------------------ | -------------------------------- |
| `*.tsx`                  | One React component, named after the file  | `note-item.tsx` → `NoteItem`     |
| `*.type.ts`              | One type                                   | `note-dto.type.ts` → `NoteDto`   |
| `*.enum.ts`              | One constant object (`as const`)           | `app-route.enum.ts` → `AppRoute` |
| `*.constant.ts`          | One constant                               | `bearer.constant.ts` → `BEARER`  |
| `*.helper.ts`            | One function                               | `format-note-date.helper.ts`     |
| `*.hook.ts`              | One hook                                   | `use-sign-out.hook.ts`           |
| `*.validation-schema.ts` | One Zod schema                             | `sign-in.validation-schema.ts`   |
| `*.module.ts`            | A class implementing infrastructure        | `base-storage.module.ts`         |
| `*.slice.ts`             | A Redux slice                              | `navigation.slice.ts`            |
| `*-api.ts`               | RTK Query endpoints of one feature         | `notes-api.ts`                   |
| `styles.module.css`      | Styles of the component in the same folder |                                  |
| `maps.ts`                | Lookup objects keyed by an enum            | `notificationTypeToIconName`     |

### 3.2 One declaration per file

A declaration lives in the file its role names. This applies to every size of declaration, including a single
number, regex, or string. Nothing checks it automatically, so it is enforced in review.

**Do**

- One exported type per `*.type.ts`, one constant per `*.constant.ts`, one constant object per `*.enum.ts`.
- Collect them through a barrel in the same folder: `constants.ts`, `enums.ts`, `types.ts`, `helpers.ts`,
  `validation-schemas.ts`.
- Keep a private helper used by one function in the same file, unexported (see `to-server-error.helper.ts`).

**Don't**

- Put two exported types in one `*.type.ts`, or three constants in one `constants.ts`.
- Declare a type, enum, or constant inside a component file "because only this component uses it". Use the
  component's `libs/` folder.
- Keep a set of user-facing messages as loose strings. They go into an enum of their own (`NotesPageMessage`).

### 3.3 Barrels

**Do**

- Create barrels only at `libs/<category>/<category>.ts` (`enums.ts`, `types.ts`, ...) and at a feature's public
  file `modules/<feature>/<feature>.ts`.
- Re-export names explicitly: `export { AppRoute } from "./app-route.enum.js";`.

**Don't**

- Create `index.ts` files (B39). The only `index` file is `src/index.tsx`.
- Use `export *` or `import * as` 🔒.
- Create a barrel for `libs/hooks` or `libs/components`. Import each hook or component from its own file.

### 3.4 Identifier naming

| Kind                           | Style                                     | Example                                        |
| ------------------------------ | ----------------------------------------- | ---------------------------------------------- |
| Variable, function, property   | camelCase, starts with a lowercase letter | `formatNoteDate`, `isLoading`                  |
| Function                       | Starts with a verb (B20)                  | `getValidClasses`, `checkIsAuthPath`           |
| Boolean                        | `is`/`has`/`should`/`can` prefix (B19)    | `isRequired`, `hasError`, `shouldOpenInNewTab` |
| Boolean-returning function     | `check` prefix 🔒                         | `checkIsServerError`                           |
| Array                          | Plural noun (B18)                         | `notes`, `NAVIGATION_ITEMS`                    |
| Constant                       | `UPPER_SNAKE_CASE` (B22)                  | `EMPTY_LENGTH`                                 |
| Constant object ("enum")       | PascalCase name, `UPPER_SNAKE` keys (B23) | `AppRoute.SIGN_IN`                             |
| Component, class, type         | PascalCase                                | `NoteItem`, `BaseStorage`, `NoteDto`           |
| Component props type           | Always `Properties`                       | `type Properties = { ... }`                    |
| Event handler inside component | `handle` + subject + event (A10)          | `handleDeleteClick`, `handleFormSubmit`        |
| Callback prop                  | `on` + event (A9)                         | `onClick`, `onSubmit`                          |
| Route with a parameter         | `$` before the parameter                  | `PROMPTS_$PROMPT_ID: "/prompts/:promptId"`     |
| Acronyms                       | First acronym uppercase (B21, B23)        | `HTTPCode`, `APIPath`                          |

**Don't** 🔒

- Abbreviate: `properties` not `props`, `parameters` not `params`, `reference` not `ref`, `error` not `err`,
  `event` not `e`, `response` not `res`, `arguments_` not `args`.
- Put the data type in the name (B17): `notes`, not `notesArray`.
- Repeat the module name inside it (A5): inside `storage`, `get()` not `getStorage()`.
- Use transliterated or non-English names (B25).

---

## 4. Exports and imports 🔒

**Do**

- Put every export at the end of the file: write `const value = …;`, then `export { value };`.
- Use named exports only.
- Write the `.js` extension on relative and alias imports, including imports of `.ts` and `.tsx` files:
  `import { Button } from "~/libs/components/button/button.js";`.
- Use the `~/` alias for imports that cross into another top-level tree (`libs`, `modules`, `pages`).
- Use relative imports within the same feature, page, or component.
- Mark type imports inline: `import { type NoteDto } from "…";`.
- Let the linter sort imports, keys, members, and exports. Run `pnpm exec eslint . --fix` and `pnpm format`.

**Don't**

- Use `export const x = …` or `export default` (config files at the root are the only exception).
- Import `useDispatch` or `useSelector` from `react-redux`. Use `useAppDispatch` / `useAppSelector`.
- Use a relative path that climbs out of the current feature (`../../../libs/...`). Use `~/`.

---

## 5. TypeScript 🔒

The compiler runs with `strict`, `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`, and
`noPropertyAccessFromIndexSignature`. ESLint runs `strict-type-checked` with zero warnings allowed.

**Do**

- Model a fixed set of values as a constant object with `as const`, and its type as `ValueOf<typeof X>`.
- Use `satisfies` to check a lookup object covers every key: `… satisfies Record<ValueOf<typeof NotificationType>, …>`.
- Treat indexed access as possibly `undefined`: `const [first] = items; if (first) { … }`.
- Read CSS-module classes with bracket access: `styles["card"]`.
- Write an optional prop that may receive `undefined` as `className?: string | undefined`.
- Prefer `unknown` and narrow it with a `check…` type guard.
- Name every number, string, or regex with meaning. Use `EMPTY_LENGTH` rather than `0` when checking a length (B31).

**Don't**

- Use the `enum` keyword or `abstract class`.
- Use `any`, non-null assertions (`!`), or `// @ts-ignore` / `eslint-disable` comments (B26).
- Write magic numbers. `HTTPCode.NOT_FOUND`, not `404`.
- Write a function with more than three parameters. Pass an options object instead.
- Use `console`. Show the user a notification, or let the error reach the error boundary.

---

## 6. React components

### 6.1 Shape

```tsx
import { getValidClasses } from "~/libs/helpers/helpers.js";

import styles from "./styles.module.css";

type Properties = {
	className?: string | undefined;
	isActive?: boolean;
	label: string;
	onClick: () => void;
};

const Chip: React.FC<Properties> = ({
	className,
	isActive = false,
	label,
	onClick,
}: Properties) => (
	<button
		className={getValidClasses(
			styles["chip"],
			isActive && styles["active"],
			className,
		)}
		onClick={onClick}
		type="button"
	>
		{label}
	</button>
);

export { Chip };
```

**Do**

- Type components as `React.FC<Properties>`. Generic form controls (`Input`, `Textarea`) use a generic function
  returning `React.JSX.Element`.
- Destructure props in the parameter list and give defaults there.
- Accept `className` on reusable components and merge it last with `getValidClasses`.
- Wrap every handler passed to a JSX prop in `useCallback`, or hoist it out of the component 🔒 (`react/jsx-no-bind`).
- Give a `<button>` an explicit `type`.
- Render the relevant loading, error, empty, and success states of every data-driven view.
- Return early for loading and error states before the main markup.
- Compose the page from small components. A component that passes ~150 lines or mixes concerns gets split (A16).

**Don't**

- Pass inline arrow functions to JSX props: `onClick={() => …}` 🔒.
- Use `useEffect` to fetch data, or to copy query data into `useState`.
- Use `useEffect` for something that can be derived during render.
- Add wrapper `<div>`s for decoration only (B4). Use a pseudo-element or the existing element.
- Use array index as `key` when items have an id.
- Read `localStorage`, `window.location`, or `document` directly in a component. Go through `libs/modules`.

### 6.2 Hooks

**Do**

- Name hooks `use…` and put each in its own `*.hook.ts` folder (`libs/hooks/use-debounce/use-debounce.hook.ts`).
- Place a hook where it is used: `libs/hooks` if app-wide, `modules/<feature>/libs/hooks` if it belongs to a
  feature, `pages/<page>/libs/hooks` if only one page needs it.
- Return a stable API: memoize returned callbacks with `useCallback`.
- Clean up every timer, listener, and subscription in the effect's return (B40).

**Don't**

- Call hooks conditionally 🔒.
- Silence `react-hooks/exhaustive-deps` 🔒. Fix the dependency instead.

### 6.3 Accessibility 🔒 (partly)

**Do**

- Label every input (`Input` and `Textarea` do this; use `isLabelHidden` to hide it visually only).
- Give icon-only buttons an `aria-label` or a `.visually-hidden` text (A2).
- Mark decorative icons `aria-hidden` (the `Icon` component does).
- Fill `alt` on content images, and use `alt=""` on decorative ones (B5).
- Use semantic elements: `main`, `nav`, `section`, `header`, `ul/li`, `time`, headings in order.
- Keep keyboard interaction and a visible `:focus-visible` outline.

**Don't**

- Put click handlers on a `div` or `span`. Use a `button` or a `Link`.
- Remove focus outlines without a replacement.

---

## 7. Styling

### 7.1 Where styles live

| File                            | Contains                                                      |
| ------------------------------- | ------------------------------------------------------------- |
| `assets/css/tokens.css`         | Design tokens as CSS custom properties on `:root`             |
| `assets/css/fonts.css`          | `@font-face` rules for local fonts (B13)                      |
| `assets/css/scaffolding.css`    | Reset, `body`, `#root`, `.page-container`, `.visually-hidden` |
| `<component>/styles.module.css` | Everything else, scoped to the component                      |

**Do**

- Style through class selectors in CSS Modules only (B10).
- Use tokens for every color, space, radius, font size, and duration: `var(--space-16)`, `var(--text-muted)`.
- Lay out with `flex` or `grid` (B8).
- Use `px` units. 🔒 `em` and `rem` are disallowed.
- Write a long hex color in full (`#ffffff`) 🔒.
- Guard animation with `prefers-reduced-motion`.
- Keep interaction states (hover, focus, active) from moving the element or its neighbors (B12).
- Give fonts a fallback chain: main font, safe font, generic family (B14).

**Don't**

- Use `!important` 🔒 (B9).
- Nest selectors 🔒 (`max-nesting-depth: 0`). Use at most two levels of descendant selectors (B11).
- Style tags or ids in a module (`.list li`, `#title`). Give the element a class.
- Add a new global class outside `scaffolding.css`.
- Hard-code a color or spacing value that a token already covers. Add a token when a value is reused.
- Use inline `style={{ … }}` for static styling.

### 7.2 Property order 🔒

`stylelint-config-recess-order` sorts declarations: positioning, box model, typography, visual, misc. `pnpm lint:css`
reports violations; `stylelint --fix` repairs them.

---

## 8. Server state (RTK Query)

All server communication goes through one `baseApi` (`libs/modules/api/base-api.ts`). Features add endpoints to it
with `injectEndpoints`.

```ts
const notesApi = baseApi
	.enhanceEndpoints({ addTagTypes: [NotesApiTag.NOTE] })
	.injectEndpoints({
		endpoints: (builder) => ({
			getNotes: builder.query<NoteDto[], undefined>({
				providesTags: [NotesApiTag.NOTE],
				query: () => ({ url: APIPath.NOTES }),
			}),
		}),
	});
```

**Do**

- Put endpoints in `modules/<feature>/<feature>-api.ts` and export only the generated hooks.
- Declare tags in a `<Feature>ApiTag` enum. Queries `providesTags`, mutations `invalidatesTags`.
- Build URLs from `APIPath` plus a feature `<Feature>ApiPath` enum. Fill parameters with `configureString`.
- Pass `extraOptions: { shouldSuppressToast: true }` on mutations whose error the UI renders itself (forms).
- Use `skipToken` or `skip` instead of calling a query with incomplete arguments.
- Use `onQueryStarted` + `upsertQueryData` / `updateQueryData` for optimistic or derived cache updates.
- Call `baseApi.util.resetApiState()` when the session ends (see `useSignOut`).

**Don't**

- Call `fetch`, axios, or `XMLHttpRequest` directly.
- Create a second `createApi`.
- Copy query data into component state or a slice.
- Store server data in a Redux slice. Slices are for client-only global state, such as the pending redirect.
- Invalidate everything to "be safe". Invalidate the tags the mutation actually changes.

---

## 9. Client state

**Do**

- Keep UI state local (`useState`) unless more than one distant part of the tree needs it.
- Use a Redux slice in `libs/modules/<name>/<name>.slice.ts` for app-wide client state, registered in `store.module.ts`.
- Read and write it through `useAppSelector` / `useAppDispatch`.
- Put state that should survive a reload in the URL (search parameters) or in `storage`.

**Don't**

- Use React context for state that changes often across the app. Context is fine for a stable dependency, such as
  a provider-scoped API.
- Use `localStorage` directly. Use `storage` with a key from `StorageKey`.

---

## 10. Forms

Forms use `useAppForm` (React Hook Form + Zod) and the controlled `Input` / `Textarea` components.

```tsx
const { clearErrors, control, handleSubmit, setError } =
	useAppForm<SignInRequestDto>({
		defaultValues: DEFAULT_SIGN_IN_PAYLOAD,
		validationSchema: signInValidationSchema,
	});

const { hasFieldErrors } = useServerFormErrors({
	clearErrors,
	error,
	fields: SIGN_IN_FIELDS,
	setError,
});
const generalError =
	hasFieldErrors || checkIsToastedError(error) ? undefined : error;
```

**Do**

- Define the Zod schema in `modules/<feature>/libs/validation-schemas/<name>.validation-schema.ts`, export it from
  the barrel as `<name>ValidationSchema`, and derive the request type with `z.infer`.
- Keep validation limits in a `<Feature>ValidationRule` enum and messages in a `<Feature>ValidationMessage` enum.
- Keep the default payload and the list of server-mappable fields as constants in the form's `libs/constants`.
- Map server validation errors onto fields with `useServerFormErrors`; show anything else with `FormAlert`.
- Use `noValidate` on `<form>` so Zod, not the browser, owns validation messages.
- Disable or show loading on the submit button while the mutation runs.
- Catch a rejected `.unwrap()` when you await it; the error is already in the mutation state.

**Don't**

- Use uncontrolled `register` and raw `<input>` elements in pages.
- Validate by hand in the submit handler.
- Show the same error twice: an error already shown as a toast (`checkIsToastedError`) is not repeated in `FormAlert`.

---

## 11. Errors and notifications

`baseQuery` (`libs/modules/api/libs/helpers/base-query.helper.ts`) normalizes every failure into a `ServerError` and
handles the cross-cutting cases:

| Server error code       | What happens                                                                    |
| ----------------------- | ------------------------------------------------------------------------------- |
| `UNAUTHENTICATED`       | Token dropped, cache reset, redirect to sign in (toast only if a token existed) |
| `FORBIDDEN`             | Toast, then redirect to `/no-access`                                            |
| `INTERNAL_SERVER_ERROR` | Toast (also used for network, timeout, and parse failures)                      |
| `VALIDATION_FAILED`     | Nothing global; the form maps `details` onto fields                             |
| Anything else           | Toast, unless the endpoint set `shouldSuppressToast`                            |

The backend is expected to answer errors as `{ code, message }`, plus `details: [{ path, message }]` for
`VALIDATION_FAILED`. Adjust `ErrorCode` and `toServerError` if your API differs.

**Do**

- Use `showNotification({ message, type })` for feedback outside React or after an action.
- Narrow unknown errors with `checkIsServerError`, `checkIsValidationError`, `getErrorMessage`.
- Let unexpected render errors reach the route `errorElement` (`ErrorPage`).

**Don't**

- Handle 401/403 in a page. `baseQuery` already does.
- Swallow an error without the UI showing it somewhere.

---

## 12. Routing and navigation

**Do**

- Add every path to `AppRoute`. Reference routes only through it.
- Register pages in `src/index.tsx` as `lazy` routes, so each page is its own chunk.
- Put signed-in pages under `PrivateRoute` → `AuthenticatedShell`.
- Navigate in components with `Link`, `ButtonLink`, `Navigate`, or `useNavigate`.
- Navigate from outside React (API layer) by dispatching `setRedirect`; `useRedirect` in `App` performs it.
- Type link targets as `NavigableRoute` so `AppRoute.ANY` cannot be linked to.

**Don't**

- Write path strings inline (`to="/notes"`).
- Use `<a href>` for internal navigation.
- Use `window.location` to navigate.

---

## 13. Configuration and environment

**Do**

- Prefix every environment variable with `VITE_APP_` and document it in `.env.example`.
- Read it in one place, `BaseConfig.envSchema`, and add it to `EnvironmentSchema`. Use it through `config.ENV`.

**Don't**

- Read `import.meta.env` anywhere else.
- Put secrets in `VITE_APP_*`. Everything prefixed with `VITE_` is shipped to the browser.
- Commit `.env`.

---

## 14. Assets and icons

**Do**

- Add an icon by placing the SVG in `assets/img`, adding a key to `IconName`, and mapping it in
  `libs/components/icon/libs/maps.ts`. Render it with `<Icon iconName={IconName.X} />`.
- Import illustrations as URLs: `import illustration from "~/assets/img/x.svg";`.
- Use SVG for vector graphics (B7). Inline SVGs need a size (B6).
- Load fonts locally from `assets/fonts` (B13).

**Don't**

- Use PNG/JPG for something that exists as SVG.
- Load fonts or icons from a CDN.

---

## 15. Security

**Do**

- Render user content as text (React escapes it). If HTML or Markdown rendering is ever needed, sanitize it (B43).
- Open external links with `NEW_TAB_LINK_ATTRIBUTES` (`rel="noreferrer"`).
- Keep the token in `storage` under `StorageKey.TOKEN`, and drop it on sign-out and on 401.

**Don't**

- Use `dangerouslySetInnerHTML` with unsanitized input.
- Log tokens, passwords, or personal data.

---

## 16. Tooling and verification

| Command            | Checks                                         |
| ------------------ | ---------------------------------------------- |
| `pnpm lint:type`   | TypeScript                                     |
| `pnpm lint:js`     | ESLint, zero warnings                          |
| `pnpm lint:css`    | Stylelint                                      |
| `pnpm lint:format` | Prettier                                       |
| `pnpm lint:fs`     | File and folder names (ls-lint)                |
| `pnpm lint:editor` | `.editorconfig` (tabs, LF, final newline)      |
| `pnpm lint:trash`  | Unused files, exports, and dependencies (knip) |
| `pnpm lint`        | All of the above                               |
| `pnpm build`       | Type check + production bundle                 |
| `pnpm format`      | Rewrites formatting                            |

**Do**

- Run `pnpm lint` and `pnpm build` before calling a change done.
- Pin exact dependency versions (B32). `saveExact` in `pnpm-workspace.yaml` does this for `pnpm add`.
- Remove code, files, and dependencies that nothing uses (B33, B34). `knip` reports them.
- Write commit messages in Conventional Commits form (`feat: add notes page`); `commitlint` checks it.

**Don't**

- Disable a lint rule in source (B26), or skip hooks with `--no-verify`.
- Leave commented-out code.

There is no test runner configured. Do not claim behavior is tested when only lint and build ran.

---

## 17. Review checklist

- [ ] Every new file is kebab-case, has a role suffix, and holds one declaration.
- [ ] New constants, types, and enums sit in the owning `libs/` folder behind a barrel.
- [ ] No magic numbers or loose user-facing strings.
- [ ] Imports use `~/` across trees, `.js` extensions, inline `type`.
- [ ] Server data comes from RTK Query; tags are provided and invalidated correctly.
- [ ] Forms use `useAppForm` + a Zod schema; server field errors are mapped.
- [ ] Loading, error, empty, and success states are rendered.
- [ ] Handlers are memoized; no inline functions in JSX props.
- [ ] Styles are in a CSS module, use tokens, no nesting, no `!important`.
- [ ] Interactive elements are labelled and keyboard reachable.
- [ ] New routes are in `AppRoute` and registered lazily.
- [ ] `pnpm lint` and `pnpm build` pass.
