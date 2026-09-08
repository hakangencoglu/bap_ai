package service

import (
	"fmt"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// KomisyonToplantiService toplantı ↔ proje ve talep ilişkilendirme süreçlerini yönetir.
// Türkçe Yorum: Proje/Talep ekleme, çıkarma, karar kaydı ve durum senkronizasyonunu uygular.
type KomisyonToplantiService struct {
	ToplantiRepo *repository.KomisyonToplantiRepository
	ProjeService *ProjeService
	TalepService *TalepService
}

// NewKomisyonToplantiService yeni bir KomisyonToplantiService oluşturur.
func NewKomisyonToplantiService(toplantiRepo *repository.KomisyonToplantiRepository, projeService *ProjeService, talepService *TalepService) *KomisyonToplantiService {
	return &KomisyonToplantiService{
		ToplantiRepo: toplantiRepo,
		ProjeService: projeService,
		TalepService: talepService,
	}
}

// AddProjeToToplanti bir projeyi veya talebi toplantı gündemine ekler.
// Türkçe Yorum: Başvurularda komisyon_bekliyor doğrulaması yapılır; taleplerde beklemede olan talep eklenir.
func (s *KomisyonToplantiService) AddProjeToToplanti(toplantiID, projeID, talepID int, gundemTipi, talepTipi string, gundemSirasi, ekleyenID int) error {
	if toplantiID <= 0 || projeID <= 0 {
		return fmt.Errorf("geçersiz toplantı veya proje ID'si")
	}
	// Türkçe Yorum: Kararları alınmış (tamamlanmış) veya iptal edilmiş toplantının gündemi değiştirilemez.
	durum, err := s.ToplantiRepo.GetToplantiDurum(toplantiID)
	if err != nil {
		return err
	}
	if durum == "tamamlandi" {
		return fmt.Errorf("kararları tamamlanmış toplantıya yeni gündem maddesi eklenemez")
	}
	if durum == "iptal" {
		return fmt.Errorf("iptal edilmiş toplantıya gündem maddesi eklenemez")
	}
	// Yalnızca yeni başvuru tiplerinde proje komisyon_bekliyor olmalı
	if talepID == 0 && (gundemTipi == "basvuru" || gundemTipi == "") {
		if err := s.assertProjeKomisyonBekliyor(projeID); err != nil {
			return err
		}
	}
	return s.ToplantiRepo.AddProjeToToplanti(toplantiID, projeID, talepID, gundemTipi, talepTipi, gundemSirasi, ekleyenID)
}

// RemoveProjeFromToplanti bir projeyi veya talebi toplantı gündeminden çıkarır.
// Türkçe Yorum: Kararları tamamlanmış toplantının gündemi geriye dönük değiştirilemez.
func (s *KomisyonToplantiService) RemoveProjeFromToplanti(toplantiID, projeID, talepID int) error {
	durum, err := s.ToplantiRepo.GetToplantiDurum(toplantiID)
	if err != nil {
		return err
	}
	if durum == "tamamlandi" {
		return fmt.Errorf("kararları tamamlanmış toplantının gündemi değiştirilemez")
	}
	return s.ToplantiRepo.RemoveProjeFromToplanti(toplantiID, projeID, talepID)
}

// GetProjectsByToplanti bir toplantıdaki gündem maddelerini (proje ve talepler) listeler.
func (s *KomisyonToplantiService) GetProjectsByToplanti(toplantiID int) ([]*models.KomisyonToplantisiProje, error) {
	return s.ToplantiRepo.GetProjectsByToplanti(toplantiID)
}

// GetToplantilerByProje bir projenin görüşüldüğü tüm toplantıları listeler.
func (s *KomisyonToplantiService) GetToplantilerByProje(projeID int) ([]*models.KomisyonToplantisiProje, error) {
	return s.ToplantiRepo.GetToplantilerByProje(projeID)
}

// SetProjeKarar toplantıdaki bir proje veya talep için karar kaydeder ve durumları senkronize eder.
// Türkçe Yorum: Talep kararlarında TalepService.OnayTalep tetiklenir; proje başvurularında ProcessWorkflowAction tetiklenir.
func (s *KomisyonToplantiService) SetProjeKarar(toplantiID, projeID, talepID int, islemYapanID int, karar, aciklama string, talepTipi string) error {
	gecerliKararlar := map[string]bool{
		"bekliyor": true, "onaylandi": true, "reddedildi": true, "ertelendi": true, "revizyon": true,
	}
	if !gecerliKararlar[karar] {
		return fmt.Errorf("geçersiz karar: %s (bekliyor|onaylandi|reddedildi|ertelendi|revizyon olmalı)", karar)
	}

	// 1. Eğer bir talep gündem maddesi ise (talepID > 0)
	if talepID > 0 {
		if err := s.ToplantiRepo.SetProjeKarar(toplantiID, projeID, talepID, karar, aciklama); err != nil {
			return err
		}
		if s.TalepService != nil && (karar == "onaylandi" || karar == "reddedildi") {
			kararTalep := models.TalepOnaylandi
			if karar == "reddedildi" {
				kararTalep = models.TalepReddedildi
			}
			err := s.TalepService.OnayTalep(&models.TalepOnayIstek{
				TalepID:   talepID,
				TalepTipi: talepTipi,
				Karar:     kararTalep,
				RedNotu:   aciklama,
			})
			if err != nil {
				return fmt.Errorf("toplantı kararı kaydedildi ancak talep güncellenemedi: %w", err)
			}
		}
		if _, syncErr := s.ToplantiRepo.SyncToplantiDurumFromProjeler(toplantiID); syncErr != nil {
			return fmt.Errorf("karar kaydedildi ancak toplantı durumu güncellenemedi: %w", syncErr)
		}
		return nil
	}

	// 2. Eğer bir proje başvurusu gündem maddesi ise (talepID == 0)
	if karar == "onaylandi" || karar == "reddedildi" || karar == "revizyon" {
		if err := s.assertProjeKomisyonBekliyor(projeID); err != nil {
			return err
		}
	}

	if err := s.ToplantiRepo.SetProjeKarar(toplantiID, projeID, 0, karar, aciklama); err != nil {
		return err
	}

	if karar == "ertelendi" || karar == "bekliyor" {
		if _, syncErr := s.ToplantiRepo.SyncToplantiDurumFromProjeler(toplantiID); syncErr != nil {
			return fmt.Errorf("karar kaydedildi ancak toplantı durumu güncellenemedi: %w", syncErr)
		}
		return nil
	}

	action := ""
	switch karar {
	case "onaylandi":
		action = models.AksiyonOnayla
	case "reddedildi":
		action = models.AksiyonReddet
	case "revizyon":
		action = models.AksiyonRevizyon
	}

	if s.ProjeService == nil {
		return fmt.Errorf("proje servisi yapılandırılmamış; durum senkronu yapılamadı")
	}
	if err := s.ProjeService.ProcessWorkflowAction(projeID, islemYapanID, action, aciklama); err != nil {
		return fmt.Errorf("toplantı kararı kaydedildi ancak proje durumu güncellenemedi: %w", err)
	}

	if _, syncErr := s.ToplantiRepo.SyncToplantiDurumFromProjeler(toplantiID); syncErr != nil {
		return fmt.Errorf("proje güncellendi ancak toplantı durumu senkronlanamadı: %w", syncErr)
	}
	return nil
}

// GetToplantiBelgeDetay PDF tutanağı için toplantı bütününü döner.
func (s *KomisyonToplantiService) GetToplantiBelgeDetay(toplantiID int) (*models.KomisyonToplantiBelge, error) {
	return s.ToplantiRepo.GetToplantiBelgeDetay(toplantiID)
}

// GetBekleyenProjeler komisyon_bekliyor durumundaki projeleri ve beklemedeki talepleri listeler.
// Türkçe Yorum: Toplantı formunda gündeme eklenecek aday projeleri ve talepleri döner.
func (s *KomisyonToplantiService) GetBekleyenProjeler() ([]*models.KomisyonBekleyenProje, error) {
	return s.ToplantiRepo.GetBekleyenProjeler()
}

// assertProjeKomisyonBekliyor projenin komisyon_bekliyor durumunda olduğunu doğrular.
func (s *KomisyonToplantiService) assertProjeKomisyonBekliyor(projeID int) error {
	if s.ProjeService == nil || s.ProjeService.ProjeRepo == nil {
		return fmt.Errorf("proje doğrulaması yapılamadı")
	}
	p, err := s.ProjeService.ProjeRepo.GetProjeByID(projeID)
	if err != nil {
		return fmt.Errorf("proje bulunamadı: %w", err)
	}
	if p.DurumAdi != models.DurumKomisyonBekliyor {
		return fmt.Errorf("yalnızca 'Komisyon Onayı Bekliyor' durumundaki projeler işlenebilir (mevcut: %s)", p.DurumAdi)
	}
	return nil
}
