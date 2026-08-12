package api

import (
	"net/http"
	"strconv"

	"bap_ai/backend/models"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// ZamanlanmisGorevHandler admin tarafındaki zamanlanmış görev kural ve log isteklerini karşılar.
// Türkçe Yorum: Kural CRUD, manuel tetikleme ve gönderim geçmişi API endpoint'lerini yönetir.
type ZamanlanmisGorevHandler struct {
	Service *service.ZamanlanmisGorevService
}

// NewZamanlanmisGorevHandler yeni bir ZamanlanmisGorevHandler oluşturur.
func NewZamanlanmisGorevHandler(service *service.ZamanlanmisGorevService) *ZamanlanmisGorevHandler {
	return &ZamanlanmisGorevHandler{Service: service}
}

// GetAllRules tüm zamanlanmış görev kurallarını döner.
// GET /api/admin/zamanlanmis-gorev/kurallar
func (h *ZamanlanmisGorevHandler) GetAllRules(c *gin.Context) {
	kurallar, err := h.Service.Repo.GetAllRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"kurallar": kurallar})
}

// CreateRule yeni bir kural ekler.
// POST /api/admin/zamanlanmis-gorev/kural
func (h *ZamanlanmisGorevHandler) CreateRule(c *gin.Context) {
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı doğrulaması başarısız"})
		return
	}
	olusturanID := int(uyeIDFloat.(float64))

	var req models.ZamanlanmisGorevKural
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri formatı: " + err.Error()})
		return
	}

	req.OlusturanID = &olusturanID
	if req.KuralAdi == "" || req.TetiklemeTipi == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kural_adi ve tetikleme_tipi zorunludur"})
		return
	}

	if err := h.Service.Repo.CreateRule(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kural kaydedilemedi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kural başarıyla eklendi", "kural": req})
}

// UpdateRule mevcut bir kuralı günceller.
// PUT /api/admin/zamanlanmis-gorev/kural/:id
func (h *ZamanlanmisGorevHandler) UpdateRule(c *gin.Context) {
	kuralID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz kural ID'si"})
		return
	}

	var req models.ZamanlanmisGorevKural
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri formatı: " + err.Error()})
		return
	}
	req.KuralID = kuralID

	if err := h.Service.Repo.UpdateRule(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kural güncellenemedi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kural başarıyla güncellendi", "kural": req})
}

// DeleteRule bir kuralı siler.
// DELETE /api/admin/zamanlanmis-gorev/kural/:id
func (h *ZamanlanmisGorevHandler) DeleteRule(c *gin.Context) {
	kuralID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz kural ID'si"})
		return
	}

	if err := h.Service.Repo.DeleteRule(kuralID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kural silinemedi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kural silindi"})
}

// TriggerManually zamanlanmış görevleri hemen tarar ve sonuçları döner.
// POST /api/admin/zamanlanmis-gorev/calistir
func (h *ZamanlanmisGorevHandler) TriggerManually(c *gin.Context) {
	sonuc, err := h.Service.ProcessScheduledRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Manuel çalıştırma hatası: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Zamanlanmış görevler başarıyla çalıştırıldı",
		"sonuc":   sonuc,
	})
}

// GetLogs gönderim loglarını listeler.
// GET /api/admin/zamanlanmis-gorev/loglar
func (h *ZamanlanmisGorevHandler) GetLogs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)

	loglar, err := h.Service.Repo.GetLogs(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Loglar getirilemedi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"loglar": loglar})
}
