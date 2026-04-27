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
    ('hakem')
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
    ('revizyon')
ON CONFLICT (durum_adi) DO NOTHING;

-- BAP türleri
INSERT INTO proje_bap_turu (bap_turu) VALUES
    ('Yüksek Lisans Tez Projesi'),
    ('Doktora Tez Projesi'),
    ('Münferit Araştırma Projesi'),
    ('Hızlı Destek Projesi'),
    ('Altyapı Projesi')
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
