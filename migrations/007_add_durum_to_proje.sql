-- ====================================================
-- Proje tablosuna durum kolonu eklenir
-- Proje durumlarını takip etmek için kullanılır
-- Olası değerler: taslak, incelemede, onaylandi, reddedildi, tamamlandi
-- ====================================================

ALTER TABLE proje ADD COLUMN IF NOT EXISTS durum VARCHAR(50) DEFAULT 'taslak';

-- Durum bazlı hızlı arama indeksi
CREATE INDEX IF NOT EXISTS idx_proje_durum ON proje(durum);
