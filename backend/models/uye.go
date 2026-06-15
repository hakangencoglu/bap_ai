package models

import "time"

// Uye yapısı, veritabanındaki uye tablosunun Go karşılığıdır.
type Uye struct {
	UyeID              int       `json:"uye_id"`
	Rol                string    `json:"rol"`
	Ad                 string    `json:"ad"`
	Soyad              string    `json:"soyad"`
	Unvan              string    `json:"unvan"`
	Bolum              string    `json:"bolum"`
	Eposta             string    `json:"eposta"`
	Telefon            string    `json:"telefon"`
	IzuUyesi           bool      `json:"izu_uyesi"`
	SifreHash          string    `json:"-"` // JSON'da şifre hash'i gizlenir
	AktifMi            bool      `json:"aktif_mi"`
	SifreDegistirZorla bool      `json:"sifre_degistir_zorla"` // Türkçe Yorum: Giriş yaptıktan sonra zorla şifre değiştirme bayrağı
	OlusturmaTarihi    time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi   time.Time `json:"guncelleme_tarihi"`
}

// UyeDetay yapısı, veritabanındaki uye_detay tablosunun Go karşılığıdır.
// Kullanıcının giriş sonrası tamamlayacağı profil bilgilerini tutar.
type UyeDetay struct {
	DetayID           int       `json:"detay_id"`
	UyeID             int       `json:"uye_id"`
	Rol               string    `json:"rol"`
	Unvan             string    `json:"unvan"`
	Bolum             string    `json:"bolum"`
	Telefon           string    `json:"telefon"`
	IzuUyesi          bool      `json:"izu_uyesi"`
	ProfilTamamlandi  bool      `json:"profil_tamamlandi"`
	OlusturmaTarihi   time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi  time.Time `json:"guncelleme_tarihi"`
}

// UyeWithDetay yapısı, uye ve uye_detay tablolarından birleştirilmiş veriyi tutar.
type UyeWithDetay struct {
	UyeID              int       `json:"uye_id"`
	Ad                 string    `json:"ad"`
	Soyad              string    `json:"soyad"`
	Eposta             string    `json:"eposta"`
	SifreHash          string    `json:"-"`
	AktifMi            bool      `json:"aktif_mi"`
	Rol                string    `json:"rol"`
	Unvan              string    `json:"unvan"`
	Bolum              string    `json:"bolum"`
	Telefon            string    `json:"telefon"`
	IzuUyesi           bool      `json:"izu_uyesi"`
	ProfilTamamlandi   bool      `json:"profil_tamamlandi"`
	SifreDegistirZorla bool      `json:"sifre_degistir_zorla"` // Türkçe Yorum: Detaylı üye yapısında şifre zorlama bayrağı
	OlusturmaTarihi    time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi   time.Time `json:"guncelleme_tarihi"`
}

// RegisterRequest yapısı, kayıt olma isteğinde gelen verileri tutar.
// Sadece temel bilgiler alınır, detay bilgiler giriş sonrası tamamlanır.
type RegisterRequest struct {
	Ad     string `json:"ad" binding:"required"`
	Soyad  string `json:"soyad" binding:"required"`
	Eposta string `json:"eposta" binding:"required,email"`
	Sifre  string `json:"sifre" binding:"required,min=6"`
}

// AdminCreateUserRequest yapısı, adminin yeni kullanıcı ekleme isteğinde gelen verileri tutar.
type AdminCreateUserRequest struct {
	Ad                 string `json:"ad" binding:"required"`
	Soyad              string `json:"soyad" binding:"required"`
	Eposta             string `json:"eposta" binding:"required,email"`
	Sifre              string `json:"sifre"`
	Rol                string `json:"rol" binding:"required"`
	Unvan              string `json:"unvan"`
	Bolum              string `json:"bolum"`
	Telefon            string `json:"telefon"`
	IzuUyesi           bool   `json:"izu_uyesi"`
	SifreDegistirZorla bool   `json:"sifre_degistir_zorla"` // Türkçe Yorum: Admin istek yapısında şifre zorlama bayrağı
}

// AdminUpdateUserRequest yapısı, adminin var olan kullanıcıyı güncelleme isteğinde gelen verileri tutar.
type AdminUpdateUserRequest struct {
	Ad       string `json:"ad" binding:"required"`
	Soyad    string `json:"soyad" binding:"required"`
	Eposta   string `json:"eposta" binding:"required,email"`
	Rol      string `json:"rol" binding:"required"`
	Unvan    string `json:"unvan"`
	Bolum    string `json:"bolum"`
	Telefon  string `json:"telefon"`
	IzuUyesi bool   `json:"izu_uyesi"`
}

// ProfilTamamlamaRequest yapısı, giriş sonrası profil tamamlama isteğinde gelen verileri tutar.
type ProfilTamamlamaRequest struct {
	Rol      string `json:"rol" binding:"required"`
	Unvan    string `json:"unvan"`
	Bolum    string `json:"bolum"`
	Telefon  string `json:"telefon"`
	IzuUyesi bool   `json:"izu_uyesi"`
}

// LoginRequest yapısı, giriş yapma isteğinde gelen verileri tutar.
type LoginRequest struct {
	Eposta string `json:"eposta" binding:"required,email"`
	Sifre  string `json:"sifre" binding:"required"`
}

// LoginResponse yapısı, başarılı giriş sonrasında dönen verileri tutar.
type LoginResponse struct {
	Token string       `json:"token"`
	Uye   UyeWithDetay `json:"uye"`
}
