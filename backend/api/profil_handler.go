package api

import (
	"net/http"

	"bap_ai/backend/models"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// ProfilHandler yapısı, profil sayfası HTTP handler'larını barındırır.
type ProfilHandler struct {
	ProfilService *service.ProfilService
	AuthService   *service.AuthService
}

// NewProfilHandler fonksiyonu, yeni bir ProfilHandler nesnesi döner.
func NewProfilHandler(profilService *service.ProfilService, authService *service.AuthService) *ProfilHandler {
	return &ProfilHandler{
		ProfilService: profilService,
		AuthService:   authService,
	}
}

// GetProfilBilgileri fonksiyonu, giriş yapan kullanıcının profil bilgilerini döner.
// GET /api/profil/bilgiler
func (h *ProfilHandler) GetProfilBilgileri(c *gin.Context) {
	// Middleware'den gelen uye_id alınır (JWT token'dan parse edilmiş)
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Kullanıcı bilgisi bulunamadı",
		})
		return
	}

	// JWT MapClaims sayıları float64 olarak tutar, int'e çevrilir
	uyeID := int(uyeIDFloat.(float64))

	// Servis katmanından profil bilgileri getirilir (üye + detay birleşik)
	uye, err := h.ProfilService.GetProfilBilgileri(uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Profil bilgileri alınamadı",
		})
		return
	}

	// Birleşik profil bilgileri döner
	c.JSON(http.StatusOK, gin.H{
		"uye_id":             uye.UyeID,
		"ad":                 uye.Ad,
		"soyad":              uye.Soyad,
		"eposta":             uye.Eposta,
		"rol":                uye.Rol,
		"unvan":              uye.Unvan,
		"bolum":              uye.Bolum,
		"fakulte_id":         uye.FakulteID,
		"bolum_id":           uye.BolumID,
		"fakulte_adi":        uye.FakulteAdi,
		"telefon":            uye.Telefon,
		"izu_uyesi":          uye.IzuUyesi,
		"aktif_mi":           uye.AktifMi,
		"profil_tamamlandi":  uye.ProfilTamamlandi,
		"olusturma_tarihi":   uye.OlusturmaTarihi,
	})
}

// TamamlaProfil fonksiyonu, giriş yapan kullanıcının profil detay bilgilerini kaydeder.
// POST /api/profil/tamamla
func (h *ProfilHandler) TamamlaProfil(c *gin.Context) {
	// Middleware'den gelen uye_id alınır
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Kullanıcı bilgisi bulunamadı",
		})
		return
	}

	// JWT MapClaims sayıları float64 olarak tutar, int'e çevrilir
	uyeID := int(uyeIDFloat.(float64))

	// Gelen JSON verisini ProfilTamamlamaRequest yapısına bağlar
	var req models.ProfilTamamlamaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    "Geçersiz istek verisi",
			"detaylar": err.Error(),
		})
		return
	}

	// Servis katmanına profil tamamlama isteği gönderilir
	if err := h.ProfilService.TamamlaProfil(uyeID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Profil bilgileri kaydedilemedi: " + err.Error(),
		})
		return
	}

	// Profil tamamlandıktan sonra güncel token üretilir
	newToken, err := h.AuthService.GenerateTokenForUye(uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Token güncellenemedi",
		})
		return
	}

	// Başarılı yanıt döner (yeni token ile)
	c.JSON(http.StatusOK, gin.H{
		"mesaj": "Profil bilgileri başarıyla kaydedildi",
		"token": newToken,
	})
}

// GetProfilProjeleri fonksiyonu, giriş yapan kullanıcının projelerini profil formatında döner.
// GET /api/profil/projeler
func (h *ProfilHandler) GetProfilProjeleri(c *gin.Context) {
	// Middleware'den gelen uye_id alınır
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Kullanıcı bilgisi bulunamadı",
		})
		return
	}

	// JWT MapClaims sayıları float64 olarak tutar, int'e çevrilir
	uyeID := int(uyeIDFloat.(float64))

	// Servis katmanından profil projeleri getirilir
	projeler, err := h.ProfilService.GetProfilProjeleri(uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Profil projeleri alınamadı",
		})
		return
	}

	// Nil slice yerine boş array döndür
	if projeler == nil {
		projeler = []models.ProfilProjeBilgisi{}
	}

	// Başarılı yanıt döner
	c.JSON(http.StatusOK, gin.H{
		"projeler": projeler,
	})
}
