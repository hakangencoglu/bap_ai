package service

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// ZamanlanmisGorevService zamanlanmış kural değerlendirme ve bildirim gönderme mantığını içerir.
// Türkçe Yorum: Kuralları proje tarihlerine göre eşleştirir, şablon parametrelerini doldurur ve E-Posta / SMS servislerini tetikler.
type ZamanlanmisGorevService struct {
	Repo          *repository.ZamanlanmisGorevRepository
	EpostaService *EpostaService
	SmsService    *SmsService
}

// NewZamanlanmisGorevService yeni bir ZamanlanmisGorevService oluşturur.
func NewZamanlanmisGorevService(repo *repository.ZamanlanmisGorevRepository, epostaService *EpostaService, smsService *SmsService) *ZamanlanmisGorevService {
	return &ZamanlanmisGorevService{
		Repo:          repo,
		EpostaService: epostaService,
		SmsService:    smsService,
	}
}

// ParseTemplate şablondaki dinamik etiketleri proje verileriyle değiştirir.
// Türkçe Yorum: {yurutucu_ad}, {proje_kodu}, {proje_baslik}, {kalan_gun}, {gecen_ay}, {bap_turu}, {tarih} etiketlerini çözümler.
func (s *ZamanlanmisGorevService) ParseTemplate(template string, p *models.KuralEslesenProje) string {
	now := time.Now()
	
	// Geçen ay ve kalan gün hesaplama
	gecenSureDays := int(now.Sub(p.BaslangicTarihi).Hours() / 24)
	if gecenSureDays < 0 {
		gecenSureDays = 0
	}
	gecenAy := int(math.Floor(float64(gecenSureDays) / 30.0))

	kalanSureDays := int(p.BitisTarihi.Sub(now).Hours() / 24)
	if kalanSureDays < 0 {
		kalanSureDays = 0
	}

	r := template
	r = strings.ReplaceAll(r, "{yurutucu_ad}", p.YurutucuAd)
	r = strings.ReplaceAll(r, "{proje_kodu}", p.ProjeKodu)
	r = strings.ReplaceAll(r, "{proje_baslik}", p.ProjeBaslik)
	r = strings.ReplaceAll(r, "{bap_turu}", p.BapTuru)
	r = strings.ReplaceAll(r, "{gecen_ay}", fmt.Sprintf("%d", gecenAy))
	r = strings.ReplaceAll(r, "{kalan_gun}", fmt.Sprintf("%d", kalanSureDays))
	r = strings.ReplaceAll(r, "{sure_ay}", fmt.Sprintf("%d", p.SureAy))
	r = strings.ReplaceAll(r, "{tarih}", now.Format("02.01.2006"))

	return r
}

// DoesProjectMatchRule projenin belirtilen kuralın zaman koşullarına uyup uymadığını kontrol eder.
func (s *ZamanlanmisGorevService) DoesProjectMatchRule(kural *models.ZamanlanmisGorevKural, p *models.KuralEslesenProje) bool {
	now := time.Now()

	switch kural.TetiklemeTipi {
	case "baslangic_sonrasi_ay":
		// Örn: Başlangıçtan 6 ay sonra
		gecenSureDays := int(now.Sub(p.BaslangicTarihi).Hours() / 24)
		gecenAy := int(math.Floor(float64(gecenSureDays) / 30.0))
		return gecenAy >= kural.ZamanDegeri

	case "bitim_oncesi_ay":
		// Örn: Bitişe 2 ay kala
		kalanSureDays := int(p.BitisTarihi.Sub(now).Hours() / 24)
		kalanAy := int(math.Ceil(float64(kalanSureDays) / 30.0))
		return kalanAy <= kural.ZamanDegeri && kalanSureDays > 0

	case "bitim_oncesi_gun":
		// Örn: Bitişe 30 gün kala
		kalanSureDays := int(p.BitisTarihi.Sub(now).Hours() / 24)
		return kalanSureDays <= kural.ZamanDegeri && kalanSureDays > 0

	case "periyodik_ay":
		// Örn: Her 3 ayda bir
		gecenSureDays := int(now.Sub(p.BaslangicTarihi).Hours() / 24)
		if gecenSureDays <= 0 {
			return false
		}
		gecenAy := int(math.Floor(float64(gecenSureDays) / 30.0))
		return gecenAy > 0 && (gecenAy%kural.ZamanDegeri == 0)

	default:
		return false
	}
}

