# Frontend template

A React + TypeScript frontend skeleton with strict conventions, extracted from the Promptomat project. It ships with
the full infrastructure (store, API layer, error handling, notifications, forms, routing), a sign-in / sign-up flow,
and one example CRUD feature (`notes`) to copy from.

- Rules: [docs/code-conventions.md](docs/code-conventions.md)
- Recipe for a new feature: [docs/adding-a-feature.md](docs/adding-a-feature.md)
- Review rubric: [docs/quality-criteria.md](docs/quality-criteria.md)

## Requirements

- Node.js 24 (`.nvmrc`)
- pnpm 11 (`corepack enable`)

## Getting started

```
cp .env.example .env
pnpm install
pnpm start:dev
```

The dev server runs on `VITE_APP_DEVELOPMENT_PORT` and proxies `VITE_APP_API_ORIGIN_URL` to
`VITE_APP_PROXY_SERVER_URL`, so the browser talks to one origin and no CORS setup is needed.

To install the git hooks (pre-commit lint, commit message check), initialise a git repository, then run:

```
pnpm ci:prepare
```

## Scripts

| Script           | Purpose                                                                       |
| ---------------- | ----------------------------------------------------------------------------- |
| `pnpm start:dev` | Vite dev server                                                               |
| `pnpm build`     | Type check + production build into `build/`                                   |
| `pnpm preview`   | Serve the production build                                                    |
| `pnpm lint`      | Every check (types, ESLint, Stylelint, Prettier, ls-lint, editorconfig, knip) |
| `pnpm format`    | Apply Prettier                                                                |

## Environment

| Variable                    | Used by        | Meaning                                     |
| --------------------------- | -------------- | ------------------------------------------- |
| `VITE_APP_NODE_ENV`         | app            | `local`, `development`, or `production`     |
| `VITE_APP_API_ORIGIN_URL`   | app, dev proxy | Base path of the API, for example `/api/v1` |
| `VITE_APP_DEVELOPMENT_PORT` | dev server     | Port of the Vite dev server                 |
| `VITE_APP_PROXY_SERVER_URL` | dev server     | Backend the dev proxy forwards to           |

## Backend contract the template expects

| Method | Path                       | Body                            | Response          |
| ------ | -------------------------- | ------------------------------- | ----------------- |
| POST   | `/auth/sign-in`            | `{ email, password }`           | `{ token, user }` |
| POST   | `/auth/sign-up`            | `{ email, nickname, password }` | `{ token, user }` |
| GET    | `/auth/authenticated-user` |                                 | `user`            |
| GET    | `/notes`                   |                                 | `NoteDto[]`       |
| POST   | `/notes`                   | `{ title, content }`            | `NoteDto`         |
| DELETE | `/notes/:id`               |                                 | empty             |

Authenticated requests carry `Authorization: Bearer <token>`. Errors are `{ code, message }` with a code from
`ErrorCode`; validation errors add `details: [{ path, message }]`. A missing or expired token must answer
`UNAUTHENTICATED`.

## Starting a new project from the template

1. Copy the folder and rename `name` in `package.json`.
2. Set `APP_NAME` in `src/libs/constants/app-name.constant.ts` and `<title>` in `index.html`.
3. Replace `src/assets/img/logo.svg`, `public/favicon.ico`, the fonts, and the colors in `tokens.css` if the brand
   changes.
4. Adapt `ErrorCode`, `AuthApiPath`, and `UserDto` to your backend.
5. Delete `modules/notes` and `pages/notes` once your first real feature exists, together with `APIPath.NOTES`,
   `AppRoute.NOTES`, its route, and its navigation item. `pnpm lint:trash` lists anything left behind.

## Docker

```
docker build --build-arg VITE_APP_NODE_ENV=production --build-arg VITE_APP_API_ORIGIN_URL=/api/v1 -t frontend .
```

The image serves `build/` with nginx and proxies `/api` to the `backend` upstream in `nginx/nginx.conf`
(`localhost:3001` by default). Point it at your backend host before deploying. `pnpm-lock.yaml` must exist, so run
`pnpm install` once before the first build.
