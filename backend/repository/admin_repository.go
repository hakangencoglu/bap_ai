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
	err := r.DB.QueryRow("SELECT COUNT(*) FROM uye WHERE aktif_mi = true").Scan(&count)
	if err != nil {
		log.Printf("GetTotalUsersCount hatası: %v", err)
	}
	return count, err
}

// GetAllUsers, sistemdeki tüm kullanıcıları döner.
func (r *AdminRepository) GetAllUsers() ([]models.Uye, error) {
	var users []models.Uye
	rows, err := r.DB.Query("SELECT uye_id, rol, unvan, ad, soyad, bolum, telefon, eposta, izu_uyesi, aktif_mi FROM uye")
	if err != nil {
		log.Printf("GetAllUsers hatası: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u models.Uye
		if err := rows.Scan(&u.UyeID, &u.Rol, &u.Unvan, &u.Ad, &u.Soyad, &u.Bolum, &u.Telefon, &u.Eposta, &u.IzuUyesi, &u.AktifMi); err == nil {
			users = append(users, u)
		}
	}
	return users, nil
}

// GetAllProjects, sistemdeki tüm projeleri döner.
func (r *AdminRepository) GetAllProjects() ([]models.Proje, error) {
	var projes []models.Proje
	rows, err := r.DB.Query(`
		SELECT p.proje_id, p.baslik_tr, COALESCE(pd.durum_adi, 'taslak'),
		       COALESCE(pbt.bap_turu, 'Münferit'), p.olusturma_tarihi
		FROM proje p
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
	`)
	if err != nil {
		log.Printf("GetAllProjects hatası: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Proje
		if err := rows.Scan(&p.ProjeID, &p.BaslikTr, &p.DurumAdi, &p.BapTuru, &p.OlusturmaTarihi); err == nil {
			projes = append(projes, p)
		}
	}
	return projes, nil
}

// UpdateUserRole, bir kullanıcının rolünü günceller.
func (r *AdminRepository) UpdateUserRole(uyeID int, rolAdi string) error {
	// Uye tablosundaki rol alanını doğrudan güncelle
	_, err := r.DB.Exec(`UPDATE uye SET rol = $1 WHERE uye_id = $2`, rolAdi, uyeID)
	if err != nil {
		log.Printf("UpdateUserRole hatası: %v", err)
	}
	return err
}

// UpdateUserStatus, bir kullanıcının aktiflik durumunu günceller.
func (r *AdminRepository) UpdateUserStatus(uyeID int, isActive bool) error {
	_, err := r.DB.Exec("UPDATE uye SET aktif_mi = $1 WHERE uye_id = $2", isActive, uyeID)
	if err != nil {
		log.Printf("UpdateUserStatus hatası: %v", err)
	}
	return err
}

// UpdateProjectStatus, projenin genel statüsünü günceller.
func (r *AdminRepository) UpdateProjectStatus(projeID int, durum string) error {
	// Durum adına göre durum_id bul ve güncelle
	_, err := r.DB.Exec(`
		UPDATE proje SET durum_id = (SELECT durum_id FROM proje_durum WHERE durum_adi = $1)
		WHERE proje_id = $2
	`, durum, projeID)
	if err != nil {
		log.Printf("UpdateProjectStatus hatası: %v", err)
	}
	return err
}

// GetProjectStats, genel sistemdeki tüm proje istatistiklerini hesaplayıp döner.
func (r *AdminRepository) GetProjectStats() (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	err := r.DB.QueryRow(`
		SELECT COUNT(*) FROM proje p
		INNER JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pd.durum_adi = 'onaylandi'
	`).Scan(&stats.AktifProje)
	if err != nil {
		return nil, err
	}

	err = r.DB.QueryRow(`
		SELECT COUNT(*) FROM proje p
		INNER JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pd.durum_adi = 'incelemede'
	`).Scan(&stats.OnayBekleyen)
	if err != nil {
		return nil, err
	}

	err = r.DB.QueryRow(`
		SELECT COUNT(*) FROM proje p
		INNER JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pd.durum_adi = 'tamamlandi'
	`).Scan(&stats.Tamamlanan)
	if err != nil {
		return nil, err
	}

	err = r.DB.QueryRow("SELECT COALESCE(SUM(toplam_butce), 0) FROM proje").Scan(&stats.ToplamButce)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// ReviewDetail, Admin sayfasında hakem yorumlarını göstermek için veri yapısı.
type ReviewDetail struct {
	DegerlendirmeID int    `json:"degerlendirme_id"`
	HakemAdSoyad    string `json:"hakem_ad_soyad"`
	Puan            int    `json:"puan"`
	Yorum           string `json:"yorum"`
	Durum           string `json:"durum"`
}

// ProjectDetail, Admin'in göreceği proje detay haritası
type ProjectDetail struct {
	Proje      models.Proje   `json:"proje"`
	Butceler   []models.Butce `json:"butceler"`
	Reviews    []ReviewDetail `json:"reviews"`
	YurutucuAd string         `json:"yurutucu_ad"`
}

// GetProjectDetailsForAdmin, bir projenin detaylı analizini döner.
func (r *AdminRepository) GetProjectDetailsForAdmin(projeID int) (*ProjectDetail, error) {
	detail := &ProjectDetail{}

	// 1. Proje Temel Bilgisi
	err := r.DB.QueryRow(`
		SELECT p.proje_id, p.baslik_tr, p.baslik_en, COALESCE(pbt.bap_turu, 'Münferit'),
		       COALESCE(pd.durum_adi, 'taslak'), p.toplam_butce, p.olusturma_tarihi
		FROM proje p
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		WHERE p.proje_id = $1
	`, projeID).Scan(
		&detail.Proje.ProjeID, &detail.Proje.BaslikTr, &detail.Proje.BaslikEn,
		&detail.Proje.BapTuru, &detail.Proje.DurumAdi,
		&detail.Proje.ToplamButce, &detail.Proje.OlusturmaTarihi,
	)
	if err != nil {
		return nil, err
	}

	// 2. Yürütücü Bilgisi
	r.DB.QueryRow(`
		SELECT COALESCE(u.ad || ' ' || u.soyad, 'Bilinmiyor')
		FROM proje_takim pt
		INNER JOIN uye u ON u.uye_id = pt.uye_id
		INNER JOIN proje_rol_tanimlama prt ON pt.proje_rol_id = prt.rol_id
		WHERE pt.proje_id = $1 AND prt.proje_rol = 'Yürütücü'
		LIMIT 1
	`, projeID).Scan(&detail.YurutucuAd)

	// 3. Bütçe Bilgileri
	var butceler []models.Butce
	rowsButce, err := r.DB.Query(`
		SELECT kalem_id, COALESCE(bk.kategori_adi, ''), aciklama, birim_fiyat, toplam_fiyat
		FROM butce b
		LEFT JOIN butce_kategori bk ON b.kategori_id = bk.kategori_id
		WHERE b.proje_id = $1
	`, projeID)
	if err == nil {
		defer rowsButce.Close()
		for rowsButce.Next() {
			var b models.Butce
			if err := rowsButce.Scan(&b.KalemID, &b.KategoriAdi, &b.Aciklama, &b.BirimFiyat, &b.ToplamFiyat); err == nil {
				butceler = append(butceler, b)
			}
		}
	}
	detail.Butceler = butceler

	// 4. Hakem Değerlendirmeleri
	var reviews []ReviewDetail
	rowsR, err := r.DB.Query(`
		SELECT d.degerlendirme_id, COALESCE(u.ad || ' ' || u.soyad, 'Silinmiş Kullanıcı'),
		       COALESCE(d.puan, 0), COALESCE(d.yorum, ''), d.durum
		FROM proje_degerlendirmeleri d
		JOIN uye u ON u.uye_id = d.hakem_id
		WHERE d.proje_id = $1
	`, projeID)
	if err == nil {
		defer rowsR.Close()
		for rowsR.Next() {
			var rd ReviewDetail
			if err := rowsR.Scan(&rd.DegerlendirmeID, &rd.HakemAdSoyad, &rd.Puan, &rd.Yorum, &rd.Durum); err == nil {
				reviews = append(reviews, rd)
			}
		}
	}
	detail.Reviews = reviews

	return detail, nil
}
