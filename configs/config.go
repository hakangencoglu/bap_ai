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
		DBName:    getEnv("DB_NAME", "bap_ai"),
		JWTSecret: getEnv("JWT_SECRET", "super-secret-key"),
	}
}

// getEnv fonksiyonu, belirtilen çevre değişkenini okur, yoksa varsayılan değeri döner.
func getEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
