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

