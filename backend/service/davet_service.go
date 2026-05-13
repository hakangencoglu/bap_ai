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
// Eğer kabul eden kişi Yürütücü rolündeyse:
//   - Projenin koordinator_id'si bu kişi olarak güncellenir
//   - Proje durumu taslak(1) → incelemede(2) yapılır
//   - Projeyi oluşturan öğrencinin rolü Araştırmacı(2) olarak güncellenir
func (s *DavetService) RespondDavet(projeID int, uyeID int, kabul bool) error {
	if kabul {
		// Daveti kabul et
		err := s.DavetRepo.UpdateDavetDurumu(projeID, uyeID, "kabul")
		if err != nil {
			return fmt.Errorf("davet kabul hatası: %v", err)
		}

		// Kabul eden kişinin projedeki rolünü kontrol et
		projeRol, err := s.DavetRepo.GetUyeProjeRol(projeID, uyeID)
		if err != nil {
			log.Printf("Proje rol kontrol hatası: %v", err)
		}

		// Eğer kabul eden kişi Yürütücü rolündeyse özel işlemler uygula
		if projeRol == "Yürütücü" {
			// 1. Projenin koordinatör ID'sini bu akademisyen olarak güncelle
			if err := s.DavetRepo.SetYurutucu(projeID, uyeID); err != nil {
				log.Printf("Yürütücü atama hatası: %v", err)
			} else {
				log.Printf("Proje #%d: Akademisyen (ID:%d) yürütücü olarak atandı", projeID, uyeID)
			}

			// 2. Önceki koordinatörün (projeyi oluşturan öğrenci) rolünü Araştırmacı(2) yap
			olusturanID, err := s.DavetRepo.GetProjeOlusturanID(projeID)
			if err == nil && olusturanID != 0 && olusturanID != uyeID {
				// Oluşturanın sistem rolünü kontrol et
				sistemRol, _ := s.DavetRepo.GetUyeSistemRol(olusturanID)
				if sistemRol == "ogrenci" {
					// Öğrencinin proje rolünü Araştırmacı(2) yap
					if err := s.DavetRepo.UpdateTeamMemberRol(projeID, olusturanID, 2); err != nil {
						log.Printf("Öğrenci rol güncelleme hatası: %v", err)
					} else {
						log.Printf("Proje #%d: Öğrenci (ID:%d) rolü Araştırmacı olarak güncellendi", projeID, olusturanID)
					}
				}
			}

			// 3. Proje durumunu taslak(1) → incelemede(2) geçir
			if err := s.DavetRepo.UpdateProjeDurumByID(projeID, 2); err != nil {
				log.Printf("Proje durumu güncelleme hatası: %v", err)
			} else {
				log.Printf("Proje #%d: Yürütücü onayı ile incelemeye alındı", projeID)
			}
		} else {
			// Yürütücü olmayan biri kabul etti — sadece yürütücü onayını kontrol et
			yurutucuOnay, err := s.DavetRepo.CheckYurutucuOnay(projeID)
			if err != nil {
				log.Printf("Yürütücü onay kontrol hatası: %v", err)
			}
			if yurutucuOnay {
				// Proje zaten yürütücü onaylı ise durumu güncellemeye gerek yok
				log.Printf("Proje #%d: Üye kabul etti, yürütücü zaten onaylamış", projeID)
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
