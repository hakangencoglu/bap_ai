-- ====================================================
-- Proje Değerlendirmeleri Tablosu
-- Hakemlerin projeleri değerlendirmelerini tutar
-- (ER diyagramında yok, mevcut sistem için korunmuştur)
-- ====================================================

CREATE TABLE IF NOT EXISTS proje_degerlendirmeleri (
    degerlendirme_id SERIAL PRIMARY KEY,
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,  -- Değerlendirilen proje
    hakem_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,      -- Değerlendiren hakem
    puan INTEGER CHECK (puan >= 0 AND puan <= 100),  -- 0-100 arası genel puan
    yorum TEXT,                                       -- Hakemin genel proje yorumu
    durum VARCHAR(50) DEFAULT 'Bekliyor',             -- Bekliyor, Onaylandı, Reddedildi, Revizyon
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(proje_id, hakem_id)                        -- Bir hakem bir projeyi bir kez değerlendirebilir
);
