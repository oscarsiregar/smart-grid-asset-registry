# Backend Service Template (Go)

Template standar backend service: **Go 1.26 · Fiber v2 · GORM/PostgreSQL · go-common**,
Clean Architecture, dengan contoh modul CRUD lengkap (`product`) beserta test di setiap lapis.

> **Aturan proyek: [CLAUDE.md](CLAUDE.md)** (ringkas, mengikat) · detail di **[docs/rules/](docs/rules/)** ·
> menambah fitur: **[docs/menambah-fitur.md](docs/menambah-fitur.md)**

## Memakai template ini untuk service baru

1. Salin folder ini, lalu ganti nama modul Go:
   ```sh
   # ganti "backend-template" dengan nama service, mis. "loan-service"
   grep -rl "backend-template" --include=*.go --include=go.mod --include=*.yaml --include=*.yml --include=Makefile . \
     | xargs sed -i 's#backend-template#loan-service#g'
   ```
2. Sesuaikan `APP_NAME` di `.env.example`, `application.name` di `values.yml`.
3. Pertahankan modul `product` sebagai referensi sampai modul pertama Anda selesai, lalu hapus
   (`domain/*/product*`, `internal/*/product/`, mock-nya, route & wiring-nya, migrasi contoh).
4. `go mod tidy && make mocks && make check`.

## Menjalankan lokal

Prasyarat: Go 1.26+, PostgreSQL, akses ke `repo.smf-indonesia.co.id` (modul privat go-common):

```sh
go env -w GOPRIVATE=repo.smf-indonesia.co.id
cp .env.example .env            # isi DB_*, INTERNAL_SERVICE_SECRET, JWT_SECRET
psql "$DATABASE_URL" -f migrations/000001_create_product.up.sql   # atau golang-migrate
make run                        # http://localhost:3030/v1/health
```

Request bisnis harus membawa `X-Internal-Secret` dan `Authorization: Bearer <JWT>` —
lihat [docs/api/product.md](docs/api/product.md).

## Perintah

| Perintah | Fungsi |
|---|---|
| `make run` | jalankan API |
| `make check` | gofmt + vet + test — **wajib sebelum push** |
| `make test` | unit test (paket yang sama dengan CI) |
| `make cover` | coverage `internal/usecase` (gate CI ≥ 80%) |
| `make lint` | golangci-lint v2 |
| `make mocks` | regenerate mock dari interface `domain/` |
| `make build` | binary ke `bin/api` |

## Struktur

```
cmd/api/                 entry point (config, DB, middleware, graceful shutdown)
domain/                  kontrak tanpa framework: entity, valueobject, errorcode, dto,
                         interface repository/service/usecase
internal/
  apperr/                katalog pesan error (ID + EN)
  bootstrap/             dependency injection + konfigurasi Fiber
  config/                load & validasi env
  handler/<modul>/       HTTP handler — 1 file = 1 endpoint
  middleware/            recover, rate limiter
  mocks/                 mock hasil mockery
  repository/psql/       gormtx, model, mapper, querysafety, <modul>/ (1 file = 1 method)
  routes/                registrasi route + rantai auth
  usecase/<modul>/       logika bisnis — 1 file = 1 use case
migrations/              SQL migrasi berurutan
docs/rules/              aturan pengembangan
docs/api/                kontrak API per modul
```

Alur request:

```
BFF ─► [Recover → CommonMiddleware → Limiter → Compress]
     ─► /v1 group ─► [InternalSecret → Identity → JWTVerifyStrict]
     ─► handler (parse) ─► usecase (permission, validasi, bisnis, tx) ─► repository (GORM) ─► PostgreSQL
     ◄─ response.Success / ErrorHandler global (AppError → status & pesan)
```

## Dokumen aturan

| # | Topik |
|---|---|
| [01](docs/rules/01-arsitektur.md) | Arsitektur & arah dependensi |
| [02](docs/rules/02-penamaan.md) | Penamaan & struktur file |
| [03](docs/rules/03-handler.md) | Handler HTTP |
| [04](docs/rules/04-usecase.md) | Usecase, otorisasi, validasi, transaksi |
| [05](docs/rules/05-repository-database.md) | Repository, GORM, migrasi, integrasi eksternal |
| [06](docs/rules/06-error-handling.md) | Error handling |
| [07](docs/rules/07-keamanan.md) | Keamanan |
| [08](docs/rules/08-testing.md) | Testing |
| [09](docs/rules/09-logging-observability.md) | Logging, health check, shutdown |
| [10](docs/rules/10-git-ci-deploy.md) | Git, CI, deploy |
