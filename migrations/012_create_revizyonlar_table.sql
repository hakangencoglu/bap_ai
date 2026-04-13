-- 012_create_revizyonlar_table.sql

CREATE TABLE IF NOT EXISTS revizyonlar (
    revizyon_id SERIAL PRIMARY KEY,
    proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
    olusturan_kisi_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
    atanan_kisi_id INTEGER REFERENCES uye(uye_id) ON DELETE SET NULL,
    aciklama TEXT NOT NULL,
    durum VARCHAR(50) DEFAULT 'Bekliyor',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Güncelleme tarihini otomatik yenilemek için trigger (eğer önceki dosyalarda tanımlanmışsa "update_updated_at_column" fonksiyonunu kullanır)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_updated_at_column') THEN
        CREATE TRIGGER update_revizyonlar_updated_at
        BEFORE UPDATE ON revizyonlar
        FOR EACH ROW
        EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;
