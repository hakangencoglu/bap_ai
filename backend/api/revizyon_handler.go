package api

import (
	"net/http"
	"strconv"

	"bap_ai/backend/models"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

type RevizyonHandler struct {
	RevizyonService *service.RevizyonService
}

func NewRevizyonHandler(revService *service.RevizyonService) *RevizyonHandler {
	return &RevizyonHandler{RevizyonService: revService}
}

// CreateRevizyon POST /api/revizyon
func (h *RevizyonHandler) CreateRevizyon(c *gin.Context) {
	// Sadece Role 2 (Akademisyen) veya Role 4 (Hakem) veya Role 1 (Admin) vb.
	// requireRoles ile middleware seviyesinde de koruyabiliriz ama biz controller'da oluşturan kişiyi de çekiyoruz.
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	var req models.Revizyon
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Hatalı girdi formu"})
		return
	}
	
	req.OlusturanKisiID = uyeID // Otomatik olarak atıyoruz

	if err := h.RevizyonService.CreateRevizyon(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Revizyon ataması oluşturulamadı"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Revizyon başarıyla eklendi"})
}

// GetAktifRevizyon GET /api/proje/:id/revizyon
func (h *RevizyonHandler) GetAktifRevizyon(c *gin.Context) {
	projeID := c.Param("id")
	if projeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}
	id, _ := strconv.Atoi(projeID)

	rev, err := h.RevizyonService.GetAktifRevizyon(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Revizyon bilgisi alınamadı"})
		return
	}

	if rev == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Aktif revizyon bulunmuyor"})
		return
	}

	c.JSON(http.StatusOK, rev)
}
