package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Veritabanı bağlantı ayarları
	// Veritabanı bağlantı ayarları
	connStr := "host=db port=5432 user=bap password=bap_admin_1 dbname=bap_app sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err == nil {
		err = db.Ping()
	}

	if err != nil {
		// Yerel bağlantı için farklı bir deneme
		connStr = "host=localhost port=5432 user=bap password=bap_admin_1 dbname=bap_app sslmode=disable"
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Fatal(err)
		}
	}

	query := `
	INSERT INTO uye (role_id, unvan, ad, soyad, bolum, iletisim_mail, iletisim_tel, password_hash, is_active, izu_akademisyen, izu_ogrenci) 
	VALUES (
		1, 
		'Prof. Dr.',
		'Sistem',
		'Yöneticisi',
		'Bilgi İşlem',
		'admin@izu.edu.tr',
		'05555555555',
		'$2a$10$XU0d2U/N5z/qP.yB2uIq/eZg4hO6/r.Q3Nq.7xO4a/yD/u0tZ8y/K', 
		true,
		true,
		false
	) ON CONFLICT (iletisim_mail) DO NOTHING;
	`

	_, err = db.Exec(query)
	if err != nil {
		log.Printf("Veritabanı ekleme hatası: %v", err)
		return
	}
	fmt.Println("Admin hesabi basariyla veritabanina eklendi!")
}
