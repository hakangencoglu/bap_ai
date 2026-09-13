package api

import (
	"net/http"
	"strconv"
	"strings"

	"bap_ai/backend/models"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// TalepHandler, proje talep HTTP isteklerini işler.
// Türkçe Yorum: Akademisyen talep gönderme, Admin/TTO listeleme ve onay/red isteklerini karşılar.
type TalepHandler struct {
	Service *service.TalepService
}

// NewTalepHandler yeni bir TalepHandler döner.
func NewTalepHandler(srv *service.TalepService) *TalepHandler {
	return &TalepHandler{Service: srv}
}

// uyeIDFromContext, JWT token'dan uye_id çeker.
func uyeIDFromContext(c *gin.Context) (int, bool) {
	v, ok := c.Get("uye_id")
	if !ok {
		return 0, false
	}
	return int(v.(float64)), true
}

// SubmitTalep, talep tipine göre doğru handler'ı çağırır.
// POST /api/talep/:tip
// Türkçe Yorum: URL parametresindeki talep tipi string'e göre switch ile ilgili servis metoduna yönlendirir.
func (h *TalepHandler) SubmitTalep(c *gin.Context) {
	uyeID, ok := uyeIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	tip := c.Param("tip")

	switch tip {
	case "ek_sure":
		var t models.TalepEkSure
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek"})
			return
		}
		t.UyeID = uyeID
		if err := h.Service.SubmitEkSure(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Ek süre talebi oluşturuldu", "talep_no": t.TalepNo})

	case "ek_butce":
		var t models.TalepEkButce
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek"})
			return
		}
		t.UyeID = uyeID
		if err := h.Service.SubmitEkButce(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Ek bütçe talebi oluşturuldu", "talep_no": t.TalepNo})

	case "fasil_aktarimi":
		var t models.TalepFasilAktarimi
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek"})
			return
		}
		t.UyeID = uyeID
		if err := h.Service.SubmitFasilAktarimi(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Fasıl aktarımı talebi oluşturuldu", "talep_no": t.TalepNo})

	case "arastirmaci":
		var t models.TalepArastirmaci
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek"})
			return
		}
		t.UyeID = uyeID
		if err := h.Service.SubmitArastirmaci(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Araştırmacı değişikliği talebi oluşturuldu", "talep_no": t.TalepNo})

	case "bursiyer":
		var t models.TalepBursiyer
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek"})
			return
		}
		t.UyeID = uyeID
		if err := h.Service.SubmitBursiyer(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Bursiyer işlem talebi oluşturuldu", "talep_no": t.TalepNo})

	case "proje_iptali":
		var t models.TalepProjeIptali
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek"})
			return
		}
		t.UyeID = uyeID
		if err := h.Service.SubmitProjeIptali(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Proje iptali talebi oluşturuldu", "talep_no": t.TalepNo})

	case "bilgi_degisimi":
		var t models.TalepBilgiDegisimi
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek"})
			return
		}
		t.UyeID = uyeID
		if err := h.Service.SubmitBilgiDegisimi(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Bilgi değişimi talebi oluşturuldu", "talep_no": t.TalepNo})

	case "proje_dondurma":
		var t models.TalepProjeDondurma
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek"})
			return
		}
		t.UyeID = uyeID
		if err := h.Service.SubmitProjeDondurma(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Proje dondurma talebi oluşturuldu", "talep_no": t.TalepNo})

	case "malzeme_guncelleme":
		var t models.TalepMalzemeGuncelleme
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek"})
			return
		}
		t.UyeID = uyeID
		if err := h.Service.SubmitMalzemeGuncelleme(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Malzeme güncelleme talebi oluşturuldu", "talep_no": t.TalepNo})

	case "avans":
		var t models.TalepAvans
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek"})
			return
		}
		t.UyeID = uyeID
		if err := h.Service.SubmitAvans(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Avans talebi oluşturuldu", "talep_no": t.TalepNo})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bilinmeyen talep tipi: " + tip})
	}
}

