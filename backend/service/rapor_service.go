package service

import (
	"fmt"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// RaporService yapısı rapor iş mantığını barındırır.
type RaporService struct {
	RaporRepo     *repository.RaporRepository
	ProjeRepo     *repository.ProjeRepository
	EpostaService *EpostaService
}

// NewRaporService yeni bir RaporService oluşturur.
func NewRaporService(raporRepo *repository.RaporRepository, projeRepo *repository.ProjeRepository, epostaService *EpostaService) *RaporService {
	return &RaporService{
		RaporRepo:     raporRepo,
		ProjeRepo:     projeRepo,
		EpostaService: epostaService,
	}
}

// SubmitAraRapor yürütücünün ara rapor veya sonuç raporu teslimini işler.
// Türkçe Yorum: Yürütücü raporu yükler, süreç geçmişine log düşülür ve TTO yetkililerine anlık bildirim/e-posta gönderilir.
func (s *RaporService) SubmitAraRapor(uyeID int, r *models.ProjeAraRapor) error {
	p, err := s.ProjeRepo.GetProjeByID(r.ProjeID)
	if err != nil || p == nil {
		return fmt.Errorf("proje bulunamadı")
	}

	if r.Baslik == "" {
		if r.RaporTuru == "sonuc_raporu" {
			r.Baslik = "Kesin Sonuç Raporu"
		} else {
			r.Baslik = fmt.Sprintf("%d. Ara Rapor", r.RaporDonemi)
		}
	}

	r.YukleyenUyeID = uyeID
	err = s.RaporRepo.CreateAraRapor(r)
	if err != nil {
		return fmt.Errorf("rapor kaydedilemedi: %w", err)
	}

	// Süreç geçmişine log ekle
	logAciklama := fmt.Sprintf("%s sisteme yüklendi ve değerlendirmeye sunuldu.", r.Baslik)
	s.ProjeRepo.UpdateProjectStatusWithLog(r.ProjeID, uyeID, p.DurumAdi, p.DurumAdi, logAciklama)

	// TTO temsilcilerine anlık E-Posta ve Sistem Bildirimi tetikle
	if s.EpostaService != nil {
		go s.EpostaService.SendStatusNotificationEmail(r.ProjeID, uyeID, p.DurumAdi, p.DurumAdi, fmt.Sprintf("Yürütücü tarafından yeni %s (%s) yüklendi. TTO onay incelemeniz beklenmektedir.", r.Baslik, r.RaporTuru))
	}

	return nil
}

// GetTTORaporTakipMatrisi TTO sorumlusunun tüm yürürlükteki projelerin ara rapor teslim durumlarını (gecikmiş, bekleyen, onaylanan, yaklaşan) izlemesini sağlar.
func (s *RaporService) GetTTORaporTakipMatrisi() ([]models.TTORaporTakipItem, error) {
	return s.RaporRepo.GetTTORaporTakipMatrisi()
}

// GetProjeRaporlari projeye ait tüm raporları döner.
func (s *RaporService) GetProjeRaporlari(projeID int) ([]models.ProjeAraRapor, error) {
	return s.RaporRepo.GetAraRaporlarByProjeID(projeID)
}

// GetBekleyenRaporlar onay bekleyen tüm raporları döner (TTO / Admin / Komisyon).
func (s *RaporService) GetBekleyenRaporlar() ([]models.ProjeAraRapor, error) {
	return s.RaporRepo.GetPendingRaporlar()
}

// DegerlendirRapor TTO veya Komisyon yetkilisinin rapor inceleme kararını işler.
// Türkçe Yorum: Karar onay, revizyon veya red olarak güncellenir. Sonuç raporu onaylanırsa proje tamamlandı yapılır.
func (s *RaporService) DegerlendirRapor(onaylayanUyeID int, req models.RaporDegerlendirmeRequest) error {
	rapor, err := s.RaporRepo.GetAraRaporByID(req.RaporID)
	if err != nil || rapor == nil {
		return fmt.Errorf("rapor bulunamadı")
	}

	err = s.RaporRepo.UpdateRaporStatus(req.RaporID, onaylayanUyeID, req.Durum, req.OnayNotu)
	if err != nil {
		return fmt.Errorf("rapor durumu güncellenemedi: %w", err)
	}

	p, pErr := s.ProjeRepo.GetProjeByID(rapor.ProjeID)
	pDurum := "yururlukte"
	if pErr == nil && p != nil {
		pDurum = p.DurumAdi
	}

	// Süreç geçmişine loglama yap
	aciklama := fmt.Sprintf("%s değerlendirildi: %s.", rapor.Baslik, req.Durum)
	if req.OnayNotu != "" {
		aciklama += " Not: " + req.OnayNotu
	}
	s.ProjeRepo.UpdateProjectStatusWithLog(rapor.ProjeID, onaylayanUyeID, pDurum, pDurum, aciklama)

	// Eğer kesin sonuç raporu onaylandıysa, proje durumunu 'tamamlandi' olarak güncelle
	if rapor.RaporTuru == "sonuc_raporu" && req.Durum == "onaylandi" {
		s.ProjeRepo.UpdateProjectStatusWithLog(rapor.ProjeID, onaylayanUyeID, pDurum, "tamamlandi", "Kesin Sonuç Raporu onaylandı. Proje başarıyla tamamlandı.")
	}

	return nil
}
