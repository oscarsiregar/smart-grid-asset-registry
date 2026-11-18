# AGENTS.md

Semua aturan proyek ada di **[CLAUDE.md](CLAUDE.md)** (ringkasan mengikat) dan
**[docs/rules/](docs/rules/)** (detail + alasan). Baca keduanya sebelum mengubah kode.

Perintah penting:

| Tujuan | Perintah |
|---|---|
| Jalankan lokal | `make run` |
| Format + vet + test (wajib sebelum push) | `make check` |
| Lint | `make lint` |
| Regenerate mock setelah ubah interface `domain/` | `make mocks` |
| Coverage usecase (gate ≥ 80%) | `make cover` |

Implementasi referensi: modul `product` (`domain/*/product*.go`, `internal/*/product/`).
