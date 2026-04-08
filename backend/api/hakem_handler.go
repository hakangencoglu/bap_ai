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
	// Auth middleware içinden Uye set edilmiş olmalı
	uyeVal, exists := c.Get("Uye")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz erişim"})
		return
	}

	uye, ok := uyeVal.(models.Uye)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi geçersiz"})
		return
	}

	// Rol kontrolü: 3 numara veya db'deki 'hakem' role_id'sidir. Şimdilik isim kontrolü yapmıyoruz çünkü jwt claim'de rolleri tam almıyoruz
	// Varsayılan rol id'miz veya direkt fonksiyona yollayabiliriz, repo kendi db rolünü kontrol ediyorsa id ile getirebilir.
	projeler, err := h.HakemService.GetProjelerByHakem(uye.UyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Projeler alınırken hata: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, projeler)
}

// SubmiteDegerlendirme, hakemin formdan yolladığı değerlendirme sonucunu kaydeder
func (h *HakemHandler) SubmitDegerlendirme(c *gin.Context) {
	uyeVal, exists := c.Get("Uye")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz erişim"})
		return
	}

	uye, ok := uyeVal.(models.Uye)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi geçersiz"})
		return
	}

	var req models.DegerlendirmeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek formu"})
		return
	}

	err := h.HakemService.SubmitDegerlendirme(uye.UyeID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Değerlendirme gönderilemedi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Değerlendirme başarıyla kaydedildi"})
}
