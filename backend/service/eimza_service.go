package service

import (
	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// EimzaService yapısı, e-imza işlemlerinin iş mantığı katmanıdır.
// Türkçe Yorum: E-İmza süreçlerinin iş mantığını yürüten servis yapısı.
type EimzaService struct {
	EimzaRepo *repository.EimzaRepository
}

// NewEimzaService yeni bir EimzaService oluşturur.
// Türkçe Yorum: EimzaService nesnesini başlatan kurucu fonksiyon.
func NewEimzaService(eimzaRepo *repository.EimzaRepository) *EimzaService {
	return &EimzaService{EimzaRepo: eimzaRepo}
}

// GetPendingSignatures kullanıcı rolüne göre imza bekleyen projeleri getirir.
// Türkçe Yorum: Kullanıcının rolüne göre imzalayabileceği bekleyen projeleri çeker.
func (s *EimzaService) GetPendingSignatures(uyeID int, rol string) ([]models.Proje, error) {
	return s.EimzaRepo.GetPendingSignatures(uyeID, rol)
}

// GetSignedDocuments kullanıcının daha önce imzaladığı projeleri getirir.
// Türkçe Yorum: Kullanıcı tarafından daha önce imzalanmış belgeleri/projeleri çeker.
func (s *EimzaService) GetSignedDocuments(uyeID int) ([]models.Proje, error) {
	return s.EimzaRepo.GetSignedDocuments(uyeID)
}

// SignDocument projenin e-imza ile imzalanması sürecini tetikler.
// Türkçe Yorum: E-İmza atılması sürecini başlatıp durumu güncelleyen servis fonksiyonu.
func (s *EimzaService) SignDocument(projeID, uyeID int, rol, imzaciAdSoyad, yontem, pin, ip, userAgent string) error {
	return s.EimzaRepo.SignDocument(projeID, uyeID, rol, imzaciAdSoyad, yontem, pin, ip, userAgent)
}
