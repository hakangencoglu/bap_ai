package service

import (
	"fmt"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// DashboardService yapısı, dashboard iş mantığını barındırır.
type DashboardService struct {
	ProjeRepo *repository.ProjeRepository
}

// NewDashboardService fonksiyonu, yeni bir DashboardService nesnesi döner.
func NewDashboardService(projeRepo *repository.ProjeRepository) *DashboardService {
	return &DashboardService{ProjeRepo: projeRepo}
}

// GetStats fonksiyonu, belirli bir üyenin dashboard istatistiklerini getirir.
func (s *DashboardService) GetStats(uyeID int) (*models.DashboardStats, error) {
	stats, err := s.ProjeRepo.GetDashboardStatsByUyeID(uyeID)
	if err != nil {
		return nil, fmt.Errorf("dashboard istatistikleri alınamadı: %w", err)
	}

	return stats, nil
}

// GetRecentProjects fonksiyonu, belirli bir üyenin son başvurularını getirir.
func (s *DashboardService) GetRecentProjects(uyeID int) ([]models.ProjeOzet, error) {
	projeler, err := s.ProjeRepo.GetRecentProjectsByUyeID(uyeID)
	if err != nil {
		return nil, fmt.Errorf("son başvurular alınamadı: %w", err)
	}

	return projeler, nil
}

// GetAllProjects fonksiyonu, belirli bir üyenin kabul ettiği tüm başvurularını getirir.
func (s *DashboardService) GetAllProjects(uyeID int) ([]models.ProjeOzet, error) {
	projeler, err := s.ProjeRepo.GetAllProjectsByUyeID(uyeID)
	if err != nil {
		return nil, fmt.Errorf("tüm başvurular alınamadı: %w", err)
	}

	return projeler, nil
}
