package repository

import (
	"database/sql"

	"bap_ai/backend/models"
)

type RevizyonRepository struct {
	DB *sql.DB
}

func NewRevizyonRepository(db *sql.DB) *RevizyonRepository {
	return &RevizyonRepository{DB: db}
}

// CreateRevizyon veritabanına yeni bir revizyon kaydı ekler
func (r *RevizyonRepository) CreateRevizyon(rev *models.Revizyon) error {
	query := `
		INSERT INTO revizyonlar (proje_id, olusturan_kisi_id, atanan_kisi_id, aciklama, durum)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING revizyon_id, created_at, updated_at
	`
	durum := "Bekliyor"
	if rev.Durum != "" {
		durum = rev.Durum
	}

	err := r.DB.QueryRow(query, rev.ProjeID, rev.OlusturanKisiID, rev.AtananKisiID, rev.Aciklama, durum).Scan(&rev.RevizyonID, &rev.CreatedAt, &rev.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

// GetAktifRevizyon projeye ait aktif ("Bekliyor" durumundaki) revizyonu çeker
func (r *RevizyonRepository) GetAktifRevizyon(projeID int) (*models.Revizyon, error) {
	query := `
		SELECT revizyon_id, proje_id, olusturan_kisi_id, atanan_kisi_id, aciklama, durum, created_at, updated_at
		FROM revizyonlar
		WHERE proje_id = $1 AND durum = 'Bekliyor'
		ORDER BY created_at DESC LIMIT 1
	`
	rev := &models.Revizyon{}
	err := r.DB.QueryRow(query, projeID).Scan(
		&rev.RevizyonID, &rev.ProjeID, &rev.OlusturanKisiID, &rev.AtananKisiID,
		&rev.Aciklama, &rev.Durum, &rev.CreatedAt, &rev.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Aktif revizyon yoksa nil dönsün
		}
		return nil, err
	}
	return rev, nil
}

// MarkRevizyonAsDone revizyonu tamamlanmış olarak işaretler
func (r *RevizyonRepository) MarkRevizyonAsDone(projeID int) error {
	query := `
		UPDATE revizyonlar SET durum = 'Tamamlandı', updated_at = CURRENT_TIMESTAMP
		WHERE proje_id = $1 AND durum = 'Bekliyor'
	`
	_, err := r.DB.Exec(query, projeID)
	return err
}
