package models

import "time"

// ProjeSozlesme, BAP Destek Programı Proje Sözleşmesi veri modelidir.
// Türkçe Yorum: Akademisyen tarafından sözleşme aşamasında doldurulan resmi bilgileri saklar.
type ProjeSozlesme struct {
	ID               int        `json:"id"`
	ProjeID          int        `json:"proje_id"`
	UyeID            int        `json:"uye_id"`
	TCKimlik         string     `json:"tc_kimlik"`
	YurutucuAdres    string     `json:"yurutucu_adres"`
	YurutucuTelefon  string     `json:"yurutucu_telefon"`
	YurutucuEposta   string     `json:"yurutucu_eposta"`
	BaslangicTarihi  string     `json:"baslangic_tarihi"`
	BitisTarihi      string     `json:"bitis_tarihi"`
	Durum            string     `json:"durum"`
	IndirildiMi      bool       `json:"indirildi_mi"`   // Sözleşme PDF'i indirildi mi (tek seferlik indirme kontrolü)
	IndirmeTarihi    *time.Time `json:"indirme_tarihi"` // PDF'in indirildiği tarih
	OlusturmaTarihi  time.Time  `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time  `json:"guncelleme_tarihi"`

	// İlişkisel veri alanları
	ProjeKodu   string `json:"proje_kodu,omitempty"`
	ProjeBaslik string `json:"proje_baslik,omitempty"`
	YurutucuAd  string `json:"yurutucu_ad,omitempty"`
}

// ProjeSozlesmeHatirlatmaInfo, e-posta hatırlatması gönderilecek aktif sözleşmelerin özet verisidir.
// Türkçe Yorum: Arka plan servisinde aylık süre kontrolleri yapılırken kullanılır.
type ProjeSozlesmeHatirlatmaInfo struct {
	SozlesmeID           int        `json:"sozlesme_id"`
	ProjeID              int        `json:"proje_id"`
	ProjeKodu            string     `json:"proje_kodu"`
	ProjeBaslik          string     `json:"proje_baslik"`
	YurutucuAd           string     `json:"yurutucu_ad"`
	YurutucuEposta       string     `json:"yurutucu_eposta"`
	BaslangicTarihi      time.Time  `json:"baslangic_tarihi"`
	BitisTarihi          time.Time  `json:"bitis_tarihi"`
	SureAy               int        `json:"sure_ay"`
	SonHatirlatmaTarihi *time.Time `json:"son_hatirlatma_tarihi,omitempty"`
	SonDonemIndeks       int        `json:"son_donem_indeks,omitempty"`
}

// SozlesmeHatirlatmaLog, gönderilen sözleşme e-posta hatırlatma kaydını saklar.
// Türkçe Yorum: Veritabanına kaydedilen aylık hatırlatma e-posta log modelidir.
type SozlesmeHatirlatmaLog struct {
	ID               int       `json:"id"`
	SozlesmeID       int       `json:"sozlesme_id"`
	ProjeID          int       `json:"proje_id"`
	GonderimTarihi   time.Time `json:"gonderim_tarihi"`
	GecenSure        string    `json:"gecen_sure"`
	KalanSure        string    `json:"kalan_sure"`
	GonderilenEposta string    `json:"gonderilen_eposta"`
	DonemIndeks      int       `json:"donem_indeks"`
	Durum            string    `json:"durum"`
}

