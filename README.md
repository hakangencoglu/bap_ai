# 🌍 BAP AI - Bilimsel Araştırma Projeleri Yönetim Sistemi

BAP AI, üniversitelerde yürütülen Bilimsel Araştırma Projeleri (BAP) süreçlerini uçtan uca dijitalleştirmek, yönetmek ve izlemek amacıyla geliştirilmiş modern, modüler ve güvenli bir web tabanlı yönetim sistemidir. 

Sistem; akademisyenler, hakemler, dekanlıklar, komisyon üyeleri ve TTO (Teknoloji Transfer Ofisi) yetkilileri gibi tüm paydaşların tek bir çatı altında koordineli şekilde çalışmasını sağlar.

---

## 🏗 Mimari Yapı ve Teknoloji Yığını

Uygulama, yüksek performans, esneklik ve kolay dağıtılabilirlik hedeflenerek modüler bir yapıda tasarlanmıştır:

```
┌─────────────────────────────────────────────────────────────┐
│                    ANA SUNUCU (App Server)                  │
│                                                             │
│   ┌─────────────────────────────────────────────────────┐   │
│   │               Docker Compose Network                │   │
│   │                                                     │   │
│   │  ┌───────────────────┐           ┌───────────────┐  │   │
│   │  │   Go App (Backend)│           │ PostgreSQL 16 │  │   │
│   │  │   + Frontend      │◄─────────►│ (db service)  │  │   │
│   │  │   Port: 8080      │           │ Port: 5432    │  │   │
│   │  └───────────────────┘           └───────────────┘  │   │
│   │                                                     │   │
│   └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

*   **Backend:** Go >= 1.23 ve yüksek performanslı **Gin Web Framework**.
*   **Database:** PostgreSQL >= 16.0 (Otomatik şema senkronizasyonu destekli).
*   **Frontend:** Go HTML Şablonları (Templates), Vanilla HTML, CSS ve JavaScript (Modern, hafif ve yüksek hızlı arayüzler, Harici Tailwind CSS veya ağır JS framework bağımlılığı olmadan).
*   **Konteynerleştirme:** Docker & Docker Compose.
*   **Otomasyon (CI/CD):** GitLab CI/CD pipeline desteği.

---

## 🔐 Temel Özellikler ve Modüller

### 1. Kullanıcı Yetkilendirme ve Rol Yönetimi (Accounts)
*   **JWT Tabanlı Güvenli Oturum:** Kullanıcı giriş ve kayıt işlemleri JWT (JSON Web Token) kullanılarak güvenli bir şekilde doğrulanır.
*   **Esnek Rol Yetkilendirme:** Sistemde bulunan roller:
    *   **Akademisyen (Araştırmacı):** Proje başvurusu yapar, ekip kurar, bütçe yönetir.
    *   **Hakem:** Atanan projeleri değerlendirir, puanlar ve yorumlar.
    *   **Admin:** Tüm sistemi yönetir, hakem atamaları gerçekleştirir, kullanıcı yetkilerini düzenler.
    *   **Dekan, Komisyon ve TTO:** Proje onay/red akışlarını takip eder ve karar verir.
*   **Zorunlu Profil Tamamlama:** İlk girişte araştırmacıların akademik bilgilerini tamamlamaları istenir.

### 2. Çok Adımlı BAP Başvuru Formu ve Otomatik Kaydetme
*   **5 Adımlı Dinamik Başvuru Akışı:** Proje Başlığı/Türü, Proje Ekibi, Bütçe Detayları, Proje Özeti ve Ek Dosyalar olmak üzere yapılandırılmış form yapısı.
*   **Otomatik Kaydetme (Autosave):** Form doldurulurken yapılan değişiklikler belirli aralıklarla arka planda otomatik olarak kaydedilir. Bağlantı kopmaları veya sayfa kapanmalarında veri kaybı önlenir.
*   **Dinamik Ekip Yönetimi:** Projeye araştırmacı, bursiyer veya danışman ekleme; davet gönderme ve onay/red mekanizması.

### 3. Hakem Değerlendirme Modülü (Hakem Sistemi)
*   **Gelişmiş Hakem Atama:** Admin paneli üzerinden projelere hakem arama ve manuel hakem atama süreci.
*   **Kabul/Red İş Akışı:** Hakemler kendilerine atanan projeleri inceleyip değerlendirmeyi kabul edebilir veya gerekçe belirterek reddedebilirler.
*   **Detaylı Değerlendirme:** Kabul edilen projeler için puanlama, görüş/yorum bildirme ve nihai tavsiye kararını içeren form yapısı.

### 4. Admin Yönetim Paneli ve Kanban Takibi
*   **İstatistik Paneli:** Toplam proje sayısı, bütçe dağılımları, onay bekleyen ve aktif projelerin anlık grafikleri.
*   **Kanban ve Liste Görünümü:** Projelerin başvuru aşamasından onaylanmasına kadar olan tüm yaşam döngüsünü (Taslak, Hakemde, Revizyonda, Dekanda, Komisyonda, Aktif vb.) görsel olarak izleme ve sürükle-bırak mantığıyla takip etme olanağı.
*   **Kullanıcı ve Rol Yönetimi:** Kullanıcı hesaplarını onaylama, engelleme veya rol değişikliklerini anlık gerçekleştirme.

### 5. Onay Süreci (Workflow) & Süreç Geçmişi
*   Projenin geçeceği onay basamakları (Dekan Onayı -> Komisyon Onayı -> TTO Aktifleşmesi) adım adım izlenir.
*   Tüm onay, red, hakem atama ve revizyon işlemleri **Süreç Geçmişi** tablosunda zaman damgalı ve işlem yapan kişi detaylarıyla loglanır.

### 6. Revizyon (Düzeltme) Yönetimi
*   Admin veya hakemler tarafından eksik veya hatalı görülen başvurular için revizyon talebi oluşturulabilir.
*   Proje sahibine revizyon talimat notları gönderilir ve proje düzenleme modunda yeniden açılır.

### 7. PDF Önizleme ve Arşivleme
*   Başvurunun son haline getirilip kilitlenmesinden önce **PDF Önizleme** motoru ile belgenin çıktısı incelenebilir.
*   Onaylanan başvurular sunucu tarafında güvenli bir şekilde arşivlenir.

---

## 🌍 Sunuculara Kurulum ve Çalıştırma (Deployment)

Uygulamanın sunucu kurulumu Docker ve Docker Compose kullanılarak oldukça pratik bir şekilde gerçekleştirilir.

### Sistem Gereksinimleri
*   Docker (v20.10+ önerilir)
*   Docker Compose (v2.0+ önerilir)
*   Git

---

### Adım Adım Kurulum Kılavuzu

#### 1. Proje Kodlarını Sunucuya Çekme
Projeyi sunucunuzda çalıştırmak istediğiniz dizine (örneğin `/opt/bap_ai`) klonlayın:
```bash
git clone <proje-repo-url> /opt/bap_ai
cd /opt/bap_ai
```

#### 2. Ortam Değişkenlerini (.env) Yapılandırma
Sistem yapılandırması `.env` dosyası üzerinden okunmaktadır. Proje kök dizinindeki `.env.example` şablon dosyasını kopyalayarak kendi `.env` dosyanızı oluşturun:
```bash
cp .env.example .env
```
Ardından oluşturduğunuz `.env` dosyası içerisindeki parametreleri (Veritabanı bağlantı bilgileri, port ve `JWT_SECRET` gibi güvenlik anahtarlarını) sunucunuzun ihtiyaçlarına göre güncelleyin.

#### 3. Uygulamayı Başlatma
Docker Compose kullanarak tüm servisleri (Go Backend ve PostgreSQL Veritabanı) arka planda başlatın ve derleyin:
```bash
docker compose up -d --build
```

#### 4. Otomatik Veritabanı Geçişleri (Migrations)
Sistem ilk kez çalıştırıldığında veya yeni tablolar eklendiğinde **otomatik geçiş mekanizması** devreye girer:
*   Uygulama ayağa kalkarken `backend/database/schema.sql` dosyasını kontrol eder.
*   Eğer veritabanında ana tablolar (`uye` vb.) yoksa, tüm şemayı otomatik olarak oluşturur.
*   Veritabanı zaten kuruluysa veri kaybını önlemek amacıyla şemayı sıfırlamaz; ancak yeni eklenen ek rolleri, süreç durumlarını ve tabloları dinamik olarak veritabanına entegre eder.

Uygulama başarıyla kurulduktan sonra tarayıcınızdan `http://<sunucu-ip-adresi>:8080` adresine giderek sisteme erişebilirsiniz.

