package models

import "time"

// TalepDurum sabitleri: Tüm talep tablolarında kullanılır.
const (
	TalepBeklemede  = "beklemede"
	TalepOnaylandi  = "onaylandi"
	TalepReddedildi = "reddedildi"
)

// TalepOzet, Admin/TTO listeleme için birleşik talep özeti.
type TalepOzet struct {
	ID               int                    `json:"id"`
	TalepNo          string                 `json:"talep_no"`
	TalepTipi        string                 `json:"talep_tipi"`
	TalepTipiEtiketi string                 `json:"talep_tipi_etiketi"`
	ProjeID          int                    `json:"proje_id"`
	ProjeKodu        string                 `json:"proje_kodu"`
	ProjeBaslik      string                 `json:"proje_baslik"`
	UyeID            int                    `json:"uye_id"`
	TalepEdenAd      string                 `json:"talep_eden_ad"`
	Durum            string                 `json:"durum"`
	Gerekce          string                 `json:"gerekce"`
	OlusturmaTarihi  time.Time              `json:"olusturma_tarihi"`
	Detay            map[string]interface{} `json:"detay"`
}

// TalepOnayIstek, Admin/TTO onay veya red işlemi için istek modeli.
type TalepOnayIstek struct {
	TalepTipi string `json:"talep_tipi" binding:"required"`
	TalepID   int    `json:"talep_id"   binding:"required"`
	Karar     string `json:"karar"      binding:"required"`
	RedNotu   string `json:"red_notu"`
}

// TalepEkSure, ek süre talebi modelidir.
type TalepEkSure struct {
	ID               int       `json:"id"`
	ProjeID          int       `json:"proje_id"`
	UyeID            int       `json:"uye_id"`
	TalepNo          string    `json:"talep_no"`
	EkSureAy         int       `json:"ek_sure_ay"`
	Gerekce          string    `json:"gerekce"`
	Durum            string    `json:"durum"`
	RedNotu          string    `json:"red_notu"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
	ProjeKodu        string    `json:"proje_kodu,omitempty"`
	ProjeBaslik      string    `json:"proje_baslik,omitempty"`
	TalepEdenAd      string    `json:"talep_eden_ad,omitempty"`
}

// TalepEkButce, ek bütçe talebi modelidir.
type TalepEkButce struct {
	ID               int       `json:"id"`
	ProjeID          int       `json:"proje_id"`
	UyeID            int       `json:"uye_id"`
	TalepNo          string    `json:"talep_no"`
	ButceKalemi      string    `json:"butce_kalemi"`
	TutarTL          float64   `json:"tutar_tl"`
	Gerekce          string    `json:"gerekce"`
	Durum            string    `json:"durum"`
	RedNotu          string    `json:"red_notu"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
	ProjeKodu        string    `json:"proje_kodu,omitempty"`
	ProjeBaslik      string    `json:"proje_baslik,omitempty"`
	TalepEdenAd      string    `json:"talep_eden_ad,omitempty"`
}

// TalepFasilAktarimi, fasıl aktarımı talebi modelidir.
type TalepFasilAktarimi struct {
	ID               int       `json:"id"`
	ProjeID          int       `json:"proje_id"`
	UyeID            int       `json:"uye_id"`
	TalepNo          string    `json:"talep_no"`
	KaynakKalem      string    `json:"kaynak_kalem"`
	HedefKalem       string    `json:"hedef_kalem"`
	TutarTL          float64   `json:"tutar_tl"`
	Gerekce          string    `json:"gerekce"`
	Durum            string    `json:"durum"`
	RedNotu          string    `json:"red_notu"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
	ProjeKodu        string    `json:"proje_kodu,omitempty"`
	ProjeBaslik      string    `json:"proje_baslik,omitempty"`
	TalepEdenAd      string    `json:"talep_eden_ad,omitempty"`
}

// TalepArastirmaci, araştırmacı ekleme/çıkarma talebi modelidir.
type TalepArastirmaci struct {
	ID               int       `json:"id"`
	ProjeID          int       `json:"proje_id"`
	UyeID            int       `json:"uye_id"`
	TalepNo          string    `json:"talep_no"`
	IslemTuru        string    `json:"islem_turu"`
	ArastirmaciAdi   string    `json:"arastirmaci_adi"`
	Gerekce          string    `json:"gerekce"`
	Durum            string    `json:"durum"`
	RedNotu          string    `json:"red_notu"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
	ProjeKodu        string    `json:"proje_kodu,omitempty"`
	ProjeBaslik      string    `json:"proje_baslik,omitempty"`
	TalepEdenAd      string    `json:"talep_eden_ad,omitempty"`
}

