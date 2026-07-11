# Deploy DRP-MikREST ke aaPanel (Standard Way)

Frontend & backend berjalan terpisah. aaPanel/Nginx serve langsung `frontend/dist/`,
dan proxy `/api/` + `/uploads/` ke backend Go.

## Prasyarat

- VPS Ubuntu 20.04 / 22.04 / 24.04
- aaPanel terinstall
- Domain pointing ke IP VPS

## 1. Install aaPanel

```bash
wget -O install.sh http://www.aapanel.com/script/install-ubuntu_6.0_en.sh && bash install.sh
```

Buka aaPanel, Install:
- **Nginx** (via App Store)
- **PostgreSQL 15+** (via App Store)
- **Go 1.23+** (via Terminal atau manual)
- **Node.js 18/20** (via App Store)

## 2. Clone Project

```bash
mkdir -p /www/wwwroot && cd /www/wwwroot
git clone https://github.com/dianrp-space/DRP-MikREST.git
cd DRP-MikREST
```

## 3. Setup Database

aaPanel → **Database** → **Add Database**:

| Field | Isi |
|-------|------|
| Database | `DRP_MikREST` |
| User | `postgres` |
| Password | **(isi sendiri)** |
| Access | `127.0.0.1` |

## 4. Build Frontend

```bash
cd /www/wwwroot/DRP-MikREST/frontend
npm install
npm run build
# hasil: /www/wwwroot/DRP-MikREST/frontend/dist/
```

## 5. Setup Backend

```bash
cd /www/wwwroot/DRP-MikREST/backend
cp .env.example .env
nano .env
```

Sesuaikan:

```ini
APP_ENV=production
APP_PORT=8080
APP_SECRET=isi-random-min-32-karakter
CORS_ORIGIN=https://domain-anda.com

DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=(password dari step 3)
DB_NAME=DRP_MikREST
DB_SSLMODE=disable
DB_MAX_CONNS=10

JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h
```

Build binary:

```bash
go mod tidy
go build -o bin/api ./cmd/api
```

Test seed (opsional, Ctrl+C setelah seed):

```bash
./bin/api -seed-email=admin@domain.com -seed-pass=Admin12345
```

## 6. Setup Website di aaPanel

### 6a. Add Site

aaPanel → **Website** → **Add Site**:

| Field | Isi |
|-------|------|
| Domain | `domain-anda.com` |
| Description | `DRP-MikREST` |
| Root Path | `/www/wwwroot/DRP-MikREST/frontend/dist` |

### 6b. SSL

aaPanel → Website → domain → **SSL** → Let's Encrypt → centang → Apply.
Aktifkan **Force HTTPS**.

### 6c. Reverse Proxy (hanya untuk /api/ dan /uploads/)

aaPanel → Website → domain → **Reverse Proxy** → **Add**:

| Field | Isi |
|-------|------|
| Name | `api` |
| Target URL | `http://127.0.0.1:8080` |
| Subdirectory | `/api/` |
| Send Domain | `$host` |

**Advanced config:**

```nginx
proxy_set_header Host $host;
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
proxy_set_header X-Forwarded-Proto $scheme;
proxy_buffering off;
```

Ulangi untuk `/uploads/`:

| Field | Isi |
|-------|------|
| Name | `uploads` |
| Target URL | `http://127.0.0.1:8080` |
| Subdirectory | `/uploads/` |
| Send Domain | `$host` |

(Advanced config sama)

## 7. Jalankan Backend

### systemd (recommended)

```bash
cat > /etc/systemd/system/DRP-MikREST.service << 'EOF'
[Unit]
Description=DRP-MikREST API Server
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/www/wwwroot/DRP-MikREST/backend
ExecStart=/www/wwwroot/DRP-MikREST/backend/bin/api -seed-email=admin@domain.com -seed-pass=Admin12345
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now DRP-MikREST
systemctl status DRP-MikREST
```

### Atau Supervisor (via aaPanel)

aaPanel → App Store → **Supervisor** → install.
Lalu **Add Daemon**:

| Field | Isi |
|-------|------|
| Name | `DRP-MikREST` |
| Run User | `root` |
| Run Dir | `/www/wwwroot/DRP-MikREST/backend` |
| Start Command | `/www/wwwroot/DRP-MikREST/backend/bin/api -seed-email=admin@domain.com -seed-pass=Admin12345` |

## 8. Verifikasi

Buka `https://domain-anda.com` → harus muncul login.
Cek API:

```bash
curl https://domain-anda.com/api/web/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@domain.com","password":"Admin12345"}'
```

## Update Aplikasi

```bash
cd /www/wwwroot/DRP-MikREST
git pull

# Frontend
cd frontend && npm install && npm run build

# Backend
cd ../backend && go build -o bin/api ./cmd/api

# Restart
systemctl restart DRP-MikREST   # systemd
# atau restart via aaPanel Supervisor
```

## Catatan

| Hal | Catatan |
|-----|---------|
| **Backup DB** | aaPanel → Database → backup, atau `pg_dump` |
| **Log** | `journalctl -u DRP-MikREST -f` (systemd), atau lihat di Supervisor |
| **Uploads** | Backend otomatis buat folder `backend/uploads/` — akses via `/uploads/` |
| **CORS** | `CORS_ORIGIN` harus diisi domain biar ga ditolak browser |
