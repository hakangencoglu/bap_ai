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
