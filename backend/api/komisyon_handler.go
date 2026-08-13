package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bap_ai/backend/models"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// KomisyonHandler komisyon toplantı HTTP isteklerini karşılar.
// Türkçe Yorum: Toplantı oluşturma, katılımcı yoklama listesi getirme ve tutanak PDF'i üretme isteklerini yönetir.
type KomisyonHandler struct {
	KomisyonService *service.KomisyonService
	ToplantiService *service.KomisyonToplantiService
	PdfService      *service.PdfService
}

// NewKomisyonHandler yeni bir KomisyonHandler oluşturur.
func NewKomisyonHandler(ks *service.KomisyonService, ts *service.KomisyonToplantiService, ps *service.PdfService) *KomisyonHandler {
	return &KomisyonHandler{
		KomisyonService: ks,
		ToplantiService: ts,
		PdfService:      ps,
	}
}

// GetCommissionMembers komisyon üye yoklama listesini döner.
// GET /api/komisyon/uyeler
// Türkçe Yorum: Komisyon toplantısında yoklama listesi oluşturulabilmesi için komisyon rolüne sahip tüm aktif üyeleri çeker.
func (h *KomisyonHandler) GetCommissionMembers(c *gin.Context) {
	uyeler, err := h.KomisyonService.GetCommissionMembers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"uyeler": uyeler})
}

// GetNextMeetingNumber sıradaki toplantı numarasını döner.
// GET /api/komisyon/toplanti/next-no
// Türkçe Yorum: Seçilen tarihin yılındaki toplam toplantı sayısına göre otomatik sıradaki numarayı (YIL/SIRA) döner.
func (h *KomisyonHandler) GetNextMeetingNumber(c *gin.Context) {
	dateStr := c.Query("date")
	var meetingDate time.Time
	var err error

	if dateStr != "" {
		meetingDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			// Alternatif ISO/datetime-local formatını dene
			meetingDate, err = time.Parse("2006-01-02T15:04", dateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz tarih formatı (YYYY-MM-DD olmalıdır)"})
				return
			}
		}
	} else {
		meetingDate = time.Now()
	}

	nextNo, err := h.KomisyonService.GetNextMeetingNumber(meetingDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"toplanti_no": nextNo})
}

