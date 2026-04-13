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

// CreateProje veritabanına yeni bir proje ekler ve oluşturan kullanıcıyı yürütücü olarak atar.
func (r *ProjeRepository) CreateProje(uyeID int, p *models.Proje) error {
	// 1. Projeyi ekle ve ID'sini al
	query := `
		INSERT INTO proje (baslik_tr, tur, baslangic_tarihi, proje_suresi, ozet_tr, durum)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING proje_id
	`
	// Durumu varsayılan olarak 'incelemede' yapıyoruz.
	durum := "incelemede"
	if p.Durum != "" {
		durum = p.Durum
	}

	// Tarihleri dönüştürmemiz gerekebilir ama form string olarak yollayacak. PostgreSQL DATE kabul eder, string 'YYYY-MM-DD' işe yarar.
	err := r.DB.QueryRow(query, p.BaslikTr, p.Tur, p.BaslangicTarihi, p.ProjeSuresi, p.OzetTr, durum).Scan(&p.ProjeID)
	if err != nil {
		return err
	}

	// 2. Proje üyesi olarak ekleyen kişiyi yürütücü atayalım
	uyeQuery := `
		INSERT INTO proje_uyeleri (proje_id, uye_id, rol)
		VALUES ($1, $2, $3)
	`
	_, err = r.DB.Exec(uyeQuery, p.ProjeID, uyeID, "Yürütücü")
	return err
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

// GetProjectsByUyeIDForProfil fonksiyonu, belirli bir üyenin tüm projelerini profil formatında getirir.
// Proje adı, tür, durum ve üyenin projedeki rolü sorgulanır.
func (r *ProjeRepository) GetProjectsByUyeIDForProfil(uyeID int) ([]models.ProfilProjeBilgisi, error) {
	query := `
		SELECT p.proje_id,
		       COALESCE(p.baslik_tr, 'Başlıksız Proje'),
		       COALESCE(p.tur, 'Münferit'),
		       COALESCE(p.durum, 'taslak'),
		       COALESCE(pu.rol, 'Araştırmacı')
		FROM proje p
		INNER JOIN proje_uyeleri pu ON p.proje_id = pu.proje_id
		WHERE pu.uye_id = $1
		ORDER BY p.created_at DESC
	`

	rows, err := r.DB.Query(query, uyeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projeler []models.ProfilProjeBilgisi
	for rows.Next() {
		var p models.ProfilProjeBilgisi
		if err := rows.Scan(&p.ProjeID, &p.BaslikTr, &p.Tur, &p.Durum, &p.UyeRol); err != nil {
			return nil, err
		}
		projeler = append(projeler, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projeler, nil
}

// GetProjeUyeleri projenin kayıtlı üyelerini getirir.
func (r *ProjeRepository) GetProjeUyeleri(projeID int) ([]models.ProjeUye, error) {
	query := `
		SELECT u.uye_id, u.ad || ' ' || u.soyad AS ad_tumu, u.role_id, pu.rol
		FROM proje_uyeleri pu
		INNER JOIN uye u ON pu.uye_id = u.uye_id
		WHERE pu.proje_id = $1
	`
	rows, err := r.DB.Query(query, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uyeler []models.ProjeUye
	for rows.Next() {
		var u models.ProjeUye
		if err := rows.Scan(&u.UyeID, &u.AdTumu, &u.RoleID, &u.Rol); err != nil {
			return nil, err
		}
		uyeler = append(uyeler, u)
	}
	return uyeler, nil
}

// GetProjeByID projeyi ID'sine göre getirir
func (r *ProjeRepository) GetProjeByID(projeID int) (*models.Proje, error) {
	query := `
		SELECT proje_id, baslik_tr, baslik_en, baslangic_tarihi, proje_suresi, toplam_tutar, etik_kurul,
		       ozet_tr, ozet_en, amac_ve_hedef, ozgunluk, metodoloji, risk_yonetimi, proje_ciktilari,
		       aktivite_bilgisi, aktivite_fizibilitesi, durum, tur, created_at, updated_at
		FROM proje WHERE proje_id = $1
	`
	p := &models.Proje{}
	err := r.DB.QueryRow(query, projeID).Scan(
		&p.ProjeID, &p.BaslikTr, &p.BaslikEn, &p.BaslangicTarihi, &p.ProjeSuresi, &p.ToplamTutar, &p.EtikKurul,
		&p.OzetTr, &p.OzetEn, &p.AmacVeHedef, &p.Ozgunluk, &p.Metodoloji, &p.RiskYonetimi, &p.ProjeCiktilari,
		&p.AktiviteBilgisi, &p.AktiviteFizibilitesi, &p.Durum, &p.Tur, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// UpdateProje mevcut bir projenin alanlarını günceller
func (r *ProjeRepository) UpdateProje(p *models.Proje) error {
	query := `
		UPDATE proje SET 
		    baslik_tr=$1, baslik_en=$2, baslangic_tarihi=$3, proje_suresi=$4, toplam_tutar=$5,
		    etik_kurul=$6, ozet_tr=$7, ozet_en=$8, amac_ve_hedef=$9, ozgunluk=$10, metodoloji=$11,
		    risk_yonetimi=$12, proje_ciktilari=$13, aktivite_bilgisi=$14, aktivite_fizibilitesi=$15,
		    durum=$16, tur=$17, updated_at=CURRENT_TIMESTAMP
		WHERE proje_id=$18
	`
	_, err := r.DB.Exec(query,
		p.BaslikTr, p.BaslikEn, p.BaslangicTarihi, p.ProjeSuresi, p.ToplamTutar,
		p.EtikKurul, p.OzetTr, p.OzetEn, p.AmacVeHedef, p.Ozgunluk, p.Metodoloji,
		p.RiskYonetimi, p.ProjeCiktilari, p.AktiviteBilgisi, p.AktiviteFizibilitesi,
		p.Durum, p.Tur, p.ProjeID,
	)
	return err
}
