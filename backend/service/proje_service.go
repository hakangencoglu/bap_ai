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

// GetProjeByID proje ID'sine göre projeyi döner.
func (s *ProjeService) GetProjeByID(projeID int) (*models.Proje, error) {
	return s.ProjeRepo.GetProjeByID(projeID)
}

// UpdateProje mevcut bir projeyi günceller ve varsa bekleyen revizyonu kapatır.
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

// workflowGecisTablo, bir proje durumundan hangi aksiyonla hangi duruma geçileceğini tanımlar.
// Türkçe Yorum: map[mevcutDurum]map[aksiyon]yeniDurum yapısındadır.
// Bu tabloya satır eklemek, yeni bir durum geçişi tanımlamak anlamına gelir.
// Komisyon oylaması ve hakem_gerekli gibi özel mantık ProcessWorkflowAction içinde ayrıca işlenir.
var workflowGecisTablo = map[string]map[string]string{
	// TTO ön inceleme aşaması
	models.DurumIncelemede: {
		models.AksiyonOnayla:   models.DurumDekanOnayiBekliyor,
		models.AksiyonReddet:   models.DurumReddedildi,
		models.AksiyonRevizyon: models.DurumRevizyon,
	},
	// Dekan onay aşaması
	models.DurumDekanOnayiBekliyor: {
		models.AksiyonOnayla:   models.DurumDekanOnayladi,
		models.AksiyonReddet:   models.DurumReddedildi,
		models.AksiyonRevizyon: models.DurumRevizyon,
	},
	// TTO dekan onayını komisyona sevk eder
	models.DurumDekanOnayladi: {
		models.AksiyonOnayla:   models.DurumKomisyonBekliyor,
		models.AksiyonReddet:   models.DurumReddedildi,
		models.AksiyonRevizyon: models.DurumRevizyon,
	},
	// Hakem ataması TTO tarafından yapılır, onaylayınca sözleşmeye geçer
	models.DurumHakemAtamaBekliyor: {
		models.AksiyonOnayla:   models.DurumSozlesmeImza,
		models.AksiyonReddet:   models.DurumReddedildi,
		models.AksiyonRevizyon: models.DurumRevizyon,
	},
	// Hakem değerlendirme aşaması
	models.DurumHakemBekliyor: {
		models.AksiyonOnayla:   models.DurumHakemOnayladi,
		models.AksiyonReddet:   models.DurumReddedildi,
		models.AksiyonRevizyon: models.DurumRevizyon,
	},
	// TTO hakem onayını sözleşmeye sevk eder
	models.DurumHakemOnayladi: {
		models.AksiyonOnayla:   models.DurumSozlesmeImza,
		models.AksiyonReddet:   models.DurumReddedildi,
		models.AksiyonRevizyon: models.DurumRevizyon,
	},
	// Sözleşme imzalama aşaması
	models.DurumSozlesmeImza: {
		models.AksiyonOnayla:   models.DurumYururlukte,
		models.AksiyonTamamla:  models.DurumYururlukte,
		models.AksiyonReddet:   models.DurumReddedildi,
		models.AksiyonRevizyon: models.DurumRevizyon,
	},
	// Geriye dönük uyumluluk: eski tto_aktif durumu
	models.DurumTTOAktif: {
		models.AksiyonOnayla:   models.DurumYururlukte,
		models.AksiyonTamamla:  models.DurumYururlukte,
		models.AksiyonReddet:   models.DurumReddedildi,
		models.AksiyonRevizyon: models.DurumRevizyon,
	},
}

// resolveKomisyonDurum komisyon_bekliyor aşamasındaki özel oylama mantığını işler.
// Türkçe Yorum: Admin/TTO/Başkan tek karar verebilir; normal komisyon üyesi oy kullanır
// ve tüm oylar tamamlanınca sonuç hesaplanır.
func (s *ProjeService) resolveKomisyonDurum(projeID int, islemYapanID int, action string, aciklama string) (string, error) {
	if action != models.AksiyonOnayla && action != models.AksiyonReddet && action != models.AksiyonRevizyon {
		return "", fmt.Errorf("geçersiz komisyon aksiyonu: %s", action)
	}

	// Kullanıcı rollerini sorgula
	var userRoles string
	err := s.ProjeRepo.DB.QueryRow(`
		SELECT COALESCE(rol, '') FROM uye WHERE uye_id = $1
	`, islemYapanID).Scan(&userRoles)
	if err != nil {
		return "", fmt.Errorf("kullanıcı bilgisi alınamadı: %w", err)
	}

	// Admin, TTO veya Komisyon Başkanı nihai kararı tek başına verir
	isYonetici := false
	for _, r := range strings.Split(userRoles, ",") {
		r = strings.TrimSpace(r)
		if r == models.RolAdmin || r == models.RolTTO || r == models.RolKomisyonBaskani {
			isYonetici = true
			break
		}
	}

	if isYonetici {
		// Türkçe Yorum: Yönetici direkt sonucu belirler, bireysel oylama kaydı gerekmez
		switch action {
		case models.AksiyonReddet:
			return models.DurumReddedildi, nil
		case models.AksiyonRevizyon:
			return models.DurumRevizyon, nil
		default:
			return models.DurumKomisyonOnayladi, nil
		}
	}

	// Normal komisyon üyesi: bireysel oyunu kaydet, sonra tüm oyları kontrol et
	if err := s.ProjeRepo.UpdateKomisyonKarar(projeID, islemYapanID, action, aciklama); err != nil {
		return "", fmt.Errorf("komisyon kararı kaydedilemedi: %w", err)
	}

	allApproved, criticalKarar, err := s.ProjeRepo.CheckAllKomisyonApproved(projeID)
	if err != nil {
		return "", fmt.Errorf("komisyon onayları kontrol edilemedi: %w", err)
	}

	switch criticalKarar {
	case models.AksiyonReddet:
		return models.DurumReddedildi, nil
	case models.AksiyonRevizyon:
		return models.DurumRevizyon, nil
	}
	if allApproved {
		return models.DurumKomisyonOnayladi, nil
	}
	// Türkçe Yorum: Henüz tüm üyeler oy kullanmadı, durum değişmez
	return models.DurumKomisyonBekliyor, nil
}

