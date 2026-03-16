-- ====================================================
-- Üye (Kullanıcı) Tablosu
-- Orijinal BAP.sql'deki `uye` tablosunun PostgreSQL versiyonu
-- Kimlik doğrulama alanları (email, password_hash) eklendi
-- ====================================================

CREATE TABLE IF NOT EXISTS uye (
    uye_id SERIAL PRIMARY KEY,
    role_id INTEGER REFERENCES roles(role_id) ON DELETE SET NULL,  -- Kullanıcı rolü
    unvan VARCHAR(100),                         -- Akademik unvan (Prof. Dr., Doç. Dr. vb.)
    ad VARCHAR(100) NOT NULL,                   -- Kullanıcı adı
    soyad VARCHAR(100) NOT NULL,                -- Kullanıcı soyadı
    bolum VARCHAR(255),                         -- Bölüm bilgisi
    iletisim_tel VARCHAR(20),                   -- İletişim telefonu (VARCHAR olarak düzeltildi)
    iletisim_mail VARCHAR(255) UNIQUE NOT NULL,  -- E-posta adresi (giriş için kullanılır)
    password_hash VARCHAR(255) NOT NULL,         -- Şifrelenmiş parola
    izu_akademisyen BOOLEAN DEFAULT FALSE,       -- İZÜ akademisyeni mi?
    izu_ogrenci BOOLEAN DEFAULT FALSE,           -- İZÜ öğrencisi mi?
    is_active BOOLEAN DEFAULT TRUE,              -- Hesap aktif mi?
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- E-posta alanı için hızlı arama indeksi
CREATE INDEX IF NOT EXISTS idx_uye_iletisim_mail ON uye(iletisim_mail);
