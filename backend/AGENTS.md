# AGENTS.md — aturan backend (WAJIB dibaca sebelum menulis kode)

File ini berlaku untuk manusia **dan** AI agent (Claude Code, Codex, Cursor, dll).
Ringkasan di sini bersifat **mengikat**; detail & alasan ada di [docs/rules/](docs/rules/).
Contoh implementasi referensi: modul **`asset_import`** (`internal/handler/asset_import`, `internal/usecase/asset_import`,
`internal/repository/psql/asset` dan `asset_import`; aturan murni di `domain/assetimport`) — tiru polanya, jangan menciptakan pola baru.
Modul contoh template `product` sudah dihapus (constitution → *Template Exceptions* no. 7); bila [docs/menambah-fitur.md](docs/menambah-fitur.md)
menyebut file `product`, padanannya ada di modul `asset_import` (mis. `create_product.go` → `create_asset_import.go`).

> **Khusus proyek ini:** tidak ada autentikasi/otorisasi (pengecualian resmi terhadap template, lihat
> `.specify/memory/constitution.md` → *Template Exceptions* no. 1). Route bisnis publik di bawah `/api`; tidak ada
> identitas user, permission, maupun middleware JWT. Aturan lain berlaku tanpa perubahan.

## Stack
Go 1.26 · Fiber v2 · GORM + PostgreSQL · `go-common` (response, apperror, logger, middleware, database) · testify + mockery v2 · sqlmock.

> **Khusus proyek ini:** `go-common` adalah library privat kantor dan **diimplementasi ulang secara lokal** di dalam
> `backend/` dengan nama, signature, dan output JSON yang sama (constitution → *Template Exceptions* no. 4). Setiap
> penyebutan go-common di aturan ini berarti paket lokal tersebut. Jangan menambah dependensi ke modul privat.

## Arsitektur (Clean Architecture) — [01-arsitektur](docs/rules/01-arsitektur.md)
```
handler  ──►  usecase  ──►  repository / service
(HTTP)        (bisnis)       (DB / API eksternal)
   │             │                 │
   └──── semua hanya bergantung pada interface di domain/ ────┘
```
1. `domain/` **tidak boleh** import `internal/`, gorm, fiber, atau go-common. Isinya: entity, valueobject, errorcode, dto, interface (repository, service, usecase).
2. Handler hanya: parse request → konversi ke `usecase.XxxInput` → panggil usecase → map ke DTO response. **Tanpa logika bisnis, tanpa akses DB.**
3. Usecase memegang **semua** logika bisnis: validasi, aturan, transaksi. Tidak tahu HTTP/Fiber.
4. Repository hanya akses data. Tidak ada aturan bisnis, tidak mengembalikan `apperr`.
5. Dependency injection **hanya** di `internal/bootstrap/setup.go`. Route **hanya** di `internal/routes/router.go`.

## Struktur file — [02-penamaan](docs/rules/02-penamaan.md)
- **1 file = 1 endpoint/use case/method repository.** `create_asset_import.go`, `find_by_asset_ids.go`, dst. Setiap file punya `_test.go` pasangannya.
- Tiap package punya `init.go` (struct + constructor `NewXxx`) dan opsional `helpers.go`.
- Nama package implementasi: `<modul>_usecase`, `<modul>_repository`; handler: `<modul>`. Import alias = nama package.
- Nama method seragam di semua layer: `GetListXxx`, `GetXxxDetail`, `CreateXxx`, `UpdateXxx`, `DeleteXxx`.
- JSON: **camelCase**. Tabel/kolom DB: **snake_case**, tabel singular. Error code: `SCREAMING_SNAKE_CASE`.

## Handler — [03-handler](docs/rules/03-handler.md)
- Parse dengan `ctx.QueryParser` / `ctx.ParamsParser` / `ctx.BodyParser` ke struct di `domain/dto`. Gagal parse → `apperr.BadRequest(errorcode.InvalidRequest, ...)`.
- Konversi tipe di handler (UUID, enum via `valueobject.ParseXxx`) sebelum masuk usecase.
- Pakai `ctx.UserContext()` (bukan `ctx` Fiber) saat memanggil usecase.
- Sukses: `response.Success(ctx, status, "Successfully <aksi> <modul>", dto)`. Delete → `204`. Create → `201`.
- Error: cukup `return err`. **Jangan** menulis response error manual.

