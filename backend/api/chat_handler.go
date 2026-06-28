package api

import (
	"net/http"
	"strings"

	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// ChatHandler yapısı, sohbet robotu HTTP handler'larını barındırır.
type ChatHandler struct {
	ChatService  *service.ChatService
	AdminService *service.AdminService
}

// NewChatHandler fonksiyonu, yeni bir ChatHandler nesnesi döner.
func NewChatHandler(chatService *service.ChatService, adminService *service.AdminService) *ChatHandler {
	// Türkçe Yorum: ChatHandler nesnesi oluşturularak referansı döndürülür.
	return &ChatHandler{
		ChatService:  chatService,
		AdminService: adminService,
	}
}

// SendMessage fonksiyonu, kullanıcının gönderdiği mesajı işler ve yapay zekadan gelen cevabı döner.
// POST /api/chat
func (h *ChatHandler) SendMessage(c *gin.Context) {
	// Türkçe Yorum: İstek gövdesi için veri yapısı
	var req struct {
		Message string `json:"message" binding:"required"`
	}

	// Gelen JSON verisini bağlama
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Geçersiz mesaj içeriği",
		})
		return
	}

	// Context'ten kullanıcı bilgilerini al
	// Türkçe Yorum: AuthMiddleware tarafından context'e yerleştirilen uye_id, eposta ve role bilgileri okunur.
	role, exists := c.Get("role")
	if !exists {
		role = "akademisyen" // Fallback rol
	}

	email, exists := c.Get("email")
	if !exists {
		email = "kullanici@izu.edu.tr" // Fallback e-posta
	}

	// Rol ve e-posta bilgilerini string olarak cast et
	roleStr, _ := role.(string)
	emailStr, _ := email.(string)

	// Türkçe Yorum: Yöneticinin panel üzerinden belirlediği rol yetki matrisine göre chatbot erişim kontrolü yapılır
	roles := strings.Split(roleStr, ",")
	allowed, err := h.AdminService.CheckPageAccess(roles, "/api/chat")
	if err != nil || !allowed {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Bu işlem için yetkiniz bulunmamaktadır. Yapay zeka asistanı yöneticiniz tarafından bu rol için kapatılmıştır.",
		})
		return
	}

	// Kullanıcı adını e-postadan veya varsayılan olarak belirle
	userName := strings.Split(emailStr, "@")[0]

	// Sohbet servisini çağır
	response, err := h.ChatService.SendChatMessage(roleStr, userName, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Yapay zeka yanıtı üretilemedi",
			"details": err.Error(),
		})
		return
	}

	// Başarılı yanıtı döndür
	c.JSON(http.StatusOK, gin.H{
		"response": response,
	})
}
