package service

import (
	"fmt"
	"log"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// DavetService, proje davet iş mantığını yönetir
type DavetService struct {
	DavetRepo *repository.DavetRepository
}

// NewDavetService, yeni bir DavetService örneği oluşturur
func NewDavetService(davetRepo *repository.DavetRepository) *DavetService {
	return &DavetService{DavetRepo: davetRepo}
}

// GetBekleyenDavetler, kullanıcının bekleyen davetlerini döner
func (s *DavetService) GetBekleyenDavetler(uyeID int) ([]models.ProjeDavet, error) {
	return s.DavetRepo.GetBekleyenDavetler(uyeID)
}

// RespondDavet, kullanıcının bir proje davetine yanıt vermesini işler
// kabul=true ise davet_durumu "kabul" yapılır, kabul=false ise üye takımdan çıkarılır
// Eğer yanıt veren yürütücüyse ve kabul ettiyse, proje durumu incelemede (2) yapılır
func (s *DavetService) RespondDavet(projeID int, uyeID int, kabul bool) error {
	if kabul {
		// Daveti kabul et
		err := s.DavetRepo.UpdateDavetDurumu(projeID, uyeID, "kabul")
		if err != nil {
			return fmt.Errorf("davet kabul hatası: %v", err)
		}

		// Yürütücü onay kontrolü: yürütücü kabul ettiyse proje incelemeye alınır
		yurutucuOnay, err := s.DavetRepo.CheckYurutucuOnay(projeID)
		if err != nil {
			log.Printf("Yürütücü onay kontrol hatası: %v", err)
		}
		if yurutucuOnay {
			// Proje durumunu taslak(1) → incelemede(2) geçir
			if err := s.DavetRepo.UpdateProjeDurumByID(projeID, 2); err != nil {
				log.Printf("Proje durumu güncelleme hatası: %v", err)
			} else {
				log.Printf("Proje #%d: Yürütücü onayı ile incelemeye alındı", projeID)
			}
		}

		return nil
	}

	// Daveti reddet → takımdan çıkar
	err := s.DavetRepo.RemoveTeamMember(projeID, uyeID)
	if err != nil {
		return fmt.Errorf("davet red hatası: %v", err)
	}

	return nil
}
