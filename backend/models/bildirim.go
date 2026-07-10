package models

import "time"

// Bildirim yapısı, veritabanındaki bildirim tablosunun Go karşılığıdır.
// Türkçe Yorum: E-posta bildirimlerinin bir kopyasının sistem içinde saklanması ve kullanıcılara gösterilmesi için kullanılır.
type Bildirim struct {
	BildirimID      int       `json:"bildirim_id"`
	UyeID           int       `json:"uye_id"`
	Baslik          string    `json:"baslik"`
	Icerik          string    `json:"icerik"`
	Okundu          bool      `json:"okundu"`
	OlusturmaTarihi time.Time `json:"olusturma_tarihi"`
}
