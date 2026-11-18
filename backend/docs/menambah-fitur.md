# Menambah Modul / Endpoint Baru — Checklist

Contoh: membuat modul **`supplier`** dengan CRUD. Kerjakan berurutan dari dalam
(domain) ke luar (route); setiap langkah punya padanan di modul `product` — salin
lalu sesuaikan.

## 0. Skema

- [ ] `migrations/00000N_create_supplier.up.sql` + `.down.sql` (tabel singular, PK UUID, kolom audit `*_by_name`, index untuk kolom filter/sort/unik).

## 1. Domain (`domain/`)

- [ ] `valueobject/supplier_status.go` (+ `_test.go`) bila ada enum — salin `product_status.go`.
- [ ] `valueobject/permission.go` — tambah `PermissionGetListSupplier = "supplier.get_list"`, dst. (satu per endpoint).
- [ ] `errorcode/errorcode.go` — `SupplierNotFound`, … **dan** tambahkan ke `All`.
- [ ] `entity/supplier.go` — struct murni.
- [ ] `repository/supplier.go` — interface + `FindSupplierFilter`.
- [ ] `usecase/supplier.go` — interface + `XxxInput/XxxOutput` + `Validate()` / `SetDefault()` (+ `supplier_test.go`).
- [ ] `dto/supplier.go` — `XxxRequest` + `SupplierResponse`.

## 2. Error messages

- [ ] `internal/apperr/apperr.go` — entri katalog ID + EN untuk setiap kode baru.

## 3. Repository (`internal/repository/psql/`)

- [ ] `model/supplier.go` (tag gorm + `TableName()`).
- [ ] `mapper/supplier.go` (+ `_test.go`).
- [ ] `supplier/init.go` (`NewSupplierRepository`, `sortableColumns`), lalu satu file per method + `_test.go` (sqlmock), `testdb_test.go`.

## 4. Mock

- [ ] `make mocks` → `internal/mocks/repository/SupplierRepository.go`, `internal/mocks/usecase/SupplierUsecase.go`.

## 5. Usecase (`internal/usecase/supplier/`)

- [ ] `init.go` (struct + `NewSupplierUsecase`, field `now`), `helpers.go`.
- [ ] Satu file per use case mengikuti urutan: log → permission → validate → aturan → tulis (tx) → log sukses.
- [ ] `_test.go` per file + `helpers_test.go` (`deps`, `newDeps`, `withinTx`, fixture).

## 6. Handler (`internal/handler/supplier/`)

- [ ] `init.go` (interface `SupplierHandler` + constructor), `helpers.go` (`caller`, `toSupplierResponse`).
- [ ] Satu file per endpoint + `_test.go`, `testutil_test.go`.

## 7. Wiring

- [ ] `internal/bootstrap/setup.go` — repository → usecase → handler, masukkan ke `RouteDependencies`.
- [ ] `internal/routes/router.go` — field di `RouteDependencies` + grup `authenticated.Group("/suppliers")`.

## 8. Dokumentasi & verifikasi

- [ ] `docs/api/supplier.md` — endpoint, permission, request/response, error.
- [ ] Daftarkan permission baru ke identity-service / role mapping.
- [ ] `make check` dan `make lint` bersih; coverage usecase ≥ 80%.
- [ ] (Opsional) smoke test end-to-end — lihat [rules/08-testing.md](rules/08-testing.md).

## Menambah endpoint ke modul yang sudah ada

Langkah 1 (permission, errorcode, Input/Output, DTO) → 2 → method repository bila perlu
(3) → `make mocks` → usecase + test (5) → handler + test (6) → method di interface
handler & route (7) → docs (8).

## Menambah integrasi eksternal

1. `domain/service/<nama>.go` — interface + Input/Output.
2. `.mockery.yaml` — aktifkan package `smart-grid-asset-registry/domain/service` (dir `internal/mocks/service`).
3. `internal/service/<nama>/init.go` (constructor menerima base URL, secret, logger; HTTP client dengan timeout) + satu file per call + test dengan `httptest.NewServer`.
4. Env baru di `config.go` (+ `Validate`), `.env.example`, `values.yml`/Secret.
5. Inject di `bootstrap/setup.go` ke usecase yang membutuhkan.
