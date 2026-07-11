<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="frontend/public/drp-mikrest.webp">
    <img src="frontend/public/drp-mikrest.webp" alt="DRP-MikREST" width="400">
  </picture>
</p>

<h3 align="center">DRP-MikREST</h3>

<p align="center">
  Aplikasi web manajemen Mikrotik (RouterOS) dengan fitur utama pembuatan voucher hotspot via API native RouterOS.
  <br>
  Alternatif open-source untuk Mikhmon.
</p>

<p align="center">
  <a href="#fitur">Fitur</a>&nbsp;|&nbsp;
  <a href="#tech-stack">Tech Stack</a>&nbsp;|&nbsp;
  <a href="#prasyarat">Prasyarat</a>&nbsp;|&nbsp;
  <a href="#instalasi">Instalasi</a>&nbsp;|&nbsp;
  <a href="#penggunaan">Penggunaan</a>&nbsp;|&nbsp;
  <a href="#struktur-proyek">Struktur</a>&nbsp;|&nbsp;
  <a href="#development">Development</a>&nbsp;|&nbsp;
  <a href="#keamanan">Keamanan</a>&nbsp;|&nbsp;
  <a href="#deployment">Deployment</a>&nbsp;|&nbsp;
  <a href="#lisensi">Lisensi</a>
</p>

---

## Fitur

- **Multi-Server Management** — Tambah, edit, hapus, dan uji koneksi ke banyak router Mikrotik sekaligus.
- **Integrasi Native RouterOS API** — Terhubung langsung via API native RouterOS port 8728 (bukan SSH/REST).
- **Generate Voucher Hotspot** — Batch generation hingga 500 voucher dengan 3 mode:
  - **Random** — Username/password acak
  - **Prefix** — Username dengan awalan khusus
  - **Same** — Username dan password sama
- **Manajemen Profile Hotspot** — Sync, buat, dan kelola profile hotspot dari/ke RouterOS.
- **Manajemen Voucher** — Lihat, nonaktifkan, aktifkan, hapus voucher. Status: `active`, `used`, `disabled`, `expired`, `failed`.
- **Monitoring User Aktif** — Lihat dan kick user hotspot aktif per server.
- **Scheduler Otomatis** — Cron scheduler untuk expirasi voucher, cleanup log, dll.
- **API Token untuk Integrasi Eksternal** — Buat token dengan scope-based access control & rate limiting.
- **Audit Log Lengkap** — Catat semua aktivitas dengan sumber (`web`, `api`, `system`).
- **Dashboard KPI** — Total server, status online/offline, jumlah voucher aktif, jumlah member aktif.
- **Pengaturan Aplikasi** — Nama aplikasi, logo, favicon, interval scheduler, retensi log.
- **Dark Theme** — UI dengan tema gelap modern menggunakan Tailwind CSS.

## Tech Stack

### Backend

| Komponen | Teknologi |
|---------|-----------|
| Bahasa | Go 1.23+ |
| Web Framework | Fiber v2 |
| Database | PostgreSQL 15+ (pgx v5) |
| RouterOS API | go-routeros/routeros v3 (native API) |
| JWT | golang-jwt/jwt v5 |
| Enkripsi | AES-256-GCM |
| Scheduling | robfig/cron v3 |
| Logging | rs/zerolog |

### Frontend

| Komponen | Teknologi |
|---------|-----------|
| Bahasa | TypeScript 5.x |
| Framework | Vue 3 (Composition API, `<script setup>`) |
| Build Tool | Vite 5.x |
| State Management | Pinia 2.x |
| Routing | Vue Router 4.x |
| Styling | Tailwind CSS 3.x |
| UI Components | SweetAlert2 |

## Prasyarat

- **Go** 1.23+
- **Node.js** 18+ / 20+
- **PostgreSQL** 15+ (berjalan di `localhost:5432`)

## Instalasi

### 1. Clone Repository

```bash
git clone https://github.com/username/DRP-MikREST.git
cd DRP-MikREST
```

### 2. Setup Database

```bash
psql "postgresql://postgres:password@localhost:5432/postgres" -c "CREATE DATABASE DRP_MikREST;"
```

Migrasi database dijalankan **otomatis** saat backend pertama kali start (folder `backend/internal/db/migrations`).

### 3. Setup Backend

```bash
cd backend
cp .env.example .env
# Edit .env: sesuaikan APP_SECRET, DB_PASSWORD, ENCRYPTION_KEY, dll.
go mod tidy
go run ./cmd/api -seed-email=admin@DRP-MikREST.test -seed-pass=Admin12345
```

