-- ================================================================
-- Migration 019: BAP Proje Türleri Tablosunun Dinamik Hale Getirilmesi
-- Admin tarafından tanımlanabilecek yeni sütunların eklenmesi
-- ================================================================

-- Sütunlar eklenir (varsayılan değerlerle)
ALTER TABLE proje_bap_turu ADD COLUMN IF NOT EXISTS butce_limiti NUMERIC(12,2) DEFAULT 0;
ALTER TABLE proje_bap_turu ADD COLUMN IF NOT EXISTS sure_limiti_ay INTEGER DEFAULT 0;
ALTER TABLE proje_bap_turu ADD COLUMN IF NOT EXISTS aktif_mi BOOLEAN DEFAULT TRUE;
ALTER TABLE proje_bap_turu ADD COLUMN IF NOT EXISTS aciklama TEXT DEFAULT '';

-- Mevcut varsayılan BAP türlerini gerçekçi değerlerle güncelle
UPDATE proje_bap_turu SET butce_limiti = 50000.00, sure_limiti_ay = 12, aciklama = 'Yüksek Lisans Tez Projesi Desteği' WHERE bap_turu = 'BAP-100';
UPDATE proje_bap_turu SET butce_limiti = 100000.00, sure_limiti_ay = 24, aciklama = 'Doktora Tez Projesi Desteği' WHERE bap_turu = 'BAP-200';
UPDATE proje_bap_turu SET butce_limiti = 150000.00, sure_limiti_ay = 18, aciklama = 'Münferit Araştırma Projesi Desteği' WHERE bap_turu = 'BAP-300';
UPDATE proje_bap_turu SET butce_limiti = 30000.00, sure_limiti_ay = 6, aciklama = 'Hızlı Destek Projesi' WHERE bap_turu = 'BAP-400';
UPDATE proje_bap_turu SET butce_limiti = 250000.00, sure_limiti_ay = 36, aciklama = 'Altyapı Projesi Desteği' WHERE bap_turu = 'BAP-500';
