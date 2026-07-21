package service

import (
	"errors"
	"time"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// ErrSozlesmeZatenIndirildi, sözleşme PDF'inin daha önce indirildiğini belirtir (tek seferlik indirme kuralı).
var ErrSozlesmeZatenIndirildi = errors.New("bu projenin sözleşmesi daha önce indirilmiştir ve tekrar indirilemez")

// SozlesmeService, proje sözleşmesi iş mantığını yönetir.
// Türkçe Yorum: Akademisyen tarafından iletilen sözleşme verilerini doğrular, kaydeder ve PDF üretir.
type SozlesmeService struct {
	Repo      *repository.SozlesmeRepository
	AdminRepo *repository.AdminRepository
	Pdf       *PdfService
}

// NewSozlesmeService yeni bir SozlesmeService örneği döner.
func NewSozlesmeService(repo *repository.SozlesmeRepository, adminRepo *repository.AdminRepository, pdf *PdfService) *SozlesmeService {
	return &SozlesmeService{Repo: repo, AdminRepo: adminRepo, Pdf: pdf}
}

// SaveSozlesme, sözleşme alanlarını doğrular ve veritabanına kaydeder.
// Türkçe Yorum: Yürürlük tarihleri PDF indirme anında hesaplandığı için burada zorunlu değildir.
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
	return s.Repo.SaveSozlesme(sz)
}

// GetSozlesme, projeye ait sözleşme kaydını getirir.
func (s *SozlesmeService) GetSozlesme(projeID int) (*models.ProjeSozlesme, error) {
	if projeID <= 0 {
		return nil, errors.New("geçersiz proje ID")
	}
	return s.Repo.GetSozlesmeByProjeID(projeID)
}

// GenerateAndMarkPDF, sözleşme PDF'ini üretir ve tek seferlik indirme kilidini uygular.
// Türkçe Yorum: Sözleşme daha önce indirilmişse tekrar indirmeye izin verilmez. Yürürlük tarihleri
// indirme anı (bugün) baz alınarak, BAP türü/proje süresine göre hesaplanıp kaydedilir.
func (s *SozlesmeService) GenerateAndMarkPDF(projeID int) ([]byte, error) {
	if projeID <= 0 {
		return nil, errors.New("geçersiz proje ID")
	}

	sz, err := s.Repo.GetSozlesmeByProjeID(projeID)
	if err != nil {
		return nil, err
	}
	if sz == nil {
		return nil, errors.New("önce sözleşme bilgilerini doldurup kaydetmelisiniz")
	}
	if sz.IndirildiMi {
		return nil, ErrSozlesmeZatenIndirildi
	}

	detail, err := s.AdminRepo.GetProjectDetailsForAdmin(projeID, false)
	if err != nil {
		return nil, errors.New("proje detayları alınamadı")
	}

	// Yürürlük tarihleri: başlangıç = indirme tarihi (bugün), bitiş = başlangıç + proje süresi (ay)
	baslangic := time.Now()
	sureAy := detail.Proje.SureAy
	if sureAy <= 0 {
		sureAy = 12
	}
	bitis := baslangic.AddDate(0, sureAy, 0)

	pdfBytes, err := s.Pdf.GenerateSozlesmePDF(detail, sz, baslangic, bitis)
	if err != nil {
		return nil, err
	}

	// PDF başarıyla üretildi; indirme kilidini uygula ve hesaplanan tarihleri kaydet.
	if err := s.Repo.MarkIndirildi(projeID, baslangic.Format("2006-01-02"), bitis.Format("2006-01-02")); err != nil {
		return nil, err
	}

	return pdfBytes, nil
}
