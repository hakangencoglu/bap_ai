package service

import (
	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

type RevizyonService struct {
	RevizyonRepo *repository.RevizyonRepository
}

func NewRevizyonService(revizyonRepo *repository.RevizyonRepository) *RevizyonService {
	return &RevizyonService{RevizyonRepo: revizyonRepo}
}

// CreateRevizyon servisi
func (s *RevizyonService) CreateRevizyon(rev *models.Revizyon) error {
	return s.RevizyonRepo.CreateRevizyon(rev)
}

func (s *RevizyonService) GetAktifRevizyon(projeID int) (*models.Revizyon, error) {
	return s.RevizyonRepo.GetAktifRevizyon(projeID)
}

func (s *RevizyonService) MarkRevizyonAsDone(projeID int) error {
	return s.RevizyonRepo.MarkRevizyonAsDone(projeID)
}
