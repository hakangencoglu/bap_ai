package api

import (
	"net/http"
	"strconv"

	"bap_ai/backend/models"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// SozlesmeHandler, proje sözleşmesi HTTP isteklerini yönetir.
// Türkçe Yorum: Akademisyenin sözleşme doldurma ve görüntüleme isteklerini karşılar.
type SozlesmeHandler struct {
	Service *service.SozlesmeService
}

// NewSozlesmeHandler yeni bir SozlesmeHandler örneği döner.
func NewSozlesmeHandler(srv *service.SozlesmeService) *SozlesmeHandler {
	return &SozlesmeHandler{Service: srv}
}

// SaveSozlesme, projeye ait sözleşme verilerini kaydeder.
// POST /api/proje/:id/sozlesme
func (h *SozlesmeHandler) SaveSozlesme(c *gin.Context) {
	uyeID, ok := uyeIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı doğrulaması başarısız"})
		return
	}
	projeID, err := strconv.Atoi(c.Param("id"))
	if err != nil || projeID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	var sz models.ProjeSozlesme
	if err := c.ShouldBindJSON(&sz); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek parametreleri"})
		return
	}
	sz.ProjeID = projeID
	sz.UyeID = uyeID

	if err := h.Service.SaveSozlesme(&sz); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sözleşme başarıyla kaydedildi",
		"sozlesme": sz,
	})
}

// GetSozlesme, projeye ait sözleşme verisini getirir.
// GET /api/proje/:id/sozlesme
func (h *SozlesmeHandler) GetSozlesme(c *gin.Context) {
	projeID, err := strconv.Atoi(c.Param("id"))
	if err != nil || projeID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	sz, err := h.Service.GetSozlesme(projeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Sözleşme verisi alınamadı"})
		return
	}
	if sz == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bu proje için henüz sözleşme doldurulmamış"})
		return
	}

	c.JSON(http.StatusOK, sz)
}