---

### Sunucu Yönetim ve Bakım Komutları

*   **Logları Canlı Takip Etmek İçin:**
    ```bash
    docker compose logs -f app
    ```
*   **Konteyner Durumlarını Kontrol Etmek İçin:**
    ```bash
    docker compose ps
    ```
*   **Uygulamayı Durdurmak İçin:**
    ```bash
    docker compose down
    ```
*   **Uygulamayı Durdurup Verileri de Temizlemek İçin (Dikkat! DB verileri silinir):**
    ```bash
    docker compose down -v
    ```

---

## 🤖 Otomatik CI/CD Dağıtımı (GitLab)

Projede GitLab CI/CD entegrasyonu tanımlıdır. `main` dalına kod gönderildiğinde (push) sunucuya otomatik olarak dağıtım yapılır.

### Gerekli GitLab CI/CD Değişkenleri
Otomatik dağıtımın çalışması için GitLab projenizin **Settings > CI/CD > Variables** menüsünden aşağıdaki değişkenleri tanımlamanız gerekmektedir:
*   `SSH_PRIVATE_KEY`: Hedef sunucuya şifresiz bağlanabilmek için SSH Private Key.
*   `SSH_KNOWN_HOSTS`: Sunucu parmak izi doğrulaması (known_hosts içeriği).
*   `SERVER_IP`: Hedef sunucunun IP adresi.
*   `DEPLOY_DIR`: Projenin sunucuda kurulduğu dizin yolu (Örn: `/opt/bap_ai`).

