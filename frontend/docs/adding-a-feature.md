# Adding a feature

A step-by-step recipe for a new backend resource with its own page, using the `notes` example as the reference.
Replace `<feature>` with a kebab-case plural (`projects`) and `<Feature>` with its PascalCase form (`Projects`).

## 1. API path

Add the resource root to `src/libs/enums/api-path.enum.ts`:

```ts
const APIPath = {
	AUTH: "/auth",
	NOTES: "/notes",
	PROJECTS: "/projects",
} as const;
```

## 2. Feature module (`src/modules/<feature>/`)

```
modules/projects/
├── projects-api.ts
├── projects.ts
└── libs/
    ├── enums/
    │   ├── enums.ts
    │   ├── project-validation-message.enum.ts
    │   ├── project-validation-rule.enum.ts
    │   ├── projects-api-path.enum.ts
    │   └── projects-api-tag.enum.ts
    ├── types/
    │   ├── types.ts
    │   ├── project-create-request-dto.type.ts
    │   └── project-dto.type.ts
    └── validation-schemas/
        ├── validation-schemas.ts
        └── project-create.validation-schema.ts
```

1. **Tag enum.** `ProjectsApiTag = { PROJECT: "project" } as const`.
2. **Path enum.** Sub-paths relative to the root: `ProjectsApiPath = { $ID: "/:id", ROOT: "/" } as const`.
3. **Validation rule and message enums.** Limits in the rule enum, messages built from the limits.
4. **Schema.** `const projectCreate = z.object({ … }).required();` exported from the barrel as
   `projectCreateValidationSchema`.
5. **Types.** The response DTO is written by hand to match the API. The request DTO is
   `z.infer<typeof projectCreate>`.
6. **Endpoints** in `projects-api.ts`. Copy the shape of `notes-api.ts`: `enhanceEndpoints` with the tag, then
   `injectEndpoints`. Export only the hooks.
7. **Public barrel** `projects.ts`. Export the types and schemas pages need. Do not export enums nobody outside uses,
   since `knip` reports unused exports.

## 3. Route

Add the path to `src/libs/enums/app-route.enum.ts`:

```ts
PROJECTS: "/projects",
PROJECTS_$PROJECT_ID: "/projects/:projectId",
```

Register it in `src/index.tsx` under the `AuthenticatedShell` children:

```tsx
{
	lazy: async () => {
		const pageModule = await import("~/pages/projects/projects.js");

		return { Component: pageModule.Projects };
	},
	path: AppRoute.PROJECTS,
},
```

Add it to the navigation in
`src/libs/components/authenticated-shell/libs/constants/navigation-items.constant.ts` if it belongs in the header.

## 4. Page (`src/pages/<feature>/`)

```
pages/projects/
├── projects.tsx
├── styles.module.css
├── components/
│   ├── project-create-form/
│   │   ├── project-create-form.tsx
│   │   ├── styles.module.css
│   │   └── libs/constants/{constants.ts, default-project-create-payload.constant.ts, project-create-fields.constant.ts}
│   └── projects-list/
│       ├── projects-list.tsx
│       ├── styles.module.css
│       └── components/project-item/{project-item.tsx, styles.module.css}
└── libs/
    └── enums/{enums.ts, projects-page-message.enum.ts}
```

1. Put all page copy in `ProjectsPageMessage`.
2. The page component composes. It holds no request logic itself.
3. The list component owns the query and renders loading (`Loader`), error (message + retry), empty, and data states.
4. The form component owns the mutation and follows `note-create-form.tsx`: `useAppForm`, `useServerFormErrors`,
   `FormAlert`, `try { await mutation(payload).unwrap(); reset(); showNotification(); } catch {}`.
5. Each item component owns its own row-level mutation (delete, toggle) and memoizes its handlers.

## 5. Verify

```
pnpm lint
pnpm build
```

Then open the page and check the loading, empty, error (stop the backend), validation error, and success paths.
