package models

// SistemSayfa yapısı, veritabanındaki sistem_sayfa tablosunun Go karşılığıdır.
type SistemSayfa struct {
	SayfaID   int    `json:"sayfa_id"`
	SayfaAdi  string `json:"sayfa_adi"`
	SayfaKodu string `json:"sayfa_kodu"`
	UrlYolu   string `json:"url_yolu"`
}

// SayfaRolYetki yapısı, veritabanındaki sayfa_rol_yetki tablosunun Go karşılığıdır.
type SayfaRolYetki struct {
	YetkiID     int `json:"yetki_id"`
	SistemRolID int `json:"sistem_rol_id"`
	SayfaID     int `json:"sayfa_id"`
}

// UpdateSayfaYetkiItem yapısı, adminin yetki matrisinden gönderdiği tekil yetki güncellemesini temsil eder.
type UpdateSayfaYetkiItem struct {
	SistemRolID int  `json:"sistem_rol_id" binding:"required"`
	SayfaID     int  `json:"sayfa_id" binding:"required"`
	Allowed     bool `json:"allowed"`
}

// UpdateSayfaYetkiRequest yapısı, adminin toplu yetki güncelleme isteğidir.
type UpdateSayfaYetkiRequest struct {
	Permissions []UpdateSayfaYetkiItem `json:"permissions" binding:"required"`
}

// SayfaYetkiMatrix yapısı, admin yetki paneline gönderilecek olan tüm sayfa, rol ve mevcut yetkilerin matrisidir.
type SayfaYetkiMatrix struct {
	Roles       []SistemRolTanimlama `json:"roles"`
	Pages       []SistemSayfa        `json:"pages"`
	Permissions []SayfaRolYetki      `json:"permissions"`
}

// CreateRoleRequest yeni rol ekleme isteğini temsil eder.
// Türkçe Yorum: Admin'in yeni rol oluştururken gönderdiği JSON gövdesi
type CreateRoleRequest struct {
	RolAdi   string `json:"rol_adi" binding:"required"`
	Sayfalar []int  `json:"sayfalar"`
}

// UpdateRoleRequest rol güncelleme isteğini temsil eder.
// Türkçe Yorum: Admin'in mevcut bir rolü güncellerken gönderdiği JSON gövdesi
type UpdateRoleRequest struct {
	RolAdi   string `json:"rol_adi" binding:"required"`
	Sayfalar []int  `json:"sayfalar"`
}
