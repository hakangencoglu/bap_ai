package models

import "time"

// Revizyon yapısı, bir projedeki revizyon detaylarını ve durumunu temsil eder.
type Revizyon struct {
	RevizyonID      int       `json:"revizyon_id"`
	ProjeID         int       `json:"proje_id"`
	OlusturanKisiID int       `json:"olusturan_kisi_id"`
	AtananKisiID    *int      `json:"atanan_kisi_id"` // Nullable olduğu için pointer kullanıyoruz
	Aciklama        string    `json:"aciklama" binding:"required"`
	Durum           string    `json:"durum"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
