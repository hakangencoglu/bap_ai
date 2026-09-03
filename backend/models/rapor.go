package models

import "time"

// ProjeAraRapor yapısı, teslim edilen gelişme ve kesin sonuç raporlarını temsil eder.
type ProjeAraRapor struct {
	RaporID          int        `json:"rapor_id"`
	ProjeID          int        `json:"proje_id"`
	YukleyenUyeID    int        `json:"yukleyen_uye_id"`
	RaporTuru        string     `json:"rapor_turu"` // ara_rapor | sonuc_raporu
	RaporDonemi      int        `json:"rapor_donemi"`
	Baslik           string     `json:"baslik"`
	Aciklama         string     `json:"aciklama"`
	DosyaURL         string     `json:"dosya_url"`
	OrijinalDosyaAdi string     `json:"orijinal_dosya_adi,omitempty"`
	Durum            string     `json:"durum"` // bekliyor | onaylandi | revizyon | reddedildi
	OnaylayanUyeID   *int       `json:"onaylayan_uye_id,omitempty"`
	OnayNotu         string     `json:"onay_notu,omitempty"`
	OnayTarihi       *time.Time `json:"onay_tarihi,omitempty"`
	OlusturmaTarihi  time.Time  `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time  `json:"guncelleme_tarihi"`

	// JOIN ile doldurulacak alanlar
	YukleyenAdSoyad  string `json:"yukleyen_ad_soyad,omitempty"`
	YukleyenUnvan    string `json:"yukleyen_unvan,omitempty"`
	OnaylayanAdSoyad string `json:"onaylayan_ad_soyad,omitempty"`
	ProjeKodu        string `json:"proje_kodu,omitempty"`
	ProjeBaslik      string `json:"proje_baslik,omitempty"`
	BapTuru          string `json:"bap_turu,omitempty"`
}

// RaporDegerlendirmeRequest yapısı, TTO/Komisyon tarafından yapılan onay/revizyon/red kararlarını tutar.
type RaporDegerlendirmeRequest struct {
	RaporID  int    `json:"rapor_id"`
	Durum    string `json:"durum"` // onaylandi | revizyon | reddedildi
	OnayNotu string `json:"onay_notu"`
}

// TTORaporTakipItem yapısı, TTO Sorumlusunun Ara Rapor Takip Ekranında listelediği tüm projelerin ve rapor teslim durumlarının özetini temsil eder.
// Türkçe Yorum: Zamanı geçmiş (gecikmiş), bekleyen, onaylanan ve yaklaşan ara raporların takibini sağlar.
type TTORaporTakipItem struct {
	ProjeID          int        `json:"proje_id"`
	ProjeKodu        string     `json:"proje_kodu"`
	ProjeBaslik      string     `json:"proje_baslik"`
	YurutucuAdSoyad  string     `json:"yurutucu_ad_soyad"`
	YurutucuEposta   string     `json:"yurutucu_eposta"`
	BapTuru          string     `json:"bap_turu"`
	BaslangicTarihi  *time.Time `json:"baslangic_tarihi,omitempty"`
	BitisTarihi      *time.Time `json:"bitis_tarihi,omitempty"`
	HesaplananDonem  int        `json:"hesaplanan_donem"`
	SonTeslimTarihi  *time.Time `json:"son_teslim_tarihi,omitempty"`
	RaporDurumu      string     `json:"rapor_durumu"` // gecikmis | bekliyor | onaylandi | revizyon | yaklasiyor | beklenmiyor
	YuklenenRaporID  *int       `json:"yuklenen_rapor_id,omitempty"`
	YuklenenDosyaURL string     `json:"yuklenen_dosya_url,omitempty"`
	YuklenmeTarihi   *time.Time `json:"yuklenme_tarihi,omitempty"`
}
