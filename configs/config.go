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

	// LDAP Ayarları
	LDAPEnabled      bool
	LDAPHost         string
	LDAPPort         string
	LDAPBaseDN       string
	LDAPBindDN       string
	LDAPBindPassword string
	LDAPUserFilter   string
	LDAPMock         bool

	// Yapay Zeka (LLM) Ayarları
	LLMProvider      string
	LLMEndpoint      string
	LLMModel         string
	GeminiAPIKey     string

	// E-posta (SMTP) Ayarları
	SMTPEnabled  bool
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
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

		// LDAP Yapılandırmaları
		LDAPEnabled:      getEnv("LDAP_ENABLED", "false") == "true",
		LDAPHost:         getEnv("LDAP_HOST", "localhost"),
		LDAPPort:         getEnv("LDAP_PORT", "389"),
		LDAPBaseDN:       getEnv("LDAP_BASE_DN", "dc=izu,dc=edu,dc=tr"),
		LDAPBindDN:       getEnv("LDAP_BIND_DN", "cn=admin,dc=izu,dc=edu,dc=tr"),
		LDAPBindPassword: getEnv("LDAP_BIND_PASSWORD", "admin123"),
		LDAPUserFilter:   getEnv("LDAP_USER_FILTER", "(&(objectClass=user)(sAMAccountName=%s))"),
		LDAPMock:         getEnv("LDAP_MOCK", "true") == "true",

		// Yapay Zeka (LLM) Yapılandırmaları
		LLMProvider:      getEnv("LLM_PROVIDER", "mock"),
		LLMEndpoint:      getEnv("LLM_ENDPOINT", "http://localhost:11434"),
		LLMModel:         getEnv("LLM_MODEL", "llama3"),
		GeminiAPIKey:     getEnv("GEMINI_API_KEY", ""),

		// E-posta (SMTP) Yapılandırmaları
		SMTPEnabled:  getEnv("SMTP_ENABLED", "false") == "true",
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),
	}
}

// getEnv fonksiyonu, belirtilen çevre değişkenini okur, yoksa varsayılan değeri döner.
func getEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
