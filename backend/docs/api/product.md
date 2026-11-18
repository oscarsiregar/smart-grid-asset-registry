# API — Product (contoh modul CRUD)

Base path: `/v1/products`. Semua endpoint membutuhkan header dari BFF:

| Header | Isi |
|---|---|
| `X-Internal-Secret` | secret bersama BFF ↔ service (403 bila salah/tidak ada) |
| `Authorization` | `Bearer <JWT HS256>` dengan klaim `id`, `name`, `permissions` (401 bila tidak valid) |

Envelope response:

```json
{ "code": 200, "status": "success", "message": "…", "data": { } }
```

## Objek `Product`

| Field | Tipe | Keterangan |
|---|---|---|
| `id` | string (UUID) | |
| `code` | string | unik, ≤ 50 karakter, di-trim |
| `name` | string | ≤ 255 karakter, di-trim |
| `description` | string \| null | |
| `price` | number | ≥ 0 |
| `status` | `"ACTIVE"` \| `"INACTIVE"` | |
| `createdAt`, `updatedAt` | string (RFC3339) | |
| `createdByName`, `updatedByName` | string | nama user dari JWT |

## Endpoint

### `GET /v1/products` — daftar

Permission: `product.get_list`

| Query | Default | Keterangan |
|---|---|---|
| `search` | — | cocok sebagian (case-insensitive) pada `code` atau `name`; `%`/`_` diperlakukan literal |
| `status` | — | `ACTIVE` / `INACTIVE` |
| `page` | 1 | mulai 1 |
| `limit` | 10 | maksimum 100 |
| `sort` | `created_at` | `code`, `name`, `price`, `status`, `created_at`, `updated_at` (lainnya → default) |
| `order` | `DESC` | `ASC` / `DESC` |

200 → `data: { "items": [Product], "totalItems": 42 }`

### `GET /v1/products/:id` — detail

Permission: `product.get_detail` · 200 → `data: Product`

### `POST /v1/products` — buat

Permission: `product.create`

```json
{ "code": "P-001", "name": "Widget", "description": "opsional", "price": 12.5, "status": "ACTIVE" }
```

201 → `data: Product`

### `PUT /v1/products/:id` — ganti seluruh field yang bisa diubah

Permission: `product.update` · body sama dengan POST (field yang tidak dikirim menjadi nilai kosong/nol) · 200 → `data: Product`

### `DELETE /v1/products/:id` — hapus (hard delete)

Permission: `product.delete` · 204 tanpa body

## Error

| HTTP | Kode | Pesan | Kapan |
|---|---|---|---|
| 400 | `INVALID_REQUEST` | `Permintaan tidak valid: <detail>` | JSON rusak, UUID/enum salah, validasi gagal |
| 401 | — | `unauthorized: missing or invalid bearer token` | JWT tidak ada/invalid |
| 403 | — | `Direct access forbidden: request must route through BFF` | `X-Internal-Secret` salah |
| 403 | `PERMISSION_DENIED` | `Anda tidak memiliki izin untuk mengakses data ini` | permission tidak dimiliki |
| 404 | `PRODUCT_NOT_FOUND` | `Produk tidak ditemukan` | id tidak ada |
| 409 | `PRODUCT_CODE_ALREADY_EXISTS` | `Kode produk <code> sudah digunakan` | kode dipakai produk lain |
| 500 | — | `Internal Server Error` | kegagalan teknis (detail hanya di log) |