// ProcessScheduledRules aktif olan tüm kuralları tarar ve uygun bildirimleri tetikler.
// Türkçe Yorum: Hem arka plan zamanlayıcısı (Cron) hem de Admin "Manuel Tetikle" butonu bu metodu çağırır.
func (s *ZamanlanmisGorevService) ProcessScheduledRules() (*models.ZamanlanmisGorevTetiklemeSonuc, error) {
	log.Printf("[ZAMANLANMIŞ-GÖREV] ProcessScheduledRules başlatıldı.")
	kurallar, err := s.Repo.GetAllRules()
	if err != nil {
		return nil, fmt.Errorf("kurallar alınamadı: %w", err)
	}

	sonuc := &models.ZamanlanmisGorevTetiklemeSonuc{}

	for _, kural := range kurallar {
		if !kural.AktifMi {
			continue
		}
		sonuc.ToplamIslenenKural++

		projeler, err := s.Repo.GetActiveProjectsForRules(kural.BapTuruID)
		if err != nil {
			log.Printf("[ZAMANLANMIŞ-GÖREV] KuralID %d için projeler çekilemedi: %v", kural.KuralID, err)
			sonuc.HataSayisi++
			continue
		}

		for _, proje := range projeler {
			sonuc.ToplamIslenenProje++

			if !s.DoesProjectMatchRule(kural, proje) {
				continue
			}

			alicilar := s.buildRecipients(kural, proje)
			for _, alici := range alicilar {
				s.sendEmailIfNeeded(kural, proje, alici, sonuc)
				s.sendSmsIfNeeded(kural, proje, alici, sonuc)
			}
		}
	}

	// TTO-İA-312 iş akış kuralı uyarınca süresi dolan hakem atamalarını kontrol et ve TTO'ya devret
	if expiredCount, err := s.ProcessExpiredHakemInvitations(); err == nil && expiredCount > 0 {
		log.Printf("[ZAMANLANMIŞ-GÖREV] TTO-İA-312 uyarınca %d adet süresi dolmuş hakem ataması TTO havuzuna aktarıldı.", expiredCount)
	}

	// BAP-200/300/400/500 projelerindeki 6 aylık ara rapor dönemi yaklaşan ve geciken yürütücülere otomatik bildirim ilet
	if reminderCount, err := s.ProcessAraRaporReminders(); err == nil && reminderCount > 0 {
		log.Printf("[ZAMANLANMIŞ-GÖREV] %d adet ara rapor hatırlatma/gecikme e-postası yürütücülere iletildi.", reminderCount)
	}

	log.Printf("[ZAMANLANMIŞ-GÖREV] İletim özeti: Kurallar: %d, Projeler: %d, E-Posta: %d, SMS: %d, Hata: %d",
		sonuc.ToplamIslenenKural, sonuc.ToplamIslenenProje, sonuc.GonderilenEposta, sonuc.GonderilenSms, sonuc.HataSayisi)

	return sonuc, nil
}

