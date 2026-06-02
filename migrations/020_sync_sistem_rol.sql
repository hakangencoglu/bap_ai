-- ================================================================
-- Migration 020: Mevcut kullanıcıların sistem_rol tablosuna senkronizasyonu
-- uye_detay tablosundaki rol bilgisine göre sistem_rol ilişki tablosu doldurulur.
-- Bu migration sadece bir kez çalıştırılmalıdır.
-- ================================================================

-- Mevcut tüm kullanıcıların rollerini sistem_rol tablosuna ekle
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT d.uye_id, srt.rol_id
FROM uye_detay d
INNER JOIN sistem_rol_tanimlama srt ON srt.rol_adi = d.rol
WHERE d.rol IS NOT NULL AND d.rol != ''
ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;
