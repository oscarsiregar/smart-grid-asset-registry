# 09 — Logging & Observability

## Logger

- Created once in `main.go`: `log := logger.New(cfg.IsProd)` (go-common; JSON in prod, text in dev).
- Injected into every usecase (`log *logrus.Logger`). There is no global logger.
- `commonMiddleware.InstallCommonMiddleware` installs the request id / trace id, the access log, the security headers and telemetry. The trace id flows through `ctx.UserContext()`.

## Pattern in a usecase

```go
log := logger.WithCtx(ctx, pu.log).WithField("usecase", "CreateProduct") // trace id included automatically

log.Warn("permission denied")                                   // access refused
log.WithError(err).Error("failed to find product by code")      // technical failure
log.WithField("productID", product.ID).Info("success")          // successful end
```

| Level | When | Example |
|---|---|---|
| `Error` | a technical failure returned as a 500 (DB, external API) — **where it happens**, once | `failed to update product` |
| `Warn` | access refused, an odd but handled condition | `permission denied` |
| `Info` | one success line per usecase, with the relevant ids & counts (`count`) | `success` |
| `Debug` | extra detail while investigating (off in prod) | — |

Rules:
1. Log messages: lower case, English, in the form `failed to <action> <object>`.
2. Contextual data goes through `WithField` / `WithFields`, not into the message string.
3. 4xx business errors (`NotFound`, `Conflict`, validation) are **not** logged as `Error` — they are normal behaviour.
4. Don't log in repositories and handlers (the usecase logs; the `ErrorHandler` logs 5xx errors that get through).
5. Don't log sensitive data (see [07-security](07-security.md)).
6. Don't log twice: when an error was logged where it happened, the upper layer just returns it.

## Health check

- `GET /v1/health` (go-common `RegisterHealthCheck`) really pings the DB (and Redis when given).
- Used by the liveness & readiness probes in `values.yml` (httpGet, not tcpSocket).
- A new critical dependency (Redis, a broker) → pass it to `RegisterHealthCheck` so the probe reflects real health.

> **This project:** the health check is `GET /api/health` (no version prefix, constitution →
> *Template Exceptions* no. 5).

## GORM log

- Non-prod: `gormlogger.Info` (every query shown) — useful during development.
- Prod: the go-common logger (slow queries & errors only). Don't enable Info in prod (data leaks & volume).

## Graceful shutdown

`main.go` waits for SIGINT/SIGTERM (`utils.NewShutdownHelper`), stops Fiber with a
10-second timeout (running requests are finished), then closes the DB connection. New resources
(Redis, brokers, workers) **must** be registered in `sh.Graceful(map[...])`.
