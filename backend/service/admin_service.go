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
// Türkçe Yorum: Şifre boşsa, şifreyi "pending" yapıp zorla değiştirme bayrağını true yapıyoruz
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
		// Şifre tanımlanmamışsa ilk giriş kontrolü için "pending" yapıyoruz ve zorla değiştirmeyi aktif ediyoruz
		passwordHash = "pending"
		req.SifreDegistirZorla = true
	}

	// Repository'ye isteği yönlendir
	return s.adminRepo.CreateUser(req, passwordHash)
}

// BulkCreateUsers toplu kullanıcı ekleme işlemini gerçekleştirir.
// Türkçe Yorum: CSV'den okunan kullanıcı istek listesini işler, başarılı ve hatalı ekleme raporu hazırlar.
func (s *AdminService) BulkCreateUsers(requests []models.AdminCreateUserRequest) (int, int, []map[string]interface{}) {
	var successCount int
	var errorCount int
	var errorDetails []map[string]interface{}

	for idx, req := range requests {
		// Her istek için CreateUser çağrılır
		err := s.CreateUser(&req)
		if err != nil {
			errorCount++
			errorDetails = append(errorDetails, map[string]interface{}{
				"satir":  idx + 2, // Başlık satırı ve 0-tabanlı indeksi telafi etmek için +2
				"eposta": req.Eposta,
				"hata":   err.Error(),
			})
		} else {
			successCount++
		}
	}

	return successCount, errorCount, errorDetails
}

// UpdateUser, admin tarafından bir kullanıcının temel ve detay bilgilerini günceller.
func (s *AdminService) UpdateUser(uyeID int, req *models.AdminUpdateUserRequest) error {
	return s.adminRepo.UpdateUser(uyeID, req)
}

// GetSayfaYetkiMatrix, sistemdeki roller, sayfalar ve yetki matrisini döner.
// Türkçe Yorum: Admin yetkilendirme sayfası için matris verilerini repository'den çeker.
func (s *AdminService) GetSayfaYetkiMatrix() (*models.SayfaYetkiMatrix, error) {
	return s.adminRepo.GetSayfaYetkiMatrix()
}

// UpdateSayfaYetki, adminin gönderdiği sayfa rol yetki güncellemelerini işler.
// Türkçe Yorum: Yetkilendirme değişikliklerini kaydetmek için repository katmanına yollar.
func (s *AdminService) UpdateSayfaYetki(permissions []models.UpdateSayfaYetkiItem) error {
	return s.adminRepo.UpdateSayfaYetki(permissions)
}

// CheckPageAccess, belirtilen rollerin ilgili sayfaya erişim hakkı olup olmadığını kontrol eder.
// Türkçe Yorum: Sayfa yönlendirme koruması için kullanıcının rollerinin URL yetkisini sorgular.
func (s *AdminService) CheckPageAccess(roles []string, path string) (bool, error) {
	return s.adminRepo.CheckPageAccess(roles, path)
}

// CreateRole yeni bir sistem rolü oluşturur ve bu role sayfa yetkileri tanımlar.
// Türkçe Yorum: Admin'in yeni rol ekleme isteğini iş mantığı katmanında işler ve repository'ye aktarır.
func (s *AdminService) CreateRole(rolAdi string, sayfaIDs []int) error {
	return s.adminRepo.CreateRole(rolAdi, sayfaIDs)
}

// UpdateRole mevcut bir sistem rolünün bilgilerini günceller.
// Türkçe Yorum: Rol adı ve izin verilen sayfaları güncelleme isteğini repository katmanına iletir.
func (s *AdminService) UpdateRole(rolID int, rolAdi string, sayfaIDs []int) error {
	return s.adminRepo.UpdateRole(rolID, rolAdi, sayfaIDs)
}

// DeleteRole sistem rolünü siler.
// Türkçe Yorum: Belirtilen rolün sistemden tamamen kaldırılması isteğini repository'ye iletir.
func (s *AdminService) DeleteRole(rolID int) error {
	return s.adminRepo.DeleteRole(rolID)
}


