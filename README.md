Bu proje, üniversitelerde kullanılan Bilimsel Araştırma Projesi (BAP) süreçlerini dijitalleştirmek amacıyla geliştirilen bir yönetim sistemidir. Uygulama modüler, ölçeklenebilir ve **Dockerize edilmiş tek bir server** üzerinde çalışmaktadır.

## 🏗 Mimari (Yeni)

```
┌─────────────────────────────────────────────────────────┐
│                    ANA SUNUCU (App Server)              │
│                                                         │
│   ┌─────────────────────────────────────────────────┐   │
│   │               Docker Compose Network            │   │
│   │                                                 │   │
│   │  ┌───────────────────┐       ┌───────────────────┐  │   │
│   │  │   Go App (Backend)│       │   PostgreSQL 16   │  │   │
│   │  │   + Frontend      │◄─────►│   (db service)    │  │   │
│   │  │   Port: 8080      │       │   Port: 5432      │  │   │
│   │  └───────────────────┘       └───────────────────┘  │   │
│   │                                                 │   │
│   └─────────────────────────────────────────────────┘   │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## 🛠 Kullanılan Teknolojiler
- **Backend:** Go >= 1.23 (Gin Framework)
- **Database:** PostgreSQL >= 16.0 (ayrı sunucuda)
- **Frontend:** Go Templates & HTML/CSS/JS
- **Deployment:** Docker & Docker Compose & GitLab CI/CD

---

## 🌍 Sunucu Üzerinde Çalıştırma (Deployment)

### Gereksinimler
- Docker
- Docker Compose

---

### Adım 1: Projeyi Çekme ve Hazırlama
```bash
git clone <proje-repo-url> /opt/bap_ai
cd /opt/bap_ai
```

### Adım 2: Çevre Değişkenlerini Ayarlama
```bash
cp .env.example .env
# Gerekirse .env içindeki şifreleri güncelleyin (varsayılanlar docker-compose uyumludur)
```

### Adım 3: Uygulamayı Başlatma
```bash
docker compose up -d --build
```

Uygulama başlatıldığında:
1. Docker network üzerinde `db` servisi (PostgreSQL) ayağa kalkar.
2. Go App (Backend) başlar ve DB'ye bağlanır.
3. `migrations/` klasöründeki SQL dosyaları otomatik çalıştırılır.
4. `http://<sunucu-ip>:8080` adresinden erişim sağlanır.

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