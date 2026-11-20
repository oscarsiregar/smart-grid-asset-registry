# 07 — Security

> **This project:** there is no authentication or authorisation (constitution → *Template Exceptions* no. 1).
> Business routes are public under `/api`; the authentication chain and permission checks below are the
> template's rules, to be restored if users are ever introduced. Every other section applies as written.

## Authentication chain (MUST, the order must not change)

This service sits **behind a BFF** (one-bff-go). Clients never call it directly.

```go
authenticated := v1.Group("",
	pkg_middleware_fiber.InternalSecretFromEnv(cfg.InternalServiceSecret), // 1
	pkg_middleware_fiber.Identity(),                                       // 2
	pkg_middleware_fiber.JWTVerifyStrict(cfg.JWTSecret),                   // 3
)
```

| # | Middleware | Rejects with | What it does |
|---|---|---|---|
| 1 | `InternalSecretFromEnv` | 403 | the request must carry the BFF's `X-Internal-Secret` |
| 2 | `Identity` | — | reads `X-User-*` into `ctx.Locals` |
| 3 | `JWTVerifyStrict` | 401 | verifies the JWT again (HMAC) and **overwrites** the identity from the headers with the JWT claims |

Background (incident of 2026-09-29/30 in another service): one route group forgot to install
`InternalSecret`, so the `X-User-Permissions` header could be forged for full
impersonation. `InternalSecret` alone is not enough either, because one secret is shared by
every service. That is why **`JWTVerifyStrict` is required** — `JWTVerify` (non-strict)
is only allowed with a reason written in the code.

Rules:
- Every business route **under `authenticated`**. Only `/v1/health` is public.
- Don't create a new group that bypasses this chain, and don't add routes to `app` or `v1` directly.
- Don't read the `X-User-*` headers yourself — use `GetUserIdentity(ctx)`.

## Authorisation

- Every endpoint has **its own permission** (`<module>.<action>`), checked in the **usecase** as the first step. There is no "logged in is enough" endpoint.
- Default deny. The permissions are documented in `docs/api/<module>.md` and registered with the identity service.
- Data-level authorisation (ownership, category, assignment) is checked after the data is read, before it is returned/changed.
- List responses only contain data the caller may see (filtered in the query, not in the frontend).

## Input & database

- Every user value goes through GORM placeholders (`Where("code = ?", code)`).
- ORDER BY: `querysafety.SafeOrderClause` + an allow-list. LIKE: `querysafety.ContainsPattern`.
- Limit sizes: Fiber's `BodyLimit` (4 MB template default), paging `MaxLimit` 100, string lengths validated in `Validate()`.
- File uploads: validate the MIME type from the content (not the extension), a size limit, a server-generated object name (UUID), stored in object storage, accessed through short-lived presigned URLs.

> **This project:** uploaded CSV files are parsed in memory and never stored (constitution → *Template
> Exceptions* no. 12).

## Secrets & configuration

- Secrets only come from the environment (Kubernetes Secret). **Not** in code, `values.yml`, the Dockerfile or the image.
- `.env` is for local use only and is in `.gitignore`; `.env.example` holds the variable names without secret values.
- Required variables are validated in `config.Validate()` — the service fails to start with the list of missing variables, instead of failing silently at runtime.
- Rotating a secret must not need a code change.

## Logging & sensitive data

- Never log: passwords, tokens/JWTs, secrets, the `Authorization`/`X-Internal-Secret` headers, raw request bodies, personal data (national ID numbers, account numbers, full email/phone).
- Log ids (productID, periodID, userID), not data contents.
- A 500 error sends no details to the client (handled by the `ErrorHandler`).

## Transport & runtime

- Containers run as **non-root** (`USER appuser`), on a minimal runtime image (alpine), with a static binary (`CGO_ENABLED=0`).
- Outbound HTTP: TLS is verified. `InsecureSkipVerify` is forbidden in production.
- The rate limiter (`middleware.LimiterConfig`) slows down clients that keep failing (20 non-2xx responses / 15 minutes).
- `RecoverConfig` makes sure one panic does not kill the process.

## Dependencies

- `go.mod` pins versions; private modules (`repo.smf-indonesia.co.id`) go through `GOPRIVATE`, not through the public proxy.
- Dependency updates go in their own MR; run `go mod tidy` + every test.
- `golangci-lint` with `gosec` enabled in CI.

## Security review checklist (every MR)

- [ ] New routes are under the `authenticated` group
- [ ] New permissions are defined & checked in the usecase (with a `permission denied` test case)
- [ ] No SQL built by string concatenation; sort/LIKE use `querysafety`
- [ ] No secrets/PII in code, logs or test fixtures
- [ ] Technical errors are not wrapped into messages that leak details
- [ ] New env variables are added to `config.Validate()` (when required) and `.env.example`
