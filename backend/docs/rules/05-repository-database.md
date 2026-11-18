# 05 — Repository & Database

## Interface (di `domain/repository/<modul>.go`)

```go
type ProductRepository interface {
	Create(ctx context.Context, trx Tx, product *entity.Product) error
	Update(ctx context.Context, trx Tx, product *entity.Product) error
	Delete(ctx context.Context, trx Tx, id uuid.UUID) error
	FindByID(ctx context.Context, trx Tx, id uuid.UUID) (*entity.Product, error)
	FindByCode(ctx context.Context, trx Tx, code string) (*entity.Product, error)
	Find(ctx context.Context, trx Tx, filter FindProductFilter) ([]*entity.Product, int64, error)
}
```

- Parameter pertama selalu `ctx`, kedua selalu `trx Tx` (boleh `nil`).
- Menerima & mengembalikan **entity**, tidak pernah model GORM.
- Filter > 3 parameter → struct `FindXxxFilter` di package `repository`.

## Implementasi

```go
func (r productRepository) FindByID(ctx context.Context, trx repository.Tx, id uuid.UUID) (*entity.Product, error) {
	db, err := gormtx.Conn(ctx, r.db, trx) // trx bila ada, kalau tidak r.db.WithContext(ctx)
	if err != nil {
		return nil, err
	}

	var product model.Product
	err = db.Where("id = ?", id).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // not found BUKAN error
		}
		return nil, err
	}
	return mapper.ToEntityProduct(&product), nil
}
```

### MUST

1. Koneksi selalu dari `gormtx.Conn(ctx, r.db, trx)` — menjamin `ctx` (timeout, trace, user id audit) ikut dan transaksi dipakai bila diberikan.
2. **Not found = `nil, nil`** untuk `FindByXxx` tunggal; list kosong = slice kosong non-nil.
3. Error DB dikembalikan apa adanya (jangan dibungkus `apperr`, jangan di-log — usecase yang log).
4. Model GORM hanya di `psql/model`, konversi lewat `psql/mapper` (`ToModelXxx`, `ToEntityXxx`, `ToEntityXxxs` → slice non-nil).
5. Repository tidak memanggil repository lain dan tidak membuka transaksi sendiri.

### MUST NOT

| Jangan | Kenapa | Pakai |
|---|---|---|
| `db.Save(m)` untuk update | Bila row tidak ada, GORM diam-diam melakukan INSERT (upsert) | `db.Model(m).Select("kolom", ...).Updates(m)` — eksplisit, zero value ikut tersimpan |
| `db.Updates(struct)` tanpa `Select` | zero value (`0`, `""`, `false`, `nil`) **dilewati** | sertakan `Select(...)` |
| `fmt.Sprintf` / konkatenasi nilai user ke SQL | SQL injection | placeholder `?` |
| `Order(request.Sort)` mentah | ORDER BY tidak bisa diparameterisasi | `querysafety.SafeOrderClause(sort, order, default, defaultOrder, allowList...)` |
| `"%"+q+"%"` untuk LIKE | `%`/`_` dari user jadi wildcard | `querysafety.ContainsPattern(q)` |
| Field model bernama `CreatedBy`/`UpdatedBy` berisi nama | audit plugin go-common menimpa dengan UUID user JWT dan **membuang** nilai non-UUID | `CreatedByName`/`UpdatedByName` (atau `CreatedBy` bertipe UUID bila memang ingin audit otomatis) |
| `db.Use(database.NewAuditPlugin())` | sudah didaftarkan `database.Connect` → `gorm.ErrRegistered` | tidak perlu apa-apa |
| Reuse chain GORM untuk COUNT lalu SELECT | statement saling bocor | `query = query.Session(&gorm.Session{})` setelah WHERE dibangun |

## Paging list

```go
query := db.Model(&model.Product{})
// ... WHERE dinamis
query = query.Session(&gorm.Session{})

var total int64
query.Count(&total)

query.Order(querysafety.SafeOrderClause(filter.Sort, filter.Order, "created_at", "DESC", sortableColumns...)).
	Limit(filter.Limit).Offset((filter.Page - 1) * filter.Limit).
	Find(&rows)
```

- `sortableColumns` = allow-list kolom yang boleh di-sort, didefinisikan di `init.go` repository.
- `Limit <= 0` berarti tanpa batas (hanya untuk pemanggilan internal; usecase selalu mengisi default).

## Skema & migrasi

- Tabel **singular** snake_case; PK `id UUID`; kolom audit `created_at`, `created_by_name`, `updated_at`, `updated_by_name` (+ `deleted_at`, `deleted_by_name` bila soft delete).
- Enum disimpan `VARCHAR` + `CHECK (status IN (...))`, dipetakan ke `valueobject` yang mengimplementasikan `sql.Scanner` & `driver.Valuer`.
- Uang `NUMERIC(18,2)`, waktu `TIMESTAMPTZ`.
- Setiap kolom yang dipakai filter/sort/unik → index (`ix_…`, `ux_…`).
- Migrasi: `migrations/NNNNNN_deskripsi.up.sql` + `.down.sql`, nomor berurutan, **idempoten** (`IF NOT EXISTS`). Kompatibel dengan golang-migrate.
- Migrasi yang sudah masuk `main` **tidak boleh diubah** — buat migrasi baru.
- Perubahan yang merusak (rename/drop kolom) dilakukan bertahap: tambah kolom baru → deploy kode yang menulis keduanya → migrasi data → hapus kolom lama di rilis berikutnya.

## Soft delete vs hard delete

- Default contoh `product`: **hard delete**.
- Bila data punya nilai audit/legal (dokumen, approval): soft delete — kolom `status = DELETED` atau `deleted_at`, dan **setiap** query `Find*` wajib memfilter row terhapus. Dokumentasikan pilihannya di usecase.

## Integrasi eksternal (`internal/service`)

- Interface di `domain/service`, implementasi `internal/service/<nama>/` (1 file = 1 call), constructor menerima base URL & secret dari config.
- HTTP client dengan **timeout** eksplisit (go-common `utils.NewTracingClient(timeout)` meneruskan trace id).
- Request selalu `http.NewRequestWithContext(ctx, ...)`; tutup `resp.Body`.
- Non-2xx → error yang jelas (`fmt.Errorf("user-management: GET /users/%s: status %d", ...)`); 404 dari upstream yang bermakna "tidak ada" → `nil, nil`.
- Jangan `InsecureSkipVerify: true` di production. Bila terpaksa untuk dev (sertifikat self-signed), jadikan konfigurasi env, default `false`.
- Test dengan `httptest.NewServer`.
