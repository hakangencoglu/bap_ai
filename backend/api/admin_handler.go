package api

import (
	"encoding/csv"
	"io"
	"net/http"
	"strconv"
	"strings"

	"bap_ai/backend/models"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// AdminHandler, admin isteklerini işler.
type AdminHandler struct {
	adminService *service.AdminService
}

// NewAdminHandler, yeni bir AdminHandler örneği oluşturur.
func NewAdminHandler(s *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: s}
}

// GetStats, admin dashboard istatistiklerini döner.
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.adminService.GetAdminDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İstatistikler alınamadı"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetAllUsers, tüm kullanıcıları döner.
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	users, err := h.adminService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcılar alınamadı"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// UpdateUserRole, bir kullanıcının rolünü günceller
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	var req struct {
		UyeID  int    `json:"uye_id" binding:"required"`
		RolAdi string `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek parametreleri"})
		return
	}

	err := h.adminService.UpdateUserRole(req.UyeID, req.RolAdi)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Rol güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rol başarıyla güncellendi"})
}

// UpdateUserStatus, bir kullanıcının aktifliğini günceller
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	var req struct {
		UyeID    int  `json:"uye_id" binding:"required"`
		AktifMi  bool `json:"aktif_mi"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek parametreleri"})
		return
	}

	err := h.adminService.UpdateUserStatus(req.UyeID, req.AktifMi)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcı durumu güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kullanıcı durumu güncellendi"})
}

