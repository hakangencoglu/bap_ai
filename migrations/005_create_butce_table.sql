-- ====================================================
-- Bütçe Tablosu
-- Orijinal BAP.sql'deki `butce` tablosunun PostgreSQL versiyonu
-- Proje bütçe kalemlerini tutar
-- ====================================================

CREATE TABLE IF NOT EXISTS butce (
    item_id SERIAL PRIMARY KEY,
    proje_id INTEGER REFERENCES proje(proje_id) ON DELETE CASCADE,  -- İlgili proje
    tur VARCHAR(500),                           -- Bütçe kalemi türü
    aciklama VARCHAR(500),                      -- Açıklama
    gerekce VARCHAR(500),                       -- Gerekçe
    urun_turu VARCHAR(500),                     -- Ürün türü
    adet INTEGER,                               -- Adet
    urun_fiyat INTEGER,                         -- Birim fiyat
    toplam_fiyat INTEGER,                       -- Toplam fiyat (adet x birim fiyat)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Proje bazlı bütçe arama indeksi
CREATE INDEX IF NOT EXISTS idx_butce_proje_id ON butce(proje_id);
