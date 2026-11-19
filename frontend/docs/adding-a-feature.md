# Adding a feature

A step-by-step recipe for a new backend resource with its own page. The reference is the CSV import feature:
`src/modules/imports` (data layer) and `src/pages/assets` (page and the import panel). Replace `<feature>` with a
kebab-case plural (`projects`) and `<Feature>` with its PascalCase form (`Projects`).

## 1. API path

Add the resource root to `src/libs/enums/api-path.enum.ts`:

```ts
const APIPath = {
	IMPORTS: "/imports",
	PROJECTS: "/projects",
} as const;
```

## 2. Feature module (`src/modules/<feature>/`)

The `imports` module looks like this:

```
modules/imports/
├── imports-api.ts
├── imports.ts
└── libs/
    ├── enums/
    │   ├── enums.ts
    │   ├── import-form-field.enum.ts
    │   ├── import-mode.enum.ts
    │   ├── import-outcome.enum.ts
    │   ├── import-validation-message.enum.ts
    │   └── imports-api-path.enum.ts
    ├── helpers/
    │   ├── helpers.ts
    │   └── create-import-form-data.helper.ts (+ .test.ts)
    ├── types/
    │   ├── types.ts
    │   ├── import-request-dto.type.ts
    │   ├── import-result-dto.type.ts
    │   └── …one file per DTO type
    └── validation-schemas/
        ├── validation-schemas.ts
        └── import-file.validation-schema.ts
```

1. **Path enum.** Sub-paths relative to the root: `ImportsApiPath = { VALIDATE: "/validate" } as const`.
2. **Tag enum**, only when the feature has queries: `ProjectsApiTag = { PROJECT: "project" } as const`.
3. **Validation message enum** (and a rule enum when there are limits).
4. **Schema.** Exported from the barrel as `<name>ValidationSchema`. Keep business rules on the server; the import
   schema only checks that a file is chosen.
5. **Types.** Response DTOs are written by hand to match the API contract (`specs/<feature>/contracts/`). The
   request DTO is `z.infer` of the schema.
6. **Endpoints** in `<feature>-api.ts`. Copy the shape of `imports-api.ts`: `baseApi.injectEndpoints` (plus
   `enhanceEndpoints` with the tags when there are queries). Pass `extraOptions: { shouldSuppressToast: true }` when
   the page renders the error itself. Export only the hooks.
7. **Public barrel** `<feature>.ts`. Export the types, enums and schemas pages need. Do not export what nobody
   outside uses, since `knip` reports unused exports.

## 3. Route

Add the path to `src/libs/enums/app-route.enum.ts`:

```ts
PROJECTS: "/projects",
```

Register it in `src/index.tsx` under the `AppShell` children:

```tsx
{
	lazy: async () => {
		const pageModule = await import("~/pages/projects/projects.js");

		return { Component: pageModule.Projects };
	},
	path: AppRoute.PROJECTS,
},
```

Add it to the navigation in `src/libs/components/app-shell/libs/constants/navigation-items.constant.ts` if it
belongs in the header.

## 4. Page (`src/pages/<feature>/`)

The `assets` page looks like this:

```
pages/assets/
├── assets.tsx (+ assets.test.tsx)
├── styles.module.css
├── libs/enums/
└── components/
    └── import-panel/
        ├── import-panel.tsx (+ test)
        ├── styles.module.css
        ├── components/{import-form, import-result, outcome-banner, rejected-rows-table, warnings-table}/
        └── libs/{constants, enums, helpers, types}/
```

1. Put all page copy in a `<Page>Message` enum (`ImportPanelMessage`, `ImportOutcomeMessage`).
2. The page component composes. It holds no request logic itself.
3. The component that owns a query or mutation renders its loading, error, empty and success states. The import
   panel reads results from the mutation state, never from a copy in `useState`.
4. Forms follow `import-form.tsx`: `useAppForm` with the feature schema, controlled field components, `noValidate`,
   and a submit button that shows loading.
5. UI state that should survive a reload lives in the URL (`use-panel.hook.ts` keeps `?panel=import`).

## 5. Verify

```
pnpm lint
pnpm build
pnpm test
```

Write tests next to the code (Vitest + React Testing Library, `fetch` mocked). Then open the page and check the
loading, empty, error (stop the backend) and success paths.