// ProcessExpiredHakemInvitations TTO-İA-312 iş akış kuralı gereğince zamanı geçen hakem davetlerini tespit edip TTO'ya bildirim ve otomatik iade yapar.
// Türkçe Yorum: Hakem değerlendirme süresi aşılırsa hakem ataması düşürülür ve proje yeni hakem atanmak üzere TTO sırasına verilir.
func (s *ZamanlanmisGorevService) ProcessExpiredHakemInvitations() (int, error) {
	query := `
		SELECT pd.proje_id, pd.hakem_id, p.proje_kodu, p.proje_baslik
		FROM proje_degerlendirmeleri pd
		JOIN proje p ON pd.proje_id = p.proje_id
		LEFT JOIN hakem_hakedis hh ON pd.proje_id = hh.proje_id AND pd.hakem_id = hh.hakem_uye_id
		WHERE pd.durum = 'Bekliyor'
		  AND (
			 (hh.son_teslim_tarihi IS NOT NULL AND hh.son_teslim_tarihi < CURRENT_TIMESTAMP)
			 OR (hh.son_teslim_tarihi IS NULL AND pd.olusturma_tarihi < CURRENT_TIMESTAMP - INTERVAL '15 days')
		  )
	`
	rows, err := s.Repo.DB.Query(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var projeID, hakemID int
		var projeKodu, projeBaslik string
		if err := rows.Scan(&projeID, &hakemID, &projeKodu, &projeBaslik); err == nil {
			// Hakem atama durumunu güncelle
			s.Repo.DB.Exec(`
				UPDATE proje_degerlendirmeleri 
				SET durum = 'Süre Doldu', atama_durumu = 'Süre Doldu', red_nedeni = 'TTO-İA-312: 15 günlük değerlendirme süresi aşıldı.' 
				WHERE proje_id = $1 AND hakem_id = $2
			`, projeID, hakemID)

			// Süreç geçmişine log ekle
			logAciklama := "TTO-İA-312 İş Akışı Uyarınca: Hakem değerlendirme süresi doldu. Proje yeni hakem atanması için TTO havuzuna iade edildi."
			s.Repo.DB.Exec(`
				INSERT INTO proje_surec_gecmisi (proje_id, islem_yapan_id, baslangic_durum, hedef_durum, aciklama)
				SELECT $1, $2, pd.durum_adi, 'hakem_atama_bekliyor', $3
				FROM proje p
				JOIN proje_durum pd ON p.durum_id = pd.durum_id
				WHERE p.proje_id = $1
			`, projeID, hakemID, logAciklama)

			// Proje durumunu hakem_atama_bekliyor yap
			s.Repo.DB.Exec(`
				UPDATE proje 
				SET durum_id = (SELECT durum_id FROM proje_durum WHERE durum_adi = 'hakem_atama_bekliyor' LIMIT 1) 
				WHERE proje_id = $1
			`, projeID)

			count++
		}
	}

	if count > 0 {
		log.Printf("[ZAMANLANMIŞ-GÖREV] %d adet süresi dolan hakem daveti TTO-İA-312 uyarınca iade edildi.", count)
	}
	return count, nil
}

// buildRecipients kural hedeflerine göre benzersiz alıcı listesi üretir.
func (s *ZamanlanmisGorevService) buildRecipients(kural *models.ZamanlanmisGorevKural, proje *models.KuralEslesenProje) []models.BildirimAliciKisi {
	hedefler := kural.AliciHedefleri
	if len(hedefler) == 0 {
		hedefler = []string{models.AliciHedefYurutucu}
	}

	seenEmails := map[string]bool{}
	seenPhones := map[string]bool{}
	var alicilar []models.BildirimAliciKisi

	addRecipient := func(item models.BildirimAliciKisi) {
		emailKey := strings.ToLower(strings.TrimSpace(item.Eposta))
		phoneKey := strings.TrimSpace(item.Telefon)
		if emailKey != "" && seenEmails[emailKey] {
			return
		}
		if phoneKey != "" && seenPhones[phoneKey] && emailKey == "" {
			return
		}
		if emailKey != "" {
			seenEmails[emailKey] = true
		}
		if phoneKey != "" {
			seenPhones[phoneKey] = true
		}
		alicilar = append(alicilar, item)
	}

	for _, hedef := range hedefler {
		switch hedef {
		case models.AliciHedefYurutucu:
			addRecipient(models.BildirimAliciKisi{
				HedefKey: models.AliciHedefYurutucu,
				AdSoyad:  proje.YurutucuAd,
				Eposta:   proje.YurutucuEposta,
				Telefon:  proje.YurutucuTelefon,
			})
		case models.AliciHedefTTO:
			ttoList, err := s.Repo.GetTTORecipients()
			if err != nil {
				log.Printf("[ZAMANLANMIŞ-GÖREV] TTO alıcıları alınamadı (Kural: %s): %v", kural.KuralAdi, err)
				continue
			}
			for _, tto := range ttoList {
				addRecipient(tto)
			}
		}
	}

	return alicilar
}

