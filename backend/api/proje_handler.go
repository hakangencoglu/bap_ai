package api

import (
	"net/http"

	"bap_ai/backend/models"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// ProjeHandler yapısı proje HTTP isteklerini karşılar.
type ProjeHandler struct {
	ProjeService *service.ProjeService
}

// NewProjeHandler yeni bir ProjeHandler oluşturur.
func NewProjeHandler(projeService *service.ProjeService) *ProjeHandler {
	return &ProjeHandler{ProjeService: projeService}
}

// CreateProje yeni bir proje başvurusu kabul eder.
// POST /api/proje
func (h *ProjeHandler) CreateProje(c *gin.Context) {
	// Middleware'den giriş yapan üye ID'sini alıyoruz
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	// Sadece modeldeki gerekli olan verileri bağlayacağız (c.BindJSON veya bind için bir DTO)
	var req models.Proje
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek formu"})
		return
	}

	roleIDFloat, _ := c.Get("role_id")
	roleID := int(roleIDFloat.(float64))

	// Rol: 3 (Öğrenci) kontrolü - Sadece bap-100 ve bap-200'e başvurabilir
	if roleID == 3 {
		if req.Tur != "bap-100" && req.Tur != "bap-200" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Öğrenci hesabıyla sadece BAP-100 ve BAP-200 türünde başvuru yapabilirsiniz."})
			return
		}
	}

	// Service katmanına iletiyoruz.
	if err := h.ProjeService.CreateProje(uyeID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proje kaydedilemedi"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Proje başvurusu başarıyla kaydedildi"})
}
