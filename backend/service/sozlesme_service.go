package service

import (
	"errors"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// SozlesmeService, proje sözleşmesi iş mantığını yönetir.
// Türkçe Yorum: Akademisyen tarafından iletilen sözleşme verilerini doğrular ve kaydeder.
type SozlesmeService struct {
	Repo *repository.SozlesmeRepository
}

// NewSozlesmeService yeni bir SozlesmeService örneği döner.
func NewSozlesmeService(repo *repository.SozlesmeRepository) *SozlesmeService {
	return &SozlesmeService{Repo: repo}
}

// SaveSozlesme, sözleşme alanlarını doğrular ve veritabanına kaydeder.
func (s *SozlesmeService) SaveSozlesme(sz *models.ProjeSozlesme) error {
	if sz.ProjeID <= 0 {
		return errors.New("geçersiz proje ID")
	}
	if sz.TCKimlik == "" || len(sz.TCKimlik) < 10 {
		return errors.New("geçerli bir T.C. Kimlik numarası giriniz")
	}
	if sz.YurutucuAdres == "" {
		return errors.New("yürütücü adresi boş olamaz")
	}
	if sz.YurutucuTelefon == "" {
		return errors.New("telefon numarası boş olamaz")
	}
	if sz.YurutucuEposta == "" {
		return errors.New("e-posta adresi boş olamaz")
	}
	if sz.BaslangicTarihi == "" || sz.BitisTarihi == "" {
		return errors.New("sözleşme başlangıç ve bitiş tarihleri zorunludur")
	}
	return s.Repo.SaveSozlesme(sz)
}

// GetSozlesme, projeye ait sözleşme kaydını getirir.
func (s *SozlesmeService) GetSozlesme(projeID int) (*models.ProjeSozlesme, error) {
	if projeID <= 0 {
		return nil, errors.New("geçersiz proje ID")
	}
	return s.Repo.GetSozlesmeByProjeID(projeID)
}
