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

			// E-Posta Bildirimi
			if kural.EpostaAktif && proje.YurutucuEposta != "" {
				if !s.Repo.IsAlreadySent(kural.KuralID, proje.ProjeID, "eposta") {
					epostaKonu := s.ParseTemplate(kural.EpostaKonu, proje)
					epostaIcerik := s.ParseTemplate(kural.EpostaSablon, proje)

					htmlBody := s.EpostaService.FormatEmailTemplate(
						fmt.Sprintf("Sayın %s,", proje.YurutucuAd),
						epostaIcerik,
						proje.ProjeKodu,
						proje.ProjeBaslik,
						proje.YurutucuAd,
						"Zamanlanmış Otomatik Bildirim",
						"",
					)

					err := s.EpostaService.SendEmailSMTP([]string{proje.YurutucuEposta}, epostaKonu, htmlBody)
					logStatus := "basarili"
					hataMsg := ""
					if err != nil {
						logStatus = "hata"
						hataMsg = err.Error()
						sonuc.HataSayisi++
						log.Printf("[ZAMANLANMIŞ-GÖREV] E-posta hatası (Proje: %s, Kural: %s): %v", proje.ProjeKodu, kural.KuralAdi, err)
					} else {
						sonuc.GonderilenEposta++
						log.Printf("[ZAMANLANMIŞ-GÖREV] E-posta gönderildi (Proje: %s, Alıcı: %s)", proje.ProjeKodu, proje.YurutucuEposta)
					}

					// Log kaydı
					s.Repo.SaveLog(&models.ZamanlanmisGorevLog{
						KuralID:    kural.KuralID,
						ProjeID:    proje.ProjeID,
						Kanal:      "eposta",
						Alici:      proje.YurutucuEposta,
						Icerik:     epostaKonu,
						Durum:      logStatus,
						HataMesaji: hataMsg,
					})
				}
			}

			// SMS Bildirimi
			if kural.SmsAktif && proje.YurutucuTelefon != "" {
				if !s.Repo.IsAlreadySent(kural.KuralID, proje.ProjeID, "sms") {
					smsIcerik := s.ParseTemplate(kural.SmsSablon, proje)

					err := s.SmsService.SendSMS(proje.YurutucuTelefon, smsIcerik)
					logStatus := "simule_edildi"
					hataMsg := ""
					if err != nil {
						logStatus = "hata"
						hataMsg = err.Error()
						sonuc.HataSayisi++
					} else {
						sonuc.GonderilenSms++
					}

					// Log kaydı
					s.Repo.SaveLog(&models.ZamanlanmisGorevLog{
						KuralID:    kural.KuralID,
						ProjeID:    proje.ProjeID,
						Kanal:      "sms",
						Alici:      proje.YurutucuTelefon,
						Icerik:     smsIcerik,
						Durum:      logStatus,
						HataMesaji: hataMsg,
					})
				}
			}
		}
	}

	log.Printf("[ZAMANLANMIŞ-GÖREV] İletim özeti: Kurallar: %d, Projeler: %d, E-Posta: %d, SMS: %d, Hata: %d",
		sonuc.ToplamIslenenKural, sonuc.ToplamIslenenProje, sonuc.GonderilenEposta, sonuc.GonderilenSms, sonuc.HataSayisi)

	return sonuc, nil
}
