package api

import (
	"net/http"
	"strconv"

	"bap_ai/backend/service"
	"github.com/gin-gonic/gin"
)

// FakulteHandler fakülte ve bölüm HTTP isteklerini yönetir.
type FakulteHandler struct {
	Service *service.FakulteService
}

// NewFakulteHandler yeni bir FakulteHandler oluşturur.
func NewFakulteHandler(service *service.FakulteService) *FakulteHandler {
	return &FakulteHandler{Service: service}
}

// GetFakulteler aktif tüm fakülteleri döner.
// GET /api/fakulteler
// Türkçe Yorum: Profil ve kullanıcı formu üzerindeki fakülte açılır kutusunu doldurur.
func (h *FakulteHandler) GetFakulteler(c *gin.Context) {
	list, err := h.Service.GetFakulteler()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fakülteler alınamadı: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"fakulteler": list})
}

// GetBolumler aktif bölümleri döner. İsteğe bağlı fakulte_id sorgu parametresi alır.
// GET /api/bolumler?fakulte_id=1
// Türkçe Yorum: Kullanıcının seçtiği fakülteye göre bölümleri dinamik filtreler.
func (h *FakulteHandler) GetBolumler(c *gin.Context) {
	fakulteIDStr := c.Query("fakulte_id")
	var fakulteID int
	if fakulteIDStr != "" {
		if id, err := strconv.Atoi(fakulteIDStr); err == nil && id > 0 {
			fakulteID = id
		}
	}

	var list interface{}
	var err error
	if fakulteID > 0 {
		list, err = h.Service.GetBolumler(fakulteID)
	} else {
		list, err = h.Service.GetBolumler()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bölümler alınamadı: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bolumler": list})
}
