package api

import (
	"net/http"

	"bap_ai/backend/models"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// DashboardHandler yapısı, dashboard HTTP handler'larını barındırır.
type DashboardHandler struct {
	DashboardService *service.DashboardService
}

// NewDashboardHandler fonksiyonu, yeni bir DashboardHandler nesnesi döner.
func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{DashboardService: dashboardService}
}

// GetStats fonksiyonu, giriş yapan kullanıcının dashboard istatistiklerini döner.
// GET /api/dashboard/stats
func (h *DashboardHandler) GetStats(c *gin.Context) {
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

	// Servis katmanından istatistikler getirilir
	stats, err := h.DashboardService.GetStats(uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "İstatistikler alınamadı",
		})
		return
	}

	// Başarılı yanıt döner
	c.JSON(http.StatusOK, stats)
}

// GetRecentProjects fonksiyonu, giriş yapan kullanıcının son başvurularını döner.
// GET /api/dashboard/recent-projects
func (h *DashboardHandler) GetRecentProjects(c *gin.Context) {
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

	// Servis katmanından son başvurular getirilir
	projeler, err := h.DashboardService.GetRecentProjects(uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Son başvurular alınamadı",
		})
		return
	}

	// Nil slice yerine boş array döndür (frontend tarafında JSON parse hatası önlenir)
	if projeler == nil {
		projeler = []models.ProjeOzet{}
	}

	// Başarılı yanıt döner
	c.JSON(http.StatusOK, gin.H{
		"projeler": projeler,
	})
}

// GetAllProjects fonksiyonu, giriş yapan kullanıcının tüm başvurularını döner.
// GET /api/dashboard/all-projects
func (h *DashboardHandler) GetAllProjects(c *gin.Context) {
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

	// Servis katmanından tüm başvurular getirilir
	projeler, err := h.DashboardService.GetAllProjects(uyeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Tüm başvurular alınamadı",
		})
		return
	}

	// Nil slice yerine boş array döndür (frontend tarafında JSON parse hatası önlenir)
	if projeler == nil {
		projeler = []models.ProjeOzet{}
	}

	// Başarılı yanıt döner
	c.JSON(http.StatusOK, gin.H{
		"projeler": projeler,
	})
}
