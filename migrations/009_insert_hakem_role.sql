-- ====================================================
-- Hakem Rolü Ekleme
-- Sisteme hakem rolünü dahil eder
-- ====================================================

INSERT INTO roller (name, description) VALUES
    ('hakem', 'Hakem - BAP projelerini değerlendirmekle yükümlü akademisyen/uzman')
ON CONFLICT (name) DO NOTHING;
