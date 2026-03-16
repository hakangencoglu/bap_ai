-- ====================================================
-- Proje Tablosu
-- Orijinal BAP.sql'deki `proje` tablosunun PostgreSQL versiyonu
-- ====================================================

CREATE TABLE IF NOT EXISTS proje (
    proje_id SERIAL PRIMARY KEY,
    baslik_tr VARCHAR(100),                     -- Proje başlığı (Türkçe)
    baslik_en VARCHAR(100),                     -- Proje başlığı (İngilizce)
    baslangic_tarihi DATE,                      -- Proje başlangıç tarihi
    proje_suresi DATE,                          -- Proje bitiş tarihi
    toplam_tutar INTEGER,                       -- Toplam bütçe tutarı
    etik_kurul BOOLEAN DEFAULT FALSE,           -- Etik kurul onayı gerekli mi?
    ozet_tr VARCHAR(1000),                      -- Proje özeti (Türkçe)
    ozet_en VARCHAR(1000),                      -- Proje özeti (İngilizce)
    amac_ve_hedef TEXT,                         -- Amaç ve hedef açıklaması
    ozgunluk TEXT,                              -- Projenin özgünlüğü
    metodoloji TEXT,                            -- Kullanılacak metodoloji
    risk_yonetimi TEXT,                         -- Risk yönetim planı
    proje_ciktilari TEXT,                       -- Beklenen proje çıktıları
    aktivite_bilgisi TEXT,                      -- Aktivite bilgileri
    aktivite_fizibilitesi TEXT,                 -- Aktivite fizibilite değerlendirmesi
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
