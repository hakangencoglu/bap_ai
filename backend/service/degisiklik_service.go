package service

import (
	"errors"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

var ErrDegisiklikProjeUyusmazligi = errors.New("değişiklik kaydı bu projeye ait değil")

// DegisiklikService, proje değişiklik geçmişi okuma iş mantığını yönetir.
type DegisiklikService struct {
	Repo *repository.DegisiklikRepository
}

// NewDegisiklikService yeni servis oluşturur.
func NewDegisiklikService(repo *repository.DegisiklikRepository) *DegisiklikService {
	return &DegisiklikService{Repo: repo}
}

// ListByProje, proje değişiklik listesini döner.
func (s *DegisiklikService) ListByProje(projeID int, olayTipi, kaynakTip string) ([]models.ProjeDegisiklik, error) {
	return s.Repo.ListByProje(projeID, olayTipi, kaynakTip)
}

// GetByID, değişiklik detayını döner; proje eşleşmesi doğrulanır.
func (s *DegisiklikService) GetByID(projeID, degisiklikID int) (*models.ProjeDegisiklik, error) {
	item, err := s.Repo.GetByID(degisiklikID)
	if err != nil {
		return nil, err
	}
	if item.ProjeID != projeID {
		return nil, ErrDegisiklikProjeUyusmazligi
	}
	return item, nil
}
