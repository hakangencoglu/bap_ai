package service

import (
	"fmt"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// ProfilService yapısı, profil sayfası iş mantığını barındırır.
type ProfilService struct {
	UyeRepo   *repository.UyeRepository
	ProjeRepo *repository.ProjeRepository
}

// NewProfilService fonksiyonu, yeni bir ProfilService nesnesi döner.
func NewProfilService(uyeRepo *repository.UyeRepository, projeRepo *repository.ProjeRepository) *ProfilService {
	return &ProfilService{
		UyeRepo:   uyeRepo,
		ProjeRepo: projeRepo,
	}
}

// GetProfilBilgileri fonksiyonu, belirli bir üyenin profil bilgilerini getirir.
// Giriş bilgileri (şifre vb.) hariç tutulur.
func (s *ProfilService) GetProfilBilgileri(uyeID int) (*models.Uye, error) {
	uye, err := s.UyeRepo.GetUyeByID(uyeID)
	if err != nil {
		return nil, fmt.Errorf("profil bilgileri alınamadı: %w", err)
	}

	return uye, nil
}

// GetProfilProjeleri fonksiyonu, belirli bir üyenin projelerini profil formatında getirir.
// Her proje için ad, tür, durum ve kullanıcının projedeki rolü döner.
func (s *ProfilService) GetProfilProjeleri(uyeID int) ([]models.ProfilProjeBilgisi, error) {
	projeler, err := s.ProjeRepo.GetProjectsByUyeIDForProfil(uyeID)
	if err != nil {
		return nil, fmt.Errorf("profil projeleri alınamadı: %w", err)
	}

	return projeler, nil
}
