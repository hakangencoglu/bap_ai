package api

import (
	"net/http"
	"strings"

	"bap_ai/backend/repository"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// ChatHandler yapısı, sohbet robotu HTTP handler'larını barındırır.
type ChatHandler struct {
	ChatService  *service.ChatService
	AdminService *service.AdminService
	ProjeRepo    *repository.ProjeRepository
	RAGService   *service.RAGService
}

// NewChatHandler fonksiyonu, yeni bir ChatHandler nesnesi döner.
func NewChatHandler(chatService *service.ChatService, adminService *service.AdminService, projeRepo *repository.ProjeRepository, ragService *service.RAGService) *ChatHandler {
	// Türkçe Yorum: ChatHandler nesnesi oluşturularak referansı döndürülür.
	return &ChatHandler{
		ChatService:  chatService,
		AdminService: adminService,
		ProjeRepo:    projeRepo,
		RAGService:   ragService,
	}
}

// SendMessage fonksiyonu, kullanıcının gönderdiği mesajı işler ve yapay zekadan gelen cevabı döner.
// POST /api/chat
func (h *ChatHandler) SendMessage(c *gin.Context) {
	// Türkçe Yorum: İstek gövdesi için veri yapısı (geçmiş dahil)
	type HistoryItem struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	var req struct {
		Message string        `json:"message" binding:"required"`
		History []HistoryItem `json:"history"`
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

	uyeIDFloat, exists := c.Get("uye_id")
	var uyeID int
	if exists {
		uyeID = int(uyeIDFloat.(float64))
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

	// Türkçe Yorum: RAG Servisi ile yetki ve mesaj bağlamına uygun filtrelenmiş veritabanı içeriği oluşturulur.
	var projectsContext string
	if h.RAGService != nil {
		ragCtx, errRAG := h.RAGService.BuildRAGContext(uyeID, roles, req.Message)
		if errRAG == nil && strings.TrimSpace(ragCtx) != "" {
			projectsContext = ragCtx
		}
	}

	// Kullanıcı adını e-postadan veya varsayılan olarak belirle
	userName := strings.Split(emailStr, "@")[0]

	// Geçmiş mesajları chat_service tipine dönüştür
	var chatHistory []service.ChatHistoryItem
	for _, h := range req.History {
		chatHistory = append(chatHistory, service.ChatHistoryItem{
			Role:    h.Role,
			Content: h.Content,
		})
	}

	// Sohbet servisini çağır
	response, err := h.ChatService.SendChatMessage(roleStr, userName, req.Message, projectsContext, chatHistory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Yapay zeka yanıtı üretilemedi",
			"details": err.Error(),
		})
		return
	}

	// Türkçe Yorum: Sohbet geçmişini kullanıcıya özel olarak veritabanına kaydet
	if h.RAGService != nil && h.RAGService.RAGRepo != nil && uyeID > 0 {
		_ = h.RAGService.RAGRepo.SaveChatMessage(uyeID, "user", req.Message)
		_ = h.RAGService.RAGRepo.SaveChatMessage(uyeID, "assistant", response)
	}

	// Başarılı yanıtı döndür
	c.JSON(http.StatusOK, gin.H{
		"response": response,
	})
}

// GetHistory kullanıcının veritabanındaki sohbet geçmişini döner
// GET /api/chat/history
func (h *ChatHandler) GetHistory(c *gin.Context) {
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı oturumu bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	if h.RAGService == nil || h.RAGService.RAGRepo == nil {
		c.JSON(http.StatusOK, gin.H{"history": []interface{}{}})
		return
	}

	history, err := h.RAGService.RAGRepo.GetUserChatHistory(uyeID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Sohbet geçmişi alınamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
	})
}

// ClearHistory kullanıcının veritabanındaki sohbet geçmişini temizler
// DELETE /api/chat/history
func (h *ChatHandler) ClearHistory(c *gin.Context) {
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı oturumu bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	if h.RAGService != nil && h.RAGService.RAGRepo != nil {
		_ = h.RAGService.RAGRepo.ClearUserChatHistory(uyeID)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sohbet geçmişi başarıyla temizlendi",
	})
}

// GetStatus fonksiyonu, yapay zeka sunucusunun durumunu döner
// GET /api/chat/status
func (h *ChatHandler) GetStatus(c *gin.Context) {
	// Türkçe Yorum: Servis üzerinden LLM bağlantı durumu kontrol edilir.
	connected := h.ChatService.CheckConnection()
	c.JSON(http.StatusOK, gin.H{
		"connected": connected,
	})
}

// SyncRAG veritabanını RAG indeksi ile senkronize eder
// POST /api/chat/sync-rag
func (h *ChatHandler) SyncRAG(c *gin.Context) {
	if h.RAGService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "RAG Servisi aktif değil"})
		return
	}

	count, err := h.RAGService.SyncDatabase()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "RAG senkronizasyonu başarısız",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "RAG indeks senkronizasyonu başarıyla tamamlandı",
		"indexed_count": count,
	})
}

