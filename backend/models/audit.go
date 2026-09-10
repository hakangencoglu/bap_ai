package models

import "time"

// AuditLog, sistemde gerçekleşen kullanıcı işlemlerini ve erişim kayıtlarını temsil eder.
// Türkçe Yorum: Güvenlik, denetim ve KVKK uyumluluğu için tüm aktif kullanıcı hareketlerini tutar.
type AuditLog struct {
	LogID     int       `json:"log_id"`
	UyeID     *int      `json:"uye_id,omitempty"`
	Email     string    `json:"email,omitempty"`
	Rol       string    `json:"rol,omitempty"`
	IslemTuru string    `json:"islem_turu"`
	Endpoint  string    `json:"endpoint"`
	IPAdresi  string    `json:"ip_adresi,omitempty"`
	DurumKodu int       `json:"durum_kodu"`
	Aciklama  string    `json:"aciklama,omitempty"`
	Tarih     time.Time `json:"tarih"`
}
