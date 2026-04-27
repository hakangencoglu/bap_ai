-- ====================================================
-- Bütçe Tablosu
-- Proje bütçe kalemlerini tutar
-- ====================================================

CREATE TABLE IF NOT EXISTS butce (
    kalem_id SERIAL PRIMARY KEY,                     -- Bütçe kalemi ID
    proje_id INTEGER REFERENCES proje(proje_id) ON DELETE CASCADE,  -- İlgili proje
    kategori_id INTEGER REFERENCES butce_kategori(kategori_id) ON DELETE SET NULL,  -- Bütçe kategorisi
    aciklama VARCHAR(500),                           -- Kalem açıklaması
    birim_ozelligi INTEGER,                          -- Birim özelliği/spec
    birim_fiyat NUMERIC(10,2) DEFAULT 0,             -- Birim fiyat
    toplam_fiyat NUMERIC(12,2) DEFAULT 0,            -- Toplam fiyat
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Proje bazlı bütçe arama indeksi
CREATE INDEX IF NOT EXISTS idx_butce_proje_id ON butce(proje_id);
