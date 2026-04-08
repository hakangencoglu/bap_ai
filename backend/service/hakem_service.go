package service

import (
	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// HakemService, hakem işlemlerine ait iş mantığını yönetir
type HakemService struct {
	HakemRepo *repository.HakemRepository
	ProjeRepo *repository.ProjeRepository
}

func NewHakemService(hakemRepo *repository.HakemRepository, projeRepo *repository.ProjeRepository) *HakemService {
	return &HakemService{
		HakemRepo: hakemRepo,
		ProjeRepo: projeRepo,
	}
}

// GetProjelerByHakem, bir hakeme ait atanan projeleri döndürür
func (s *HakemService) GetProjelerByHakem(hakemID int) ([]models.HakemProjeOzet, error) {
	return s.HakemRepo.GetProjelerByHakemID(hakemID)
}

// SubmitDegerlendirme, puanlamayı kaydeder ve gerekirse genel proje durumunu günceller.
func (s *HakemService) SubmitDegerlendirme(hakemID int, req models.DegerlendirmeRequest) error {
	// Puanı kaydet
	if err := s.HakemRepo.SubmitDegerlendirme(hakemID, req); err != nil {
		return err
	}

	// Tüm değerlendirmeleri al
	degerlendirmeler, err := s.HakemRepo.GetAllDegerlendirmeByProjeID(req.ProjeID)
	if err != nil {
		return nil // Hata loglanabilir ama asıl puanlama kaydedildi
	}

	// Basit karar mekanizması:
	// Eğer projeye atanan hiçbir hakemin durumu 'Bekliyor' değilse, proje tamamlanmış demektir.
	hepsiTamam := true
	for _, d := range degerlendirmeler {
		if d.Durum == "Bekliyor" {
			hepsiTamam = false
			break
		}
	}

	// Eğer tüm hakemler değerlendirmesini yaptıysa ve proje durumu da değişmeliyse, puan ortalaması vs alınabilir.
	// Şimdilik sadece projeyi "degerlendirildi" veya "sonuclandi" aşamasına getirebiliriz.
	// Not: Admin onayı da eklenecekse burası opsiyoneldir.
	_ = hepsiTamam // Sonraki aşamada eklenecek

	return nil
}
