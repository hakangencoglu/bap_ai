-- ====================================================
-- Üye (Kullanıcı) Tablosu
-- Sistemdeki kullanıcıların temel bilgilerini tutar
-- Auth alanları (sifre_hash, aktif_mi) korunmuştur
-- ====================================================

CREATE TABLE IF NOT EXISTS uye (
    uye_id SERIAL PRIMARY KEY,
    rol VARCHAR(100),                               -- Kullanıcı rolü (metin olarak)
    ad VARCHAR(100) NOT NULL,                       -- Kullanıcı adı
    soyad VARCHAR(100) NOT NULL,                    -- Kullanıcı soyadı
    unvan VARCHAR(100),                             -- Akademik unvan (Prof. Dr., Doç. Dr. vb.)
    bolum VARCHAR(255),                             -- Bölüm bilgisi
    eposta VARCHAR(255) UNIQUE NOT NULL,             -- E-posta adresi (giriş için kullanılır)
    telefon VARCHAR(15),                            -- İletişim telefonu
    izu_uyesi BOOLEAN DEFAULT FALSE,                -- İZÜ üyesi mi?
    sifre_hash VARCHAR(255) NOT NULL,               -- Şifrelenmiş parola (auth için)
    aktif_mi BOOLEAN DEFAULT TRUE,                  -- Hesap aktif mi? (auth için)
    olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- E-posta alanı için hızlı arama indeksi
CREATE INDEX IF NOT EXISTS idx_uye_eposta ON uye(eposta);