// sendEmailIfNeeded seçili alıcıya e-posta bildirimi gönderir.
func (s *ZamanlanmisGorevService) sendEmailIfNeeded(kural *models.ZamanlanmisGorevKural, proje *models.KuralEslesenProje, alici models.BildirimAliciKisi, sonuc *models.ZamanlanmisGorevTetiklemeSonuc) {
	if !kural.EpostaAktif || strings.TrimSpace(alici.Eposta) == "" {
		return
	}
	if s.Repo.IsAlreadySent(kural.KuralID, proje.ProjeID, "eposta", alici.Eposta) {
		return
	}

	epostaKonu := s.ParseTemplate(kural.EpostaKonu, proje)
	epostaIcerik := s.ParseTemplate(kural.EpostaSablon, proje)
	htmlBody := s.EpostaService.FormatEmailTemplate(
		fmt.Sprintf("Sayın %s,", alici.AdSoyad),
		epostaIcerik,
		proje.ProjeKodu,
		proje.ProjeBaslik,
		alici.AdSoyad,
		"Zamanlanmış Otomatik Bildirim",
		"",
	)

	err := s.EpostaService.SendEmailSMTP([]string{alici.Eposta}, epostaKonu, htmlBody)
	logStatus := "basarili"
	hataMsg := ""
	if err != nil {
		logStatus = "hata"
		hataMsg = err.Error()
		sonuc.HataSayisi++
		log.Printf("[ZAMANLANMIŞ-GÖREV] E-posta hatası (Proje: %s, Alıcı: %s): %v", proje.ProjeKodu, alici.Eposta, err)
	} else {
		sonuc.GonderilenEposta++
		log.Printf("[ZAMANLANMIŞ-GÖREV] E-posta gönderildi (Proje: %s, Alıcı: %s)", proje.ProjeKodu, alici.Eposta)
	}

	s.Repo.SaveLog(&models.ZamanlanmisGorevLog{
		KuralID:    kural.KuralID,
		ProjeID:    proje.ProjeID,
		Kanal:      "eposta",
		Alici:      alici.Eposta,
		Icerik:     epostaKonu,
		Durum:      logStatus,
		HataMesaji: hataMsg,
	})
}

// sendSmsIfNeeded seçili alıcıya SMS bildirimi gönderir.
func (s *ZamanlanmisGorevService) sendSmsIfNeeded(kural *models.ZamanlanmisGorevKural, proje *models.KuralEslesenProje, alici models.BildirimAliciKisi, sonuc *models.ZamanlanmisGorevTetiklemeSonuc) {
	if !kural.SmsAktif || strings.TrimSpace(alici.Telefon) == "" {
		return
	}
	if s.Repo.IsAlreadySent(kural.KuralID, proje.ProjeID, "sms", alici.Telefon) {
		return
	}

	smsIcerik := s.ParseTemplate(kural.SmsSablon, proje)
	err := s.SmsService.SendSMS(alici.Telefon, smsIcerik)
	logStatus := "simule_edildi"
	hataMsg := ""
	if err != nil {
		logStatus = "hata"
		hataMsg = err.Error()
		sonuc.HataSayisi++
	} else {
		sonuc.GonderilenSms++
	}

	s.Repo.SaveLog(&models.ZamanlanmisGorevLog{
		KuralID:    kural.KuralID,
		ProjeID:    proje.ProjeID,
		Kanal:      "sms",
		Alici:      alici.Telefon,
		Icerik:     smsIcerik,
		Durum:      logStatus,
		HataMesaji: hataMsg,
	})
}

