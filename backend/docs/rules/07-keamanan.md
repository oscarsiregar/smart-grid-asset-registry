# 07 — Keamanan

## Rantai autentikasi (MUST, urutan tidak boleh diubah)

Service ini berada **di belakang BFF** (one-bff-go). Client tidak pernah memanggilnya langsung.

```go
authenticated := v1.Group("",
	pkg_middleware_fiber.InternalSecretFromEnv(cfg.InternalServiceSecret), // 1
	pkg_middleware_fiber.Identity(),                                       // 2
	pkg_middleware_fiber.JWTVerifyStrict(cfg.JWTSecret),                   // 3
)
```

| # | Middleware | Menolak dengan | Fungsi |
|---|---|---|---|
| 1 | `InternalSecretFromEnv` | 403 | request harus membawa `X-Internal-Secret` milik BFF |
| 2 | `Identity` | — | membaca `X-User-*` ke `ctx.Locals` |
| 3 | `JWTVerifyStrict` | 401 | memverifikasi ulang JWT (HMAC) dan **menimpa** identitas dari header dengan klaim JWT |

Latar belakang (insiden 2026-09-29/30 di service lain): sebuah grup route lupa memasang
`InternalSecret`, sehingga header `X-User-Permissions` bisa dipalsukan untuk
impersonasi penuh. `InternalSecret` saja juga tidak cukup karena satu secret dipakai
semua service. Karena itu **`JWTVerifyStrict` wajib** — `JWTVerify` (non-strict)
hanya boleh dengan alasan tertulis di kode.

Aturan:
- Semua route bisnis **di bawah `authenticated`**. Hanya `/v1/health` yang publik.
- Jangan membuat grup baru yang melewati rantai ini, jangan menambah route di `app` atau `v1` langsung.
- Jangan membaca header `X-User-*` sendiri — gunakan `GetUserIdentity(ctx)`.

## Otorisasi

- Setiap endpoint punya **permission sendiri** (`<modul>.<aksi>`), dicek di **usecase** sebagai langkah pertama. Tidak ada endpoint "cukup login".
- Default deny. Daftar permission didokumentasikan di `docs/api/<modul>.md` dan didaftarkan ke identity-service.
- Otorisasi tingkat data (kepemilikan, kategori, penugasan) dicek setelah data dibaca, sebelum data dikembalikan/diubah.
- Respons list hanya memuat data yang boleh dilihat caller (filter di query, bukan di frontend).

## Input & database

- Semua nilai user lewat placeholder GORM (`Where("code = ?", code)`).
- ORDER BY: `querysafety.SafeOrderClause` + allow-list. LIKE: `querysafety.ContainsPattern`.
- Batasi ukuran: `BodyLimit` Fiber (4 MB default template), `MaxLimit` paging 100, panjang string divalidasi di `Validate()`.
- Upload file: validasi MIME dari isi (bukan ekstensi), batas ukuran, nama objek di-generate server (UUID), simpan di object storage, akses via presigned URL berumur pendek.

## Secret & konfigurasi

- Secret hanya dari environment (Kubernetes Secret). **Tidak** di kode, `values.yml`, Dockerfile, atau image.
- `.env` hanya untuk lokal dan ada di `.gitignore`; `.env.example` berisi nama variabel tanpa nilai rahasia.
- Variabel wajib divalidasi di `config.Validate()` — service gagal start dengan daftar variabel yang hilang, bukan gagal diam-diam saat runtime.
- Rotasi secret tidak boleh butuh perubahan kode.

## Logging & data sensitif

- Jangan log: password, token/JWT, secret, header `Authorization`/`X-Internal-Secret`, body request mentah, data pribadi (NIK, nomor rekening, email/telepon lengkap).
- Log id (productID, periodID, userID), bukan isi data.
- Error 500 tidak mengirim detail ke client (ditangani `ErrorHandler`).

## Transport & runtime

- Container berjalan sebagai **non-root** (`USER appuser`), image runtime minimal (alpine), binary statis (`CGO_ENABLED=0`).
- Outbound HTTP: TLS diverifikasi. `InsecureSkipVerify` dilarang di production.
- Rate limiter (`middleware.LimiterConfig`) membatasi klien yang terus gagal (20 respons non-2xx / 15 menit).
- `RecoverConfig` memastikan satu panic tidak mematikan proses.

## Dependensi

- `go.mod` mem-pin versi; modul privat (`repo.smf-indonesia.co.id`) lewat `GOPRIVATE`, tidak lewat proxy publik.
- Update dependensi lewat MR tersendiri, jalankan `go mod tidy` + seluruh test.
- `golangci-lint` dengan `gosec` aktif di CI.

## Checklist review keamanan (setiap MR)

- [ ] Route baru di bawah grup `authenticated`
- [ ] Permission baru didefinisikan & dicek di usecase (dengan test kasus `permission denied`)
- [ ] Tidak ada SQL dari string concat; sort/LIKE memakai `querysafety`
- [ ] Tidak ada secret/PII di kode, log, atau test fixture
- [ ] Error teknis tidak dibungkus menjadi pesan yang membocorkan detail
- [ ] Env baru ditambahkan ke `config.Validate()` (bila wajib) dan `.env.example`
