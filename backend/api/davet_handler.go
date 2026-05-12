package api

import (
	"net/http"

	"bap_ai/backend/models"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// DavetHandler, proje davet HTTP isteklerini yönetir
type DavetHandler struct {
	DavetService *service.DavetService
}

// NewDavetHandler, yeni bir DavetHandler örneği oluşturur
func NewDavetHandler(davetService *service.DavetService) *DavetHandler {
	return &DavetHandler{DavetService: davetService}
}

// GetBekleyenDavetler, giriş yapan kullanıcının bekleyen proje davetlerini döner
// GET /api/davetler
func (h *DavetHandler) GetBekleyenDavetler(c *gin.Context) {
	// Middleware'den kullanıcı ID'sini al
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	// Bekleyen davetleri getir
	davetler, err := h.DavetService.GetBekleyenDavetler(uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Davetler alınamadı"})
		return
	}

	// Null kontrolü — boş dizi dön
	if davetler == nil {
		davetler = []models.ProjeDavet{}
	}

	c.JSON(http.StatusOK, gin.H{"davetler": davetler})
}

// RespondDavet, kullanıcının bir proje davetine yanıt vermesini işler
// POST /api/davet/yanit
func (h *DavetHandler) RespondDavet(c *gin.Context) {
	// Middleware'den kullanıcı ID'sini al
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	// İstek gövdesini parse et
	var req models.DavetYanit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek formatı"})
		return
	}

	// Davet yanıtını işle
	err := h.DavetService.RespondDavet(req.ProjeID, uyeID, req.Kabul)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Davet yanıtı işlenemedi"})
		return
	}

	mesaj := "Davet reddedildi"
	if req.Kabul {
		mesaj = "Davet kabul edildi"
	}
	c.JSON(http.StatusOK, gin.H{"message": mesaj})
}
