package models

import "time"

// Proje yapısı, veritabanındaki proje tablosunun Go karşılığıdır.
type Proje struct {
	ProjeID              int       `json:"proje_id"`
	BaslikTr             string    `json:"baslik_tr"`
	BaslikEn             string    `json:"baslik_en"`
	BaslangicTarihi      string    `json:"baslangic_tarihi"`
	ProjeSuresi          string    `json:"proje_suresi"`
	ToplamTutar          int       `json:"toplam_tutar"`
	EtikKurul            bool      `json:"etik_kurul"`
	OzetTr               string    `json:"ozet_tr"`
	OzetEn               string    `json:"ozet_en"`
	AmacVeHedef          string    `json:"amac_ve_hedef"`
	Ozgunluk             string    `json:"ozgunluk"`
	Metodoloji           string    `json:"metodoloji"`
	RiskYonetimi         string    `json:"risk_yonetimi"`
	ProjeCiktilari       string    `json:"proje_ciktilari"`
	AktiviteBilgisi      string    `json:"aktivite_bilgisi"`
	AktiviteFizibilitesi string    `json:"aktivite_fizibilitesi"`
	Durum                string    `json:"durum"`
	Tur                  string    `json:"tur"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// DashboardStats yapısı, dashboard sayfasında gösterilecek istatistikleri tutar.
type DashboardStats struct {
	AktifProje   int `json:"aktif_proje"`
	OnayBekleyen int `json:"onay_bekleyen"`
	Tamamlanan   int `json:"tamamlanan"`
	ToplamButce  int `json:"toplam_butce"`
}

// ProjeOzet yapısı, dashboard'daki son başvurular tablosu için özet proje bilgisi tutar.
type ProjeOzet struct {
	ProjeID   int    `json:"proje_id"`
	BaslikTr  string `json:"baslik_tr"`
	Tur       string `json:"tur"`
	Tarih     string `json:"tarih"`
	Durum     string `json:"durum"`
}

// ProfilProjeBilgisi yapısı, profil sayfasındaki proje kartları için bilgi tutar.
// Proje adı, tür (alan anahtar kelimeleri), durum ve kullanıcının projedeki rolünü içerir.
type ProfilProjeBilgisi struct {
	ProjeID  int    `json:"proje_id"`
	BaslikTr string `json:"baslik_tr"`
	Tur      string `json:"tur"`
	Durum    string `json:"durum"`
	UyeRol   string `json:"uye_rol"`
}

// ProjeUye yapısı, projeye kayıtlı üyelerin modal vs işlemlerde listelenmesi için oluşturuldu.
type ProjeUye struct {
	UyeID  int    `json:"uye_id"`
	AdTumu string `json:"ad_tumu"` // "Ad Soyad"
	RoleID int    `json:"role_id"` // 3 (Öğrenci) veya 2 (Akademisyen) filtresi için
	Rol    string `json:"rol"`     // Projedeki rolü ("Yürütücü", "Araştırmacı" vb.)
}

