package api

import (
	"net/http"
	"strings"

	"bap_ai/backend/models"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// FeedbackHandler geri bildirim HTTP uç noktalarını yönetir.
// Türkçe Yorum: Geri bildirim alma isteğini karşılayan ve rol bazlı erişim kontrolünü doğrulayan handler katmanı.
type FeedbackHandler struct {
	FeedbackService *service.FeedbackService
	AdminService    *service.AdminService
}

// NewFeedbackHandler yeni bir FeedbackHandler örneği oluşturur.
// Türkçe Yorum: FeedbackHandler yapısını başlatan kurucu fonksiyon.
func NewFeedbackHandler(feedbackService *service.FeedbackService, adminService *service.AdminService) *FeedbackHandler {
	return &FeedbackHandler{
		FeedbackService: feedbackService,
		AdminService:    adminService,
	}
}

// SubmitFeedback kullanıcının gönderdiği geri bildirimi işler.
// Türkçe Yorum: Rol yetkisini sorgular, isteği doğrular ve e-posta/DB sürecini başlatır.
func (h *FeedbackHandler) SubmitFeedback(c *gin.Context) {
	// Context'ten rol ve üye ID bilgilerini al
	role, exists := c.Get("role")
	if !exists {
		role = "akademisyen"
	}
	roleStr, _ := role.(string)

	uyeIDVal, exists := c.Get("uye_id")
	var uyeID int
	if exists {
		switch v := uyeIDVal.(type) {
		case float64:
			uyeID = int(v)
		case int:
			uyeID = v
		}
	}

	if uyeID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Oturum bilgisi geçersiz"})
		return
	}

	// 1. Admin yetki matrisine göre geri bildirim modülüne erişim hakkı kontrolü
	roles := strings.Split(roleStr, ",")
	allowed, err := h.AdminService.CheckPageAccess(roles, "/api/feedback")
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Geri bildirim modülünü kullanma yetkiniz bulunmamaktadır.",
		})
		return
	}

	// 2. Request body parse et
	var req models.CreateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lütfen mesaj ve konu alanlarını eksiksiz giriniz."})
		return
	}

	// 3. Servise ilet ve işlemi tamamla
	if err := h.FeedbackService.SendFeedback(uyeID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Geri bildirim iletilirken bir sorun oluştu: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Geri bildiriminiz sistem yöneticilerine başarıyla iletildi. Teşekkür ederiz!",
	})
}
