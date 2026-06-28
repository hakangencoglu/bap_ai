# PROGRESS.md

🏗️ 1. Altyapı ve Yapılandırma

    [x] Proje ana dizin yapısının oluşturulması (bap_ai/)

    [x] AGENTS.md ve README.md dosyalarının hazırlanması

    [x] Go modülünün başlatılması (go mod init bap_ai)

    [x] Environment (.env) ve config yapısının kurulması

    [x] Docker ve docker-compose yapılandırması

    [x] Veritabanı migration sistemi (15 migration dosyası)

🔐 2. Accounts (Kullanıcı İşlemleri)

Sorumlular: Backend Dev, DB Admin, Frontend Dev

    [x] DB: Users ve Roles tablolarının tasarımı (Migrations)

    [x] Backend: JWT tabanlı kimlik doğrulama servisinin yazılması

    [x] Backend: Kayıt olma ve Giriş yapma fonksiyonları (Türkçe yorum satırlı)

    [x] Frontend: Login ve Register sayfalarının HTML/CSS tasarımı

    [x] Backend: AuthMiddleware ve AdminMiddleware (rol tabanlı erişim kontrolü)

    [x] Backend: RequireRoles middleware fonksiyonu

📝 3. BAP Başvuru Sistemi

Sorumlular: Backend Dev, DB Admin, Frontend Dev

    [x] DB: BAP başvuru formları ve türleri için modüler tablo yapısı

    [x] Frontend: Dinamik başvuru formu arayüzü (5 adımlı multi-step form)

    [x] Backend: Form verilerini işleyen servis katmanı (proje_handler, proje_service, proje_repository)

    [x] Frontend/JS: Belirli aralıklarla tetiklenen Otomatik Kaydetme (Autosave) mekanizması

    [x] Backend: Proje oluşturma, getirme ve güncelleme (CRUD) endpoint'leri

    [x] Backend: Proje takımı yönetimi (yürütücü otomatik atama)

📊 4. Dashboard ve Profil

Sorumlular: Backend Dev, Frontend Dev

    [x] Frontend: Ana sayfa (Dashboard) — istatistikler, son başvurular tablosu

    [x] Backend: Dashboard istatistikleri endpoint'i (aktif proje, onay bekleyen, tamamlanan, bütçe)

    [x] Backend: Son başvurular endpoint'i

    [x] Frontend: Profil sayfası — kişisel bilgiler, proje kartları

    [x] Backend: Profil bilgileri ve profil projeleri endpoint'leri

    [x] Frontend/JS: Rol bazlı dinamik sidebar menüsü

    [x] Frontend: Tema değiştirme (dark/light mode) ve i18n (çoklu dil desteği)

⚖️ 5. Hakem Sistemi

Sorumlular: Backend Dev, DB Admin, Frontend Dev

    [x] DB: Proje değerlendirmeleri tablosu (proje_degerlendirmeleri)

    [x] Backend: Otomatik hakem atama mekanizması (rastgele 2 hakem)

    [x] Frontend: Hakem dashboard sayfası — atanan projeler tablosu

    [x] Frontend: Hakem değerlendirme formu (puan, yorum, karar)

    [x] Backend: Değerlendirme kaydetme endpoint'i

    [x] Frontend: Proje detay modalı (hakem görünümü)

🛡️ 6. Admin Paneli

Sorumlular: Backend Dev, Frontend Dev

    [x] Frontend: Admin dashboard — istatistikler, proje yönetimi, kullanıcı yönetimi

    [x] Backend: Admin istatistikleri, tüm projeler, tüm kullanıcılar endpoint'leri

    [x] Backend: Kullanıcı rol ve durum güncelleme endpoint'leri

    [x] Backend: Proje durum güncelleme endpoint'i

    [x] Frontend: Admin proje durum raporları sayfası (Kanban + liste görünümü)

    [x] Backend: Proje detay endpoint'i (hakem yorumları, bütçe bilgileri)

🔄 7. Revizyon Sistemi

Sorumlular: Backend Dev, Frontend Dev

    [x] DB: Revizyonlar tablosu

    [x] Backend: Revizyon oluşturma ve aktif revizyon getirme endpoint'leri

    [x] Frontend: Revizyon atama modalı (ekip üyesi seçimi + talimat notu)

    [x] Frontend: Revizyon düzenleme modu (başvuru formunun edit hali)

    [x] Backend: Revizyon tamamlandığında otomatik durum güncelleme

