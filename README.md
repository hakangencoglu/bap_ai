# BAP AI - Bilimsel Araştırma Projesi Sistemi

Bu proje, üniversitelerde kullanılan Bilimsel Araştırma Projesi (BAP) süreçlerini dijitalleştirmek amacıyla geliştirilen bir yönetim sistemidir.

## Teknolojiler
- **Backend:** Go 1.25 (Gin Framework)
- **Database:** PostgreSQL 16.0
- **Frontend:** Go Templates & Static Assets (HTML/CSS/JS)

## Kurulum ve Çalıştırma

### Docker ile Çalıştırma (Önerilen)

Projeyi Docker kullanarak en hızlı şekilde ayağa kaldırabilirsiniz. Bu yöntemle Go ve PostgreSQL bağımlılıkları otomatik olarak kurulur ve yapılandırılır.

**Gereksinimler:**
- Docker
- Docker Compose

**Adımlar:**
1. Proje ana dizinindeyken aşağıdaki komutu çalıştırın:
   ```bash
   docker-compose up -d --build
   ```
2. Uygulama başlatıldıktan sonra tarayıcıdan veya API istemcinizden şu adrese erişebilirsiniz:
   `http://localhost:8080`

**Önemli Notlar:**
- `docker-compose` başlatıldığında `migrations` klasöründeki SQL dosyaları otomatik olarak çalıştırılarak veritabanı şeması oluşturulur.
- Logları takip etmek için: `docker-compose logs -f app` komutunu kullanabilirsiniz.

### Yerel Geliştirme (Local Development)

1. Bağımlılıkları yükleyin:
   ```bash
   go mod download
   ```
2. `.env` dosyasındaki `DB_HOST` değerini `localhost` (veya yerel DB IP'niz) olarak güncelleyin.
3. Uygulamayı başlatın:
   ```bash
   go run cmd/server/main.go
   ```

## Proje Yapısı
- `cmd/server/`: Uygulamanın giriş noktası.
- `backend/`: API, Servis ve Repository katmanları.
- `frontend/`: HTML şablonları (`templates`) ve statik dosyalar (`static`).
- `migrations/`: Veritabanı SQL şema dosyaları.
- `configs/`: Çevresel değişkenlerin yönetimi.