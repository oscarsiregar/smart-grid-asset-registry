# 09 — Logging & Observability

## Logger

- Dibuat sekali di `main.go`: `log := logger.New(cfg.IsProd)` (go-common; JSON di prod, teks di dev).
- Diinjeksi ke setiap usecase (`log *logrus.Logger`). Tidak ada logger global.
- `commonMiddleware.InstallCommonMiddleware` memasang request id / trace id, access log, header keamanan, dan telemetry. Trace id mengalir lewat `ctx.UserContext()`.

## Pola di usecase

```go
log := logger.WithCtx(ctx, pu.log).WithField("usecase", "CreateProduct") // trace id otomatis

log.Warn("permission denied")                                   // akses ditolak
log.WithError(err).Error("failed to find product by code")      // kegagalan teknis
log.WithField("productID", product.ID).Info("success")          // akhir sukses
```

| Level | Kapan | Contoh |
|---|---|---|
| `Error` | kegagalan teknis yang dikembalikan sebagai 500 (DB, API eksternal) — **di titik kejadian**, sekali | `failed to update product` |
| `Warn` | penolakan akses, kondisi aneh tapi tertangani | `permission denied` |
| `Info` | satu baris sukses per usecase, dengan id relevan & jumlah (`count`) | `success` |
| `Debug` | detail tambahan saat investigasi (mati di prod) | — |

Aturan:
1. Pesan log: huruf kecil, Bahasa Inggris, format `failed to <aksi> <objek>`.
2. Data kontekstual lewat `WithField` / `WithFields`, bukan disisipkan di string pesan.
3. Error bisnis 4xx (`NotFound`, `Conflict`, validasi) **tidak** di-log sebagai `Error` — itu perilaku normal.
4. Jangan log di repository dan handler (usecase yang log; `ErrorHandler` me-log 5xx yang lolos).
5. Jangan log data sensitif (lihat [07-keamanan](07-keamanan.md)).
6. Jangan log ganda: bila error sudah di-log saat terjadi, lapis atas cukup mengembalikannya.

## Health check

- `GET /v1/health` (go-common `RegisterHealthCheck`) benar-benar mem-ping DB (dan Redis bila diberikan).
- Dipakai liveness & readiness probe di `values.yml` (httpGet, bukan tcpSocket).
- Menambah dependensi kritis baru (Redis, broker) → teruskan ke `RegisterHealthCheck` agar probe mencerminkan kesehatan nyata.

## GORM log

- Non-prod: `gormlogger.Info` (semua query tampil) — berguna saat development.
- Prod: logger go-common (query lambat & error saja). Jangan aktifkan Info di prod (bocor data & volume).

## Graceful shutdown

`main.go` menunggu SIGINT/SIGTERM (`utils.NewShutdownHelper`), menghentikan Fiber dengan
timeout 10 detik (request berjalan diselesaikan), lalu menutup koneksi DB. Resource baru
(Redis, broker, worker) **wajib** didaftarkan di `sh.Graceful(map[...])`.