// TalepBursiyer, bursiyer işlem talebi modelidir.
type TalepBursiyer struct {
	ID               int       `json:"id"`
	ProjeID          int       `json:"proje_id"`
	UyeID            int       `json:"uye_id"`
	TalepNo          string    `json:"talep_no"`
	BursiyerKimlik   string    `json:"bursiyer_kimlik"`
	BursiyerAdi      string    `json:"bursiyer_adi"`
	IslemTuru        string    `json:"islem_turu"`
	Gerekce          string    `json:"gerekce"`
	Durum            string    `json:"durum"`
	RedNotu          string    `json:"red_notu"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
	ProjeKodu        string    `json:"proje_kodu,omitempty"`
	ProjeBaslik      string    `json:"proje_baslik,omitempty"`
	TalepEdenAd      string    `json:"talep_eden_ad,omitempty"`
}

// TalepProjeIptali, proje iptali talebi modelidir.
type TalepProjeIptali struct {
	ID               int       `json:"id"`
	ProjeID          int       `json:"proje_id"`
	UyeID            int       `json:"uye_id"`
	TalepNo          string    `json:"talep_no"`
	Gerekce          string    `json:"gerekce"`
	Durum            string    `json:"durum"`
	RedNotu          string    `json:"red_notu"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
	ProjeKodu        string    `json:"proje_kodu,omitempty"`
	ProjeBaslik      string    `json:"proje_baslik,omitempty"`
	TalepEdenAd      string    `json:"talep_eden_ad,omitempty"`
}

// TalepBilgiDegisimi, proje bilgi değişimi talebi modelidir.
type TalepBilgiDegisimi struct {
	ID               int       `json:"id"`
	ProjeID          int       `json:"proje_id"`
	UyeID            int       `json:"uye_id"`
	TalepNo          string    `json:"talep_no"`
	DegisiklikTanimi string    `json:"degisiklik_tanimi"`
	Gerekce          string    `json:"gerekce"`
	Durum            string    `json:"durum"`
	RedNotu          string    `json:"red_notu"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
	ProjeKodu        string    `json:"proje_kodu,omitempty"`
	ProjeBaslik      string    `json:"proje_baslik,omitempty"`
	TalepEdenAd      string    `json:"talep_eden_ad,omitempty"`
}

// TalepProjeDondurma, proje dondurma talebi modelidir.
type TalepProjeDondurma struct {
	ID               int       `json:"id"`
	ProjeID          int       `json:"proje_id"`
	UyeID            int       `json:"uye_id"`
	TalepNo          string    `json:"talep_no"`
	DondurmaAy       int       `json:"dondurma_sure_ay"`
	Gerekce          string    `json:"gerekce"`
	Durum            string    `json:"durum"`
	RedNotu          string    `json:"red_notu"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
	ProjeKodu        string    `json:"proje_kodu,omitempty"`
	ProjeBaslik      string    `json:"proje_baslik,omitempty"`
	TalepEdenAd      string    `json:"talep_eden_ad,omitempty"`
}

// TalepMalzemeGuncelleme, malzeme güncelleme talebi modelidir.
type TalepMalzemeGuncelleme struct {
	ID               int       `json:"id"`
	ProjeID          int       `json:"proje_id"`
	UyeID            int       `json:"uye_id"`
	TalepNo          string    `json:"talep_no"`
	GuncellemeTanimi string    `json:"guncelleme_tanimi"`
	Gerekce          string    `json:"gerekce"`
	Durum            string    `json:"durum"`
	RedNotu          string    `json:"red_notu"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
	ProjeKodu        string    `json:"proje_kodu,omitempty"`
	ProjeBaslik      string    `json:"proje_baslik,omitempty"`
	TalepEdenAd      string    `json:"talep_eden_ad,omitempty"`
}

// TalepAvans, avans talebi modelidir.
type TalepAvans struct {
	ID               int       `json:"id"`
	ProjeID          int       `json:"proje_id"`
	UyeID            int       `json:"uye_id"`
	TalepNo          string    `json:"talep_no"`
	ButceKalemi      string    `json:"butce_kalemi"`
	TutarTL          float64   `json:"tutar_tl"`
	Gerekce          string    `json:"gerekce"`
	Durum            string    `json:"durum"`
	RedNotu          string    `json:"red_notu"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
	ProjeKodu        string    `json:"proje_kodu,omitempty"`
	ProjeBaslik      string    `json:"proje_baslik,omitempty"`
	TalepEdenAd      string    `json:"talep_eden_ad,omitempty"`
}
