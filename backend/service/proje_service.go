package service

import (
	"fmt"
	"strings"

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

// ProcessWorkflowAction onay mekanizmasındaki kararları işler (TTO, Dekan, Komisyon, Hakem kararları)
// Yeni iş akışı: incelemede (TTO) → dekan_onayi_bekliyor → komisyon_bekliyor → hakem_bekliyor → sozlesme_imza → yururlukte
// ve durum geçişlerini loglayarak gerçekleştirir.
func (s *ProjeService) ProcessWorkflowAction(projeID int, islemYapanID int, action string, aciklama string) error {
	p, err := s.ProjeRepo.GetProjeByID(projeID)
	if err != nil {
		return fmt.Errorf("proje bulunamadı: %w", err)
	}

	var yeniDurum string
	switch p.DurumAdi {
	case "incelemede":
		// Türkçe Yorum: TTO yetkilisi ön inceleme aşamasındaki (incelemede) bir projeyi onaylarsa dekan onayına gönderir.
		switch action {
		case "onayla":
			yeniDurum = "dekan_onayi_bekliyor"
		case "reddet":
			yeniDurum = "reddedildi"
		case "revizyon":
			yeniDurum = "revizyon"
		default:
			return fmt.Errorf("geçersiz işlem: %s", action)
		}
	case "dekan_onayi_bekliyor":
		// Türkçe Yorum: Dekan onaylarsa durum dekan_onayladi olur (TTO ekranına düşer).
		switch action {
		case "onayla":
			yeniDurum = "dekan_onayladi"
		case "reddet":
			yeniDurum = "reddedildi"
		case "revizyon":
			yeniDurum = "revizyon"
		default:
			return fmt.Errorf("geçersiz işlem: %s", action)
		}
	case "dekan_onayladi":
		// Türkçe Yorum: TTO yetkilisi dekanın onayladığı projeyi komisyon onayına sevk eder.
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
		// Türkçe Yorum: Komisyon üyesi karar verdiğinde bu karar 'proje_komisyon_onay' tablosuna yazılır.
		// Eğer işlem yapan rol admin ise doğrudan tüm süreci karara bağlayabilir (geriye dönük uyumluluk/kolaylık için).
		// Değilse, sadece kendi oyunu günceller ve tüm komisyon onaylarının tamamlanıp tamamlanmadığını kontrol eder.
		
		if action != "onayla" && action != "reddet" && action != "revizyon" {
			return fmt.Errorf("geçersiz işlem: %s", action)
		}

		// Komisyon üyesinin bireysel kararını kaydet
		err = s.ProjeRepo.UpdateKomisyonKarar(projeID, islemYapanID, action, aciklama)
		if err != nil {
			return fmt.Errorf("komisyon kararı kaydedilemedi: %w", err)
		}

		// Tüm komisyon üyelerinin kararlarını kontrol et
		allApproved, criticalKarar, err := s.ProjeRepo.CheckAllKomisyonApproved(projeID)
		if err != nil {
			return fmt.Errorf("komisyon onayları kontrol edilemedi: %w", err)
		}

		if criticalKarar == "reddet" {
			yeniDurum = "reddedildi"
		} else if criticalKarar == "revizyon" {
			yeniDurum = "revizyon"
		} else if allApproved {
			yeniDurum = "komisyon_onayladi"
		} else {
			yeniDurum = "komisyon_bekliyor"
		}
	case "komisyon_onayladi":
		// Türkçe Yorum: TTO yetkilisi komisyonun onayladığı projeyi hakem atamaya sevk eder veya doğrudan sözleşmeye gönderir.
		// BAP türünün hakem gerektirip gerektirmediğini kontrol ediyoruz.
		var hakemGerekli bool
		err = s.ProjeRepo.DB.QueryRow(`
			SELECT COALESCE(pbt.hakem_gerekli, false)
			FROM proje p
			JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
			WHERE p.proje_id = $1
		`, projeID).Scan(&hakemGerekli)
		if err != nil {
			hakemGerekli = true // hata durumunda varsayılan güvenli
		}

		if action == "onayla" && !hakemGerekli {
			action = "onayla_hakemsiz"
		}

		switch action {
		case "onayla":
			yeniDurum = "hakem_atama_bekliyor"
		case "onayla_hakemsiz":
			yeniDurum = "sozlesme_imza"
		case "reddet":
			yeniDurum = "reddedildi"
		case "revizyon":
			yeniDurum = "revizyon"
		default:
			return fmt.Errorf("geçersiz işlem: %s", action)
		}
	case "hakem_bekliyor":
		// Türkçe Yorum: Hakem onaylayınca durum hakem_onayladi olur (TTO ekranına düşer).
		switch action {
		case "onayla":
			yeniDurum = "hakem_onayladi"
		case "reddet":
			yeniDurum = "reddedildi"
		case "revizyon":
			yeniDurum = "revizyon"
		default:
			return fmt.Errorf("geçersiz işlem: %s", action)
		}
	case "hakem_onayladi":
		// Türkçe Yorum: TTO yetkilisi hakemin onayladığı projeyi sözleşme aşamasına sevk eder.
		switch action {
		case "onayla":
			yeniDurum = "sozlesme_imza"
		case "reddet":
			yeniDurum = "reddedildi"
		case "revizyon":
			yeniDurum = "revizyon"
		default:
			return fmt.Errorf("geçersiz işlem: %s", action)
		}
	case "sozlesme_imza":
		// Türkçe Yorum: Sözleşme imzalandıktan sonra proje yürürlükte durumuna geçer.
		switch action {
		case "tamamla", "onayla":
			yeniDurum = "yururlukte"
		case "reddet":
			yeniDurum = "reddedildi"
		case "revizyon":
			yeniDurum = "revizyon"
		default:
			return fmt.Errorf("geçersiz işlem: %s", action)
		}
	// Türkçe Yorum: Eski tto_aktif durumu geriye dönük uyumluluk için korunuyor.
	case "tto_aktif":
		switch action {
		case "tamamla", "onayla":
			yeniDurum = "yururlukte"
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

	// Türkçe Yorum: Eğer hedef durum komisyon_bekliyor ise ve proje bu duruma yeni geçiyorsa komisyon oylama kayıtları açılır.
	if yeniDurum == "komisyon_bekliyor" && p.DurumAdi != "komisyon_bekliyor" {
		if err := s.ProjeRepo.CreateKomisyonOnayRecords(projeID); err != nil {
			return fmt.Errorf("komisyon oylama kayıtları oluşturulamadı: %w", err)
		}
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

// GetWorkflowHistoryByUyeID belirli bir kullanıcının geçmiş onay kararlarını çeker.
func (s *ProjeService) GetWorkflowHistoryByUyeID(uyeID int) ([]models.ProjeSurecGecmisi, error) {
	return s.ProjeRepo.GetWorkflowHistoryByUyeID(uyeID)
}

// GetProjectsForWorkflow rol bazında onay bekleyen projeleri listeler.
// Türkçe Yorum: Kullanıcının sahip olduğu tüm rollere (virgülle ayrılmış olabilir) göre onay bekleyen projeleri çeker ve tekil olarak birleştirir.
// Yeni iş akışı: TTO = ön inceleme (incelemede) + sözleşme (sozlesme_imza), Hakem = hakem_bekliyor
func (s *ProjeService) GetProjectsForWorkflow(rol string, uyeID int) ([]models.Proje, error) {
	roles := strings.Split(rol, ",")
	var allProjects []models.Proje
	seen := make(map[int]bool)
	hasWorkflowRole := false

	for _, r := range roles {
		r = strings.TrimSpace(r)
		
		// Admin ise süreçteki tüm onay bekleyen projeleri görsün
		if r == "admin" {
			hasWorkflowRole = true
			for _, d := range []string{"incelemede", "dekan_onayi_bekliyor", "dekan_onayladi", "komisyon_bekliyor", "komisyon_onayladi", "hakem_bekliyor", "hakem_onayladi", "sozlesme_imza", "tto_aktif"} {
				projeler, err := s.ProjeRepo.GetProjectsForWorkflow(r, d, 0)
				if err != nil {
					return nil, err
				}
				for _, p := range projeler {
					if !seen[p.ProjeID] {
						seen[p.ProjeID] = true
						allProjects = append(allProjects, p)
					}
				}
			}
			continue
		}

		// Türkçe Yorum: TTO rolü için süreçteki tüm taslak olmayan projeleri takip amaçlı gösterir.
		if r == "tto" {
			hasWorkflowRole = true
			durumlar := []string{
				"incelemede",
				"dekan_onayi_bekliyor",
				"dekan_onayladi",
				"komisyon_bekliyor",
				"komisyon_onayladi",
				"hakem_atama_bekliyor",
				"hakem_bekliyor",
				"hakem_onayladi",
				"sozlesme_imza",
				"tto_aktif",
				"yururlukte",
				"reddedildi",
				"revizyon",
				"tamamlandi",
			}
			for _, d := range durumlar {
				projeler, err := s.ProjeRepo.GetProjectsForWorkflow(r, d, 0)
				if err != nil {
					return nil, err
				}
				for _, p := range projeler {
					if !seen[p.ProjeID] {
						seen[p.ProjeID] = true
						allProjects = append(allProjects, p)
					}
				}
			}
			continue
		}

		// Türkçe Yorum: Hakem rolü için hakem_bekliyor durumundaki projeleri gösterir.
		if r == "hakem" {
			hasWorkflowRole = true
			projeler, err := s.ProjeRepo.GetProjectsForWorkflow(r, "hakem_bekliyor", 0)
			if err != nil {
				return nil, err
			}
			for _, p := range projeler {
				if !seen[p.ProjeID] {
					seen[p.ProjeID] = true
					allProjects = append(allProjects, p)
				}
			}
			continue
		}

		var durum string
		switch r {
		case "dekan":
			durum = "dekan_onayi_bekliyor"
		case "komisyon":
			durum = "komisyon_bekliyor"
		default:
			continue
		}

		hasWorkflowRole = true
		projeler, err := s.ProjeRepo.GetProjectsForWorkflow(r, durum, uyeID)
		if err != nil {
			return nil, err
		}
		for _, p := range projeler {
			if !seen[p.ProjeID] {
				seen[p.ProjeID] = true
				allProjects = append(allProjects, p)
			}
		}
	}

	if !hasWorkflowRole {
		return nil, fmt.Errorf("onay akışı için yetkili rol bulunamadı: %s", rol)
	}

	return allProjects, nil
}

