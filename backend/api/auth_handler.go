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
