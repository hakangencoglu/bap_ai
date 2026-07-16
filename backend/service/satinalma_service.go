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
	SatinalmaRepo    *repository.SatinalmaRepository
	ProjeRepo        *repository.ProjeRepository
	OnPurchaseAction func(talepID int, eventType string, islemYapanID int)
}

// NewSatinalmaService yeni bir SatinalmaService nesnesi oluşturur.
// Türkçe Yorum: SatinalmaService için dependency injection kurucusu.
func NewSatinalmaService(satinalmaRepo *repository.SatinalmaRepository, projeRepo *repository.ProjeRepository) *SatinalmaService {
	return &SatinalmaService{
		SatinalmaRepo: satinalmaRepo,
		ProjeRepo:     projeRepo,
	}
}

// CreatePurchaseRequests yeni bir satın alma talebi grubu (toplu talep) oluşturur.
// Türkçe Yorum: Proje yetki kontrolü yapar, proje durumunun aktif (yururlukte) olduğunu ve talep edilen toplam tutarın ilgili bütçe kalemindeki rezerve (onaylı + bekleyen) bakiyeyi aşmadığını denetler.
func (s *SatinalmaService) CreatePurchaseRequests(reqs []*models.SatinalmaTalebi, requestorRole string) error {
	if len(reqs) == 0 {
		return fmt.Errorf("en az bir satın alma kalemi gönderilmelidir")
	}

	firstReq := reqs[0]

	// Yetkilendirme Kontrolü: Admin dışındaki tüm kullanıcıların projenin ekibinde olması zorunludur.
	isAdmin := false
	for _, r := range strings.Split(requestorRole, ",") {
		if strings.TrimSpace(r) == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		isUye, err := s.ProjeRepo.IsProjeUyesi(firstReq.ProjeID, firstReq.UyeID)
		if err != nil {
			return fmt.Errorf("proje yetki kontrolü yapılamadı: %w", err)
		}
		if !isUye {
			return fmt.Errorf("bu proje için satın alma talebi oluşturma yetkiniz bulunmamaktadır")
		}
	}

	// 1. Projeyi sorgula ve durumunu kontrol et
	proje, err := s.ProjeRepo.GetProjeByID(firstReq.ProjeID)
	if err != nil {
		return fmt.Errorf("proje bilgisi alınamadı: %w", err)
	}

	// Türkçe Yorum: TTO onayından geçerek aktifleşmiş projelerin durumu 'yururlukte' olmalıdır.
	if proje.DurumAdi != "yururlukte" {
		return fmt.Errorf("satın alma talebi sadece TTO tarafından onaylanmış ve sözleşmesi imzalanmış (aktif) projeler için yapılabilir")
	}

	// 2. Rezerve bütçe (onaylı + bekleyen) limit kontrolünü yap
	kalanButce, err := s.SatinalmaRepo.GetReservedBudget(firstReq.ProjeID, firstReq.KalemID)
	if err != nil {
		return fmt.Errorf("rezerve bütçe bilgisi sorgulanamadı: %w", err)
	}

	// Toplam talep tutarını hesapla
	var toplamTalepTutar float64
	for _, req := range reqs {
		toplamTalepTutar += float64(req.Miktar) * req.BirimFiyat
	}

	if toplamTalepTutar > kalanButce {
		return fmt.Errorf("talep edilen toplam tutar (%.2f ₺), bu bütçe kaleminin onaylanmış ve bekleyen taleplerden kalan limitini (%.2f ₺) aşmaktadır", toplamTalepTutar, kalanButce)
	}

	// 3. Talepleri veritabanına ekle
	err = s.SatinalmaRepo.CreatePurchaseRequests(reqs)
	if err == nil && s.OnPurchaseAction != nil {
		for _, req := range reqs {
			go s.OnPurchaseAction(req.TalepID, "create", req.UyeID)
		}
	}
	return err
}

// CreatePurchaseRequest yeni bir satın alma talebi oluşturur.
// Türkçe Yorum: Geriye dönük uyumluluk için tekli satın alma talebi ekleme isteklerini toplu ekleme metoduna yönlendirir.
func (s *SatinalmaService) CreatePurchaseRequest(req *models.SatinalmaTalebi, requestorRole string) error {
	return s.CreatePurchaseRequests([]*models.SatinalmaTalebi{req}, requestorRole)
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
	err = s.SatinalmaRepo.UpdatePurchaseStatus(talepID, status, redNedeni)
	if err == nil && s.OnPurchaseAction != nil {
		go s.OnPurchaseAction(talepID, "update", 0)
	}
	return err
}