---

## 💻 Yerel Geliştirme (Local Development)

Projeyi kendi bilgisayarınızda geliştirmek istiyorsanız aşağıdaki adımları takip edebilirsiniz:

1.  Bilgisayarınızda **Go (>= 1.23)** ve **PostgreSQL (>= 16)** kurulu olduğundan emin olun.
2.  PostgreSQL üzerinde `bap_app` adında boş bir veritabanı oluşturun.
3.  Proje ana dizininde `.env` dosyasını oluşturup `DB_HOST=localhost` ve kendi lokal veritabanı kullanıcı bilgilerinizle düzenleyin.
4.  Bağımlılıkları yükleyin:
    ```bash
    go mod download
    ```
5.  Uygulamayı başlatın:
    ```bash
    go run cmd/server/main.go
    ```
6.  Tarayıcınızdan `http://localhost:8080` adresine girerek yerel çalışmayı test edin.

---

## 📁 Proje Dizin Yapısı

Proje, Go standartlarına ve modüler yapısına uygun şekilde aşağıdaki gibi organize edilmiştir:

```text
bap_ai/
├── README.md                    # 🌍 Proje Tanıtımı ve Dağıtım Kılavuzu
├── AGENTS.md                    # 🛠️ Geliştirme Standartları (Ajan Kılavuzu)
├── PROGRESS.md                  # 📈 Proje Gelişim Aşamaları ve Log Geçmişi
├── Dockerfile                   # Uygulama Docker derleme dosyası
├── docker-compose.yml           # Çoklu konteyner yapılandırma dosyası
├── cmd/                         # Uygulamanın giriş noktası (Entry Point)
│   └── server/
│       └── main.go              # Sunucuyu başlatan ana dosya
├── backend/                     # Backend Çalışma Alanı
│   ├── api/                     # HTTP Katmanı (Handlers & Middlewares)
│   ├── service/                 # İş Mantığı Katmanı (Business Logic)
│   ├── repository/              # Veritabanı SQL Sorgu Katmanı (Go)
│   ├── models/                  # Veritabanı Struct Yapıları
│   └── database/                # DB Bağlantı ve Otomatik Migration/Şema Dosyaları
├── frontend/                    # Frontend Çalışma Alanı
│   ├── templates/               # Dinamik HTML Şablonları (Templates)
│   └── static/                  # CSS, JS, Görseller ve Tasarım Dosyaları
├── configs/                     # Uygulama ve Çevre Yapılandırmaları
├── uploads/                     # Sunucu Tarafında Yüklenen Dosyalar ve PDF Arşivi
├── tests/                       # Entegrasyon ve Birim Testleri
├── go.mod                       # Go Modül Tanımları
└── go.sum                       # Go Bağımlılık İmzaları
```