package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"bap_ai/backend/models"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// RaporHandler yapısı HTTP isteklerini işler.
type RaporHandler struct {
	Service *service.RaporService
}

// NewRaporHandler yeni bir RaporHandler oluşturur.
func NewRaporHandler(service *service.RaporService) *RaporHandler {
	return &RaporHandler{Service: service}
}

// SubmitAraRapor, yürütücünün ara rapor veya sonuç raporu yüklemesini işler.
// POST /api/proje/:id/ara-rapor
func (h *RaporHandler) SubmitAraRapor(c *gin.Context) {
	projeIDStr := c.Param("id")
	projeID, err := strconv.Atoi(projeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	uyeIDVal, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz erişim"})
		return
	}
	uyeID := uyeIDVal.(int)

	var req models.ProjeAraRapor
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek gövdesi"})
		return
	}

	req.ProjeID = projeID
	if err := h.Service.SubmitAraRapor(uyeID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Rapor başarıyla yüklendi ve onay sürecine sevk edildi", "data": req})
}

// GetProjeRaporlari, projeye ait tüm rapor teslimlerini döner.
// GET /api/proje/:id/ara-raporlar
func (h *RaporHandler) GetProjeRaporlari(c *gin.Context) {
	projeIDStr := c.Param("id")
	projeID, err := strconv.Atoi(projeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	list, err := h.Service.GetProjeRaporlari(projeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Raporlar çekilemedi"})
		return
	}

	c.JSON(http.StatusOK, list)
}

// GetBekleyenRaporlar, TTO ve Admin için inceleme bekleyen raporları döner.
// GET /api/admin/ara-raporlar/bekleyen
func (h *RaporHandler) GetBekleyenRaporlar(c *gin.Context) {
	list, err := h.Service.GetBekleyenRaporlar()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bekleyen raporlar getirilemedi"})
		return
	}

	c.JSON(http.StatusOK, list)
}

// DegerlendirRapor, TTO / Admin yetkilisinin rapor inceleme kararını kaydeder.
// POST /api/admin/ara-rapor/degerlendir
func (h *RaporHandler) DegerlendirRapor(c *gin.Context) {
	uyeIDVal, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz erişim"})
		return
	}
	uyeID := uyeIDVal.(int)

	var req models.RaporDegerlendirmeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek gövdesi"})
		return
	}

	if req.Durum != "onaylandi" && req.Durum != "revizyon" && req.Durum != "reddedildi" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz durum (onaylandi, revizyon veya reddedildi olmalı)"})
		return
	}

	if err := h.Service.DegerlendirRapor(uyeID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rapor değerlendirmesi başarıyla kaydedildi"})
}

// UploadRaporDosya, ara rapor dosyasını (PDF/Word/ZIP) sunucuya yükler ve dosya URL'sini döner.
// POST /api/proje/:id/ara-rapor/upload
func (h *RaporHandler) UploadRaporDosya(c *gin.Context) {
	projeIDStr := c.Param("id")
	projeID, err := strconv.Atoi(projeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	file, err := c.FormFile("dosya")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lütfen yüklenecek dosyayı seçiniz"})
		return
	}

	uploadDir := filepath.Join("./uploads/ara_raporlar", fmt.Sprintf("%d", projeID))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Sunucu klasörü oluşturulamadı"})
		return
	}

	ext := filepath.Ext(file.Filename)
	uniqueFilename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), sanitizeRaporFilename(file.Filename), ext)
	dst := filepath.Join(uploadDir, uniqueFilename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Dosya kaydedilemedi"})
		return
	}

	fileURL := fmt.Sprintf("/uploads/ara_raporlar/%d/%s", projeID, uniqueFilename)
	c.JSON(http.StatusOK, gin.H{
		"message": "Dosya başarıyla yüklendi",
		"dosya_url": fileURL,
		"orijinal_dosya_adi": file.Filename,
	})
}

func sanitizeRaporFilename(name string) string {
	name = filepath.Base(name)
	ext := filepath.Ext(name)
	nameWithoutExt := name[:len(name)-len(ext)]
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	safe := reg.ReplaceAllString(nameWithoutExt, "_")
	if len(safe) > 50 {
		safe = safe[:50]
	}
	return safe
}
