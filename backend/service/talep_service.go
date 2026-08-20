package service

import (
	"errors"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// TalepService, proje talep iş mantığını yönetir.
// Türkçe Yorum: Akademisyen tarafından gönderilen talepleri doğrulayıp repository'e yönlendirir.
type TalepService struct {
	Repo *repository.TalepRepository
}

// NewTalepService yeni bir TalepService nesnesi döner.
func NewTalepService(repo *repository.TalepRepository) *TalepService {
	return &TalepService{Repo: repo}
}

// SubmitEkSure, ek süre talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitEkSure(t *models.TalepEkSure) error {
	if t.EkSureAy <= 0 {
		return errors.New("ek süre en az 1 ay olmalıdır")
	}
	if t.Gerekce == "" {
		return errors.New("gerekçe boş olamaz")
	}
	return s.Repo.CreateEkSure(t)
}

// SubmitEkButce, ek bütçe talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitEkButce(t *models.TalepEkButce) error {
	if t.TutarTL <= 0 {
		return errors.New("tutar 0'dan büyük olmalıdır")
	}
	if t.ButceKalemi == "" || t.Gerekce == "" {
		return errors.New("bütçe kalemi ve gerekçe zorunludur")
	}
	return s.Repo.CreateEkButce(t)
}

// SubmitFasilAktarimi, fasıl aktarımı talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitFasilAktarimi(t *models.TalepFasilAktarimi) error {
	if t.TutarTL <= 0 {
		return errors.New("tutar 0'dan büyük olmalıdır")
	}
	if t.KaynakKalem == "" || t.HedefKalem == "" || t.Gerekce == "" {
		return errors.New("kaynak kalem, hedef kalem ve gerekçe zorunludur")
	}
	return s.Repo.CreateFasilAktarimi(t)
}

// SubmitArastirmaci, araştırmacı değişikliği talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitArastirmaci(t *models.TalepArastirmaci) error {
	if t.IslemTuru == "" || t.ArastirmaciAdi == "" || t.Gerekce == "" {
		return errors.New("işlem türü, araştırmacı adı ve gerekçe zorunludur")
	}
	return s.Repo.CreateArastirmaci(t)
}

// SubmitBursiyer, bursiyer işlem talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitBursiyer(t *models.TalepBursiyer) error {
	if t.BursiyerKimlik == "" || t.BursiyerAdi == "" || t.IslemTuru == "" || t.Gerekce == "" {
		return errors.New("bursiyer kimlik, ad, işlem türü ve gerekçe zorunludur")
	}
	return s.Repo.CreateBursiyer(t)
}

// SubmitProjeIptali, proje iptali talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitProjeIptali(t *models.TalepProjeIptali) error {
	if t.Gerekce == "" {
		return errors.New("gerekçe zorunludur")
	}
	return s.Repo.CreateProjeIptali(t)
}

// SubmitBilgiDegisimi, bilgi değişimi talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitBilgiDegisimi(t *models.TalepBilgiDegisimi) error {
	if t.DegisiklikTanimi == "" || t.Gerekce == "" {
		return errors.New("değişiklik tanımı ve gerekçe zorunludur")
	}
	return s.Repo.CreateBilgiDegisimi(t)
}

// SubmitProjeDondurma, proje dondurma talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitProjeDondurma(t *models.TalepProjeDondurma) error {
	if t.DondurmaAy <= 0 {
		return errors.New("dondurma süresi en az 1 ay olmalıdır")
	}
	if t.Gerekce == "" {
		return errors.New("gerekçe zorunludur")
	}
	return s.Repo.CreateProjeDondurma(t)
}

// SubmitMalzemeGuncelleme, malzeme güncelleme talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitMalzemeGuncelleme(t *models.TalepMalzemeGuncelleme) error {
	if t.GuncellemeTanimi == "" || t.Gerekce == "" {
		return errors.New("güncelleme tanımı ve gerekçe zorunludur")
	}
	return s.Repo.CreateMalzemeGuncelleme(t)
}

// SubmitAvans, avans talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitAvans(t *models.TalepAvans) error {
	if t.TutarTL <= 0 {
		return errors.New("tutar 0'dan büyük olmalıdır")
	}
	if t.ButceKalemi == "" || t.Gerekce == "" {
		return errors.New("bütçe kalemi ve gerekçe zorunludur")
	}
	return s.Repo.CreateAvans(t)
}

// GetAllTalepler, Admin/TTO için tüm talepleri listeler.
// sadeceBekleyen true ise yalnızca 'beklemede' durumundakiler gelir.
func (s *TalepService) GetAllTalepler(sadeceBekleyen bool) ([]models.TalepOzet, error) {
	return s.Repo.GetAllTalepler(sadeceBekleyen)
}

// GetTaleplerByUye, akademisyene ait tüm talepleri listeler.
// Türkçe Yorum: Akademisyenin kendi taleplerini çekmesi için kullanılır.
func (s *TalepService) GetTaleplerByUye(uyeID int, sadeceBekleyen bool) ([]models.TalepOzet, error) {
	return s.Repo.GetTaleplerByUye(uyeID, sadeceBekleyen)
}

// OnayTalep, bir talebi onaylar veya reddeder.
// Türkçe Yorum: talep_tipi string olarak gelir, doğru tabloya yönlendirir.
func (s *TalepService) OnayTalep(istek *models.TalepOnayIstek) error {
	if istek.Karar != models.TalepOnaylandi && istek.Karar != models.TalepReddedildi {
		return errors.New("geçersiz karar: 'onaylandi' veya 'reddedildi' olmalıdır")
	}
	switch istek.TalepTipi {
	case "ek_sure":
		return s.Repo.UpdateEkSureDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	case "ek_butce":
		return s.Repo.UpdateEkButceDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	case "fasil_aktarimi":
		return s.Repo.UpdateFasilAktarimiDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	case "arastirmaci":
		return s.Repo.UpdateArastirmaciDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	case "bursiyer":
		return s.Repo.UpdateBursiyerDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	case "proje_iptali":
		return s.Repo.UpdateProjeIptaliDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	case "bilgi_degisimi":
		return s.Repo.UpdateBilgiDegisimiDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	case "proje_dondurma":
		return s.Repo.UpdateProjeDondurmaD(istek.TalepID, istek.Karar, istek.RedNotu)
	case "malzeme_guncelleme":
		return s.Repo.UpdateMalzemeGuncellemeDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	case "avans":
		return s.Repo.UpdateAvansDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	default:
		return errors.New("bilinmeyen talep tipi: " + istek.TalepTipi)
	}
}