## Usecase — [04-usecase](docs/rules/04-usecase.md)
- Signature: `func (u xxxUsecase) Method(ctx, input usecase.MethodInput) (usecase.MethodOutput, error)`.
- Urutan wajib: **log setup → validasi input → aturan bisnis → tulis (dalam transaksi bila >1 query tulis / read-then-write) → log sukses**.
- `input.Validate()` (di domain) mengembalikan `error` biasa; usecase membungkusnya `apperr.BadRequest(errorcode.InvalidRequest, err.Error())`.
- Repository `nil, nil` = tidak ditemukan → usecase yang memutuskan `apperr.NotFound(...)`.
- Transaksi: `u.txManager.WithinTx(ctx, func(ctx, trx) error {...})`, teruskan `trx` ke **setiap** pemanggilan repository di dalamnya.
- Waktu dari `u.now()` (diinjeksi), ID dari `uuid.New()`.

## Repository & DB — [05-repository-database](docs/rules/05-repository-database.md)
- Signature: `(ctx context.Context, trx repository.Tx, ...)`. Ambil koneksi via `gormtx.Conn(ctx, r.db, trx)`.
- Model GORM (`internal/repository/psql/model`) **tidak pernah** keluar dari layer repository — konversi via `mapper.ToEntityXxx / ToModelXxx`.
- Not found → `return nil, nil` (pakai `errors.Is(err, gorm.ErrRecordNotFound)`).
- **Jangan pakai `db.Save`** untuk update (fallback ke INSERT). Pakai `Model(...).Select(kolom...).Updates(...)`.
- **Jangan** menamai field `CreatedBy`/`UpdatedBy` kecuali berisi UUID — audit plugin go-common akan menimpa/membuangnya. Pakai `CreatedByName`/`UpdatedByName`.
- Semua nilai user → placeholder `?`. ORDER BY → `querysafety.SafeOrderClause` + allow-list. LIKE → `querysafety.ContainsPattern`.
- Perubahan skema → file migrasi baru di `migrations/` (`NNNNNN_deskripsi.up.sql` + `.down.sql`). Jangan ubah migrasi yang sudah dirilis.

## Error — [06-error-handling](docs/rules/06-error-handling.md)
- Error bisnis **hanya** lewat `internal/apperr` (`BadRequest`, `Forbidden`, `NotFound`, `Conflict`, `UnprocessableEntity`).
- Kode baru → tambahkan di `domain/errorcode` (+ `All`) **dan** satu pesan berbahasa Inggris di katalog `apperr` (proyek ini: katalog hanya EN, constitution → *Template Exceptions* no. 6). Test akan gagal bila lupa.
- Error teknis (DB, jaringan) dikembalikan apa adanya → otomatis jadi 500 "Internal Server Error" (detail tidak bocor ke client, tetap ter-log).

## Keamanan — [07-keamanan](docs/rules/07-keamanan.md)
- Proyek ini tanpa auth: route bisnis publik di bawah `/api`, tanpa InternalSecret/Identity/JWT dan tanpa cek permission.
  Bila suatu saat ada konsep user, kembali ke aturan template di [07-keamanan](docs/rules/07-keamanan.md).
- Tidak ada secret di kode, `.env`, `values.yml`, log, atau commit. Config wajib divalidasi di `config.Validate()`.
- Jangan log password/token/PII/body request mentah.

## Testing — [08-testing](docs/rules/08-testing.md)
- Setiap file logika punya test. Table-driven, mock **per kasus** (mockery `EXPECT()`), `StubWithinTx` untuk transaksi.
- Usecase: tiap cabang (validasi, not found, conflict, error repo, rollback, sukses). Coverage `internal/usecase` ≥ **80%** (gate di main/staging).
- Repository: sqlmock (regexp) — sukses, error DB, trx diberikan, Tx tipe salah.
- Handler: `app.Test` + `utils.NewErrorHandler` — status code, mapping body, parse error.
- Setelah ubah interface di `domain/`: `make mocks`. Jangan edit file mock hasil generate.

## Logging — [09-logging-observability](docs/rules/09-logging-observability.md)
- `log := logger.WithCtx(ctx, u.log).WithField("usecase", "NamaMethod")` di awal setiap usecase.
- `Error` + `WithError(err)` untuk kegagalan teknis, `Info("success")` di akhir. Jangan log error yang sudah jadi `apperr` 4xx sebagai Error.

## Perintah penting
| Tujuan | Perintah |
|---|---|
| Jalankan lokal | `make run` |
| Format + vet + test | `make check` |
| Lint | `make lint` |
| Regenerate mock setelah ubah interface `domain/` | `make mocks` |
| Coverage usecase (gate ≥ 80%) | `make cover` |

## Menambah fitur baru
Ikuti checklist langkah demi langkah di **[docs/menambah-fitur.md](docs/menambah-fitur.md)**.
