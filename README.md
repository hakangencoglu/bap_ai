# BAP AI - Bilimsel Araştırma Projesi Sistemi 🚀

Bu proje, üniversitelerde kullanılan Bilimsel Araştırma Projesi (BAP) süreçlerini dijitalleştirmek amacıyla geliştirilen bir yönetim sistemidir. Uygulama baştan sona modüler, ölçeklenebilir ve **tamamen Dockerize edilmiş** bir yapıda çalışmaktadır.

## 🛠 Kullanılan Teknolojiler
- **Backend:** Go >= 1.23 (Gin Framework)
- **Database:** PostgreSQL >= 16.0
- **Frontend:** Go Templates & HTML/CSS/JS (Go sunucusu üzerinden statik olarak yayınlanır)
- **Deployment:** Docker & Docker Compose & GitLab CI/CD

---

## 🌍 Sunucu Üzerinde Çalıştırma (Deployment)

Projemiz, hiçbir ek bağımlılık kurmadan (Go vb.) doğrudan **Docker** kullanılarak her ortamda çalıştırılabilir. Sunucunuzda projeyi ayağa kaldırmak için aşağıdaki adımları sıfırdan takip edebilirsiniz.

### 1. Gereksinimler
Sunucunuzda aşağıdaki araçların kurulu olması yeterlidir:
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

### 2. Projeyi Sunucuya Çekme
Projeyi sunucunuza (örneğin `/opt/bap_ai` dizinine) klonlayın veya indirin.
```bash
git clone <proje-repo-url> /opt/bap_ai
cd /opt/bap_ai
```

### 3. Çevre Değişkenlerini (Environment Variables) Ayarlama
Docker yapılandırmasının düzgün çalışması için gerekli ortam değişkenlerini oluşturmanız gerekir. Projede yer alan `.env.example` dosyasını kopyalayarak işe başlayın:
```bash
cp .env.example .env
```
`.env` dosyasını açarak (örneğin `nano .env`) veritabanı şifresi veya JWT anahtarı gibi bilgileri sunucunuza (production) uygun şekilde güncelleyin. Docker iç ağı kullanıldığı için `DB_HOST=db` olarak kalmalıdır.

### 4. Uygulamayı Başlatma
Konfigürasyonları tamamladıktan sonra projenin tüm imajlarını oluşturup arkaplanda başlatmak için şu komutu çalıştırın:
```bash
docker-compose up -d --build
```
Bu komut sırasıyla şunları yapacaktır:
- PostgreSQL veritabanını başlatır.
- `migrations/` klasöründeki SQL dosyalarını çalıştırarak tüm tablolarınızı oluşturur.
- Go backend uygulamasını (frontend dosyalarını da statik olarak sunacak şekilde) derler ve çalıştırır.

Uygulamanız başlatıldıktan sonra `http://<sunucu-ip>:8080` adresi üzerinden erişim sağlayabilirsiniz.
Logları canlı takip etmek için:
```bash
docker-compose logs -f
```

---

## 🤖 Otomatik CI/CD Deployment (GitLab)

Bu projede ayrıca **GitLab CI/CD** kullanılarak tam otomatik dağıtım (deployment) süreci yapılandırılmıştır. `main` dalına kod gönderdiğinizde, proje GitLab Runner üzerinden derlenir ve Docker registry'sine atılır. 

Sunucuya otomatik dağıtımı manuel tetiklemek veya entegre etmek için:
1. GitLab üzerindeki **CI/CD > Variables** kısmında `SSH_PRIVATE_KEY`, `SSH_KNOWN_HOSTS`, `SERVER_IP`, `DEPLOY_DIR` (varsayılan: `/opt/bap_ai`) vb. değerlerin doğru tanımlandığından emin olun.
2. GitLab pipelines arayüzünde "deploy_prod" aşamasını çalıştırdığınızda, sunucuya SSH ile bağlanılır, en son Docker imajı çekilir ve projeniz otomatik ayağa kalkar!

---

## 💻 Yerel Geliştirme (Local Development)

Projeyi Docker olmadan, yerel bilgisayarınızda (örneğin geliştirme yaparken) çalıştırmak isterseniz:
1. Go (`>= 1.23`) kurulu olduğundan emin olun.
2. `go mod download` ile bağımlılıkları yükleyin.
3. Kendi yerel veritabanınız için `.env` dosyasındaki `DB_HOST` değerini (`localhost` vs.) ayarlayın.
4. `go run cmd/server/main.go` komutuyla projeyi başlatın. İstekleri yerel olarak (http://localhost:8080) test edin.