package service

import (
	"fmt"
	"strings"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// SatinalmaService yapısı, satın alma işlemlerine ait iş mantığını yönetir.
// Türkçe Yorum: Satın alma talepleri oluşturulurken bütçe kalemi limit kontrolü ve proje durum doğrulaması yapan servis katmanıdır.
type SatinalmaService struct {
	SatinalmaRepo *repository.SatinalmaRepository
	ProjeRepo     *repository.ProjeRepository
}

// NewSatinalmaService yeni bir SatinalmaService nesnesi oluşturur.
// Türkçe Yorum: SatinalmaService için dependency injection kurucusu.
func NewSatinalmaService(satinalmaRepo *repository.SatinalmaRepository, projeRepo *repository.ProjeRepository) *SatinalmaService {
	return &SatinalmaService{
		SatinalmaRepo: satinalmaRepo,
		ProjeRepo:     projeRepo,
	}
}

// CreatePurchaseRequest yeni bir satın alma talebi oluşturur.
// Türkçe Yorum: Proje yetki kontrolü yapar, proje durumunun aktif (tamamlandi) olduğunu ve talep edilen tutarın ilgili bütçe kalemindeki rezerve (onaylı + bekleyen) bakiyeyi aşmadığını denetler.
func (s *SatinalmaService) CreatePurchaseRequest(req *models.SatinalmaTalebi, requestorRole string) error {
	// Yetkilendirme Kontrolü: Admin dışındaki tüm kullanıcıların projenin ekibinde olması zorunludur.
	isAdmin := false
	for _, r := range strings.Split(requestorRole, ",") {
		if strings.TrimSpace(r) == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		isUye, err := s.ProjeRepo.IsProjeUyesi(req.ProjeID, req.UyeID)
		if err != nil {
			return fmt.Errorf("proje yetki kontrolü yapılamadı: %w", err)
		}
		if !isUye {
			return fmt.Errorf("bu proje için satın alma talebi oluşturma yetkiniz bulunmamaktadır")
		}
	}

	// 1. Projeyi sorgula ve durumunu kontrol et
	proje, err := s.ProjeRepo.GetProjeByID(req.ProjeID)
	if err != nil {
		return fmt.Errorf("proje bilgisi alınamadı: %w", err)
	}

	// Türkçe Yorum: TTO onayından geçerek aktifleşmiş projelerin durumu 'tamamlandi' olmalıdır.
	if proje.DurumAdi != "tamamlandi" {
		return fmt.Errorf("satın alma talebi sadece TTO tarafından onaylanmış ve sözleşmesi imzalanmış (aktif) projeler için yapılabilir")
	}

	// 2. Rezerve bütçe (onaylı + bekleyen) limit kontrolünü yap
	kalanButce, err := s.SatinalmaRepo.GetReservedBudget(req.ProjeID, req.KalemID)
	if err != nil {
		return fmt.Errorf("rezerve bütçe bilgisi sorgulanamadı: %w", err)
	}

	talepTutar := float64(req.Miktar) * req.BirimFiyat
	if talepTutar > kalanButce {
		return fmt.Errorf("talep edilen toplam tutar (%.2f ₺), bu bütçe kaleminin onaylanmış ve bekleyen taleplerden kalan limitini (%.2f ₺) aşmaktadır", talepTutar, kalanButce)
	}

	// 3. Talebi veritabanına ekle
	return s.SatinalmaRepo.CreatePurchaseRequest(req)
}

// GetPurchaseRequestsByProject bir projeye ait tüm talepleri listeler.
// Türkçe Yorum: Belirli bir proje altındaki tüm satın alma işlemlerini yetkilendirme kontrolü yaparak listeler.
func (s *SatinalmaService) GetPurchaseRequestsByProject(projeID int, requestorID int, requestorRole string) ([]models.SatinalmaTalebi, error) {
	// Yetkilendirme Kontrolü: Admin ve TTO rolleri dışındaki kullanıcıların proje üyesi olması zorunludur.
	hasAccess := false
	for _, r := range strings.Split(requestorRole, ",") {
		rClean := strings.TrimSpace(r)
		if rClean == "admin" || rClean == "tto" {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		isUye, err := s.ProjeRepo.IsProjeUyesi(projeID, requestorID)
		if err != nil {
			return nil, fmt.Errorf("proje yetki kontrolü yapılamadı: %w", err)
		}
		if !isUye {
			return nil, fmt.Errorf("bu projenin satın alma taleplerini görüntüleme yetkiniz bulunmamaktadır")
		}
	}

	return s.SatinalmaRepo.GetPurchaseRequestsByProject(projeID)
}

// GetAllPurchaseRequests tüm sistemdeki satın alma taleplerini listeler.
// Türkçe Yorum: TTO yetkilileri için tüm talepleri listeler.
func (s *SatinalmaService) GetAllPurchaseRequests() ([]models.SatinalmaTalebi, error) {
	return s.SatinalmaRepo.GetAllPurchaseRequests()
}

// UpdatePurchaseStatus satın alma talebini onaylar veya reddeder.
// Türkçe Yorum: TTO yetkilisinin verdiği karara göre talebi 'Onaylandı' veya 'Reddedildi' durumuna getirir. Onay durumunda bütçe limitini tekrar doğrular.
func (s *SatinalmaService) UpdatePurchaseStatus(talepID int, status string, redNedeni string) error {
	// 1. Talebi bul
	talep, err := s.SatinalmaRepo.GetPurchaseRequestByID(talepID)
	if err != nil {
		return fmt.Errorf("satın alma talebi bulunamadı: %w", err)
	}

	if talep.Durum != "Beklemede" {
		return fmt.Errorf("sadece 'Beklemede' durumundaki satın alma talepleri güncellenebilir")
	}

	// 2. Eğer onaylanıyorsa kalan bütçeyi son bir kez daha kontrol et
	if status == "Onaylandı" {
		kalanButce, err := s.SatinalmaRepo.GetRemainingBudget(talep.ProjeID, talep.KalemID)
		if err != nil {
			return fmt.Errorf("bütçe kalemi kalan limiti doğrulanamadı: %w", err)
		}

		if talep.ToplamFiyat > kalanButce {
			return fmt.Errorf("bu satın alma talebi onaylandığında bütçe kalem limiti (Kalan: %.2f ₺) aşılacaktır", kalanButce)
		}
	}

	// 3. Durumu güncelle
	return s.SatinalmaRepo.UpdatePurchaseStatus(talepID, status, redNedeni)
}
