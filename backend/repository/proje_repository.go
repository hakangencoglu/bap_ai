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
	// durum_id=2 (incelemede) varsayılan olarak atanır
	query := `
		INSERT INTO proje (baslik_tr, bap_turu_id, sure_ay, toplam_butce, koordinator_id, durum_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING proje_id
	`
	// Varsayılan durum: incelemede (durum_id=2)
	durumID := 2
	if p.DurumID != nil {
		durumID = *p.DurumID
	}

	err := r.DB.QueryRow(query, p.BaslikTr, p.BapTuruID, p.SureAy, p.ToplamButce, uyeID, durumID).Scan(&p.ProjeID)
	if err != nil {
		return err
	}

	// 2. Proje takımına oluşturan kişiyi yürütücü olarak ekle (proje_rol_id=1 → Yürütücü)
	takimQuery := `
		INSERT INTO proje_takim (proje_id, uye_id, proje_rol_id)
		VALUES ($1, $2, 1)
	`
	_, err = r.DB.Exec(takimQuery, p.ProjeID, uyeID)
	return err
}

// GetDashboardStatsByUyeID fonksiyonu, belirli bir üyenin proje istatistiklerini getirir.
// proje_takim tablosu üzerinden üyeye ait projelerin durumlarına göre sayılar hesaplanır.
func (r *ProjeRepository) GetDashboardStatsByUyeID(uyeID int) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	// Aktif proje sayısı: durum_adi 'onaylandi' olan projeler
	queryAktif := `
		SELECT COUNT(*) FROM proje p
		INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id
		INNER JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pt.uye_id = $1 AND pd.durum_adi = 'onaylandi'
	`
	err := r.DB.QueryRow(queryAktif, uyeID).Scan(&stats.AktifProje)
	if err != nil {
		return nil, err
	}

	// Onay bekleyen proje sayısı: durum_adi 'incelemede' veya 'taslak'
	queryBekleyen := `
		SELECT COUNT(*) FROM proje p
		INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id
		INNER JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pt.uye_id = $1 AND pd.durum_adi IN ('incelemede', 'taslak')
	`
	err = r.DB.QueryRow(queryBekleyen, uyeID).Scan(&stats.OnayBekleyen)
	if err != nil {
		return nil, err
	}

	// Tamamlanan proje sayısı: durum_adi 'tamamlandi'
	queryTamamlanan := `
		SELECT COUNT(*) FROM proje p
		INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id
		INNER JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pt.uye_id = $1 AND pd.durum_adi = 'tamamlandi'
	`
	err = r.DB.QueryRow(queryTamamlanan, uyeID).Scan(&stats.Tamamlanan)
	if err != nil {
		return nil, err
	}

	// Toplam bütçe: üyeye ait projelerin toplam_butce toplamı
	queryButce := `
		SELECT COALESCE(SUM(p.toplam_butce), 0) FROM proje p
		INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id
		WHERE pt.uye_id = $1
	`
	err = r.DB.QueryRow(queryButce, uyeID).Scan(&stats.ToplamButce)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetRecentProjectsByUyeID fonksiyonu, belirli bir üyenin son 2 proje başvurusunu getirir.
// Tarih sırasına göre en yeniden en eskiye doğru sıralanır.
func (r *ProjeRepository) GetRecentProjectsByUyeID(uyeID int) ([]models.ProjeOzet, error) {
	query := `
		SELECT p.proje_id,
		       COALESCE(p.baslik_tr, 'Başlıksız Proje'),
		       COALESCE(pbt.bap_turu, 'Münferit'),
		       TO_CHAR(p.olusturma_tarihi, 'DD.MM.YYYY'),
		       COALESCE(pd.durum_adi, 'taslak')
		FROM proje p
		INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pt.uye_id = $1
		ORDER BY p.olusturma_tarihi DESC
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
		if err := rows.Scan(&p.ProjeID, &p.BaslikTr, &p.BapTuru, &p.Tarih, &p.DurumAdi); err != nil {
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
func (r *ProjeRepository) GetProjectsByUyeIDForProfil(uyeID int) ([]models.ProfilProjeBilgisi, error) {
	query := `
		SELECT p.proje_id,
		       COALESCE(p.baslik_tr, 'Başlıksız Proje'),
		       COALESCE(pbt.bap_turu, 'Münferit'),
		       COALESCE(pd.durum_adi, 'taslak'),
		       COALESCE(prt.proje_rol, 'Araştırmacı')
		FROM proje p
		INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		LEFT JOIN proje_rol_tanimlama prt ON pt.proje_rol_id = prt.rol_id
		WHERE pt.uye_id = $1
		ORDER BY p.olusturma_tarihi DESC
	`

	rows, err := r.DB.Query(query, uyeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projeler []models.ProfilProjeBilgisi
	for rows.Next() {
		var p models.ProfilProjeBilgisi
		if err := rows.Scan(&p.ProjeID, &p.BaslikTr, &p.BapTuru, &p.DurumAdi, &p.UyeRol); err != nil {
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
		SELECT u.uye_id, u.ad || ' ' || u.soyad AS ad_tumu, COALESCE(u.rol, 'belirsiz'), COALESCE(prt.proje_rol, 'Araştırmacı')
		FROM proje_takim pt
		INNER JOIN uye u ON pt.uye_id = u.uye_id
		LEFT JOIN proje_rol_tanimlama prt ON pt.proje_rol_id = prt.rol_id
		WHERE pt.proje_id = $1
	`
	rows, err := r.DB.Query(query, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uyeler []models.ProjeUye
	for rows.Next() {
		var u models.ProjeUye
		if err := rows.Scan(&u.UyeID, &u.AdTumu, &u.Rol, &u.ProjeRol); err != nil {
			return nil, err
		}
		uyeler = append(uyeler, u)
	}
	return uyeler, nil
}

// GetProjeByID projeyi ID'sine göre getirir
func (r *ProjeRepository) GetProjeByID(projeID int) (*models.Proje, error) {
	query := `
		SELECT p.proje_id, p.baslik_tr, p.baslik_en, p.sure_ay, p.toplam_butce, p.etik_kurul,
		       p.etik_kurul_no, p.koordinator_id, p.durum_id, p.bap_turu_id,
		       p.olusturma_tarihi, p.guncelleme_tarihi,
		       COALESCE(pd.durum_adi, 'taslak'), COALESCE(pbt.bap_turu, 'Münferit')
		FROM proje p
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		WHERE p.proje_id = $1
	`
	p := &models.Proje{}
	err := r.DB.QueryRow(query, projeID).Scan(
		&p.ProjeID, &p.BaslikTr, &p.BaslikEn, &p.SureAy, &p.ToplamButce, &p.EtikKurul,
		&p.EtikKurulNo, &p.KoordinatorID, &p.DurumID, &p.BapTuruID,
		&p.OlusturmaTarihi, &p.GuncellemeTarihi,
		&p.DurumAdi, &p.BapTuru,
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
		    baslik_tr=$1, baslik_en=$2, sure_ay=$3, toplam_butce=$4,
		    etik_kurul=$5, etik_kurul_no=$6, koordinator_id=$7,
		    durum_id=$8, bap_turu_id=$9, guncelleme_tarihi=CURRENT_TIMESTAMP
		WHERE proje_id=$10
	`
	_, err := r.DB.Exec(query,
		p.BaslikTr, p.BaslikEn, p.SureAy, p.ToplamButce,
		p.EtikKurul, p.EtikKurulNo, p.KoordinatorID,
		p.DurumID, p.BapTuruID, p.ProjeID,
	)
	return err
}
