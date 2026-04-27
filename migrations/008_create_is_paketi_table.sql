-- ====================================================
-- İş Paketi Tablosu
-- Projelerin iş paketlerini (görev dağılımı) tutar
-- ====================================================

CREATE TABLE IF NOT EXISTS is_paketi (
    paket_id SERIAL PRIMARY KEY,
    proje_id INTEGER REFERENCES proje(proje_id) ON DELETE CASCADE,  -- İlgili proje
    paket_adi VARCHAR(500),                         -- İş paketinin adı
    paket_amaci VARCHAR(500),                       -- İş paketinin amacı
    baslangic_tarihi DATE,                          -- Başlangıç tarihi
    bitis_tarihi DATE,                              -- Bitiş tarihi
    olusturma_tarihi DATE DEFAULT CURRENT_DATE,     -- Oluşturulma tarihi
    guncelleme_tarihi DATE DEFAULT CURRENT_DATE     -- Güncelleme tarihi
);

-- Proje bazlı iş paketi arama indeksi
CREATE INDEX IF NOT EXISTS idx_is_paketi_proje_id ON is_paketi(proje_id);
