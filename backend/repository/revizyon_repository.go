package repository

import (
	"database/sql"

	"bap_ai/backend/models"
)

// RevizyonRepository yapısı, revizyon tablosuna erişim sorgularını barındırır.
type RevizyonRepository struct {
	DB *sql.DB
}

// NewRevizyonRepository fonksiyonu, yeni bir RevizyonRepository döner.
func NewRevizyonRepository(db *sql.DB) *RevizyonRepository {
	return &RevizyonRepository{DB: db}
}

// CreateRevizyon veritabanına yeni bir revizyon kaydı ekler
func (r *RevizyonRepository) CreateRevizyon(rev *models.Revizyon) error {
	// Türkçe Yorum: Revizyon oluşturulurken revizyon talep edilen bölüm bilgisini de kaydediyoruz.
	query := `
		INSERT INTO proje_revizyon (proje_id, olusturan_kisi_id, atanan_kisi_id, aciklama, durum, revizyon_bolum)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING revizyon_id, olusturma_tarihi, guncelleme_tarihi
	`
	durum := "Bekliyor"
	if rev.Durum != "" {
		durum = rev.Durum
	}

	err := r.DB.QueryRow(query, rev.ProjeID, rev.OlusturanKisiID, rev.AtananKisiID, rev.Aciklama, durum, rev.RevizyonBolum).Scan(
		&rev.RevizyonID, &rev.OlusturmaTarihi, &rev.GuncellemeTarihi,
	)
	return err
}

// GetAktifRevizyon projeye ait aktif ("Bekliyor" durumundaki) revizyonu çeker
func (r *RevizyonRepository) GetAktifRevizyon(projeID int) (*models.Revizyon, error) {
	// Türkçe Yorum: Aktif revizyon kaydını ararken revizyon_bolum alanını da sorguya dahil ediyoruz.
	query := `
		SELECT revizyon_id, proje_id, olusturan_kisi_id, atanan_kisi_id, aciklama, durum, revizyon_bolum,
		       olusturma_tarihi, guncelleme_tarihi
		FROM proje_revizyon
		WHERE proje_id = $1 AND durum = 'Bekliyor'
		ORDER BY olusturma_tarihi DESC LIMIT 1
	`
	rev := &models.Revizyon{}
	err := r.DB.QueryRow(query, projeID).Scan(
		&rev.RevizyonID, &rev.ProjeID, &rev.OlusturanKisiID, &rev.AtananKisiID,
		&rev.Aciklama, &rev.Durum, &rev.RevizyonBolum, &rev.OlusturmaTarihi, &rev.GuncellemeTarihi,
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
		UPDATE proje_revizyon SET durum = 'Tamamlandı', guncelleme_tarihi = CURRENT_TIMESTAMP
		WHERE proje_id = $1 AND durum = 'Bekliyor'
	`
	_, err := r.DB.Exec(query, projeID)
	return err
}
