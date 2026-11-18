# 04 — Usecase (Logika Bisnis)

Usecase adalah **satu-satunya** tempat keputusan bisnis dibuat: siapa boleh apa,
data mana yang valid, apa yang terjadi bila konflik, dan apa yang harus atomik.

## Kontrak (di `domain/usecase/<modul>.go`)

```go
type ProductUsecase interface {
	CreateProduct(ctx context.Context, input CreateProductInput) (CreateProductOutput, error)
	...
}

type CreateProductInput struct {
	Code   string
	...
	ActorName   string   // untuk kolom *_by_name
	Permissions []string // dari identitas caller
}

type CreateProductOutput struct {
	Product *entity.Product
}

// Validate: bentuk input saja (wajib isi, panjang, rentang, enum valid).
// Mengembalikan error biasa — domain tidak tahu apperr.
func (i CreateProductInput) Validate() error { ... }

// SetDefault: default paging/sorting, batas maksimum limit.
func (i *GetListProductInput) SetDefault() { ... }
```

- Input/Output selalu struct **nilai** bernama `<Method>Input` / `<Method>Output` — bahkan jika hanya satu field. Menambah field nanti tidak mengubah signature.
- Method tanpa data balikan (delete) boleh mengembalikan `error` saja.
- Output membawa **entity**, bukan DTO. Mapping ke JSON adalah urusan handler.

## Urutan langkah (MUST)

```go
func (pu productUsecase) UpdateProduct(ctx context.Context, input usecase.UpdateProductInput) (usecase.UpdateProductOutput, error) {
	// 1. logger ber-context
	log := logger.WithCtx(ctx, pu.log).WithField("usecase", "UpdateProduct")

	// 2. otorisasi — PALING AWAL, sebelum validasi atau query apa pun
	if !hasPermission(input.Permissions, valueobject.PermissionUpdateProduct) {
		log.Warn("permission denied")
		return usecase.UpdateProductOutput{}, apperr.Forbidden(errorcode.PermissionDenied)
	}

	// 3. validasi bentuk input
	if err := input.Validate(); err != nil {
		return usecase.UpdateProductOutput{}, apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	// 4-5. aturan bisnis + penulisan, atomik
	err := pu.txManager.WithinTx(ctx, func(ctx context.Context, trx repository.Tx) error {
		product, err := pu.productRepository.FindByID(ctx, trx, input.ID)
		if err != nil { log.WithError(err).Error("failed to find product by id"); return err }
		if product == nil { return apperr.NotFound(errorcode.ProductNotFound) }
		... // cek konflik, mutasi entity
		return pu.productRepository.Update(ctx, trx, product)
	})
	if err != nil {
		return usecase.UpdateProductOutput{}, err
	}

	// 6. log sukses dengan id yang relevan
	log.WithField("productID", input.ID).Info("success")
	return usecase.UpdateProductOutput{Product: product}, nil
}
```

## Otorisasi

- Satu permission per endpoint, didefinisikan di `domain/valueobject/permission.go`.
- Default **deny**: tidak ada permission = `403`.
- Otorisasi berbasis data (mis. hanya debitur kategori Syariah, hanya reviewer yang ditugaskan)
  juga di usecase, **setelah** data dibaca: pakai permission berpasangan
  (`xxx_sharia` / `xxx_conventional`) + helper `hasScopedPermissions` / `isContractCategoryPermitted`
  seperti di repo referensi, dan error `Forbidden` yang spesifik.
- Jangan mengandalkan frontend / BFF untuk menyembunyikan tombol sebagai otorisasi.

## Validasi: dua tingkat

| Jenis | Tempat | Error |
|---|---|---|
| Bentuk (wajib, panjang, rentang, enum) | `Input.Validate()` di domain | `400 INVALID_REQUEST` |
| Aturan yang butuh data (unik, status sesuai, relasi ada) | usecase | `404` / `409` / `422` dengan errorcode spesifik |

- Normalisasi (trim, upper-case kode) di usecase sebelum disimpan, dan dipakai konsisten untuk pengecekan keunikan.
- Pengecekan unik di usecase memberi pesan 409 yang ramah; **constraint unik di DB tetap wajib** sebagai jaminan saat ada request bersamaan.

## Transaksi

Gunakan `WithinTx` bila:
- ada **lebih dari satu** query tulis, atau
- ada pola **baca-lalu-tulis** yang hasilnya bergantung pada bacaan (cek unik lalu insert, cek status lalu update).

Aturan:
1. Semua pemanggilan repository di dalam callback **harus** menerima `trx` (bukan `nil`).
2. Kembalikan error dari callback → otomatis rollback. Jangan panggil `Commit/Rollback` manual.
3. Jangan memanggil API eksternal di dalam transaksi (menahan koneksi DB). Ambil data eksternal **sebelum** `WithinTx`.
4. Hasil yang perlu dikembalikan ditangkap lewat variabel di luar closure (`var product *entity.Product`).
5. Satu query tulis tunggal tidak perlu transaksi (lihat `DeleteProduct`).

## Lain-lain

- Waktu: `pu.now()` (field `now func() time.Time` diisi `time.Now` di constructor) — agar test bisa memaku waktu.
- ID baru: `uuid.New()` di usecase, bukan default DB, supaya entity lengkap sebelum disimpan dan bisa dikembalikan.
- Paging: `input.SetDefault()` lalu teruskan ke `repository.FindXxxFilter`. Limit maksimum `usecase.MaxLimit` (100).
- Usecase tidak memanggil usecase lain. Kode bersama → `helpers.go` di package itu.
- Fungsi privat yang dipanggil beberapa usecase tetap dites di `helpers_test.go`.
