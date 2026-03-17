package models

import "time"

// Butce yapısı, veritabanındaki butce tablosunun Go karşılığıdır.
type Butce struct {
	ItemID      int       `json:"item_id"`
	ProjeID     int       `json:"proje_id"`
	Tur         string    `json:"tur"`
	Aciklama    string    `json:"aciklama"`
	Gerekce     string    `json:"gerekce"`
	UrunTuru    string    `json:"urun_turu"`
	Adet        int       `json:"adet"`
	UrunFiyat   int       `json:"urun_fiyat"`
	ToplamFiyat int       `json:"toplam_fiyat"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
