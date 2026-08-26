package models

import "time"

// Satın alma talep durum sabitleri.
const (
	SatinalmaDurumBeklemede       = "Beklemede"
	SatinalmaDurumOnaylandi       = "Onaylandı"
	SatinalmaDurumReddedildi      = "Reddedildi"
	SatinalmaDurumKapatildi       = "Kapatildi"
	SatinalmaDurumIptalEdildi     = "IptalEdildi"
)

// Mutabakat / ödeme durum sabitleri.
const (
	OdemeDurumMutabakatBekliyor = "mutabakat_bekliyor"
	OdemeDurumOnaylandi         = "onaylandi"
	OdemeDurumReddedildi        = "reddedildi"
)

// Fark yönü sabitleri.
const (
	FarkYonuFazla = "fazla"
	FarkYonuEksik = "eksik"
	FarkYonuEsit  = "esit"
)

// Mutabakat karar sabitleri.
const (
	MutabakatKararOnayla = "onayla"
	MutabakatKararIptal  = "iptal"
)

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
	Durum            string    `json:"durum"` // Beklemede, Onaylandı, Reddedildi, Kapatildi, IptalEdildi
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
	RevizeEdenAdSoyad  string     `json:"revize_eden_ad_soyad,omitempty"`

	// Türkçe Yorum: Mutabakat sonrası fiili tutar özeti (listeleme için).
	FiiliTutar     *float64 `json:"fiili_tutar,omitempty"`
	MutabakatDurum *string  `json:"mutabakat_durum,omitempty"`
}

// EffectiveAmount onaylı/revize tutarı (taahhüt) döner.
// Türkçe Yorum: Revize tutar varsa onu, yoksa orijinal toplam fiyatı taahhüt olarak kullanır.
func (t *SatinalmaTalebi) EffectiveAmount() float64 {
	if t.RevizeToplamFiyat != nil {
		return *t.RevizeToplamFiyat
	}
	return t.ToplamFiyat
}