// GetAllTalepler, Admin/TTO için tüm talepleri, akademisyen için kendi taleplerini listeler.
// GET /api/talepler?sadece_bekleyen=true
// Türkçe Yorum: Giriş yapan kullanıcının rolüne göre tümünü veya kendi taleplerini döner.
func (h *TalepHandler) GetAllTalepler(c *gin.Context) {
	sadeceBekleyen := c.Query("sadece_bekleyen") == "true"

	uyeID, hasUye := uyeIDFromContext(c)
	roleVal, _ := c.Get("role")
	roleStr, _ := roleVal.(string)

	// Türkçe Yorum: Admin/TTO/komisyon/dekan tüm talepleri görür; akademisyen yalnızca kendi taleplerini alır.
	isPrivileged := false
	for _, r := range strings.Split(roleStr, ",") {
		r = strings.TrimSpace(r)
		if r == "admin" || r == "tto" || r == "komisyon" || r == "dekan" {
			isPrivileged = true
			break
		}
	}

	var list []models.TalepOzet
	var err error

	if isPrivileged {
		list, err = h.Service.GetAllTalepler(sadeceBekleyen)
	} else if hasUye && uyeID > 0 {
		list, err = h.Service.GetTaleplerByUye(uyeID, sadeceBekleyen)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Talepler alınamadı"})
		return
	}
	if list == nil {
		list = []models.TalepOzet{}
	}
	c.JSON(http.StatusOK, gin.H{"talepler": list})
}

// OnayTalep, Admin/TTO tarafından talep onay veya reddini işler.
// POST /api/talep/onay
func (h *TalepHandler) OnayTalep(c *gin.Context) {
	var istek models.TalepOnayIstek
	if err := c.ShouldBindJSON(&istek); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek parametreleri"})
		return
	}
	if uyeID, ok := uyeIDFromContext(c); ok {
		istek.IslemiYapanID = uyeID
	}
	if err := h.Service.OnayTalep(&istek); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Talep durumu güncellendi"})
}

// GetTaleplerByProje, projeye ait akademisyen taleplerini listeler.
// GET /api/proje/:id/talepler
// Türkçe Yorum: Akademisyen kendi projesi için geçmiş taleplerini görüntüler.
func (h *TalepHandler) GetTaleplerByProje(c *gin.Context) {
	projeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}
	repo := h.Service.Repo
	type sonuc struct {
		EkSure           interface{} `json:"ek_sure"`
		EkButce          interface{} `json:"ek_butce"`
		FasilAktarimi    interface{} `json:"fasil_aktarimi"`
		Arastirmaci      interface{} `json:"arastirmaci"`
		Bursiyer         interface{} `json:"bursiyer"`
		ProjeIptali      interface{} `json:"proje_iptali"`
		BilgiDegisimi    interface{} `json:"bilgi_degisimi"`
		ProjeDondurma    interface{} `json:"proje_dondurma"`
		MalzemeGuncelleme interface{} `json:"malzeme_guncelleme"`
		Avans            interface{} `json:"avans"`
	}
	s := sonuc{}
	s.EkSure, _ = repo.GetEkSureByProje(projeID)
	s.EkButce, _ = repo.GetEkButceByProje(projeID)
	s.FasilAktarimi, _ = repo.GetFasilAktarimiByProje(projeID)
	s.Arastirmaci, _ = repo.GetArastirmaciByProje(projeID)
	s.Bursiyer, _ = repo.GetBursiyerByProje(projeID)
	s.ProjeIptali, _ = repo.GetProjeIptaliByProje(projeID)
	s.BilgiDegisimi, _ = repo.GetBilgiDegisimiByProje(projeID)
	s.ProjeDondurma, _ = repo.GetProjeDondurmaByProje(projeID)
	s.MalzemeGuncelleme, _ = repo.GetMalzemeGuncellemeByProje(projeID)
	s.Avans, _ = repo.GetAvansByProje(projeID)
	c.JSON(http.StatusOK, s)
}
