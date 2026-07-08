package api

import (
	"net/http"
	"strconv"
	"strings"

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
	// Türkçe Yorum: Çoklu rolleri virgülle ayrılmış şekilde alıp split ederek öğrenci rolü kontrolü yapıyoruz.
	role, _ := c.Get("role")
	roleStr, _ := role.(string)

	isOgrenci := false
	roles := strings.Split(roleStr, ",")
	for _, r := range roles {
		if strings.TrimSpace(r) == "ogrenci" {
			isOgrenci = true
			break
		}
	}

	// Öğrenci sadece belirli BAP türlerine başvurabilir
	if isOgrenci {
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

	// Durumu "taslak" olarak ayarla (durum_id=1)
	durumID := 1
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

// DeleteTaslakProje taslak durumundaki bir projeyi siler.
// DELETE /api/proje/:id
func (h *ProjeHandler) DeleteTaslakProje(c *gin.Context) {
	// URL'den proje ID'sini al
	projeIDStr := c.Param("id")
	projeID, err := strconv.Atoi(projeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	// Middleware'den giriş yapan üye ID'sini al
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	// Taslak projeyi sil (yetki ve durum kontrolü repository'de yapılır)
	err = h.ProjeService.DeleteTaslakProje(projeID, uyeID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Taslak proje başarıyla silindi"})
}

// GetWorkflowProjects onay bekleyen veya süreçteki projeleri yetkili role göre listeler.
// GET /api/workflow/projects
func (h *ProjeHandler) GetWorkflowProjects(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Rol bilgisi bulunamadı"})
		return
	}
	roleStr := role.(string)

	uyeIDFloat, existsUye := c.Get("uye_id")
	uyeID := 0
	if existsUye {
		uyeID = int(uyeIDFloat.(float64))
	}

	projeler, err := h.ProjeService.GetProjectsForWorkflow(roleStr, uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"projects": projeler})
}

// HandleWorkflowAction onay/red/revizyon kararlarını işleyen endpointtir.
// POST /api/workflow/action
func (h *ProjeHandler) HandleWorkflowAction(c *gin.Context) {
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	islemYapanID := int(uyeIDFloat.(float64))

	var req struct {
		ProjeID  int    `json:"proje_id" binding:"required"`
		Action   string `json:"action" binding:"required"` // onayla, reddet, revizyon
		Aciklama string `json:"aciklama"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek parametreleri"})
		return
	}

	err := h.ProjeService.ProcessWorkflowAction(req.ProjeID, islemYapanID, req.Action, req.Aciklama)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "İşlem başarıyla gerçekleştirildi"})
}

// GetSurecGecmisi projenin geçmiş tüm onay ve revizyon durum değişikliklerini tarihçesiyle döner.
// GET /api/proje/:id/surec-gecmisi
func (h *ProjeHandler) GetSurecGecmisi(c *gin.Context) {
	projeIDStr := c.Param("id")
	projeID, err := strconv.Atoi(projeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	gecmis, err := h.ProjeService.GetProjeSurecGecmisi(projeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Süreç geçmişi getirilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"gecmis": gecmis})
}

// GetWorkflowHistory kullanıcının kendi geçmiş workflow (onay/red/revizyon) işlemlerini döner.
// GET /api/workflow/history
func (h *ProjeHandler) GetWorkflowHistory(c *gin.Context) {
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	uyeID := int(uyeIDFloat.(float64))

	gecmis, err := h.ProjeService.GetWorkflowHistoryByUyeID(uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Geçmiş kararlar getirilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"history": gecmis})
}

// ProjectExtrasRequest, proje ek verilerini tek istekte toplamak için kullanılan yapıdır.
type ProjectExtrasRequest struct {
	ProjeDetay                 *models.ProjeDetay                       `json:"proje_detay"`
	IsPaketleri                []repository.IsPaketiInput               `json:"is_paketleri"`
	ButceKalemleri             []repository.ButceKalemiInput            `json:"butce_kalemleri"`
	RiskYonetimi               []repository.RiskInput                   `json:"risk_yonetimi"`
	ArastirmaAmaci             string                                   `json:"arastirma_amaci"`
	YayinEtki                  []repository.YayinEtkiInput              `json:"yayin_etki"`
	YayginlastirmaEtkinlikleri []repository.YayginlastirmaEtkinlikInput `json:"yayginlastirma_etkinlikleri"`
}

// SaveProjectExtras projeye ait ek verileri (iş paketleri, bütçe, detaylar) tek istekte kaydeder.
// POST /api/proje/:id/extras
func (h *ProjeHandler) SaveProjectExtras(c *gin.Context) {
	// Proje ID'sini URL parametresinden al
	projeIDStr := c.Param("id")
	projeID, err := strconv.Atoi(projeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	// İstek gövdesini parse et
	var req ProjectExtrasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek formatı"})
		return
	}

	// 1. Proje detaylarını kaydet (özet, anahtar kelimeler, hedefler vb.)
	if req.ProjeDetay != nil {
		req.ProjeDetay.ProjeID = projeID
		if err := h.ProjeService.ProjeRepo.SaveProjeDetay(req.ProjeDetay); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Proje detayları kaydedilemedi"})
			return
		}
	}

	// 2. İş paketlerini kaydet
	if req.IsPaketleri != nil {
		if err := h.ProjeService.ProjeRepo.SaveIsPaketleri(projeID, req.IsPaketleri); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "İş paketleri kaydedilemedi"})
			return
		}
	}

	// 3. Bütçe kalemlerini kaydet
	if req.ButceKalemleri != nil {
		if err := h.ProjeService.ProjeRepo.SaveButceKalemleri(projeID, req.ButceKalemleri); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Bütçe kalemleri kaydedilemedi"})
			return
		}
	}

	// 4. Risk yönetimi kayıtlarını kaydet
	if req.RiskYonetimi != nil {
		if err := h.ProjeService.ProjeRepo.SaveRiskYonetimi(projeID, req.RiskYonetimi); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Risk yönetimi kaydedilemedi"})
			return
		}
	}

	// 5. Araştırma olanakları bilgisini kaydet
	if req.ArastirmaAmaci != "" {
		uyeIDFloat, _ := c.Get("uye_id")
		uyeID := int(uyeIDFloat.(float64))
		if err := h.ProjeService.ProjeRepo.SaveArastirma(projeID, uyeID, req.ArastirmaAmaci); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Araştırma olanakları kaydedilemedi"})
			return
		}
	}

	// 6. Yaygın etki çıktılarını kaydet
	if req.YayinEtki != nil {
		if err := h.ProjeService.ProjeRepo.SaveYayinEtki(projeID, req.YayinEtki); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Yaygın etki verileri kaydedilemedi"})
			return
		}
	}

	// 7. Yaygınlaştırma etkinliklerini kaydet
	if req.YayginlastirmaEtkinlikleri != nil {
		if err := h.ProjeService.ProjeRepo.SaveYayginlastirmaEtkinlik(projeID, req.YayginlastirmaEtkinlikleri); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Yaygınlaştırma etkinlikleri kaydedilemedi"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proje ek verileri başarıyla kaydedildi"})
}

// GetProjeDetaylar, projenin bütçe, iş paketleri, risk ve araştırma bilgilerini döner.
func (h *ProjeHandler) GetProjeDetaylar(c *gin.Context) {
	projeIDStr := c.Param("id")
	projeID, err := strconv.Atoi(projeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	butceler, isPaketleri, riskler, arastirma, err := h.ProjeService.ProjeRepo.GetProjeDetaylar(projeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proje detayları alınamadı"})
		return
	}

	yayinEtki, yayginlastirmaEtkinlikleri, _ := h.ProjeService.ProjeRepo.GetProjeYayinEtkiBilgiler(projeID)

	c.JSON(http.StatusOK, gin.H{
		"butceler":                    butceler,
		"is_paketleri":                isPaketleri,
		"risk_yonetimi":               riskler,
		"arastirma":                   arastirma,
		"yayin_etki":                  yayinEtki,
		"yayginlastirma_etkinlikleri": yayginlastirmaEtkinlikleri,
	})
}

// GetDBStatus geçici olarak veritabanındaki komisyon ve oylama durumlarını sorgular.
func (h *ProjeHandler) GetDBStatus(c *gin.Context) {
	db := h.ProjeService.ProjeRepo.DB

	type KomisyonUye struct {
		ID        int    `json:"id"`
		Ad        string `json:"ad"`
		Soyad     string `json:"soyad"`
		Rol       string `json:"rol"`
		Aktif     bool   `json:"aktif"`
		SistemRol string `json:"sistem_rol"`
	}
	rows, err := db.Query(`
		SELECT u.uye_id, u.ad, u.soyad, u.rol, u.aktif_mi, COALESCE(srt.rol_adi, 'yok')
		FROM uye u
		LEFT JOIN sistem_rol sr ON u.uye_id = sr.uye_id
		LEFT JOIN sistem_rol_tanimlama srt ON sr.sistem_rol_id = srt.rol_id
		WHERE u.rol = 'komisyon' OR srt.rol_adi = 'komisyon'
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var uyeler []KomisyonUye
	for rows.Next() {
		var u KomisyonUye
		rows.Scan(&u.ID, &u.Ad, &u.Soyad, &u.Rol, &u.Aktif, &u.SistemRol)
		uyeler = append(uyeler, u)
	}

	type OylamaKayit struct {
		UyeID    int    `json:"uye_id"`
		Ad       string `json:"ad"`
		Soyad    string `json:"soyad"`
		Karar    string `json:"karar"`
		Aciklama string `json:"aciklama"`
	}
	rows2, err := db.Query(`
		SELECT pko.komisyon_uye_id, u.ad, u.soyad, pko.karar, COALESCE(pko.aciklama, '')
		FROM  proje_komisyon_onay pko
		JOIN  uye u ON pko.komisyon_uye_id = u.uye_id
		WHERE pko.proje_id = 1
	`)
	var oylamalar []OylamaKayit
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var o OylamaKayit
			rows2.Scan(&o.UyeID, &o.Ad, &o.Soyad, &o.Karar, &o.Aciklama)
			oylamalar = append(oylamalar, o)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"komisyon_uyeleri": uyeler,
		"oylamalar_proje_1": oylamalar,
	})
}

// GetButceKategorileri veritabanındaki bütçe kategorilerini döner.
// GET /api/butce-kategorileri
// Türkçe Bilgilendirme: Sistemdeki tüm bütçe kategorilerini (Makine-Teçhizat, Sarf vb.) JSON dizisi olarak döner.
func (h *ProjeHandler) GetButceKategorileri(c *gin.Context) {
	list, err := h.ProjeService.GetButceKategorileri()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bütçe kategorileri alınamadı"})
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetSistemRolleri veritabanındaki sistem rollerini ve Türkçe etiketlerini döner.
// GET /api/sistem-rolleri
// Türkçe Bilgilendirme: Sistemdeki tüm rolleri ve Türkçe karşılıklarını JSON dizisi olarak döner.
func (h *ProjeHandler) GetSistemRolleri(c *gin.Context) {
	list, err := h.ProjeService.GetSistemRolleri()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Sistem rolleri listesi alınamadı"})
		return
	}
	c.JSON(http.StatusOK, list)
}