// resolveKomisyonOnayladiDurum komisyon_onayladi aşamasında hakem gereksinimini kontrol eder.
// Türkçe Yorum: BAP türüne göre hakem gerekli ise hakem_atama_bekliyor, değilse sozlesme_imza durumuna geçilir.
func (s *ProjeService) resolveKomisyonOnayladiDurum(projeID int, action string) (string, error) {
	switch action {
	case models.AksiyonReddet:
		return models.DurumReddedildi, nil
	case models.AksiyonRevizyon:
		return models.DurumRevizyon, nil
	case models.AksiyonOnayla:
		// Türkçe Yorum: BAP türünde hakem değerlendirmesi gerekli mi kontrol ediyoruz
		var hakemGerekli bool
		err := s.ProjeRepo.DB.QueryRow(`
			SELECT COALESCE(pbt.hakem_gerekli, false)
			FROM proje p
			JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
			WHERE p.proje_id = $1
		`, projeID).Scan(&hakemGerekli)
		if err != nil {
			// Hata durumunda güvenli varsayılan: hakem gerekli
			hakemGerekli = true
		}
		if hakemGerekli {
			return models.DurumHakemAtamaBekliyor, nil
		}
		return models.DurumSozlesmeImza, nil
	default:
		return "", fmt.Errorf("geçersiz işlem: %s", action)
	}
}

