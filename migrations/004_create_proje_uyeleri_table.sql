-- ====================================================
-- Proje Üyeleri Tablosu
-- Orijinal BAP.sql'deki `proje_uyeleri` tablosunun PostgreSQL versiyonu
-- Proje ile üye arasındaki ilişkiyi (çoka çok) tutar
-- ====================================================

CREATE TABLE IF NOT EXISTS proje_uyeleri (
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    uye_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    rol VARCHAR(100),                           -- Üyenin projedeki rolü (yürütücü, araştırmacı vb.)
    PRIMARY KEY (proje_id, uye_id)
);

-- Üye bazlı hızlı arama indeksi
CREATE INDEX IF NOT EXISTS idx_proje_uyeleri_uye_id ON proje_uyeleri(uye_id);
