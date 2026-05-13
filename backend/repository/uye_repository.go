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
		INSERT INTO uye (rol, ad, soyad, unvan, bolum, eposta, telefon, izu_uyesi, sifre_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING uye_id, aktif_mi, olusturma_tarihi, guncelleme_tarihi
	`
	err := r.DB.QueryRow(
		query,
		uye.Rol,
		uye.Ad,
		uye.Soyad,
		uye.Unvan,
		uye.Bolum,
		uye.Eposta,
		uye.Telefon,
		uye.IzuUyesi,
		uye.SifreHash,
	).Scan(&uye.UyeID, &uye.AktifMi, &uye.OlusturmaTarihi, &uye.GuncellemeTarihi)

	return err
}

// GetUyeByEmail fonksiyonu, e-posta adresine göre üyeyi veritabanından getirir.
func (r *UyeRepository) GetUyeByEmail(email string) (*models.Uye, error) {
	uye := &models.Uye{}

	// E-posta adresine göre tek bir satır sorgulanır
	query := `
		SELECT uye_id, rol, ad, soyad, unvan, bolum, eposta, telefon,
		       sifre_hash, izu_uyesi, aktif_mi, olusturma_tarihi, guncelleme_tarihi
		FROM uye
		WHERE eposta = $1
	`
	err := r.DB.QueryRow(query, email).Scan(
		&uye.UyeID,
		&uye.Rol,
		&uye.Ad,
		&uye.Soyad,
		&uye.Unvan,
		&uye.Bolum,
		&uye.Eposta,
		&uye.Telefon,
		&uye.SifreHash,
		&uye.IzuUyesi,
		&uye.AktifMi,
		&uye.OlusturmaTarihi,
		&uye.GuncellemeTarihi,
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
		SELECT uye_id, rol, ad, soyad, unvan, bolum, eposta, telefon,
		       sifre_hash, izu_uyesi, aktif_mi, olusturma_tarihi, guncelleme_tarihi
		FROM uye
		WHERE uye_id = $1
	`
	err := r.DB.QueryRow(query, id).Scan(
		&uye.UyeID,
		&uye.Rol,
		&uye.Ad,
		&uye.Soyad,
		&uye.Unvan,
		&uye.Bolum,
		&uye.Eposta,
		&uye.Telefon,
		&uye.SifreHash,
		&uye.IzuUyesi,
		&uye.AktifMi,
		&uye.OlusturmaTarihi,
		&uye.GuncellemeTarihi,
	)
	if err != nil {
		return nil, err
	}

	return uye, nil
}

// AkademisyenOzet yapısı, yürütücü seçimi için gerekli özet bilgileri tutar.
type AkademisyenOzet struct {
	UyeID int    `json:"uye_id"`
	Ad    string `json:"ad"`
	Soyad string `json:"soyad"`
	Unvan string `json:"unvan"`
	Bolum string `json:"bolum"`
}

// GetUyelerByRol fonksiyonu, belirtilen role sahip aktif kullanıcıları döner.
func (r *UyeRepository) GetUyelerByRol(rol string) ([]AkademisyenOzet, error) {
	query := `
		SELECT uye_id, ad, soyad, COALESCE(unvan, ''), COALESCE(bolum, '')
		FROM uye
		WHERE rol = $1 AND aktif_mi = true
		ORDER BY ad, soyad
	`
	rows, err := r.DB.Query(query, rol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uyeler []AkademisyenOzet
	for rows.Next() {
		var u AkademisyenOzet
		if err := rows.Scan(&u.UyeID, &u.Ad, &u.Soyad, &u.Unvan, &u.Bolum); err != nil {
			return nil, err
		}
		uyeler = append(uyeler, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return uyeler, nil
}

// UyeAramaOzet yapısı, kullanıcı arama sonuçları için özet bilgileri tutar.
type UyeAramaOzet struct {
	UyeID int    `json:"uye_id"`
	Ad    string `json:"ad"`
	Soyad string `json:"soyad"`
	Unvan string `json:"unvan"`
	Bolum string `json:"bolum"`
	Rol   string `json:"rol"`
}

// SearchUyeler fonksiyonu, ad/soyad/e-posta ile aktif kullanıcı araması yapar.
func (r *UyeRepository) SearchUyeler(query string) ([]UyeAramaOzet, error) {
	searchQuery := `
		SELECT uye_id, ad, soyad, COALESCE(unvan, ''), COALESCE(bolum, ''), COALESCE(rol, '')
		FROM uye
		WHERE aktif_mi = true
		  AND (ad ILIKE '%' || $1 || '%' OR soyad ILIKE '%' || $1 || '%' OR eposta ILIKE '%' || $1 || '%')
		ORDER BY ad, soyad
		LIMIT 20
	`
	rows, err := r.DB.Query(searchQuery, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uyeler []UyeAramaOzet
	for rows.Next() {
		var u UyeAramaOzet
		if err := rows.Scan(&u.UyeID, &u.Ad, &u.Soyad, &u.Unvan, &u.Bolum, &u.Rol); err != nil {
			return nil, err
		}
		uyeler = append(uyeler, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return uyeler, nil
}
