package models

import "time"

// ProjeSozlesme, BAP Destek Programı Proje Sözleşmesi veri modelidir.
// Türkçe Yorum: Akademisyen tarafından sözleşme aşamasında doldurulan resmi bilgileri saklar.
type ProjeSozlesme struct {
	ID              int       `json:"id"`
	ProjeID         int       `json:"proje_id"`
	UyeID           int       `json:"uye_id"`
	TCKimlik        string    `json:"tc_kimlik"`
	YurutucuAdres   string    `json:"yurutucu_adres"`
	YurutucuTelefon string    `json:"yurutucu_telefon"`
	YurutucuEposta  string    `json:"yurutucu_eposta"`
	BaslangicTarihi string    `json:"baslangic_tarihi"`
	BitisTarihi     string    `json:"bitis_tarihi"`
	Durum           string    `json:"durum"`
	OlusturmaTarihi time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`

	// İlişkisel veri alanları
	ProjeKodu   string `json:"proje_kodu,omitempty"`
	ProjeBaslik string `json:"proje_baslik,omitempty"`
	YurutucuAd  string `json:"yurutucu_ad,omitempty"`
}
