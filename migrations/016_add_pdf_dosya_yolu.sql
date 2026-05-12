-- ================================================================
-- Migration 016: Proje tablosuna PDF dosya yolu sütunu eklenmesi
-- Bu sütun, nihai onaylanan başvuru PDF'inin sunucu yolunu tutar.
-- ================================================================

ALTER TABLE proje ADD COLUMN IF NOT EXISTS pdf_dosya_yolu TEXT;

-- Yorum: Bu alan nullable'dır çünkü PDF ancak başvuru
-- onaylandıktan sonra üretilir ve kaydedilir.
