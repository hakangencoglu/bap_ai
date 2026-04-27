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