📄 8. Önizleme ve PDF Modülü

Sorumlular: Backend Dev, Frontend Dev

    [x] Backend: Başvuru verilerini PDF formatına dönüştüren motorun kurulması

    [x] Frontend: PDF önizleme ekranının (Preview) entegrasyonu

    [x] Backend: Nihai başvurunun onaylanması ve dosya saklama mantığı

🐛 9. Bug Fix Geçmişi

    [x] (2026-05-04) hakem_handler.go: Context anahtarı düzeltmesi ("Uye" → "uye_id") — Hakem API'leri çalışmıyordu
    [x] (2026-05-04) hakem_repository.go: AssignRandomHakem sorgusu düzeltmesi (sistem_rol → uye.rol)
    [x] (2026-05-04) application_form.html: Edit modunda .page-subtitle → .page-description
    [x] (2026-05-04) application_form.html: bap_turu_id ve sure_ay backend'e doğru gönderilmiyordu
    [x] (2026-05-04) application_form.html: Öğrenci kısıtlama döngüsü live collection hatası
    [x] (2026-05-04) anasayfa.html: Profil dropdown linki "#" → "/profil"

🔄 10. Hakem Atama Sistemi (Elle Atama & Kabul/Red)

Sorumlular: Backend Dev, DB Admin, Frontend Dev

    [x] DB: Migration 021 — atama_durumu, red_nedeni, karar_tarihi sütunları (proje_degerlendirmeleri)

    [x] Backend: Admin hakem atama endpoint'leri (POST /admin/hakem-ata, GET /admin/hakemler, GET /admin/projeler/degerlendirme-bekleyen)

    [x] Backend: Hakem kabul/red karar endpoint'i (POST /hakem/karar)

    [x] Backend: Atama durumu kontrolü ile değerlendirme kısıtlaması (sadece kabul edenler değerlendirebilir)

    [x] Frontend: Admin Hakem Atama sayfası (admin_hakem_atama.html) — hakem arama, atama modalı, mevcut hakemler kartı

    [x] Frontend: Admin dashboard sidebar ve Hakem Ata butonu güncellenmesi

    [x] Frontend: Hakem dashboard güncellenmesi — atama durumu badge'ları, Kabul/Red butonları, red nedeni modalı

🤖 11. Yapay Zeka Sohbet Asistanı (Chatbot) Modülü

Sorumlular: Backend Dev, Frontend Dev

    [x] Yapılandırma: .env ve configs/config.go dosyalarına LLM çevre değişkenlerinin eklenmesi
    [x] Backend: ChatService (Ollama, OpenAI ve Gemini desteği) ve ChatHandler (POST /api/chat) katmanlarının oluşturulması
    [x] Backend: Çevrimdışı/Yerel çalışmada devreye giren akıllı Türkçe soru-cevap motoru
    [x] Frontend: styles.css ve chatbot.js ile premium cam efektli, Light/Dark mod uyumlu sohbet arayüzü tasarımı
    [x] Frontend: i18n.js üzerinden tüm dashboard sayfalarına dinamik entegrasyon
    [x] Doğrulama: Derleme testleri, Docker Compose ayağa kaldırma ve curl istekleri ile doğrulama adımları
    [x] Erişim Yönetimi: Chatbot'un sistem_sayfa tablosuna bir sayfa/özellik olarak eklenmesi
    [x] Erişim Yönetimi: Admin Sayfa Yetkileri yönetim matrisine chatbot entegrasyonu (Dinamik SQL migration)
    [x] Erişim Yönetimi: İstemci tarafında check-page-access sorgusu ile yetkisiz rollere chatbot gösterim engeli
    [x] Erişim Yönetimi: Sunucu tarafında chat_handler yetki kontrolü ile yetkisiz API çağrılarına 403 Forbidden engeli
    [x] Görsel İyileştirme: Yapay Zeka sunucu bağlantısını gösteren dinamik durum noktası (Yeşil/Kırmızı LED ışık)
    [x] Arayüz Tasarımı: Üzerine gelmek yerine sayfayı daraltarak açılan sabitlenmiş (pinned/docked) sağ panel düzeni