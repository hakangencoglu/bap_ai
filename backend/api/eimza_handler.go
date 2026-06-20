package api

import (
	"fmt"
	"net/http"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// EimzaHandler yapısı e-imza ile ilgili HTTP isteklerini karşılar.
// Türkçe Yorum: E-İmza HTTP endpoint'lerini barındıran handler yapısı.
type EimzaHandler struct {
	EimzaService *service.EimzaService
	UyeRepo      *repository.UyeRepository
}

// NewEimzaHandler yeni bir EimzaHandler oluşturur.
// Türkçe Yorum: Yeni EimzaHandler nesnesini döndüren kurucu fonksiyon.
func NewEimzaHandler(eimzaService *service.EimzaService, uyeRepo *repository.UyeRepository) *EimzaHandler {
	return &EimzaHandler{
		EimzaService: eimzaService,
		UyeRepo:      uyeRepo,
	}
}

// GetPendingSignatures kullanıcı için e-imza bekleyen belgeleri döner.
// GET /api/eimza/pending
// Türkçe Yorum: Kullanıcının rolüne göre imza bekleyen projeleri listeleyen endpoint.
func (h *EimzaHandler) GetPendingSignatures(c *gin.Context) {
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Rol bilgisi bulunamadı"})
		return
	}
	roleStr, ok := role.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz rol formatı"})
		return
	}

	projeler, err := h.EimzaService.GetPendingSignatures(uyeID, roleStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bekleyen imzalar getirilemedi: " + err.Error()})
		return
	}

	if projeler == nil {
		projeler = []models.Proje{}
	}

	c.JSON(http.StatusOK, gin.H{"projects": projeler})
}

// GetSignedDocuments kullanıcının daha önce e-imza ile imzaladığı belgeleri döner.
// GET /api/eimza/signed
// Türkçe Yorum: Kullanıcının e-imzayla imzaladığı projelerin geçmişini listeleyen endpoint.
func (h *EimzaHandler) GetSignedDocuments(c *gin.Context) {
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	projeler, err := h.EimzaService.GetSignedDocuments(uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İmzalanmış belgeler getirilemedi: " + err.Error()})
		return
	}

	if projeler == nil {
		projeler = []models.Proje{}
	}

	c.JSON(http.StatusOK, gin.H{"projects": projeler})
}

// SignDocument belgeyi e-imza ile imzalar.
// POST /api/eimza/sign
// Türkçe Yorum: Projeyi simüle edilmiş mobil/akıllı kart yöntemiyle imzalayan endpoint.
func (h *EimzaHandler) SignDocument(c *gin.Context) {
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Rol bilgisi bulunamadı"})
		return
	}
	roleStr, ok := role.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz rol formatı"})
		return
	}

	var req models.EimzaSignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "İstek gövdesi geçersiz: " + err.Error()})
		return
	}

	// Kullanıcı bilgilerinden ad-soyadı getir
	uye, err := h.UyeRepo.GetUyeByID(uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcı detayları alınamadı: " + err.Error()})
		return
	}

	imzaciAdSoyad := fmt.Sprintf("%s %s", uye.Ad, uye.Soyad)
	if uye.Unvan != "" {
		imzaciAdSoyad = fmt.Sprintf("%s %s %s", uye.Unvan, uye.Ad, uye.Soyad)
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// İmzalama işlemini gerçekleştir
	err = h.EimzaService.SignDocument(req.ProjeID, uyeID, roleStr, imzaciAdSoyad, req.Yontem, req.PinKodu, ip, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Belge imzalanamadı: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Belge başarıyla e-imza ile imzalandı",
		"proje_id":  req.ProjeID,
		"imzaci":    imzaciAdSoyad,
		"ip_adresi": ip,
	})
}
