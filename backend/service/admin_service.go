package service

import (
	"bap_ai/backend/models"
	"bap_ai/backend/repository"
	"fmt"

	"golang.org/x/crypto/bcrypt"
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
func (s *AdminService) UpdateUserRole(uyeID int, rolAdi string) error {
	return s.adminRepo.UpdateUserRole(uyeID, rolAdi)
}

// UpdateUserStatus, kullanıcının aktiflik durumunu günceller.
func (s *AdminService) UpdateUserStatus(uyeID int, isActive bool) error {
	return s.adminRepo.UpdateUserStatus(uyeID, isActive)
}

// AssignHakem, admin tarafından bir projeye belirli bir hakemi atar.
func (s *AdminService) AssignHakem(projeID int, hakemID int) error {
	return s.adminRepo.AssignHakemToProje(projeID, hakemID)
}

// UpdateProjectStatus, projenin genel durumunu günceller.
func (s *AdminService) UpdateProjectStatus(projeID int, durum string) error {
	return s.adminRepo.UpdateProjectStatus(projeID, durum)
}

// GetProjectDetailsForAdmin, yöneticiler için proje detayını getirir.
func (s *AdminService) GetProjectDetailsForAdmin(projeID int) (*repository.ProjectDetail, error) {
	return s.adminRepo.GetProjectDetailsForAdmin(projeID)
}

// GetDegerlendirilmemisProjeleri, hakem atanması gereken projeleri döner.
func (s *AdminService) GetDegerlendirilmemisProjeleri() ([]models.Proje, error) {
	return s.adminRepo.GetDegerlendirilmemisProjeleri()
}

// GetHakemListesi, aktif hakem kullanıcılarını döner.
func (s *AdminService) GetHakemListesi() ([]models.Uye, error) {
	return s.adminRepo.GetHakemListesi()
}

// GetProjeyeAtananHakemler, bir projeye atanan hakemlerin durumlarını döner.
func (s *AdminService) GetProjeyeAtananHakemler(projeID int) ([]repository.AtananHakemDetay, error) {
	return s.adminRepo.GetProjeyeAtananHakemler(projeID)
}

// GetBapTurleri, sistemdeki BAP proje türlerini döner.
func (s *AdminService) GetBapTurleri(onlyActive bool) ([]models.ProjeBapTuru, error) {
	return s.adminRepo.GetBapTurleri(onlyActive)
}

// CreateBapTuru, yeni bir BAP proje türü oluşturur.
func (s *AdminService) CreateBapTuru(bt *models.ProjeBapTuru) error {
	return s.adminRepo.CreateBapTuru(bt)
}

// UpdateBapTuru, mevcut bir BAP proje türünü günceller.
func (s *AdminService) UpdateBapTuru(bt *models.ProjeBapTuru) error {
	return s.adminRepo.UpdateBapTuru(bt)
}

// CreateUser, admin tarafından yeni bir kullanıcı ekleme işlemini gerçekleştirir.
// Eğer istekte şifre belirtilmemişse, kullanıcının ilk girişte şifre oluşturması için şifresi "pending" olarak atanır.
func (s *AdminService) CreateUser(req *models.AdminCreateUserRequest) error {
	var passwordHash string
	if req.Sifre != "" {
		// Şifreyi bcrypt ile hashle
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Sifre), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("şifre hashlenemedi: %w", err)
		}
		passwordHash = string(hashedPassword)
	} else {
		// Şifre tanımlanmamışsa ilk giriş kontrolü için "pending" yapıyoruz
		passwordHash = "pending"
	}

	// Repository'ye isteği yönlendir
	return s.adminRepo.CreateUser(req, passwordHash)
}

// UpdateUser, admin tarafından bir kullanıcının temel ve detay bilgilerini günceller.
func (s *AdminService) UpdateUser(uyeID int, req *models.AdminUpdateUserRequest) error {
	return s.adminRepo.UpdateUser(uyeID, req)
}