// ProcessWorkflowAction onay mekanizmasındaki kararları işler (TTO, Dekan, Komisyon, Hakem kararları).
// Türkçe Yorum: Durum geçişleri workflowGecisTablo üzerinden çözülür. Komisyon oylaması ve
// hakem zorunluluğu gibi özel mantık ayrı yardımcı fonksiyonlara taşınmıştır.
func (s *ProjeService) ProcessWorkflowAction(projeID int, islemYapanID int, action string, aciklama string) error {
	p, err := s.ProjeRepo.GetProjeByID(projeID)
	if err != nil {
		return fmt.Errorf("proje bulunamadı: %w", err)
	}

	var yeniDurum string

	switch p.DurumAdi {
	case models.DurumKomisyonBekliyor:
		// Türkçe Yorum: Komisyon oylaması özel mantık içerdiği için ayrı fonksiyonda işleniyor
		yeniDurum, err = s.resolveKomisyonDurum(projeID, islemYapanID, action, aciklama)
		if err != nil {
			return err
		}

	case models.DurumKomisyonOnayladi:
		// Türkçe Yorum: Hakem gereksinimi kontrolü için ayrı fonksiyon çağrılıyor
		yeniDurum, err = s.resolveKomisyonOnayladiDurum(projeID, action)
		if err != nil {
			return err
		}

	default:
		// Türkçe Yorum: Diğer tüm durumlar için geçiş tablosu kullanılır
		gecisler, ok := workflowGecisTablo[p.DurumAdi]
		if !ok {
			return fmt.Errorf("bu proje durumu için onay süreci işletilemez: %s", p.DurumAdi)
		}
		yeniDurum, ok = gecisler[action]
		if !ok {
			return fmt.Errorf("geçersiz işlem '%s' → durum '%s'", action, p.DurumAdi)
		}
	}

	// Türkçe Yorum: Durum değişmeyecekse (komisyon henüz tamamlanmadı) sadece güncelleme loglamadan çık
	if yeniDurum == p.DurumAdi {
		return nil
	}

	// Türkçe Yorum: Komisyon_bekliyor durumuna ilk kez geçiliyorsa komisyon oylama kayıtları açılır
	if yeniDurum == models.DurumKomisyonBekliyor && p.DurumAdi != models.DurumKomisyonBekliyor {
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
// Türkçe Bilgilendirme: İstek yapan rolünün yetkisine göre maskeleme parametresini repository'ye iletir.
func (s *ProjeService) GetProjeSurecGecmisi(projeID int, isAdminOrTTO bool) ([]models.ProjeSurecGecmisi, error) {
	return s.ProjeRepo.GetProjeSurecGecmisi(projeID, isAdminOrTTO)
}

// GetWorkflowHistoryByUyeID belirli bir kullanıcının geçmiş onay kararlarını çeker.
func (s *ProjeService) GetWorkflowHistoryByUyeID(uyeID int) ([]models.ProjeSurecGecmisi, error) {
	return s.ProjeRepo.GetWorkflowHistoryByUyeID(uyeID)
}

// GetProjectsForWorkflow rol bazında onay bekleyen projeleri listeler.
// Türkçe Yorum: Kullanıcının sahip olduğu tüm rollere göre onay bekleyen projeleri çeker ve tekil olarak birleştirir.
func (s *ProjeService) GetProjectsForWorkflow(rol string, uyeID int) ([]models.Proje, error) {
	roles := strings.Split(rol, ",")
	var allProjects []models.Proje
	seen := make(map[int]bool)
	hasWorkflowRole := false

	// appendUniq: tekrar eden proje ID'lerini filtreleyen yardımcı fonksiyon
	appendUniq := func(projeler []models.Proje) {
		for _, p := range projeler {
			if !seen[p.ProjeID] {
				seen[p.ProjeID] = true
				allProjects = append(allProjects, p)
			}
		}
	}

	for _, r := range roles {
		r = strings.TrimSpace(r)

		// Admin ise süreçteki tüm onay bekleyen projeleri görsün
		if r == models.RolAdmin {
			hasWorkflowRole = true
			adminDurumlar := []string{
				models.DurumIncelemede, models.DurumDekanOnayiBekliyor,
				models.DurumDekanOnayladi, models.DurumKomisyonBekliyor,
				models.DurumKomisyonOnayladi, models.DurumHakemBekliyor,
				models.DurumHakemOnayladi, models.DurumSozlesmeImza, models.DurumTTOAktif,
			}
			for _, d := range adminDurumlar {
				projeler, _ := s.ProjeRepo.GetProjectsForWorkflow(r, d, 0)
				appendUniq(projeler)
			}
			continue
		}

		// Türkçe Yorum: TTO rolü için süreçteki tüm projeleri takip amaçlı gösterir
		if r == models.RolTTO {
			hasWorkflowRole = true
			ttoDurumlar := []string{
				models.DurumIncelemede, models.DurumDekanOnayiBekliyor,
				models.DurumDekanOnayladi, models.DurumKomisyonBekliyor,
				models.DurumKomisyonOnayladi, models.DurumHakemAtamaBekliyor,
				models.DurumHakemBekliyor, models.DurumHakemOnayladi,
				models.DurumSozlesmeImza, models.DurumTTOAktif,
				models.DurumYururlukte, models.DurumReddedildi,
				models.DurumRevizyon, models.DurumTamamlandi,
			}
			for _, d := range ttoDurumlar {
				projeler, _ := s.ProjeRepo.GetProjectsForWorkflow(r, d, 0)
				appendUniq(projeler)
			}
			continue
		}

		// Türkçe Yorum: Hakem rolü için hakem_bekliyor durumundaki projeleri gösterir
		if r == models.RolHakem {
			hasWorkflowRole = true
			projeler, _ := s.ProjeRepo.GetProjectsForWorkflow(r, models.DurumHakemBekliyor, 0)
			appendUniq(projeler)
			continue
		}

		// Türkçe Yorum: Diğer roller için rol→durum eşlemesi
		rolDurumMap := map[string]string{
			models.RolDekan:           models.DurumDekanOnayiBekliyor,
			models.RolKomisyon:        models.DurumKomisyonBekliyor,
			models.RolKomisyonBaskani: models.DurumKomisyonBekliyor,
		}

		durum, ok := rolDurumMap[r]
		if !ok {
			continue
		}

		hasWorkflowRole = true
		projeler, _ := s.ProjeRepo.GetProjectsForWorkflow(r, durum, uyeID)
		appendUniq(projeler)
	}

	if !hasWorkflowRole {
		return nil, fmt.Errorf("onay akışı için yetkili rol bulunamadı: %s", rol)
	}

	return allProjects, nil
}

// GetButceKategorileri bütçe kategorilerini döner.
// Türkçe Bilgilendirme: ProjeRepository'den bütçe kategorilerini alıp handler'a iletir.
func (s *ProjeService) GetButceKategorileri() ([]models.ButceKategori, error) {
	return s.ProjeRepo.GetButceKategorileri()
}

// GetSistemRolleri sistemdeki tüm rolleri döner.
// Türkçe Bilgilendirme: ProjeRepository'den sistem rollerini alıp handler'a iletir.
func (s *ProjeService) GetSistemRolleri() ([]models.SistemRolTanimlama, error) {
	return s.ProjeRepo.GetSistemRolleri()
}

// GetProjeDurumlari proje durum tanımlarını döner.
// Türkçe Bilgilendirme: ProjeRepository'den proje durum tanımlarını alıp handler'a iletir.
func (s *ProjeService) GetProjeDurumlari() ([]models.ProjeDurumTanim, error) {
	return s.ProjeRepo.GetProjeDurumlari()
}
