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
// Türkçe Yorum: Akademisyen tarafından gönderilen tekli veya toplu satın alma talebini (JSON içindeki items dizisi veya düz parametrelerle) karşılar ve servis katmanına toplu olarak yollar.
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

	// 2. İstek gövdesini bind et (hem eski düz parametreleri hem de yeni items dizisini destekler)
	var req struct {
		ProjeID    int     `json:"proje_id" binding:"required"`
		KalemID    int     `json:"kalem_id" binding:"required"`
		MalzemeAdi string  `json:"malzeme_adi"`
		Miktar     int     `json:"miktar"`
		BirimFiyat float64 `json:"birim_fiyat"`
		Gerekce    string  `json:"gerekce"`
		Items      []struct {
			MalzemeAdi string  `json:"malzeme_adi" binding:"required"`
			Miktar     int     `json:"miktar" binding:"required"`
			BirimFiyat float64 `json:"birim_fiyat" binding:"required"`
			Gerekce    string  `json:"gerekce" binding:"required"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek parametreleri"})
		return
	}

	// Geriye dönük uyumluluk: Eğer items boşsa ve düz malzeme parametreleri doluysa items'a ekle
	var itemsList []struct {
		MalzemeAdi string  `json:"malzeme_adi" binding:"required"`
		Miktar     int     `json:"miktar" binding:"required"`
		BirimFiyat float64 `json:"birim_fiyat" binding:"required"`
		Gerekce    string  `json:"gerekce" binding:"required"`
	}

	if len(req.Items) > 0 {
		itemsList = req.Items
	} else if req.MalzemeAdi != "" && req.Miktar > 0 && req.BirimFiyat > 0 {
		itemsList = append(itemsList, struct {
			MalzemeAdi string  `json:"malzeme_adi" binding:"required"`
			Miktar     int     `json:"miktar" binding:"required"`
			BirimFiyat float64 `json:"birim_fiyat" binding:"required"`
			Gerekce    string  `json:"gerekce" binding:"required"`
		}{
			MalzemeAdi: req.MalzemeAdi,
			Miktar:     req.Miktar,
			BirimFiyat: req.BirimFiyat,
			Gerekce:    req.Gerekce,
		})
	}

	if len(itemsList) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Satın alınacak malzeme bilgisi eksik"})
		return
	}

	// 3. Modelleri oluştur ve servise yolla
	var talepler []*models.SatinalmaTalebi
	for _, item := range itemsList {
		t := &models.SatinalmaTalebi{
			ProjeID:    req.ProjeID,
			UyeID:      uyeID,
			KalemID:    req.KalemID,
			MalzemeAdi: item.MalzemeAdi,
			Miktar:     item.Miktar,
			BirimFiyat: item.BirimFiyat,
			Gerekce:    item.Gerekce,
		}
		talepler = append(talepler, t)
	}

	err := h.Service.CreatePurchaseRequests(talepler, roleStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Satın alma talebi/talepleri başarıyla oluşturuldu",
		"talepler": talepler,
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

// RevisePurchaseRequest satın alma talebi için bütçe/fiyat revizyonu yapar.
// POST /api/satinalma/revize
// Türkçe Yorum: Admin veya TTO yetkilisi tarafından gönderilen yeni birim fiyat ve gerekçeyi alarak ilgili bütçe kalemini günceller.
func (h *SatinalmaHandler) RevisePurchaseRequest(c *gin.Context) {
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
		TalepID        int     `json:"talep_id" binding:"required"`
		YeniBirimFiyat float64 `json:"yeni_birim_fiyat" binding:"required,gt=0"`
		Gerekce        string  `json:"gerekce" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veya eksik istek parametreleri"})
		return
	}

	// 3. Servis katmanını çağır
	err := h.Service.RevisePurchaseRequest(req.TalepID, req.YeniBirimFiyat, req.Gerekce, uyeID, roleStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Satın alma talebi başarıyla revize edildi"})
}

// GetProjectBudgetReport projenin bütçe kalemi bazlı harcama raporunu döner.
// GET /api/satinalma/proje/:id/butce-raporu
// Türkçe Yorum: Planlanan, harcanan, ödenen ve kalan tutarları bütçe kalemi bazında listeler.
func (h *SatinalmaHandler) GetProjectBudgetReport(c *gin.Context) {
	projeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID'si"})
		return
	}

	rapor, err := h.Service.GetProjectBudgetReport(projeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rapor)
}
