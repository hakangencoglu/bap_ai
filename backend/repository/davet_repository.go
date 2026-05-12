package repository

import (
	"database/sql"
	"log"

	"bap_ai/backend/models"
)

// DavetRepository, proje davet işlemleri için veritabanı erişimini sağlar
type DavetRepository struct {
	DB *sql.DB
}

// NewDavetRepository, yeni bir DavetRepository örneği oluşturur
func NewDavetRepository(db *sql.DB) *DavetRepository {
	return &DavetRepository{DB: db}
}

// GetBekleyenDavetler, kullanıcının bekleyen proje davetlerini getirir
func (r *DavetRepository) GetBekleyenDavetler(uyeID int) ([]models.ProjeDavet, error) {
	query := `
		SELECT pt.proje_id, COALESCE(p.baslik_tr, ''), COALESCE(pbt.bap_turu, 'Münferit'),
		       COALESCE(davet_eden.ad || ' ' || davet_eden.soyad, 'Bilinmiyor'),
		       COALESCE(prt.proje_rol, 'Araştırmacı'), pt.davet_durumu,
		       TO_CHAR(p.olusturma_tarihi, 'DD.MM.YYYY')
		FROM proje_takim pt
		INNER JOIN proje p ON pt.proje_id = p.proje_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN proje_rol_tanimlama prt ON pt.proje_rol_id = prt.rol_id
		LEFT JOIN uye davet_eden ON p.koordinator_id = davet_eden.uye_id
		WHERE pt.uye_id = $1 AND pt.davet_durumu = 'beklemede'
		ORDER BY p.olusturma_tarihi DESC
	`

	rows, err := r.DB.Query(query, uyeID)
	if err != nil {
		log.Printf("GetBekleyenDavetler hatası: %v", err)
		return nil, err
	}
	defer rows.Close()

	var davetler []models.ProjeDavet
	for rows.Next() {
		var d models.ProjeDavet
		if err := rows.Scan(&d.ProjeID, &d.BaslikTr, &d.BapTuru, &d.DavetEdenAd,
			&d.ProjeRol, &d.DavetDurumu, &d.DavetTarihi); err == nil {
			davetler = append(davetler, d)
		}
	}
	return davetler, nil
}

// UpdateDavetDurumu, belirli bir projedeki üyenin davet durumunu günceller
func (r *DavetRepository) UpdateDavetDurumu(projeID int, uyeID int, durum string) error {
	_, err := r.DB.Exec(`
		UPDATE proje_takim SET davet_durumu = $1
		WHERE proje_id = $2 AND uye_id = $3
	`, durum, projeID, uyeID)
	if err != nil {
		log.Printf("UpdateDavetDurumu hatası: %v", err)
	}
	return err
}

// RemoveTeamMember, red durumunda üyeyi proje takımından çıkarır
func (r *DavetRepository) RemoveTeamMember(projeID int, uyeID int) error {
	_, err := r.DB.Exec(`
		DELETE FROM proje_takim WHERE proje_id = $1 AND uye_id = $2
	`, projeID, uyeID)
	if err != nil {
		log.Printf("RemoveTeamMember hatası: %v", err)
	}
	return err
}

// CheckYurutucuOnay, projenin yürütücüsünün daveti kabul edip etmediğini kontrol eder
// true dönerse yürütücü kabul etmiştir
func (r *DavetRepository) CheckYurutucuOnay(projeID int) (bool, error) {
	var davetDurumu string
	err := r.DB.QueryRow(`
		SELECT COALESCE(pt.davet_durumu, 'beklemede')
		FROM proje_takim pt
		INNER JOIN proje_rol_tanimlama prt ON pt.proje_rol_id = prt.rol_id
		WHERE pt.proje_id = $1 AND prt.proje_rol = 'Yürütücü'
		LIMIT 1
	`, projeID).Scan(&davetDurumu)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		log.Printf("CheckYurutucuOnay hatası: %v", err)
		return false, err
	}

	return davetDurumu == "kabul", nil
}

// UpdateProjeDurumByID, projenin durum_id'sini doğrudan günceller
func (r *DavetRepository) UpdateProjeDurumByID(projeID int, durumID int) error {
	_, err := r.DB.Exec(`UPDATE proje SET durum_id = $1 WHERE proje_id = $2`, durumID, projeID)
	if err != nil {
		log.Printf("UpdateProjeDurumByID hatası: %v", err)
	}
	return err
}

// AddTeamMemberWithInvite, proje takımına davet durumu "beklemede" ile yeni üye ekler
func (r *DavetRepository) AddTeamMemberWithInvite(projeID int, uyeID int, rolID int) error {
	_, err := r.DB.Exec(`
		INSERT INTO proje_takim (proje_id, uye_id, proje_rol_id, davet_durumu)
		VALUES ($1, $2, $3, 'beklemede')
		ON CONFLICT (proje_id, uye_id) DO NOTHING
	`, projeID, uyeID, rolID)
	if err != nil {
		log.Printf("AddTeamMemberWithInvite hatası: %v", err)
	}
	return err
}
