package service

import (
	"errors"
	"fmt"
	"time"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// ErrSozlesmeZatenIndirildi, sözleşme PDF'inin daha önce indirildiğini belirtir (tek seferlik indirme kuralı).
var ErrSozlesmeZatenIndirildi = errors.New("bu projenin sözleşmesi daha önce indirilmiştir ve tekrar indirilemez")

// SozlesmeService, proje sözleşmesi iş mantığını yönetir.
// Türkçe Yorum: Akademisyen tarafından iletilen sözleşme verilerini doğrular, kaydeder ve PDF üretir.
type SozlesmeService struct {
	Repo         *repository.SozlesmeRepository
	AdminRepo    *repository.AdminRepository
	Pdf          *PdfService
	EpostaService *EpostaService // Türkçe Yorum: Durum geçişi bildirimlerini tetiklemek için
}

// NewSozlesmeService yeni bir SozlesmeService örneği döner.
func NewSozlesmeService(repo *repository.SozlesmeRepository, adminRepo *repository.AdminRepository, pdf *PdfService, epostaService *EpostaService) *SozlesmeService {
	return &SozlesmeService{Repo: repo, AdminRepo: adminRepo, Pdf: pdf, EpostaService: epostaService}
}

// SaveSozlesme, sözleşme alanlarını doğrular ve veritabanına kaydeder.
// Türkçe Yorum: Bitiş tarihi proje süresine (sure_ay) göre başlangıçtan hesaplanır; istemci değeri ezilir.
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
	if sz.BaslangicTarihi == "" {
		return errors.New("sözleşme başlangıç tarihi zorunludur")
	}
	baslangic, err := time.Parse("2006-01-02", sz.BaslangicTarihi)
	if err != nil {
		return errors.New("geçersiz sözleşme başlangıç tarihi")
	}

	// Türkçe Yorum: Sözleşme başlangıcı, komisyon karar tarihinden itibaren en geç 2 ay içinde olmalıdır.
	if err := s.validateBaslangicVsKomisyonKarar(sz.ProjeID, baslangic); err != nil {
		return err
	}

	// Proje başvurusundaki süreye göre bitiş tarihini zorunlu olarak hesapla
	sureAy := 12
	if detail, dErr := s.AdminRepo.GetProjectDetailsForAdmin(sz.ProjeID, false); dErr == nil && detail != nil && detail.Proje.SureAy > 0 {
		sureAy = detail.Proje.SureAy
	}
	sz.BitisTarihi = baslangic.AddDate(0, sureAy, 0).Format("2006-01-02")

	err = s.Repo.SaveSozlesme(sz)
	if err != nil {
		return err
	}

	// Türkçe Yorum: Sözleşme kaydedildiğinde, eğer proje durumu 'sozlesme_imza' ise durumu 'sozlesme_dolduruldu' olarak güncelle.
	var currentDurum string
	err = s.Repo.DB.QueryRow(`
		SELECT pd.durum_adi 
		FROM proje p
		JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE p.proje_id = $1
	`, sz.ProjeID).Scan(&currentDurum)
	if err == nil && currentDurum == "sozlesme_imza" {
		var targetDurumID int
		err = s.Repo.DB.QueryRow(`SELECT durum_id FROM proje_durum WHERE durum_adi = 'sozlesme_dolduruldu'`).Scan(&targetDurumID)
		if err == nil {
			_, err = s.Repo.DB.Exec(`
				UPDATE proje 
				SET durum_id = $1, guncelleme_tarihi = CURRENT_TIMESTAMP 
				WHERE proje_id = $2
			`, targetDurumID, sz.ProjeID)
			if err == nil {
				// Türkçe Yorum: Süreç geçmişine kayıt yaz (schema: baslangic_durum / hedef_durum)
				_, _ = s.Repo.DB.Exec(`
					INSERT INTO proje_surec_gecmisi (proje_id, islem_yapan_id, baslangic_durum, hedef_durum, aciklama)
					VALUES ($1, $2, 'sozlesme_imza', 'sozlesme_dolduruldu', 'Yürütücü sözleşme bilgilerini doldurdu.')
				`, sz.ProjeID, sz.UyeID)
				// Türkçe Yorum: E-posta bildirimini asenkron olarak tetikle
				if s.EpostaService != nil {
					go s.EpostaService.SendStatusNotificationEmail(sz.ProjeID, sz.UyeID, "sozlesme_imza", "sozlesme_dolduruldu", "Yürütücü sözleşme bilgilerini doldurdu.")
				}
			}
		}
	}

	return nil
}

// GetSozlesme, projeye ait sözleşme kaydını getirir.
func (s *SozlesmeService) GetSozlesme(projeID int) (*models.ProjeSozlesme, error) {
	if projeID <= 0 {
		return nil, errors.New("geçersiz proje ID")
	}
	return s.Repo.GetSozlesmeByProjeID(projeID)
}

// GetKomisyonKararTarihi projenin komisyon onay karar tarihini döner.
func (s *SozlesmeService) GetKomisyonKararTarihi(projeID int) (*time.Time, error) {
	return s.Repo.GetKomisyonKararTarihi(projeID)
}

// validateBaslangicVsKomisyonKarar sözleşme başlangıcının komisyon kararından itibaren 2 ay içinde olduğunu doğrular.
// Türkçe Yorum: Alt sınır karar tarihi, üst sınır karar tarihi + 2 ay.
func (s *SozlesmeService) validateBaslangicVsKomisyonKarar(projeID int, baslangic time.Time) error {
	kararPtr, err := s.Repo.GetKomisyonKararTarihi(projeID)
	if err != nil {
		return err
	}
	if kararPtr == nil {
		return errors.New("bu proje için komisyon onay kararı bulunamadı; sözleşme başlangıç tarihi doğrulanamıyor")
	}
	karar := *kararPtr
	sonTarih := karar.AddDate(0, 2, 0)
	bas := time.Date(baslangic.Year(), baslangic.Month(), baslangic.Day(), 0, 0, 0, 0, time.UTC)

	if bas.Before(karar) {
		return fmt.Errorf(
			"sözleşme başlangıç tarihi (%s), komisyon karar tarihinden (%s) önce olamaz",
			bas.Format("02.01.2006"), karar.Format("02.01.2006"),
		)
	}
	if bas.After(sonTarih) {
		return fmt.Errorf(
			"sözleşme başlangıç tarihi (%s), komisyon karar tarihinden (%s) itibaren 2 ay içinde olmalıdır (son gün: %s)",
			bas.Format("02.01.2006"), karar.Format("02.01.2006"), sonTarih.Format("02.01.2006"),
		)
	}
	return nil
}

// GenerateAndMarkPDF, sözleşme PDF'ini üretir ve tek seferlik indirme kilidini uygular.
// Türkçe Yorum: Sözleşme daha önce indirilmişse tekrar indirmeye izin verilmez. Formda kaydedilen veya
// otomatik hesaplanan yürürlük tarihleri esas alınarak PDF üretilir.
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

	// Yürürlük: başlangıç kayıtlıysa onu kullan; bitiş her zaman proje süresine göre hesaplanır
	baslangic := time.Now()
	sureAy := detail.Proje.SureAy
	if sureAy <= 0 {
		sureAy = 12
	}
	if sz.BaslangicTarihi != "" {
		if t, err := time.Parse("2006-01-02", sz.BaslangicTarihi); err == nil {
			baslangic = t
		}
	}
	// Türkçe Yorum: PDF indirmeden önce 2 aylık komisyon penceresi yeniden doğrulanır.
	if err := s.validateBaslangicVsKomisyonKarar(projeID, baslangic); err != nil {
		return nil, err
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
