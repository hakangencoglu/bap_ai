-- ====================================================
-- DB Admin: Kurulum Fakülte Verileri Göçü (Migration)
-- Üniversite bünyesindeki tüm fakülte ve enstitü tanımlarının eklenmesi
-- ====================================================

INSERT INTO fakulte (fakulte_adi, fakulte_kodu, kisa_ad) VALUES
('Eğitim Fakültesi', 'EF', 'Eğitim'),
('Hukuk Fakültesi', 'HF', 'Hukuk'),
('İnsan ve Toplum Bilimleri Fakültesi', 'İTBF', 'İnsan ve Toplum'),
('İslami İlimler Fakültesi', 'İİF', 'İslami İlimler'),
('İşletme ve Yönetim Bilimleri Fakültesi', 'İYBF', 'İşletme'),
('Lisansüstü Eğitim Enstitüsü', 'LEE', 'Lisansüstü'),
('Mühendislik ve Doğa Bilimleri Fakültesi', 'MDBF', 'Mühendislik'),
('Rektörlüğe Bağlı Bölümler', 'RBB', 'Rektörlük'),
('Sağlık Bilimleri Fakültesi', 'SBF', 'Sağlık Bilimleri'),
('Spor Bilimleri Fakültesi', 'SPBF', 'Spor Bilimleri')
ON CONFLICT (fakulte_kodu) DO UPDATE SET fakulte_adi = EXCLUDED.fakulte_adi, kisa_ad = EXCLUDED.kisa_ad;
