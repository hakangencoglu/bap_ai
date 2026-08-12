package service

import (
	"log"
	"time"
)

// ZamanlanmisGorevScheduler arka planda zamanlanmış kuralları periyodik olarak tarayan motordur.
// Türkçe Yorum: Goroutine ve time.Ticker kullanarak sistem arka planında otomatik bildirim döngüsünü yürütür.
type ZamanlanmisGorevScheduler struct {
	Service *ZamanlanmisGorevService
}

// NewZamanlanmisGorevScheduler yeni bir ZamanlanmisGorevScheduler oluşturur.
func NewZamanlanmisGorevScheduler(service *ZamanlanmisGorevService) *ZamanlanmisGorevScheduler {
	return &ZamanlanmisGorevScheduler{Service: service}
}

// StartScheduler periyodik olarak (ör. 6 saatte veya 24 saatte bir) kuralları çalıştıran arka plan döngüsünü başlatır.
func (s *ZamanlanmisGorevScheduler) StartScheduler(interval time.Duration) {
	go func() {
		log.Printf("[SCHEDULER] Zamanlanmış Görev Motoru başlatıldı (Periyot: %v)", interval)

		// Sunucu ilk açıldığında 10 saniye bekleyip ilk taramayı yap
		time.Sleep(10 * time.Second)
		if _, err := s.Service.ProcessScheduledRules(); err != nil {
			log.Printf("[SCHEDULER] İlk zamanlanmış görev çalıştırma hatası: %v", err)
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			log.Printf("[SCHEDULER] Periyodik zamanlanmış görev taraması çalışıyor...")
			if _, err := s.Service.ProcessScheduledRules(); err != nil {
				log.Printf("[SCHEDULER] Periyodik zamanlanmış görev hatası: %v", err)
			}
		}
	}()
}
