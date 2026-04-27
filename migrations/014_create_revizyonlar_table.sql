-- ====================================================
-- Revizyonlar Tablosu
-- Proje revizyonlarını ve durumlarını takip eder
-- (ER diyagramında yok, mevcut sistem için korunmuştur)
-- ====================================================

CREATE TABLE IF NOT EXISTS revizyonlar (
    revizyon_id SERIAL PRIMARY KEY,
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,     -- İlgili proje
    olusturan_kisi_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,  -- Revizyonu oluşturan kişi
    atanan_kisi_id INTEGER REFERENCES uye(uye_id) ON DELETE SET NULL,           -- Revizyonun atandığı kişi
    aciklama TEXT NOT NULL,                            -- Revizyon açıklaması
    durum VARCHAR(50) DEFAULT 'Bekliyor',              -- Bekliyor, Tamamlandı
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
