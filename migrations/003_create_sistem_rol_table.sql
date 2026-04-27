-- ====================================================
-- Sistem Rol Tablosu
-- Kullanıcı-rol atamasını yönetir
-- Bir kullanıcının hangi sistem rolüne sahip olduğunu tutar
-- ====================================================

CREATE TABLE IF NOT EXISTS sistem_rol (
    rol_id SERIAL PRIMARY KEY,
    uye_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,    -- Kullanıcı referansı
    sistem_rol_id INTEGER NOT NULL REFERENCES sistem_rol_tanimlama(rol_id) ON DELETE CASCADE,  -- Rol tanımı referansı
    UNIQUE(uye_id, sistem_rol_id)                   -- Bir kullanıcıya aynı rol bir kez atanabilir
);

-- Kullanıcı bazlı hızlı arama indeksi
CREATE INDEX IF NOT EXISTS idx_sistem_rol_uye_id ON sistem_rol(uye_id);
