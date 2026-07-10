package service

import (
	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// BildirimService bildirim iş mantığını yönetir.
// Türkçe Yorum: Bildirimlerin iş mantığını yürüten ve repository katmanına erişen servis sınıfıdır.
type BildirimService struct {
	BildirimRepo *repository.BildirimRepository
}

// NewBildirimService yeni bir BildirimService oluşturur.
// Türkçe Yorum: BildirimService için dependency injection kurucusu.
func NewBildirimService(bildirimRepo *repository.BildirimRepository) *BildirimService {
	return &BildirimService{BildirimRepo: bildirimRepo}
}

// GetBildirimler kullanıcının bildirimlerini getirir.
// Türkçe Yorum: Kullanıcının bildirimlerini almak için repository katmanını tetikler.
func (s *BildirimService) GetBildirimler(uyeID int) ([]models.Bildirim, error) {
	return s.BildirimRepo.GetBildirimlerByUyeID(uyeID)
}

// MarkAsRead bildirimi okundu olarak işaretler.
// Türkçe Yorum: Belirli bir bildirimi okundu yapmak için repository katmanını tetikler.
func (s *BildirimService) MarkAsRead(bildirimID int, uyeID int) error {
	return s.BildirimRepo.MarkAsRead(bildirimID, uyeID)
}

// MarkAllAsRead tüm bildirimleri okundu olarak işaretler.
// Türkçe Yorum: Tüm bildirimleri okundu yapmak için repository katmanını tetikler.
func (s *BildirimService) MarkAllAsRead(uyeID int) error {
	return s.BildirimRepo.MarkAllAsRead(uyeID)
}
