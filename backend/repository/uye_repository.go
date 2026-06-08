package repository

import (
	"database/sql"
	"fmt"

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
// Sadece temel bilgiler (ad, soyad, eposta, sifre_hash) kaydedilir.
func (r *UyeRepository) CreateUye(uye *models.Uye) error {
	// INSERT sorgusu ile yeni üye eklenir ve otomatik oluşan alanlar geri alınır
	query := `
		INSERT INTO uye (ad, soyad, eposta, sifre_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING uye_id, aktif_mi, olusturma_tarihi, guncelleme_tarihi
	`
	err := r.DB.QueryRow(
		query,
		uye.Ad,
		uye.Soyad,
		uye.Eposta,
		uye.SifreHash,
	).Scan(&uye.UyeID, &uye.AktifMi, &uye.OlusturmaTarihi, &uye.GuncellemeTarihi)

	return err
}

// CreateUyeDetay fonksiyonu, üye detay bilgilerini veritabanına kaydeder.
func (r *UyeRepository) CreateUyeDetay(detay *models.UyeDetay) error {
	// Üye detay bilgileri oluşturulur
	query := `
		INSERT INTO uye_detay (uye_id, rol, unvan, bolum, telefon, izu_uyesi, profil_tamamlandi)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING detay_id, olusturma_tarihi, guncelleme_tarihi
	`
	err := r.DB.QueryRow(
		query,
		detay.UyeID,
		detay.Rol,
		detay.Unvan,
		detay.Bolum,
		detay.Telefon,
		detay.IzuUyesi,
		detay.ProfilTamamlandi,
	).Scan(&detay.DetayID, &detay.OlusturmaTarihi, &detay.GuncellemeTarihi)

	return err
}

// GetUyeDetayByUyeID fonksiyonu, üye ID'sine göre detay bilgilerini döner.
func (r *UyeRepository) GetUyeDetayByUyeID(uyeID int) (*models.UyeDetay, error) {
	detay := &models.UyeDetay{}

	// Üye ID'sine göre detay bilgileri sorgulanır
	query := `
		SELECT detay_id, uye_id, COALESCE(rol, ''), COALESCE(unvan, ''), COALESCE(bolum, ''),
		       COALESCE(telefon, ''), izu_uyesi, profil_tamamlandi, olusturma_tarihi, guncelleme_tarihi
		FROM uye_detay
		WHERE uye_id = $1
	`
	err := r.DB.QueryRow(query, uyeID).Scan(
		&detay.DetayID,
		&detay.UyeID,
		&detay.Rol,
		&detay.Unvan,
		&detay.Bolum,
		&detay.Telefon,
		&detay.IzuUyesi,
		&detay.ProfilTamamlandi,
		&detay.OlusturmaTarihi,
		&detay.GuncellemeTarihi,
	)
	if err != nil {
		return nil, err
	}

	return detay, nil
}

// UpdateUyeDetay fonksiyonu, mevcut üye detay bilgilerini günceller.
func (r *UyeRepository) UpdateUyeDetay(detay *models.UyeDetay) error {
	// Üye detay bilgileri güncellenir
	query := `
		UPDATE uye_detay
		SET rol = $1, unvan = $2, bolum = $3, telefon = $4, izu_uyesi = $5,
		    profil_tamamlandi = $6, guncelleme_tarihi = CURRENT_TIMESTAMP
		WHERE uye_id = $7
		RETURNING guncelleme_tarihi
	`
	err := r.DB.QueryRow(
		query,
		detay.Rol,
		detay.Unvan,
		detay.Bolum,
		detay.Telefon,
		detay.IzuUyesi,
		detay.ProfilTamamlandi,
		detay.UyeID,
	).Scan(&detay.GuncellemeTarihi)

	return err
}

// GetUyeByEmail fonksiyonu, e-posta adresine göre üyeyi detay bilgileriyle birlikte getirir.
func (r *UyeRepository) GetUyeByEmail(email string) (*models.UyeWithDetay, error) {
	uye := &models.UyeWithDetay{}

	// E-posta adresine göre üye ve detay bilgileri JOIN ile sorgulanır
	query := `
		SELECT u.uye_id, u.ad, u.soyad, u.eposta, u.sifre_hash, u.aktif_mi,
		       COALESCE(d.rol, ''), COALESCE(d.unvan, ''), COALESCE(d.bolum, ''),
		       COALESCE(d.telefon, ''), COALESCE(d.izu_uyesi, FALSE),
		       COALESCE(d.profil_tamamlandi, FALSE),
		       u.olusturma_tarihi, u.guncelleme_tarihi
		FROM uye u
		LEFT JOIN uye_detay d ON u.uye_id = d.uye_id
		WHERE u.eposta = $1
	`
	err := r.DB.QueryRow(query, email).Scan(
		&uye.UyeID,
		&uye.Ad,
		&uye.Soyad,
		&uye.Eposta,
		&uye.SifreHash,
		&uye.AktifMi,
		&uye.Rol,
		&uye.Unvan,
		&uye.Bolum,
		&uye.Telefon,
		&uye.IzuUyesi,
		&uye.ProfilTamamlandi,
		&uye.OlusturmaTarihi,
		&uye.GuncellemeTarihi,
	)
	if err != nil {
		return nil, err
	}

	return uye, nil
}

// GetUyeByID fonksiyonu, üye ID'sine göre üyeyi detay bilgileriyle birlikte getirir.
func (r *UyeRepository) GetUyeByID(id int) (*models.UyeWithDetay, error) {
	uye := &models.UyeWithDetay{}

	// ID'ye göre üye ve detay bilgileri JOIN ile sorgulanır
	query := `
		SELECT u.uye_id, u.ad, u.soyad, u.eposta, u.sifre_hash, u.aktif_mi,
		       COALESCE(d.rol, ''), COALESCE(d.unvan, ''), COALESCE(d.bolum, ''),
		       COALESCE(d.telefon, ''), COALESCE(d.izu_uyesi, FALSE),
		       COALESCE(d.profil_tamamlandi, FALSE),
		       u.olusturma_tarihi, u.guncelleme_tarihi
		FROM uye u
		LEFT JOIN uye_detay d ON u.uye_id = d.uye_id
		WHERE u.uye_id = $1
	`
	err := r.DB.QueryRow(query, id).Scan(
		&uye.UyeID,
		&uye.Ad,
		&uye.Soyad,
		&uye.Eposta,
		&uye.SifreHash,
		&uye.AktifMi,
		&uye.Rol,
		&uye.Unvan,
		&uye.Bolum,
		&uye.Telefon,
		&uye.IzuUyesi,
		&uye.ProfilTamamlandi,
		&uye.OlusturmaTarihi,
		&uye.GuncellemeTarihi,
	)
	if err != nil {
		return nil, err
	}

	return uye, nil
}

// UpdateUyeRol fonksiyonu, uye tablosundaki rol alanını günceller (geriye dönük uyumluluk).
func (r *UyeRepository) UpdateUyeRol(uyeID int, rol string) error {
	// Üye tablosundaki rol alanı da güncellenir
	query := `UPDATE uye SET rol = $1, guncelleme_tarihi = CURRENT_TIMESTAMP WHERE uye_id = $2`
	_, err := r.DB.Exec(query, rol, uyeID)
	return err
}

// UpsertSistemRol fonksiyonu, kullanıcının sistem rolünü sistem_rol tablosunda günceller.
// Önce kullanıcının mevcut rollerini temizler, ardından yeni rolü atar.
// Bu sayede sistem_rol tablosu her zaman güncel kalır.
func (r *UyeRepository) UpsertSistemRol(uyeID int, rolAdi string) error {
	// 1. Kullanıcının mevcut sistem rollerini temizle
	_, err := r.DB.Exec(`DELETE FROM sistem_rol WHERE uye_id = $1`, uyeID)
	if err != nil {
		return fmt.Errorf("mevcut sistem rolleri temizlenemedi: %w", err)
	}

	// 2. Yeni rolün ID'sini sistem_rol_tanimlama tablosundan bul ve ata
	insertQuery := `
		INSERT INTO sistem_rol (uye_id, sistem_rol_id)
		SELECT $1, rol_id FROM sistem_rol_tanimlama WHERE rol_adi = $2
		ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING
	`
	_, err = r.DB.Exec(insertQuery, uyeID, rolAdi)
	if err != nil {
		return fmt.Errorf("sistem rolü atanamadı: %w", err)
	}

	return nil
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
	// uye_detay tablosundan rol filtrelemesi yapılır
	query := `
		SELECT u.uye_id, u.ad, u.soyad, COALESCE(d.unvan, ''), COALESCE(d.bolum, '')
		FROM uye u
		LEFT JOIN uye_detay d ON u.uye_id = d.uye_id
		WHERE d.rol = $1 AND u.aktif_mi = true
		ORDER BY u.ad, u.soyad
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
	// uye_detay tablosu ile JOIN yapılarak arama gerçekleştirilir
	searchQuery := `
		SELECT u.uye_id, u.ad, u.soyad, COALESCE(d.unvan, ''), COALESCE(d.bolum, ''), COALESCE(d.rol, '')
		FROM uye u
		LEFT JOIN uye_detay d ON u.uye_id = d.uye_id
		WHERE u.aktif_mi = true
		  AND (u.ad ILIKE '%' || $1 || '%' OR u.soyad ILIKE '%' || $1 || '%' OR u.eposta ILIKE '%' || $1 || '%')
		ORDER BY u.ad, u.soyad
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

// UpdateUyePassword kullanıcının şifresini günceller
func (r *UyeRepository) UpdateUyePassword(uyeID int, hashedPassword string) error {
	query := `UPDATE uye SET sifre_hash = $1, guncelleme_tarihi = CURRENT_TIMESTAMP WHERE uye_id = $2`
	_, err := r.DB.Exec(query, hashedPassword, uyeID)
	return err
}

