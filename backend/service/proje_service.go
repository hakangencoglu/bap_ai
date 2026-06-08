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

// DeleteTaslakProje taslak durumundaki bir projeyi siler.
// Yetki ve durum kontrolü repository katmanında yapılır.
func (s *ProjeService) DeleteTaslakProje(projeID int, uyeID int) error {
	return s.ProjeRepo.DeleteTaslakProje(projeID, uyeID)
}

// ProcessWorkflowAction onay mekanizmasındaki kararları işler (Dekan, Komisyon, TTO kararları)
// ve durum geçişlerini loglayarak gerçekleştirir.
func (s *ProjeService) ProcessWorkflowAction(projeID int, islemYapanID int, action string, aciklama string) error {
	p, err := s.ProjeRepo.GetProjeByID(projeID)
	if err != nil {
		return fmt.Errorf("proje bulunamadı: %w", err)
	}

	var yeniDurum string
	switch p.DurumAdi {
	case "dekan_onayi_bekliyor":
		switch action {
		case "onayla":
			yeniDurum = "komisyon_bekliyor"
		case "reddet":
			yeniDurum = "reddedildi"
		case "revizyon":
			yeniDurum = "revizyon"
		default:
			return fmt.Errorf("geçersiz işlem: %s", action)
		}
	case "komisyon_bekliyor":
		switch action {
		case "onayla":
			yeniDurum = "tto_aktif"
		case "reddet":
			yeniDurum = "reddedildi"
		case "revizyon":
			yeniDurum = "revizyon"
		default:
			return fmt.Errorf("geçersiz işlem: %s", action)
		}
	case "tto_aktif":
		switch action {
		case "tamamla", "onayla":
			yeniDurum = "tamamlandi"
		case "reddet":
			yeniDurum = "reddedildi"
		case "revizyon":
			yeniDurum = "revizyon"
		default:
			return fmt.Errorf("geçersiz işlem: %s", action)
		}
	default:
		return fmt.Errorf("bu proje durumu için onay süreci işletilemez: %s", p.DurumAdi)
	}

	err = s.ProjeRepo.UpdateProjectStatusWithLog(projeID, islemYapanID, p.DurumAdi, yeniDurum, aciklama)
	if err != nil {
		return fmt.Errorf("proje durum güncellenirken hata oluştu: %w", err)
	}

	return nil
}

// GetProjeSurecGecmisi projenin geçmiş durum değişikliklerini listeler.
func (s *ProjeService) GetProjeSurecGecmisi(projeID int) ([]models.ProjeSurecGecmisi, error) {
	return s.ProjeRepo.GetProjeSurecGecmisi(projeID)
}

// GetProjectsForWorkflow rol bazında onay bekleyen projeleri listeler.
func (s *ProjeService) GetProjectsForWorkflow(rol string) ([]models.Proje, error) {
	var durum string
	switch rol {
	case "dekan":
		durum = "dekan_onayi_bekliyor"
	case "komisyon":
		durum = "komisyon_bekliyor"
	case "tto":
		durum = "tto_aktif"
	default:
		return nil, fmt.Errorf("geçersiz rol: %s", rol)
	}
	return s.ProjeRepo.GetProjectsForWorkflow(rol, durum)
}

