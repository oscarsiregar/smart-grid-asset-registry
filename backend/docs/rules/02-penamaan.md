# 02 — Penamaan & Struktur File

## Satu file = satu unit

| Lapis | Satu file berisi | Contoh |
|---|---|---|
| handler | 1 endpoint | `internal/handler/product/create_product.go` |
| usecase | 1 use case (method interface) | `internal/usecase/product/create_product.go` |
| repository | 1 method repository | `internal/repository/psql/product/find_by_code.go` |
| service | 1 panggilan API eksternal | `internal/service/user_management/get_user_by_id.go` |
| valueobject | 1 enum | `domain/valueobject/product_status.go` |
| domain lain | 1 file per modul | `domain/dto/product.go`, `domain/usecase/product.go` |

Setiap file di atas yang berisi logika **wajib** punya pasangan `_test.go` dengan nama sama.

File pendukung per package:

- `init.go` — interface handler (khusus handler), struct implementasi, constructor `NewXxx`.
- `helpers.go` — fungsi privat bersama dalam package (`hasPermission`, `toProductResponse`).
- `testutil_test.go` / `testdb_test.go` / `helpers_test.go` — helper khusus test.

## Package & import alias

| Lokasi | Nama package | Alias import di bootstrap |
|---|---|---|
| `internal/handler/product` | `product` | (tanpa alias) |
| `internal/usecase/product` | `product_usecase` | `product_usecase` |
| `internal/repository/psql/product` | `product_repository` | `product_repository` |
| `internal/service/user_management` | `user_management_service` | `user_management_service` |
| mock | `mocks` | `mocksrepo`, `mocksusecase`, `mockssvc` |
| go-common middleware fiber | — | `pkg_middleware_fiber` |

Nama folder modul = snake_case singular nama agregat (`product`, `covenant_config`).

## Nama method (seragam di SEMUA lapis)

| Operasi | Usecase / Handler | Repository |
|---|---|---|
| Daftar + paging | `GetListProduct` | `Find(filter)` → `([]*E, total, error)` |
| Detail | `GetProductDetail` | `FindByID`, `FindByCode`, `FindByXxx` |
| Semua tanpa paging | — | `FindAll` |
| Buat | `CreateProduct` | `Create` |
| Ubah | `UpdateProduct` | `Update` |
| Hapus | `DeleteProduct` | `Delete`, `DeleteByXxx` |
| Aksi domain | kata kerja + objek: `ApprovePeriod`, `SubmitPeriodForVerification` | — |

Nama handler, usecase, Input/Output, DTO, permission, dan nama file harus **sejalan**:

```
CreateProduct  →  create_product.go
               →  usecase.CreateProductInput / CreateProductOutput
               →  dto.CreateProductRequest  (response umum: dto.ProductResponse)
               →  valueobject.PermissionCreateProduct = "product.create"
```

## Tipe

- Struct implementasi **privat** (`productUsecase`), constructor publik mengembalikan **interface** (`usecase.ProductUsecase`).
- Receiver: singkatan nama struct (`pu productUsecase`, `r productRepository`, `h *productHandler`). Usecase & repository receiver by value; handler pointer.
- Field opsional/nullable → pointer (`*string`, `*int`, `*valueobject.X`). Field wajib → nilai biasa.
- ID: `uuid.UUID` di entity/usecase; `string` di DTO.
- Uang/angka: `float64` (DB `NUMERIC`). Bila presisi desimal krusial (akuntansi), sepakati tipe decimal di level proyek.
- Waktu: `time.Time` di entity; string RFC3339 di DTO response.

## Konvensi string

| Hal | Format | Contoh |
|---|---|---|
| JSON field | camelCase | `totalItems`, `createdByName` |
| Query param | camelCase | `?search=&page=&limit=&sort=&order=` |
| Path | kebab-case, plural | `/v1/products/:id`, `/v1/report-types` |
| Tabel DB | snake_case **singular** | `product`, `debtor_covenant_config` |
| Kolom DB | snake_case | `created_at`, `updated_by_name` |
| Index | `ix_<tabel>_<kolom>` / unik `ux_<tabel>_<kolom>` | `ux_product_code` |
| Enum value | SCREAMING_SNAKE_CASE | `ACTIVE`, `PENDING_APPROVAL` |
| Error code | SCREAMING_SNAKE_CASE | `PRODUCT_NOT_FOUND` |
| Permission | `<modul>.<aksi>` snake_case | `product.get_list`, `period.approve_sharia` |
| Env var | SCREAMING_SNAKE_CASE | `DB_MAX_OPEN_CONNS` |
| Pesan sukses | `"Successfully <aksi> <modul>"` | `"Successfully get list product"` |

## Komentar & bahasa

- Kode, identifier, komentar kode: **Bahasa Inggris**.
- Pesan error ke client: **Bahasa Indonesia** (katalog `apperr`, dengan padanan EN).
- Dokumentasi (`docs/`, README): Bahasa Indonesia.
- Komentar menjelaskan **kenapa**, bukan apa. Komentar insiden/perbaikan sertakan tanggal & alasan.
