-- ====================================================
-- Roller Tablosu
-- Sistemdeki kullanıcı rollerini tanımlar
-- ====================================================

CREATE TABLE IF NOT EXISTS roles (
    role_id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,           -- Örn: admin, akademisyen, ogrenci
    description TEXT,                           -- Rol açıklaması
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Varsayılan rollerin eklenmesi
INSERT INTO roles (name, description) VALUES
    ('admin', 'Sistem Yöneticisi - Tüm yetkilere sahip kullanıcı'),
    ('akademisyen', 'Akademisyen - BAP başvurusu yapabilen ve değerlendirebilen kullanıcı'),
    ('ogrenci', 'Öğrenci - BAP başvurusu yapabilen kullanıcı')
ON CONFLICT (name) DO NOTHING;
