package service

import (
	"fmt"
	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// ProjeService yapısı proje iş mantığını barındırır.
type ProjeService struct {
	ProjeRepo    *repository.ProjeRepository
	HakemRepo    *repository.HakemRepository
	RevizyonRepo *repository.RevizyonRepository
}

// NewProjeService fonksiyonu yeni bir ProjeService oluşturur.
func NewProjeService(projeRepo *repository.ProjeRepository, hakemRepo *repository.HakemRepository, revizyonRepo *repository.RevizyonRepository) *ProjeService {
	return &ProjeService{
		ProjeRepo:    projeRepo,
		HakemRepo:    hakemRepo,
		RevizyonRepo: revizyonRepo,
	}
}

// CreateProje veritabanına bir proje ekler ve üye-proje bağlantısını sağlar.
// uyeRol parametresi ile öğrenci/akademisyen rolüne göre proje rolü belirlenir.
// Ayrıca otomatik hakem atar.
func (s *ProjeService) CreateProje(uyeID int, p *models.Proje, uyeRol string) error {
	err := s.ProjeRepo.CreateProje(uyeID, p, uyeRol)
	if err != nil {
		return fmt.Errorf("proje oluşturulurken bir hata meydana geldi: %w", err)
	}

	// Proje başarıyla oluşturulduysa, rastgele 2 hakem ata
	if s.HakemRepo != nil {
		s.HakemRepo.AssignRandomHakem(p.ProjeID, 2)
		// Not: Atama başarısız olsa bile (örneğin hakem yoksa) projeyi hata ile bölmemek adına hatayı yutabilir veya loglayabiliriz.
	}

	return nil
}

// GetProjeByID
func (s *ProjeService) GetProjeByID(projeID int) (*models.Proje, error) {
	return s.ProjeRepo.GetProjeByID(projeID)
}

// UpdateProje
func (s *ProjeService) UpdateProje(p *models.Proje) error {
	err := s.ProjeRepo.UpdateProje(p)
	if err == nil && s.RevizyonRepo != nil {
		s.RevizyonRepo.MarkRevizyonAsDone(p.ProjeID)
	}
	return err
}
