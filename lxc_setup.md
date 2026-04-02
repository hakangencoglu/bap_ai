# LXC Dual-Container Setup Guide

Bu rehber, BAP AI projesini iki ayrı LXC konteyneri üzerinde (biri veritabanı, diğeri uygulama) nasıl kuracağınızı açıklar.

## Mimari Genel Bakış

```
┌─────────────────┐     TCP/5432     ┌─────────────────┐
│  LXC: bap-app   │ ──────────────►  │  LXC: bap-db    │
│  (Uygulama)     │                  │  (PostgreSQL)   │
│  Port: 8080     │                  │  Port: 5432     │
└─────────────────┘                  └─────────────────┘
```

---

## LXC 1: Veritabanı Sunucusu (`bap-db`)

### 1. PostgreSQL Kurulumu
```bash
sudo apt update
sudo apt install postgresql-16
```

### 2. Ağ Yapılandırması

PostgreSQL'in dış bağlantıları kabul etmesi için `/etc/postgresql/16/main/postgresql.conf` dosyasını düzenleyin:
```conf
listen_addresses = '*'
```

`/etc/postgresql/16/main/pg_hba.conf` dosyasına uygulama konteynerinin erişim izni ekleyin:
```conf
# TYPE  DATABASE        USER            ADDRESS                 METHOD
host    bap_app         bap             10.0.3.0/24             scram-sha-256
```

> **Not:** `10.0.3.0/24` kısmını kendi LXC ağ aralığınıza göre ayarlayın.

### 3. Veritabanı ve Kullanıcı Oluşturma
```bash
sudo -u postgres psql
CREATE DATABASE bap_app;
CREATE USER bap WITH PASSWORD 'bap_admin_1';
GRANT ALL PRIVILEGES ON DATABASE bap_app TO bap;
\c bap_app
GRANT ALL ON SCHEMA public TO bap;
```

### 4. Firewall Ayarları (Opsiyonel)
```bash
sudo ufw allow from 10.0.3.0/24 to any port 5432
```

### 5. Servisi Yeniden Başlatma
```bash
sudo systemctl restart postgresql
sudo systemctl enable postgresql
```

---

## LXC 2: Uygulama Sunucusu (`bap-app`)

### Seçenek A: Docker ile Çalıştırma (Önerilen)

#### 1. Docker Kurulumu
```bash
sudo apt update
sudo apt install docker.io docker-compose-plugin
sudo systemctl enable docker
```

#### 2. Projeyi Klonlama
```bash
git clone <proje-repo-url> /opt/bap_ai
cd /opt/bap_ai
```

#### 3. Çevre Değişkenlerini Ayarlama
```bash
cp .env.example .env
nano .env
```

`DB_HOST` değerini veritabanı konteynerinin IP adresi ile güncelleyin:
```env
DB_HOST=10.0.3.100   # bap-db konteynerinin IP adresi
DB_PORT=5432
DB_USER=bap
DB_PASSWORD=bap_admin_1
DB_NAME=bap_app
```

#### 4. Uygulamayı Başlatma
```bash
docker compose up -d --build
```

#### 5. Durum Kontrolü
```bash
docker ps                    # Konteyner durumu ve health check
docker compose logs -f       # Canlı loglar
```

---

### Seçenek B: Doğrudan Çalıştırma (Go ile)

#### 1. Go Kurulumu
```bash
sudo apt update
sudo apt install golang-go   # veya resmi Go binary'si
```

#### 2. Çevre Değişkenlerini Ayarlama
`.env` dosyasında `DB_HOST` kısmına bap-db konteynerinin IP adresini yazın:
```env
DB_HOST=10.0.3.100
DB_PORT=5432
DB_USER=bap
DB_PASSWORD=bap_admin_1
DB_NAME=bap_app
```

#### 3. Derleme ve Çalıştırma
```bash
cd /opt/bap_ai
go mod download
go build -o main ./cmd/server/main.go
./main
```

> **Not:** Uygulama başlatıldığında migration'lar otomatik çalıştırılır.

---

## Bağlantı Testi

Uygulama konteynerinden veritabanına erişimi test etmek için:
```bash
# bap-app konteynerinden
sudo apt install postgresql-client
psql -h <bap-db-ip> -U bap -d bap_app
```
