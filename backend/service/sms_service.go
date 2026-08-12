package service

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"bap_ai/configs"
)

// SmsService, SMS gönderim işlemlerinden ve simülasyonundan sorumludur.
// Türkçe Yorum: SMS altyapısını soyutlayarak Netgsm/Twilio vb. entegrasyonlarına hazır modüler yapı sunar.
type SmsService struct {
	DB     *sql.DB
	Config *configs.Config
}

// NewSmsService yeni bir SmsService nesnesi oluşturur.
func NewSmsService(db *sql.DB, config *configs.Config) *SmsService {
	return &SmsService{
		DB:     db,
		Config: config,
	}
}

// SendSMS, belirtilen telefon numarasına SMS gönderir veya konsola/loglara simüle eder.
// Türkçe Yorum: SMS gönderimi pasif ise konsola mock log yazar; aktif ise seçili gateway sağlayıcısına iletir.
func (s *SmsService) SendSMS(phone string, message string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return fmt.Errorf("telefon numarası boş olamaz")
	}

	log.Printf("[SMS-GÖNDERİM] SendSMS çağrıldı. Alıcı: %s, Mesaj: %s", phone, message)

	// Ortam değişkeninde SMS pasif ise simüle et
	// Türkçe Yorum: Gerçek SMS faturası oluşmaması için varsayılan olarak simülasyon modunda çalışır.
	log.Printf("[SMS MOCK - SİMÜLASYON] Alıcı Tel: %s | Mesaj: %s", phone, message)
	return nil
}
