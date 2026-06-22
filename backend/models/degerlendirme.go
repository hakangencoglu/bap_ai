package models

import "time"

// IsPaketi yapısı, veritabanındaki is_paketi tablosunun Go karşılığıdır.
type IsPaketi struct {
	PaketID     int    `json:"paket_id"`
	ProjeID     int    `json:"proje_id"`
	PaketAdi    string `json:"paket_adi"`
	PaketAmaci  string `json:"paket_amaci"`
	BaslangicAy int    `json:"baslangic_ay"`
	BitisAy     int    `json:"bitis_ay"`
}

// ProjeDegerlendirme yapısı, bir projenin hakemler tarafından değerlendirmesini tutar.
type ProjeDegerlendirme struct {
	DegerlendirmeID  int        `json:"degerlendirme_id"`
	ProjeID          int        `json:"proje_id"`
	HakemID          int        `json:"hakem_id"`
	Puan             int        `json:"puan"`
	Yorum            string     `json:"yorum"`
	Durum            string     `json:"durum"`           // Bekliyor, Onaylandı, Reddedildi, Revizyon (öğrenci görebilir)
	AtamaDurumu      string     `json:"atama_durumu"`    // Atandı, Kabul Edildi, Reddedildi (admin/akademisyen görebilir)
	RedNedeni        string     `json:"red_nedeni"`      // Hakem atamayı reddettiğinde sebebi
	KararTarihi      *time.Time `json:"karar_tarihi"`    // Atama kabul/red kararının tarihi
	OlusturmaTarihi  time.Time  `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time  `json:"guncelleme_tarihi"`
}

// HakemProjeOzet yapısı, hakem dashboard'ında gösterilecek atanmış projenin özetidir.
type HakemProjeOzet struct {
	ProjeID      int    `json:"proje_id"`
	BaslikTr     string `json:"baslik_tr"`
	BapTuru      string `json:"bap_turu"`
	DurumAdi     string `json:"durum_adi"`      // Projenin genel durumu
	HakemDurum   string `json:"hakem_durum"`     // Bekliyor, Onaylandı... (değerlendirme durumu)
	AtamaDurumu  string `json:"atama_durumu"`    // Atandı, Kabul Edildi, Reddedildi (atama durumu)
	Puan         *int   `json:"puan"`
	Tarih        string `json:"tarih"`
}

// DegerlendirmeRequest yapısı, hakemin projeyi puanlama/yorumlama isteğidir.
type DegerlendirmeRequest struct {
	ProjeID       int    `json:"proje_id" binding:"required"`
	Puan          int    `json:"puan" binding:"required"`
	Yorum         string `json:"yorum" binding:"required"`
	Durum         string `json:"durum" binding:"required"` // Onaylandı, Reddedildi, Revizyon
	RevizyonBolum string `json:"revizyon_bolum"`           // Revizyon talep edilen bölüm (ör. proje_bilgileri, proje_ekibi, butce_kalemleri, is_paketleri)
}

// HakemKararRequest yapısı, hakemin atamayı kabul veya reddetme isteğidir.
type HakemKararRequest struct {
	ProjeID   int    `json:"proje_id" binding:"required"`
	Karar     string `json:"karar" binding:"required"`     // "kabul" veya "red"
	RedNedeni string `json:"red_nedeni"`                    // Sadece red durumunda gerekli
}

// AdminHakemAtamaRequest yapısı, admin'in projeye hakem atama isteğidir.
type AdminHakemAtamaRequest struct {
	ProjeID int `json:"proje_id" binding:"required"`
	HakemID int `json:"hakem_id" binding:"required"`
}