// ProcessAraRaporReminders BAP-200/300/400/500 projelerindeki 6 aylık ara rapor teslim dönemi yaklaşan ve zamanı geçen yürütücülere otomatik bildirim gönderir.
// Türkçe Yorum: Teslim zamanına 15 gün kala hatırlatma e-postası, zamanı geçenler için ise uyarı e-postası iletir.
func (s *ZamanlanmisGorevService) ProcessAraRaporReminders() (int, error) {
	query := `
		SELECT 
			p.proje_id,
			p.proje_kodu,
			COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), p.proje_kodu) AS proje_baslik,
			COALESCE(u.unvan || ' ' || u.ad || ' ' || u.soyad, '') AS yurutucu_ad,
			COALESCE(u.eposta, '') AS yurutucu_eposta,
			COALESCE(pbt.bap_turu, 'BAP-200') AS bap_turu,
			ps.olusturma_tarihi AS baslangic_tarihi
		FROM proje p
		JOIN proje_durum pd ON p.durum_id = pd.durum_id
		LEFT JOIN uye u ON p.koordinator_id = u.uye_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN (
			SELECT proje_id, MAX(olusturma_tarihi) AS olusturma_tarihi
			FROM proje_surec_gecmisi
			WHERE hedef_durum = 'yururlukte'
			GROUP BY proje_id
		) ps ON p.proje_id = ps.proje_id
		WHERE pd.durum_adi = 'yururlukte'
		  AND ps.olusturma_tarihi IS NOT NULL
	`
	rows, err := s.Repo.DB.Query(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var projeID int
		var projeKodu, projeBaslik, yurutucuAd, yurutucuEposta, bapTuru string
		var baslangicTarihi time.Time

		if err := rows.Scan(&projeID, &projeKodu, &projeBaslik, &yurutucuAd, &yurutucuEposta, &bapTuru, &baslangicTarihi); err == nil {
			if strings.TrimSpace(yurutucuEposta) == "" {
				continue
			}

			monthsSinceStart := int(time.Since(baslangicTarihi).Hours() / (24 * 30))
			donem := (monthsSinceStart / 6) + 1
			if donem < 1 {
				donem = 1
			}
			sonTeslim := baslangicTarihi.AddDate(0, donem*6, 0)

			// Bu dönem için rapor yüklenmiş mi kontrol et
			var raporCount int
			s.Repo.DB.QueryRow(`SELECT COUNT(*) FROM proje_ara_rapor WHERE proje_id = $1 AND rapor_donemi = $2`, projeID, donem).Scan(&raporCount)
			if raporCount > 0 {
				continue
			}

			daysLeft := int(time.Until(sonTeslim).Hours() / 24)

			// 1. Hatırlatma: Teslim tarihine 15 gün kala
			if daysLeft > 0 && daysLeft <= 15 {
				konu := fmt.Sprintf("[BAP AI] Ara Rapor Teslim Zamanı Yaklaşıyor - Proje: %s", projeKodu)
				mesaj := fmt.Sprintf("%d. Ara Rapor teslim tarihinize %d gün kalmıştır. Lütfen raporunuzu sisteme yükleyiniz.", donem, daysLeft)
				if s.EpostaService != nil {
					htmlBody := s.EpostaService.FormatEmailTemplate(fmt.Sprintf("Sayın %s,", yurutucuAd), mesaj, projeKodu, projeBaslik, yurutucuAd, "Ara Rapor Hatırlatma", "6 Aylık Periyodik Ara Rapor Teslimi")
					s.EpostaService.SendEmailSMTP([]string{yurutucuEposta}, konu, htmlBody)
					count++
				}
			}

			// 2. Gecikme Uyarısı: Teslim tarihi geçmişse
			if daysLeft < 0 {
				konu := fmt.Sprintf("[BAP AI] UYARI: Ara Rapor Teslimi Gecikti - Proje: %s", projeKodu)
				mesaj := fmt.Sprintf("%d. Ara Rapor teslim tarihiniz (%s) geçmiştir. Lütfen en kısa sürede raporunuzu sisteme yükleyiniz.", donem, sonTeslim.Format("02.01.2006"))
				if s.EpostaService != nil {
					htmlBody := s.EpostaService.FormatEmailTemplate(fmt.Sprintf("Sayın %s,", yurutucuAd), mesaj, projeKodu, projeBaslik, yurutucuAd, "Ara Rapor Gecikti", "6 Aylık Periyodik Ara Rapor Gecikme Uyarısı")
					s.EpostaService.SendEmailSMTP([]string{yurutucuEposta}, konu, htmlBody)
					count++
				}
			}
		}
	}

	return count, nil
}
