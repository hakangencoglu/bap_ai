package models

import "time"

// Fakulte, üniversite bünyesindeki fakülte veya enstitü birimlerini temsil eder.
// Türkçe Yorum: Fakülte ID, tam adı, kodu ve aktiflik durumunu barındırır.
type Fakulte struct {
	FakulteID        int       `json:"fakulte_id"`
	FakulteAdi       string    `json:"fakulte_adi"`
	FakulteKodu      string    `json:"fakulte_kodu"`
	KisaAd           string    `json:"kisa_ad"`
	Aktif            bool      `json:"aktif"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
}

// Bolum, üniversite bünyesindeki akademik bölümleri temsil eder.
// Türkçe Yorum: Bölüm ID, adı, kodu ve aktiflik durumunu barındırır.
type Bolum struct {
	BolumID          int       `json:"bolum_id"`
	BolumAdi         string    `json:"bolum_adi"`
	BolumKodu        string    `json:"bolum_kodu"`
	KisaAd           string    `json:"kisa_ad"`
	Aktif            bool      `json:"aktif"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
}

// FakulteBolum, fakülte ile bölüm arasındaki çoka-çok / bire-çok ilişkiyi temsil eder.
// Türkçe Yorum: Bir fakülteye bağlı bölümleri eşleştiren ara ilişki tablosu modelidir.
type FakulteBolum struct {
	ID               int       `json:"id"`
	FakulteID        int       `json:"fakulte_id"`
	BolumID          int       `json:"bolum_id"`
	FakulteAdi       string    `json:"fakulte_adi,omitempty"`
	BolumAdi         string    `json:"bolum_adi,omitempty"`
	Aktif            bool      `json:"aktif"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
}
