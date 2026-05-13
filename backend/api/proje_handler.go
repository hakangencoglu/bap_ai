package api

import (
	"net/http"
	"strconv"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// ProjeHandler yapısı proje HTTP isteklerini karşılar.
type ProjeHandler struct {
	ProjeService *service.ProjeService
	UyeRepo      *repository.UyeRepository
	DavetRepo    *repository.DavetRepository
}

// NewProjeHandler yeni bir ProjeHandler oluşturur.
func NewProjeHandler(projeService *service.ProjeService, uyeRepo *repository.UyeRepository, davetRepo *repository.DavetRepository) *ProjeHandler {
	return &ProjeHandler{ProjeService: projeService, UyeRepo: uyeRepo, DavetRepo: davetRepo}
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

	var req models.Proje
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek formu"})
		return
	}

	// Rol kontrolü: artık string rol kullanıyoruz
	role, _ := c.Get("role")
	roleStr, _ := role.(string)

	// Öğrenci sadece belirli BAP türlerine başvurabilir
	if roleStr == "ogrenci" {
		// bap_turu_id ile kontrol (1=Yüksek Lisans, 2=Doktora)
		if req.BapTuruID != nil && *req.BapTuruID > 2 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Öğrenci hesabıyla sadece Yüksek Lisans ve Doktora projelerine başvurabilirsiniz."})
			return
		}
	}

	// Service katmanına rol bilgisiyle birlikte iletiyoruz.
	if err := h.ProjeService.CreateProje(uyeID, &req, roleStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proje kaydedilemedi"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Proje başvurusu başarıyla kaydedildi", "proje_id": req.ProjeID})
}

// GetUyeler projenin kayıtlı üyelerini döner
// GET /api/proje/:id/uyeler
func (h *ProjeHandler) GetUyeler(c *gin.Context) {
	projeID := c.Param("id")
	if projeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	id, _ := strconv.Atoi(projeID)
	uyeler, err := h.ProjeService.ProjeRepo.GetProjeUyeleri(id)
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

	// Durumu "incelemede" olarak ayarla (durum_id=2)
	durumID := 2
	req.DurumID = &durumID

	if err := h.ProjeService.UpdateProje(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proje güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proje başarıyla güncellendi"})
}

// GetAkademisyenler akademisyen rolündeki kullanıcıları döner (Yürütücü seçimi için)
// GET /api/akademisyenler
func (h *ProjeHandler) GetAkademisyenler(c *gin.Context) {
	uyeler, err := h.UyeRepo.GetUyelerByRol("akademisyen")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Akademisyen listesi alınamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"akademisyenler": uyeler})
}

// AddTeamMember projeye yeni ekip üyesi ekler (davet durumu: beklemede)
// POST /api/proje/:id/takim
func (h *ProjeHandler) AddTeamMember(c *gin.Context) {
	projeID := c.Param("id")
	if projeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}
	id, _ := strconv.Atoi(projeID)

	// İstek gövdesini parse et
	var req struct {
		UyeID int `json:"uye_id" binding:"required"`
		RolID int `json:"rol_id"` // Varsayılan: 2 (Araştırmacı)
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek formatı"})
		return
	}

	// Rol ID varsayılanı Araştırmacı (2)
	if req.RolID == 0 {
		req.RolID = 2
	}

	// Davet ile ekle (davet_durumu = beklemede)
	if err := h.DavetRepo.AddTeamMemberWithInvite(id, req.UyeID, req.RolID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ekip üyesi eklenemedi"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Ekip üyesine davet gönderildi"})
}

// SearchUyeler kayıtlı kullanıcılar arasında arama yapar (ekip üyesi ekleme için)
// GET /api/uyeler/ara?q=...
func (h *ProjeHandler) SearchUyeler(c *gin.Context) {
	query := c.Query("q")
	if query == "" || len(query) < 2 {
		c.JSON(http.StatusOK, gin.H{"uyeler": []interface{}{}})
		return
	}

	uyeler, err := h.UyeRepo.SearchUyeler(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcı araması başarısız"})
		return
	}

	// Null kontrolü — boş dizi dön
	if uyeler == nil {
		uyeler = []repository.UyeAramaOzet{}
	}

	c.JSON(http.StatusOK, gin.H{"uyeler": uyeler})
}
