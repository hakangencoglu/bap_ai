package api

import (
	"net/http"
	"strconv"

	"bap_ai/backend/models"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// SatinalmaHandler satın alma HTTP isteklerini işler.
// Türkçe Yorum: Satın alma talebi oluşturma, listeleme ve TTO onaylama isteklerini karşılayan handler katmanıdır.
type SatinalmaHandler struct {
	Service *service.SatinalmaService
}

// NewSatinalmaHandler yeni bir SatinalmaHandler nesnesi döner.
// Türkçe Yorum: SatinalmaHandler için dependency injection kurucusu.
func NewSatinalmaHandler(srv *service.SatinalmaService) *SatinalmaHandler {
	return &SatinalmaHandler{Service: srv}
}

// CreatePurchaseRequest akademisyenin satın alma talebi oluşturmasını sağlar.
// POST /api/satinalma/talep
func (h *SatinalmaHandler) CreatePurchaseRequest(c *gin.Context) {
	// 1. Giriş yapan üye bilgilerini al
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	// Giriş yapan üyenin rolünü al
	roleVal, existsRole := c.Get("role")
	roleStr := ""
	if existsRole {
		roleStr, _ = roleVal.(string)
	}

	// 2. İstek gövdesini bind et
	var req struct {
		ProjeID    int     `json:"proje_id" binding:"required"`
		KalemID    int     `json:"kalem_id" binding:"required"`
		MalzemeAdi string  `json:"malzeme_adi" binding:"required"`
		Miktar     int     `json:"miktar" binding:"required"`
		BirimFiyat float64 `json:"birim_fiyat" binding:"required"`
		Gerekce    string  `json:"gerekce" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek parametreleri"})
		return
	}

	// 3. Model oluştur ve servise yolla
	t := &models.SatinalmaTalebi{
		ProjeID:    req.ProjeID,
		UyeID:      uyeID,
		KalemID:    req.KalemID,
		MalzemeAdi: req.MalzemeAdi,
		Miktar:     req.Miktar,
		BirimFiyat: req.BirimFiyat,
		Gerekce:    req.Gerekce,
	}

	err := h.Service.CreatePurchaseRequest(t, roleStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Satın alma talebi başarıyla oluşturuldu",
		"talep":   t,
	})
}

// GetPurchaseRequestsByProject bir projeye ait tüm satın alma taleplerini döner.
// GET /api/satinalma/proje/:id
func (h *SatinalmaHandler) GetPurchaseRequestsByProject(c *gin.Context) {
	projeIDStr := c.Param("id")
	projeID, err := strconv.Atoi(projeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	// 1. Giriş yapan üye bilgilerini al
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	// Giriş yapan üyenin rolünü al
	roleVal, existsRole := c.Get("role")
	roleStr := ""
	if existsRole {
		roleStr, _ = roleVal.(string)
	}

	talepler, err := h.Service.GetPurchaseRequestsByProject(projeID, uyeID, roleStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"talepler": talepler})
}

// GetAllPurchaseRequests tüm projelerdeki satın alma taleplerini listeler.
// GET /api/satinalma/tum
func (h *SatinalmaHandler) GetAllPurchaseRequests(c *gin.Context) {
	talepler, err := h.Service.GetAllPurchaseRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"talepler": talepler})
}

// HandlePurchaseApproval satın alma talebini TTO yetkilisi adına onaylar veya reddeder.
// POST /api/satinalma/onay
func (h *SatinalmaHandler) HandlePurchaseApproval(c *gin.Context) {
	var req struct {
		TalepID   int    `json:"talep_id" binding:"required"`
		Status    string `json:"status" binding:"required"` // Onaylandı, Reddedildi
		RedNedeni string `json:"red_nedeni"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek parametreleri"})
		return
	}

	if req.Status != "Onaylandı" && req.Status != "Reddedildi" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz karar durumu. Sadece 'Onaylandı' veya 'Reddedildi' kabul edilir."})
		return
	}

	if req.Status == "Reddedildi" && req.RedNedeni == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reddedilen talepler için red nedeni girilmesi zorunludur."})
		return
	}

	err := h.Service.UpdatePurchaseStatus(req.TalepID, req.Status, req.RedNedeni)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Satın alma talebi başarıyla güncellendi"})
}
