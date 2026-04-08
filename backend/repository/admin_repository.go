package repository

import (
	"database/sql"
	"log"

	"bap_ai/backend/models"
)

// AdminRepository, admin işlemleri için veritabanı erişimini sağlar.
type AdminRepository struct {
	DB *sql.DB
}

// NewAdminRepository, yeni bir AdminRepository örneği oluşturur.
func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{DB: db}
}

// GetTotalUsersCount, sistemdeki toplam aktif kullanıcı sayısını döner.
func (r *AdminRepository) GetTotalUsersCount() (int64, error) {
	var count int64
	err := r.DB.QueryRow("SELECT COUNT(*) FROM uye WHERE is_active = true").Scan(&count)
	if err != nil {
		log.Printf("GetTotalUsersCount hatası: %v", err)
	}
	return count, err
}

// GetAllUsers, sistemdeki tüm kullanıcıları döner.
func (r *AdminRepository) GetAllUsers() ([]models.Uye, error) {
	var users []models.Uye
	rows, err := r.DB.Query("SELECT uye_id, role_id, unvan, ad, soyad, bolum, iletisim_tel, iletisim_mail, izu_akademisyen, is_active FROM uye")
	if err != nil {
		log.Printf("GetAllUsers hatası: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u models.Uye
		if err := rows.Scan(&u.UyeID, &u.RoleID, &u.Unvan, &u.Ad, &u.Soyad, &u.Bolum, &u.IletisimTel, &u.IletisimMail, &u.IzuAkademisyen, &u.IsActive); err == nil {
			users = append(users, u)
		}
	}
	return users, nil
}

// GetAllProjects, sistemdeki tüm projeleri döner.
func (r *AdminRepository) GetAllProjects() ([]models.Proje, error) {
	var projes []models.Proje
	rows, err := r.DB.Query("SELECT proje_id, baslik_tr, durum, tur, created_at FROM proje")
	if err != nil {
		log.Printf("GetAllProjects hatası: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Proje
		if err := rows.Scan(&p.ProjeID, &p.BaslikTr, &p.Durum, &p.Tur, &p.CreatedAt); err == nil {
			projes = append(projes, p)
		}
	}
	return projes, nil
}

// UpdateUserRole, bir kullanıcının rolünü günceller.
func (r *AdminRepository) UpdateUserRole(uyeID int, roleID int) error {
	_, err := r.DB.Exec("UPDATE uye SET role_id = $1 WHERE uye_id = $2", roleID, uyeID)
	if err != nil {
		log.Printf("UpdateUserRole hatası: %v", err)
	}
	return err
}

// UpdateUserStatus, bir kullanıcının aktiflik durumunu günceller.
func (r *AdminRepository) UpdateUserStatus(uyeID int, isActive bool) error {
	_, err := r.DB.Exec("UPDATE uye SET is_active = $1 WHERE uye_id = $2", isActive, uyeID)
	if err != nil {
		log.Printf("UpdateUserStatus hatası: %v", err)
	}
	return err
}

// UpdateProjectStatus, projenin genel statüsünü günceller.
func (r *AdminRepository) UpdateProjectStatus(projeID int, durum string) error {
	_, err := r.DB.Exec("UPDATE proje SET durum = $1 WHERE proje_id = $2", durum, projeID)
	if err != nil {
		log.Printf("UpdateProjectStatus hatası: %v", err)
	}
	return err
}

// GetProjectStats, genel sistemdeki tüm proje istatistiklerini hesaplayıp döner.
func (r *AdminRepository) GetProjectStats() (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	// Bütün projeler için Aktif (onaylandi) sayısı:
	err := r.DB.QueryRow("SELECT COUNT(*) FROM proje WHERE durum = 'onaylandi'").Scan(&stats.AktifProje)
	if err != nil {
		return nil, err
	}

	// Onay bekleyen proje sayısı (incelemede)
	err = r.DB.QueryRow("SELECT COUNT(*) FROM proje WHERE durum = 'incelemede'").Scan(&stats.OnayBekleyen)
	if err != nil {
		return nil, err
	}

	// Tamamlanan proje sayısı
	err = r.DB.QueryRow("SELECT COUNT(*) FROM proje WHERE durum = 'tamamlandi'").Scan(&stats.Tamamlanan)
	if err != nil {
		return nil, err
	}

	// Toplam bütçe (Bütün projeler)
	err = r.DB.QueryRow("SELECT COALESCE(SUM(toplam_tutar), 0) FROM proje").Scan(&stats.ToplamButce)
	if err != nil {
		return nil, err
	}

	return stats, nil
}


