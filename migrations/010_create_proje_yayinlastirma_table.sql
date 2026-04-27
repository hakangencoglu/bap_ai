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
