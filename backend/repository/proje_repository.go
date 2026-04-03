package repository

import (
	"database/sql"

	"bap_ai/backend/models"
)

// ProjeRepository yapısı, proje tablosuna erişim sorgularını barındırır.
type ProjeRepository struct {
	DB *sql.DB
}

// NewProjeRepository fonksiyonu, yeni bir ProjeRepository nesnesi döner.
func NewProjeRepository(db *sql.DB) *ProjeRepository {
	return &ProjeRepository{DB: db}
}

// GetDashboardStatsByUyeID fonksiyonu, belirli bir üyenin proje istatistiklerini getirir.
// proje_uyeleri tablosu üzerinden üyeye ait projelerin durumlarına göre sayılar hesaplanır.
func (r *ProjeRepository) GetDashboardStatsByUyeID(uyeID int) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	// Aktif proje sayısı: durum 'onaylandi' olan projeler
	queryAktif := `
		SELECT COUNT(*) FROM proje p
		INNER JOIN proje_uyeleri pu ON p.proje_id = pu.proje_id
		WHERE pu.uye_id = $1 AND p.durum = 'onaylandi'
	`
	err := r.DB.QueryRow(queryAktif, uyeID).Scan(&stats.AktifProje)
	if err != nil {
		return nil, err
	}

	// Onay bekleyen proje sayısı: durum 'incelemede' veya 'taslak' olan projeler
	queryBekleyen := `
		SELECT COUNT(*) FROM proje p
		INNER JOIN proje_uyeleri pu ON p.proje_id = pu.proje_id
		WHERE pu.uye_id = $1 AND p.durum IN ('incelemede', 'taslak')
	`
	err = r.DB.QueryRow(queryBekleyen, uyeID).Scan(&stats.OnayBekleyen)
	if err != nil {
		return nil, err
	}

	// Tamamlanan proje sayısı: durum 'tamamlandi' olan projeler
	queryTamamlanan := `
		SELECT COUNT(*) FROM proje p
		INNER JOIN proje_uyeleri pu ON p.proje_id = pu.proje_id
		WHERE pu.uye_id = $1 AND p.durum = 'tamamlandi'
	`
	err = r.DB.QueryRow(queryTamamlanan, uyeID).Scan(&stats.Tamamlanan)
	if err != nil {
		return nil, err
	}

	// Toplam bütçe: üyeye ait projelerin toplam_tutar toplamı
	queryButce := `
		SELECT COALESCE(SUM(p.toplam_tutar), 0) FROM proje p
		INNER JOIN proje_uyeleri pu ON p.proje_id = pu.proje_id
		WHERE pu.uye_id = $1
	`
	err = r.DB.QueryRow(queryButce, uyeID).Scan(&stats.ToplamButce)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetRecentProjectsByUyeID fonksiyonu, belirli bir üyenin son 5 proje başvurusunu getirir.
// Tarih sırasına göre en yeniden en eskiye doğru sıralanır.
func (r *ProjeRepository) GetRecentProjectsByUyeID(uyeID int) ([]models.ProjeOzet, error) {
	query := `
		SELECT p.proje_id,
		       COALESCE(p.baslik_tr, 'Başlıksız Proje'),
		       COALESCE(p.tur, 'Münferit'),
		       TO_CHAR(p.created_at, 'DD.MM.YYYY'),
		       COALESCE(p.durum, 'taslak')
		FROM proje p
		INNER JOIN proje_uyeleri pu ON p.proje_id = pu.proje_id
		WHERE pu.uye_id = $1
		ORDER BY p.created_at DESC
		LIMIT 2
	`

	rows, err := r.DB.Query(query, uyeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projeler []models.ProjeOzet
	for rows.Next() {
		var p models.ProjeOzet
		if err := rows.Scan(&p.ProjeID, &p.BaslikTr, &p.Tur, &p.Tarih, &p.Durum); err != nil {
			return nil, err
		}
		projeler = append(projeler, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projeler, nil
}
