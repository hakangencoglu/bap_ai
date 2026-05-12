-- ================================================================
-- Migration 017: BAP proje türlerini BAP-100..500 olarak güncelle
-- Mevcut kayıtları yeni isimlendirmeye uyumlu hale getirir.
-- ================================================================

-- Mevcut eski isimleri güncelle (varsa)
UPDATE proje_bap_turu SET bap_turu = 'BAP-100' WHERE bap_turu = 'Yüksek Lisans Tez Projesi';
UPDATE proje_bap_turu SET bap_turu = 'BAP-200' WHERE bap_turu = 'Doktora Tez Projesi';
UPDATE proje_bap_turu SET bap_turu = 'BAP-300' WHERE bap_turu = 'Münferit Araştırma Projesi';
UPDATE proje_bap_turu SET bap_turu = 'BAP-400' WHERE bap_turu = 'Hızlı Destek Projesi';
UPDATE proje_bap_turu SET bap_turu = 'BAP-500' WHERE bap_turu = 'Altyapı Projesi';

-- Eğer hiç kayıt yoksa yeni ekle
INSERT INTO proje_bap_turu (bap_turu) VALUES
    ('BAP-100'),
    ('BAP-200'),
    ('BAP-300'),
    ('BAP-400'),
    ('BAP-500')
ON CONFLICT (bap_turu) DO NOTHING;
