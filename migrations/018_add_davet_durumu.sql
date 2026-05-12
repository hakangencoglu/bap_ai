-- ================================================================
-- Migration 018: Proje takım tablosuna davet durumu sütunu ekleme
-- Ekip üyeleri davet edildiğinde beklemede, kabul veya red durumuna geçer
-- ================================================================

-- Davet durumu sütunu eklenir (varsayılan: beklemede)
ALTER TABLE proje_takim ADD COLUMN IF NOT EXISTS davet_durumu VARCHAR(20) DEFAULT 'beklemede';

-- Mevcut kayıtları otomatik kabul olarak işaretle (geriye uyumluluk)
UPDATE proje_takim SET davet_durumu = 'kabul' WHERE davet_durumu IS NULL OR davet_durumu = 'beklemede';
