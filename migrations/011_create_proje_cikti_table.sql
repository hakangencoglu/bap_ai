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
