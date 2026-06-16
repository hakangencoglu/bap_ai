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
    rol_adi VARCHAR(100) UNIQUE NOT NULL            -- Örn: admin, akademisyen, ogrenci, hakem
);

-- Proje Rol Tanımlama: Proje içindeki roller
CREATE TABLE IF NOT EXISTS proje_rol_tanimlama (
    rol_id SERIAL PRIMARY KEY,
    proje_rol VARCHAR(100) UNIQUE NOT NULL           -- Örn: Yürütücü, Araştırmacı, Danışman
);

-- Proje Durum Tanımlama: Proje durumları
CREATE TABLE IF NOT EXISTS proje_durum (
    durum_id SERIAL PRIMARY KEY,
    durum_adi VARCHAR(100) UNIQUE NOT NULL            -- Örn: taslak, incelemede, onaylandi, reddedildi, tamamlandi
);

-- Proje BAP Türü: BAP proje türleri
CREATE TABLE IF NOT EXISTS proje_bap_turu (
    bap_turu_id SERIAL PRIMARY KEY,
    bap_turu VARCHAR(200) UNIQUE NOT NULL             -- Örn: Yüksek Lisans, Doktora, Münferit
);

-- Proje Çıktı Türü: Çıktı türlerinin tanımları
CREATE TABLE IF NOT EXISTS proje_cikti_turu (
    cikti_turu_id SERIAL PRIMARY KEY,
    cikti_turu VARCHAR(200) UNIQUE NOT NULL           -- Örn: Makale, Tez, Patent, Rapor
);

-- Proje Etik: Etik durum tanımları
CREATE TABLE IF NOT EXISTS proje_etik (
    etik_id SERIAL PRIMARY KEY,
    etik_durumu VARCHAR(200) UNIQUE NOT NULL           -- Örn: Gerekli, Gerekli Değil, Alındı
);

-- Bütçe Tanım: Bütçe türü tanımları
CREATE TABLE IF NOT EXISTS butce_tanim (
    tanim_tip_id SERIAL PRIMARY KEY,
    tanim_adi VARCHAR(200) NOT NULL                   -- Örn: Sarf Malzeme, Hizmet Alımı, Yolluk
);

-- Bütçe Kategori: Bütçe kategorileri
CREATE TABLE IF NOT EXISTS butce_kategori (
    kategori_id SERIAL PRIMARY KEY,
    kategori_adi VARCHAR(200) UNIQUE NOT NULL          -- Örn: Donanım, Yazılım, Seyahat
);

-- Olanak Türü: Olanak türleri
CREATE TABLE IF NOT EXISTS olanak_tur (
    olanak_tur_id SERIAL PRIMARY KEY,
    tur_adi VARCHAR(200) NOT NULL                     -- Örn: Laboratuvar, Kütüphane
);

-- ====================================================
-- Varsayılan Lookup Verileri
-- ====================================================

-- Sistem rolleri
INSERT INTO sistem_rol_tanimlama (rol_adi) VALUES
    ('admin'),
    ('akademisyen'),
    ('ogrenci'),
    ('hakem'),
    ('dekan'),
    ('komisyon'),
    ('tto')
ON CONFLICT (rol_adi) DO NOTHING;

-- Proje rolleri
INSERT INTO proje_rol_tanimlama (proje_rol) VALUES
    ('Yürütücü'),
    ('Araştırmacı'),
    ('Danışman'),
    ('Bursiyer')
ON CONFLICT (proje_rol) DO NOTHING;

-- Proje durumları
INSERT INTO proje_durum (durum_adi) VALUES
    ('taslak'),
    ('incelemede'),
    ('onaylandi'),
    ('reddedildi'),
    ('tamamlandi'),
    ('revizyon'),
    ('dekan_onayi_bekliyor'),
    ('komisyon_bekliyor'),
    ('tto_aktif')
ON CONFLICT (durum_adi) DO NOTHING;

-- BAP türleri
INSERT INTO proje_bap_turu (bap_turu) VALUES
    ('BAP-100'),
    ('BAP-200'),
    ('BAP-300'),
    ('BAP-400'),
    ('BAP-500')
ON CONFLICT (bap_turu) DO NOTHING;

-- Çıktı türleri
INSERT INTO proje_cikti_turu (cikti_turu) VALUES
    ('SCI/SSCI Makale'),
    ('Ulusal Makale'),
    ('Tez'),
    ('Patent'),
    ('Bildiri'),
    ('Rapor'),
    ('Kitap/Kitap Bölümü')
ON CONFLICT (cikti_turu) DO NOTHING;

