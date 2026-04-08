-- ====================================================
-- Proje Değerlendirmeleri Tablosu
-- Hakemlerin projeleri değerlendirmelerini tutar
-- ====================================================

CREATE TABLE IF NOT EXISTS proje_degerlendirmeleri (
    degerlendirme_id SERIAL PRIMARY KEY,
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    hakem_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    puan INTEGER CHECK (puan >= 0 AND puan <= 100), -- 0-100 arası genel puan
    yorum TEXT,                                       -- Hakemin genel proje yorumu
    durum VARCHAR(50) DEFAULT 'Bekliyor',             -- Bekliyor, Onaylandı, Reddedildi, Revizyon
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(proje_id, hakem_id)                        -- Bir hakem bir projeyi bir kez değerlendirebilir
);
