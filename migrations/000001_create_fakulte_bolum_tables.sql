-- ====================================================
-- DB Admin: Fakülte, Bölüm ve Fakülte-Bölüm İlişki Tabloları Göçü (Migration)
-- ====================================================

-- 1. Fakülte Tablosu
CREATE TABLE IF NOT EXISTS fakulte (
    fakulte_id       SERIAL PRIMARY KEY,
    fakulte_adi      VARCHAR(255) NOT NULL UNIQUE,
    fakulte_kodu     VARCHAR(50) UNIQUE,
    kisa_ad          VARCHAR(100),
    aktif            BOOLEAN NOT NULL DEFAULT TRUE,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fakulte_aktif ON fakulte(aktif);

-- 2. Bölüm Tablosu
CREATE TABLE IF NOT EXISTS bolum (
    bolum_id         SERIAL PRIMARY KEY,
    bolum_adi        VARCHAR(255) NOT NULL,
    bolum_kodu       VARCHAR(50) UNIQUE,
    kisa_ad          VARCHAR(100),
    aktif            BOOLEAN NOT NULL DEFAULT TRUE,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bolum_aktif ON bolum(aktif);

-- 3. Fakülte - Bölüm İlişki Tablosu (Junction / İlişki Tablosu)
CREATE TABLE IF NOT EXISTS fakulte_bolum (
    id               SERIAL PRIMARY KEY,
    fakulte_id       INTEGER NOT NULL REFERENCES fakulte(fakulte_id) ON DELETE CASCADE,
    bolum_id         INTEGER NOT NULL REFERENCES bolum(bolum_id) ON DELETE CASCADE,
    aktif            BOOLEAN NOT NULL DEFAULT TRUE,
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_fakulte_bolum UNIQUE (fakulte_id, bolum_id)
);

CREATE INDEX IF NOT EXISTS idx_fakulte_bolum_fakulte ON fakulte_bolum(fakulte_id);
CREATE INDEX IF NOT EXISTS idx_fakulte_bolum_bolum ON fakulte_bolum(bolum_id);
CREATE INDEX IF NOT EXISTS idx_fakulte_bolum_aktif ON fakulte_bolum(aktif);

-- ====================================================
-- Başlangıç Referans Verileri (Seed Data)
-- ====================================================

-- Örnek / Başlangıç Fakülte Tanımları
INSERT INTO fakulte (fakulte_adi, fakulte_kodu, kisa_ad) VALUES
('Mühendislik ve Doğa Bilimleri Fakültesi', 'MDBF', 'Mühendislik'),
('İnsan ve Toplum Bilimleri Fakültesi', 'ITBF', 'İnsan ve Toplum'),
('İşletme ve Yönetim Bilimleri Fakültesi', 'IYBF', 'İşletme'),
('İslami İlimler Fakültesi', 'IIF', 'İslami İlimler'),
('Hukuk Fakültesi', 'HF', 'Hukuk'),
('Sağlık Bilimleri Fakültesi', 'SBF', 'Sağlık Bilimleri'),
('Tıp Fakültesi', 'TF', 'Tıp'),
('Eğitim Fakültesi', 'EF', 'Eğitim')
ON CONFLICT (fakulte_adi) DO NOTHING;

-- Örnek / Başlangıç Bölüm Tanımları
INSERT INTO bolum (bolum_adi, bolum_kodu, kisa_ad) VALUES
('Bilgisayar Mühendisliği', 'CENG', 'Bilgisayar Müh.'),
('Yazılım Mühendisliği', 'SENG', 'Yazılım Müh.'),
('Elektrik-Elektronik Mühendisliği', 'EEEN', 'Elek.-Elektr. Müh.'),
('Endüstri Mühendisliği', 'IE', 'Endüstri Müh.'),
('Psikoloji', 'PSYK', 'Psikoloji'),
('Siyaset Bilimi ve Uluslararası İlişkiler', 'POLS', 'Siyaset Bil.'),
('Tarih', 'HIST', 'Tarih'),
('Türk Dili ve Edebiyatı', 'TDE', 'Türk Dili'),
('İşletme', 'BUS', 'İşletme'),
('İktisat', 'ECON', 'İktisat'),
('Uluslararası Ticaret ve Finansman', 'IBF', 'Ulus. Ticaret'),
('Temel İslam Bilimleri', 'TIB', 'Temel İslam'),
('Hukuk', 'LAW', 'Hukuk'),
('Hemşirelik', 'NURS', 'Hemşirelik'),
('Beslenme ve Diyetetik', 'NUT', 'Beslenme'),
('Fizyoterapi ve Rehabilitasyon', 'FTR', 'Fizyoterapi'),
('Temel Tıp Bilimleri', 'MED', 'Tıp'),
('Özel Eğitim Öğretmenliği', 'SPED', 'Özel Eğitim'),
('Rehberlik ve Psikolojik Danışmanlık', 'RPD', 'PDR')
ON CONFLICT (bolum_kodu) DO NOTHING;

-- Fakülte - Bölüm İlişkilerinin Oluşturulması
INSERT INTO fakulte_bolum (fakulte_id, bolum_id)
SELECT f.fakulte_id, b.bolum_id
FROM fakulte f, bolum b
WHERE (f.fakulte_kodu = 'MDBF' AND b.bolum_kodu IN ('CENG', 'SENG', 'EEEN', 'IE'))
   OR (f.fakulte_kodu = 'ITBF' AND b.bolum_kodu IN ('PSYK', 'POLS', 'HIST', 'TDE'))
   OR (f.fakulte_kodu = 'IYBF' AND b.bolum_kodu IN ('BUS', 'ECON', 'IBF'))
   OR (f.fakulte_kodu = 'IIF'  AND b.bolum_kodu IN ('TIB'))
   OR (f.fakulte_kodu = 'HF'   AND b.bolum_kodu IN ('LAW'))
   OR (f.fakulte_kodu = 'SBF'  AND b.bolum_kodu IN ('NURS', 'NUT', 'FTR'))
   OR (f.fakulte_kodu = 'TF'   AND b.bolum_kodu IN ('MED'))
ON CONFLICT (fakulte_id, bolum_id) DO NOTHING;

-- 4. Kullanıcı (Üye) Tablosuna Fakülte ve Bölüm İlişkisi Eklenmesi
ALTER TABLE uye ADD COLUMN IF NOT EXISTS fakulte_id INTEGER REFERENCES fakulte(fakulte_id) ON DELETE SET NULL;
ALTER TABLE uye ADD COLUMN IF NOT EXISTS bolum_id INTEGER REFERENCES bolum(bolum_id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_uye_fakulte ON uye(fakulte_id);
CREATE INDEX IF NOT EXISTS idx_uye_bolum ON uye(bolum_id);
