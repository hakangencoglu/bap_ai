package models

import "time"

// KomisyonToplantisi komisyon toplantı verilerini temsil eder.
// Türkçe Yorum: Toplantı genel bilgileri, tarihi, gündemi ve kararları burada saklanır.
type KomisyonToplantisi struct {
	ToplantiID      int                        `json:"toplanti_id"`
	ToplantiNo      string                     `json:"toplanti_no"`
	Tarih           time.Time                  `json:"tarih"`
	Gundem          string                     `json:"gundem"`
	Karar           string                     `json:"karar"`
	OlusturanID     int                        `json:"olusturan_id"`
	OlusturmaTarihi time.Time                  `json:"olusturma_tarihi"`
	Katilimcilar    []*KomisyonToplantiKatilim `json:"katilimcilar"`
}

// KomisyonToplantiKatilim komisyon toplantısına katılan üyelerin katılım durumlarını temsil eder.
// Türkçe Yorum: Hangi komisyon üyesinin toplantıya katıldığı veya katılmadığı bu struct ile eşleştirilir.
type KomisyonToplantiKatilim struct {
	UyeID   int    `json:"uye_id"`
	Ad      string `json:"ad"`
	Soyad   string `json:"soyad"`
	Unvan   string `json:"unvan"`
	Bolum   string `json:"bolum"`
	Katildi bool   `json:"katildi"`
}
