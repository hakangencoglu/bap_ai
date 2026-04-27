-- ====================================================
-- Varsayılan Veriler
-- Sistem başlangıç verileri (admin kullanıcı vb.)
-- ====================================================

-- Varsayılan Admin Kullanıcısı
INSERT INTO uye (rol, ad, soyad, unvan, bolum, eposta, telefon, izu_uyesi, sifre_hash, aktif_mi)
VALUES (
    'admin',
    'Sistem',
    'Yöneticisi',
    'Prof. Dr.',
    'Bilgi İşlem',
    'admin@izu.edu.tr',
    '05555555555',
    true,
    '$2a$10$XU0d2U/N5z/qP.yB2uIq/eZg4hO6/r.Q3Nq.7xO4a/yD/u0tZ8y/K',
    true
) ON CONFLICT (eposta) DO NOTHING;

-- Admin kullanıcısına sistem rolü ata
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT u.uye_id, srt.rol_id
FROM uye u, sistem_rol_tanimlama srt
WHERE u.eposta = 'admin@izu.edu.tr' AND srt.rol_adi = 'admin'
ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;
