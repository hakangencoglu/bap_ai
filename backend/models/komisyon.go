package models

import "time"

// ProjeKomisyonOnay komisyon üyelerinin projeler hakkındaki tekil onay ve karar durumlarını tutar.
// Türkçe Yorum: Her bir komisyon üyesinin projeye verdiği oylamayı temsil eder.
type ProjeKomisyonOnay struct {
	OnayID           int       `json:"onay_id"`
	ProjeID          int       `json:"proje_id"`
	KomisyonUyeID    int       `json:"komisyon_uye_id"`
	Karar            string    `json:"karar"` // bekliyor, onayla, reddet, revizyon
	Aciklama         string    `json:"aciklama"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
}
