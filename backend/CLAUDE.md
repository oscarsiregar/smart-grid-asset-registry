# Aturan Proyek (WAJIB dibaca sebelum menulis kode)

File ini berlaku untuk manusia **dan** AI agent (Claude Code, Codex, Cursor, dll).
Ringkasan di sini bersifat **mengikat**; detail & alasan ada di [docs/rules/](docs/rules/).
Contoh implementasi referensi: modul **`product`** — tiru polanya, jangan menciptakan pola baru.

## Stack
Go 1.26 · Fiber v2 · GORM + PostgreSQL · `repo.smf-indonesia.co.id/package-helper/go-common` (response, apperror, logger, middleware, database) · testify + mockery v2 · sqlmock.

## Arsitektur (Clean Architecture) — [01-arsitektur](docs/rules/01-arsitektur.md)
```
handler  ──►  usecase  ──►  repository / service
(HTTP)        (bisnis)       (DB / API eksternal)
   │             │                 │
   └──── semua hanya bergantung pada interface di domain/ ────┘
```
1. `domain/` **tidak boleh** import `internal/`, gorm, fiber, atau go-common. Isinya: entity, valueobject, errorcode, dto, interface (repository, service, usecase).
2. Handler hanya: parse request → konversi ke `usecase.XxxInput` → panggil usecase → map ke DTO response. **Tanpa logika bisnis, tanpa akses DB.**
3. Usecase memegang **semua** logika bisnis: otorisasi (permission), validasi, aturan, transaksi. Tidak tahu HTTP/Fiber.
4. Repository hanya akses data. Tidak ada aturan bisnis, tidak mengembalikan `apperr`.
5. Dependency injection **hanya** di `internal/bootstrap/setup.go`. Route **hanya** di `internal/routes/router.go`.

## Struktur file — [02-penamaan](docs/rules/02-penamaan.md)
- **1 file = 1 endpoint/use case/method repository.** `create_product.go`, `find_by_id.go`, dst. Setiap file punya `_test.go` pasangannya.
- Tiap package punya `init.go` (struct + constructor `NewXxx`) dan opsional `helpers.go`.
- Nama package implementasi: `<modul>_usecase`, `<modul>_repository`; handler: `<modul>`. Import alias = nama package.
- Nama method seragam di semua layer: `GetListXxx`, `GetXxxDetail`, `CreateXxx`, `UpdateXxx`, `DeleteXxx`.
- JSON: **camelCase**. Tabel/kolom DB: **snake_case**, tabel singular. Error code: `SCREAMING_SNAKE_CASE`. Permission: `"<modul>.<aksi>"`.

## Handler — [03-handler](docs/rules/03-handler.md)
- Parse dengan `ctx.QueryParser` / `ctx.ParamsParser` / `ctx.BodyParser` ke struct di `domain/dto`. Gagal parse → `apperr.BadRequest(errorcode.InvalidRequest, ...)`.
- Konversi tipe di handler (UUID, enum via `valueobject.ParseXxx`) sebelum masuk usecase.
- Identitas & permission dari `pkg_middleware_fiber.GetUserIdentity(ctx)` → diteruskan ke Input (`Permissions`, `ActorName`).
- Pakai `ctx.UserContext()` (bukan `ctx` Fiber) saat memanggil usecase.
- Sukses: `response.Success(ctx, status, "Successfully <aksi> <modul>", dto)`. Delete → `204`. Create → `201`.
- Error: cukup `return err`. **Jangan** menulis response error manual.

## Usecase — [04-usecase](docs/rules/04-usecase.md)
- Signature: `func (u xxxUsecase) Method(ctx, input usecase.MethodInput) (usecase.MethodOutput, error)`.
- Urutan wajib: **log setup → cek permission → validasi input → aturan bisnis → tulis (dalam transaksi bila >1 query tulis / read-then-write) → log sukses**.
- Permission dicek **di usecase**, pertama kali, dengan konstanta dari `domain/valueobject/permission.go`.
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
- Kode baru → tambahkan di `domain/errorcode` (+ `All`) **dan** pesan ID+EN di katalog `apperr`. Test akan gagal bila lupa.
- Error teknis (DB, jaringan) dikembalikan apa adanya → otomatis jadi 500 "Internal Server Error" (detail tidak bocor ke client, tetap ter-log).

## Keamanan — [07-keamanan](docs/rules/07-keamanan.md)
- Semua route bisnis di grup `authenticated` (InternalSecret → Identity → **JWTVerifyStrict**). Tidak ada route bisnis di luar grup itu.
- Setiap endpoint punya permission sendiri & dicek di usecase. Default: **deny**.
- Tidak ada secret di kode, `.env`, `values.yml`, log, atau commit. Config wajib divalidasi di `config.Validate()`.
- Jangan log password/token/PII/body request mentah.

## Testing — [08-testing](docs/rules/08-testing.md)
- Setiap file logika punya test. Table-driven, mock **per kasus** (mockery `EXPECT()`), `StubWithinTx` untuk transaksi.
- Usecase: tiap cabang (permission, validasi, not found, conflict, error repo, rollback, sukses). Coverage `internal/usecase` ≥ **80%** (gate di main/staging).
- Repository: sqlmock (regexp) — sukses, error DB, trx diberikan, Tx tipe salah.
- Handler: `app.Test` + `utils.NewErrorHandler` — status code, mapping body, parse error.
- Setelah ubah interface di `domain/`: `make mocks`. Jangan edit file mock hasil generate.

## Logging — [09-logging-observability](docs/rules/09-logging-observability.md)
- `log := logger.WithCtx(ctx, u.log).WithField("usecase", "NamaMethod")` di awal setiap usecase.
- `Warn` untuk permission denied, `Error` + `WithError(err)` untuk kegagalan teknis, `Info("success")` di akhir. Jangan log error yang sudah jadi `apperr` 4xx sebagai Error.

## Menambah fitur baru
Ikuti checklist langkah demi langkah di **[docs/menambah-fitur.md](docs/menambah-fitur.md)**.
