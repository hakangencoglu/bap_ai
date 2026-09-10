package repository

import (
	"database/sql"
	"bap_ai/backend/models"
)

// AuditRepository, sistem işlem ve denetim loglarının veritabanı işlemlerini yönetir.
type AuditRepository struct {
	DB *sql.DB
}

// NewAuditRepository, yeni bir AuditRepository örneği oluşturur.
func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{DB: db}
}

// LogKaydet, kullanıcı hareketlerini sistem_islem_log tablosuna kaydeder.
// Türkçe Yorum: Asenkron veya senkron çağrılarak her işlemin denetim kaydını tutar.
func (r *AuditRepository) LogKaydet(log *models.AuditLog) error {
	if r == nil || r.DB == nil || log == nil {
		return nil
	}

	var uyeIDInterface interface{}
	if log.UyeID != nil && *log.UyeID > 0 {
		uyeIDInterface = *log.UyeID
	} else {
		uyeIDInterface = nil
	}

	_, err := r.DB.Exec(`
		INSERT INTO sistem_islem_log (uye_id, email, rol, islem_turu, endpoint, ip_adresi, durum_kodu, aciklama)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, uyeIDInterface, log.Email, log.Rol, log.IslemTuru, log.Endpoint, log.IPAdresi, log.DurumKodu, log.Aciklama)
	return err
}

// LogListele, sistemdeki son denetim loglarını listeler (Admin paneli için).
func (r *AuditRepository) LogListele(limit int) ([]models.AuditLog, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.DB.Query(`
		SELECT l.log_id, l.uye_id, COALESCE(l.email, ''), COALESCE(l.rol, ''),
		       l.islem_turu, l.endpoint, COALESCE(l.ip_adresi, ''), l.durum_kodu,
		       COALESCE(l.aciklama, ''), l.tarih
		FROM sistem_islem_log l
		ORDER BY l.tarih DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.AuditLog
	for rows.Next() {
		var item models.AuditLog
		var uyeID sql.NullInt64
		if err := rows.Scan(&item.LogID, &uyeID, &item.Email, &item.Rol,
			&item.IslemTuru, &item.Endpoint, &item.IPAdresi, &item.DurumKodu,
			&item.Aciklama, &item.Tarih); err != nil {
			continue
		}
		if uyeID.Valid {
			idVal := int(uyeID.Int64)
			item.UyeID = &idVal
		}
		list = append(list, item)
	}
	return list, nil
}
