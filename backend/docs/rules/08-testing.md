# 08 — Testing

## Target

| Lapis | Alat | Yang wajib dites | Coverage |
|---|---|---|---|
| `domain/valueobject` | testify | Parse (valid/invalid), Scan (string, []byte, tipe salah, nilai salah), Value, String, IsValid | 100% |
| `domain/usecase` (Validate/SetDefault) | testify, table-driven | setiap aturan validasi & default | 100% |
| `internal/usecase` | mockery mocks + `StubWithinTx` | **setiap cabang**: sukses, permission denied, validasi, not found, conflict, setiap error repo, rollback | **≥ 80% (gate CI)**, target 100% |
| `internal/repository/psql` | sqlmock (regexp) | SQL yang dihasilkan, sukses, error DB, trx diberikan, Tx tipe salah, not found → nil,nil | tinggi |
| `internal/handler` | `fiber.App.Test` + mock usecase | parse error → 400, input diteruskan benar, mapping response, status sukses, AppError → status, error biasa → 500 | tinggi |
| mapper / helpers / querysafety | testify | semua field, nil, slice kosong | 100% |

Tidak dites (dikecualikan di CI): `cmd/`, `internal/bootstrap`, `internal/routes`, `internal/mocks`, package yang hanya deklarasi (`domain/repository|service|dto|entity|errorcode`, `psql/model`).

## Pola test usecase (table-driven, mock per kasus)

```go
tests := []struct {
	name    string
	input   usecase.CreateProductInput
	setup   func(d deps)          // pasang EXPECT() khusus kasus ini
	wantErr error
}{
	{
		name:  "error code already exists rolls back",
		input: validInput,
		setup: func(d deps) {
			d.withinTx()                                    // WithinTx menjalankan callback dgn d.trx
			d.repo.EXPECT().FindByCode(mock.Anything, d.trx, "P-001").Return(sampleProduct(), nil)
			d.trx.EXPECT().Rollback().Return(nil)           // memastikan rollback terjadi
		},
		wantErr: apperr.Conflict(errorcode.ProductCodeAlreadyExists, "P-001"),
	},
}
for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		d := newDeps(t)       // mock BARU tiap kasus
		tt.setup(d)
		output, err := d.usecase().CreateProduct(ctx, tt.input)
		assert.Equal(t, tt.wantErr, err)
	})
}
```

Aturan:
1. **Mock dibuat per kasus** (`newDeps(t)`). Mockery mendaftarkan `AssertExpectations` lewat `t.Cleanup`, jadi panggilan yang diharapkan tapi tidak terjadi → gagal; panggilan tak terduga → panic/gagal. Ini yang membuktikan "repository tidak dipanggil saat permission ditolak".
2. Gunakan `EXPECT()` (typed), bukan `On("Method")`.
3. Pastikan transaksi benar: argumen `d.trx` (bukan `mock.Anything`) untuk repo yang harus berada di dalam transaksi, dan `Commit()` / `Rollback()` diharapkan secara eksplisit.
4. Waktu dipaku: `now: func() time.Time { return fixedNow }`.
5. Cocokkan entity yang dikirim ke repo dengan `mock.MatchedBy(func(p *entity.Product) bool {...})` untuk memverifikasi field yang dihitung usecase (trim, audit, timestamp).
6. Bandingkan error dengan `assert.Equal(t, tt.wantErr, err)` — `apperr` menghasilkan nilai yang bisa dibandingkan, sehingga kode & pesan ikut terverifikasi.

## Pola test repository (sqlmock)

```go
gdb, mock := newMockDB(t) // regexp matcher; ExpectationsWereMet dicek di Cleanup
mock.ExpectQuery(`SELECT \* FROM "product" WHERE id = \$1 ORDER BY "product"\."id" LIMIT \$2`).
	WithArgs(id, 1).
	WillReturnRows(productRow(sqlmock.NewRows(productColumns), id, "P-001"))
```

- Tulis SQL yang diharapkan **selengkap mungkin** (bukan hanya `SELECT`) — inilah yang menangkap regresi seperti `Save` → upsert atau zero value yang tidak ter-update.
- Write dengan GORM default dibungkus transaksi: `ExpectBegin` → `ExpectExec` → `ExpectCommit` / `ExpectRollback`.
- Kasus wajib per method: sukses, error DB, `wrapTx(gdb)`, `unsupportedTx{}`; plus `nil, nil` untuk not found.

## Pola test handler

```go
app := newTestApp(NewProductHandler(uc), identityWith(valueobject.PermissionCreateProduct))
resp := doRequest(t, app, "POST", "/products", body)
require.Equal(t, fiber.StatusCreated, resp.StatusCode)
```

- `newTestApp` memakai `utils.NewErrorHandler` yang sama dengan produksi, dan middleware yang menyuntikkan identitas seperti `Identity()`.
- Untuk input invalid, buat mock usecase **tanpa** EXPECT — membuktikan usecase tidak dipanggil.
- Verifikasi bentuk JSON (camelCase, `items` bukan `null`, format waktu RFC3339).

## Mock

- Dihasilkan dari interface `domain/` oleh mockery (`.mockery.yaml`): `make mocks`.
- **Jangan edit** file hasil generate. Helper manual (mis. `txhelpers.go`) diberi komentar "Hand-written" dan tidak ditimpa mockery.
- Commit mock hasil generate bersama perubahan interface.

## Perintah

```sh
make test     # unit test (paket sama dengan CI)
make cover    # coverage internal/usecase, gate 80%
go test ./internal/usecase/product/ -run TestCreateProduct -v   # satu test
```

## Smoke test end-to-end (sebelum rilis besar)

1. Jalankan Postgres sementara, terapkan `migrations/*.up.sql`.
2. `make run` dengan `.env` yang menunjuk DB tersebut.
3. Panggil endpoint dengan header `X-Internal-Secret` dan `Authorization: Bearer <JWT HS256 berisi id, name, permissions>`.
4. Verifikasi: 403 tanpa secret, 401 tanpa JWT, 403 tanpa permission, alur CRUD, 409 duplikat, 404 setelah delete.
