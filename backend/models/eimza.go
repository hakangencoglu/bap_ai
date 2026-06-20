package models

import "time"

// ProjeImza yapısı, veritabanındaki proje_imza tablosunun Go karşılığıdır.
// Türkçe Yorum: Elektronik imza kayıtlarını modelleyen yapı.
type ProjeImza struct {
	ImzaID          int       `json:"imza_id"`
	ProjeID         int       `json:"proje_id"`
	UyeID           int       `json:"uye_id"`
	ImzaciAdSoyad   string    `json:"imzaci_ad_soyad"`
	Rol             string    `json:"rol"`
	ImzaTarihi      time.Time `json:"imza_tarihi"`
	ImzaDurumu      string    `json:"imza_durumu"`
	ImzaToken       string    `json:"imza_token"`
	IpAdresi        string    `json:"ip_adresi"`
	TarayiciBilgisi string    `json:"tarayici_bilgisi"`
}

// EimzaSignRequest yapısı, e-imza gönderme isteği için kullanılır.
// Türkçe Yorum: E-İmza atma talebini alan istek yapısı.
type EimzaSignRequest struct {
	ProjeID  int    `json:"proje_id" binding:"required"`
	Yontem   string `json:"yontem" binding:"required"`   // "mobil" veya "token"
	PinKodu  string `json:"pin_kodu" binding:"required"` // E-imza PIN kodu
	Telefon  string `json:"telefon,omitempty"`           // Mobil imza için isteğe bağlı telefon
	Operator string `json:"operator,omitempty"`          // Mobil imza için isteğe bağlı operatör
}
