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

	// Projelerin veritabanı bağlamını (context) LLM için hazırla
	var projectsContext string
	if uyeID > 0 {
		var sb strings.Builder

		// Yardımcı fonksiyon: Projenin risk yönetimi, bütçe ve iş paketlerini getirir
		fetchProjectDetails := func(projeID int) string {
			var detailSb strings.Builder

			// 1. Riskleri getir
			riskQuery := `
				SELECT COALESCE(risk_aciklamasi, ''), COALESCE(cozum_plani, '') 
				FROM risk_yonetimi 
				WHERE proje_id = $1
			`
			riskRows, err := h.ProjeRepo.DB.Query(riskQuery, projeID)
			if err == nil {
				hasRisk := false
				for riskRows.Next() {
					var risk, cozum string
					if err := riskRows.Scan(&risk, &cozum); err == nil {
						if !hasRisk {
							detailSb.WriteString("    * Risk Yönetimi:\n")
							hasRisk = true
						}
						detailSb.WriteString(fmt.Sprintf("      - Risk: %s | Çözüm Planı: %s\n", risk, cozum))
					}
				}
				riskRows.Close()
			}

			// 2. Bütçe kalemlerini getir
			butceQuery := `
				SELECT COALESCE(aciklama, ''), COALESCE(toplam_tutar, 0.0) 
				FROM butce 
				WHERE proje_id = $1
			`
			butceRows, err := h.ProjeRepo.DB.Query(butceQuery, projeID)
			if err == nil {
				hasButce := false
				for butceRows.Next() {
					var aciklama string
					var tutar float64
					if err := butceRows.Scan(&aciklama, &tutar); err == nil {
						if !hasButce {
							detailSb.WriteString("    * Bütçe Kalemleri:\n")
							hasButce = true
						}
						detailSb.WriteString(fmt.Sprintf("      - Açıklama: %s | Tutar: %.2f TL\n", aciklama, tutar))
					}
				}
				butceRows.Close()
			}

			// 3. İş paketlerini getir
			isPaketiQuery := `
				SELECT COALESCE(paket_adi, ''), COALESCE(paket_amaci, ''), baslangic_ay, bitis_ay 
				FROM is_paketi 
				WHERE proje_id = $1
				ORDER BY baslangic_ay ASC
			`
			isRows, err := h.ProjeRepo.DB.Query(isPaketiQuery, projeID)
			if err == nil {
				hasIs := false
				for isRows.Next() {
					var ad, amac string
					var baslangic, bitis int
					if err := isRows.Scan(&ad, &amac, &baslangic, &bitis); err == nil {
						if !hasIs {
							detailSb.WriteString("    * İş Paketleri:\n")
							hasIs = true
						}
						detailSb.WriteString(fmt.Sprintf("      - Paket Adı: %s | Amacı: %s | Süre: Ay %d - Ay %d (%d Ay)\n", 
							ad, amac, baslangic, bitis, bitis-baslangic+1))
					}
				}
				isRows.Close()
			}

			// 4. Satın alma taleplerini getir ve harcanan/kalan bütçeyi hesapla
			saQuery := `
				SELECT COALESCE(malzeme_adi, ''), miktar, birim_fiyat, toplam_fiyat, COALESCE(durum, 'Beklemede')
				FROM satinalma_talebi
				WHERE proje_id = $1
			`
			saRows, err := h.ProjeRepo.DB.Query(saQuery, projeID)
			if err == nil {
				hasSa := false
				var toplamHarcanan float64
				var saDetails strings.Builder
				
				for saRows.Next() {
					var malzeme string
					var miktar int
					var birimFiyat, toplamFiyat float64
					var durum string
					if err := saRows.Scan(&malzeme, &miktar, &birimFiyat, &toplamFiyat, &durum); err == nil {
						if !hasSa {
							saDetails.WriteString("    * Satın Alma Talepleri:\n")
							hasSa = true
						}
						saDetails.WriteString(fmt.Sprintf("      - Malzeme: %s | Miktar: %d | Birim Fiyat: %.2f TL | Toplam: %.2f TL | Durum: %s\n", malzeme, miktar, birimFiyat, toplamFiyat, durum))
						// Sadece Onaylanan satın alma talepleri bütçeden düşülür (harcanır)
						if durum == "Onaylandı" {
							toplamHarcanan += toplamFiyat
						}
					}
				}
				saRows.Close()

				// Projenin toplam bütçesini bulup kalan bütçeyi hesaplayalım ve context'e yazalım
				var toplamButce float64
				errButce := h.ProjeRepo.DB.QueryRow("SELECT toplam_butce FROM proje WHERE proje_id = $1", projeID).Scan(&toplamButce)
				if errButce == nil {
					kalanButce := toplamButce - toplamHarcanan
					detailSb.WriteString(fmt.Sprintf("    * Bütçe Durumu: Toplam Bütçe: %.2f TL | Harcanan Bütçe (Onaylı Talepler): %.2f TL | Kalan Bütçe: %.2f TL\n", toplamButce, toplamHarcanan, kalanButce))
				}
				
				if hasSa {
					detailSb.WriteString(saDetails.String())
				} else {
					detailSb.WriteString("    * Satın Alma Talepleri: Henüz satın alma talebi oluşturulmamıştır.\n")
				}
			}

			return detailSb.String()
		}

		if strings.Contains(roleStr, "admin") || strings.Contains(roleStr, "dekan") || strings.Contains(roleStr, "komisyon") || strings.Contains(roleStr, "tto") {
			// Yönetim rolleri için son 25 projeyi yükle
			query := `
				SELECT p.proje_id, COALESCE(p.proje_kodu, ''), COALESCE(p.baslik_tr, 'Başlıksız'), COALESCE(pbt.bap_turu, 'Münferit'),
				       COALESCE(pd.durum_adi, 'taslak'), p.toplam_butce, p.sure_ay, COALESCE(u.ad || ' ' || u.soyad, 'Bilinmiyor')
				FROM proje p
				LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
				LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
				LEFT JOIN uye u ON p.koordinator_id = u.uye_id
				ORDER BY p.olusturma_tarihi DESC
				LIMIT 25
			`
			rows, err := h.ProjeRepo.DB.Query(query)
			if err == nil {
				defer rows.Close()
				sb.WriteString("SİSTEMDEKİ SON 25 BAP PROJESİ VE DETAYLARI:\n")
				type ProjeTemp struct {
					pID, pSure int
					pKod, pBaslik, pTuru, pDurum, pKoord string
					pButce float64
				}
				var tempProjects []ProjeTemp
				for rows.Next() {
					var p ProjeTemp
					if err := rows.Scan(&p.pID, &p.pKod, &p.pBaslik, &p.pTuru, &p.pDurum, &p.pButce, &p.pSure, &p.pKoord); err == nil {
						tempProjects = append(tempProjects, p)
					}
				}
				for _, p := range tempProjects {
					sb.WriteString(fmt.Sprintf("- Kod: %s (ID: %d), Başlık: %s, Tür: %s, Aşama/Durum: %s, Toplam Bütçe: %.2f TL, Süre: %d Ay, Koordinatör: %s\n",
						p.pKod, p.pID, p.pBaslik, p.pTuru, p.pDurum, p.pButce, p.pSure, p.pKoord))
					sb.WriteString(fetchProjectDetails(p.pID))
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
				sb.WriteString("SİZE ATANAN DEĞERLENDİRME PROJELERİ VE DETAYLARI:\n")
				type ProjeTemp struct {
					pID, pPuan, pSure int
					pKod, pBaslik, pTuru, pDurum, pHakemDurum string
					pButce float64
				}
				var tempProjects []ProjeTemp
				for rows.Next() {
					var p ProjeTemp
					if err := rows.Scan(&p.pID, &p.pKod, &p.pBaslik, &p.pTuru, &p.pDurum, &p.pHakemDurum, &p.pPuan, &p.pButce, &p.pSure); err == nil {
						tempProjects = append(tempProjects, p)
					}
				}
				for _, p := range tempProjects {
					sb.WriteString(fmt.Sprintf("- Kod: %s (ID: %d), Başlık: %s, Tür: %s, Proje Durumu: %s, Değerlendirme Durumunuz: %s, Verdiğiniz Puan: %d, Bütçe: %.2f TL, Süre: %d Ay\n",
						p.pKod, p.pID, p.pBaslik, p.pTuru, p.pDurum, p.pHakemDurum, p.pPuan, p.pButce, p.pSure))
					sb.WriteString(fetchProjectDetails(p.pID))
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
				sb.WriteString("PROJELERİNİZ VE DETAYLARI:\n")
				type ProjeTemp struct {
					pID, pSure int
					pKod, pBaslik, pTuru, pDurum string
					pButce float64
				}
				var tempProjects []ProjeTemp
				for rows.Next() {
					var p ProjeTemp
					if err := rows.Scan(&p.pID, &p.pKod, &p.pBaslik, &p.pTuru, &p.pDurum, &p.pButce, &p.pSure); err == nil {
						tempProjects = append(tempProjects, p)
					}
				}
				for _, p := range tempProjects {
					sb.WriteString(fmt.Sprintf("- Kod: %s (ID: %d), Başlık: %s, Tür: %s, Aşama/Durum: %s, Toplam Bütçe: %.2f TL, Süre: %d Ay\n",
						p.pKod, p.pID, p.pBaslik, p.pTuru, p.pDurum, p.pButce, p.pSure))
					sb.WriteString(fetchProjectDetails(p.pID))
				}
			}
		}
		projectsContext = sb.String()
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

