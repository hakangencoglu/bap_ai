package models

import "time"

// Feedback veritabanındaki geri_bildirim tablosunun Go karşılığıdır.
// Türkçe Yorum: Geri bildirim kayıtlarının temel veri yapısı.
type Feedback struct {
	ID              int       `json:"id"`
	UyeID           int       `json:"uye_id"`
	Konu            string    `json:"konu"`
	Mesaj           string    `json:"mesaj"`
	SayfaURL        string    `json:"sayfa_url"`
	Durum           string    `json:"durum"`
	OlusturmaTarihi time.Time `json:"olusturma_tarihi"`
}

// CreateFeedbackRequest kullanıcının geri bildirim modülünden gönderdiği DTO verisidir.
// Türkçe Yorum: Frontend'den gelen geri bildirim isteğinin JSON gövdesi.
type CreateFeedbackRequest struct {
	Konu     string `json:"konu" binding:"required"`
	Mesaj    string `json:"mesaj" binding:"required"`
	SayfaURL string `json:"sayfa_url"`
}

// FeedbackSenderInfo e-posta imzasında gösterilecek tüm kullanıcı bilgilerini temsil eder.
// Türkçe Yorum: Gönderenin ad, unvan, iletişim ve sistemdeki rol/yetki detaylarını tutan yapı.
type FeedbackSenderInfo struct {
	UyeID      int      `json:"uye_id"`
	Ad         string   `json:"ad"`
	Soyad      string   `json:"soyad"`
	Unvan      string   `json:"unvan"`
	Bolum      string   `json:"bolum"`
	Eposta     string   `json:"eposta"`
	Telefon    string   `json:"telefon"`
	Roles      []string `json:"roles"`       // Örn: ["akademisyen", "hakem"]
	RoleLabels []string `json:"role_labels"` // Örn: ["Akademisyen", "Hakem"]
}
