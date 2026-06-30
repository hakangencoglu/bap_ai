package api

import (
	"fmt"
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
}

// NewChatHandler fonksiyonu, yeni bir ChatHandler nesnesi döner.
func NewChatHandler(chatService *service.ChatService, adminService *service.AdminService, projeRepo *repository.ProjeRepository) *ChatHandler {
	// Türkçe Yorum: ChatHandler nesnesi oluşturularak referansı döndürülür.
	return &ChatHandler{
		ChatService:  chatService,
		AdminService: adminService,
		ProjeRepo:    projeRepo,
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

	// Projelerin veritabanı bağlamını (context) LLM için hazırla
	var projectsContext string
	if uyeID > 0 {
		var sb strings.Builder
		if strings.Contains(roleStr, "admin") || strings.Contains(roleStr, "dekan") || strings.Contains(roleStr, "komisyon") || strings.Contains(roleStr, "tto") {
			// Yönetim rolleri için son 50 projeyi yükle
			query := `
				SELECT p.proje_id, COALESCE(p.proje_kodu, ''), COALESCE(p.baslik_tr, 'Başlıksız'), COALESCE(pbt.bap_turu, 'Münferit'),
				       COALESCE(pd.durum_adi, 'taslak'), p.toplam_butce, p.sure_ay, COALESCE(u.ad || ' ' || u.soyad, 'Bilinmiyor')
				FROM proje p
				LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
				LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
				LEFT JOIN uye u ON p.koordinator_id = u.uye_id
				ORDER BY p.olusturma_tarihi DESC
				LIMIT 50
			`
			rows, err := h.ProjeRepo.DB.Query(query)
			if err == nil {
				defer rows.Close()
				sb.WriteString("SİSTEMDEKİ SON 50 BAP PROJESİ:\n")
				for rows.Next() {
					var pID, pSure int
					var pKod, pBaslik, pTuru, pDurum, pKoord string
					var pButce float64
					if err := rows.Scan(&pID, &pKod, &pBaslik, &pTuru, &pDurum, &pButce, &pSure, &pKoord); err == nil {
						sb.WriteString(fmt.Sprintf("- Kod: %s (ID: %d), Başlık: %s, Tür: %s, Aşama/Durum: %s, Toplam Bütçe: %.2f TL, Süre: %d Ay, Koordinatör: %s\n",
							pKod, pID, pBaslik, pTuru, pDurum, pButce, pSure, pKoord))
					}
				}
			}
		} else if strings.Contains(roleStr, "hakem") {
			// Hakem için atanmış projeleri yükle
			query := `
				SELECT p.proje_id, COALESCE(p.proje_kodu, ''), COALESCE(p.baslik_tr, 'Başlıksız'), COALESCE(pbt.bap_turu, 'Münferit'),
				       COALESCE(pd.durum_adi, 'taslak'), pdeg.durum, COALESCE(pdeg.puan, 0), p.toplam_butce, p.sure_ay
				FROM proje p
				INNER JOIN proje_degerlendirmeleri pdeg ON p.proje_id = pdeg.proje_id
				LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
				LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
				WHERE pdeg.hakem_id = $1
				ORDER BY pdeg.olusturma_tarihi DESC
			`
			rows, err := h.ProjeRepo.DB.Query(query, uyeID)
			if err == nil {
				defer rows.Close()
				sb.WriteString("SİZE ATANAN DEĞERLENDİRME PROJELERİ:\n")
				for rows.Next() {
					var pID, pPuan, pSure int
					var pKod, pBaslik, pTuru, pDurum, pHakemDurum string
					var pButce float64
					if err := rows.Scan(&pID, &pKod, &pBaslik, &pTuru, &pDurum, &pHakemDurum, &pPuan, &pButce, &pSure); err == nil {
						sb.WriteString(fmt.Sprintf("- Kod: %s (ID: %d), Başlık: %s, Tür: %s, Proje Durumu: %s, Değerlendirme Durumunuz: %s, Verdiğiniz Puan: %d, Bütçe: %.2f TL, Süre: %d Ay\n",
							pKod, pID, pBaslik, pTuru, pDurum, pHakemDurum, pPuan, pButce, pSure))
					}
				}
			}
		} else {
			// Akademisyen/Öğrenci için kendi projelerini yükle
			query := `
				SELECT p.proje_id, COALESCE(p.proje_kodu, ''), COALESCE(p.baslik_tr, 'Başlıksız'), COALESCE(pbt.bap_turu, 'Münferit'),
				       COALESCE(pd.durum_adi, 'taslak'), p.toplam_butce, p.sure_ay
				FROM proje p
				INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id
				LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
				LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
				WHERE pt.uye_id = $1 AND pt.davet_durumu = 'kabul'
				ORDER BY p.olusturma_tarihi DESC
			`
			rows, err := h.ProjeRepo.DB.Query(query, uyeID)
			if err == nil {
				defer rows.Close()
				sb.WriteString("PROJELERİNİZ:\n")
				for rows.Next() {
					var pID, pSure int
					var pKod, pBaslik, pTuru, pDurum string
					var pButce float64
					if err := rows.Scan(&pID, &pKod, &pBaslik, &pTuru, &pDurum, &pButce, &pSure); err == nil {
						sb.WriteString(fmt.Sprintf("- Kod: %s (ID: %d), Başlık: %s, Tür: %s, Aşama/Durum: %s, Toplam Bütçe: %.2f TL, Süre: %d Ay\n",
							pKod, pID, pBaslik, pTuru, pDurum, pButce, pSure))
					}
				}
			}
		}
		projectsContext = sb.String()
	}

	// Kullanıcı adını e-postadan veya varsayılan olarak belirle
	userName := strings.Split(emailStr, "@")[0]

	// Sohbet servisini çağır
	response, err := h.ChatService.SendChatMessage(roleStr, userName, req.Message, projectsContext)
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

// GetStatus fonksiyonu, yapay zeka sunucusunun durumunu döner
// GET /api/chat/status
func (h *ChatHandler) GetStatus(c *gin.Context) {
	// Türkçe Yorum: Servis üzerinden LLM bağlantı durumu kontrol edilir.
	connected := h.ChatService.CheckConnection()
	c.JSON(http.StatusOK, gin.H{
		"connected": connected,
	})
}

