package database

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"bap_ai/configs"

	_ "github.com/lib/pq" // PostgreSQL sürücüsü
)

// DB, uygulamanın kullanacağı veritabanı bağlantı havuzunu tutar.
var DB *sql.DB

// Connect fonksiyonu, uzak PostgreSQL veritabanı sunucusuna bağlantıyı başlatır ve havuz ayarlarını yapılandırır.
func Connect() {
	// Konfigürasyondan veritabanı bağlantı bilgilerini alır
	cfg := configs.AppConfig

	// Bağlantı dizesini oluşturur (uzak DB sunucusu için)
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

	// Uzak DB sunucusu için bağlantı havuzu yapılandırması
	configureConnectionPool(cfg)

	// Bağlantının gerçekten çalışıp çalışmadığını kontrol eder (retry mekanizması)
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		err = DB.Ping()
		if err == nil {
			break
		}
		// Bekleme süresi kademeli olarak artar (2s, 4s, 6s, ...)
		waitTime := time.Duration((i+1)*2) * time.Second
		log.Printf("Veritabanı sunucusuna bağlanılamıyor (%d/%d), %v sonra tekrar denenecek... Hata: %v\n",
			i+1, maxRetries, waitTime, err)
		time.Sleep(waitTime)
	}

	if err != nil {
		log.Fatalf("Veritabanı sunucusuna erişilemiyor (host: %s, port: %s): %v",
			cfg.DBHost, cfg.DBPort, err)
	}

	log.Printf("PostgreSQL veritabanına başarıyla bağlanıldı. (Sunucu: %s:%s, DB: %s)\n",
		cfg.DBHost, cfg.DBPort, cfg.DBName)
}

// configureConnectionPool fonksiyonu, uzak veritabanı sunucusu için bağlantı havuzu parametrelerini ayarlar.
func configureConnectionPool(cfg *configs.Config) {
	// Maksimum açık bağlantı sayısı
	maxOpen, err := strconv.Atoi(cfg.DBMaxOpenConns)
	if err != nil {
		maxOpen = 25
	}
	DB.SetMaxOpenConns(maxOpen)

	// Maksimum boşta bekleyen bağlantı sayısı
	maxIdle, err := strconv.Atoi(cfg.DBMaxIdleConns)
	if err != nil {
		maxIdle = 10
	}
	DB.SetMaxIdleConns(maxIdle)

	// Bağlantının maksimum yaşam süresi (dakika cinsinden)
	lifetimeMin, err := strconv.Atoi(cfg.DBConnMaxLifetimeMin)
	if err != nil {
		lifetimeMin = 5
	}
	DB.SetConnMaxLifetime(time.Duration(lifetimeMin) * time.Minute)

	log.Printf("DB bağlantı havuzu yapılandırıldı: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%dm\n",
		maxOpen, maxIdle, lifetimeMin)
}
