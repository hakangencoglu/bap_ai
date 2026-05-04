package repository

import (
	"database/sql"
	"fmt"

	"bap_ai/backend/models"
)

// HakemRepository, hakem değerlendirme işlemlerini gerçekleştirir.
type HakemRepository struct {
	DB *sql.DB
}

// NewHakemRepository fonksiyonu yeni bir HakemRepository döner.
func NewHakemRepository(db *sql.DB) *HakemRepository {
	return &HakemRepository{DB: db}
}

// AssignRandomHakem, belirtilen sayıda rastgele hakemi projeye atar.
func (r *HakemRepository) AssignRandomHakem(projeID int, count int) error {
	// Doğrudan uye tablosundaki rol alanından hakem kullanıcıları sorgulanır
	queryRandomHakem := `
		SELECT uye_id FROM uye
		WHERE rol = 'hakem' AND aktif_mi = true
		ORDER BY RANDOM() LIMIT $1
	`
	rows, err := r.DB.Query(queryRandomHakem, count)
	if err != nil {
		return err
	}
	defer rows.Close()

	var hakemIDs []int
	for rows.Next() {
		var uid int
		if err := rows.Scan(&uid); err != nil {
			return err
		}
		hakemIDs = append(hakemIDs, uid)
	}

	if len(hakemIDs) == 0 {
		return fmt.Errorf("atanacak aktif hakem bulunamadı")
	}

	for _, hid := range hakemIDs {
		insertQuery := `
			INSERT INTO proje_degerlendirmeleri (proje_id, hakem_id, durum)
			VALUES ($1, $2, 'Bekliyor')
			ON CONFLICT (proje_id, hakem_id) DO NOTHING
		`
		r.DB.Exec(insertQuery, projeID, hid)
	}
	return nil
}

// GetProjelerByHakemID, bir hakeme atanmış tüm projeleri getirir
func (r *HakemRepository) GetProjelerByHakemID(hakemID int) ([]models.HakemProjeOzet, error) {
	query := `
		SELECT p.proje_id, COALESCE(p.baslik_tr, 'Başlıksız Proje'),
		       COALESCE(pbt.bap_turu, 'Münferit'), COALESCE(pd.durum_adi, 'taslak'),
		       pdeg.durum, pdeg.puan, TO_CHAR(pdeg.olusturma_tarihi, 'DD.MM.YYYY')
		FROM proje p
		INNER JOIN proje_degerlendirmeleri pdeg ON p.proje_id = pdeg.proje_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pdeg.hakem_id = $1
		ORDER BY pdeg.olusturma_tarihi DESC
	`
	rows, err := r.DB.Query(query, hakemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projeler []models.HakemProjeOzet
	for rows.Next() {
		var p models.HakemProjeOzet
		if err := rows.Scan(&p.ProjeID, &p.BaslikTr, &p.BapTuru, &p.DurumAdi, &p.HakemDurum, &p.Puan, &p.Tarih); err != nil {
			return nil, err
		}
		projeler = append(projeler, p)
	}
	return projeler, nil
}

// SubmitDegerlendirme, hakemin yaptığı değerlendirmeyi DB'ye kaydeder
func (r *HakemRepository) SubmitDegerlendirme(hakemID int, req models.DegerlendirmeRequest) error {
	query := `
		UPDATE proje_degerlendirmeleri
		SET puan = $1, yorum = $2, durum = $3, guncelleme_tarihi = CURRENT_TIMESTAMP
		WHERE proje_id = $4 AND hakem_id = $5
	`
	res, err := r.DB.Exec(query, req.Puan, req.Yorum, req.Durum, req.ProjeID, hakemID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("değerlendirme bulunamadı veya güncellenemedi")
	}
	return nil
}

// GetAllDegerlendirmeByProjeID, bir projenin tüm hakem değerlendirmelerini getirir.
func (r *HakemRepository) GetAllDegerlendirmeByProjeID(projeID int) ([]models.ProjeDegerlendirme, error) {
	query := `
		SELECT degerlendirme_id, proje_id, hakem_id, COALESCE(puan, 0), COALESCE(yorum, ''), durum
		FROM proje_degerlendirmeleri WHERE proje_id = $1
	`
	rows, err := r.DB.Query(query, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var degs []models.ProjeDegerlendirme
	for rows.Next() {
		var d models.ProjeDegerlendirme
		if err := rows.Scan(&d.DegerlendirmeID, &d.ProjeID, &d.HakemID, &d.Puan, &d.Yorum, &d.Durum); err != nil {
			return nil, err
		}
		degs = append(degs, d)
	}
	return degs, nil
}
