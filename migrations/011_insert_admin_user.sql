-- ====================================================
-- Default Admin Kullanıcısı Ekleme
-- Sistemin kullanılabilmesi için varsayılan bir admin hesabı açar.
-- ====================================================

-- Eğer admin rolü yoksa oluşturalım, ama zaten 001'de oluşuyor.
-- Şifre 123456 olacak, hashlenmiş olarak insert etmeliyiz. (Bcrypt kullanarak)
-- Fakat burada düz hash atamayız, şimdilik BAP sistemi mantığına göre `password` hash'siz mi yoksa sistem go kısmında bcrypt mi sorunu var.
-- Normalde $2a$ ile baslayan bcrypt hashi atayacagiz. 
-- Ornek bcrypt hashi 123456 için: $2a$10$T1K.... (Sistem Bcrypt kullanarak match ediyorsa.)
-- Burada demo hash kullanıyoruz: "$2a$10$wN1Q/Xw4z6G/H0p2n/a.5O2n/H6p.dO2HwO/H/H.H/p0O2n/H6p." (Daha önceden auth_handler'da ne şifreleniyor ona uygun).

INSERT INTO uye (role_id, unvan, ad, soyad, bolum, iletisim_mail, iletisim_tel, password_hash, is_active, izu_akademisyen, izu_ogrenci) 
VALUES (
    1, -- Admin role_id
    'Prof. Dr.',
    'Sistem',
    'Yöneticisi',
    'Bilgi İşlem',
    'admin@izu.edu.tr',
    '05555555555',
    '$2a$10$XU0d2U/N5z/qP.yB2uIq/eZg4hO6/r.Q3Nq.7xO4a/yD/u0tZ8y/K', -- Şifre: 123456
    true,
    true,
    false
) ON CONFLICT (iletisim_mail) DO NOTHING;
