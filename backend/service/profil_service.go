package service

import (
	"fmt"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// ProfilService yapısı, profil sayfası iş mantığını barındırır.
type ProfilService struct {
	UyeRepo   *repository.UyeRepository
	ProjeRepo *repository.ProjeRepository
}

// NewProfilService fonksiyonu, yeni bir ProfilService nesnesi döner.
func NewProfilService(uyeRepo *repository.UyeRepository, projeRepo *repository.ProjeRepository) *ProfilService {
	return &ProfilService{
		UyeRepo:   uyeRepo,
		ProjeRepo: projeRepo,
	}
}

// GetProfilBilgileri fonksiyonu, belirli bir üyenin profil bilgilerini getirir.
// Üye ve detay tabloları birleştirilmiş olarak döner.
func (s *ProfilService) GetProfilBilgileri(uyeID int) (*models.UyeWithDetay, error) {
	uye, err := s.UyeRepo.GetUyeByID(uyeID)
	if err != nil {
		return nil, fmt.Errorf("profil bilgileri alınamadı: %w", err)
	}

	return uye, nil
}

// GetProfilProjeleri fonksiyonu, belirli bir üyenin projelerini profil formatında getirir.
// Her proje için ad, tür, durum ve kullanıcının projedeki rolü döner.
func (s *ProfilService) GetProfilProjeleri(uyeID int) ([]models.ProfilProjeBilgisi, error) {
	projeler, err := s.ProjeRepo.GetProjectsByUyeIDForProfil(uyeID)
	if err != nil {
		return nil, fmt.Errorf("profil projeleri alınamadı: %w", err)
	}

	return projeler, nil
}

// TamamlaProfil fonksiyonu, kullanıcının profil detay bilgilerini tamamlar.
// Profil tamamlandı olarak işaretlenir ve rol ataması yapılır.
func (s *ProfilService) TamamlaProfil(uyeID int, req *models.ProfilTamamlamaRequest) error {
	// Türkçe Yorum: Telefon numarasını temizle ve doğrula
	cleanedPhone, err := cleanAndValidatePhone(req.Telefon)
	if err != nil {
		return err
	}
	req.Telefon = cleanedPhone

	// Mevcut detay kaydını kontrol et
	detay, err := s.UyeRepo.GetUyeDetayByUyeID(uyeID)
	if err != nil {
		// Detay kaydı yoksa yeni oluştur
		detay = &models.UyeDetay{
			UyeID:            uyeID,
			Rol:              req.Rol,
			Unvan:            req.Unvan,
			Bolum:            req.Bolum,
			Telefon:          req.Telefon,
			IzuUyesi:         req.IzuUyesi,
			ProfilTamamlandi: true,
		}
		if err := s.UyeRepo.CreateUyeDetay(detay); err != nil {
			return fmt.Errorf("profil detay kaydı oluşturulamadı: %w", err)
		}
	} else {
		// Mevcut detay kaydını güncelle
		detay.Rol = req.Rol
		detay.Unvan = req.Unvan
		detay.Bolum = req.Bolum
		detay.Telefon = req.Telefon
		detay.IzuUyesi = req.IzuUyesi
		detay.ProfilTamamlandi = true

		if err := s.UyeRepo.UpdateUyeDetay(detay); err != nil {
			return fmt.Errorf("profil detay kaydı güncellenemedi: %w", err)
		}
	}

	// Bölüm adını gerekirse bolum tablosundan çöz
	bolumMetin := req.Bolum
	if req.BolumID != nil && *req.BolumID > 0 {
		var bAdi string
		_ = s.UyeRepo.DB.QueryRow("SELECT bolum_adi FROM bolum WHERE bolum_id = $1", *req.BolumID).Scan(&bAdi)
		if bAdi != "" {
			bolumMetin = bAdi
			req.Bolum = bAdi
			detay.Bolum = bAdi
		}
	}

	// Geriye dönük uyumluluk: uye tablosundaki rol ve fakülte/bölüm alanlarını güncelle
	if err := s.UyeRepo.UpdateUyeRol(uyeID, req.Rol); err != nil {
		return fmt.Errorf("üye rol güncellenemedi: %w", err)
	}

	if err := s.UyeRepo.UpdateUyeFakulteBolum(uyeID, req.FakulteID, req.BolumID, bolumMetin); err != nil {
		return fmt.Errorf("üye fakülte/bölüm bilgisi güncellenemedi: %w", err)
	}

	// Sistem rol tablosunu güncelle (sistem_rol ilişki tablosu)
	if err := s.UyeRepo.UpsertSistemRol(uyeID, req.Rol); err != nil {
		return fmt.Errorf("sistem rolü atanamadı: %w", err)
	}

	return nil
}
