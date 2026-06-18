package service

import (
	"fmt"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// HakemService, hakem işlemlerine ait iş mantığını yönetir
type HakemService struct {
	HakemRepo *repository.HakemRepository
	ProjeRepo *repository.ProjeRepository
	AdminRepo *repository.AdminRepository
}

func NewHakemService(hakemRepo *repository.HakemRepository, projeRepo *repository.ProjeRepository, adminRepo *repository.AdminRepository) *HakemService {
	return &HakemService{
		HakemRepo: hakemRepo,
		ProjeRepo: projeRepo,
		AdminRepo: adminRepo,
	}
}

// GetProjelerByHakem, bir hakeme ait atanan projeleri döndürür
func (s *HakemService) GetProjelerByHakem(hakemID int) ([]models.HakemProjeOzet, error) {
	return s.HakemRepo.GetProjelerByHakemID(hakemID)
}

// KabulRedKarar, hakemin atamayı kabul veya reddetme kararını işler
func (s *HakemService) KabulRedKarar(hakemID int, req models.HakemKararRequest) error {
	// Karar değeri doğrulanır
	var atamaDurumu string
	switch req.Karar {
	case "kabul":
		atamaDurumu = "Kabul Edildi"
	case "red":
		atamaDurumu = "Reddedildi"
		// Red durumunda neden gerekli
		if req.RedNedeni == "" {
			return fmt.Errorf("red durumunda red nedeni belirtilmelidir")
		}
	default:
		return fmt.Errorf("geçersiz karar değeri: %s (kabul veya red olmalı)", req.Karar)
	}

	return s.HakemRepo.UpdateAtamaKarar(hakemID, req.ProjeID, atamaDurumu, req.RedNedeni)
}

// SubmitDegerlendirme, puanlamayı kaydeder ve gerekirse genel proje durumunu günceller.
func (s *HakemService) SubmitDegerlendirme(hakemID int, req models.DegerlendirmeRequest) error {
	// Puanı kaydet (sadece atamayı kabul etmiş hakemler değerlendirme yapabilir)
	if err := s.HakemRepo.SubmitDegerlendirme(hakemID, req); err != nil {
		return err
	}

	// Eğer değerlendirme "Revizyon" ise, proje durumunu güncelleyip revizyon kaydı oluştururuz
	// Türkçe Yorum: Hakem revizyon istediğinde projenin durumunu 'revizyon' yaparız ve akademisyen için revizyon bildirimi açarız.
	if req.Durum == "Revizyon" {
		p, err := s.ProjeRepo.GetProjeByID(req.ProjeID)
		if err != nil {
			return fmt.Errorf("proje bulunamadı: %w", err)
		}

		err = s.ProjeRepo.UpdateProjectStatusWithLog(req.ProjeID, hakemID, p.DurumAdi, "revizyon", req.Yorum)
		if err != nil {
			return fmt.Errorf("proje durumu güncellenemedi: %w", err)
		}

		var atananID *int
		if p.KoordinatorID != nil {
			atananID = p.KoordinatorID
		}

		insertQuery := `
			INSERT INTO revizyonlar (proje_id, olusturan_kisi_id, atanan_kisi_id, aciklama, durum, revizyon_bolum)
			VALUES ($1, $2, $3, $4, 'Bekliyor', $5)
		`
		_, err = s.HakemRepo.DB.Exec(insertQuery, req.ProjeID, hakemID, atananID, req.Yorum, req.RevizyonBolum)
		if err != nil {
			return fmt.Errorf("revizyon kaydı oluşturulamadı: %w", err)
		}
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

// IsHakemAssigned, hakemin projeye atanıp atanmadığını sorgular.
// Türkçe Yorum: Hakemin projeyi görme/değerlendirme yetkisi olup olmadığını kontrol eder.
func (s *HakemService) IsHakemAssigned(hakemID int, projeID int) (bool, error) {
	return s.HakemRepo.IsHakemAssigned(hakemID, projeID)
}

// GetProjectDetailsForHakem, hakemin projenin tüm detaylarını görmesini sağlar.
// Türkçe Yorum: Hakem detay sayfası için admin yetkili fonksiyonu üzerinden projenin tüm detaylarını çeker.
func (s *HakemService) GetProjectDetailsForHakem(projeID int) (*repository.ProjectDetail, error) {
	return s.AdminRepo.GetProjectDetailsForAdmin(projeID)
}
