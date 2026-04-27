package models

import "time"

// IsPaketi yapısı, veritabanındaki is_paketi tablosunun Go karşılığıdır.
type IsPaketi struct {
	PaketID          int    `json:"paket_id"`
	ProjeID          int    `json:"proje_id"`
	PaketAdi         string `json:"paket_adi"`
	PaketAmaci       string `json:"paket_amaci"`
	BaslangicTarihi  string `json:"baslangic_tarihi"`
	BitisTarihi      string `json:"bitis_tarihi"`
	OlusturmaTarihi  string `json:"olusturma_tarihi"`
	GuncellemeTarihi string `json:"guncelleme_tarihi"`
}

// ProjeDegerlendirme yapısı, bir projenin hakemler tarafından değerlendirmesini tutar.
type ProjeDegerlendirme struct {
	DegerlendirmeID  int       `json:"degerlendirme_id"`
	ProjeID          int       `json:"proje_id"`
	HakemID          int       `json:"hakem_id"`
	Puan             int       `json:"puan"`
	Yorum            string    `json:"yorum"`
	Durum            string    `json:"durum"` // Bekliyor, Onaylandı, Reddedildi, Revizyon
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
}

// HakemProjeOzet yapısı, hakem dashboard'ında gösterilecek atanmış projenin özetidir.
type HakemProjeOzet struct {
	ProjeID    int    `json:"proje_id"`
	BaslikTr   string `json:"baslik_tr"`
	BapTuru    string `json:"bap_turu"`
	DurumAdi   string `json:"durum_adi"`     // Projenin genel durumu
	HakemDurum string `json:"hakem_durum"`   // Bekliyor, Onaylandı...
	Puan       *int   `json:"puan"`
	Tarih      string `json:"tarih"`
}

// DegerlendirmeRequest yapısı, hakemin projeyi puanlama/yorumlama isteğidir.
type DegerlendirmeRequest struct {
	ProjeID int    `json:"proje_id" binding:"required"`
	Puan    int    `json:"puan" binding:"required"`
	Yorum   string `json:"yorum" binding:"required"`
	Durum   string `json:"durum" binding:"required"` // Onaylandı, Reddedildi, Revizyon
}
