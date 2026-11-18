# 03 — Handler (HTTP)

Handler adalah **adapter tipis** antara Fiber dan usecase. Jika handler Anda lebih
dari ±50 baris atau memuat `if` yang bukan soal parsing, logika itu milik usecase.

## Anatomi standar

```go
// CreateProduct godoc
// POST /v1/products
func (h *productHandler) CreateProduct(ctx *fiber.Ctx) error {
	// 1. parse ke DTO request
	var request dto.CreateProductRequest
	if err := ctx.BodyParser(&request); err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	// 2. konversi tipe (UUID, enum) — gagal = 400
	status, err := valueobject.ParseProductStatus(request.Status)
	if err != nil {
		return apperr.BadRequest(errorcode.InvalidRequest, err.Error())
	}

	// 3. identitas caller
	actorName, permissions := caller(ctx)

	// 4. panggil SATU usecase dengan ctx.UserContext()
	output, err := h.productUsecase.CreateProduct(ctx.UserContext(), usecase.CreateProductInput{...})
	if err != nil {
		return err // ErrorHandler global yang memformat
	}

	// 5. map Output → DTO response
	return response.Success(ctx, fiber.StatusCreated, "Successfully create product", toProductResponse(output.Product))
}
```

## Aturan

1. **Parsing**: `QueryParser` (GET list), `ParamsParser` (path `:id`), `BodyParser` (POST/PUT/PATCH).
   Untuk PUT/PATCH yang punya path param + body, panggil `ParamsParser` lalu `BodyParser` pada struct yang sama.
2. **Konversi**: UUID dengan `uuid.Parse`, enum dengan `valueobject.ParseXxx`. Pesan error parse UUID: `"id must be a valid UUID"` (jangan bocorkan detail parser).
3. **Identitas**: hanya dari `pkg_middleware_fiber.GetUserIdentity(ctx)` (diisi middleware). Caller tanpa identitas → `Permissions: []string{}` sehingga usecase menolak.
   Jangan pernah membaca header `X-User-*` sendiri.
4. **Context**: selalu `ctx.UserContext()` — membawa trace id dan user id untuk audit plugin. Jangan meneruskan `*fiber.Ctx` ke usecase.
5. **Status sukses**:

   | Operasi | Status | Body |
   |---|---|---|
   | GET list / detail | 200 | `response.Success` |
   | POST create | 201 | entitas yang dibuat |
   | PUT / PATCH | 200 | entitas setelah diubah |
   | DELETE | 204 | kosong (`ctx.SendStatus`) |
   | Aksi async | 202 | status job |

6. **Error**: `return err`. Jangan `c.Status(...).JSON(...)` untuk error. `utils.NewErrorHandler` (di `bootstrap/fiber.go`) memetakan `apperror.AppError` ke status-nya dan menyembunyikan error lain sebagai 500.
7. **Response list**: `{ "items": [...], "totalItems": N }`. `items` **tidak boleh `null`** — inisialisasi `make([]T, 0, n)`.
8. **Mapping** entity → DTO di helper `toXxxResponse` (satu bentuk response per entitas untuk semua endpoint).
9. Handler **tidak** memanggil repository, service, atau usecase lain, dan tidak memanggil lebih dari satu method usecase.

## Envelope response (dari go-common)

```json
{ "code": 200, "status": "success", "message": "Successfully get list product", "data": { ... } }
{ "code": 404, "status": "error",   "message": "Produk tidak ditemukan",        "data": null }
```

## Registrasi route

Hanya di `internal/routes/router.go`, di bawah grup `authenticated`:

```go
productGroup := authenticated.Group("/products")
productGroup.Get("/", deps.ProductHandler.GetListProduct)
productGroup.Post("/", deps.ProductHandler.CreateProduct)
productGroup.Get("/:id", deps.ProductHandler.GetProductDetail)
productGroup.Put("/:id", deps.ProductHandler.UpdateProduct)
productGroup.Delete("/:id", deps.ProductHandler.DeleteProduct)
```

Route statis (`/months`, `/bulk-generate`) didaftarkan **sebelum** route ber-parameter (`/:id`) agar tidak tertangkap sebagai id.

## Konvensi REST

| Kebutuhan | Method & path |
|---|---|
| list | `GET /v1/<modul-plural>` |
| detail | `GET /v1/<modul-plural>/:id` |
| buat | `POST /v1/<modul-plural>` |
| ganti penuh | `PUT /v1/<modul-plural>/:id` |
| ubah sebagian / aksi | `PATCH /v1/<modul-plural>/:id[/<aksi>]` (mis. `/submit`, `/approve`) |
| hapus | `DELETE /v1/<modul-plural>/:id` |
| sub-resource | `/v1/periods/:id/report-compliance/:catalogId` |
