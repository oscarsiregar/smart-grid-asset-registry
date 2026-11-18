# 10 — Git, CI & Deploy

## Branch

| Prefix | Untuk |
|---|---|
| `feat/<deskripsi>` | fitur baru |
| `fix/<deskripsi>` | perbaikan bug |
| `refactor/<deskripsi>` | perubahan struktur tanpa ubah perilaku |
| `chore/<deskripsi>` | dependensi, tooling, CI |
| `docs/<deskripsi>` | dokumentasi |

- Dibuat dari `main`, masuk lewat Merge Request. Tidak ada push langsung ke `main` / `staging`.
- Alur environment: feature branch → `staging` (uji) → `main` (produksi).

## Commit — Conventional Commits

```
<type>(<scope>): <ringkasan imperatif, huruf kecil, tanpa titik>

<badan opsional: KENAPA perubahan ini, bukan apa>
```

Contoh:
```
feat(product): add bulk status update endpoint
fix(period): reject approve when reviewer is not assigned
docs: document maker-checker-approval status flow
```

- Satu commit = satu perubahan logis yang lolos test.
- Jangan commit: `.env`, binary (`*.exe`, `bin/`), `coverage.out`, file IDE.

## Merge Request

Checklist sebelum minta review:
- [ ] `make check` lulus (gofmt, vet, test)
- [ ] `make lint` bersih
- [ ] Test baru untuk setiap cabang logika baru; coverage usecase tidak turun
- [ ] Mock di-regenerate bila interface `domain/` berubah
- [ ] Migrasi baru (bila ada) + `.down.sql`
- [ ] `docs/api/<modul>.md` diperbarui bila kontrak API berubah
- [ ] `.env.example` + `config.Validate()` diperbarui bila ada env baru
- [ ] Checklist keamanan di [07-keamanan](07-keamanan.md)

Reviewer memeriksa juga: arah dependensi antar lapis, penamaan sesuai [02-penamaan](02-penamaan.md), tidak ada logika bisnis di handler/repository.

## CI (`.gitlab-ci.yml`)

| Job | Isi | Memblokir? |
|---|---|---|
| `lint` | `golangci-lint run` (konfigurasi `.golangci.yml`) | ya |
| `go_test` | gofmt check, `go vet`, unit test (tanpa paket wiring/mock/deklarasi) | ya |
| `coverage_check` | coverage `internal/usecase` ≥ 80% | hanya di `main`/`staging` |
| include `Infra/supply-chain` | build image, push ke Harbor, deploy | sesuai template Infra |

Modul privat go-common diambil dengan `CI_JOB_TOKEN` (`git config url.insteadOf`) dan `GOPRIVATE`.

## Build & image

- `Dockerfile` multi-stage: build `golang:1.26.x-alpine` (`CGO_ENABLED=0`, `-trimpath -ldflags="-s -w"`) → runtime `alpine` minimal, user non-root, `tzdata`.
- Konfigurasi dari env var (viper `AutomaticEnv`), **tidak ada** `.env` di image.

## Deploy (Kubernetes / Helm `values.yml`)

- Secret (`DB_PASSWORD`, `JWT_SECRET`, `INTERNAL_SERVICE_SECRET`) dari Kubernetes Secret; `values.yml` hanya env non-rahasia.
- `FIBER_PREFORK=false` dan `GOMAXPROCS` disesuaikan limit CPU.
- Probe `httpGet /v1/health` dengan `tcpSocket: null` (chart bersama default ke tcpSocket dan Helm menggabungkan map → Deployment ditolak bila keduanya ada).
- Sebelum deploy ke suatu environment: rebuild dari source terbaru dan pastikan env var menunjuk ke environment itu (bukan `localhost` dari `.env` lokal).
- Migrasi DB dijalankan **sebelum** versi kode yang membutuhkannya (job migrasi / pipeline), dan harus kompatibel mundur dengan versi kode yang sedang berjalan.

## Versi Go & tooling

| Tool | Versi |
|---|---|
| Go | 1.26.x (lihat `go.mod`) |
| golangci-lint | v2.13.x |
| mockery | v2.53.x |
| go-common | dipin di `go.mod` (`v1.0.0`) |

Naik versi Go/tool lewat MR `chore/` tersendiri dan perbarui tabel ini.
