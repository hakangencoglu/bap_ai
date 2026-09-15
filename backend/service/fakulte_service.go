package service

import (
	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// FakulteService fakülte ve bölüm iş mantıklarını yürütür.
type FakulteService struct {
	Repo *repository.FakulteRepository
}

// NewFakulteService yeni bir FakulteService oluşturur.
func NewFakulteService(repo *repository.FakulteRepository) *FakulteService {
	return &FakulteService{Repo: repo}
}

// GetFakulteler tüm aktif fakülteleri getirir.
// Türkçe Yorum: Arayüzdeki fakülte açılır listesini besler.
func (s *FakulteService) GetFakulteler() ([]models.Fakulte, error) {
	return s.Repo.GetFakulteler()
}

// GetBolumler seçili fakülteye ait veya tüm aktif bölümleri getirir.
// Türkçe Yorum: Seçilen fakülteye göre dinamik olarak bölümleri listeler.
func (s *FakulteService) GetBolumler(fakulteID ...int) ([]models.Bolum, error) {
	return s.Repo.GetBolumler(fakulteID...)
}
