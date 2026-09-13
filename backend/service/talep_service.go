package service

import (
	"errors"
	"fmt"
	"log"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// TalepService, proje talep iş mantığını yönetir.
// Türkçe Yorum: Akademisyen taleplerini doğrular, repository'e yönlendirir ve audit kaydı tutar.
type TalepService struct {
	Repo  *repository.TalepRepository
	Audit *repository.DegisiklikRepository
}

// NewTalepService yeni bir TalepService nesnesi döner.
func NewTalepService(repo *repository.TalepRepository, audit *repository.DegisiklikRepository) *TalepService {
	return &TalepService{Repo: repo, Audit: audit}
}

// SubmitEkSure, ek süre talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitEkSure(t *models.TalepEkSure) error {
	if t.EkSureAy <= 0 {
		return errors.New("ek süre en az 1 ay olmalıdır")
	}
	if t.Gerekce == "" {
		return errors.New("gerekçe boş olamaz")
	}
	if err := s.Repo.CreateEkSure(t); err != nil {
		return err
	}
	s.recordOlusturma("ek_sure", t.ProjeID, t.ID, t.TalepNo, t.UyeID,
		fmt.Sprintf("Ek süre talebi oluşturuldu (%d ay)", t.EkSureAy), t)
	return nil
}

// SubmitEkButce, ek bütçe talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitEkButce(t *models.TalepEkButce) error {
	if t.TutarTL <= 0 {
		return errors.New("tutar 0'dan büyük olmalıdır")
	}
	if t.ButceKalemi == "" || t.Gerekce == "" {
		return errors.New("bütçe kalemi ve gerekçe zorunludur")
	}
	if err := s.Repo.CreateEkButce(t); err != nil {
		return err
	}
	s.recordOlusturma("ek_butce", t.ProjeID, t.ID, t.TalepNo, t.UyeID,
		fmt.Sprintf("Ek bütçe talebi oluşturuldu (%s, %.2f ₺)", t.ButceKalemi, t.TutarTL), t)
	return nil
}

// SubmitFasilAktarimi, fasıl aktarımı talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitFasilAktarimi(t *models.TalepFasilAktarimi) error {
	if t.TutarTL <= 0 {
		return errors.New("tutar 0'dan büyük olmalıdır")
	}
	if t.KaynakKalem == "" || t.HedefKalem == "" || t.Gerekce == "" {
		return errors.New("kaynak kalem, hedef kalem ve gerekçe zorunludur")
	}
	if err := s.Repo.CreateFasilAktarimi(t); err != nil {
		return err
	}
	s.recordOlusturma("fasil_aktarimi", t.ProjeID, t.ID, t.TalepNo, t.UyeID,
		fmt.Sprintf("Fasıl aktarımı talebi oluşturuldu (%s → %s, %.2f ₺)", t.KaynakKalem, t.HedefKalem, t.TutarTL), t)
	return nil
}

// SubmitArastirmaci, araştırmacı değişikliği talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitArastirmaci(t *models.TalepArastirmaci) error {
	if t.IslemTuru == "" || t.ArastirmaciAdi == "" || t.Gerekce == "" {
		return errors.New("işlem türü, araştırmacı adı ve gerekçe zorunludur")
	}
	if err := s.Repo.CreateArastirmaci(t); err != nil {
		return err
	}
	s.recordOlusturma("arastirmaci", t.ProjeID, t.ID, t.TalepNo, t.UyeID,
		fmt.Sprintf("Araştırmacı talebi oluşturuldu (%s: %s)", t.IslemTuru, t.ArastirmaciAdi), t)
	return nil
}

// SubmitBursiyer, bursiyer işlem talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitBursiyer(t *models.TalepBursiyer) error {
	if t.BursiyerKimlik == "" || t.BursiyerAdi == "" || t.IslemTuru == "" || t.Gerekce == "" {
		return errors.New("bursiyer kimlik, ad, işlem türü ve gerekçe zorunludur")
	}
	if err := s.Repo.CreateBursiyer(t); err != nil {
		return err
	}
	payload := map[string]interface{}{
		"id": t.ID, "proje_id": t.ProjeID, "talep_no": t.TalepNo,
		"bursiyer_adi": t.BursiyerAdi, "islem_turu": t.IslemTuru, "gerekce": t.Gerekce,
		"bursiyer_kimlik": maskKimlik(t.BursiyerKimlik),
	}
	s.recordOlusturma("bursiyer", t.ProjeID, t.ID, t.TalepNo, t.UyeID,
		fmt.Sprintf("Bursiyer talebi oluşturuldu (%s: %s)", t.IslemTuru, t.BursiyerAdi), payload)
	return nil
}

