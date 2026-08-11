package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// KomisyonToplantiHandler toplantı ↔ proje HTTP isteklerini karşılar.
// Türkçe Yorum: Proje ekleme/çıkarma, karar kaydı ve tutanak PDF endpoint'lerini yönetir.
type KomisyonToplantiHandler struct {
	ToplantiService *service.KomisyonToplantiService
	PdfService      *service.PdfService
}

// NewKomisyonToplantiHandler yeni bir KomisyonToplantiHandler oluşturur.
func NewKomisyonToplantiHandler(ts *service.KomisyonToplantiService, ps *service.PdfService) *KomisyonToplantiHandler {
	return &KomisyonToplantiHandler{
		ToplantiService: ts,
		PdfService:      ps,
	}
}

// AddProjeToToplanti toplantıya proje ekler.
// POST /api/komisyon/toplanti/:id/projeler
// Türkçe Yorum: Raportör, komisyon_bekliyor durumundaki projeleri toplantı gündemine ekler.
func (h *KomisyonToplantiHandler) AddProjeToToplanti(c *gin.Context) {
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	ekleyenID := int(uyeIDFloat.(float64))

	toplantiID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz toplantı ID'si"})
		return
	}

	var req struct {
		ProjeID      int `json:"proje_id" binding:"required"`
		GundemSirasi int `json:"gundem_sirasi"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "proje_id zorunludur"})
		return
	}

	if err := h.ToplantiService.AddProjeToToplanti(toplantiID, req.ProjeID, req.GundemSirasi, ekleyenID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Proje toplantı gündemine eklendi."})
}

// RemoveProjeFromToplanti toplantıdan proje çıkarır.
// DELETE /api/komisyon/toplanti/:id/projeler/:proje_id
func (h *KomisyonToplantiHandler) RemoveProjeFromToplanti(c *gin.Context) {
	toplantiID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz toplantı ID'si"})
		return
	}
	projeID, err := strconv.Atoi(c.Param("proje_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID'si"})
		return
	}
	if err := h.ToplantiService.RemoveProjeFromToplanti(toplantiID, projeID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Proje toplantı gündeminden çıkarıldı."})
}

// GetProjectsByToplanti bir toplantıdaki projeleri listeler.
// GET /api/komisyon/toplanti/:id/projeler
func (h *KomisyonToplantiHandler) GetProjectsByToplanti(c *gin.Context) {
	toplantiID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz toplantı ID'si"})
		return
	}
	projeler, err := h.ToplantiService.GetProjectsByToplanti(toplantiID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"projeler": projeler})
}

// SetProjeKarar toplantıdaki bir proje için karar kaydeder.
// PUT /api/komisyon/toplanti/:id/projeler/:proje_id/karar
// Türkçe Yorum: Raportör onay/red/erteleme kararını bu endpoint üzerinden gönderir.
func (h *KomisyonToplantiHandler) SetProjeKarar(c *gin.Context) {
	toplantiID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz toplantı ID'si"})
		return
	}
	projeID, err := strconv.Atoi(c.Param("proje_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID'si"})
		return
	}

	var req struct {
		Karar    string `json:"karar" binding:"required"`
		Aciklama string `json:"aciklama"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "karar alanı zorunludur"})
		return
	}

	if err := h.ToplantiService.SetProjeKarar(toplantiID, projeID, req.Karar, req.Aciklama); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Karar kaydedildi: " + req.Karar})
}

// GetToplantiBelgePDF toplantı tutanak PDF'ini indirir.
// GET /api/komisyon/toplanti/:id/tutanak
// Türkçe Yorum: Toplantı detayları + katılımcılar + proje kararları tek PDF belgede sunulur.
func (h *KomisyonToplantiHandler) GetToplantiBelgePDF(c *gin.Context) {
	toplantiID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz toplantı ID'si"})
		return
	}

	belge, err := h.ToplantiService.GetToplantiBelgeDetay(toplantiID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pdfBytes, err := h.PdfService.GenerateToplantTutanakPDF(belge)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF tutanak oluşturulamadı: " + err.Error()})
		return
	}

	filename := fmt.Sprintf("BAP_Komisyon_Tutanak_%s.pdf", strings.ReplaceAll(belge.Toplanti.ToplantiNo, "/", "-"))
	c.Header("Content-Disposition", "inline; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
