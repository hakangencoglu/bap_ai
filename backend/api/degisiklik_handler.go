package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// DegisiklikHandler, proje değişiklik geçmişi HTTP isteklerini işler.
type DegisiklikHandler struct {
	Service *service.DegisiklikService
}

// NewDegisiklikHandler yeni handler oluşturur.
func NewDegisiklikHandler(srv *service.DegisiklikService) *DegisiklikHandler {
	return &DegisiklikHandler{Service: srv}
}

// ListDegisiklikler, projeye ait değişiklik listesini döner.
// GET /api/proje/:id/degisiklikler
func (h *DegisiklikHandler) ListDegisiklikler(c *gin.Context) {
	projeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}
	list, err := h.Service.ListByProje(projeID, c.Query("olay_tipi"), c.Query("kaynak_tip"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Değişiklikler alınamadı"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"degisiklikler": list})
}

// GetDegisiklik, tek değişiklik detayını (önce/sonra) döner.
// GET /api/proje/:id/degisiklikler/:degisiklik_id
func (h *DegisiklikHandler) GetDegisiklik(c *gin.Context) {
	projeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}
	degID, err := strconv.Atoi(c.Param("degisiklik_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz değişiklik ID"})
		return
	}
	item, err := h.Service.GetByID(projeID, degID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, service.ErrDegisiklikProjeUyusmazligi) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Değişiklik kaydı bulunamadı"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Değişiklik detayı alınamadı"})
		return
	}
	c.JSON(http.StatusOK, item)
}
