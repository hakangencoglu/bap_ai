package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

// RunSchema veritabanı şemasını schema.sql dosyasından okuyarak uygular
func RunSchema(db *sql.DB, schemaPath string) error {
	// Veritabanının daha önce kurulup kurulmadığını kontrol et
	var exists bool
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = 'uye'
	);`
	
	// 'uye' tablosunun varlığını sorgula
	err := db.QueryRow(query).Scan(&exists)
	if err != nil {
		return fmt.Errorf("veritabanı durumu kontrol edilemedi: %v", err)
	}

	// Eğer tablo zaten varsa şema kurulumunu atla (veri kaybı ve sıfırlanmayı önlemek için)
	// Ancak yeni eklenen tabloları, durumları ve rolleri dinamik olarak senkronize et
	if exists {
		log.Println("Şema: Veritabanı zaten kurulu, dinamik senkronizasyon adımları çalıştırılıyor...")
		
		// E-İmza tablosunu oluştur
		// Türkçe Yorum: E-İmza kayıtlarını tutacak olan tabloyu ve ilgili indeksleri dinamik olarak oluşturuyoruz.
		imzaTableQuery := `
			CREATE TABLE IF NOT EXISTS proje_imza (
				imza_id SERIAL PRIMARY KEY,
				proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
				uye_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
				imzaci_ad_soyad VARCHAR(255) NOT NULL,
				rol VARCHAR(100) NOT NULL,
				imza_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				imza_durumu VARCHAR(50) DEFAULT 'İmzalandı',
				imza_token VARCHAR(255),
				ip_adresi VARCHAR(50),
				tarayici_bilgisi VARCHAR(255),
				UNIQUE(proje_id, uye_id)
			);
			CREATE INDEX IF NOT EXISTS idx_proje_imza_proje_id ON proje_imza(proje_id);
			CREATE INDEX IF NOT EXISTS idx_proje_imza_uye_id ON proje_imza(uye_id);
		`
		if _, err := db.Exec(imzaTableQuery); err != nil {
			return fmt.Errorf("proje_imza tablosu oluşturulamadı: %v", err)
		}

		// Yeni ve temel sistem rollerini ekle
		roleQuery := `
			INSERT INTO sistem_rol_tanimlama (rol_adi) VALUES
				('admin'),
				('akademisyen'),
				('ogrenci'),
				('hakem'),
				('dekan'),
				('komisyon'),
				('tto')
			ON CONFLICT (rol_adi) DO NOTHING;
		`
		if _, err := db.Exec(roleQuery); err != nil {
			return fmt.Errorf("yeni sistem rolleri eklenemedi: %v", err)
		}

		// Yeni proje durumlarını ekle
		statusQuery := `
			INSERT INTO proje_durum (durum_adi) VALUES
				('revizyon'),
				('dekan_onayi_bekliyor'),
				('komisyon_bekliyor'),
				('tto_aktif')
			ON CONFLICT (durum_adi) DO NOTHING;
		`
		if _, err := db.Exec(statusQuery); err != nil {
			return fmt.Errorf("yeni proje durumları eklenemedi: %v", err)
		}

		// Süreç geçmişi tablosunu oluştur
		historyTableQuery := `
			CREATE TABLE IF NOT EXISTS proje_surec_gecmisi (
				gecmis_id SERIAL PRIMARY KEY,
				proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
				islem_yapan_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
				baslangic_durum VARCHAR(100),
				hedef_durum VARCHAR(100),
				aciklama TEXT,
				olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_proje_surec_gecmisi_proje_id ON proje_surec_gecmisi(proje_id);
		`
		if _, err := db.Exec(historyTableQuery); err != nil {
			return fmt.Errorf("proje_surec_gecmisi tablosu oluşturulamadı: %v", err)
		}

		// Hakem atama akışı sütunlarını ekle (atama_durumu, red_nedeni, karar_tarihi)
		hakemAtamaQuery := `
			ALTER TABLE proje_degerlendirmeleri
				ADD COLUMN IF NOT EXISTS atama_durumu VARCHAR(50) DEFAULT 'Atandı';
			ALTER TABLE proje_degerlendirmeleri
				ADD COLUMN IF NOT EXISTS red_nedeni TEXT;
			ALTER TABLE proje_degerlendirmeleri
				ADD COLUMN IF NOT EXISTS karar_tarihi TIMESTAMP WITH TIME ZONE;
		`
		if _, err := db.Exec(hakemAtamaQuery); err != nil {
			return fmt.Errorf("hakem atama sütunları eklenemedi: %v", err)
		}

		// Zorunlu şifre değiştirme sütununu ekle
		// Türkçe Yorum: Uye tablosuna zorunlu şifre değiştirme sütununu ekliyoruz.
		uyeSifreZorlaQuery := `
			ALTER TABLE uye
				ADD COLUMN IF NOT EXISTS sifre_degistir_zorla BOOLEAN DEFAULT FALSE;
		`
		if _, err := db.Exec(uyeSifreZorlaQuery); err != nil {
			return fmt.Errorf("zorunlu şifre değiştirme sütunu eklenemedi: %v", err)
		}

		// BAP türüne hakem ve bursiyer sütunlarını ekle
		// Türkçe Yorum: BAP türü tanımlarken hakem ve bursiyer gereksinimlerini belirleyebilmek için yeni sütunlar ekleniyor.
		bapHakemBursiyerQuery := `
			ALTER TABLE proje_bap_turu
				ADD COLUMN IF NOT EXISTS hakem_gerekli BOOLEAN DEFAULT FALSE;
			ALTER TABLE proje_bap_turu
				ADD COLUMN IF NOT EXISTS hakem_sayisi INTEGER DEFAULT 0;
			ALTER TABLE proje_bap_turu
				ADD COLUMN IF NOT EXISTS bursiyer_gerekli BOOLEAN DEFAULT FALSE;
			ALTER TABLE proje_bap_turu
				ADD COLUMN IF NOT EXISTS bursiyer_sayisi INTEGER DEFAULT 0;
		`
		if _, err := db.Exec(bapHakemBursiyerQuery); err != nil {
			return fmt.Errorf("BAP hakem/bursiyer sütunları eklenemedi: %v", err)
		}

		// Revizyon tablosuna revizyon_bolum sütununu ekle
		// Türkçe Yorum: Revizyonun hangi bölümü etkilediğini belirten revizyon_bolum sütununu ekliyoruz.
		revizyonBolumQuery := `
			ALTER TABLE revizyonlar
				ADD COLUMN IF NOT EXISTS revizyon_bolum VARCHAR(100);
		`
		if _, err := db.Exec(revizyonBolumQuery); err != nil {
			return fmt.Errorf("revizyon tablosuna revizyon_bolum sütunu eklenemedi: %v", err)
		}

		// Yetki tablosu ve tohum verileri senkronizasyonu
		// Türkçe Yorum: Dinamik sayfa-rol yetkilendirme tablolarını oluşturup tohum (seed) verilerini ekliyoruz.
		yetkiQuery := `
			CREATE TABLE IF NOT EXISTS sistem_sayfa (
				sayfa_id SERIAL PRIMARY KEY,
				sayfa_adi VARCHAR(200) NOT NULL,
				sayfa_kodu VARCHAR(100) UNIQUE NOT NULL,
				url_yolu VARCHAR(255) UNIQUE NOT NULL
			);
			
			CREATE TABLE IF NOT EXISTS sayfa_rol_yetki (
				yetki_id SERIAL PRIMARY KEY,
				sistem_rol_id INTEGER NOT NULL REFERENCES sistem_rol_tanimlama(rol_id) ON DELETE CASCADE,
				sayfa_id INTEGER NOT NULL REFERENCES sistem_sayfa(sayfa_id) ON DELETE CASCADE,
				UNIQUE(sistem_rol_id, sayfa_id)
			);
			
			CREATE INDEX IF NOT EXISTS idx_sayfa_rol_yetki_rol ON sayfa_rol_yetki(sistem_rol_id);
			CREATE INDEX IF NOT EXISTS idx_sayfa_rol_yetki_sayfa ON sayfa_rol_yetki(sayfa_id);

			INSERT INTO sistem_sayfa (sayfa_adi, sayfa_kodu, url_yolu) VALUES
				('Anasayfa', 'anasayfa', '/anasayfa'),
				('Yeni Başvuru Formu', 'basvuru', '/basvuru'),
				('Profil Sayfası', 'profil', '/profil'),
				('Hakem Dashboard', 'hakem_dashboard', '/hakem/dashboard'),
				('Proje Değerlendirme', 'hakem_degerlendirme', '/hakem/degerlendirme'),
				('Admin Dashboard', 'admin_dashboard', '/admin/dashboard'),
				('Hakem Atama', 'admin_hakem_atama', '/admin/hakem-atama'),
				('Proje Durum Raporları', 'admin_projects_status', '/admin/projects/status'),
				('Dekan Dashboard', 'dekan_dashboard', '/dekan/dashboard'),
				('Komisyon Dashboard', 'komisyon_dashboard', '/komisyon/dashboard'),
				('TTO Dashboard', 'tto_dashboard', '/tto/dashboard'),
				('E-İmza Paneli', 'eimza', '/eimza')
			ON CONFLICT (sayfa_kodu) DO NOTHING;

			-- Admin yetkileri (Tüm sayfalara erişebilir)
			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT r.rol_id, s.sayfa_id FROM sistem_rol_tanimlama r, sistem_sayfa s
			WHERE r.rol_adi = 'admin'
			ON CONFLICT DO NOTHING;

			-- Akademisyen yetkileri (Anasayfa, Başvuru ve Profil görebilir)
			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT r.rol_id, s.sayfa_id FROM sistem_rol_tanimlama r, sistem_sayfa s
			WHERE r.rol_adi = 'akademisyen' AND s.sayfa_kodu IN ('anasayfa', 'basvuru', 'profil')
			ON CONFLICT DO NOTHING;

			-- Öğrenci yetkileri (Anasayfa, Başvuru ve Profil görebilir)
			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT r.rol_id, s.sayfa_id FROM sistem_rol_tanimlama r, sistem_sayfa s
			WHERE r.rol_adi = 'ogrenci' AND s.sayfa_kodu IN ('anasayfa', 'basvuru', 'profil')
			ON CONFLICT DO NOTHING;

			-- Hakem yetkileri (Anasayfa, Profil, Hakem Dashboard ve Değerlendirme görebilir)
			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT r.rol_id, s.sayfa_id FROM sistem_rol_tanimlama r, sistem_sayfa s
			WHERE r.rol_adi = 'hakem' AND s.sayfa_kodu IN ('anasayfa', 'profil', 'hakem_dashboard', 'hakem_degerlendirme')
			ON CONFLICT DO NOTHING;

			-- Dekan yetkileri (Anasayfa, Profil ve Dekan Dashboard görebilir)
			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT r.rol_id, s.sayfa_id FROM sistem_rol_tanimlama r, sistem_sayfa s
			WHERE r.rol_adi = 'dekan' AND s.sayfa_kodu IN ('anasayfa', 'profil', 'dekan_dashboard')
			ON CONFLICT DO NOTHING;

			-- Komisyon yetkileri (Anasayfa, Profil ve Komisyon Dashboard görebilir)
			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT r.rol_id, s.sayfa_id FROM sistem_rol_tanimlama r, sistem_sayfa s
			WHERE r.rol_adi = 'komisyon' AND s.sayfa_kodu IN ('anasayfa', 'profil', 'komisyon_dashboard')
			ON CONFLICT DO NOTHING;

			-- TTO yetkileri (Anasayfa, Profil ve TTO Dashboard görebilir)
			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT r.rol_id, s.sayfa_id FROM sistem_rol_tanimlama r, sistem_sayfa s
			WHERE r.rol_adi = 'tto' AND s.sayfa_kodu IN ('anasayfa', 'profil', 'tto_dashboard')
			ON CONFLICT DO NOTHING;

			-- E-İmza Paneli yetkileri (Tüm rollere eimza sayfası yetkisi verilir)
			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT r.rol_id, s.sayfa_id FROM sistem_rol_tanimlama r, sistem_sayfa s
			WHERE s.sayfa_kodu = 'eimza'
			ON CONFLICT DO NOTHING;
		`
		if _, err := db.Exec(yetkiQuery); err != nil {
			return fmt.Errorf("yetki tablolari ve tohum verileri yuklenemedi: %v", err)
		}

		log.Println("Şema: Dinamik senkronizasyon başarıyla tamamlandı.")
		return nil
	}

	// Şema dosyasını oku
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("şema dosyası okunamadı (%s): %v", schemaPath, err)
	}

	// Şemayı çalıştır
	_, err = db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("şema çalıştırılamadı (%s): %v", schemaPath, err)
	}

	log.Printf("Veritabanı şeması başarıyla uygulandı: %s\n", schemaPath)
	return nil
}

