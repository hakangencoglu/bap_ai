package tests

import (
	"testing"
	"bap_ai/backend/database"
	"bap_ai/configs"
)

// TestDatabaseConnection veritabanı bağlantısının bir bütün olarak çalışıp çalışmadığını doğrular.
func TestDatabaseConnection(t *testing.T) {
	// Konfigürasyonu yükle
	configs.LoadConfig()

	// Bağlantıyı başlat
	database.Connect()

	// DB nesnesinin aktif olduğunu kontrol et
	if database.DB == nil {
		t.Fatal("Hata: Veritabanı bağlantısı (database.DB) başlatılamadı.")
	}

	// Ping atarak fiziksel bağlantıyı doğrula
	err := database.DB.Ping()
	if err != nil {
		t.Fatalf("Hata: Veritabanına ping atılamadı: %v", err)
	}

	t.Log("Başarılı: Veritabanı bağlantısı başarıyla sağlandı.")
}
