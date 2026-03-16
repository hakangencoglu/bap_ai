-- ====================================================
-- İş Paketleri Tablosu
-- Orijinal BAP.sql'deki `is_paketleri` tablosunun PostgreSQL versiyonu
-- Projelerin iş paketlerini (görev dağılımı) tutar
-- ====================================================

CREATE TABLE IF NOT EXISTS is_paketleri (
    paket_id SERIAL PRIMARY KEY,
    proje_id INTEGER REFERENCES proje(proje_id) ON DELETE CASCADE,  -- İlgili proje
    paket_isim VARCHAR(500),                    -- İş paketinin adı
    paket_amac VARCHAR(500),                    -- İş paketinin amacı
    baslangic_tarih DATE,                       -- Başlangıç tarihi
    bitis_tarih DATE,                           -- Bitiş tarihi
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Proje bazlı iş paketi arama indeksi
CREATE INDEX IF NOT EXISTS idx_is_paketleri_proje_id ON is_paketleri(proje_id);
