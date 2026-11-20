# 10 — Git, CI & Deploy

## Branches

| Prefix | For |
|---|---|
| `feat/<description>` | new features |
| `fix/<description>` | bug fixes |
| `refactor/<description>` | structural changes without behaviour changes |
| `chore/<description>` | dependencies, tooling, CI |
| `docs/<description>` | documentation |

- Created from `main`, merged through a Merge Request. No direct pushes to `main` / `staging`.
- Environment flow: feature branch → `staging` (testing) → `main` (production).

## Commits — Conventional Commits

```
<type>(<scope>): <imperative summary, lower case, no full stop>

<optional body: WHY this change, not what>
```

Examples:
```
feat(product): add bulk status update endpoint
fix(period): reject approve when reviewer is not assigned
docs: document maker-checker-approval status flow
```

- One commit = one logical change that passes the tests.
- Don't commit: `.env`, binaries (`*.exe`, `bin/`), `coverage.out`, IDE files.

## Merge Requests

Checklist before asking for a review:
- [ ] `make check` passes (gofmt, vet, test)
- [ ] `make lint` is clean
- [ ] New tests for every new logic branch; usecase coverage does not drop
- [ ] Mocks regenerated when a `domain/` interface changes
- [ ] New migrations (if any) + `.down.sql`
- [ ] `docs/api/<module>.md` updated when the API contract changes
- [ ] `.env.example` + `config.Validate()` updated for new env variables
- [ ] The security checklist in [07-security](07-security.md)

Reviewers also check: the dependency direction between layers, naming as in [02-naming](02-naming.md), no business logic in handlers/repositories.

## CI (`.gitlab-ci.yml`)

| Job | Contents | Blocking? |
|---|---|---|
| `lint` | `golangci-lint run` (configured in `.golangci.yml`) | yes |
| `go_test` | gofmt check, `go vet`, unit tests (without the wiring/mock/declaration packages) | yes |
| `coverage_check` | `internal/usecase` coverage ≥ 80% | only on `main`/`staging` |
| include `Infra/supply-chain` | build the image, push to Harbor, deploy | as in the Infra template |

The private go-common module is fetched with `CI_JOB_TOKEN` (`git config url.insteadOf`) and `GOPRIVATE`.

> **This project:** CI and deployment run on GitHub Actions to one Azure VM (constitution → *Template
> Exceptions* no. 15, see `deploy/README.md`); go-common is re-implemented locally (no. 4).

## Build & image

- A multi-stage `Dockerfile`: build on `golang:1.26.x-alpine` (`CGO_ENABLED=0`, `-trimpath -ldflags="-s -w"`) → a minimal `alpine` runtime, a non-root user, `tzdata`.
- Configuration from env variables (viper `AutomaticEnv`); **no** `.env` in the image.

## Deploy (Kubernetes / Helm `values.yml`)

- Secrets (`DB_PASSWORD`, `JWT_SECRET`, `INTERNAL_SERVICE_SECRET`) from a Kubernetes Secret; `values.yml` holds only non-secret env.
- `FIBER_PREFORK=false` and `GOMAXPROCS` set to the CPU limit.
- Probes `httpGet /v1/health` with `tcpSocket: null` (the shared chart defaults to tcpSocket and Helm merges maps → the Deployment is rejected when both are present).
- Before deploying to an environment: rebuild from the latest source and make sure the env variables point at that environment (not the `localhost` of a local `.env`).
- DB migrations run **before** the code version that needs them (a migration job / pipeline), and must stay backwards compatible with the code version that is running.

## Go & tooling versions

| Tool | Version |
|---|---|
| Go | 1.26.x (see `go.mod`) |
| golangci-lint | v2.13.x |
| mockery | v2.53.x |
| go-common | pinned in `go.mod` (`v1.0.0`) |

Upgrade Go/tools in a separate `chore/` MR and update this table.
