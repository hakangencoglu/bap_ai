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
	// Türkçe Yorum: Veritabanı zaten kuruluysa herhangi bir şema veya veri değişikliği yapmadan doğrudan başarılı şekilde döner.
	if exists {
		log.Println("Şema: Veritabanı zaten kurulu. Veritabanı şemasında herhangi bir güncelleme veya değişiklik yapılmadı.")
		
		// Türkçe Yorum: Mevcut veritabanında proje_kodu sütunu yoksa eklenir ve mevcut kayıtlar için sıralı şekilde proje kodları üretilerek doldurulur.
		alterQuery := `
			ALTER TABLE proje ADD COLUMN IF NOT EXISTS proje_kodu VARCHAR(100) UNIQUE;
			
			WITH numbered_projects AS (
				SELECT 
					p.proje_id,
					pbt.bap_turu,
					COALESCE(EXTRACT(YEAR FROM p.olusturma_tarihi), EXTRACT(YEAR FROM CURRENT_TIMESTAMP)) as yil,
					ROW_NUMBER() OVER (
						PARTITION BY p.bap_turu_id, EXTRACT(YEAR FROM p.olusturma_tarihi) 
						ORDER BY p.olusturma_tarihi, p.proje_id
					) as sira_no
				FROM proje p
				LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
			)
			UPDATE proje p
			SET proje_kodu = COALESCE(REPLACE(np.bap_turu, '-', ''), 'BAP') || '-' || TO_CHAR(np.yil, 'FM9999') || '-' || LPAD(np.sira_no::text, 3, '0')
			FROM numbered_projects np
			WHERE p.proje_id = np.proje_id AND p.proje_kodu IS NULL;
		`
		if _, err := db.Exec(alterQuery); err != nil {
			log.Printf("Uyarı: Proje kodu sütunu veya backfill işlemi uygulanamadı: %v", err)
		} else {
			log.Println("Bilgi: Proje kodu sütunu kontrol edildi ve mevcut boş kayıtlar için geriye dönük proje kodları oluşturuldu.")
		}

		// Türkçe Yorum: Zaten kurulu olan veritabanı için chatbot yetkilendirme alanlarını kontrol edip dinamik olarak ekliyoruz.
		chatbotPageQuery := `
			INSERT INTO sistem_sayfa (sayfa_adi, sayfa_kodu, url_yolu)
			VALUES ('Yapay Zeka Asistanı (Chatbot)', 'chatbot', '/api/chat')
			ON CONFLICT (sayfa_kodu) DO NOTHING;

			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT srt.rol_id, ss.sayfa_id
			FROM sistem_rol_tanimlama srt, sistem_sayfa ss
			WHERE ss.sayfa_kodu = 'chatbot' AND srt.rol_adi IN ('akademisyen', 'ogrenci', 'hakem', 'dekan', 'komisyon', 'tto')
			ON CONFLICT DO NOTHING;
		`
		if _, err := db.Exec(chatbotPageQuery); err != nil {
			log.Printf("Uyarı: Chatbot yetki alanları dinamik olarak eklenemedi: %v", err)
		} else {
			log.Println("Bilgi: Chatbot yetki alanları ve varsayılan rolleri veritabanına dinamik olarak eklendi.")
		}

		// Türkçe Yorum: 'yururlukte' durumunu ekleyip, mevcut 'tamamlandi' projeleri bu duruma taşıyoruz.
		yururlukteStatusQuery := `
			INSERT INTO proje_durum (durum_adi) VALUES ('yururlukte') ON CONFLICT (durum_adi) DO NOTHING;
			UPDATE proje 
			SET durum_id = (SELECT durum_id FROM proje_durum WHERE durum_adi = 'yururlukte')
			WHERE durum_id = (SELECT durum_id FROM proje_durum WHERE durum_adi = 'tamamlandi');
		`
		if _, err := db.Exec(yururlukteStatusQuery); err != nil {
			log.Printf("Uyarı: 'yururlukte' durum göçü uygulanamadı: %v", err)
		} else {
			log.Println("Bilgi: 'yururlukte' durum göçü ve proje güncellemeleri başarıyla uygulandı.")
		}
		
		// Türkçe Yorum: Mevcut veritabanına yeni iş akışı için gerekli eksik tüm durumları ekle.
		newStatusQuery := `
			INSERT INTO proje_durum (durum_adi) VALUES ('dekan_onayi_bekliyor') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('dekan_onayladi') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('komisyon_bekliyor') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('komisyon_onayladi') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('hakem_atama_bekliyor') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('hakem_bekliyor') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('hakem_onayladi') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('sozlesme_imza') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('tto_aktif') ON CONFLICT (durum_adi) DO NOTHING;
		`
		if _, err := db.Exec(newStatusQuery); err != nil {
			log.Printf("Uyarı: Yeni durum kayıtları (hakem_bekliyor, sozlesme_imza) eklenemedi: %v", err)
		} else {
			log.Println("Bilgi: Yeni iş akışı durum kayıtları (hakem_bekliyor, sozlesme_imza) başarıyla eklendi.")
		}

		// Türkçe Yorum: Yeni proje_asama tablosunu oluştur (iş akışı onay masaları).
		// Bu tablo; "Dekan Onayına Sun", "Komisyona Sun", "Hakeme Sun" gibi
		// kullanıcıya gösterilen Türkçe aşama isimlerini tutar.
		projeAsamaQuery := `
			CREATE TABLE IF NOT EXISTS proje_asama (
				asama_id   SERIAL PRIMARY KEY,
				asama_kodu VARCHAR(100) UNIQUE NOT NULL,
				asama_adi  VARCHAR(200) NOT NULL,
				sira_no    INTEGER DEFAULT 0
			);

			INSERT INTO proje_asama (asama_kodu, asama_adi, sira_no) VALUES
				('tto_on_inceleme',   'TTO Ön İnceleme',  1),
				('dekan_onayina_sun', 'Dekan Onayına Sun', 2),
				('komisyona_sun',     'Komisyona Sun',      3),
				('hakeme_sun',        'Hakeme Sun',         4),
				('sozlesme_imza',     'Sözleşme İmzası',    5)
			ON CONFLICT (asama_kodu) DO NOTHING;

			ALTER TABLE proje ADD COLUMN IF NOT EXISTS asama_id INTEGER REFERENCES proje_asama(asama_id) ON DELETE SET NULL;
			CREATE INDEX IF NOT EXISTS idx_proje_asama_id ON proje(asama_id);
		`
		if _, err := db.Exec(projeAsamaQuery); err != nil {
			log.Printf("Uyarı: proje_asama tablosu veya asama_id sütunu oluşturulamadı: %v", err)
		} else {
			log.Println("Bilgi: proje_asama tablosu ve proje.asama_id sütunu başarıyla kontrol edildi/oluşturuldu.")
		}

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

