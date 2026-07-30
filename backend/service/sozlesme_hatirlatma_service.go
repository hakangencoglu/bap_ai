package service

import (
	"fmt"
	"log"
	"time"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// SozlesmeHatirlatmaService, proje sözleşmelerinin aylık süre takibini ve yürütücülere e-posta hatırlatmasını yönetir.
// Türkçe Yorum: Sözleşme imzalandıktan sonra her ay geçen ve kalan süre bilgileri hesaplanıp yürütücüye iletilir.
type SozlesmeHatirlatmaService struct {
	SozlesmeRepo  *repository.SozlesmeRepository
	EpostaService *EpostaService
}

// NewSozlesmeHatirlatmaService yeni bir SozlesmeHatirlatmaService örneği döner.
// Türkçe Yorum: Servis yapısını başlatmak için kullanılan kurucu fonksiyon.
func NewSozlesmeHatirlatmaService(sozlesmeRepo *repository.SozlesmeRepository, epostaService *EpostaService) *SozlesmeHatirlatmaService {
	return &SozlesmeHatirlatmaService{
		SozlesmeRepo:  sozlesmeRepo,
		EpostaService: epostaService,
	}
}

// CalculateElapsedTime başlangıç tarihi ile verilen zaman arasındaki geçen süreyi Türkçe olarak hesaplar.
// Türkçe Yorum: Başlangıç tarihinden bugüne kadar geçen yılı, ayı ve günü "X Ay Y Gün" formatında döner.
func CalculateElapsedTime(baslangic time.Time, refDate time.Time) string {
	if refDate.Before(baslangic) {
		return "Henüz başlamadı"
	}

	years := refDate.Year() - baslangic.Year()
	months := int(refDate.Month()) - int(baslangic.Month())
	days := refDate.Day() - baslangic.Day()

	if days < 0 {
		prevMonth := refDate.AddDate(0, -1, 0)
		daysInPrevMonth := time.Date(prevMonth.Year(), prevMonth.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		days += daysInPrevMonth
		months--
	}
	if months < 0 {
		months += 12
		years--
	}

	totalMonths := years*12 + months

	if totalMonths > 0 && days > 0 {
		return fmt.Sprintf("%d Ay %d Gün", totalMonths, days)
	} else if totalMonths > 0 {
		return fmt.Sprintf("%d Ay", totalMonths)
	} else {
		return fmt.Sprintf("%d Gün", days)
	}
}

// CalculateRemainingTime bitiş tarihine kadar kalan süreyi Türkçe olarak hesaplar.
// Türkçe Yorum: Referans tarihten bitiş tarihine kadar kalan süreyi "X Ay Y Gün" formatında döner.
func CalculateRemainingTime(bitis time.Time, refDate time.Time) string {
	if refDate.After(bitis) {
		return "Süre doldu"
	}

	years := bitis.Year() - refDate.Year()
	months := int(bitis.Month()) - int(refDate.Month())
	days := bitis.Day() - refDate.Day()

	if days < 0 {
		prevMonth := bitis.AddDate(0, -1, 0)
		daysInPrevMonth := time.Date(prevMonth.Year(), prevMonth.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		days += daysInPrevMonth
		months--
	}
	if months < 0 {
		months += 12
		years--
	}

	totalMonths := years*12 + months

	if totalMonths > 0 && days > 0 {
		return fmt.Sprintf("%d Ay %d Gün", totalMonths, days)
	} else if totalMonths > 0 {
		return fmt.Sprintf("%d Ay", totalMonths)
	} else {
		return fmt.Sprintf("%d Gün", days)
	}
}

// GetCurrentMonthIndex sözleşme başlangıcından itibaren geçen tam ay sayısını döner.
// Türkçe Yorum: Hatırlatmanın kaçıncı ay dönemi olduğunu tespit etmek için kullanılır.
func GetCurrentMonthIndex(baslangic time.Time, refDate time.Time) int {
	if refDate.Before(baslangic) {
		return 0
	}
	years := refDate.Year() - baslangic.Year()
	months := int(refDate.Month()) - int(baslangic.Month())
	if refDate.Day() < baslangic.Day() {
		months--
	}
	return years*12 + months
}

// IsMonthlyReminderDue sözleşmenin aylık e-posta hatırlatma zamanının gelip gelmediğini kontrol eder.
// Türkçe Yorum: Sözleşme başlayalı en az 1 ay geçmişse ve bu ay dönemi için henüz e-posta atılmamışsa true döner.
func IsMonthlyReminderDue(info models.ProjeSozlesmeHatirlatmaInfo, refDate time.Time) bool {
	if info.BaslangicTarihi.IsZero() || refDate.Before(info.BaslangicTarihi) {
		return false
	}

	currentMonthIndex := GetCurrentMonthIndex(info.BaslangicTarihi, refDate)
	if currentMonthIndex < 1 {
		return false
	}

	if info.SonHatirlatmaTarihi == nil {
		return true
	}

	// Son hatırlatmanın üzerinden en az 25 gün geçmiş olmalı ve dönem indeksi artmış olmalı
	daysSinceLast := refDate.Sub(*info.SonHatirlatmaTarihi).Hours() / 24
	if currentMonthIndex > info.SonDonemIndeks && daysSinceLast >= 25 {
		return true
	}

	return false
}

// CheckAndSendMonthlyReminders tüm aktif sözleşmeleri tarar ve zamanı gelen hatırlatma e-postalarını iletir.
// Türkçe Yorum: Veritabanındaki aktif sözleşmeleri getirir, aylık e-postaları gönderir ve sonucu loglar.
func (s *SozlesmeHatirlatmaService) CheckAndSendMonthlyReminders() (int, error) {
	log.Println("[SÖZLEŞME-HATIRLATMA] Aylık süre hatırlatma taraması başlatıldı...")

	contracts, err := s.SozlesmeRepo.GetActiveSignedContractsForReminder()
	if err != nil {
		return 0, fmt.Errorf("sözleşmeler alınırken hata: %w", err)
	}

	now := time.Now()
	sentCount := 0

	for _, info := range contracts {
		if IsMonthlyReminderDue(info, now) {
			gecenSure := CalculateElapsedTime(info.BaslangicTarihi, now)
			kalanSure := CalculateRemainingTime(info.BitisTarihi, now)
			donemIndeks := GetCurrentMonthIndex(info.BaslangicTarihi, now)
			if donemIndeks < 1 {
				donemIndeks = 1
			}

			err := s.EpostaService.SendContractMonthlyReminderEmail(info, gecenSure, kalanSure)
			durum := "gonderildi"
			if err != nil {
				log.Printf("[SÖZLEŞME-HATIRLATMA] E-posta gönderilemedi (%s): %v", info.ProjeKodu, err)
				durum = "hata"
			} else {
				sentCount++
			}

			logItem := &models.SozlesmeHatirlatmaLog{
				SozlesmeID:       info.SozlesmeID,
				ProjeID:          info.ProjeID,
				GecenSure:        gecenSure,
				KalanSure:        kalanSure,
				GonderilenEposta: info.YurutucuEposta,
				DonemIndeks:      donemIndeks,
				Durum:            durum,
			}

			if logErr := s.SozlesmeRepo.SaveReminderLog(logItem); logErr != nil {
				log.Printf("[SÖZLEŞME-HATIRLATMA] Log kaydı atılamadı (%s): %v", info.ProjeKodu, logErr)
			}
		}
	}

	log.Printf("[SÖZLEŞME-HATIRLATMA] Tarama tamamlandı. Toplam gönderilen e-posta sayısı: %d", sentCount)
	return sentCount, nil
}

// StartHatirlatmaScheduler arka planda periyodik olarak süre kontrolünü çalıştırır.
// Türkçe Yorum: Uygulama ayağa kalktığında Goroutine içerisinde periyodik zamanlayıcı başlatır.
func (s *SozlesmeHatirlatmaService) StartHatirlatmaScheduler(interval time.Duration) {
	log.Printf("[SÖZLEŞME-HATIRLATMA] Arka plan zamanlayıcı başlatıldı (Periyot: %v)", interval)
	go func() {
		// Başlangıçta 10 saniye bekle ve ilk kontrolü yap
		time.Sleep(10 * time.Second)
		if _, err := s.CheckAndSendMonthlyReminders(); err != nil {
			log.Printf("[SÖZLEŞME-HATIRLATMA] Zamanlayıcı ilk çalıştırma hatası: %v", err)
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if _, err := s.CheckAndSendMonthlyReminders(); err != nil {
				log.Printf("[SÖZLEŞME-HATIRLATMA] Periyodik tarama hatası: %v", err)
			}
		}
	}()
}
