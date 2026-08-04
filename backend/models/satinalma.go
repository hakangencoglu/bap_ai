package models

import "time"

// SatinalmaTalebi yapısı, satın alma talebi tablosunun veritabanı modelidir.
// Türkçe Yorum: Bu struct, veritabanındaki satinalma_talebi tablosunu temsil eder ve API dönüşlerinde kullanılır.
type SatinalmaTalebi struct {
	TalepID          int       `json:"talep_id"`
	TalepNo          string    `json:"talep_no"`
	ProjeID          int       `json:"proje_id"`
	UyeID            int       `json:"uye_id"`
	KalemID          int       `json:"kalem_id"`
	MalzemeAdi       string    `json:"malzeme_adi"`
	Miktar           int       `json:"miktar"`
	BirimFiyat       float64   `json:"birim_fiyat"`
	ToplamFiyat      float64   `json:"toplam_fiyat"`
	Durum            string    `json:"durum"` // 'Beklemede', 'Onaylandı', 'Reddedildi'
	Gerekce          string    `json:"gerekce"`
	RedNedeni        *string   `json:"red_nedeni,omitempty"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`

	// Türkçe Yorum: Listeleme ekranlarında gösterilecek olan, ilişkili tablolardan join ile çekilecek ek alanlar.
	UyeAdSoyad       string  `json:"uye_ad_soyad,omitempty"`
	ProjeBaslik      string  `json:"proje_baslik,omitempty"`
	ProjeKodu        string  `json:"proje_kodu,omitempty"`
	KalemAciklama    string  `json:"kalem_aciklama,omitempty"`
	ButceKategoriAdi string  `json:"butce_kategori_adi,omitempty"`
	MevcutButce      float64 `json:"mevcut_butce,omitempty"`
	KalanButce       float64 `json:"kalan_butce,omitempty"`

	// Türkçe Yorum: Revizyon (Bütçe Düzenleme) ile ilgili eklenen yeni alanlar
	RevizeBirimFiyat  *float64   `json:"revize_birim_fiyat,omitempty"`
	RevizeToplamFiyat *float64   `json:"revize_toplam_fiyat,omitempty"`
	RevizyonGerekcesi *string    `json:"revizyon_gerekcesi,omitempty"`
	RevizeEdenID      *int       `json:"revize_eden_id,omitempty"`
	RevizyonTarihi    *time.Time `json:"revizyon_tarihi,omitempty"`
	RevizeEdenAdSoyad string     `json:"revize_eden_ad_soyad,omitempty"`
}