// SubmitProjeIptali, proje iptali talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitProjeIptali(t *models.TalepProjeIptali) error {
	if t.Gerekce == "" {
		return errors.New("gerekçe zorunludur")
	}
	if err := s.Repo.CreateProjeIptali(t); err != nil {
		return err
	}
	s.recordOlusturma("proje_iptali", t.ProjeID, t.ID, t.TalepNo, t.UyeID,
		"Proje iptali talebi oluşturuldu", t)
	return nil
}

// SubmitBilgiDegisimi, bilgi değişimi talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitBilgiDegisimi(t *models.TalepBilgiDegisimi) error {
	if t.DegisiklikTanimi == "" || t.Gerekce == "" {
		return errors.New("değişiklik tanımı ve gerekçe zorunludur")
	}
	if err := s.Repo.CreateBilgiDegisimi(t); err != nil {
		return err
	}
	s.recordOlusturma("bilgi_degisimi", t.ProjeID, t.ID, t.TalepNo, t.UyeID,
		"Bilgi değişimi talebi oluşturuldu", t)
	return nil
}

// SubmitProjeDondurma, proje dondurma talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitProjeDondurma(t *models.TalepProjeDondurma) error {
	if t.DondurmaAy <= 0 {
		return errors.New("dondurma süresi en az 1 ay olmalıdır")
	}
	if t.Gerekce == "" {
		return errors.New("gerekçe zorunludur")
	}
	if err := s.Repo.CreateProjeDondurma(t); err != nil {
		return err
	}
	s.recordOlusturma("proje_dondurma", t.ProjeID, t.ID, t.TalepNo, t.UyeID,
		fmt.Sprintf("Proje dondurma talebi oluşturuldu (%d ay)", t.DondurmaAy), t)
	return nil
}

// SubmitMalzemeGuncelleme, malzeme güncelleme talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitMalzemeGuncelleme(t *models.TalepMalzemeGuncelleme) error {
	if t.GuncellemeTanimi == "" || t.Gerekce == "" {
		return errors.New("güncelleme tanımı ve gerekçe zorunludur")
	}
	if err := s.Repo.CreateMalzemeGuncelleme(t); err != nil {
		return err
	}
	s.recordOlusturma("malzeme_guncelleme", t.ProjeID, t.ID, t.TalepNo, t.UyeID,
		"Malzeme güncelleme talebi oluşturuldu", t)
	return nil
}

// SubmitAvans, avans talebini doğrulayıp kaydeder.
func (s *TalepService) SubmitAvans(t *models.TalepAvans) error {
	if t.TutarTL <= 0 {
		return errors.New("tutar 0'dan büyük olmalıdır")
	}
	if t.ButceKalemi == "" || t.Gerekce == "" {
		return errors.New("bütçe kalemi ve gerekçe zorunludur")
	}
	if err := s.Repo.CreateAvans(t); err != nil {
		return err
	}
	s.recordOlusturma("avans", t.ProjeID, t.ID, t.TalepNo, t.UyeID,
		fmt.Sprintf("Avans talebi oluşturuldu (%s, %.2f ₺)", t.ButceKalemi, t.TutarTL), t)
	return nil
}

// GetAllTalepler, Admin/TTO için tüm talepleri listeler.
func (s *TalepService) GetAllTalepler(sadeceBekleyen bool) ([]models.TalepOzet, error) {
	return s.Repo.GetAllTalepler(sadeceBekleyen)
}

// GetTaleplerByUye, akademisyene ait tüm talepleri listeler.
func (s *TalepService) GetTaleplerByUye(uyeID int, sadeceBekleyen bool) ([]models.TalepOzet, error) {
	return s.Repo.GetTaleplerByUye(uyeID, sadeceBekleyen)
}

