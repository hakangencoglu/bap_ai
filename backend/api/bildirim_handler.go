package api

import (
	"net/http"
	"strconv"

	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// BildirimHandler yapısı bildirim HTTP isteklerini karşılar.
// Türkçe Yorum: Kullanıcıların bildirim listesini almasını ve okundu işlemleri yapmasını sağlayan HTTP denetleyicisidir.
type BildirimHandler struct {
	BildirimService *service.BildirimService
}

// NewBildirimHandler yeni bir BildirimHandler oluşturur.
// Türkçe Yorum: BildirimHandler için dependency injection kurucusu.
func NewBildirimHandler(bildirimService *service.BildirimService) *BildirimHandler {
	return &BildirimHandler{BildirimService: bildirimService}
}

// GetBildirimler kullanıcının bildirimlerini döner.
// GET /api/bildirimler
// Türkçe Yorum: Giriş yapan kullanıcının veritabanındaki tüm sistem içi bildirimlerini JSON formatında döner.
func (h *BildirimHandler) GetBildirimler(c *gin.Context) {
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	list, err := h.BildirimService.GetBildirimler(uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bildirimler alınamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bildirimler": list})
}

// MarkAsRead belirtilen bildirimi okundu olarak işaretler.
// POST /api/bildirimler/:id/oku
// Türkçe Yorum: URL'den alınan bildirim ID'sine göre ilgili kaydı okundu olarak işaretler.
func (h *BildirimHandler) MarkAsRead(c *gin.Context) {
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	bildirimIDStr := c.Param("id")
	bildirimID, err := strconv.Atoi(bildirimIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz bildirim ID"})
		return
	}

	err = h.BildirimService.MarkAsRead(bildirimID, uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bildirim güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bildirim okundu olarak işaretlendi"})
}

// MarkAllAsRead tüm bildirimleri okundu olarak işaretler.
// POST /api/bildirimler/oku-hepsi
// Türkçe Yorum: Giriş yapan kullanıcının tüm okunmamış bildirimlerini tek seferde okundu olarak günceller.
func (h *BildirimHandler) MarkAllAsRead(c *gin.Context) {
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	err := h.BildirimService.MarkAllAsRead(uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bildirimler güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tüm bildirimler okundu olarak işaretlendi"})
}