Server berjalan di `http://localhost:8080`. Admin awal dibuat otomatis jika tabel `users` masih kosong.

### 4. Setup Frontend

```bash
cd frontend
cp .env.example .env   # opsional, default mengarah ke localhost:8080
npm install
npm run dev
```

Frontend berjalan di `http://localhost:5173`.

## Penggunaan

1. Buka `http://localhost:5173`, login dengan admin awal.
2. Tambah server Mikrotik (host, port 8728, username, password).
3. Klik **Test Koneksi** untuk memastikan router terjangkau.
4. Buka server &rarr; Voucher &rarr; **+ Generate Voucher**.
5. Pilih profile, jumlah, mode (random/prefix/same), dan pola.
6. Voucher dibuat di database & di-push ke RouterOS via API native.
7. Untuk integrasi sistem lain:
   - Buat **API Token** di halaman Tokens
   - Panggil `POST /api/v1/vouchers` dengan header `Authorization: Bearer <token>`

### Endpoint API Utama

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| `POST` | `/api/web/auth/login` | Login email + password |
| `GET` | `/api/web/servers` | Daftar server router |
| `POST` | `/api/web/servers` | Tambah server baru |
| `POST` | `/api/web/vouchers/generate` | Generate voucher batch |
| `POST` | `/api/web/tokens` | Buat API token |
| `POST` | `/api/v1/vouchers` | Generate voucher via API token (eksternal) |

## Struktur Proyek

```
DRP-MikREST/
├── backend/
│   ├── cmd/api/main.go              # Entry point backend
│   ├── internal/
│   │   ├── api/v1/router.go         # Route API eksternal
│   │   ├── config/                  # Konfigurasi dari .env
│   │   ├── crypto/                  # AES-256-GCM encrypt/decrypt
│   │   ├── db/
│   │   │   ├── db.go                # Koneksi database
│   │   │   ├── migrate.go           # Migration runner
│   │   │   └── migrations/          # File migrasi SQL
│   │   ├── handler/                 # HTTP handlers
│   │   ├── middleware/              # JWT, API token, rate limit, security
│   │   ├── models/                  # Data models
│   │   ├── repository/              # Data access layer
│   │   ├── routeros/                # RouterOS API client
│   │   ├── scheduler/               # Cron scheduler
│   │   ├── service/                 # Business logic
│   │   └── util/                    # Utility functions
│   ├── .env.example
│   ├── go.mod
│   └── Makefile
├── frontend/
│   ├── public/                      # Static assets (logo, favicon)
│   ├── src/
│   │   ├── components/              # Vue components
│   │   ├── composables/             # Vue composables (useApi, useSettings)
│   │   ├── layouts/                 # Layout components
│   │   ├── router/                  # Vue Router config
│   │   ├── stores/                  # Pinia stores
│   │   └── views/                   # Halaman aplikasi
│   ├── .env.example
│   ├── package.json
│   ├── tailwind.config.js
│   └── vite.config.ts
├── DEPLOY_AAPANEL.md                # Panduan deployment aaPanel
└── README.md
```

## Development

### Backend

```bash
cd backend
make tidy         # go mod tidy
make dev          # hot-reload dengan air
make run          # run tanpa hot-reload
make build        # build binary
make vet          # go vet
make test         # go test
make migrate-new name=add_index   # buat migrasi baru
```

### Frontend

```bash
cd frontend
npm run dev       # development server
npm run build     # build production
npm run typecheck # type checking TypeScript
npm run preview   # preview production build
```

## Keamanan

- **Password RouterOS** — Dienkripsi AES-256-GCM sebelum disimpan di database.
- **API Token** — Disimpan sebagai hash SHA-256 (bukan plaintext).
- **Password User** — Di-hash dengan bcrypt.
- **JWT** — Sesi web menggunakan access + refresh token dengan TTL yang dapat dikonfigurasi.
- **Rate Limiting** — Login: 10 percobaan/menit. API Token: per-token rate limit.
- **Security Headers** — Middleware secure headers di semua response.
- **CORS** — Dilindungi dengan konfigurasi origin yang ketat.

Jangan pernah commit file `.env`. Gunakan `.env.example` sebagai template.

## Deployment

Lihat [DEPLOY_AAPANEL.md](./DEPLOY_AAPANEL.md) untuk panduan deployment lengkap menggunakan **aaPanel** + Nginx reverse proxy + systemd service.

## Lisensi

Hak cipta dilindungi. Lihat repository untuk informasi lisensi lebih lanjut.