-- Etik durumları
INSERT INTO proje_etik (etik_durumu) VALUES
    ('Gerekli'),
    ('Gerekli Değil'),
    ('Alındı'),
    ('Başvuru Yapıldı')
ON CONFLICT (etik_durumu) DO NOTHING;

-- Bütçe kategorileri
INSERT INTO butce_kategori (kategori_adi) VALUES
    ('Makine-Teçhizat'),
    ('Sarf Malzeme'),
    ('Hizmet Alımı'),
    ('Seyahat (Yolluk)'),
    ('Yazılım'),
    ('Yayın/Basım')
ON CONFLICT (kategori_adi) DO NOTHING;




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
    rol VARCHAR(100),                               -- Kullanıcı rolü (metin olarak)
    ad VARCHAR(100) NOT NULL,                       -- Kullanıcı adı
    soyad VARCHAR(100) NOT NULL,                    -- Kullanıcı soyadı
    unvan VARCHAR(100),                             -- Akademik unvan (Prof. Dr., Doç. Dr. vb.)
    bolum VARCHAR(255),                             -- Bölüm bilgisi
    eposta VARCHAR(255) UNIQUE NOT NULL,             -- E-posta adresi (giriş için kullanılır)
    telefon VARCHAR(15),                            -- İletişim telefonu
    izu_uyesi BOOLEAN DEFAULT FALSE,                -- İZÜ üyesi mi?
    sifre_hash VARCHAR(255) NOT NULL,               -- Şifrelenmiş parola (auth için)
    aktif_mi BOOLEAN DEFAULT TRUE,                  -- Hesap aktif mi? (auth için)
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
    uye_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,    -- Kullanıcı referansı
    sistem_rol_id INTEGER NOT NULL REFERENCES sistem_rol_tanimlama(rol_id) ON DELETE CASCADE,  -- Rol tanımı referansı
    UNIQUE(uye_id, sistem_rol_id)                   -- Bir kullanıcıya aynı rol bir kez atanabilir
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
    baslik_tr VARCHAR(500),                         -- Proje başlığı (Türkçe)
    baslik_en VARCHAR(500),                         -- Proje başlığı (İngilizce)
    sure_ay INTEGER,                                -- Proje süresi (ay cinsinden)
    toplam_butce NUMERIC(12,2) DEFAULT 0,           -- Toplam bütçe tutarı
    etik_kurul BOOLEAN DEFAULT FALSE,               -- Etik kurul onayı gerekli mi?
    etik_kurul_no INTEGER,                          -- Etik kurul numarası
    koordinator_id INTEGER REFERENCES uye(uye_id) ON DELETE SET NULL,  -- Proje koordinatörü (üye FK)
    durum_id INTEGER REFERENCES proje_durum(durum_id) ON DELETE SET NULL,  -- Proje durumu (FK)
    bap_turu_id INTEGER REFERENCES proje_bap_turu(bap_turu_id) ON DELETE SET NULL,  -- BAP türü (FK)
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Durum bazlı hızlı arama indeksi
CREATE INDEX IF NOT EXISTS idx_proje_durum_id ON proje(durum_id);
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
    proje_id INTEGER PRIMARY KEY REFERENCES proje(proje_id) ON DELETE CASCADE,  -- Proje referansı (1'e 1 ilişki)
    ozet VARCHAR(500),                              -- Proje özeti (abstract)
    anahtar_kelimeler VARCHAR(100),                 -- Anahtar kelimeler (keywords)
    hedefler VARCHAR(500),                          -- Proje hedefleri (objectives)
    ozgunluk VARCHAR(100),                          -- Projenin özgünlüğü (originality)
    metodoloji VARCHAR(500)                         -- Kullanılacak metodoloji (methodology)
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
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,     -- Proje referansı
    uye_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,           -- Üye referansı
    proje_rol_id INTEGER REFERENCES proje_rol_tanimlama(rol_id) ON DELETE SET NULL,  -- Proje rolü referansı
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
    kalem_id SERIAL PRIMARY KEY,                     -- Bütçe kalemi ID
    proje_id INTEGER REFERENCES proje(proje_id) ON DELETE CASCADE,  -- İlgili proje
    kategori_id INTEGER REFERENCES butce_kategori(kategori_id) ON DELETE SET NULL,  -- Bütçe kategorisi
    aciklama VARCHAR(500),                           -- Kalem açıklaması
    birim_ozelligi INTEGER,                          -- Birim özelliği/spec
    birim_fiyat NUMERIC(10,2) DEFAULT 0,             -- Birim fiyat
    toplam_fiyat NUMERIC(12,2) DEFAULT 0,            -- Toplam fiyat
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
    proje_id INTEGER REFERENCES proje(proje_id) ON DELETE CASCADE,  -- İlgili proje
    paket_adi VARCHAR(500),                         -- İş paketinin adı
    paket_amaci VARCHAR(500),                       -- İş paketinin amacı
    baslangic_tarihi DATE,                          -- Başlangıç tarihi
    bitis_tarihi DATE,                              -- Bitiş tarihi
    olusturma_tarihi DATE DEFAULT CURRENT_DATE,     -- Oluşturulma tarihi
    guncelleme_tarihi DATE DEFAULT CURRENT_DATE     -- Güncelleme tarihi
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
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,  -- Proje referansı
    paket_id INTEGER REFERENCES is_paketi(paket_id) ON DELETE SET NULL,      -- İş paketi referansı (opsiyonel)
    risk_aciklamasi VARCHAR(500),                    -- Risk açıklaması
    cozum_plani VARCHAR(500)                        -- Çözüm planı
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
    proje_id INTEGER PRIMARY KEY REFERENCES proje(proje_id) ON DELETE CASCADE,  -- Proje referansı (1'e 1)
    yayin_turu VARCHAR(200),                         -- Yayın türü
    yayin_ciktisi VARCHAR(200),                      -- Çıktı bilgisi
    tahmini_yayin_tarihi DATE                        -- Tahmini yayın tarihi
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
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,  -- Proje referansı
    cikti_turu_id INTEGER REFERENCES proje_cikti_turu(cikti_turu_id) ON DELETE SET NULL,  -- Çıktı türü referansı
    aciklama VARCHAR(500),                           -- Çıktı açıklaması
    cikti_periyodu DATE                              -- Çıktı periyodu/tarihi
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
    proje_id INTEGER PRIMARY KEY REFERENCES proje(proje_id) ON DELETE CASCADE,  -- Proje referansı (1'e 1)
    olusturan_id INTEGER REFERENCES uye(uye_id) ON DELETE SET NULL,             -- Oluşturan kişi
    arastirma_amaci VARCHAR(500)                     -- Araştırma amacı
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
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,  -- Değerlendirilen proje
    hakem_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,      -- Değerlendiren hakem
    puan INTEGER CHECK (puan >= 0 AND puan <= 100),  -- 0-100 arası genel puan
    yorum TEXT,                                       -- Hakemin genel proje yorumu
    durum VARCHAR(50) DEFAULT 'Bekliyor',             -- Bekliyor, Onaylandı, Reddedildi, Revizyon
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(proje_id, hakem_id)                        -- Bir hakem bir projeyi bir kez değerlendirebilir
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
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,     -- İlgili proje
    olusturan_kisi_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,  -- Revizyonu oluşturan kişi
    atanan_kisi_id INTEGER REFERENCES uye(uye_id) ON DELETE SET NULL,           -- Revizyonun atandığı kişi
    aciklama TEXT NOT NULL,                            -- Revizyon açıklaması
    durum VARCHAR(50) DEFAULT 'Bekliyor',              -- Bekliyor, Tamamlandı
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
INSERT INTO uye (rol, ad, soyad, unvan, bolum, eposta, telefon, izu_uyesi, sifre_hash, aktif_mi)
VALUES (
    'admin',
    'Sistem',
    'Yöneticisi',
    'Prof. Dr.',
    'Bilgi İşlem',
    'admin@izu.edu.tr',
    '05555555555',
    true,
    '$2a$10$XU0d2U/N5z/qP.yB2uIq/eZg4hO6/r.Q3Nq.7xO4a/yD/u0tZ8y/K',
    true
) ON CONFLICT (eposta) DO NOTHING;

-- Admin kullanıcısına sistem rolü ata
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT u.uye_id, srt.rol_id
FROM uye u, sistem_rol_tanimlama srt
WHERE u.eposta = 'admin@izu.edu.tr' AND srt.rol_adi = 'admin'
ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;

-- Varsayılan Admin1 Kullanıcısı
INSERT INTO uye (rol, ad, soyad, unvan, bolum, eposta, telefon, izu_uyesi, sifre_hash, aktif_mi)
VALUES (
    'admin',
    'Admin1',
    'Yöneticisi',
    'Prof. Dr.',
    'Bilgi İşlem',
    'admin1@izu.edu.tr',
    '05555555556',
    true,
    '$2a$10$uUSxvVDTYDu4KjZXbnPx3OOJVvppVRYJcm4Dlhs0Mx8xRcmcD46ri',
    true
) ON CONFLICT (eposta) DO NOTHING;

-- Admin1 kullanıcısına sistem rolü ata
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT u.uye_id, srt.rol_id
FROM uye u, sistem_rol_tanimlama srt
WHERE u.eposta = 'admin1@izu.edu.tr' AND srt.rol_adi = 'admin'
ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;

-- Varsayılan Öğrenci Kullanıcısı
INSERT INTO uye (rol, ad, soyad, unvan, bolum, eposta, telefon, izu_uyesi, sifre_hash, aktif_mi)
VALUES (
    'ogrenci',
    'Öğrenci',
    'Kullanıcısı',
    NULL,
    'Bilgisayar Mühendisliği',
    'ogrenci@izu.edu.tr',
    '05555555557',
    true,
    '$2a$10$hWT4OaAqjGOvMW/aFMfzoOae5O53wSjgffRVHmoN/75Y3m5DRajY2',
    true
) ON CONFLICT (eposta) DO NOTHING;

-- Öğrenci kullanıcısına sistem rolü ata
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT u.uye_id, srt.rol_id
FROM uye u, sistem_rol_tanimlama srt
WHERE u.eposta = 'ogrenci@izu.edu.tr' AND srt.rol_adi = 'ogrenci'
ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;

-- Varsayılan Akademisyen Kullanıcısı
INSERT INTO uye (rol, ad, soyad, unvan, bolum, eposta, telefon, izu_uyesi, sifre_hash, aktif_mi)
VALUES (
    'akademisyen',
    'Akademisyen',
    'Kullanıcısı',
    'Prof. Dr.',
    'Bilgisayar Mühendisliği',
    'akademisyen@izu.edu.tr',
    '05555555558',
    true,
    '$2a$10$kHe1CybkltpFDSnd7PvEleMkyNcisc0C9s.Poz8v94lt0POdHfUt6',
    true
) ON CONFLICT (eposta) DO NOTHING;

-- Akademisyen kullanıcısına sistem rolü ata
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT u.uye_id, srt.rol_id
FROM uye u, sistem_rol_tanimlama srt
WHERE u.eposta = 'akademisyen@izu.edu.tr' AND srt.rol_adi = 'akademisyen'
ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;





-- ==========================================
-- Migration: 016_add_pdf_dosya_yolu.sql
-- ==========================================
-- ================================================================
-- Migration 016: Proje tablosuna PDF dosya yolu sütunu eklenmesi
-- Bu sütun, nihai onaylanan başvuru PDF'inin sunucu yolunu tutar.
-- ================================================================

ALTER TABLE proje ADD COLUMN IF NOT EXISTS pdf_dosya_yolu TEXT;

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
UPDATE proje_bap_turu SET bap_turu = 'BAP-100' WHERE bap_turu = 'Yüksek Lisans Tez Projesi';
UPDATE proje_bap_turu SET bap_turu = 'BAP-200' WHERE bap_turu = 'Doktora Tez Projesi';
UPDATE proje_bap_turu SET bap_turu = 'BAP-300' WHERE bap_turu = 'Münferit Araştırma Projesi';
UPDATE proje_bap_turu SET bap_turu = 'BAP-400' WHERE bap_turu = 'Hızlı Destek Projesi';
UPDATE proje_bap_turu SET bap_turu = 'BAP-500' WHERE bap_turu = 'Altyapı Projesi';

-- Eğer hiç kayıt yoksa yeni ekle
INSERT INTO proje_bap_turu (bap_turu) VALUES
    ('BAP-100'),
    ('BAP-200'),
    ('BAP-300'),
    ('BAP-400'),
    ('BAP-500')
ON CONFLICT (bap_turu) DO NOTHING;




-- ==========================================
-- Migration: 018_add_davet_durumu.sql
-- ==========================================
-- ================================================================
-- Migration 018: Proje takım tablosuna davet durumu sütunu ekleme
-- Ekip üyeleri davet edildiğinde beklemede, kabul veya red durumuna geçer
-- ================================================================

-- Davet durumu sütunu eklenir (varsayılan: beklemede)
ALTER TABLE proje_takim ADD COLUMN IF NOT EXISTS davet_durumu VARCHAR(20) DEFAULT 'beklemede';

-- Mevcut kayıtları otomatik kabul olarak işaretle (geriye uyumluluk)
UPDATE proje_takim SET davet_durumu = 'kabul' WHERE davet_durumu IS NULL OR davet_durumu = 'beklemede';




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
    uye_id INTEGER UNIQUE NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,  -- Üye ile 1:1 ilişki
    rol VARCHAR(100),                               -- Kullanıcı rolü (akademisyen, ogrenci, hakem, admin)
    unvan VARCHAR(100),                             -- Akademik unvan (Prof. Dr., Doç. Dr. vb.)
    bolum VARCHAR(255),                             -- Bölüm / Fakülte bilgisi
    telefon VARCHAR(15),                            -- İletişim telefonu
    izu_uyesi BOOLEAN DEFAULT FALSE,                -- İZÜ üyesi mi?
    profil_tamamlandi BOOLEAN DEFAULT FALSE,        -- Profil bilgileri tamamlandı mı?
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
INSERT INTO uye_detay (uye_id, rol, unvan, bolum, telefon, izu_uyesi, profil_tamamlandi)
SELECT 
    uye_id,
    COALESCE(rol, 'ogrenci'),
    unvan,
    bolum,
    telefon,
    izu_uyesi,
    TRUE  -- Mevcut kullanıcıların profili zaten tamamlanmış kabul edilir
FROM uye
WHERE NOT EXISTS (
    SELECT 1 FROM uye_detay WHERE uye_detay.uye_id = uye.uye_id
);




-- ==========================================
-- Migration: 019_make_bap_turu_dynamic.sql
-- ==========================================
-- ================================================================
-- Migration 019: BAP Proje Türleri Tablosunun Dinamik Hale Getirilmesi
-- Admin tarafından tanımlanabilecek yeni sütunların eklenmesi
-- ================================================================

-- Sütunlar eklenir (varsayılan değerlerle)
ALTER TABLE proje_bap_turu ADD COLUMN IF NOT EXISTS butce_limiti NUMERIC(12,2) DEFAULT 0;
ALTER TABLE proje_bap_turu ADD COLUMN IF NOT EXISTS sure_limiti_ay INTEGER DEFAULT 0;
ALTER TABLE proje_bap_turu ADD COLUMN IF NOT EXISTS aktif_mi BOOLEAN DEFAULT TRUE;
ALTER TABLE proje_bap_turu ADD COLUMN IF NOT EXISTS aciklama TEXT DEFAULT '';
ALTER TABLE proje_bap_turu ADD COLUMN IF NOT EXISTS hakem_gerekli BOOLEAN DEFAULT FALSE;
ALTER TABLE proje_bap_turu ADD COLUMN IF NOT EXISTS hakem_sayisi INTEGER DEFAULT 0;
ALTER TABLE proje_bap_turu ADD COLUMN IF NOT EXISTS bursiyer_gerekli BOOLEAN DEFAULT FALSE;
ALTER TABLE proje_bap_turu ADD COLUMN IF NOT EXISTS bursiyer_sayisi INTEGER DEFAULT 0;

-- Mevcut varsayılan BAP türlerini gerçekçi değerlerle güncelle
UPDATE proje_bap_turu SET butce_limiti = 50000.00, sure_limiti_ay = 12, aciklama = 'Yüksek Lisans Tez Projesi Desteği' WHERE bap_turu = 'BAP-100';
UPDATE proje_bap_turu SET butce_limiti = 100000.00, sure_limiti_ay = 24, aciklama = 'Doktora Tez Projesi Desteği' WHERE bap_turu = 'BAP-200';
UPDATE proje_bap_turu SET butce_limiti = 150000.00, sure_limiti_ay = 18, aciklama = 'Münferit Araştırma Projesi Desteği' WHERE bap_turu = 'BAP-300';
UPDATE proje_bap_turu SET butce_limiti = 30000.00, sure_limiti_ay = 6, aciklama = 'Hızlı Destek Projesi' WHERE bap_turu = 'BAP-400';
UPDATE proje_bap_turu SET butce_limiti = 250000.00, sure_limiti_ay = 36, aciklama = 'Altyapı Projesi Desteği' WHERE bap_turu = 'BAP-500';




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
SELECT d.uye_id, srt.rol_id
FROM uye_detay d
INNER JOIN sistem_rol_tanimlama srt ON srt.rol_adi = d.rol
WHERE d.rol IS NOT NULL AND d.rol != ''
ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;




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
SELECT d.uye_id, srt.rol_id
FROM uye_detay d
INNER JOIN sistem_rol_tanimlama srt ON srt.rol_adi = d.rol
WHERE d.rol IS NOT NULL AND d.rol != ''
ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;




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

-- Mevcut kayıtları geriye dönük uyumluluk için 'Kabul Edildi' olarak işaretle
UPDATE proje_degerlendirmeleri
    SET atama_durumu = 'Kabul Edildi'
    WHERE atama_durumu IS NULL OR atama_durumu = 'Atandı';


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
