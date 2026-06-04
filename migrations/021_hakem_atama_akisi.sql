-- ================================================================
-- Migration 021: Hakem Atama Kabul/Red Akışı
-- Hakemlerin atamayı kabul veya reddetme sürecini yönetir.
-- Mevcut 'durum' alanına dokunulmaz (öğrenci görünümü korunur).
-- Yeni 'atama_durumu' alanı admin/akademisyen görünümü için eklenir.
-- ================================================================

-- Hakem atama kabul/red durumu sütunu eklenir
-- Değerler: 'Atandı', 'Kabul Edildi', 'Reddedildi'
ALTER TABLE proje_degerlendirmeleri
    ADD COLUMN IF NOT EXISTS atama_durumu VARCHAR(50) DEFAULT 'Atandı';

-- Hakemin atamayı reddetme sebebini tutmak için alan
ALTER TABLE proje_degerlendirmeleri
    ADD COLUMN IF NOT EXISTS red_nedeni TEXT;

-- Hakemin atamayı kabul/red ettiği tarih
ALTER TABLE proje_degerlendirmeleri
    ADD COLUMN IF NOT EXISTS karar_tarihi TIMESTAMP WITH TIME ZONE;

-- Mevcut kayıtları geriye dönük uyumluluk için 'Kabul Edildi' olarak işaretle
UPDATE proje_degerlendirmeleri
    SET atama_durumu = 'Kabul Edildi'
    WHERE atama_durumu IS NULL OR atama_durumu = 'Atandı';
