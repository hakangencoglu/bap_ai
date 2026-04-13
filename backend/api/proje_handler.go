package api

import (
	"net/http"
	"strconv"

	"bap_ai/backend/models"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// ProjeHandler yapısı proje HTTP isteklerini karşılar.
type ProjeHandler struct {
	ProjeService *service.ProjeService
}

// NewProjeHandler yeni bir ProjeHandler oluşturur.
func NewProjeHandler(projeService *service.ProjeService) *ProjeHandler {
	return &ProjeHandler{ProjeService: projeService}
}

// CreateProje yeni bir proje başvurusu kabul eder.
// POST /api/proje
func (h *ProjeHandler) CreateProje(c *gin.Context) {
	// Middleware'den giriş yapan üye ID'sini alıyoruz
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	// Sadece modeldeki gerekli olan verileri bağlayacağız (c.BindJSON veya bind için bir DTO)
	var req models.Proje
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek formu"})
		return
	}

	roleIDFloat, _ := c.Get("role_id")
	roleID := int(roleIDFloat.(float64))

	// Rol: 3 (Öğrenci) kontrolü - Sadece bap-100 ve bap-200'e başvurabilir
	if roleID == 3 {
		if req.Tur != "bap-100" && req.Tur != "bap-200" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Öğrenci hesabıyla sadece BAP-100 ve BAP-200 türünde başvuru yapabilirsiniz."})
			return
		}
	}

	// Service katmanına iletiyoruz.
	if err := h.ProjeService.CreateProje(uyeID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proje kaydedilemedi"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Proje başvurusu başarıyla kaydedildi"})
}

// GetUyeler projenin kayıtlı üyelerini döner
// GET /api/proje/:id/uyeler
func (h *ProjeHandler) GetUyeler(c *gin.Context) {
	projeID := c.Param("id")
	if projeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	// strconv ile int'e çevirip repository'den üyeleri alabiliriz.
	importStr, _ := strconv.Atoi(projeID)
	uyeler, err := h.ProjeService.ProjeRepo.GetProjeUyeleri(importStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Uyeler getirilirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"uyeler": uyeler})
}

// GetProje projeyi ID'sine göre getirir
// GET /api/proje/:id
func (h *ProjeHandler) GetProje(c *gin.Context) {
	projeID := c.Param("id")
	if projeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}
	id, _ := strconv.Atoi(projeID)

	p, err := h.ProjeService.GetProjeByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proje bulunamadı"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// UpdateProje mevcut projeyi günceller (Revizyon giderme/Düzeltme için)
// PUT /api/proje/:id
func (h *ProjeHandler) UpdateProje(c *gin.Context) {
	projeID := c.Param("id")
	id, _ := strconv.Atoi(projeID)

	var req models.Proje
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek formu"})
		return
	}
	
	req.ProjeID = id

	// Burada revizyonu tamamlanmış duruma getirmek için proje durumunu girmeliyiz
	// Tasarıma göre, öğrenci revizeyi gönderdiğinde süreç "Akademisyen Onayına" dönecek. Biz burada "incelemede" kullanabiliriz.
	// Status should be set back to what is needed
	req.Durum = "incelemede" // "Yürütücü Onayında" olarak sistemde ayrı bir durum yoksa incelemede yeterlidir.

	if err := h.ProjeService.UpdateProje(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proje güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proje başarıyla güncellendi"})
}
