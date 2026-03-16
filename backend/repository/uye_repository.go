package repository

import (
	"database/sql"

	"bap_ai/backend/models"
)

// UyeRepository yapısı, üye tablosuna erişim sorgularını barındırır.
type UyeRepository struct {
	DB *sql.DB
}

// NewUyeRepository fonksiyonu, yeni bir UyeRepository nesnesi döner.
func NewUyeRepository(db *sql.DB) *UyeRepository {
	return &UyeRepository{DB: db}
}

// CreateUye fonksiyonu, yeni bir üyeyi veritabanına kaydeder.
func (r *UyeRepository) CreateUye(uye *models.Uye) error {
	// INSERT sorgusu ile yeni üye eklenir ve otomatik oluşan alanlar geri alınır
	query := `
		INSERT INTO uye (role_id, unvan, ad, soyad, bolum, iletisim_tel, iletisim_mail, password_hash, izu_akademisyen, izu_ogrenci)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING uye_id, is_active, created_at, updated_at
	`
	err := r.DB.QueryRow(
		query,
		uye.RoleID,
		uye.Unvan,
		uye.Ad,
		uye.Soyad,
		uye.Bolum,
		uye.IletisimTel,
		uye.IletisimMail,
		uye.PasswordHash,
		uye.IzuAkademisyen,
		uye.IzuOgrenci,
	).Scan(&uye.UyeID, &uye.IsActive, &uye.CreatedAt, &uye.UpdatedAt)

	return err
}

// GetUyeByEmail fonksiyonu, e-posta adresine göre üyeyi veritabanından getirir.
func (r *UyeRepository) GetUyeByEmail(email string) (*models.Uye, error) {
	uye := &models.Uye{}

	// E-posta adresine göre tek bir satır sorgulanır
	query := `
		SELECT uye_id, role_id, unvan, ad, soyad, bolum, iletisim_tel, iletisim_mail, 
		       password_hash, izu_akademisyen, izu_ogrenci, is_active, created_at, updated_at
		FROM uye
		WHERE iletisim_mail = $1
	`
	err := r.DB.QueryRow(query, email).Scan(
		&uye.UyeID,
		&uye.RoleID,
		&uye.Unvan,
		&uye.Ad,
		&uye.Soyad,
		&uye.Bolum,
		&uye.IletisimTel,
		&uye.IletisimMail,
		&uye.PasswordHash,
		&uye.IzuAkademisyen,
		&uye.IzuOgrenci,
		&uye.IsActive,
		&uye.CreatedAt,
		&uye.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return uye, nil
}

// GetUyeByID fonksiyonu, üye ID'sine göre üyeyi veritabanından getirir.
func (r *UyeRepository) GetUyeByID(id int) (*models.Uye, error) {
	uye := &models.Uye{}

	// ID'ye göre tek bir satır sorgulanır
	query := `
		SELECT uye_id, role_id, unvan, ad, soyad, bolum, iletisim_tel, iletisim_mail,
		       password_hash, izu_akademisyen, izu_ogrenci, is_active, created_at, updated_at
		FROM uye
		WHERE uye_id = $1
	`
	err := r.DB.QueryRow(query, id).Scan(
		&uye.UyeID,
		&uye.RoleID,
		&uye.Unvan,
		&uye.Ad,
		&uye.Soyad,
		&uye.Bolum,
		&uye.IletisimTel,
		&uye.IletisimMail,
		&uye.PasswordHash,
		&uye.IzuAkademisyen,
		&uye.IzuOgrenci,
		&uye.IsActive,
		&uye.CreatedAt,
		&uye.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return uye, nil
}
