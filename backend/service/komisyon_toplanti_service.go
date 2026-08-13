package service

import (
	"fmt"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// KomisyonToplantiService toplantı ↔ proje ilişkilendirme süreçlerini yönetir.
// Türkçe Yorum: Proje ekleme, çıkarma, karar kaydı ve proje durum senkronunu uygular.
type KomisyonToplantiService struct {
	ToplantiRepo *repository.KomisyonToplantiRepository
	ProjeService *ProjeService
}

// NewKomisyonToplantiService yeni bir KomisyonToplantiService oluşturur.
func NewKomisyonToplantiService(toplantiRepo *repository.KomisyonToplantiRepository, projeService *ProjeService) *KomisyonToplantiService {
	return &KomisyonToplantiService{
		ToplantiRepo: toplantiRepo,
		ProjeService: projeService,
	}
}

// AddProjeToToplanti bir projeyi toplantı gündemine ekler.
// Türkçe Yorum: Yalnızca komisyon_bekliyor durumundaki projeler eklenebilir.
func (s *KomisyonToplantiService) AddProjeToToplanti(toplantiID, projeID, gundemSirasi, ekleyenID int) error {
	if toplantiID <= 0 || projeID <= 0 {
		return fmt.Errorf("geçersiz toplantı veya proje ID'si")
	}
	if err := s.assertProjeKomisyonBekliyor(projeID); err != nil {
		return err
	}
	return s.ToplantiRepo.AddProjeToToplanti(toplantiID, projeID, gundemSirasi, ekleyenID)
}

// RemoveProjeFromToplanti bir projeyi toplantı gündeminden çıkarır.
func (s *KomisyonToplantiService) RemoveProjeFromToplanti(toplantiID, projeID int) error {
	return s.ToplantiRepo.RemoveProjeFromToplanti(toplantiID, projeID)
}

// GetProjectsByToplanti bir toplantıdaki projeleri listeler.
func (s *KomisyonToplantiService) GetProjectsByToplanti(toplantiID int) ([]*models.KomisyonToplantisiProje, error) {
	return s.ToplantiRepo.GetProjectsByToplanti(toplantiID)
}

// GetToplantilerByProje bir projenin görüşüldüğü tüm toplantıları listeler.
func (s *KomisyonToplantiService) GetToplantilerByProje(projeID int) ([]*models.KomisyonToplantisiProje, error) {
	return s.ToplantiRepo.GetToplantilerByProje(projeID)
}

// SetProjeKarar toplantıdaki bir proje için karar kaydeder ve proje durumunu senkronize eder.
// Türkçe Yorum: onaylandi/reddedildi/revizyon → ProcessWorkflowAction; ertelendi yalnızca köprü kaydı.
func (s *KomisyonToplantiService) SetProjeKarar(toplantiID, projeID, islemYapanID int, karar, aciklama string) error {
	gecerliKararlar := map[string]bool{
		"bekliyor": true, "onaylandi": true, "reddedildi": true, "ertelendi": true, "revizyon": true,
	}
	if !gecerliKararlar[karar] {
		return fmt.Errorf("geçersiz karar: %s (bekliyor|onaylandi|reddedildi|ertelendi|revizyon olmalı)", karar)
	}

	// Durum değiştiren kararlarda proje komisyon_bekliyor olmalı
	if karar == "onaylandi" || karar == "reddedildi" || karar == "revizyon" {
		if err := s.assertProjeKomisyonBekliyor(projeID); err != nil {
			return err
		}
	}

	if err := s.ToplantiRepo.SetProjeKarar(toplantiID, projeID, karar, aciklama); err != nil {
		return err
	}

	// Türkçe Yorum: Erteleme ve bekliyor durumları proje durumunu değiştirmez; toplantı durumu yine senkronlanır.
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

	// Türkçe Yorum: Tüm gündem projeleri nihai karara bağlandıysa toplantı tamamlandı olur.
	if _, syncErr := s.ToplantiRepo.SyncToplantiDurumFromProjeler(toplantiID); syncErr != nil {
		return fmt.Errorf("proje güncellendi ancak toplantı durumu senkronlanamadı: %w", syncErr)
	}
	return nil
}

// GetToplantiBelgeDetay PDF tutanağı için toplantı bütününü döner.
func (s *KomisyonToplantiService) GetToplantiBelgeDetay(toplantiID int) (*models.KomisyonToplantiBelge, error) {
	return s.ToplantiRepo.GetToplantiBelgeDetay(toplantiID)
}

// GetBekleyenProjeler komisyon_bekliyor durumundaki projeleri listeler.
// Türkçe Yorum: Toplantı formunda gündeme eklenecek aday projeleri döner.
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
