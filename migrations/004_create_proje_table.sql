-- ====================================================
-- Proje Ana Tablosu
-- BAP projelerinin temel bilgilerini tutar
-- ====================================================

CREATE TABLE IF NOT EXISTS proje (
    proje_id SERIAL PRIMARY KEY,
    baslik_tr VARCHAR(500),                         -- Proje başlığı (Türkçe)
    baslik_en VARCHAR(500),                         -- Proje başlığı (İngilizce)
    sure_ay INTEGER,                                -- Proje süresi (ay cinsinden)
    toplam_butce NUMERIC(12,2) DEFAULT 0,           -- Toplam bütçe tutarı
    etik_kurul BOOLEAN DEFAULT FALSE,               -- Etik kurul onayı gerekli mi?
    etik_kurul_no INTEGER,                          -- Etik kurul numarası
    koordinator_id INTEGER REFERENCES uye(uye_id) ON DELETE SET NULL,  -- Proje koordinatörü (üye FK)
    durum_id INTEGER REFERENCES proje_durum(durum_id) ON DELETE SET NULL,  -- Proje durumu (FK)
    bap_turu_id INTEGER REFERENCES proje_bap_turu(bap_turu_id) ON DELETE SET NULL,  -- BAP türü (FK)
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Durum bazlı hızlı arama indeksi
CREATE INDEX IF NOT EXISTS idx_proje_durum_id ON proje(durum_id);
-- Koordinatör bazlı hızlı arama indeksi
CREATE INDEX IF NOT EXISTS idx_proje_koordinator_id ON proje(koordinator_id);
