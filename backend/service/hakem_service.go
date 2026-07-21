package service

import (
	"fmt"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// HakemService, hakem işlemlerine ait iş mantığını yönetir
type HakemService struct {
	HakemRepo *repository.HakemRepository
	ProjeRepo *repository.ProjeRepository
	AdminRepo *repository.AdminRepository
}

func NewHakemService(hakemRepo *repository.HakemRepository, projeRepo *repository.ProjeRepository, adminRepo *repository.AdminRepository) *HakemService {
	return &HakemService{
		HakemRepo: hakemRepo,
		ProjeRepo: projeRepo,
		AdminRepo: adminRepo,
	}
}

// GetProjelerByHakem, bir hakeme ait atanan projeleri döndürür
func (s *HakemService) GetProjelerByHakem(hakemID int) ([]models.HakemProjeOzet, error) {
	return s.HakemRepo.GetProjelerByHakemID(hakemID)
}

// KabulRedKarar, hakemin atamayı kabul veya reddetme kararını işler
func (s *HakemService) KabulRedKarar(hakemID int, req models.HakemKararRequest) error {
	// Karar değeri doğrulanır
	var atamaDurumu string
	switch req.Karar {
	case "kabul":
		atamaDurumu = "Kabul Edildi"
		// Türkçe Yorum: Hakem, projeyi kabul edip değerlendirmeye başlamadan önce
		// gizlilik taahhütnamesini onaylamak zorundadır. Onaylanmadan kabul edilemez.
		if !req.TaahhutnameOnay {
			return fmt.Errorf("projeyi kabul edebilmek için hakem gizlilik taahhütnamesini onaylamanız gerekmektedir")
		}
	case "red":
		atamaDurumu = "Reddedildi"
		// Red durumunda neden gerekli
		if req.RedNedeni == "" {
			return fmt.Errorf("red durumunda red nedeni belirtilmelidir")
		}
	default:
		return fmt.Errorf("geçersiz karar değeri: %s (kabul veya red olmalı)", req.Karar)
	}

	if err := s.HakemRepo.UpdateAtamaKarar(hakemID, req.ProjeID, atamaDurumu, req.RedNedeni, req.TaahhutnameOnay); err != nil {
		return err
	}

	// Hakem kabul/red kararını süreç geçmişine logla
	aciklama := fmt.Sprintf("Hakem atama daveti %s.", atamaDurumu)
	if req.RedNedeni != "" {
		aciklama += " Red nedeni: " + req.RedNedeni
	}
	logQuery := `
		INSERT INTO proje_surec_gecmisi (proje_id, islem_yapan_id, baslangic_durum, hedef_durum, aciklama)
		SELECT $1, $2, pd.durum_adi, pd.durum_adi, $3
		FROM proje p
		JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE p.proje_id = $1
	`
	s.HakemRepo.DB.Exec(logQuery, req.ProjeID, hakemID, aciklama)

	// Türkçe Yorum: Eğer hakem daveti reddettiyse, proje durumunu tekrar 'hakem_atama_bekliyor' (TTO onay/atama sırası) yapıyoruz.
	if req.Karar == "red" {
		p, err := s.ProjeRepo.GetProjeByID(req.ProjeID)
		if err == nil && p != nil {
			logAciklama := fmt.Sprintf("Hakem atama daveti reddedildi. Red Nedeni: %s. Proje yeni hakem atanması için TTO'ya iade edildi.", req.RedNedeni)
			s.ProjeRepo.UpdateProjectStatusWithLog(req.ProjeID, hakemID, p.DurumAdi, "hakem_atama_bekliyor", logAciklama)
		}
	}

	return nil
}

// SubmitDegerlendirme, puanlamayı kaydeder, süreç geçmişine loglar ve gerekirse proje durumunu günceller.
func (s *HakemService) SubmitDegerlendirme(hakemID int, req models.DegerlendirmeRequest) error {
	// Puanı kaydet (sadece atamayı kabul etmiş hakemler değerlendirme yapabilir)
	if err := s.HakemRepo.SubmitDegerlendirme(hakemID, req); err != nil {
		return err
	}

	p, err := s.ProjeRepo.GetProjeByID(req.ProjeID)
	if err != nil {
		return fmt.Errorf("proje bulunamadı: %w", err)
	}

	// Revizyon durumunda proje statüsü değiştirilir, revizyon kaydı oluşturulur ve erken dönülür
	if req.Durum == "Revizyon" {
		err = s.ProjeRepo.UpdateProjectStatusWithLog(req.ProjeID, hakemID, p.DurumAdi, "revizyon", req.Yorum)
		if err != nil {
			return fmt.Errorf("proje durumu güncellenemedi: %w", err)
		}

		var atananID *int
		if p.KoordinatorID != nil {
			atananID = p.KoordinatorID
		}

		insertQuery := `
			INSERT INTO revizyonlar (proje_id, olusturan_kisi_id, atanan_kisi_id, aciklama, durum, revizyon_bolum)
			VALUES ($1, $2, $3, $4, 'Bekliyor', $5)
		`
		_, err = s.HakemRepo.DB.Exec(insertQuery, req.ProjeID, hakemID, atananID, req.Yorum, req.RevizyonBolum)
		if err != nil {
			return fmt.Errorf("revizyon kaydı oluşturulamadı: %w", err)
		}
		return nil
	}

	// Türkçe Yorum: Reddedildi durumunda proje statüsü doğrudan 'hakem_atama_bekliyor' (TTO) yapılarak iade edilir ve süreç sonlandırılır.
	if req.Durum == "Reddedildi" {
		aciklama := fmt.Sprintf("Hakem değerlendirmesi tamamlandı: Reddedildi. Puan: %d. Yorum: %s. Karar verilmesi için proje TTO'ya iade edildi.", req.Puan, req.Yorum)
		err = s.ProjeRepo.UpdateProjectStatusWithLog(req.ProjeID, hakemID, p.DurumAdi, "hakem_atama_bekliyor", aciklama)
		if err != nil {
			return fmt.Errorf("proje durumu güncellenemedi: %w", err)
		}
		return nil
	}

	// Onaylandı / Reddedildi durumunda hakem kararını süreç geçmişine logla
	aciklama := fmt.Sprintf("Hakem değerlendirmesi tamamlandı: %s. Puan: %d", req.Durum, req.Puan)
	if req.Yorum != "" {
		aciklama += ". Yorum: " + req.Yorum
	}
	logQuery := `
		INSERT INTO proje_surec_gecmisi (proje_id, islem_yapan_id, baslangic_durum, hedef_durum, aciklama)
		VALUES ($1, $2, $3, $3, $4)
	`
	s.HakemRepo.DB.Exec(logQuery, req.ProjeID, hakemID, p.DurumAdi, aciklama)

	// Tüm kabul edilmiş hakemlerin değerlendirmelerini kontrol et
	degerlendirmeler, err := s.HakemRepo.GetAllDegerlendirmeByProjeID(req.ProjeID)
	if err != nil {
		return nil
	}

	// Sadece atamayı kabul etmiş hakemler sayılır; bunlar arasında bekleyen var mı?
	herhangiKabulEdilenVar := false
	kabulEdilenBekliyor := false
	hepsiOnayladi := true
	dusukPuanVar := false
	for _, d := range degerlendirmeler {
		if d.AtamaDurumu == "Kabul Edildi" {
			herhangiKabulEdilenVar = true
			if d.Durum == "Bekliyor" {
				kabulEdilenBekliyor = true
				break
			}
			if d.Durum != "Onaylandı" {
				hepsiOnayladi = false
			}
			// Türkçe Yorum: 70 puanın altı düşük puan olarak kabul edilir
			if d.Puan < 70 {
				dusukPuanVar = true
			}
		}
	}

	// Tüm kabul eden hakemler değerlendirmesini bitirdiyse projeyi ilerlet
	if herhangiKabulEdilenVar && !kabulEdilenBekliyor {
		yeniDurum := "hakem_onayladi"
		ilerlemeAciklamasi := "Tüm hakem değerlendirmeleri tamamlandı. Proje TTO sevk onayına sunuldu."
		
		// Türkçe Yorum: Hakemlerden biri onaylamazsa veya düşük puan verirse proje doğrudan reddedilmez; karar için TTO'ya iade edilir.
		if !hepsiOnayladi || dusukPuanVar {
			yeniDurum = "hakem_atama_bekliyor"
			ilerlemeAciklamasi = "Tüm hakem değerlendirmeleri tamamlandı. Bir veya daha fazla hakem projeyi onaylamadı veya düşük puan verdi. Karar verilmesi için proje TTO'ya iade edildi."
		}
		s.ProjeRepo.UpdateProjectStatusWithLog(req.ProjeID, hakemID, p.DurumAdi, yeniDurum, ilerlemeAciklamasi)
	}

	return nil
}

// IsHakemAssigned, hakemin projeye atanıp atanmadığını sorgular.
// Türkçe Yorum: Hakemin projeyi görme/değerlendirme yetkisi olup olmadığını kontrol eder.
func (s *HakemService) IsHakemAssigned(hakemID int, projeID int) (bool, error) {
	return s.HakemRepo.IsHakemAssigned(hakemID, projeID)
}

// GetProjectDetailsForHakem, hakemin projenin tüm detaylarını görmesini sağlar.
// Türkçe Yorum: Hakem detay sayfası için yetki parametresi false olarak iletilir (hakem adları maskelenir).
func (s *HakemService) GetProjectDetailsForHakem(projeID int) (*repository.ProjectDetail, error) {
	return s.AdminRepo.GetProjectDetailsForAdmin(projeID, false)
}

// GetDegerlendirmeQuestions, hakem değerlendirme başlıklarını ve alt sorularını döner.
// Türkçe Yorum: Formun dinamik oluşması için gerekli başlık ve soruları repository'den çeker.
func (s *HakemService) GetDegerlendirmeQuestions() ([]models.HakemDegerlendirmeBaslik, error) {
	return s.HakemRepo.GetDegerlendirmeQuestions()
}
