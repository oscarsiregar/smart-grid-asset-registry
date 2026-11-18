# 01 — Arsitektur

## Prinsip

Service ini memakai **Clean Architecture** sederhana dengan tiga lapis kerja
(handler → usecase → repository/service) dan satu lapis kontrak (`domain/`).
Tujuannya: logika bisnis bisa dites tanpa HTTP dan tanpa database, dan setiap
lapis bisa diganti tanpa menyentuh lapis lain.

```
                 ┌──────────────────────────── domain/ (kontrak, tanpa framework) ───────────────────────────┐
                 │ entity · valueobject · errorcode · dto · repository(interface) · service(interface) · usecase(interface + Input/Output) │
                 └───────────────▲─────────────────────────▲───────────────────────────▲─────────────────────┘
                                 │                         │                           │
HTTP ─► routes ─► internal/handler ─► internal/usecase ─► internal/repository/psql ─► PostgreSQL
                    (parse & map)     (aturan bisnis)  └► internal/service/*       ─► API eksternal
```

## Struktur folder

```
cmd/api/main.go                  entry point: config, DB, Fiber, middleware, graceful shutdown
domain/
  entity/                        struct domain murni (tanpa tag gorm/json)
  valueobject/                   enum (Parse/Scan/Value/String/IsValid) + konstanta permission
  errorcode/                     konstanta kode error (+ daftar All)
  dto/                           kontrak HTTP: Request (query/params/json tag) & Response (json camelCase)
  repository/                    interface repository + Tx/TransactionManager + struct filter
  service/                       interface integrasi keluar (API lain, storage, broker)
  usecase/                       interface usecase + XxxInput/XxxOutput + Validate()/SetDefault()
internal/
  apperr/                        katalog pesan error (ID + EN) → apperror go-common
  bootstrap/                     composition root (DI) + konfigurasi Fiber
  config/                        load & validasi env
  handler/<modul>/               HTTP handler, 1 file = 1 endpoint
  middleware/                    middleware lokal (recover, limiter)
  mocks/                         hasil mockery (+ txhelpers.go manual)
  repository/psql/
    gormtx/                      adapter transaksi GORM ↔ domain Tx
    model/                       model GORM (tag + TableName)
    mapper/                      model ↔ entity
    querysafety/                 sanitasi ORDER BY & LIKE
    <modul>/                     implementasi repository, 1 file = 1 method
  routes/router.go               satu-satunya tempat registrasi route
  service/<nama>/                implementasi domain/service, 1 file = 1 call
  usecase/<modul>/               logika bisnis, 1 file = 1 use case
migrations/                      SQL migrasi berurutan
docs/                            aturan, panduan, kontrak API
```

## Aturan dependensi (MUST)

| Package | Boleh import | Tidak boleh import |
|---|---|---|
| `domain/*` | stdlib, `uuid`, sesama `domain/*` | `internal/*`, gorm, fiber, go-common, viper |
| `internal/handler` | `domain/*`, `internal/apperr`, fiber, go-common (response, middleware) | `internal/repository`, `internal/usecase/<impl>`, gorm |
| `internal/usecase` | `domain/*`, `internal/apperr`, go-common `logger`, logrus | fiber, gorm, `internal/repository`, `internal/handler` |
| `internal/repository/psql` | `domain/*`, gorm, sesama `psql/*` | `internal/apperr`, fiber, `internal/usecase` |
| `internal/service` | `domain/*`, net/http, go-common `utils` | `internal/repository`, `internal/usecase` |
| `internal/bootstrap` | semua (composition root) | — |

> Kenapa `domain/` steril? Karena interface & tipe di sana dipakai oleh semua lapis.
> Kalau `domain` mengimpor gorm/fiber, perubahan framework merambat ke mana-mana dan
> mock tidak bisa dibuat tanpa menarik seluruh dependensi.

## Tanggung jawab per lapis

| Lapis | Melakukan | TIDAK melakukan |
|---|---|---|
| Handler | parse query/params/body, konversi tipe (UUID, enum), ambil identitas caller, panggil 1 usecase, map Output → DTO, pilih status sukses | cek permission, aturan bisnis, query DB, transaksi, format response error |
| Usecase | permission, validasi, aturan bisnis, orkestrasi repo/service, transaksi, logging | parsing HTTP, SQL, format JSON |
| Repository | 1 operasi data, mapping model↔entity, propagasi error teknis | aturan bisnis, `apperr`, memanggil repository lain, membuka transaksi sendiri |
| Service (outbound) | panggil API eksternal, mapping respons ke Output domain, timeout | aturan bisnis |

## Composition root

`internal/bootstrap/setup.go` adalah satu-satunya tempat `NewXxx(...)` konkret
dipanggil. Urutan: **services → repositories → txManager → usecases → handlers**,
lalu dikembalikan sebagai `routes.RouteDependencies`. Tidak ada variabel global
atau `init()` yang menyimpan dependensi.

## Kapan membuat modul baru vs memperluas modul

- **Modul baru** (`internal/{handler,usecase,repository/psql}/<modul>`) untuk setiap
  agregat/tabel utama yang punya endpoint sendiri.
- Endpoint yang mengubah data milik agregat lain tetap ditaruh di modul **pemilik
  route** (mis. `/periods/:id/…` di modul `period_*`), dan usecase-nya boleh memakai
  beberapa repository.
- Usecase **tidak** memanggil usecase lain. Logika bersama → fungsi di `helpers.go`
  modul tersebut, atau method pada entity bila murni domain.
