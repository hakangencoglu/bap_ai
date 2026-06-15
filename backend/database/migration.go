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
		
		// Yeni sistem rollerini ekle
		roleQuery := `
			INSERT INTO sistem_rol_tanimlama (rol_adi) VALUES
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

