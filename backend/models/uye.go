package models

import "time"

// Uye yapısı, veritabanındaki uye tablosunun Go karşılığıdır.
type Uye struct {
	UyeID          int       `json:"uye_id"`
	RoleID         int       `json:"role_id"`
	Unvan          string    `json:"unvan"`
	Ad             string    `json:"ad"`
	Soyad          string    `json:"soyad"`
	Bolum          string    `json:"bolum"`
	IletisimTel    string    `json:"iletisim_tel"`
	IletisimMail   string    `json:"iletisim_mail"`
	PasswordHash   string    `json:"-"` // JSON'da şifre hash'i gizlenir
	IzuAkademisyen bool      `json:"izu_akademisyen"`
	IzuOgrenci     bool      `json:"izu_ogrenci"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// RegisterRequest yapısı, kayıt olma isteğinde gelen verileri tutar.
type RegisterRequest struct {
	Ad           string `json:"ad" binding:"required"`
	Soyad        string `json:"soyad" binding:"required"`
	IletisimMail string `json:"iletisim_mail" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"`
	Unvan        string `json:"unvan"`
	Bolum        string `json:"bolum"`
	IletisimTel  string `json:"iletisim_tel"`
}

// LoginRequest yapısı, giriş yapma isteğinde gelen verileri tutar.
type LoginRequest struct {
	IletisimMail string `json:"iletisim_mail" binding:"required,email"`
	Password     string `json:"password" binding:"required"`
}

// LoginResponse yapısı, başarılı giriş sonrasında dönen verileri tutar.
type LoginResponse struct {
	Token string `json:"token"`
	Uye   Uye    `json:"uye"`
}
