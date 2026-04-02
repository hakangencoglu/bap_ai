package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config yapısı, uygulamada kullanılacak tüm çevresel değişkenleri tutar.
type Config struct {
	Port      string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	JWTSecret string

	// Veritabanı bağlantı havuzu ayarları (uzak DB sunucusu için önemli)
	DBMaxOpenConns      string
	DBMaxIdleConns      string
	DBConnMaxLifetimeMin string
}

// AppConfig, uygulamanın genel konfigürasyonunu bellekte tutar.
var AppConfig *Config

// LoadConfig fonksiyonu, .env dosyasını sisteme yükler ve Config struct'ına atar.
func LoadConfig() {
	// İlk olarak proje kök dizinindeki .env dosyasını arar
	err := godotenv.Load()
	if err != nil {
		// Bulunamazsa configs dizinindeki .env dosyasını denemeye çalışır
		err = godotenv.Load("configs/.env")
		if err != nil {
			log.Println("Uyarı: .env dosyası bulunamadı, sistem ortam değişkenleri kullanılacak.")
		}
	}

	AppConfig = &Config{
		Port:      getEnv("PORT", "8080"),
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnv("DB_PORT", "5432"),
		DBUser:    getEnv("DB_USER", "postgres"),
		DBPass:    getEnv("DB_PASSWORD", "postgres"),
		DBName:    getEnv("DB_NAME", "bap_app"),
		JWTSecret: getEnv("JWT_SECRET", "super-secret-key"),

		// Uzak veritabanı sunucusu için bağlantı havuzu yapılandırması
		DBMaxOpenConns:       getEnv("DB_MAX_OPEN_CONNS", "25"),
		DBMaxIdleConns:       getEnv("DB_MAX_IDLE_CONNS", "10"),
		DBConnMaxLifetimeMin: getEnv("DB_CONN_MAX_LIFETIME_MIN", "5"),
	}
}

// getEnv fonksiyonu, belirtilen çevre değişkenini okur, yoksa varsayılan değeri döner.
func getEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
