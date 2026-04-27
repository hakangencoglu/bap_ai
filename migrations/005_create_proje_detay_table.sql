-- ====================================================
-- Proje Detay Tablosu
-- Projenin akademik detay alanlarını tutar
-- (Özet, anahtar kelimeler, hedefler, özgünlük, metodoloji)
-- ====================================================

CREATE TABLE IF NOT EXISTS proje_detay (
    proje_id INTEGER PRIMARY KEY REFERENCES proje(proje_id) ON DELETE CASCADE,  -- Proje referansı (1'e 1 ilişki)
    ozet VARCHAR(500),                              -- Proje özeti (abstract)
    anahtar_kelimeler VARCHAR(100),                 -- Anahtar kelimeler (keywords)
    hedefler VARCHAR(500),                          -- Proje hedefleri (objectives)
    ozgunluk VARCHAR(100),                          -- Projenin özgünlüğü (originality)
    metodoloji VARCHAR(500)                         -- Kullanılacak metodoloji (methodology)
);
