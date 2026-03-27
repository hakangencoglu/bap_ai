# LXC Dual-Container Setup Guide

Bu rehber, BAP AI projesini iki ayrı LXC konteyneri üzerinde (biri veritabanı, diğeri backend) nasıl kuracağınızı açıklar.

## LXC 1: Veritabanı (Sunucu: `bap-db`)

### 1. Kurulum
```bash
sudo apt update
sudo apt install postgresql-16
```

### 2. Ağ Yapılandırması
PostgreSQL'in dış bağlantıları kabul etmesi için `/etc/postgresql/16/main/postgresql.conf` dosyasını düzenleyin:
```conf
listen_addresses = '*'
```

Ardından `/etc/postgresql/16/main/pg_hba.conf` dosyasına ikinci LXC konteynerinin (uygulama) IP adresini ekleyin:
```conf
# TYPE  DATABASE        USER            ADDRESS                 METHOD
host    bap_app         bap             10.0.3.0/24            scram-sha-256
```

### 3. Veritabanı ve Kullanıcı Oluşturma
```bash
sudo -u postgres psql
CREATE DATABASE bap_app;
CREATE USER bap WITH PASSWORD 'bap_admin_1';
GRANT ALL PRIVILEGES ON DATABASE bap_app TO bap;
```

---

## LXC 2: Uygulama (Sunucu: `bap-app`)

### 1. Proje Yapılandırması (.env)
Bu konteynerdeki `.env` dosyasında `DB_HOST` kısmına LXC 1'in IP adresini yazın:
```env
DB_HOST=10.0.3.100 # Örnek: Veritabanı konteynerinin IP adresi
DB_PORT=5432
DB_USER=bap
DB_PASSWORD=bap_admin_1
DB_NAME=bap_app
```

### 2. Çalıştırma
```bash
go run cmd/server/main.go
```
