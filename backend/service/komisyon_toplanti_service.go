package service

import (
	"fmt"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// KomisyonToplantiService toplantı ↔ proje ilişkilendirme süreçlerini yönetir.
// Türkçe Yorum: Proje ekleme, çıkarma ve karar kaydı iş kurallarını uygular.
type KomisyonToplantiService struct {
	ToplantiRepo *repository.KomisyonToplantiRepository
}

// NewKomisyonToplantiService yeni bir KomisyonToplantiService oluşturur.
func NewKomisyonToplantiService(toplantiRepo *repository.KomisyonToplantiRepository) *KomisyonToplantiService {
	return &KomisyonToplantiService{ToplantiRepo: toplantiRepo}
}

// AddProjeToToplanti bir projeyi toplantı gündemine ekler.
// Türkçe Yorum: komisyon_bekliyor durumundaki projeler raportör tarafından toplantıya eklenir.
func (s *KomisyonToplantiService) AddProjeToToplanti(toplantiID, projeID, gundemSirasi, ekleyenID int) error {
	if toplantiID <= 0 || projeID <= 0 {
		return fmt.Errorf("geçersiz toplantı veya proje ID'si")
	}
	return s.ToplantiRepo.AddProjeToToplanti(toplantiID, projeID, gundemSirasi, ekleyenID)
}

// RemoveProjeFromToplanti bir projeyi toplantı gündeminden çıkarır.
func (s *KomisyonToplantiService) RemoveProjeFromToplanti(toplantiID, projeID int) error {
	return s.ToplantiRepo.RemoveProjeFromToplanti(toplantiID, projeID)
}

// GetProjectsByToplanti bir toplantıdaki projeleri listeler.
func (s *KomisyonToplantiService) GetProjectsByToplanti(toplantiID int) ([]*models.KomisyonToplantisiProje, error) {
	return s.ToplantiRepo.GetProjectsByToplanti(toplantiID)
}

// GetToplantilerByProje bir projenin görüşüldüğü tüm toplantıları listeler.
func (s *KomisyonToplantiService) GetToplantilerByProje(projeID int) ([]*models.KomisyonToplantisiProje, error) {
	return s.ToplantiRepo.GetToplantilerByProje(projeID)
}

// SetProjeKarar toplantıdaki bir proje için karar kaydeder.
// Türkçe Yorum: Geçerli kararlar: bekliyor | onaylandi | reddedildi | ertelendi
func (s *KomisyonToplantiService) SetProjeKarar(toplantiID, projeID int, karar, aciklama string) error {
	gecerliKararlar := map[string]bool{
		"bekliyor": true, "onaylandi": true, "reddedildi": true, "ertelendi": true,
	}
	if !gecerliKararlar[karar] {
		return fmt.Errorf("geçersiz karar: %s (bekliyor|onaylandi|reddedildi|ertelendi olmalı)", karar)
	}
	return s.ToplantiRepo.SetProjeKarar(toplantiID, projeID, karar, aciklama)
}

// GetToplantiBelgeDetay PDF tutanağı için toplantı bütününü döner.
func (s *KomisyonToplantiService) GetToplantiBelgeDetay(toplantiID int) (*models.KomisyonToplantiBelge, error) {
	return s.ToplantiRepo.GetToplantiBelgeDetay(toplantiID)
}