// UpdateProjectStatus, projenin durumunu günceller.
func (h *AdminHandler) UpdateProjectStatus(c *gin.Context) {
	var req struct {
		ProjeID int    `json:"proje_id" binding:"required"`
		Durum   string `json:"durum" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek parametreleri"})
		return
	}

	err := h.adminService.UpdateProjectStatus(req.ProjeID, req.Durum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proje durumu güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proje durumu başarıyla güncellendi"})
}

// GetAllProjects, sistemdeki projeleri döner (Eğer servis bağlanırsa).
func (h *AdminHandler) GetAllProjects(c *gin.Context) {
	// Projeleri çekmek için
	projes, err := h.adminService.GetAllProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Projeler alınamadı"})
		return
	}
	c.JSON(http.StatusOK, projes)
}

// GetProjectDetails, admin için istenen projenin spesifik detaylarını (hakem, bütçe) döner.
func (h *AdminHandler) GetProjectDetails(c *gin.Context) {
	idStr := c.Param("id")
	projeID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	details, err := h.adminService.GetProjectDetailsForAdmin(projeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proje detayları getirilemedi"})
		return
	}

	c.JSON(http.StatusOK, details)
}

// AssignHakem, admin tarafından projeye hakem ataması yapar
func (h *AdminHandler) AssignHakem(c *gin.Context) {
	var req models.AdminHakemAtamaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek parametreleri"})
		return
	}

	err := h.adminService.AssignHakem(req.ProjeID, req.HakemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hakem ataması yapılamadı: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Hakem başarıyla atandı"})
}

// GetDegerlendirilmemisProjeleri, hakem atanması gereken projeleri döner
func (h *AdminHandler) GetDegerlendirilmemisProjeleri(c *gin.Context) {
	projeler, err := h.adminService.GetDegerlendirilmemisProjeleri()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Projeler alınamadı"})
		return
	}
	c.JSON(http.StatusOK, projeler)
}

// GetHakemListesi, sistemdeki aktif hakemleri döner
func (h *AdminHandler) GetHakemListesi(c *gin.Context) {
	hakemler, err := h.adminService.GetHakemListesi()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hakem listesi alınamadı"})
		return
	}
	c.JSON(http.StatusOK, hakemler)
}

// GetProjeyeAtananHakemler, belirtilen projeye atanan hakemleri döner
func (h *AdminHandler) GetProjeyeAtananHakemler(c *gin.Context) {
	idStr := c.Param("id")
	projeID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	hakemler, err := h.adminService.GetProjeyeAtananHakemler(projeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hakem bilgileri getirilemedi"})
		return
	}
	c.JSON(http.StatusOK, hakemler)
}

// GetBapTurleri, sistemdeki tüm BAP proje türlerini döner (Admin için)
// GET /api/admin/bap-turleri
func (h *AdminHandler) GetBapTurleri(c *gin.Context) {
	list, err := h.adminService.GetBapTurleri(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "BAP türleri listesi alınamadı"})
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetBapTurleriPublic, sadece aktif olan BAP proje türlerini döner (Kullanıcılar için)
// GET /api/bap-turleri
func (h *AdminHandler) GetBapTurleriPublic(c *gin.Context) {
	list, err := h.adminService.GetBapTurleri(true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "BAP türleri listesi alınamadı"})
		return
	}
	c.JSON(http.StatusOK, list)
}

// CreateBapTuru, yeni bir BAP proje türü oluşturur (Admin için)
// POST /api/admin/bap-turu
func (h *AdminHandler) CreateBapTuru(c *gin.Context) {
	var req models.ProjeBapTuru
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri formatı"})
		return
	}

	if err := h.adminService.CreateBapTuru(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "BAP türü oluşturulamadı"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "BAP türü başarıyla oluşturuldu", "data": req})
}

// UpdateBapTuru, mevcut bir BAP proje türünü günceller (Admin için)
// PUT /api/admin/bap-turu/:id
func (h *AdminHandler) UpdateBapTuru(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz BAP türü ID"})
		return
	}

	var req models.ProjeBapTuru
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri formatı"})
		return
	}

	req.BapTuruID = id
	if err := h.adminService.UpdateBapTuru(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "BAP türü güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "BAP türü başarıyla güncellendi"})
}

// CreateUser, admin tarafından yeni bir kullanıcı eklenmesini sağlar
// POST /api/admin/user
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req models.AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri formatı", "detay": err.Error()})
		return
	}

	if err := h.adminService.CreateUser(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcı oluşturulamadı: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Kullanıcı başarıyla oluşturuldu"})
}

// UpdateUser, admin tarafından bir kullanıcının bilgilerini günceller.
// PUT /api/admin/user/:id
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	uyeID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz kullanıcı ID"})
		return
	}

	var req models.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri formatı", "detay": err.Error()})
		return
	}

	if err := h.adminService.UpdateUser(uyeID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcı bilgileri güncellenemedi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kullanıcı bilgileri başarıyla güncellendi"})
}

// BulkCreateUsers admin tarafından CSV dosyası ile toplu kullanıcı eklenmesini sağlar
// POST /api/admin/users/bulk
// Türkçe Yorum: Admin'in yüklediği CSV dosyasını okuyup, kolon adlarına göre eşleştirerek toplu üye kaydı yapar.
func (h *AdminHandler) BulkCreateUsers(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dosya yüklenemedi", "detay": err.Error()})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Dosya açılamadı", "detay": err.Error()})
		return
	}
	defer src.Close()

	reader := csv.NewReader(src)
	// Virgül veya noktalı virgül desteği ekleyelim (özellikle Türkçe Excel CSV'lerinde noktalı virgül kullanılır)
	// Türkçe Yorum: Noktalı virgül veya virgül ayracını desteklemek için kontrol ekliyoruz
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	// İlk satırı (başlıklar) oku
	headers, err := reader.Read()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV başlıkları okunamadı", "detay": err.Error()})
		return
	}

	// Eğer noktalı virgülle ayrılmışsa başlık tek bir elemandır ve ";" içerir, o zaman separator değiştirelim
	if len(headers) == 1 && strings.Contains(headers[0], ";") {
		// Dosyayı başa alıp separatorü değiştirip yeniden okuyalım
		if _, seekErr := src.Seek(0, 0); seekErr == nil {
			reader = csv.NewReader(src)
			reader.Comma = ';'
			reader.LazyQuotes = true
			reader.FieldsPerRecord = -1
			headers, err = reader.Read()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "CSV başlıkları okunamadı", "detay": err.Error()})
				return
			}
		}
	}

	// Başlıkları normalize et
	var normalizedHeaders []string
	for _, h := range headers {
		normalizedHeaders = append(normalizedHeaders, normalizeHeader(h))
	}

	var requests []models.AdminCreateUserRequest

	// Diğer satırları oku
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "CSV satırı okunamadı", "detay": err.Error()})
			return
		}

		// Boş satırları atla
		allEmpty := true
		for _, val := range record {
			if strings.TrimSpace(val) != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			continue
		}

		req := models.AdminCreateUserRequest{
			Rol: "akademisyen", // Varsayılan rol
		}

		// Kolon değerlerini ata
		for i, val := range record {
			if i >= len(normalizedHeaders) {
				continue
			}
			val = strings.TrimSpace(val)
			colName := normalizedHeaders[i]

			switch colName {
			case "ad":
				req.Ad = val
			case "soyad":
				req.Soyad = val
			case "eposta":
				req.Eposta = val
			case "sifre":
				req.Sifre = val
			case "rol":
				req.Rol = val
			case "unvan":
				req.Unvan = val
			case "bolum":
				req.Bolum = val
			case "telefon":
				req.Telefon = val
			case "izu_uyesi":
				req.IzuUyesi = strings.ToLower(val) == "true" || val == "1" || strings.ToLower(val) == "yes" || strings.ToLower(val) == "evet"
			case "sifre_degistir_zorla":
				req.SifreDegistirZorla = strings.ToLower(val) == "true" || val == "1" || strings.ToLower(val) == "yes" || strings.ToLower(val) == "evet"
			}
		}

		requests = append(requests, req)
	}

	// Servis katmanına gönder
	success, failed, details := h.adminService.BulkCreateUsers(requests)

	c.JSON(http.StatusOK, gin.H{
		"message":         "Toplu kullanıcı ekleme işlemi tamamlandı",
		"basarili_sayisi": success,
		"hatali_sayisi":   failed,
		"hatalar":         details,
	})
}

// normalizeHeader, CSV başlıklarını normalize ederek standart isimlere dönüştürür.
// Türkçe Yorum: Kolon başlıklarını Türkçe karakterlerden ve büyük/küçük harf farklılıklarından arındırır.
func normalizeHeader(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	h = strings.ReplaceAll(h, "ı", "i")
	h = strings.ReplaceAll(h, "ş", "s")
	h = strings.ReplaceAll(h, "ö", "o")
	h = strings.ReplaceAll(h, "ü", "u")
	h = strings.ReplaceAll(h, "ğ", "g")
	h = strings.ReplaceAll(h, "ç", "c")

	switch h {
	case "ad", "firstname", "first name", "name", "first_name":
		return "ad"
	case "soyad", "lastname", "last name", "surname", "last_name":
		return "soyad"
	case "eposta", "email", "e-mail", "username":
		return "eposta"
	case "sifre", "password":
		return "sifre"
	case "rol", "role":
		return "rol"
	case "unvan", "title", "academic_title":
		return "unvan"
	case "bolum", "department", "dept":
		return "bolum"
	case "telefon", "phone", "telephone", "phone_number":
		return "telefon"
	case "izu_uyesi", "izu_member", "izu":
		return "izu_uyesi"
	case "sifre_degistir_zorla", "force_password_change", "forcepasswordchange", "force_password":
		return "sifre_degistir_zorla"
	}
	return h
}
