-- ====================================================
-- Araştırma Tablosu
-- Projeye bağlı araştırma bilgilerini tutar
-- ====================================================

CREATE TABLE IF NOT EXISTS arastirma (
    proje_id INTEGER PRIMARY KEY REFERENCES proje(proje_id) ON DELETE CASCADE,  -- Proje referansı (1'e 1)
    olusturan_id INTEGER REFERENCES uye(uye_id) ON DELETE SET NULL,             -- Oluşturan kişi
    arastirma_amaci VARCHAR(500)                     -- Araştırma amacı
);
