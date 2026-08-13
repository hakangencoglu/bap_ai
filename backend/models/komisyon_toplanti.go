package models

import "time"

// KomisyonToplantisi komisyon toplantı verilerini temsil eder.
// Türkçe Yorum: Toplantı genel bilgileri, tarihi, gündemi ve kararları burada saklanır.
type KomisyonToplantisi struct {
	ToplantiID      int                        `json:"toplanti_id"`
	ToplantiNo      string                     `json:"toplanti_no"`
	Tarih           time.Time                  `json:"tarih"`
	Gundem          string                     `json:"gundem"`
	Karar           string                     `json:"karar"`
	Durum           string                     `json:"durum"`
	OlusturanID     int                        `json:"olusturan_id"`
	OlusturmaTarihi time.Time                  `json:"olusturma_tarihi"`
	Katilimcilar    []*KomisyonToplantiKatilim `json:"katilimcilar"`
	Projeler        []*KomisyonToplantisiProje `json:"projeler,omitempty"`
	// Katılım oranı sayaçları: liste ekranlarında katılımcı detayı çekilmeden oran gösterilebilsin.
	KatilimciSayisi int `json:"katilimci_sayisi"` // Yoklamaya alınan (davetli) komisyon üyesi sayısı
	KatilanSayisi   int `json:"katilan_sayisi"`   // Yoklamada "katıldı" işaretlenen üye sayısı
}

// KomisyonToplantiKatilim komisyon toplantısına katılan üyelerin katılım durumlarını temsil eder.
type KomisyonToplantiKatilim struct {
	UyeID   int    `json:"uye_id"`
	Ad      string `json:"ad"`
	Soyad   string `json:"soyad"`
	Unvan   string `json:"unvan"`
	Bolum   string `json:"bolum"`
	Katildi bool   `json:"katildi"`
}

// KomisyonToplantisiProje bir toplantıda görüşülen proje ve bu toplantıya özel kararı temsil eder.
// Türkçe Yorum: Köprü tablo (komisyon_toplanti_proje) satırlarını karşılar.
type KomisyonToplantisiProje struct {
	ID              int        `json:"id"`
	ToplantiID      int        `json:"toplanti_id"`
	ProjeID         int        `json:"proje_id"`
	GundemSirasi    *int       `json:"gundem_sirasi"`
	Karar           string     `json:"karar"`
	KararAciklamasi string     `json:"karar_aciklamasi"`
	KararTarihi     *time.Time `json:"karar_tarihi"`
	EkleyenID       *int       `json:"ekleyen_id"`
	OlusturmaTarihi time.Time  `json:"olusturma_tarihi"`
	ProjeKodu       string     `json:"proje_kodu"`
	ProjeBaslik     string     `json:"proje_baslik"`
	YurutucuAd      string     `json:"yurutucu_ad"`
	YurutucuUnvan   string     `json:"yurutucu_unvan"`
	MevcutDurum     string     `json:"mevcut_durum"`
}

// KomisyonToplantiBelge PDF tutanak üretimi için gerekli tüm veriyi bir arada tutar.
type KomisyonToplantiBelge struct {
	Toplanti     *KomisyonToplantisi        `json:"toplanti"`
	Katilimcilar []*KomisyonToplantiKatilim `json:"katilimcilar"`
	Projeler     []*KomisyonToplantisiProje `json:"projeler"`
}

// KomisyonBekleyenProje toplantı gündemine eklenebilecek komisyon_bekliyor projeyi temsil eder.
type KomisyonBekleyenProje struct {
	ProjeID       int     `json:"proje_id"`
	ProjeKodu     string  `json:"proje_kodu"`
	ProjeBaslik   string  `json:"proje_baslik"`
	YurutucuAd    string  `json:"yurutucu_ad"`
	YurutucuUnvan string  `json:"yurutucu_unvan"`
	BapTuru       string  `json:"bap_turu"`
	ToplamButce   float64 `json:"toplam_butce"`
}
