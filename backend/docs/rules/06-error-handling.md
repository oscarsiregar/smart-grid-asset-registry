# 06 — Error Handling

## Alur

```
usecase: return apperr.NotFound(errorcode.ProductNotFound)
   │  (apperror.AppError{Code:404, Message:"Produk tidak ditemukan"})
handler: return err
   │
bootstrap/fiber.go → ErrorHandler = utils.NewErrorHandler(log)   (go-common)
   ├─ *apperror.AppError → status = Code, message = Message   (log bila ≥ 500)
   ├─ *fiber.Error       → status = Code
   └─ error lain         → 500 "Internal Server Error" (detail TIDAK dikirim, tapi di-log)
```

Akibatnya: error teknis (koneksi DB putus, query salah) **tidak pernah bocor** ke
client, dan error bisnis selalu punya status & pesan yang benar.

## Kapan memakai apa

| Situasi | Fungsi | HTTP |
|---|---|---|
| Body/query/path tidak bisa diparse, UUID/enum salah, `Validate()` gagal | `apperr.BadRequest(errorcode.InvalidRequest, detail)` | 400 |
| Tidak punya permission / bukan pemilik data | `apperr.Forbidden(errorcode.PermissionDenied)` atau kode spesifik | 403 |
| Data yang dirujuk tidak ada | `apperr.NotFound(errorcode.XxxNotFound)` | 404 |
| Bentrok state: duplikat, versi berubah | `apperr.Conflict(errorcode.XxxAlreadyExists, arg)` | 409 |
| Input valid tapi melanggar aturan bisnis (status salah, tidak berlaku) | `apperr.UnprocessableEntity(errorcode.Xxx, arg)` | 422 |
| Kegagalan teknis | `return err` apa adanya | 500 |

401 (token) dan 403 BFF (secret) dihasilkan middleware go-common — usecase tidak perlu menanganinya.

## Menambah kode error baru

1. `domain/errorcode/errorcode.go` — tambah konstanta **dan** masukkan ke slice `All`.
2. `internal/apperr/apperr.go` — tambah entri katalog:

   ```go
   errorcode.ProductCodeAlreadyExists: {
       ID: "Kode produk %s sudah digunakan",
       EN: "Product code %s is already in use",
   },
   ```
3. `TestCatalog_EveryErrorCodeHasAMessage` memastikan langkah 2 tidak terlewat.

## Aturan pesan

- Pesan ke client **Bahasa Indonesia**, sopan, bisa ditindaklanjuti user ("Kode produk P-001 sudah digunakan"), bukan jargon teknis.
- Argumen `%s` hanya untuk nilai yang aman ditampilkan (kode, id yang dikirim user). Jangan menyisipkan isi error DB, query, stack trace, atau data milik user lain.
- Satu kode = satu makna. Lebih baik kode spesifik (`PERIOD_NOT_EDITABLE`) daripada kode generik dengan pesan berbeda-beda.
- Usecase & handler **tidak boleh** memanggil `apperror.*` go-common langsung — selalu lewat `apperr` agar pesan terpusat.

## Error dari domain

`Input.Validate()` dan `valueobject.ParseXxx` mengembalikan `error` biasa (domain tidak
mengenal `apperr`). Lapis pemanggil yang membungkus:

```go
if err := input.Validate(); err != nil {
	return Output{}, apperr.BadRequest(errorcode.InvalidRequest, err.Error())
}
```

Karena pesan dari `Validate()` dikirim ke client, tulis pesannya singkat dan aman
(`"code is required"`, `"price must not be negative"`).

## Panic

- `middleware.RecoverConfig()` (dipasang pertama di `main.go`) mengubah panic menjadi 500 untuk request itu saja — service tetap hidup.
- Panic **bukan** mekanisme kontrol alur. Jangan `panic` untuk error yang bisa dikembalikan.
- `log.Fatal` hanya boleh di `main.go` / bootstrap saat startup.

## Wrapping

Saat menambah konteks pada error teknis gunakan `%w` agar `errors.Is/As` tetap bekerja:

```go
return nil, fmt.Errorf("user-management: get user %s: %w", id, err)
```

Jangan bandingkan error dengan `==`; gunakan `errors.Is` (linter `errorlint` aktif).
