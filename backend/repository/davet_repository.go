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
		SELECT pt.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'), ''), COALESCE(pbt.bap_turu, 'Münferit'),
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
		if err := rows.Scan(&d.ProjeID, &d.ProjeKodu, &d.BaslikTr, &d.BapTuru, &d.DavetEdenAd,
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

// AddTeamMemberWithInvite, proje takımına yeni üye ekler (yürütücü ise doğrudan "kabul", diğer üyeler için "beklemede" olarak).
// Türkçe Yorum: Yürütücü seçilen kişi için davet beklemeden doğrudan kabul edilmiş sayılır ve projenin koordinatörü güncellenir.
func (r *DavetRepository) AddTeamMemberWithInvite(projeID int, uyeID int, rolID int) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if rolID == 1 {
		// 1. Projedeki mevcut yürütücünün rolünü araştırmacıya (2) düşür
		_, err = tx.Exec(`
			UPDATE proje_takim 
			SET proje_rol_id = 2 
			WHERE proje_id = $1 AND proje_rol_id = 1
		`, projeID)
		if err != nil {
			log.Printf("AddTeamMemberWithInvite eski yürütücü düşürme hatası: %v", err)
			return err
		}

		// 2. Yeni yürütücüyü 'kabul' durumunda ekle / güncelle
		_, err = tx.Exec(`
			INSERT INTO proje_takim (proje_id, uye_id, proje_rol_id, davet_durumu)
			VALUES ($1, $2, 1, 'kabul')
			ON CONFLICT (proje_id, uye_id) 
			DO UPDATE SET proje_rol_id = 1, davet_durumu = 'kabul'
		`, projeID, uyeID)
		if err != nil {
			log.Printf("AddTeamMemberWithInvite yeni yürütücü ekleme hatası: %v", err)
			return err
		}

		// 3. Proje tablosundaki koordinator_id'yi güncelle
		_, err = tx.Exec(`
			UPDATE proje 
			SET koordinator_id = $1, guncelleme_tarihi = CURRENT_TIMESTAMP 
			WHERE proje_id = $2
		`, uyeID, projeID)
		if err != nil {
			log.Printf("AddTeamMemberWithInvite proje tablosu yürütücü güncelleme hatası: %v", err)
			return err
		}
	} else {
		// Yürütücü dışındaki üyeleri 'beklemede' davet durumu ile ekle
		_, err = tx.Exec(`
			INSERT INTO proje_takim (proje_id, uye_id, proje_rol_id, davet_durumu)
			VALUES ($1, $2, $3, 'beklemede')
			ON CONFLICT (proje_id, uye_id) DO NOTHING
		`, projeID, uyeID, rolID)
		if err != nil {
			log.Printf("AddTeamMemberWithInvite üye ekleme hatası: %v", err)
			return err
		}
	}

	return tx.Commit()
}

// SetYurutucu, akademisyen daveti kabul ettiğinde projenin koordinatör ID'sini günceller
func (r *DavetRepository) SetYurutucu(projeID int, uyeID int) error {
	_, err := r.DB.Exec(`UPDATE proje SET koordinator_id = $1, guncelleme_tarihi = CURRENT_TIMESTAMP WHERE proje_id = $2`, uyeID, projeID)
	if err != nil {
		log.Printf("SetYurutucu hatası: %v", err)
	}
	return err
}

// GetUyeProjeRol, belirli bir üyenin projedeki rol adını döner (Yürütücü, Araştırmacı vb.)
func (r *DavetRepository) GetUyeProjeRol(projeID int, uyeID int) (string, error) {
	var rolAdi string
	err := r.DB.QueryRow(`
		SELECT COALESCE(prt.proje_rol, 'Araştırmacı')
		FROM proje_takim pt
		LEFT JOIN proje_rol_tanimlama prt ON pt.proje_rol_id = prt.rol_id
		WHERE pt.proje_id = $1 AND pt.uye_id = $2
	`, projeID, uyeID).Scan(&rolAdi)
	if err != nil {
		log.Printf("GetUyeProjeRol hatası: %v", err)
		return "", err
	}
	return rolAdi, nil
}

// UpdateTeamMemberRol, bir ekip üyesinin proje rolünü günceller
func (r *DavetRepository) UpdateTeamMemberRol(projeID int, uyeID int, rolID int) error {
	_, err := r.DB.Exec(`
		UPDATE proje_takim SET proje_rol_id = $1
		WHERE proje_id = $2 AND uye_id = $3
	`, rolID, projeID, uyeID)
	if err != nil {
		log.Printf("UpdateTeamMemberRol hatası: %v", err)
	}
	return err
}

// GetProjeOlusturanID, projenin ilk oluşturan kişisinin ID'sini döner (koordinator_id üzerinden)
func (r *DavetRepository) GetProjeOlusturanID(projeID int) (int, error) {
	var koordinatorID int
	err := r.DB.QueryRow(`SELECT COALESCE(koordinator_id, 0) FROM proje WHERE proje_id = $1`, projeID).Scan(&koordinatorID)
	if err != nil {
		log.Printf("GetProjeOlusturanID hatası: %v", err)
	}
	return koordinatorID, err
}

// GetUyeSistemRol, bir üyenin sistemdeki rollerini döner (akademisyen, ogrenci vb.)
func (r *DavetRepository) GetUyeSistemRol(uyeID int) (string, error) {
	var rol string
	query := `
		SELECT COALESCE((
		    SELECT string_agg(srt.rol_adi, ',') 
		    FROM sistem_rol sr 
		    INNER JOIN sistem_rol_tanimlama srt ON sr.sistem_rol_id = srt.rol_id 
		    WHERE sr.uye_id = u.uye_id
		), d.rol, u.rol, '')
		FROM uye u
		LEFT JOIN uye_detay d ON u.uye_id = d.uye_id
		WHERE u.uye_id = $1
	`
	err := r.DB.QueryRow(query, uyeID).Scan(&rol)
	if err != nil {
		log.Printf("GetUyeSistemRol hatası: %v", err)
	}
	return rol, err
}
