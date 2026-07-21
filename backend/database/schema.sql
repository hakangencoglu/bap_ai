-- ==========================================
-- Migration: 001_create_lookup_tables.sql
-- ==========================================
-- ====================================================
-- Lookup (Referans) Tabloları
-- Sistemdeki tüm tanımlama/referans tablolarını oluşturur
-- ====================================================
-- Sistem Rol Tanımlama: Kullanıcı rollerinin isimleri
CREATE TABLE IF NOT EXISTS sistem_rol_tanimlama (
    rol_id SERIAL PRIMARY KEY,
    rol_adi VARCHAR(100) UNIQUE NOT NULL, -- Örn: admin, akademisyen, ogrenci, hakem
    rol_etiketi VARCHAR(100) -- Örn: Sistem Yöneticisi, Akademisyen
);
-- Proje Rol Tanımlama: Proje içindeki roller
CREATE TABLE IF NOT EXISTS proje_rol_tanimlama (
    rol_id SERIAL PRIMARY KEY,
    proje_rol VARCHAR(100) UNIQUE NOT NULL -- Örn: Yürütücü, Araştırmacı, Danışman
);
-- Proje Durum Tanımlama: Projenin genel durumu (taslak, incelemede, yururlukte vb.)
CREATE TABLE IF NOT EXISTS proje_durum (
    durum_id SERIAL PRIMARY KEY,
    durum_adi VARCHAR(100) UNIQUE NOT NULL, -- Örn: taslak, incelemede, reddedildi, revizyon, yururlukte, tamamlandi
    durum_etiketi VARCHAR(100) -- Örn: Taslak, İncelemede
);
-- Proje Aşama Tanımlama: İş akışındaki onay masaları (Dekan Onayına Sun, Komisyona Sun vb.)
-- asama_kodu dahili kod, asama_adi kullanıcıya gösterilen Türkçe addır.
CREATE TABLE IF NOT EXISTS proje_asama (
    asama_id SERIAL PRIMARY KEY,
    asama_kodu VARCHAR(100) UNIQUE NOT NULL,
    -- Örn: dekan_onayina_sun
    asama_adi VARCHAR(200) NOT NULL,
    -- Örn: Dekan Onayına Sun
    sira_no INTEGER DEFAULT 0 -- Akıştaki sıra
);
-- Proje BAP Türü: BAP proje türleri
CREATE TABLE IF NOT EXISTS proje_bap_turu (
    bap_turu_id SERIAL PRIMARY KEY,
    bap_turu VARCHAR(200) UNIQUE NOT NULL -- Örn: Yüksek Lisans, Doktora, Münferit
);
-- Proje Çıktı Türü: Çıktı türlerinin tanımları
CREATE TABLE IF NOT EXISTS proje_cikti_turu (
    cikti_turu_id SERIAL PRIMARY KEY,
    cikti_turu VARCHAR(200) UNIQUE NOT NULL -- Örn: Makale, Tez, Patent, Rapor
);
-- Proje Etik: Etik durum tanımları
CREATE TABLE IF NOT EXISTS proje_etik (
    etik_id SERIAL PRIMARY KEY,
    etik_durumu VARCHAR(200) UNIQUE NOT NULL -- Örn: Gerekli, Gerekli Değil, Alındı
);
-- Bütçe Tanım: Bütçe türü tanımları
CREATE TABLE IF NOT EXISTS butce_tanim (
    tanim_tip_id SERIAL PRIMARY KEY,
    tanim_adi VARCHAR(200) NOT NULL -- Örn: Sarf Malzeme, Hizmet Alımı, Yolluk
);
-- Bütçe Kategori: Bütçe kategorileri
CREATE TABLE IF NOT EXISTS butce_kategori (
    kategori_id SERIAL PRIMARY KEY,
    kategori_adi VARCHAR(200) UNIQUE NOT NULL -- Örn: Donanım, Yazılım, Seyahat
);
-- Olanak Türü: Olanak türleri
CREATE TABLE IF NOT EXISTS olanak_tur (
    olanak_tur_id SERIAL PRIMARY KEY,
    tur_adi VARCHAR(200) NOT NULL -- Örn: Laboratuvar, Kütüphane
);
-- ====================================================
-- Varsayılan Lookup Verileri
-- ====================================================
-- Sistem rolleri
INSERT INTO sistem_rol_tanimlama (rol_adi, rol_etiketi)
VALUES ('admin', 'Sistem Yöneticisi'),
    ('akademisyen', 'Akademisyen'),
    ('ogrenci', 'Öğrenci'),
    ('hakem', 'Hakem'),
    ('dekan', 'Fakülte Dekanı'),
    ('komisyon', 'BAP Komisyon Üyesi'),
    ('tto', 'TTO Temsilcisi') ON CONFLICT (rol_adi) DO UPDATE SET rol_etiketi = EXCLUDED.rol_etiketi;
-- Proje rolleri
INSERT INTO proje_rol_tanimlama (proje_rol)
VALUES ('Yürütücü'),
    ('Araştırmacı'),
    ('Danışman'),
    ('Bursiyer') ON CONFLICT (proje_rol) DO NOTHING;
-- Proje genel durumları (iş akışı ara durumları artık proje_asama tablosunda)
INSERT INTO proje_durum (durum_adi, durum_etiketi)
VALUES ('taslak', 'Taslak'),
    ('incelemede', 'İncelemede'),
    ('dekan_onayi_bekliyor', 'Dekan Onayı Bekliyor'),
    ('dekan_onayladi', 'Dekan Onayladı'),
    ('komisyon_bekliyor', 'Komisyon Onayı Bekliyor'),
    ('komisyon_onayladi', 'Komisyon Onayladı'),
    ('hakem_atama_bekliyor', 'Hakem Atama Bekleniyor'),
    ('hakem_bekliyor', 'Hakem İncelemesinde'),
    ('hakem_onayladi', 'Hakem Onayladı'),
    ('sozlesme_imza', 'Sözleşme / İmza Aşaması'),
    ('tto_aktif', 'TTO Onayı Bekliyor'),
    ('onaylandi', 'Onaylandı'),
    ('reddedildi', 'Reddedildi'),
    ('tamamlandi', 'Onaylandı (Tamamlandı)'),
    ('revizyon', 'Revizyon Gerekli'),
    ('yururlukte', 'Yürürlükte (Aktif)') ON CONFLICT (durum_adi) DO UPDATE SET durum_etiketi = EXCLUDED.durum_etiketi;
-- Proje aşamaları (iş akışı onay masaları)
INSERT INTO proje_asama (asama_kodu, asama_adi, sira_no)
VALUES ('tto_on_inceleme', 'TTO Ön İnceleme', 1),
    ('dekan_onayina_sun', 'Dekan Onayına Sun', 2),
    ('komisyona_sun', 'Komisyona Sun', 3),
    ('hakeme_sun', 'Hakeme Sun', 4),
    ('sozlesme_imza', 'Sözleşme İmzası', 5) ON CONFLICT (asama_kodu) DO NOTHING;
-- BAP türleri
INSERT INTO proje_bap_turu (bap_turu)
VALUES ('BAP-100'),
    ('BAP-200'),
    ('BAP-300'),
    ('BAP-400'),
    ('BAP-500') ON CONFLICT (bap_turu) DO NOTHING;
-- Çıktı türleri
INSERT INTO proje_cikti_turu (cikti_turu)
VALUES ('SCI/SSCI Makale'),
    ('Ulusal Makale'),
    ('Tez'),
    ('Patent'),
    ('Bildiri'),
    ('Rapor'),
    ('Kitap/Kitap Bölümü') ON CONFLICT (cikti_turu) DO NOTHING;
-- Etik durumları
INSERT INTO proje_etik (etik_durumu)
VALUES ('Gerekli'),
    ('Gerekli Değil'),
    ('Alındı'),
    ('Başvuru Yapıldı') ON CONFLICT (etik_durumu) DO NOTHING;
-- Bütçe kategorileri
INSERT INTO butce_kategori (kategori_adi)
VALUES ('Makine-Teçhizat'),
    ('Sarf Malzeme'),
    ('Hizmet Alımı'),
    ('Seyahat (Yolluk)'),
    ('Yazılım'),
    ('Yayın/Basım') ON CONFLICT (kategori_adi) DO NOTHING;
