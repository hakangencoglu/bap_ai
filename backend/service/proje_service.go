package service

import (
	"fmt"
	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// ProjeService yapısı proje iş mantığını barındırır.
type ProjeService struct {
	ProjeRepo *repository.ProjeRepository
}

// NewProjeService fonksiyonu yeni bir ProjeService oluşturur.
func NewProjeService(projeRepo *repository.ProjeRepository) *ProjeService {
	return &ProjeService{ProjeRepo: projeRepo}
}

// CreateProje veritabanına bir proje ekler ve üye-proje bağlantısını sağlar.
func (s *ProjeService) CreateProje(uyeID int, p *models.Proje) error {
	err := s.ProjeRepo.CreateProje(uyeID, p)
	if err != nil {
		return fmt.Errorf("proje oluşturulurken bir hata meydana geldi: %w", err)
	}
	return nil
}
