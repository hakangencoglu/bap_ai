package models

import "time"

// Butce yapısı, veritabanındaki butce tablosunun Go karşılığıdır.
type Butce struct {
	KalemID          int       `json:"kalem_id"`
	ProjeID          int       `json:"proje_id"`
	KategoriID       *int      `json:"kategori_id"`          // Nullable FK → butce_kategori
	Aciklama         string    `json:"aciklama"`
	BirimOzelligi    int       `json:"birim_ozelligi"`
	BirimFiyat       float64   `json:"birim_fiyat"`
	ToplamFiyat      float64   `json:"toplam_fiyat"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
	// JOIN ile doldurulacak alan
	KategoriAdi string `json:"kategori_adi,omitempty"` // butce_kategori tablosundan
}
