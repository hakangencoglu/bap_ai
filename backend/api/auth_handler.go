package api

import (
	"net/http"

	"bap_ai/backend/models"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler yapısı, kimlik doğrulama HTTP handler'larını barındırır.
type AuthHandler struct {
	AuthService *service.AuthService
}

// NewAuthHandler fonksiyonu, yeni bir AuthHandler nesnesi döner.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: authService}
}

// Register fonksiyonu, yeni kullanıcı kayıt isteğini işler.
// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	// Gelen JSON verisini RegisterRequest yapısına bağlar
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Geçersiz istek verisi",
			"detaylar": err.Error(),
		})
		return
	}

	// Servis katmanına kayıt isteği gönderilir
	uye, err := h.AuthService.Register(&req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Başarılı kayıt yanıtı döner
	c.JSON(http.StatusCreated, gin.H{
		"mesaj": "Kullanıcı başarıyla kaydedildi",
		"uye":   uye,
	})
}

// Login fonksiyonu, kullanıcı giriş isteğini işler.
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	// Gelen JSON verisini LoginRequest yapısına bağlar
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Geçersiz istek verisi",
			"detaylar": err.Error(),
		})
		return
	}

	// Servis katmanına giriş isteği gönderilir
	response, err := h.AuthService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Başarılı giriş yanıtı döner
	c.JSON(http.StatusOK, gin.H{
		"mesaj": "Giriş başarılı",
		"token": response.Token,
		"uye":   response.Uye,
	})
}

// SetPassword fonksiyonu, kullanıcının ilk şifresini belirleme isteğini işler.
// POST /api/auth/set-password
func (h *AuthHandler) SetPassword(c *gin.Context) {
	var req struct {
		Eposta string `json:"eposta" binding:"required,email"`
		Sifre  string `json:"sifre" binding:"required,min=6"`
	}

	// Gelen JSON verisi kontrol edilir
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    "Geçersiz istek verisi",
			"detaylar": err.Error(),
		})
		return
	}

	// Servis katmanında şifre güncelleme ve giriş işlemi tetiklenir
	response, err := h.AuthService.SetPassword(req.Eposta, req.Sifre)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Başarılı işlem sonucunda token ve üye bilgileri dönülür
	c.JSON(http.StatusOK, gin.H{
		"mesaj": "Şifre başarıyla tanımlandı",
		"token": response.Token,
		"uye":   response.Uye,
	})
}

