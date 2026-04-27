package service

import (
	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// AdminStats yapısı, admin dashboard'ında gösterilecek istatistikleri tutar.
type AdminStats struct {
	ToplamProje   int   `json:"toplam_proje"`
	OnayBekleyen  int   `json:"onay_bekleyen"`
	ToplamKullanici int64 `json:"toplam_kullanici"`
	ToplamButce   float64 `json:"toplam_butce"`
}

// AdminService, admin işlemleri için iş kurallarını barındırır.
type AdminService struct {
	adminRepo *repository.AdminRepository
	projeRepo *repository.ProjeRepository
}

// NewAdminService, yeni bir AdminService örneği oluşturur.
func NewAdminService(adminRepo *repository.AdminRepository, projeRepo *repository.ProjeRepository) *AdminService {
	return &AdminService{
		adminRepo: adminRepo,
		projeRepo: projeRepo,
	}
}

// GetAdminDashboardStats, dashboard için özet istatistikleri hazırlar.
func (s *AdminService) GetAdminDashboardStats() (*AdminStats, error) {
	// Toplam proje istatistiklerini admin repository'den alalım
	projeStats, err := s.adminRepo.GetProjectStats()
	if err != nil {
		return nil, err
	}

	// Kullanıcı sayısını admin repository'den alalım
	userCount, err := s.adminRepo.GetTotalUsersCount()
	if err != nil {
		userCount = 0
	}

	return &AdminStats{
		ToplamProje:   projeStats.AktifProje + projeStats.OnayBekleyen + projeStats.Tamamlanan,
		OnayBekleyen:  projeStats.OnayBekleyen,
		ToplamKullanici: userCount,
		ToplamButce:   projeStats.ToplamButce,
	}, nil
}

// GetAllProjects, sistemdeki tüm projeleri döner.
func (s *AdminService) GetAllProjects() ([]models.Proje, error) {
	return s.adminRepo.GetAllProjects()
}

// GetAllUsers, sistemdeki tüm kullanıcıları döner.
func (s *AdminService) GetAllUsers() ([]models.Uye, error) {
	return s.adminRepo.GetAllUsers()
}

// UpdateUserRole, kullanıcının rolünü günceller.
func (s *AdminService) UpdateUserRole(uyeID int, roleID int) error {
	return s.adminRepo.UpdateUserRole(uyeID, roleID)
}

// UpdateUserStatus, kullanıcının aktiflik durumunu günceller.
func (s *AdminService) UpdateUserStatus(uyeID int, isActive bool) error {
	return s.adminRepo.UpdateUserStatus(uyeID, isActive)
}

// AssignHakem, bir projeye hakem ataması yapar.
func (s *AdminService) AssignHakem(projeID int, hakemID int) error {
	// Hakem ataması logic'i. Aslında hakem repository'sinde AddDegerlendirme yapılabilir
	// Şimdilik sadece implemente edilecek yeri bırakalım.
	// Yeni bir degerlendirme kaydı olusturulacak:
	return nil 
}

// UpdateProjectStatus, projenin genel durumunu günceller.
func (s *AdminService) UpdateProjectStatus(projeID int, durum string) error {
	return s.adminRepo.UpdateProjectStatus(projeID, durum)
}

// GetProjectDetailsForAdmin, yöneticiler için proje detayını getirir.
func (s *AdminService) GetProjectDetailsForAdmin(projeID int) (*repository.ProjectDetail, error) {
	return s.adminRepo.GetProjectDetailsForAdmin(projeID)
}

