# drp-mikrest

Aplikasi web manajemen Mikrotik (RouterOS) dengan fitur utama **pembuatan
voucher hotspot** (seperti Mikhmon) via API native RouterOS (port 8728).

## Stack

- **Backend**: Go (Fiber) + pgx (PostgreSQL) + go-routeros/routeros v3
- **Frontend**: Nuxt 3 + Vue 3 + Pinia + Tailwind CSS
- **Database**: PostgreSQL 15+

Lihat [PLAN.md](./PLAN.md) untuk perencanaan lengkap (skema, fitur, roadmap).

## Prasyarat

- Go 1.23+
- Node.js 18+ / 20+
- PostgreSQL 15+ (berjalan di localhost:5432)

## Setup Database

```bash
psql "postgresql://postgres:34267793@localhost:5432/postgres" -c "CREATE DATABASE drp_mikrest;"
```

Migrasi dijalankan otomatis saat backend start (folder
`backend/internal/db/migrations`).

## Setup Backend

```bash
cd backend
cp .env.example .env       # sesuaikan nilai, terutama APP_SECRET & DB_PASSWORD
go mod tidy
go run ./cmd/api -seed-email=admin@drp.test -seed-pass=Admin12345
```

Server mendengarkan di `http://localhost:8080`.
Admin awal dibuat jika tabel users masih kosong.

Endpoint utama:
- `POST /api/web/auth/login` - login email+password
- `GET  /api/web/servers` - list server router
- `POST /api/web/servers` - tambah server (host, port, user, pass)
- `POST /api/web/vouchers/generate` - generate voucher batch
- `POST /api/web/tokens` - buat API token per user
- `POST /api/v1/vouchers` - generate voucher via API token (external)

Lihat dokumentasi API lengkap di PLAN.md bagian 11.

## Setup Frontend

```bash
cd frontend
cp .env.example .env       # opsional, default arah ke localhost:8080
npm install
npm run dev
```

Frontend di `http://localhost:5173`.

## Penggunaan

1. Buka `http://localhost:5173`, login dengan admin awal.
2. Tambah server Mikrotik (host, port 8728, user, password).
3. Klik "Test Koneksi" untuk memastikan router terjangkau.
4. Buka server -> Voucher -> "+ Generate Voucher".
5. Pilih profile, jumlah, mode (random/prefix/same), pola.
6. Voucher dibuat di DB & di-push ke RouterOS via API.
7. Untuk integrasi sistem lain: buat API Token di halaman Token,
   lalu panggil `POST /api/v1/vouchers` dengan `Authorization: Bearer <token>`.

## Catatan Keamanan

- Password server RouterOS dienkripsi AES-256-GCM sebelum disimpan.
- API token disimpan sebagai hash SHA-256 (bukan plaintext).
- JWT untuk sesi web, API token untuk integrasi eksternal.
- Jangan commit file `.env`. Gunakan `.env.example` sebagai template.

## Pengembangan

```bash
# backend
make vet          # go vet
make build        # build binary
make tidy         # go mod tidy
make migrate-new name=add_index   # buat file migrasi baru

# frontend
npm run build     # build production
npm run typecheck # cek tipe TypeScript
```

## Lisensi

Lihat repository. (TBD)