// OnayTalep, bir talebi onaylar veya reddeder; önce/sonra audit yazar.
func (s *TalepService) OnayTalep(istek *models.TalepOnayIstek) error {
	if istek.Karar != models.TalepOnaylandi && istek.Karar != models.TalepReddedildi {
		return errors.New("geçersiz karar: 'onaylandi' veya 'reddedildi' olmalıdır")
	}

	projeID, talepNo, oncekiTalep, err := s.Audit.GetTalepMeta(istek.TalepTipi, istek.TalepID)
	if err != nil {
		return fmt.Errorf("talep bilgisi alınamadı: %w", err)
	}
	oncekiDetaylar, err := s.captureOnayOnceki(istek.TalepTipi, projeID, oncekiTalep)
	if err != nil {
		return fmt.Errorf("önceki durum yakalanamadı: %w", err)
	}

	switch istek.TalepTipi {
	case "ek_sure":
		err = s.Repo.UpdateEkSureDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	case "ek_butce":
		err = s.Repo.UpdateEkButceDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	case "fasil_aktarimi":
		err = s.Repo.ApproveFasilAktarimi(istek.TalepID, istek.Karar, istek.RedNotu)
	case "arastirmaci":
		err = s.Repo.UpdateArastirmaciDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	case "bursiyer":
		err = s.Repo.UpdateBursiyerDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	case "proje_iptali":
		err = s.Repo.UpdateProjeIptaliDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	case "bilgi_degisimi":
		err = s.Repo.UpdateBilgiDegisimiDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	case "proje_dondurma":
		err = s.Repo.UpdateProjeDondurmaD(istek.TalepID, istek.Karar, istek.RedNotu)
	case "malzeme_guncelleme":
		err = s.Repo.UpdateMalzemeGuncellemeDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	case "avans":
		err = s.Repo.UpdateAvansDurum(istek.TalepID, istek.Karar, istek.RedNotu)
	default:
		return errors.New("bilinmeyen talep tipi: " + istek.TalepTipi)
	}
	if err != nil {
		return err
	}

	_, _, sonrakiTalep, _ := s.Audit.GetTalepMeta(istek.TalepTipi, istek.TalepID)
	sonrakiDetaylar, _ := s.captureOnaySonraki(istek.TalepTipi, projeID, sonrakiTalep, oncekiDetaylar)

	olaySonek := "onay"
	ozet := talepTipiEtiket(istek.TalepTipi) + " onaylandı"
	if istek.Karar == models.TalepReddedildi {
		olaySonek = "red"
		ozet = talepTipiEtiket(istek.TalepTipi) + " reddedildi"
	}

	detaylar := mergeDetayOncekiSonraki(oncekiDetaylar, sonrakiDetaylar)
	_, auditErr := s.Audit.RecordChange(models.DegisiklikKayitIstek{
		ProjeID:       projeID,
		OlayTipi:      istek.TalepTipi + "_" + olaySonek,
		KaynakTip:     "talep_" + istek.TalepTipi,
		KaynakID:      istek.TalepID,
		TalepNo:       talepNo,
		Ozet:          ozet,
		IslemiYapanID: istek.IslemiYapanID,
		Detaylar:      detaylar,
		VersiyonArtir: istek.Karar == models.TalepOnaylandi,
	})
	if auditErr != nil {
		log.Printf("Uyarı: talep onay audit kaydı yazılamadı (%s #%d): %v", istek.TalepTipi, istek.TalepID, auditErr)
	}
	return nil
}

// recordOlusturma, talep oluşturma audit kaydı yazar.
func (s *TalepService) recordOlusturma(tip string, projeID, kaynakID int, talepNo string, uyeID int, ozet string, payload interface{}) {
	if s.Audit == nil {
		return
	}
	_, err := s.Audit.RecordChange(models.DegisiklikKayitIstek{
		ProjeID:       projeID,
		OlayTipi:      tip + "_olusturma",
		KaynakTip:     "talep_" + tip,
		KaynakID:      kaynakID,
		TalepNo:       talepNo,
		Ozet:          ozet,
		IslemiYapanID: uyeID,
		Detaylar: []models.DegisiklikDetayIstek{{
			VarlikTip: "talep",
			VarlikID:  kaynakID,
			Onceki:    nil,
			Sonraki:   payload,
		}},
		VersiyonArtir: false,
	})
	if err != nil {
		log.Printf("Uyarı: talep oluşturma audit kaydı yazılamadı (%s #%d): %v", tip, kaynakID, err)
	}
}

// captureOnayOnceki, onay öncesi etkilenen varlıkları yakalar.
func (s *TalepService) captureOnayOnceki(tip string, projeID int, talep map[string]interface{}) ([]models.DegisiklikDetayIstek, error) {
	var detaylar []models.DegisiklikDetayIstek
	talepID := 0
	if id, ok := talep["id"].(int); ok {
		talepID = id
	}
	detaylar = append(detaylar, models.DegisiklikDetayIstek{VarlikTip: "talep", VarlikID: talepID, Onceki: talep})

	switch tip {
	case "ek_sure", "proje_iptali", "proje_dondurma", "bilgi_degisimi":
		proje, err := s.Audit.CaptureProje(projeID)
		if err != nil {
			return nil, err
		}
		detaylar = append(detaylar, models.DegisiklikDetayIstek{VarlikTip: "proje", VarlikID: projeID, Onceki: proje})
	case "ek_butce", "avans":
		kalem, _ := talep["butce_kalemi"].(string)
		if kalem != "" {
			butce, err := s.Audit.CaptureButceByKategori(projeID, kalem)
			if err != nil {
				return nil, err
			}
			kid := 0
			if v, ok := butce["kategori_id"].(int); ok {
				kid = v
			}
			detaylar = append(detaylar, models.DegisiklikDetayIstek{VarlikTip: "proje_butce", VarlikID: kid, Onceki: butce})
		}
	case "fasil_aktarimi":
		for _, key := range []string{"kaynak_kalem", "hedef_kalem"} {
			kalem, _ := talep[key].(string)
			if kalem == "" {
				continue
			}
			b, err := s.Audit.CaptureButceByKategori(projeID, kalem)
			if err != nil {
				return nil, err
			}
			kid := 0
			if v, ok := b["kategori_id"].(int); ok {
				kid = v
			}
			detaylar = append(detaylar, models.DegisiklikDetayIstek{VarlikTip: "proje_butce", VarlikID: kid, Onceki: b})
		}
	case "arastirmaci", "bursiyer":
		takim, err := s.Audit.CaptureTakim(projeID)
		if err != nil {
			return nil, err
		}
		detaylar = append(detaylar, models.DegisiklikDetayIstek{VarlikTip: "proje_takim", VarlikID: projeID, Onceki: takim})
	case "malzeme_guncelleme":
		butce, err := s.Audit.CaptureButceOzet(projeID)
		if err != nil {
			return nil, err
		}
		detaylar = append(detaylar, models.DegisiklikDetayIstek{VarlikTip: "proje_butce", VarlikID: projeID, Onceki: butce})
	}
	return detaylar, nil
}

