package api

import (
	"net/http"

	"bap_ai/backend/models"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// ProfilHandler yapısı, profil sayfası HTTP handler'larını barındırır.
type ProfilHandler struct {
	ProfilService *service.ProfilService
}

// NewProfilHandler fonksiyonu, yeni bir ProfilHandler nesnesi döner.
func NewProfilHandler(profilService *service.ProfilService) *ProfilHandler {
	return &ProfilHandler{ProfilService: profilService}
}

// GetProfilBilgileri fonksiyonu, giriş yapan kullanıcının profil bilgilerini döner.
// GET /api/profil/bilgiler
func (h *ProfilHandler) GetProfilBilgileri(c *gin.Context) {
	// Middleware'den gelen uye_id alınır (JWT token'dan parse edilmiş)
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Kullanıcı bilgisi bulunamadı",
		})
		return
	}

	// JWT MapClaims sayıları float64 olarak tutar, int'e çevrilir
	uyeID := int(uyeIDFloat.(float64))

	// Servis katmanından profil bilgileri getirilir
	uye, err := h.ProfilService.GetProfilBilgileri(uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Profil bilgileri alınamadı",
		})
		return
	}

	// Giriş bilgileri hariç yanıt döner (sifre_hash zaten json:"-" ile gizli)
	c.JSON(http.StatusOK, gin.H{
		"uye_id":            uye.UyeID,
		"rol":               uye.Rol,
		"ad":                uye.Ad,
		"soyad":             uye.Soyad,
		"unvan":             uye.Unvan,
		"bolum":             uye.Bolum,
		"telefon":           uye.Telefon,
		"eposta":            uye.Eposta,
		"izu_uyesi":         uye.IzuUyesi,
		"aktif_mi":          uye.AktifMi,
		"olusturma_tarihi":  uye.OlusturmaTarihi,
	})
}

// GetProfilProjeleri fonksiyonu, giriş yapan kullanıcının projelerini profil formatında döner.
// GET /api/profil/projeler
func (h *ProfilHandler) GetProfilProjeleri(c *gin.Context) {
	// Middleware'den gelen uye_id alınır
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Kullanıcı bilgisi bulunamadı",
		})
		return
	}

	// JWT MapClaims sayıları float64 olarak tutar, int'e çevrilir
	uyeID := int(uyeIDFloat.(float64))

	// Servis katmanından profil projeleri getirilir
	projeler, err := h.ProfilService.GetProfilProjeleri(uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Profil projeleri alınamadı",
		})
		return
	}

	// Nil slice yerine boş array döndür
	if projeler == nil {
		projeler = []models.ProfilProjeBilgisi{}
	}

	// Başarılı yanıt döner
	c.JSON(http.StatusOK, gin.H{
		"projeler": projeler,
	})
}