// SatinalmaOdeme onay sonrası fiili ödeme / mutabakat kaydını temsil eder.
// Türkçe Yorum: Taahhüt ile fatura bedeli arasındaki farkı TTO yönetir.
type SatinalmaOdeme struct {
	OdemeID          int        `json:"odeme_id"`
	TalepID          int        `json:"talep_id"`
	TalepNo          string     `json:"talep_no"`
	ProjeID          int        `json:"proje_id"`
	KalemID          int        `json:"kalem_id"`
	TaahhutTutari    float64    `json:"taahhut_tutari"`
	FiiliTutar       float64    `json:"fiili_tutar"`
	FarkTutari       float64    `json:"fark_tutari"`
	FarkYonu         string     `json:"fark_yonu"`
	FarkKalemID      *int       `json:"fark_kalem_id,omitempty"`
	FaturaNo         *string    `json:"fatura_no,omitempty"`
	FaturaTarihi     *time.Time `json:"fatura_tarihi,omitempty"`
	OdemeTarihi      *time.Time `json:"odeme_tarihi,omitempty"`
	ParaBirimi       string     `json:"para_birimi"`
	EvrakYolu        *string    `json:"evrak_yolu,omitempty"`
	Durum            string     `json:"durum"`
	TtoUyeID         *int       `json:"tto_uye_id,omitempty"`
	TtoGerekce       *string    `json:"tto_gerekce,omitempty"`
	KararTarihi      *time.Time `json:"karar_tarihi,omitempty"`
	OlusturmaTarihi  time.Time  `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time  `json:"guncelleme_tarihi"`

	// Join alanları
	MalzemeAdi       string `json:"malzeme_adi,omitempty"`
	ProjeKodu        string `json:"proje_kodu,omitempty"`
	ProjeBaslik      string `json:"proje_baslik,omitempty"`
	TtoAdSoyad       string `json:"tto_ad_soyad,omitempty"`
	ButceKategoriAdi string `json:"butce_kategori_adi,omitempty"`
	FarkKalemKategoriAdi string `json:"fark_kalem_kategori_adi,omitempty"`
}

// ButceHareket bütçe ledger satırını temsil eder.
// Türkçe Yorum: Rezervasyon, fiili harcama, iade vb. hareketlerin audit kaydıdır.
type ButceHareket struct {
	HareketID      int       `json:"hareket_id"`
	ProjeID        int       `json:"proje_id"`
	KalemID        int       `json:"kalem_id"`
	KaynakTip      string    `json:"kaynak_tip"`
	KaynakID       int       `json:"kaynak_id"`
	HareketTip     string    `json:"hareket_tip"`
	Tutar          float64   `json:"tutar"`
	Aciklama       *string   `json:"aciklama,omitempty"`
	IslemiYapanID  *int      `json:"islemi_yapan_id,omitempty"`
	OlusturmaTarihi time.Time `json:"olusturma_tarihi"`
}

// MutabakatIstek TTO'nun fiili tutar mutabakat isteğidir.
// Türkçe Yorum: karar=onayla → kapatır; karar=iptal → taahhüdü serbest bırakır.
type MutabakatIstek struct {
	TalepID      int     `json:"talep_id" binding:"required"`
	FiiliTutar   float64 `json:"fiili_tutar"`
	FaturaNo     string  `json:"fatura_no"`
	FaturaTarihi string  `json:"fatura_tarihi"` // YYYY-MM-DD
	OdemeTarihi  string  `json:"odeme_tarihi"`  // YYYY-MM-DD
	Gerekce      string  `json:"gerekce"`
	Karar        string  `json:"karar" binding:"required"` // onayla | iptal
	FarkKalemID  int     `json:"fark_kalem_id"`            // Fark varsa hangi bütçe kalemine yansıyacağı
}

// ButceHarcamaRaporKalemi bir proje bütçe kaleminin harcama özetini tutar.
// Türkçe Yorum: Planlanan, taahhüt, fiili, bekleyen ve kullanılabilir tutarları raporlar.
type ButceHarcamaRaporKalemi struct {
	KalemID       int     `json:"kalem_id"`
	KategoriAdi   string  `json:"kategori_adi"`
	Aciklama      string  `json:"aciklama"`
	Planlanan     float64 `json:"planlanan"`
	Taahhut       float64 `json:"taahhut"`
	Fiili         float64 `json:"fiili"`
	Bekleyen      float64 `json:"bekleyen"`
	Kullanilabilir float64 `json:"kullanilabilir"`
	// Geriye dönük uyumluluk alanları
	Harcanan float64 `json:"harcanan"` // taahhüt + bekleyen + fiili
	Odenen   float64 `json:"odenen"`   // fiili (gerçek ödenen)
	Kalan    float64 `json:"kalan"`    // kullanilabilir
}

// ProjeButceHarcamaRaporu projenin bütçe kalemi bazlı harcama raporunu temsil eder.
type ProjeButceHarcamaRaporu struct {
	ProjeID     int                       `json:"proje_id"`
	ProjeKodu   string                    `json:"proje_kodu"`
	ProjeBaslik string                    `json:"proje_baslik"`
	Kalemler    []ButceHarcamaRaporKalemi `json:"kalemler"`
	ToplamPlanlanan      float64          `json:"toplam_planlanan"`
	ToplamTaahhut        float64          `json:"toplam_taahhut"`
	ToplamFiili          float64          `json:"toplam_fiili"`
	ToplamBekleyen       float64          `json:"toplam_bekleyen"`
	ToplamKullanilabilir float64          `json:"toplam_kullanilabilir"`
	// Geriye dönük uyumluluk
	ToplamHarcanan float64 `json:"toplam_harcanan"`
	ToplamOdenen   float64 `json:"toplam_odenen"`
	ToplamKalan    float64 `json:"toplam_kalan"`
}
