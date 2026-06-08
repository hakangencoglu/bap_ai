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

-- Varsayılan Admin1 Kullanıcısı
INSERT INTO uye (rol, ad, soyad, unvan, bolum, eposta, telefon, izu_uyesi, sifre_hash, aktif_mi)
VALUES (
    'admin',
    'Admin1',
    'Yöneticisi',
    'Prof. Dr.',
    'Bilgi İşlem',
    'admin1@izu.edu.tr',
    '05555555556',
    true,
    '$2a$10$uUSxvVDTYDu4KjZXbnPx3OOJVvppVRYJcm4Dlhs0Mx8xRcmcD46ri',
    true
) ON CONFLICT (eposta) DO NOTHING;

-- Admin1 kullanıcısına sistem rolü ata
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT u.uye_id, srt.rol_id
FROM uye u, sistem_rol_tanimlama srt
WHERE u.eposta = 'admin1@izu.edu.tr' AND srt.rol_adi = 'admin'
ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;

-- Varsayılan Öğrenci Kullanıcısı
INSERT INTO uye (rol, ad, soyad, unvan, bolum, eposta, telefon, izu_uyesi, sifre_hash, aktif_mi)
VALUES (
    'ogrenci',
    'Öğrenci',
    'Kullanıcısı',
    NULL,
    'Bilgisayar Mühendisliği',
    'ogrenci@izu.edu.tr',
    '05555555557',
    true,
    '$2a$10$hWT4OaAqjGOvMW/aFMfzoOae5O53wSjgffRVHmoN/75Y3m5DRajY2',
    true
) ON CONFLICT (eposta) DO NOTHING;

-- Öğrenci kullanıcısına sistem rolü ata
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT u.uye_id, srt.rol_id
FROM uye u, sistem_rol_tanimlama srt
WHERE u.eposta = 'ogrenci@izu.edu.tr' AND srt.rol_adi = 'ogrenci'
ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;

-- Varsayılan Akademisyen Kullanıcısı
INSERT INTO uye (rol, ad, soyad, unvan, bolum, eposta, telefon, izu_uyesi, sifre_hash, aktif_mi)
VALUES (
    'akademisyen',
    'Akademisyen',
    'Kullanıcısı',
    'Prof. Dr.',
    'Bilgisayar Mühendisliği',
    'akademisyen@izu.edu.tr',
    '05555555558',
    true,
    '$2a$10$kHe1CybkltpFDSnd7PvEleMkyNcisc0C9s.Poz8v94lt0POdHfUt6',
    true
) ON CONFLICT (eposta) DO NOTHING;

-- Akademisyen kullanıcısına sistem rolü ata
INSERT INTO sistem_rol (uye_id, sistem_rol_id)
SELECT u.uye_id, srt.rol_id
FROM uye u, sistem_rol_tanimlama srt
WHERE u.eposta = 'akademisyen@izu.edu.tr' AND srt.rol_adi = 'akademisyen'
ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;

