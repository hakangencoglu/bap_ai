-- ====================================================
-- Proje tablosuna tür kolonu eklenir
-- BAP proje türünü belirtir (Yüksek Lisans, Doktora, Münferit vb.)
-- ====================================================

ALTER TABLE proje ADD COLUMN IF NOT EXISTS tur VARCHAR(100) DEFAULT 'Münferit';
