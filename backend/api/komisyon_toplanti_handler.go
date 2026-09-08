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

// GetBekleyenProjeler komisyon onayı bekleyen projeleri listeler.
// GET /api/komisyon/bekleyen-projeler
func (h *KomisyonToplantiHandler) GetBekleyenProjeler(c *gin.Context) {
	projeler, err := h.ToplantiService.GetBekleyenProjeler()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"projeler": projeler})
}

// AddProjeToToplanti toplantıya proje veya talep ekler.
// POST /api/komisyon/toplanti/:id/projeler
// Türkçe Yorum: Raportör, komisyon_bekliyor durumundaki projeleri veya beklemedeki talepleri toplantı gündemine ekler.
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
		ProjeID      int    `json:"proje_id" binding:"required"`
		TalepID      int    `json:"talep_id"`
		GundemTipi   string `json:"gundem_tipi"`
		TalepTipi    string `json:"talep_tipi"`
		GundemSirasi int    `json:"gundem_sirasi"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "proje_id zorunludur"})
		return
	}

	if err := h.ToplantiService.AddProjeToToplanti(toplantiID, req.ProjeID, req.TalepID, req.GundemTipi, req.TalepTipi, req.GundemSirasi, ekleyenID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Gündem maddesi toplantı gündemine eklendi."})
}

// RemoveProjeFromToplanti toplantıdan proje veya talep çıkarır.
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
	talepID, _ := strconv.Atoi(c.Query("talep_id"))

	if err := h.ToplantiService.RemoveProjeFromToplanti(toplantiID, projeID, talepID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Gündem maddesi toplantı gündeminden çıkarıldı."})
}

// GetProjectsByToplanti bir toplantıdaki projeleri ve talepleri listeler.
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

// SetProjeKarar toplantıdaki bir proje veya talep için karar kaydeder ve durumunu günceller.
// PUT /api/komisyon/toplanti/:id/projeler/:proje_id/karar
// Türkçe Yorum: Başkan/raportör kararını bu endpoint üzerinden gönderir; workflow senkronu serviste yapılır.
func (h *KomisyonToplantiHandler) SetProjeKarar(c *gin.Context) {
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	islemYapanID := int(uyeIDFloat.(float64))

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
		TalepID   int    `json:"talep_id"`
		TalepTipi string `json:"talep_tipi"`
		Karar     string `json:"karar" binding:"required"`
		Aciklama  string `json:"aciklama"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "karar alanı zorunludur"})
		return
	}

	if err := h.ToplantiService.SetProjeKarar(toplantiID, projeID, req.TalepID, islemYapanID, req.Karar, req.Aciklama, req.TalepTipi); err != nil {
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