// captureOnaySonraki, onay sonrası varlıkları yakalar.
func (s *TalepService) captureOnaySonraki(tip string, projeID int, talep map[string]interface{}, onceki []models.DegisiklikDetayIstek) ([]models.DegisiklikDetayIstek, error) {
	sonraki := make([]models.DegisiklikDetayIstek, 0, len(onceki))
	talepID := 0
	if id, ok := talep["id"].(int); ok {
		talepID = id
	}
	for _, d := range onceki {
		item := models.DegisiklikDetayIstek{VarlikTip: d.VarlikTip, VarlikID: d.VarlikID}
		switch d.VarlikTip {
		case "talep":
			item.Sonraki = talep
			item.VarlikID = talepID
		case "proje":
			p, err := s.Audit.CaptureProje(projeID)
			if err != nil {
				return nil, err
			}
			item.Sonraki = p
		case "proje_butce":
			if tip == "malzeme_guncelleme" {
				b, err := s.Audit.CaptureButceOzet(projeID)
				if err != nil {
					return nil, err
				}
				item.Sonraki = b
			} else if oncekiMap, ok := d.Onceki.(map[string]interface{}); ok {
				kalem, _ := oncekiMap["kategori_adi"].(string)
				b, err := s.Audit.CaptureButceByKategori(projeID, kalem)
				if err != nil {
					return nil, err
				}
				item.Sonraki = b
			}
		case "proje_takim":
			t, err := s.Audit.CaptureTakim(projeID)
			if err != nil {
				return nil, err
			}
			item.Sonraki = t
		}
		sonraki = append(sonraki, item)
	}
	return sonraki, nil
}

// mergeDetayOncekiSonraki, önce/sonra detay listelerini birleştirir.
func mergeDetayOncekiSonraki(onceki, sonraki []models.DegisiklikDetayIstek) []models.DegisiklikDetayIstek {
	out := make([]models.DegisiklikDetayIstek, 0, len(onceki))
	for i, o := range onceki {
		item := models.DegisiklikDetayIstek{
			VarlikTip: o.VarlikTip,
			VarlikID:  o.VarlikID,
			Onceki:    o.Onceki,
		}
		if i < len(sonraki) {
			item.Sonraki = sonraki[i].Sonraki
			if sonraki[i].VarlikID != 0 {
				item.VarlikID = sonraki[i].VarlikID
			}
		}
		out = append(out, item)
	}
	return out
}

// talepTipiEtiket, talep tipi için Türkçe etiket döner.
func talepTipiEtiket(tip string) string {
	labels := map[string]string{
		"ek_sure": "Ek süre", "ek_butce": "Ek bütçe", "fasil_aktarimi": "Fasıl aktarımı",
		"arastirmaci": "Araştırmacı", "bursiyer": "Bursiyer", "proje_iptali": "Proje iptali",
		"bilgi_degisimi": "Bilgi değişimi", "proje_dondurma": "Proje dondurma",
		"malzeme_guncelleme": "Malzeme güncelleme", "avans": "Avans",
	}
	if l, ok := labels[tip]; ok {
		return l
	}
	return tip
}

// maskKimlik, kimlik numarasının ortasını gizler.
func maskKimlik(kimlik string) string {
	if len(kimlik) < 5 {
		return "***"
	}
	masked := kimlik[:2]
	for i := 0; i < len(kimlik)-4; i++ {
		masked += "*"
	}
	return masked + kimlik[len(kimlik)-2:]
}
