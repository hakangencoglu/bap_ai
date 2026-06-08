package repository

import (
	"database/sql"
	"fmt"

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

// CreateProje veritabanına yeni bir proje ekler ve oluşturan kullanıcıyı uygun rolle atar.
// Öğrenci oluşturuyorsa Araştırmacı (proje_rol_id=2) olarak, akademisyen oluşturuyorsa Yürütücü (proje_rol_id=1) olarak atanır.
func (r *ProjeRepository) CreateProje(uyeID int, p *models.Proje, uyeRol string) error {
	// 1. Projeyi ekle ve ID'sini al
	// durum_id=1 (taslak) varsayılan olarak atanır — yürütücü kabul ettikten sonra incelemeye geçer
	query := `
		INSERT INTO proje (baslik_tr, bap_turu_id, sure_ay, toplam_butce, koordinator_id, durum_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING proje_id
	`
	// Varsayılan durum: taslak (durum_id=1)
	durumID := 1
	if p.DurumID != nil {
		durumID = *p.DurumID
	}

	err := r.DB.QueryRow(query, p.BaslikTr, p.BapTuruID, p.SureAy, p.ToplamButce, uyeID, durumID).Scan(&p.ProjeID)
	if err != nil {
		return err
	}

	// 2. Proje takımına oluşturan kişiyi uygun rolle ekle
	// Öğrenci → Araştırmacı (proje_rol_id=2), Akademisyen → Yürütücü (proje_rol_id=1)
	projeRolID := 1 // Varsayılan: Yürütücü
	if uyeRol == "ogrenci" {
		projeRolID = 2 // Araştırmacı
	}

	// Projeyi oluşturan kişi otomatik olarak daveti kabul etmiş sayılır
	takimQuery := `
		INSERT INTO proje_takim (proje_id, uye_id, proje_rol_id, davet_durumu)
		VALUES ($1, $2, $3, 'kabul')
	`
	_, err = r.DB.Exec(takimQuery, p.ProjeID, uyeID, projeRolID)
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
		SELECT u.uye_id, u.ad || ' ' || u.soyad AS ad_tumu, COALESCE(d.rol, 'belirsiz'), COALESCE(prt.proje_rol, 'Araştırmacı')
		FROM proje_takim pt
		INNER JOIN uye u ON pt.uye_id = u.uye_id
		LEFT JOIN uye_detay d ON u.uye_id = d.uye_id
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

// UpdateProjeDurum sadece projenin durum_id alanını günceller (onaylama, reddetme vb. için)
func (r *ProjeRepository) UpdateProjeDurum(projeID int, durumID int) error {
	query := `UPDATE proje SET durum_id = $1, guncelleme_tarihi = CURRENT_TIMESTAMP WHERE proje_id = $2`
	_, err := r.DB.Exec(query, durumID, projeID)
	return err
}

// SavePDFPath oluşturulan PDF dosyasının sunucu yolunu proje tablosuna kaydeder
func (r *ProjeRepository) SavePDFPath(projeID int, pdfPath string) error {
	query := `UPDATE proje SET pdf_dosya_yolu = $1, guncelleme_tarihi = CURRENT_TIMESTAMP WHERE proje_id = $2`
	_, err := r.DB.Exec(query, pdfPath, projeID)
	return err
}

// DeleteTaslakProje taslak durumundaki bir projeyi siler.
// Sadece taslak (durum_adi='taslak') ve koordinatörü olan kullanıcı silebilir.
// İlişkili alt tablolar ON DELETE CASCADE ile otomatik temizlenir.
func (r *ProjeRepository) DeleteTaslakProje(projeID int, uyeID int) error {
	// Projenin taslak olduğunu ve kullanıcının koordinatör veya takım üyesi olduğunu doğrula
	var count int
	checkQuery := `
		SELECT COUNT(*) FROM proje p
		INNER JOIN proje_durum pd ON p.durum_id = pd.durum_id
		INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id
		WHERE p.proje_id = $1 AND pt.uye_id = $2 AND pd.durum_adi = 'taslak'
	`
	err := r.DB.QueryRow(checkQuery, projeID, uyeID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("proje bulunamadı, taslak değil veya yetkiniz yok")
	}

	// Projeyi sil (CASCADE ile alt tablolar temizlenir)
	deleteQuery := `DELETE FROM proje WHERE proje_id = $1`
	_, err = r.DB.Exec(deleteQuery, projeID)
	return err
}

// UpdateProjectStatusWithLog projenin durumunu günceller ve bu değişikliği süreç geçmişi tablosuna kaydeder.
// Bu işlem bir transaction (veri tabanı işlemi) kapsamında gerçekleştirilir.
func (r *ProjeRepository) UpdateProjectStatusWithLog(projeID int, islemYapanID int, baslangicDurum, yeniDurum, aciklama string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Yeni durumun durum_id değerini bul
	var durumID int
	err = tx.QueryRow(`SELECT durum_id FROM proje_durum WHERE durum_adi = $1`, yeniDurum).Scan(&durumID)
	if err != nil {
		return fmt.Errorf("hedef durum (%s) bulunamadı: %v", yeniDurum, err)
	}

	// 2. Projenin durumunu güncelle
	_, err = tx.Exec(`UPDATE proje SET durum_id = $1, guncelleme_tarihi = CURRENT_TIMESTAMP WHERE proje_id = $2`, durumID, projeID)
	if err != nil {
		return fmt.Errorf("proje durumu güncellenemedi: %v", err)
	}

	// 3. Süreç geçmişi tablosuna log kaydı ekle
	logQuery := `
		INSERT INTO proje_surec_gecmisi (proje_id, islem_yapan_id, baslangic_durum, hedef_durum, aciklama)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.Exec(logQuery, projeID, islemYapanID, baslangicDurum, yeniDurum, aciklama)
	if err != nil {
		return fmt.Errorf("süreç geçmişi kaydedilemedi: %v", err)
	}

	return tx.Commit()
}

// GetProjeSurecGecmisi projenin geçmiş onay/red/revizyon süreç kayıtlarını getirir.
// Hangi durumdan hangi duruma, kimin tarafından ne zaman ve hangi açıklamayla geçildiğini listeler.
func (r *ProjeRepository) GetProjeSurecGecmisi(projeID int) ([]models.ProjeSurecGecmisi, error) {
	query := `
		SELECT g.gecmis_id, g.proje_id, g.islem_yapan_id, g.baslangic_durum, g.hedef_durum, g.aciklama, g.olusturma_tarihi,
		       COALESCE(u.ad || ' ' || u.soyad, '') as ad_tumu, COALESCE(u.unvan, '') as unvan
		FROM proje_surec_gecmisi g
		LEFT JOIN uye u ON g.islem_yapan_id = u.uye_id
		WHERE g.proje_id = $1
		ORDER BY g.olusturma_tarihi ASC
	`
	rows, err := r.DB.Query(query, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gecmis []models.ProjeSurecGecmisi
	for rows.Next() {
		var g models.ProjeSurecGecmisi
		err := rows.Scan(
			&g.GecmisID, &g.ProjeID, &g.IslemYapanID, &g.BaslangicDurum, &g.HedefDurum, &g.Aciklama, &g.OlusturmaTarihi,
			&g.IslemYapanAdTumu, &g.IslemYapanUnvan,
		)
		if err != nil {
			return nil, err
		}
		gecmis = append(gecmis, g)
	}
	return gecmis, nil
}

// GetProjectsForWorkflow belirli bir aşamadaki (durum_adi) tüm projeleri listeler.
// Bu fonksiyon onay vericilerin (Dekan, Komisyon, TTO) onay bekleyen listeleri için kullanılır.
func (r *ProjeRepository) GetProjectsForWorkflow(rol string, durum string) ([]models.Proje, error) {
	query := `
		SELECT p.proje_id, p.baslik_tr, p.baslik_en, p.sure_ay, p.toplam_butce, p.etik_kurul,
		       p.etik_kurul_no, p.koordinator_id, p.durum_id, p.bap_turu_id,
		       p.olusturma_tarihi, p.guncelleme_tarihi,
		       COALESCE(pd.durum_adi, ''), COALESCE(pbt.bap_turu, ''),
		       COALESCE(u.unvan || ' ' || u.ad || ' ' || u.soyad, u.ad || ' ' || u.soyad, '') as koordinator_ad_soyad
		FROM proje p
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN uye u ON p.koordinator_id = u.uye_id
		WHERE pd.durum_adi = $1
		ORDER BY p.guncelleme_tarihi DESC
	`
	rows, err := r.DB.Query(query, durum)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projeler []models.Proje
	for rows.Next() {
		var p models.Proje
		err := rows.Scan(
			&p.ProjeID, &p.BaslikTr, &p.BaslikEn, &p.SureAy, &p.ToplamButce, &p.EtikKurul,
			&p.EtikKurulNo, &p.KoordinatorID, &p.DurumID, &p.BapTuruID,
			&p.OlusturmaTarihi, &p.GuncellemeTarihi,
			&p.DurumAdi, &p.BapTuru, &p.KoordinatorAdSoyad,
		)
		if err != nil {
			return nil, err
		}
		projeler = append(projeler, p)
	}
	return projeler, nil
}

