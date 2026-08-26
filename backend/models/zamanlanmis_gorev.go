package models

import "time"

// ZamanlanmisGorevKural admin tarafından tanımlanan otomatik bildirim kuralını temsil eder.
// Türkçe Yorum: BAP türüne ve süre aralığına göre tetiklenen e-posta ve SMS kurallarını saklar.
type ZamanlanmisGorevKural struct {
	KuralID          int        `json:"kural_id"`
	KuralAdi         string     `json:"kural_adi"`
	BapTuruID        *int       `json:"bap_turu_id"`       // NULL ise tüm BAP türleri için geçerlidir
	BapTuruAdi       string     `json:"bap_turu_adi,omitempty"`
	TetiklemeTipi    string     `json:"tetikleme_tipi"`   // baslangic_sonrasi_ay, bitim_oncesi_ay, bitim_oncesi_gun, periyodik_ay
	ZamanDegeri      int        `json:"zaman_degeri"`     // Örn: 3 (3. ay), 30 (bitime 30 gün kala)
	EpostaAktif      bool       `json:"eposta_aktif"`
	SmsAktif         bool       `json:"sms_aktif"`
	EpostaKonu       string     `json:"eposta_konu"`
	EpostaSablon     string     `json:"eposta_sablon"`    // HTML/metin şablonu ({yurutucu_ad}, {proje_kodu}...)
	SmsSablon        string     `json:"sms_sablon"`       // SMS metin şablonu
	AktifMi          bool       `json:"aktif_mi"`
	OlusturanID      *int       `json:"olusturan_id"`
	OlusturmaTarihi  time.Time  `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
	AliciHedefleri   []string   `json:"alici_hedefleri"` // yurutucu, tto
}

// Bildirim alıcı hedef sabitleri.
const (
	AliciHedefYurutucu = "yurutucu"
	AliciHedefTTO      = "tto"
)

// BildirimAliciKisi gönderim hedefindeki kişinin iletişim bilgisini taşır.
type BildirimAliciKisi struct {
	HedefKey string
	AdSoyad  string
	Eposta   string
	Telefon  string
}

// ZamanlanmisGorevLog gönderilen otomatik bildirimlerin geçmiş kaydını temsil eder.
// Türkçe Yorum: E-Posta ve SMS gönderim sonuçları, alıcılar ve tarihler bu struct ile modellenir.
type ZamanlanmisGorevLog struct {
	LogID          int       `json:"log_id"`
	KuralID        int       `json:"kural_id"`
	KuralAdi       string    `json:"kural_adi,omitempty"`
	ProjeID        int       `json:"proje_id"`
	ProjeKodu      string    `json:"proje_kodu,omitempty"`
	Kanal          string    `json:"kanal"`          // eposta, sms
	Alici          string    `json:"alici"`          // E-posta adresi veya Telefon numarası
	Icerik         string    `json:"icerik"`
	Durum          string    `json:"durum"`          // basarili, hata, simule_edildi
	HataMesaji     string    `json:"hata_mesaji,omitempty"`
	GonderimTarihi time.Time `json:"gonderim_tarihi"`
}

// KuralEslesenProje tetikleme zamanı gelmiş aktif projeyi temsil eder.
// Türkçe Yorum: Şablon parametrelerini doldurmak için gerekli proje ve yürütücü verilerini taşır.
type KuralEslesenProje struct {
	ProjeID         int
	ProjeKodu       string
	ProjeBaslik     string
	BapTuru         string
	SureAy          int
	BaslangicTarihi time.Time
	BitisTarihi     time.Time
	YurutucuID      int
	YurutucuAd      string
	YurutucuEposta  string
	YurutucuTelefon string
}

// ZamanlanmisGorevTetiklemeSonuc manuel çalıştırma sonucunu özetler.
type ZamanlanmisGorevTetiklemeSonuc struct {
	ToplamIslenenKural int `json:"toplam_islenen_kural"`
	ToplamIslenenProje int `json:"toplam_islenen_proje"`
	GonderilenEposta   int `json:"gonderilen_eposta"`
	GonderilenSms      int `json:"gonderilen_sms"`
	HataSayisi         int `json:"hata_sayisi"`
}
