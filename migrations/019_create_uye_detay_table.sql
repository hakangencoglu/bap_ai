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