// CreateMeeting yeni bir toplantı kaydeder.
// POST /api/komisyon/toplanti
// Türkçe Yorum: Gündem, kararlar ve katılımcı yoklama durumlarını alarak yeni bir toplantı kaydı açar.
func (h *KomisyonHandler) CreateMeeting(c *gin.Context) {
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	var req struct {
		Tarih        string `json:"tarih" binding:"required"`
		Gundem       string `json:"gundem" binding:"required"`
		Karar        string `json:"karar" binding:"required"`
		Katilimcilar []struct {
			UyeID   int  `json:"uye_id" binding:"required"`
			Katildi bool `json:"katildi"`
		} `json:"katilimcilar" binding:"required"`
		Projeler []struct {
			ProjeID      int `json:"proje_id" binding:"required"`
			GundemSirasi int `json:"gundem_sirasi"`
		} `json:"projeler"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek parametreleri"})
		return
	}

	meetingDate, err := time.Parse("2006-01-02", req.Tarih)
	if err != nil {
		meetingDate, err = time.Parse("2006-01-02T15:04", req.Tarih)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz tarih formatı"})
			return
		}
	}

	// Katılımcıları model formatına dönüştür
	var katilimcilar []*models.KomisyonToplantiKatilim
	// Yoklama verilerini eşleştirmek için komisyon üyelerini de çekiyoruz (Ad/Soyad/Unvan/Bölüm için)
	uyeler, err := h.KomisyonService.GetCommissionMembers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Üye listesi doğrulanamadı"})
		return
	}

	uyeMap := make(map[int]*models.UyeWithDetay)
	for _, u := range uyeler {
		uyeMap[u.UyeID] = u
	}

	for _, k := range req.Katilimcilar {
		u, ok := uyeMap[k.UyeID]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz katılımcı üye ID'si: " + strconv.Itoa(k.UyeID)})
			return
		}
		katilimcilar = append(katilimcilar, &models.KomisyonToplantiKatilim{
			UyeID:   k.UyeID,
			Ad:      u.Ad,
			Soyad:   u.Soyad,
			Unvan:   u.Unvan,
			Bolum:   u.Bolum,
			Katildi: k.Katildi,
		})
	}

	meeting := &models.KomisyonToplantisi{
		Tarih:        meetingDate,
		Gundem:       req.Gundem,
		Karar:        req.Karar,
		Durum:        "tamamlandi",
		OlusturanID:  uyeID,
		Katilimcilar: katilimcilar,
	}

	err = h.KomisyonService.CreateMeeting(meeting)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Türkçe Yorum: Toplantıya seçilen projeler köprü tabloya bekliyor kararıyla eklenir.
	var eklenen []int
	var eklemeHatalari []string
	if h.ToplantiService != nil {
		for i, p := range req.Projeler {
			sira := p.GundemSirasi
			if sira <= 0 {
				sira = i + 1
			}
			if err := h.ToplantiService.AddProjeToToplanti(meeting.ToplantiID, p.ProjeID, sira, uyeID); err != nil {
				eklemeHatalari = append(eklemeHatalari, fmt.Sprintf("proje %d: %s", p.ProjeID, err.Error()))
				continue
			}
			eklenen = append(eklenen, p.ProjeID)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":         "Komisyon toplantısı başarıyla kaydedildi.",
		"toplanti":        meeting,
		"eklenen_projeler": eklenen,
		"proje_hatalari":  eklemeHatalari,
	})
}

// GetMeetingPDF toplantı raporunu PDF formatında üretir.
// GET /api/komisyon/toplanti/:id/pdf
// Türkçe Yorum: Toplantıyı ID'sine göre sorgular ve DejaVuSans Türkçe font destekli PDF dosyasını tarayıcıya yollar.
func (h *KomisyonHandler) GetMeetingPDF(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz toplantı ID'si"})
		return
	}

	meeting, err := h.KomisyonService.GetMeetingByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if meeting == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Toplantı bulunamadı"})
		return
	}

	pdfBytes, err := h.PdfService.GenerateCommissionMeetingPDF(meeting)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("BAP_Komisyon_Toplanti_%s.pdf", strings.ReplaceAll(meeting.ToplantiNo, "/", "-"))
	c.Header("Content-Disposition", "inline; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// GetMeetingsList tüm komisyon toplantılarının listesini döner.
// GET /api/komisyon/toplantilar
// Türkçe Yorum: Geçmiş toplantıları listelemek için kullanılır.
func (h *KomisyonHandler) GetMeetingsList(c *gin.Context) {
	toplantilar, err := h.KomisyonService.ListMeetings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"toplantilar": toplantilar})
}

// DeleteMeeting komisyon toplantısını kalıcı siler (yalnızca admin1).
// DELETE /api/komisyon/toplanti/:id
func (h *KomisyonHandler) DeleteMeeting(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz toplantı ID"})
		return
	}
	if err := h.KomisyonService.DeleteMeeting(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Toplantı silindi"})
}

// PreviewMeetingPDF toplantı kararını kaydetmeden önce PDF formatında önizleme olarak üretir.
// POST /api/komisyon/toplanti/preview-pdf
// Türkçe Yorum: Veritabanına kaydetmeden, sadece gelen form parametreleriyle geçici bir komisyon toplantısı PDF'i oluşturup döner.
func (h *KomisyonHandler) PreviewMeetingPDF(c *gin.Context) {
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	var req struct {
		Tarih        string `json:"tarih" binding:"required"`
		Gundem       string `json:"gundem" binding:"required"`
		Karar        string `json:"karar" binding:"required"`
		Katilimcilar []struct {
			UyeID   int  `json:"uye_id" binding:"required"`
			Katildi bool `json:"katildi"`
		} `json:"katilimcilar" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek parametreleri"})
		return
	}

	meetingDate, err := time.Parse("2006-01-02", req.Tarih)
	if err != nil {
		meetingDate, err = time.Parse("2006-01-02T15:04", req.Tarih)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz tarih formatı"})
			return
		}
	}

	// Katılımcıları model formatına dönüştür
	var katilimcilar []*models.KomisyonToplantiKatilim
	uyeler, err := h.KomisyonService.GetCommissionMembers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Üye listesi doğrulanamadı"})
		return
	}

	uyeMap := make(map[int]*models.UyeWithDetay)
	for _, u := range uyeler {
		uyeMap[u.UyeID] = u
	}

	for _, k := range req.Katilimcilar {
		u, ok := uyeMap[k.UyeID]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz katılımcı üye ID'si: " + strconv.Itoa(k.UyeID)})
			return
		}
		katilimcilar = append(katilimcilar, &models.KomisyonToplantiKatilim{
			UyeID:   k.UyeID,
			Ad:      u.Ad,
			Soyad:   u.Soyad,
			Unvan:   u.Unvan,
			Bolum:   u.Bolum,
			Katildi: k.Katildi,
		})
	}

	// Geçici bir toplantı numarası üret (yıl / ÖNİZLEME)
	year := meetingDate.Year()
	toplantiNo := fmt.Sprintf("%d/ÖNİZLEME", year)

	meeting := &models.KomisyonToplantisi{
		ToplantiNo:   toplantiNo,
		Tarih:        meetingDate,
		Gundem:       req.Gundem,
		Karar:        req.Karar,
		OlusturanID:  uyeID,
		Katilimcilar: katilimcilar,
	}

	pdfBytes, err := h.PdfService.GenerateCommissionMeetingPDF(meeting)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF önizlemesi oluşturulamadı: " + err.Error()})
		return
	}

	c.Header("Content-Disposition", "inline; filename=BAP_Komisyon_Toplanti_Onizleme.pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