-- ==========================================
-- Migration: 002_create_uye_table.sql
-- ==========================================
-- ====================================================
-- Üye (Kullanıcı) Tablosu
-- Sistemdeki kullanıcıların temel bilgilerini tutar
-- Auth alanları (sifre_hash, aktif_mi) korunmuştur
-- ====================================================
CREATE TABLE IF NOT EXISTS uye (
    uye_id SERIAL PRIMARY KEY,
    rol VARCHAR(100),
    -- Kullanıcı rolü (metin olarak)
    ad VARCHAR(100) NOT NULL,
    -- Kullanıcı adı
    soyad VARCHAR(100) NOT NULL,
    -- Kullanıcı soyadı
    unvan VARCHAR(100),
    -- Akademik unvan (Prof. Dr., Doç. Dr. vb.)
    bolum VARCHAR(255),
    -- Bölüm bilgisi
    eposta VARCHAR(255) UNIQUE NOT NULL,
    -- E-posta adresi (giriş için kullanılır)
    telefon VARCHAR(15),
    -- İletişim telefonu
    izu_uyesi BOOLEAN DEFAULT FALSE,
    -- İZÜ üyesi mi?
    sifre_hash VARCHAR(255) NOT NULL,
    -- Şifrelenmiş parola (auth için)
    aktif_mi BOOLEAN DEFAULT TRUE,
    -- Hesap aktif mi? (auth için)
    sifre_degistir_zorla BOOLEAN DEFAULT FALSE,
    -- Şifre değiştirmeye zorla
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- E-posta alanı için hızlı arama indeksi
CREATE INDEX IF NOT EXISTS idx_uye_eposta ON uye(eposta);
-- ==========================================
-- Migration: 003_create_sistem_rol_table.sql
-- ==========================================
-- ====================================================
-- Sistem Rol Tablosu
-- Kullanıcı-rol atamasını yönetir
-- Bir kullanıcının hangi sistem rolüne sahip olduğunu tutar
-- ====================================================
CREATE TABLE IF NOT EXISTS sistem_rol (
    rol_id SERIAL PRIMARY KEY,
    uye_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    -- Kullanıcı referansı
    sistem_rol_id INTEGER NOT NULL REFERENCES sistem_rol_tanimlama(rol_id) ON DELETE CASCADE,
    -- Rol tanımı referansı
    UNIQUE(uye_id, sistem_rol_id) -- Bir kullanıcıya aynı rol bir kez atanabilir
);
-- Kullanıcı bazlı hızlı arama indeksi
CREATE INDEX IF NOT EXISTS idx_sistem_rol_uye_id ON sistem_rol(uye_id);
-- ==========================================
-- Migration: 004_create_proje_table.sql
-- ==========================================
-- ====================================================
-- Proje Ana Tablosu
-- BAP projelerinin temel bilgilerini tutar
-- ====================================================
CREATE TABLE IF NOT EXISTS proje (
    proje_id SERIAL PRIMARY KEY,
    proje_kodu VARCHAR(100) UNIQUE,
    -- Proje kodu (örn: 2026-BAP100-003)
    baslik_tr VARCHAR(500),
    -- Proje başlığı (Türkçe)
    baslik_en VARCHAR(500),
    -- Proje başlığı (İngilizce)
    sure_ay INTEGER,
    -- Proje süresi (ay cinsinden)
    toplam_butce NUMERIC(12, 2) DEFAULT 0,
    -- Toplam bütçe tutarı
    etik_kurul BOOLEAN DEFAULT FALSE,
    -- Etik kurul onayı gerekli mi?
    etik_kurul_no INTEGER,
    -- Etik kurul numarası
    koordinator_id INTEGER REFERENCES uye(uye_id) ON DELETE
    SET NULL,
        -- Proje koordinatörü (üye FK)
        durum_id INTEGER REFERENCES proje_durum(durum_id) ON DELETE
    SET NULL,
        -- Genel proje durumu (FK → proje_durum)
        asama_id INTEGER REFERENCES proje_asama(asama_id) ON DELETE
    SET NULL,
        -- Onay akışı aşaması (FK → proje_asama); NULL ise aktif iş akışı yok
        bap_turu_id INTEGER REFERENCES proje_bap_turu(bap_turu_id) ON DELETE
    SET NULL,
        -- BAP türü (FK)
        olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- Durum bazlı hızlı arama indeksi
CREATE INDEX IF NOT EXISTS idx_proje_durum_id ON proje(durum_id);
-- Aşama bazlı hızlı arama indeksi
CREATE INDEX IF NOT EXISTS idx_proje_asama_id ON proje(asama_id);
-- Koordinatör bazlı hızlı arama indeksi
CREATE INDEX IF NOT EXISTS idx_proje_koordinator_id ON proje(koordinator_id);
-- ==========================================
-- Migration: 005_create_proje_detay_table.sql
-- ==========================================
-- ====================================================
-- Proje Detay Tablosu
-- Projenin akademik detay alanlarını tutar
-- (Özet, anahtar kelimeler, hedefler, özgünlük, metodoloji)
-- ====================================================
CREATE TABLE IF NOT EXISTS proje_detay (
    proje_id INTEGER PRIMARY KEY REFERENCES proje(proje_id) ON DELETE CASCADE,
    -- Proje referansı (1'e 1 ilişki)
    ozet VARCHAR(500),
    -- Proje özeti (abstract)
    anahtar_kelimeler VARCHAR(100),
    -- Anahtar kelimeler (keywords)
    hedefler VARCHAR(500),
    -- Proje hedefleri (objectives)
    ozgunluk VARCHAR(100),
    -- Projenin özgünlüğü (originality)
    metodoloji VARCHAR(500) -- Kullanılacak metodoloji (methodology)
);
-- ==========================================
-- Migration: 006_create_proje_takim_table.sql
-- ==========================================
-- ====================================================
-- Proje Takım Tablosu
-- Proje ile üye arasındaki ilişkiyi (çoka çok) tutar
-- Her üyenin projede hangi rolde olduğunu belirtir
-- ====================================================
CREATE TABLE IF NOT EXISTS proje_takim (
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    -- Proje referansı
    uye_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    -- Üye referansı
    proje_rol_id INTEGER REFERENCES proje_rol_tanimlama(rol_id) ON DELETE
    SET NULL,
        -- Proje rolü referansı
        PRIMARY KEY (proje_id, uye_id)
);
-- Üye bazlı hızlı arama indeksi
CREATE INDEX IF NOT EXISTS idx_proje_takim_uye_id ON proje_takim(uye_id);
-- ==========================================
-- Migration: 007_create_butce_table.sql
-- ==========================================
-- ====================================================
-- Bütçe Tablosu
-- Proje bütçe kalemlerini tutar
-- ====================================================
CREATE TABLE IF NOT EXISTS butce (
    kalem_id SERIAL PRIMARY KEY,
    -- Bütçe kalemi ID
    proje_id INTEGER REFERENCES proje(proje_id) ON DELETE CASCADE,
    -- İlgili proje
    kategori_id INTEGER REFERENCES butce_kategori(kategori_id) ON DELETE
    SET NULL,
        -- Bütçe kategorisi
        aciklama VARCHAR(500),
        -- Kalem açıklaması
        birim_ozelligi INTEGER,
        -- Birim özelliği/spec
        birim_fiyat NUMERIC(10, 2) DEFAULT 0,
        -- Birim fiyat
        toplam_fiyat NUMERIC(12, 2) DEFAULT 0,
        -- Toplam fiyat
        olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- Proje bazlı bütçe arama indeksi
CREATE INDEX IF NOT EXISTS idx_butce_proje_id ON butce(proje_id);
-- ==========================================
-- Migration: 008_create_is_paketi_table.sql
-- ==========================================
-- ====================================================
-- İş Paketi Tablosu
-- Projelerin iş paketlerini (görev dağılımı) tutar
-- ====================================================
CREATE TABLE IF NOT EXISTS is_paketi (
    paket_id SERIAL PRIMARY KEY,
    proje_id INTEGER REFERENCES proje(proje_id) ON DELETE CASCADE,
    -- İlgili proje
    paket_adi VARCHAR(500),
    -- İş paketinin adı
    paket_amaci VARCHAR(500),
    -- İş paketinin amacı
    baslangic_tarihi DATE,
    -- Başlangıç tarihi
    bitis_tarihi DATE,
    -- Bitiş tarihi
    olusturma_tarihi DATE DEFAULT CURRENT_DATE,
    -- Oluşturulma tarihi
    guncelleme_tarihi DATE DEFAULT CURRENT_DATE -- Güncelleme tarihi
);
-- Proje bazlı iş paketi arama indeksi
CREATE INDEX IF NOT EXISTS idx_is_paketi_proje_id ON is_paketi(proje_id);
-- ==========================================
-- Migration: 009_create_risk_yonetimi_table.sql
-- ==========================================
-- ====================================================
-- Risk Yönetimi Tablosu
-- Proje ve iş paketlerine bağlı riskleri ve çözüm planlarını tutar
-- ====================================================
CREATE TABLE IF NOT EXISTS risk_yonetimi (
    risk_id SERIAL PRIMARY KEY,
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    -- Proje referansı
    paket_id INTEGER REFERENCES is_paketi(paket_id) ON DELETE
    SET NULL,
        -- İş paketi referansı (opsiyonel)
        risk_aciklamasi VARCHAR(500),
        -- Risk açıklaması
        cozum_plani VARCHAR(500) -- Çözüm planı
);
-- Proje bazlı risk arama indeksi
CREATE INDEX IF NOT EXISTS idx_risk_yonetimi_proje_id ON risk_yonetimi(proje_id);
-- ==========================================
-- Migration: 010_create_proje_yayinlastirma_table.sql
-- ==========================================
-- ====================================================
-- Proje Yayınlaştırma Tablosu
-- Proje yayınlaştırma bilgilerini tutar
-- ====================================================
CREATE TABLE IF NOT EXISTS proje_yayinlastirma (
    proje_id INTEGER PRIMARY KEY REFERENCES proje(proje_id) ON DELETE CASCADE,
    -- Proje referansı (1'e 1)
    yayin_turu VARCHAR(200),
    -- Yayın türü
    yayin_ciktisi VARCHAR(200),
    -- Çıktı bilgisi
    tahmini_yayin_tarihi DATE -- Tahmini yayın tarihi
);
-- ==========================================
-- Migration: 011_create_proje_cikti_table.sql
-- ==========================================
-- ====================================================
-- Proje Çıktı Tablosu
-- Projenin beklenen çıktılarını tutar
-- ====================================================
CREATE TABLE IF NOT EXISTS proje_cikti (
    cikti_id SERIAL PRIMARY KEY,
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    -- Proje referansı
    cikti_turu_id INTEGER REFERENCES proje_cikti_turu(cikti_turu_id) ON DELETE
    SET NULL,
        -- Çıktı türü referansı
        aciklama VARCHAR(500),
        -- Çıktı açıklaması
        cikti_periyodu DATE -- Çıktı periyodu/tarihi
);
-- Proje bazlı çıktı arama indeksi
CREATE INDEX IF NOT EXISTS idx_proje_cikti_proje_id ON proje_cikti(proje_id);
-- ==========================================
-- Migration: 012_create_arastirma_table.sql
-- ==========================================
-- ====================================================
-- Araştırma Tablosu
-- Projeye bağlı araştırma bilgilerini tutar
-- ====================================================
CREATE TABLE IF NOT EXISTS arastirma (
    proje_id INTEGER PRIMARY KEY REFERENCES proje(proje_id) ON DELETE CASCADE,
    -- Proje referansı (1'e 1)
    olusturan_id INTEGER REFERENCES uye(uye_id) ON DELETE
    SET NULL,
        -- Oluşturan kişi
        arastirma_amaci VARCHAR(500) -- Araştırma amacı
);
-- ==========================================
-- Migration: 013_create_proje_degerlendirmeleri_table.sql
-- ==========================================
-- ====================================================
-- Proje Değerlendirmeleri Tablosu
-- Hakemlerin projeleri değerlendirmelerini tutar
-- (ER diyagramında yok, mevcut sistem için korunmuştur)
-- ====================================================
CREATE TABLE IF NOT EXISTS proje_degerlendirmeleri (
    degerlendirme_id SERIAL PRIMARY KEY,
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    -- Değerlendirilen proje
    hakem_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    -- Değerlendiren hakem
    puan INTEGER CHECK (
        puan >= 0
        AND puan <= 100
    ),
    -- 0-100 arası genel puan
    yorum TEXT,
    -- Hakemin genel proje yorumu
    durum VARCHAR(50) DEFAULT 'Bekliyor',
    -- Bekliyor, Onaylandı, Reddedildi, Revizyon
    taahhutname_onay_tarihi TIMESTAMP WITH TIME ZONE,
    -- Hakemin gizlilik taahhütnamesini onayladığı tarih (kabul için zorunlu)
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(proje_id, hakem_id) -- Bir hakem bir projeyi bir kez değerlendirebilir
);
-- ==========================================
-- Migration: 014_create_revizyonlar_table.sql
-- ==========================================
-- ====================================================
-- Revizyonlar Tablosu
-- Proje revizyonlarını ve durumlarını takip eder
-- (ER diyagramında yok, mevcut sistem için korunmuştur)
-- ====================================================
CREATE TABLE IF NOT EXISTS revizyonlar (
    revizyon_id SERIAL PRIMARY KEY,
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    -- İlgili proje
    olusturan_kisi_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    -- Revizyonu oluşturan kişi
    atanan_kisi_id INTEGER REFERENCES uye(uye_id) ON DELETE
    SET NULL,
        -- Revizyonun atandığı kişi
        aciklama TEXT NOT NULL,
        -- Revizyon açıklaması
        durum VARCHAR(50) DEFAULT 'Bekliyor',
        -- Bekliyor, Tamamlandı
        revizyon_bolum VARCHAR(100),
        -- Revizyon istenen bölüm
        olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- ==========================================
-- Migration: 015_insert_default_data.sql
-- ==========================================
-- ====================================================
-- Varsayılan Veriler
-- Sistem başlangıç verileri (admin kullanıcı vb.)
-- ====================================================
-- Varsayılan Admin Kullanıcısı
INSERT INTO uye (
        rol,
        ad,
        soyad,
        unvan,
        bolum,
        eposta,
        telefon,
        izu_uyesi,
        sifre_hash,
        aktif_mi
    )
VALUES (
        'admin',
        'Sistem',
        'Yöneticisi',
        'Prof. Dr.',
        'Bilgi İşlem',
        'admin@izu.edu.tr',
        '5555555555',
        true,
        '$2a$10$uUSxvVDTYDu4KjZXbnPx3OOJVvppVRYJcm4Dlhs0Mx8xRcmcD46ri',
        true
    ) ON CONFLICT (eposta) DO NOTHING;
-- Admin kullanıcısına sistem rolü ata
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT u.uye_id,
    srt.rol_id
FROM uye u,
    sistem_rol_tanimlama srt
WHERE u.eposta = 'admin@izu.edu.tr'
    AND srt.rol_adi = 'admin' ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;
-- Varsayılan Admin1 Kullanıcısı
INSERT INTO uye (
        rol,
        ad,
        soyad,
        unvan,
        bolum,
        eposta,
        telefon,
        izu_uyesi,
        sifre_hash,
        aktif_mi
    )
VALUES (
        'admin',
        'Admin1',
        'Yöneticisi',
        'Prof. Dr.',
        'Bilgi İşlem',
        'admin1@izu.edu.tr',
        '5555555556',
        true,
        '$2a$10$uUSxvVDTYDu4KjZXbnPx3OOJVvppVRYJcm4Dlhs0Mx8xRcmcD46ri',
        true
    ) ON CONFLICT (eposta) DO NOTHING;
-- Admin1 kullanıcısına sistem rolü ata
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT u.uye_id,
    srt.rol_id
FROM uye u,
    sistem_rol_tanimlama srt
WHERE u.eposta = 'admin1@izu.edu.tr'
    AND srt.rol_adi = 'admin' ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;
-- Varsayılan Öğrenci Kullanıcısı
INSERT INTO uye (
        rol,
        ad,
        soyad,
        unvan,
        bolum,
        eposta,
        telefon,
        izu_uyesi,
        sifre_hash,
        aktif_mi
    )
VALUES (
        'ogrenci',
        'Öğrenci',
        'Kullanıcısı',
        NULL,
        'Bilgisayar Mühendisliği',
        'ogrenci@izu.edu.tr',
        '5555555557',
        true,
        '$2a$10$hWT4OaAqjGOvMW/aFMfzoOae5O53wSjgffRVHmoN/75Y3m5DRajY2',
        true
    ) ON CONFLICT (eposta) DO NOTHING;
-- Öğrenci kullanıcısına sistem rolü ata
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT u.uye_id,
    srt.rol_id
FROM uye u,
    sistem_rol_tanimlama srt
WHERE u.eposta = 'ogrenci@izu.edu.tr'
    AND srt.rol_adi = 'ogrenci' ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;
-- Varsayılan Akademisyen Kullanıcısı
INSERT INTO uye (
        rol,
        ad,
        soyad,
        unvan,
        bolum,
        eposta,
        telefon,
        izu_uyesi,
        sifre_hash,
        aktif_mi
    )
VALUES (
        'akademisyen',
        'Akademisyen',
        'Kullanıcısı',
        'Prof. Dr.',
        'Bilgisayar Mühendisliği',
        'akademisyen@izu.edu.tr',
        '5555555558',
        true,
        '$2a$10$kHe1CybkltpFDSnd7PvEleMkyNcisc0C9s.Poz8v94lt0POdHfUt6',
        true
    ) ON CONFLICT (eposta) DO NOTHING;
-- Akademisyen kullanıcısına sistem rolü ata
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT u.uye_id,
    srt.rol_id
FROM uye u,
    sistem_rol_tanimlama srt
WHERE u.eposta = 'akademisyen@izu.edu.tr'
    AND srt.rol_adi = 'akademisyen' ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;
-- Varsayılan Dekan Kullanıcısı
INSERT INTO uye (
        rol,
        ad,
        soyad,
        unvan,
        bolum,
        eposta,
        telefon,
        izu_uyesi,
        sifre_hash,
        aktif_mi
    )
VALUES (
        'dekan',
        'Dekan',
        'Kullanıcısı',
        'Prof. Dr.',
        'Bilgisayar Mühendisliği',
        'dekan@izu.edu.tr',
        '5555555559',
        true,
        '$2a$10$VUSdpqx1BFQqGA/MtcP9me7hK1PrQFScEiJocDzFhPpYCX24flH5u',
        true
    ) ON CONFLICT (eposta) DO NOTHING;
-- Dekan kullanıcısına sistem rolü ata
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT u.uye_id,
    srt.rol_id
FROM uye u,
    sistem_rol_tanimlama srt
WHERE u.eposta = 'dekan@izu.edu.tr'
    AND srt.rol_adi = 'dekan' ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;
-- Varsayılan TTO Kullanıcısı
INSERT INTO uye (
        rol,
        ad,
        soyad,
        unvan,
        bolum,
        eposta,
        telefon,
        izu_uyesi,
        sifre_hash,
        aktif_mi
    )
VALUES (
        'tto',
        'TTO',
        'Uzmanı',
        'Dr.',
        'Teknoloji Transfer Ofisi',
        'tto@izu.edu.tr',
        '5555555553',
        true,
        '$2a$10$rj4nxdqm9EN.wDQM/H0ZkOJquMceS41lk1INHgnBOg5LH1Xc/zfai',
        true
    ) ON CONFLICT (eposta) DO NOTHING;
-- TTO kullanıcısına sistem rolünü ata
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT u.uye_id,
    srt.rol_id
FROM uye u,
    sistem_rol_tanimlama srt
WHERE u.eposta = 'tto@izu.edu.tr'
    AND srt.rol_adi = 'tto' ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;
-- ==========================================
-- Migration: 016_add_pdf_dosya_yolu.sql
-- ==========================================
-- ================================================================
-- Migration 016: Proje tablosuna PDF dosya yolu sütunu eklenmesi
-- Bu sütun, nihai onaylanan başvuru PDF'inin sunucu yolunu tutar.
-- ================================================================
ALTER TABLE proje
ADD COLUMN IF NOT EXISTS pdf_dosya_yolu TEXT;
-- Yorum: Bu alan nullable'dır çünkü PDF ancak başvuru
-- onaylandıktan sonra üretilir ve kaydedilir.
-- ==========================================
-- Migration: 017_update_bap_turleri.sql
-- ==========================================
-- ================================================================
-- Migration 017: BAP proje türlerini BAP-100..500 olarak güncelle
-- Mevcut kayıtları yeni isimlendirmeye uyumlu hale getirir.
-- ================================================================
-- Mevcut eski isimleri güncelle (varsa)
UPDATE proje_bap_turu
SET bap_turu = 'BAP-100'
WHERE bap_turu = 'Lisans Tez Projesi';
UPDATE proje_bap_turu
SET bap_turu = 'BAP-200'
WHERE bap_turu = 'Yüksek Lisans Tez Projesi';
UPDATE proje_bap_turu
SET bap_turu = 'BAP-300'
WHERE bap_turu = 'Doktora Tez Projesi';
UPDATE proje_bap_turu
SET bap_turu = 'BAP-400'
WHERE bap_turu = 'Akademisyen Araştırma Projesi';
UPDATE proje_bap_turu
SET bap_turu = 'BAP-500'
WHERE bap_turu = 'Bilimsel Etkinlik Destek Projesi';
-- Eğer hiç kayıt yoksa yeni ekle
INSERT INTO proje_bap_turu (bap_turu)
VALUES ('BAP-100'),
    ('BAP-200'),
    ('BAP-300'),
    ('BAP-400'),
    ('BAP-500') ON CONFLICT (bap_turu) DO NOTHING;
-- ==========================================
-- Migration: 018_add_davet_durumu.sql
-- ==========================================
-- ================================================================
-- Migration 018: Proje takım tablosuna davet durumu sütunu ekleme
-- Ekip üyeleri davet edildiğinde beklemede, kabul veya red durumuna geçer
-- ================================================================
-- Davet durumu sütunu eklenir (varsayılan: beklemede)
ALTER TABLE proje_takim
ADD COLUMN IF NOT EXISTS davet_durumu VARCHAR(20) DEFAULT 'beklemede';
-- Mevcut kayıtları otomatik kabul olarak işaretle (geriye uyumluluk)
UPDATE proje_takim
SET davet_durumu = 'kabul'
WHERE davet_durumu IS NULL
    OR davet_durumu = 'beklemede';
-- ==========================================
-- Migration: 019_create_uye_detay_table.sql
-- ==========================================
-- ====================================================
-- Üye Detay Tablosu
-- Kullanıcıların giriş sonrası tamamlayacağı profil bilgilerini tutar
-- uye tablosu ile 1:1 ilişki (uye_id üzerinden)
-- ====================================================
CREATE TABLE IF NOT EXISTS uye_detay (
    detay_id SERIAL PRIMARY KEY,
    uye_id INTEGER UNIQUE NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    -- Üye ile 1:1 ilişki
    rol VARCHAR(100),
    -- Kullanıcı rolü (akademisyen, ogrenci, hakem, admin)
    unvan VARCHAR(100),
    -- Akademik unvan (Prof. Dr., Doç. Dr. vb.)
    bolum VARCHAR(255),
    -- Bölüm / Fakülte bilgisi
    telefon VARCHAR(15),
    -- İletişim telefonu
    izu_uyesi BOOLEAN DEFAULT FALSE,
    -- İZÜ üyesi mi?
    profil_tamamlandi BOOLEAN DEFAULT FALSE,
    -- Profil bilgileri tamamlandı mı?
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- Üye ID'si için hızlı arama indeksi
CREATE INDEX IF NOT EXISTS idx_uye_detay_uye_id ON uye_detay(uye_id);
-- Rol alanı için filtreleme indeksi
CREATE INDEX IF NOT EXISTS idx_uye_detay_rol ON uye_detay(rol);
-- ====================================================
-- Mevcut Veri Taşıma (Migration)
-- uye tablosundaki detay alanlarını uye_detay tablosuna taşır
-- ====================================================
INSERT INTO uye_detay (
        uye_id,
        rol,
        unvan,
        bolum,
        telefon,
        izu_uyesi,
        profil_tamamlandi
    )
SELECT uye_id,
    COALESCE(rol, 'ogrenci'),
    unvan,
    bolum,
    telefon,
    izu_uyesi,
    TRUE -- Mevcut kullanıcıların profili zaten tamamlanmış kabul edilir
FROM uye
WHERE NOT EXISTS (
        SELECT 1
        FROM uye_detay
        WHERE uye_detay.uye_id = uye.uye_id
    );
-- ==========================================
-- Migration: 019_make_bap_turu_dynamic.sql
-- ==========================================
-- ================================================================
-- Migration 019: BAP Proje Türleri Tablosunun Dinamik Hale Getirilmesi
-- Admin tarafından tanımlanabilecek yeni sütunların eklenmesi
-- ================================================================
-- Sütunlar eklenir (varsayılan değerlerle)
ALTER TABLE proje_bap_turu
ADD COLUMN IF NOT EXISTS butce_limiti NUMERIC(12, 2) DEFAULT 0;
ALTER TABLE proje_bap_turu
ADD COLUMN IF NOT EXISTS sure_limiti_ay INTEGER DEFAULT 0;
ALTER TABLE proje_bap_turu
ADD COLUMN IF NOT EXISTS aktif_mi BOOLEAN DEFAULT TRUE;
ALTER TABLE proje_bap_turu
ADD COLUMN IF NOT EXISTS aciklama TEXT DEFAULT '';
ALTER TABLE proje_bap_turu
ADD COLUMN IF NOT EXISTS hakem_gerekli BOOLEAN DEFAULT FALSE;
ALTER TABLE proje_bap_turu
ADD COLUMN IF NOT EXISTS hakem_sayisi INTEGER DEFAULT 0;
ALTER TABLE proje_bap_turu
ADD COLUMN IF NOT EXISTS bursiyer_gerekli BOOLEAN DEFAULT FALSE;
ALTER TABLE proje_bap_turu
ADD COLUMN IF NOT EXISTS bursiyer_sayisi INTEGER DEFAULT 0;
-- Mevcut varsayılan BAP türlerini gerçekçi değerlerle güncelle
UPDATE proje_bap_turu
SET butce_limiti = 50000.00,
    sure_limiti_ay = 12,
    aciklama = 'Lisans Tez Projesi Desteği'
WHERE bap_turu = 'BAP-100';
UPDATE proje_bap_turu
SET butce_limiti = 100000.00,
    sure_limiti_ay = 24,
    aciklama = 'Yüksek Lisans Tez Projesi Desteği'
WHERE bap_turu = 'BAP-200';
UPDATE proje_bap_turu
SET butce_limiti = 150000.00,
    sure_limiti_ay = 36,
    aciklama = 'Doktora Tez Projesi Desteği'
WHERE bap_turu = 'BAP-300';
UPDATE proje_bap_turu
SET butce_limiti = 30000.00,
    sure_limiti_ay = 6,
    aciklama = 'Akademisyen Araştırma Projesi Desteği'
WHERE bap_turu = 'BAP-400';
UPDATE proje_bap_turu
SET butce_limiti = 250000.00,
    sure_limiti_ay = 36,
    aciklama = 'Bilimsel Etkinlik Destek Projesi'
WHERE bap_turu = 'BAP-500';
-- ==========================================
-- Migration: 020_sync_sistem_rol.sql
-- ==========================================
-- ================================================================
-- Migration 020: Mevcut kullanıcıların sistem_rol tablosuna senkronizasyonu
-- uye_detay tablosundaki rol bilgisine göre sistem_rol ilişki tablosu doldurulur.
-- Bu migration sadece bir kez çalıştırılmalıdır.
-- ================================================================
-- Mevcut tüm kullanıcıların rollerini sistem_rol tablosuna ekle
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT d.uye_id,
    srt.rol_id
FROM uye_detay d
    INNER JOIN sistem_rol_tanimlama srt ON srt.rol_adi = d.rol
WHERE d.rol IS NOT NULL
    AND d.rol != '' ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;
-- ==========================================
-- Migration: 021_hakem_atama_akisi.sql
-- ==========================================
-- ================================================================
-- Migration 021: Hakem Atama Kabul/Red Akışı
-- Hakemlerin atamayı kabul veya reddetme sürecini yönetir.
-- Mevcut 'durum' alanına dokunulmaz (öğrenci görünümü korunur).
-- Yeni 'atama_durumu' alanı admin/akademisyen görünümü için eklenir.
-- ================================================================
-- Hakem atama kabul/red durumu sütunu eklenir
-- Değerler: 'Atandı', 'Kabul Edildi', 'Reddedildi'
ALTER TABLE proje_degerlendirmeleri
ADD COLUMN IF NOT EXISTS atama_durumu VARCHAR(50) DEFAULT 'Atandı';
-- Hakemin atamayı reddetme sebebini tutmak için alan
ALTER TABLE proje_degerlendirmeleri
ADD COLUMN IF NOT EXISTS red_nedeni TEXT;
-- ==========================================
-- Migration: 020_sync_sistem_rol.sql
-- ==========================================
-- ================================================================
-- Migration 020: Mevcut kullanıcıların sistem_rol tablosuna senkronizasyonu
-- uye_detay tablosundaki rol bilgisine göre sistem_rol ilişki tablosu doldurulur.
-- Bu migration sadece bir kez çalıştırılmalıdır.
-- ================================================================
-- Mevcut tüm kullanıcıların rollerini sistem_rol tablosuna ekle
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT d.uye_id,
    srt.rol_id
FROM uye_detay d
    INNER JOIN sistem_rol_tanimlama srt ON srt.rol_adi = d.rol
WHERE d.rol IS NOT NULL
    AND d.rol != '' ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;
-- ==========================================
-- Migration: 021_hakem_atama_akisi.sql
-- ==========================================
-- ================================================================
-- Migration 021: Hakem Atama Kabul/Red Akışı
-- Hakemlerin atamayı kabul veya reddetme sürecini yönetir.
-- Mevcut 'durum' alanına dokunulmaz (öğrenci görünümü korunur).
-- Yeni 'atama_durumu' alanı admin/akademisyen görünümü için eklenir.
-- ================================================================
-- Hakem atama kabul/red durumu sütunu eklenir
-- Değerler: 'Atandı', 'Kabul Edildi', 'Reddedildi'
ALTER TABLE proje_degerlendirmeleri
ADD COLUMN IF NOT EXISTS atama_durumu VARCHAR(50) DEFAULT 'Atandı';
-- Hakemin atamayı reddetme sebebini tutmak için alan
ALTER TABLE proje_degerlendirmeleri
ADD COLUMN IF NOT EXISTS red_nedeni TEXT;
-- Hakemin atamayı kabul/red ettiği tarih
ALTER TABLE proje_degerlendirmeleri
ADD COLUMN IF NOT EXISTS karar_tarihi TIMESTAMP WITH TIME ZONE;
-- Hakemin gizlilik taahhütnamesini onayladığı tarih (kabul için zorunlu)
ALTER TABLE proje_degerlendirmeleri
ADD COLUMN IF NOT EXISTS taahhutname_onay_tarihi TIMESTAMP WITH TIME ZONE;
-- Mevcut kayıtları geriye dönük uyumluluk için 'Kabul Edildi' olarak işaretle
UPDATE proje_degerlendirmeleri
SET atama_durumu = 'Kabul Edildi'
WHERE atama_durumu IS NULL
    OR atama_durumu = 'Atandı';
-- ================================================================
-- Proje Süreç Geçmişi Tablosu
-- Onay, Red, Revizyon logları ve tarihçesi
-- ================================================================
CREATE TABLE IF NOT EXISTS proje_surec_gecmisi (
    gecmis_id SERIAL PRIMARY KEY,
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    islem_yapan_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    baslangic_durum VARCHAR(100),
    hedef_durum VARCHAR(100),
    aciklama TEXT,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_proje_surec_gecmisi_proje_id ON proje_surec_gecmisi(proje_id);
-- ==========================================
-- Migration: 022_kaynakca_ve_yayin_etki.sql
-- ==========================================
-- ================================================================
-- Migration 022: Kaynakça alanı, Yaygın Etki ve Yaygınlaştırma tabloları
-- Proje başvuru formunun 6. bölümü için gerekli tablolar eklendi.
-- ================================================================
-- proje_detay tablosuna kaynakça alanı ekleniyor
ALTER TABLE proje_detay
ADD COLUMN IF NOT EXISTS kaynakca TEXT;
-- Projeden Elde Edilmesi Öngörülen Çıktılar tablosu (sabit satır türleri)
CREATE TABLE IF NOT EXISTS proje_yayin_etki (
    id SERIAL PRIMARY KEY,
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    cikti_turu VARCHAR(100) NOT NULL,
    -- bilimsel_akademik, ekonomik_ticari_sosyal, arastirmaci_yetistirme, olusturulmasina_yonelik
    ongorul_cikti TEXT,
    -- Öngörülen çıktılar
    zaman_araligi VARCHAR(200) -- Elde edilme zaman aralığı
);
CREATE INDEX IF NOT EXISTS idx_proje_yayin_etki_proje_id ON proje_yayin_etki(proje_id);
-- Çıktıların Paylaşımı ve Yaygınlaştırılması tablosu (dinamik satırlar)
CREATE TABLE IF NOT EXISTS proje_yayginlastirma_etkinlik (
    id SERIAL PRIMARY KEY,
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    etkinlik_turu VARCHAR(500),
    -- Etkinlik türü
    paydas VARCHAR(500),
    -- Paydaş / Olası Kullanıcılar
    zaman_sure VARCHAR(200),
    -- Etkinliğin Zamanı ve Süresi
    sira_no INTEGER DEFAULT 1 -- Sıra numarası
);
CREATE INDEX IF NOT EXISTS idx_proje_yayginlastirma_etkinlik_proje_id ON proje_yayginlastirma_etkinlik(proje_id);
-- ====================================================
-- E-İmza (Elektronik İmza) Kayıtları Tablosu
-- ====================================================
CREATE TABLE IF NOT EXISTS proje_imza (
    imza_id SERIAL PRIMARY KEY,
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    uye_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    imzaci_ad_soyad VARCHAR(255) NOT NULL,
    rol VARCHAR(100) NOT NULL,
    imza_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    imza_durumu VARCHAR(50) DEFAULT 'İmzalandı',
    imza_token VARCHAR(255),
    ip_adresi VARCHAR(50),
    tarayici_bilgisi VARCHAR(255),
    UNIQUE(proje_id, uye_id)
);
CREATE INDEX IF NOT EXISTS idx_proje_imza_proje_id ON proje_imza(proje_id);
CREATE INDEX IF NOT EXISTS idx_proje_imza_uye_id ON proje_imza(uye_id);
-- ================================================================
-- Sistem Sayfa Yetkilendirme Tabloları
-- Dinamik sayfa-rol yetkilendirmesi için gerekli tablolar
-- ================================================================
-- Sistem Sayfaları Tanımlama Tablosu
CREATE TABLE IF NOT EXISTS sistem_sayfa (
    sayfa_id SERIAL PRIMARY KEY,
    sayfa_adi VARCHAR(200) NOT NULL,
    sayfa_kodu VARCHAR(100) UNIQUE NOT NULL,
    -- Örn: basvuru, hakem_dashboard vb.
    url_yolu VARCHAR(255) UNIQUE NOT NULL -- Sayfa URL'si
);
-- Sayfa Rol Yetkileri Eşleştirme Tablosu
CREATE TABLE IF NOT EXISTS sayfa_rol_yetki (
    yetki_id SERIAL PRIMARY KEY,
    sistem_rol_id INTEGER NOT NULL REFERENCES sistem_rol_tanimlama(rol_id) ON DELETE CASCADE,
    sayfa_id INTEGER NOT NULL REFERENCES sistem_sayfa(sayfa_id) ON DELETE CASCADE,
    UNIQUE(sistem_rol_id, sayfa_id)
);
-- Sayfalar için varsayılan indeksler
CREATE INDEX IF NOT EXISTS idx_sayfa_rol_yetki_rol ON sayfa_rol_yetki(sistem_rol_id);
CREATE INDEX IF NOT EXISTS idx_sayfa_rol_yetki_sayfa ON sayfa_rol_yetki(sayfa_id);
-- ====================================================
-- Varsayılan Sayfa Tanımları (Seed Verisi)
-- ====================================================
INSERT INTO sistem_sayfa (sayfa_adi, sayfa_kodu, url_yolu)
VALUES ('Anasayfa', 'anasayfa', '/anasayfa'),
    ('Yeni Başvuru Formu', 'basvuru', '/basvuru'),
    ('Profil Sayfası', 'profil', '/profil'),
    (
        'Hakem Dashboard',
        'hakem_dashboard',
        '/hakem/dashboard'
    ),
    (
        'Proje Değerlendirme',
        'hakem_degerlendirme',
        '/hakem/degerlendirme'
    ),
    (
        'Admin Dashboard',
        'admin_dashboard',
        '/admin/dashboard'
    ),
    (
        'Hakem Atama',
        'admin_hakem_atama',
        '/admin/hakem-atama'
    ),
    (
        'Proje Durum Raporları',
        'admin_projects_status',
        '/admin/projects/status'
    ),
    (
        'Dekan Dashboard',
        'dekan_dashboard',
        '/dekan/dashboard'
    ),
    (
        'Komisyon Dashboard',
        'komisyon_dashboard',
        '/komisyon/dashboard'
    ),
    (
        'TTO Dashboard',
        'tto_dashboard',
        '/tto/dashboard'
    ),
    ('E-İmza Paneli', 'eimza', '/eimza'),
    ('Projelerim', 'projelerim', '/projelerim'),
    (
        'Yapay Zeka Asistanı (Chatbot)',
        'chatbot',
        '/api/chat'
    ),
    (
        'Satın Alma Talepleri (Araştırmacı)',
        'satinalma_arastirmaci',
        '/satinalma'
    ),
    (
        'Satın Alma Yönetimi (TTO)',
        'satinalma_tto',
        '/tto/satinalma'
    ) ON CONFLICT (sayfa_kodu) DO NOTHING;
-- ====================================================
-- Varsayılan Sayfa Yetkileri (Seed Verisi)
-- ====================================================
-- Admin yetkileri (Tüm sayfalara erişebilir)
INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
SELECT r.rol_id,
    s.sayfa_id
FROM sistem_rol_tanimlama r,
    sistem_sayfa s
WHERE r.rol_adi = 'admin' ON CONFLICT DO NOTHING;
-- Akademisyen yetkileri (Anasayfa, Başvuru, Profil ve Projelerim görebilir)
INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
SELECT r.rol_id,
    s.sayfa_id
FROM sistem_rol_tanimlama r,
    sistem_sayfa s
WHERE r.rol_adi = 'akademisyen'
    AND s.sayfa_kodu IN ('anasayfa', 'basvuru', 'profil', 'projelerim') ON CONFLICT DO NOTHING;
-- Öğrenci yetkileri (Anasayfa, Başvuru, Profil ve Projelerim görebilir)
INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
SELECT r.rol_id,
    s.sayfa_id
FROM sistem_rol_tanimlama r,
    sistem_sayfa s
WHERE r.rol_adi = 'ogrenci'
    AND s.sayfa_kodu IN ('anasayfa', 'basvuru', 'profil', 'projelerim') ON CONFLICT DO NOTHING;
-- Hakem yetkileri (Anasayfa, Profil, Hakem Dashboard ve Değerlendirme görebilir)
INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
SELECT r.rol_id,
    s.sayfa_id
FROM sistem_rol_tanimlama r,
    sistem_sayfa s
WHERE r.rol_adi = 'hakem'
    AND s.sayfa_kodu IN (
        'anasayfa',
        'profil',
        'hakem_dashboard',
        'hakem_degerlendirme'
    ) ON CONFLICT DO NOTHING;
-- Dekan yetkileri (Anasayfa, Profil ve Dekan Dashboard görebilir)
INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
SELECT r.rol_id,
    s.sayfa_id
FROM sistem_rol_tanimlama r,
    sistem_sayfa s
WHERE r.rol_adi = 'dekan'
    AND s.sayfa_kodu IN ('anasayfa', 'profil', 'dekan_dashboard') ON CONFLICT DO NOTHING;
-- Komisyon yetkileri (Anasayfa, Profil ve Komisyon Dashboard görebilir)
INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
SELECT r.rol_id,
    s.sayfa_id
FROM sistem_rol_tanimlama r,
    sistem_sayfa s
WHERE r.rol_adi = 'komisyon'
    AND s.sayfa_kodu IN ('anasayfa', 'profil', 'komisyon_dashboard') ON CONFLICT DO NOTHING;
-- TTO yetkileri (Anasayfa, Profil ve TTO Dashboard görebilir)
INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
SELECT r.rol_id,
    s.sayfa_id
FROM sistem_rol_tanimlama r,
    sistem_sayfa s
WHERE r.rol_adi = 'tto'
    AND s.sayfa_kodu IN (
        'anasayfa',
        'profil',
        'tto_dashboard',
        'satinalma_tto'
    ) ON CONFLICT DO NOTHING;
-- Araştırmacı (akademisyen/öğrenci) satın alma yetkileri
INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
SELECT r.rol_id,
    s.sayfa_id
FROM sistem_rol_tanimlama r,
    sistem_sayfa s
WHERE s.sayfa_kodu = 'satinalma_arastirmaci'
    AND r.rol_adi IN ('akademisyen', 'ogrenci') ON CONFLICT DO NOTHING;
-- E-İmza Paneli yetkileri (Sadece admin rolüne eimza sayfası yetkisi verilir)
INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
SELECT r.rol_id,
    s.sayfa_id
FROM sistem_rol_tanimlama r,
    sistem_sayfa s
WHERE s.sayfa_kodu = 'eimza'
    AND r.rol_adi = 'admin' ON CONFLICT DO NOTHING;
-- Chatbot yetkileri (Sadece admin rolüne chatbot sayfası yetkisi verilir)
INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
SELECT r.rol_id,
    s.sayfa_id
FROM sistem_rol_tanimlama r,
    sistem_sayfa s
WHERE s.sayfa_kodu = 'chatbot'
    AND r.rol_adi = 'admin' ON CONFLICT DO NOTHING;
-- ================================================================
-- Migration 023: Form güncellemeleri
-- - proje_detay: TEXT alanlar, İngilizce özet ve anahtar kelimeler
-- - is_paketi: tarih yerine ay tabanlı sütunlar
-- ================================================================
-- proje_detay: metin alanlarını TEXT'e çevir (250 kelime desteği)
ALTER TABLE proje_detay
ALTER COLUMN ozet TYPE TEXT;
ALTER TABLE proje_detay
ALTER COLUMN hedefler TYPE TEXT;
ALTER TABLE proje_detay
ALTER COLUMN ozgunluk TYPE TEXT;
ALTER TABLE proje_detay
ALTER COLUMN metodoloji TYPE TEXT;
-- proje_detay: İngilizce alanlar ekle
ALTER TABLE proje_detay
ADD COLUMN IF NOT EXISTS ozet_en TEXT;
ALTER TABLE proje_detay
ADD COLUMN IF NOT EXISTS anahtar_kelimeler_en VARCHAR(500);
-- is_paketi: tarih tabanlı alanları kaldır, ay tabanlı alanlar ekle
ALTER TABLE is_paketi
ADD COLUMN IF NOT EXISTS baslangic_ay INTEGER;
ALTER TABLE is_paketi
ADD COLUMN IF NOT EXISTS bitis_ay INTEGER;
ALTER TABLE is_paketi DROP COLUMN IF EXISTS baslangic_tarihi;
ALTER TABLE is_paketi DROP COLUMN IF EXISTS bitis_tarihi;
-- ================================================================
-- Migration 024: Satın Alma Talepleri Tablosu
-- Projesi onaylanmış ve aktif olan (sözleşmesi imzalanmış) projelerin
-- bütçe kalemleri üzerinden satın alma talepleri yapılabilmesini sağlar.
-- ================================================================
CREATE TABLE IF NOT EXISTS satinalma_talebi (
    talep_id SERIAL PRIMARY KEY,
    talep_no VARCHAR(100),
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    uye_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE
    SET NULL,
        -- Talebi oluşturan akademisyen
        kalem_id INTEGER NOT NULL REFERENCES butce(kalem_id) ON DELETE CASCADE,
        -- Hangi bütçe kaleminden satın alınacağı
        malzeme_adi VARCHAR(500) NOT NULL,
        -- Malzeme / hizmet adı
        miktar INTEGER NOT NULL CHECK (miktar > 0),
        -- Satın alınacak miktar
        birim_fiyat NUMERIC(10, 2) NOT NULL CHECK (birim_fiyat >= 0),
        -- Birim fiyatı
        toplam_fiyat NUMERIC(12, 2) NOT NULL CHECK (toplam_fiyat >= 0),
        -- Toplam tutar (Go tarafında hesaplanıp yazılır)
        durum VARCHAR(50) DEFAULT 'Beklemede',
        -- 'Beklemede', 'Onaylandı', 'Reddedildi'
        gerekce TEXT NOT NULL,
        -- Gerekçe açıklaması
        red_nedeni TEXT,
        -- Varsa red gerekçesi
        olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_satinalma_talebi_proje_id ON satinalma_talebi(proje_id);
CREATE INDEX IF NOT EXISTS idx_satinalma_talebi_kalem_id ON satinalma_talebi(kalem_id);
-- ================================================================
-- Hakem Havuzu Test Kullanıcıları
-- Şifre: hakem123 (Bcrypt Hash)
-- ================================================================
INSERT INTO uye (
        rol,
        ad,
        soyad,
        unvan,
        bolum,
        eposta,
        izu_uyesi,
        sifre_hash,
        aktif_mi
    )
VALUES (
        'hakem',
        'Hakem 1',
        'Test',
        'Prof. Dr.',
        'Bilgisayar Mühendisliği',
        'hakem1@izu.edu.tr',
        true,
        '$2a$10$77st4J5b6/2ZfuEpPxi12.wPh7YX0gfHNwUHQ/Q2vqhRDp73zd2Fy',
        true
    ),
    (
        'hakem',
        'Hakem 2',
        'Test',
        'Prof. Dr.',
        'Endüstri Mühendisliği',
        'hakem2@izu.edu.tr',
        true,
        '$2a$10$77st4J5b6/2ZfuEpPxi12.wPh7YX0gfHNwUHQ/Q2vqhRDp73zd2Fy',
        true
    ),
    (
        'hakem',
        'Hakem 3',
        'Test',
        'Prof. Dr.',
        'Yazılım Mühendisliği',
        'hakem3@izu.edu.tr',
        true,
        '$2a$10$77st4J5b6/2ZfuEpPxi12.wPh7YX0gfHNwUHQ/Q2vqhRDp73zd2Fy',
        true
    ),
    (
        'hakem',
        'Hakem 4',
        'Test',
        'Doç. Dr.',
        'Bilgisayar Mühendisliği',
        'hakem4@izu.edu.tr',
        true,
        '$2a$10$77st4J5b6/2ZfuEpPxi12.wPh7YX0gfHNwUHQ/Q2vqhRDp73zd2Fy',
        true
    ),
    (
        'hakem',
        'Hakem 5',
        'Test',
        'Doç. Dr.',
        'Gıda Mühendisliği',
        'hakem5@izu.edu.tr',
        true,
        '$2a$10$77st4J5b6/2ZfuEpPxi12.wPh7YX0gfHNwUHQ/Q2vqhRDp73zd2Fy',
        true
    ),
    (
        'hakem',
        'Hakem 6',
        'Test',
        'Prof. Dr.',
        'Elektrik-Elektronik Mühendisliği',
        'hakem6@izu.edu.tr',
        true,
        '$2a$10$77st4J5b6/2ZfuEpPxi12.wPh7YX0gfHNwUHQ/Q2vqhRDp73zd2Fy',
        true
    ) ON CONFLICT (eposta) DO NOTHING;
-- Eklenen hakemlerin sistem_rol tablosuna atamalarının yapılması
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT u.uye_id,
    (
        SELECT rol_id
        FROM sistem_rol_tanimlama
        WHERE rol_adi = 'hakem'
    )
FROM uye u
WHERE u.eposta IN (
        'hakem1@izu.edu.tr',
        'hakem2@izu.edu.tr',
        'hakem3@izu.edu.tr',
        'hakem4@izu.edu.tr',
        'hakem5@izu.edu.tr',
        'hakem6@izu.edu.tr'
    ) ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;
-- ================================================================
-- Komisyon Üyeleri ve Çoklu Onay Tablosu Tanımları
-- ================================================================
-- 5 adet varsayılan komisyon üyesi eklenmesi
INSERT INTO uye (
        rol,
        ad,
        soyad,
        unvan,
        bolum,
        eposta,
        telefon,
        izu_uyesi,
        sifre_hash,
        aktif_mi
    )
VALUES (
        'komisyon',
        'Ahmet',
        'Yılmaz',
        'Prof. Dr.',
        'Bilgisayar Mühendisliği',
        'komisyon1@izu.edu.tr',
        '5555555561',
        true,
        '$2a$10$rj4nxdqm9EN.wDQM/H0ZkOJquMceS41lk1INHgnBOg5LH1Xc/zfai',
        true
    ),
    (
        'komisyon',
        'Mehmet',
        'Kaya',
        'Prof. Dr.',
        'Endüstri Mühendisliği',
        'komisyon2@izu.edu.tr',
        '5555555562',
        true,
        '$2a$10$rj4nxdqm9EN.wDQM/H0ZkOJquMceS41lk1INHgnBOg5LH1Xc/zfai',
        true
    ),
    (
        'komisyon',
        'Ayşe',
        'Demir',
        'Prof. Dr.',
        'İşletme',
        'komisyon3@izu.edu.tr',
        '5555555563',
        true,
        '$2a$10$rj4nxdqm9EN.wDQM/H0ZkOJquMceS41lk1INHgnBOg5LH1Xc/zfai',
        true
    ),
    (
        'komisyon',
        'Fatma',
        'Çelik',
        'Prof. Dr.',
        'Mimarlık',
        'komisyon4@izu.edu.tr',
        '5555555564',
        true,
        '$2a$10$rj4nxdqm9EN.wDQM/H0ZkOJquMceS41lk1INHgnBOg5LH1Xc/zfai',
        true
    ),
    (
        'komisyon',
        'Mustafa',
        'Şahin',
        'Prof. Dr.',
        'Hukuk',
        'komisyon5@izu.edu.tr',
        '5555555565',
        true,
        '$2a$10$rj4nxdqm9EN.wDQM/H0ZkOJquMceS41lk1INHgnBOg5LH1Xc/zfai',
        true
    ) ON CONFLICT (eposta) DO NOTHING;
-- Komisyon üyelerine sistem rolü atama
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT u.uye_id,
    srt.rol_id
FROM uye u,
    sistem_rol_tanimlama srt
WHERE u.eposta IN (
        'komisyon1@izu.edu.tr',
        'komisyon2@izu.edu.tr',
        'komisyon3@izu.edu.tr',
        'komisyon4@izu.edu.tr',
        'komisyon5@izu.edu.tr'
    )
    AND srt.rol_adi = 'komisyon' ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;
-- Komisyon üyelerinin detay profillerini oluşturma
INSERT INTO uye_detay (
        uye_id,
        rol,
        unvan,
        bolum,
        telefon,
        izu_uyesi,
        profil_tamamlandi
    )
SELECT u.uye_id,
    'komisyon',
    u.unvan,
    u.bolum,
    u.telefon,
    true,
    true
FROM uye u
WHERE u.eposta IN (
        'komisyon1@izu.edu.tr',
        'komisyon2@izu.edu.tr',
        'komisyon3@izu.edu.tr',
        'komisyon4@izu.edu.tr',
        'komisyon5@izu.edu.tr'
    )
    AND NOT EXISTS (
        SELECT 1
        FROM uye_detay ud
        WHERE ud.uye_id = u.uye_id
    );
-- Çoklu komisyon onay tablosu
CREATE TABLE IF NOT EXISTS proje_komisyon_onay (
    onay_id SERIAL PRIMARY KEY,
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    komisyon_uye_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    karar VARCHAR(50) DEFAULT 'bekliyor',
    -- bekliyor, onayla, reddet, revizyon
    aciklama TEXT,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(proje_id, komisyon_uye_id)
);
CREATE INDEX IF NOT EXISTS idx_proje_komisyon_onay_proje ON proje_komisyon_onay(proje_id);
CREATE INDEX IF NOT EXISTS idx_proje_komisyon_onay_uye ON proje_komisyon_onay(komisyon_uye_id);

-- ================================================================
-- Migration 025: Hakem Değerlendirme Sorularının Dinamikleştirilmesi
-- ================================================================
CREATE TABLE IF NOT EXISTS hakem_degerlendirme_basliklari (
    baslik_id SERIAL PRIMARY KEY,
    baslik_adi VARCHAR(200) UNIQUE NOT NULL,
    maksimum_puan INTEGER NOT NULL,
    sira_no INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS hakem_degerlendirme_sorulari (
    soru_id SERIAL PRIMARY KEY,
    soru_kodu VARCHAR(10) NOT NULL,
    soru_metni TEXT UNIQUE NOT NULL,
    maksimum_puan INTEGER DEFAULT 5,
    sira_no INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS hakem_degerlendirme_baslik_soru (
    baslik_id INTEGER NOT NULL REFERENCES hakem_degerlendirme_basliklari(baslik_id) ON DELETE CASCADE,
    soru_id INTEGER NOT NULL REFERENCES hakem_degerlendirme_sorulari(soru_id) ON DELETE CASCADE,
    PRIMARY KEY (baslik_id, soru_id)
);

CREATE TABLE IF NOT EXISTS proje_degerlendirme_soru_cevaplari (
    cevap_id SERIAL PRIMARY KEY,
    degerlendirme_id INTEGER NOT NULL REFERENCES proje_degerlendirmeleri(degerlendirme_id) ON DELETE CASCADE,
    baslik_id INTEGER NOT NULL REFERENCES hakem_degerlendirme_basliklari(baslik_id) ON DELETE CASCADE,
    soru_id INTEGER NOT NULL REFERENCES hakem_degerlendirme_sorulari(soru_id) ON DELETE CASCADE,
    puan_degeri VARCHAR(50) NOT NULL,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(degerlendirme_id, baslik_id, soru_id)
);

-- Başlıklar
INSERT INTO hakem_degerlendirme_basliklari (baslik_adi, maksimum_puan, sira_no) VALUES
('Özgün Değer', 24, 1),
('Projenin Yönetimi', 20, 2),
('Projenin Yaygın Etkisi', 20, 3),
('Yapılabilirlik: Ekipman/Ortam', 10, 4),
('Yapılabilirlik: Süre', 16, 5),
('Yapılabilirlik: Bütçe', 10, 6)
ON CONFLICT (baslik_adi) DO UPDATE SET maksimum_puan = EXCLUDED.maksimum_puan, sira_no = EXCLUDED.sira_no;

-- Sorular
INSERT INTO hakem_degerlendirme_sorulari (soru_kodu, soru_metni, maksimum_puan, sira_no) VALUES
('A', 'Yerel, ulusal veya uluslararası bir soruna bilimsel çözüm getirmektedir.', 6, 1),
('B', 'Yöntem, kuram veya ortaya koyacağı bilgi açısından bilimsel ya da teknolojik bir yenilik getirmektedir.', 5, 2),
('C', 'Yeni, farklı bakış sunan ve tamamlayıcı bilimsel bir araştırma sorusu ortaya atmaktadır.', 5, 3),
('D', 'Temel ve güncel bilimsel kaynaklara dayalı literatür taraması ile bilimsel tutarlılığı, bütünlüğü vurgulanmış ve diğer bilimsel çalışmalarla ilişki kurulmuştur.', 4, 4),
('E', 'Araştırmanın amacı (problem/hipotez) açıkça belirtilmiştir.', 4, 5),

('A', 'Araştırmanın amacını (problem/hipotez) test edecek bilimsel araştırma yöntemleri açıkça belirtilmiştir.', 7, 1),
('B', 'Araştırmada proje yönetim araçları kullanılmıştır.', 7, 2),
('C', 'Veri toplama yöntemleri ve araçları (varsa geliştirilme süreçleri) belirtilmiştir.', 6, 3),

('A', 'Bulgular, evrensel ve/veya ulusal düzeyde araştırmacılar tarafından ilgili bilimsel alanda kullanılabilir özelliktedir.', 5, 1),
('B', 'Araştırmacı/Yürütücü elde edilecek bulgularıyla yeni projelere düşünsel kaynak oluşturma ya da ileri bilimsel araştırma üretme potansiyeli vardır.', 5, 2),
('C', 'Desteklenecek projenin lisansüstü tezi üretme veya araştırmacı/öğrenci yetiştirilmesine katkı sağlama potansiyeli vardır.', 5, 3),
('D', 'Yayın, patent, ödül, yarışma derecesi, bildiri ile tescil edilecek çıktılar elde etme potansiyeli vardır.', 5, 4),

('A', 'Projenin yürütüleceği bölümün/merkezin altyapısı, ortamı ve olanakları yeterlidir.', 5, 1),
('B', 'Proje kapsamında istenilen ek ekipman mevcut altyapı ve proje ile uyumludur.', 5, 2),

('A', 'Önerilen araştırma süresi gerçekçidir.', 6, 1),
('B', 'Projede her bir iş paketinin hangi sürede gerçekleştirileceği detaylandırılmıştır.', 5, 2),
('C', 'Projenin başarısını olumsuz yönde etkileyebilecek riskler ve alınacak tedbirler (B Planı) belirtilmiştir.', 5, 3),

('A', 'Önerilen bütçe gerçekçidir ve bütçenin hazırlanmasında ekonomiklik dikkate alınmıştır.', 5, 1),
('B', 'Talep edilen destek iş paketleriyle uyumlu hazırlanmıştır.', 5, 2)
ON CONFLICT (soru_metni) DO UPDATE SET soru_kodu = EXCLUDED.soru_kodu, maksimum_puan = EXCLUDED.maksimum_puan, sira_no = EXCLUDED.sira_no;

-- İlişkilendirmeler (3. Tablo)
-- Özgün Değer (1)
INSERT INTO hakem_degerlendirme_baslik_soru (baslik_id, soru_id)
SELECT b.baslik_id, s.soru_id FROM hakem_degerlendirme_basliklari b, hakem_degerlendirme_sorulari s
WHERE b.baslik_adi = 'Özgün Değer' AND s.soru_metni IN (
    'Yerel, ulusal veya uluslararası bir soruna bilimsel çözüm getirmektedir.',
    'Yöntem, kuram veya ortaya koyacağı bilgi açısından bilimsel ya da teknolojik bir yenilik getirmektedir.',
    'Yeni, farklı bakış sunan ve tamamlayıcı bilimsel bir araştırma sorusu ortaya atmaktadır.',
    'Temel ve güncel bilimsel kaynaklara dayalı literatür taraması ile bilimsel tutarlılığı, bütünlüğü vurgulanmış ve diğer bilimsel çalışmalarla ilişki kurulmuştur.',
    'Araştırmanın amacı (problem/hipotez) açıkça belirtilmiştir.'
) ON CONFLICT DO NOTHING;

-- Projenin Yönetimi (2)
INSERT INTO hakem_degerlendirme_baslik_soru (baslik_id, soru_id)
SELECT b.baslik_id, s.soru_id FROM hakem_degerlendirme_basliklari b, hakem_degerlendirme_sorulari s
WHERE b.baslik_adi = 'Projenin Yönetimi' AND s.soru_metni IN (
    'Araştırmanın amacını (problem/hipotez) test edecek bilimsel araştırma yöntemleri açıkça belirtilmiştir.',
    'Araştırmada proje yönetim araçları kullanılmıştır.',
    'Veri toplama yöntemleri ve araçları (varsa geliştirilme süreçleri) belirtilmiştir.'
) ON CONFLICT DO NOTHING;

-- Projenin Yaygın Etkisi (3)
INSERT INTO hakem_degerlendirme_baslik_soru (baslik_id, soru_id)
SELECT b.baslik_id, s.soru_id FROM hakem_degerlendirme_basliklari b, hakem_degerlendirme_sorulari s
WHERE b.baslik_adi = 'Projenin Yaygın Etkisi' AND s.soru_metni IN (
    'Bulgular, evrensel ve/veya ulusal düzeyde araştırmacılar tarafından ilgili bilimsel alanda kullanılabilir özelliktedir.',
    'Araştırmacı/Yürütücü elde edilecek bulgularıyla yeni projelere düşünsel kaynak oluşturma ya da ileri bilimsel araştırma üretme potansiyeli vardır.',
    'Desteklenecek projenin lisansüstü tezi üretme veya araştırmacı/öğrenci yetiştirilmesine katkı sağlama potansiyeli vardır.',
    'Yayın, patent, ödül, yarışma derecesi, bildiri ile tescil edilecek çıktılar elde etme potansiyeli vardır.'
) ON CONFLICT DO NOTHING;

-- Yapılabilirlik: Ekipman/Ortam (4)
INSERT INTO hakem_degerlendirme_baslik_soru (baslik_id, soru_id)
SELECT b.baslik_id, s.soru_id FROM hakem_degerlendirme_basliklari b, hakem_degerlendirme_sorulari s
WHERE b.baslik_adi = 'Yapılabilirlik: Ekipman/Ortam' AND s.soru_metni IN (
    'Projenin yürütüleceği bölümün/merkezin altyapısı, ortamı ve olanakları yeterlidir.',
    'Proje kapsamında istenilen ek ekipman mevcut altyapı ve proje ile uyumludur.'
) ON CONFLICT DO NOTHING;

-- Yapılabilirlik: Süre (5)
INSERT INTO hakem_degerlendirme_baslik_soru (baslik_id, soru_id)
SELECT b.baslik_id, s.soru_id FROM hakem_degerlendirme_basliklari b, hakem_degerlendirme_sorulari s
WHERE b.baslik_adi = 'Yapılabilirlik: Süre' AND s.soru_metni IN (
    'Önerilen araştırma süresi gerçekçidir.',
    'Projede her bir iş paketinin hangi sürede gerçekleştirileceği detaylandırılmıştır.',
    'Projenin başarısını olumsuz yönde etkileyebilecek riskler ve alınacak tedbirler (B Planı) belirtilmiştir.'
) ON CONFLICT DO NOTHING;

-- Yapılabilirlik: Bütçe (6)
INSERT INTO hakem_degerlendirme_baslik_soru (baslik_id, soru_id)
SELECT b.baslik_id, s.soru_id FROM hakem_degerlendirme_basliklari b, hakem_degerlendirme_sorulari s
WHERE b.baslik_adi = 'Yapılabilirlik: Bütçe' AND s.soru_metni IN (
    'Önerilen bütçe gerçekçidir ve bütçenin hazırlanmasında ekonomiklik dikkate alınmıştır.',
    'Talep edilen destek iş paketleriyle uyumlu hazırlanmıştır.'
) ON CONFLICT DO NOTHING;

-- ====================================================
-- Bildirim Tablosu
-- Kullanıcı e-posta bildirimlerinin sistem içi kopyalarını tutar
-- ====================================================
CREATE TABLE IF NOT EXISTS bildirim (
    bildirim_id SERIAL PRIMARY KEY,
    uye_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    baslik VARCHAR(255) NOT NULL,
    icerik TEXT NOT NULL,
    okundu BOOLEAN DEFAULT FALSE,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_bildirim_uye_id ON bildirim(uye_id);

-- ====================================================
-- Komisyon Başkanı Toplantı Modülü Tabloları ve Seed Verileri
-- ====================================================
CREATE TABLE IF NOT EXISTS komisyon_toplantisi (
    toplanti_id SERIAL PRIMARY KEY,
    toplanti_no VARCHAR(100) NOT NULL UNIQUE,
    tarih TIMESTAMP WITH TIME ZONE NOT NULL,
    gundem TEXT NOT NULL,
    karar TEXT NOT NULL,
    olusturan_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE SET NULL,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS komisyon_toplanti_katilimci (
    toplanti_id INTEGER NOT NULL REFERENCES komisyon_toplantisi(toplanti_id) ON DELETE CASCADE,
    uye_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    katildi BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (toplanti_id, uye_id)
);

INSERT INTO sistem_rol_tanimlama (rol_adi, rol_etiketi)
VALUES ('komisyon_baskani', 'BAP Komisyon Başkanı') 
ON CONFLICT (rol_adi) DO UPDATE SET rol_etiketi = EXCLUDED.rol_etiketi;

INSERT INTO sistem_sayfa (sayfa_adi, sayfa_kodu, url_yolu)
VALUES ('Komisyon Başkanı Dashboard', 'komisyon_baskani_dashboard', '/komisyon/baskan/dashboard') 
ON CONFLICT (sayfa_kodu) DO NOTHING;

INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
SELECT r.rol_id, s.sayfa_id
FROM sistem_rol_tanimlama r, sistem_sayfa s
WHERE r.rol_adi IN ('admin', 'komisyon_baskani') AND s.sayfa_kodu = 'komisyon_baskani_dashboard'
ON CONFLICT DO NOTHING;

-- ==========================================
-- BAP Proje Talep Tabloları
-- Her talep türü için ayrı tablo.
-- Tüm tablolar: proje_id, uye_id (talep eden),
-- talep_no (otomatik), durum, gerekce, tarih alanları içerir.
-- ==========================================

-- 1) Ek Süre Talebi
CREATE TABLE IF NOT EXISTS talep_ek_sure (
    id               SERIAL PRIMARY KEY,
    proje_id         INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    uye_id           INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    talep_no         VARCHAR(100) NOT NULL UNIQUE,
    ek_sure_ay       INTEGER NOT NULL,                        -- Kaç ay ek süre isteniyor
    gerekce          TEXT NOT NULL,
    durum            VARCHAR(50) NOT NULL DEFAULT 'beklemede', -- beklemede / onaylandi / reddedildi
    red_notu         TEXT,                                    -- Reddedilirse açıklama
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2) Ek Bütçe Talebi
CREATE TABLE IF NOT EXISTS talep_ek_butce (
    id               SERIAL PRIMARY KEY,
    proje_id         INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    uye_id           INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    talep_no         VARCHAR(100) NOT NULL UNIQUE,
    butce_kalemi     VARCHAR(200) NOT NULL,                   -- Hangi bütçe kalemine ek isteniyor
    tutar_tl         NUMERIC(15,2) NOT NULL,                  -- TL tutarı
    gerekce          TEXT NOT NULL,
    durum            VARCHAR(50) NOT NULL DEFAULT 'beklemede',
    red_notu         TEXT,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3) Fasıl Aktarımı (Kalemler Arası Aktarım) Talebi
CREATE TABLE IF NOT EXISTS talep_fasil_aktarimi (
    id               SERIAL PRIMARY KEY,
    proje_id         INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    uye_id           INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    talep_no         VARCHAR(100) NOT NULL UNIQUE,
    kaynak_kalem     VARCHAR(200) NOT NULL,                   -- Aktarım yapılacak kaynak kalem
    hedef_kalem      VARCHAR(200) NOT NULL,                   -- Aktarım yapılacak hedef kalem
    tutar_tl         NUMERIC(15,2) NOT NULL,
    gerekce          TEXT NOT NULL,
    durum            VARCHAR(50) NOT NULL DEFAULT 'beklemede',
    red_notu         TEXT,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4) Araştırmacı Ekleme/Çıkarma Talebi
CREATE TABLE IF NOT EXISTS talep_arastirmaci (
    id               SERIAL PRIMARY KEY,
    proje_id         INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    uye_id           INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    talep_no         VARCHAR(100) NOT NULL UNIQUE,
    islem_turu       VARCHAR(20) NOT NULL,                    -- 'ekleme' veya 'cikarma'
    arastirmaci_adi  VARCHAR(200) NOT NULL,                   -- Eklenecek/çıkarılacak araştırmacı adı
    gerekce          TEXT NOT NULL,
    durum            VARCHAR(50) NOT NULL DEFAULT 'beklemede',
    red_notu         TEXT,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5) Bursiyer İşlemleri Talebi
CREATE TABLE IF NOT EXISTS talep_bursiyer (
    id               SERIAL PRIMARY KEY,
    proje_id         INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    uye_id           INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    talep_no         VARCHAR(100) NOT NULL UNIQUE,
    bursiyer_kimlik  VARCHAR(50) NOT NULL,                    -- TC kimlik numarası
    bursiyer_adi     VARCHAR(200) NOT NULL,
    islem_turu       VARCHAR(50) NOT NULL,                    -- 'eklenmesi' / 'cikarilmasi' / 'degistirilmesi'
    gerekce          TEXT NOT NULL,
    durum            VARCHAR(50) NOT NULL DEFAULT 'beklemede',
    red_notu         TEXT,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 6) Proje İptali Talebi
CREATE TABLE IF NOT EXISTS talep_proje_iptali (
    id               SERIAL PRIMARY KEY,
    proje_id         INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    uye_id           INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    talep_no         VARCHAR(100) NOT NULL UNIQUE,
    gerekce          TEXT NOT NULL,
    durum            VARCHAR(50) NOT NULL DEFAULT 'beklemede',
    red_notu         TEXT,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 7) Proje Bilgi Değişimi Talebi
CREATE TABLE IF NOT EXISTS talep_bilgi_degisimi (
    id               SERIAL PRIMARY KEY,
    proje_id         INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    uye_id           INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    talep_no         VARCHAR(100) NOT NULL UNIQUE,
    degisiklik_tanimi VARCHAR(500) NOT NULL,                  -- Hangi bilgi değiştirilmek isteniyor
    gerekce          TEXT NOT NULL,
    durum            VARCHAR(50) NOT NULL DEFAULT 'beklemede',
    red_notu         TEXT,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 8) Proje Dondurma Talebi
CREATE TABLE IF NOT EXISTS talep_proje_dondurma (
    id               SERIAL PRIMARY KEY,
    proje_id         INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    uye_id           INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    talep_no         VARCHAR(100) NOT NULL UNIQUE,
    dondurma_sure_ay INTEGER NOT NULL,                        -- Kaç ay dondurulacak
    gerekce          TEXT NOT NULL,
    durum            VARCHAR(50) NOT NULL DEFAULT 'beklemede',
    red_notu         TEXT,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 9) Malzeme Güncelleme Talebi
CREATE TABLE IF NOT EXISTS talep_malzeme_guncelleme (
    id               SERIAL PRIMARY KEY,
    proje_id         INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    uye_id           INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    talep_no         VARCHAR(100) NOT NULL UNIQUE,
    guncelleme_tanimi VARCHAR(500) NOT NULL,                  -- Hangi malzeme güncellenmek isteniyor
    gerekce          TEXT NOT NULL,
    durum            VARCHAR(50) NOT NULL DEFAULT 'beklemede',
    red_notu         TEXT,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 10) Avans Talebi
CREATE TABLE IF NOT EXISTS talep_avans (
    id               SERIAL PRIMARY KEY,
    proje_id         INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    uye_id           INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    talep_no         VARCHAR(100) NOT NULL UNIQUE,
    butce_kalemi     VARCHAR(200) NOT NULL,                   -- Hangi bütçe kaleminden avans isteniyor
    tutar_tl         NUMERIC(15,2) NOT NULL,
    gerekce          TEXT NOT NULL,
    durum            VARCHAR(50) NOT NULL DEFAULT 'beklemede',
    red_notu         TEXT,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ON CONFLICT DO NOTHING;
-- =====================================================================
-- BAP PROJE SÖZLEŞMESİ TABLOSU
-- Türkçe Yorum: Akademisyen tarafından sözleşme aşamasında doldurulan
-- resmi BAP Destek Programı Proje Sözleşmesi verilerini saklar.
-- =====================================================================
CREATE TABLE IF NOT EXISTS proje_sozlesme (
    id SERIAL PRIMARY KEY,
    proje_id INT NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    uye_id INT NOT NULL REFERENCES uye(uye_id),
    tc_kimlik VARCHAR(11) NOT NULL,
    yurutucu_adres TEXT NOT NULL,
    yurutucu_telefon VARCHAR(20) NOT NULL,
    yurutucu_eposta VARCHAR(100) NOT NULL,
    baslangic_tarihi DATE NOT NULL,
    bitis_tarihi DATE NOT NULL,
    durum VARCHAR(20) DEFAULT 'dolduruldu',
    olusturma_tarihi TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_proje_sozlesme UNIQUE (proje_id)
);
