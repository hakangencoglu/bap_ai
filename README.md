# BAP AI - Bilimsel Araştırma Projesi Sistemi 🚀

Bu proje, üniversitelerde kullanılan Bilimsel Araştırma Projesi (BAP) süreçlerini dijitalleştirmek amacıyla geliştirilen bir yönetim sistemidir. Uygulama modüler, ölçeklenebilir ve **iki sunuculu mimari** üzerinde çalışmaktadır.

## 🏗 Mimari

```
┌─────────────────────────┐         ┌─────────────────────────┐
│   ANA SUNUCU (App)      │         │   DB SUNUCUSU           │
│                         │         │                         │
│  ┌───────────────────┐  │         │  ┌───────────────────┐  │
│  │  Docker Container │  │  TCP    │  │   PostgreSQL 16   │  │
│  │  ┌─────────────┐  │  │ ──────► │  │   (Built-in)      │  │
│  │  │ Go Backend  │  │  │  5432   │  │                   │  │
│  │  │ + Frontend  │  │  │         │  └───────────────────┘  │
│  │  └─────────────┘  │  │         │                         │
│  └───────────────────┘  │         │  Veriler burada saklanır│
│                         │         │                         │
│  Port: 8080             │         │  Port: 5432             │
└─────────────────────────┘         └─────────────────────────┘
```

## 🛠 Kullanılan Teknolojiler
- **Backend:** Go >= 1.23 (Gin Framework)
- **Database:** PostgreSQL >= 16.0 (ayrı sunucuda)
- **Frontend:** Go Templates & HTML/CSS/JS
- **Deployment:** Docker & Docker Compose & GitLab CI/CD

---

## 🌍 Sunucu Üzerinde Çalıştırma (Deployment)

### Gereksinimler

| Sunucu | Gereksinimler |
|--------|---------------|
| **Ana Sunucu (App)** | Docker, Docker Compose |
| **DB Sunucusu** | PostgreSQL >= 16.0 |

---

### Adım 1: DB Sunucusu Kurulumu

DB sunucusunda PostgreSQL'in kurulu ve yapılandırılmış olması gerekir.

#### 1.1 PostgreSQL Ağ Yapılandırması
`/etc/postgresql/16/main/postgresql.conf` dosyasında:
```conf
listen_addresses = '*'
```

`/etc/postgresql/16/main/pg_hba.conf` dosyasında uygulama sunucusunun erişimine izin verin:
```conf
# TYPE  DATABASE        USER            ADDRESS                 METHOD
host    bap_app         bap             <APP_SUNUCU_IP>/32      scram-sha-256
```

#### 1.2 Veritabanı ve Kullanıcı Oluşturma
```bash
sudo -u postgres psql
CREATE DATABASE bap_app;
CREATE USER bap WITH PASSWORD 'bap_admin_1';
GRANT ALL PRIVILEGES ON DATABASE bap_app TO bap;
\c bap_app
GRANT ALL ON SCHEMA public TO bap;
```

PostgreSQL servisini yeniden başlatın:
```bash
sudo systemctl restart postgresql
```

---

### Adım 2: Uygulama Sunucusu Kurulumu

#### 2.1 Projeyi Çekme
```bash
git clone <proje-repo-url> /opt/bap_ai
cd /opt/bap_ai
```

#### 2.2 Çevre Değişkenlerini Ayarlama
```bash
cp .env.example .env
nano .env
```

**Önemli:** `DB_HOST` değerini DB sunucusunun IP adresi ile güncelleyin:
```env
DB_HOST=<DB_SUNUCU_IP>
DB_PORT=5432
DB_USER=bap
DB_PASSWORD=bap_admin_1
DB_NAME=bap_app
```

#### 2.3 Uygulamayı Başlatma
```bash
docker compose up -d --build
```

Uygulama başlatıldığında:
1. DB sunucusuna bağlanır
2. `migrations/` klasöründeki SQL dosyalarını otomatik çalıştırır
3. `http://<sunucu-ip>:8080` adresinden erişim sağlanır

Logları takip etmek için:
```bash
docker compose logs -f
```

Konteyner sağlık durumunu kontrol etmek için:
```bash
docker ps
```

---

## 🤖 Otomatik CI/CD Deployment (GitLab)

`main` dalına kod gönderildiğinde GitLab CI/CD ile otomatik dağıtım yapılır:

1. GitLab **CI/CD > Variables** kısmında aşağıdaki değişkenlerin tanımlı olduğundan emin olun:
   - `SSH_PRIVATE_KEY`, `SSH_KNOWN_HOSTS`, `SERVER_IP`, `DEPLOY_DIR`
2. Pipeline'da "deploy_prod" aşamasını çalıştırın

---

## 💻 Yerel Geliştirme (Local Development)

1. Go (`>= 1.23`) kurulu olduğundan emin olun
2. `.env` dosyasında `DB_HOST` değerini erişilebilir bir PostgreSQL sunucusuna ayarlayın
3. Bağımlılıkları yükleyin ve başlatın:
```bash
go mod download
go run cmd/server/main.go
```
4. `http://localhost:8080` adresinden test edin