package api

import (
	"bap_ai/backend/models"
	"bap_ai/backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HakemHandler struct {
	HakemService *service.HakemService
}

func NewHakemHandler(hs *service.HakemService) *HakemHandler {
	return &HakemHandler{HakemService: hs}
}

// GetAtananProjeler, giriş yapmış hakemin, kendine atanmış projelerini listeler
func (h *HakemHandler) GetAtananProjeler(c *gin.Context) {
	// Middleware'den gelen uye_id alınır (JWT token'dan parse edilmiş)
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}

	// JWT MapClaims sayıları float64 olarak tutar, int'e çevrilir
	uyeID := int(uyeIDFloat.(float64))

	// Hakeme atanan projeler servis katmanından getirilir
	projeler, err := h.HakemService.GetProjelerByHakem(uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Projeler alınırken hata: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, projeler)
}

// SubmitDegerlendirme, hakemin formdan yolladığı değerlendirme sonucunu kaydeder
func (h *HakemHandler) SubmitDegerlendirme(c *gin.Context) {
	// Middleware'den gelen uye_id alınır
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}

	// JWT MapClaims sayıları float64 olarak tutar, int'e çevrilir
	uyeID := int(uyeIDFloat.(float64))

	var req models.DegerlendirmeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek formu"})
		return
	}

	err := h.HakemService.SubmitDegerlendirme(uyeID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Değerlendirme gönderilemedi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Değerlendirme başarıyla kaydedildi"})
}

// KabulRedKarar, hakemin kendine atanan projeyi kabul veya reddetme kararını işler
func (h *HakemHandler) KabulRedKarar(c *gin.Context) {
	// Middleware'den gelen uye_id alınır
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}

	// JWT MapClaims sayıları float64 olarak tutar, int'e çevrilir
	uyeID := int(uyeIDFloat.(float64))

	var req models.HakemKararRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek formu"})
		return
	}

	// Servis katmanı üzerinden karar işlenir
	err := h.HakemService.KabulRedKarar(uyeID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Karar işlenemedi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Karar başarıyla kaydedildi"})
}
