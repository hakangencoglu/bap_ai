package api

import (
	"bap_ai/backend/models"
	"bap_ai/backend/service"
	"net/http"
	"strconv"

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

// GetProjeDetay, hakemin yetkili olduğu projenin tüm detaylarını döndürür.
// Türkçe Yorum: Hakeme atanan projenin tüm detaylarını yetki kontrolü yaptıktan sonra JSON olarak döner.
func (h *HakemHandler) GetProjeDetay(c *gin.Context) {
	projeIDStr := c.Param("id")
	projeID, err := strconv.Atoi(projeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	hakemID := int(uyeIDFloat.(float64))

	// Yetki kontrolü: hakem bu projeye atanmış mı?
	assigned, err := h.HakemService.IsHakemAssigned(hakemID, projeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Yetki kontrolü yapılamadı"})
		return
	}
	if !assigned {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bu projenin detaylarını görüntüleme yetkiniz yoktur."})
		return
	}

	details, err := h.HakemService.GetProjectDetailsForHakem(projeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proje detayları alınamadı"})
		return
	}

	c.JSON(http.StatusOK, details)
}

// GetDegerlendirmeQuestions, dinamik değerlendirme sorularını döner.
// Türkçe Yorum: Hakem değerlendirme sayfasının soruları dinamik çekmesi için JSON döner.
func (h *HakemHandler) GetDegerlendirmeQuestions(c *gin.Context) {
	questions, err := h.HakemService.GetDegerlendirmeQuestions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Değerlendirme soruları alınamadı: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, questions)
}
