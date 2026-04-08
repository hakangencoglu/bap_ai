package api

import (
	"net/http"

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
		UyeID  int `json:"uye_id" binding:"required"`
		RoleID int `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek parametreleri"})
		return
	}

	err := h.adminService.UpdateUserRole(req.UyeID, req.RoleID)
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
		IsActive bool `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek parametreleri"})
		return
	}

	err := h.adminService.UpdateUserStatus(req.UyeID, req.IsActive)
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
