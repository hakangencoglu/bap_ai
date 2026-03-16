package database

import (
	"database/sql"
	"fmt"
	"log"

	"bap_ai/configs"

	_ "github.com/lib/pq" // PostgreSQL sürücüsü
)

// DB, uygulamanın kullanacağı veritabanı bağlantı havuzunu tutar.
var DB *sql.DB

// Connect fonksiyonu, PostgreSQL veritabanına bağlantıyı başlatır.
func Connect() {
	// Konfigürasyondan veritabanı bağlantı bilgilerini alır
	cfg := configs.AppConfig

	// Bağlantı dizesini oluşturur
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName,
	)

	var err error
	// Veritabanı bağlantısını açar
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Veritabanı bağlantısı açılamadı: %v", err)
	}

	// Bağlantının gerçekten çalışıp çalışmadığını kontrol eder
	if err = DB.Ping(); err != nil {
		log.Fatalf("Veritabanına erişilemiyor: %v", err)
	}

	log.Println("PostgreSQL veritabanına başarıyla bağlanıldı.")
}
