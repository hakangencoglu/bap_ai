package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

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
// Türkçe Yorum: Bu fonksiyon, proje eklemeyi ve benzersiz proje kodu (proje_kodu) oluşturmayı bir transaction içinde yürütür.
func (r *ProjeRepository) CreateProje(uyeID int, p *models.Proje, uyeRol string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Projeyi ekle ve ID'si ile oluşturulma tarihini al
	// durum_id=1 (taslak) varsayılan olarak atanır
	query := `
		INSERT INTO proje (baslik_tr, bap_turu_id, sure_ay, toplam_butce, koordinator_id, durum_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING proje_id, olusturma_tarihi
	`
	durumID := 1
	if p.DurumID != nil {
		durumID = *p.DurumID
	}

	var olusturmaTarihi time.Time
	err = tx.QueryRow(query, p.BaslikTr, p.BapTuruID, p.SureAy, p.ToplamButce, uyeID, durumID).Scan(&p.ProjeID, &olusturmaTarihi)
	if err != nil {
		return err
	}
	p.OlusturmaTarihi = olusturmaTarihi

	// 2. BAP türü adını ve yılını alarak benzersiz bir Proje Kodu oluştur
	var bapTuru string = "BAP"
	if p.BapTuruID != nil {
		err = tx.QueryRow(`SELECT bap_turu FROM proje_bap_turu WHERE bap_turu_id = $1`, *p.BapTuruID).Scan(&bapTuru)
		if err != nil {
			return err
		}
	}
	cleanBapTuru := strings.ReplaceAll(bapTuru, "-", "")
	year := olusturmaTarihi.Year()

	var maxSeq int
	seqQuery := `
		SELECT COALESCE(MAX(CAST(SUBSTRING(proje_kodu FROM '\d{3}$') AS INTEGER)), 0)
		FROM proje
		WHERE bap_turu_id = $1 AND EXTRACT(YEAR FROM olusturma_tarihi) = $2
	`
	_ = tx.QueryRow(seqQuery, p.BapTuruID, year).Scan(&maxSeq)
	nextSeq := maxSeq + 1
	// Türkçe Yorum: Proje kodunu yil-bapturu-numara (örn: 2026-BAP100-003) formatında oluşturuyoruz
	projeKodu := fmt.Sprintf("%d-%s-%03d", year, cleanBapTuru, nextSeq)

	// Proje kodunu güncelle
	_, err = tx.Exec(`UPDATE proje SET proje_kodu = $1 WHERE proje_id = $2`, projeKodu, p.ProjeID)
	if err != nil {
		return err
	}
	p.ProjeKodu = projeKodu

	// 3. Proje takımına oluşturan kişiyi uygun rolle ekle
	projeRolID := 1 // Varsayılan: Yürütücü
	roles := strings.Split(uyeRol, ",")
	isOnlyOgrenci := true
	for _, r := range roles {
		r = strings.TrimSpace(r)
		if r != "ogrenci" && r != "" {
			isOnlyOgrenci = false
		}
	}
	if isOnlyOgrenci && len(roles) > 0 {
		projeRolID = 2 // Araştırmacı
	}

	// Projeyi oluşturan kişi otomatik olarak daveti kabul etmiş sayılır
	takimQuery := `
		INSERT INTO proje_takim (proje_id, uye_id, proje_rol_id, davet_durumu)
		VALUES ($1, $2, $3, 'kabul')
	`
	_, err = tx.Exec(takimQuery, p.ProjeID, uyeID, projeRolID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetDashboardStatsByUyeID fonksiyonu, belirli bir üyenin proje istatistiklerini getirir.
// proje_takim tablosu üzerinden üyeye ait projelerin durumlarına göre sayılar hesaplanır.
func (r *ProjeRepository) GetDashboardStatsByUyeID(uyeID int) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	// Aktif proje sayısı: durum_adi 'onaylandi', 'tto_aktif' veya 'yururlukte' olan ve daveti kabul edilmiş projeler
	queryAktif := `
		SELECT COUNT(*) FROM proje p
		INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id
		INNER JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pt.uye_id = $1 AND pd.durum_adi IN ('onaylandi', 'tto_aktif', 'yururlukte') AND pt.davet_durumu = 'kabul'
	`
	err := r.DB.QueryRow(queryAktif, uyeID).Scan(&stats.AktifProje)
	if err != nil {
		return nil, err
	}

	// Onay bekleyen proje sayısı: durum_adi 'incelemede' veya 'taslak' ve daveti kabul edilmiş projeler
	queryBekleyen := `
		SELECT COUNT(*) FROM proje p
		INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id
		INNER JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pt.uye_id = $1 AND pd.durum_adi IN ('incelemede', 'taslak') AND pt.davet_durumu = 'kabul'
	`
	err = r.DB.QueryRow(queryBekleyen, uyeID).Scan(&stats.OnayBekleyen)
	if err != nil {
		return nil, err
	}

	// Tamamlanan proje sayısı: durum_adi 'tamamlandi' ve daveti kabul edilmiş projeler
	queryTamamlanan := `
		SELECT COUNT(*) FROM proje p
		INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id
		INNER JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pt.uye_id = $1 AND pd.durum_adi = 'tamamlandi' AND pt.davet_durumu = 'kabul'
	`
	err = r.DB.QueryRow(queryTamamlanan, uyeID).Scan(&stats.Tamamlanan)
	if err != nil {
		return nil, err
	}

	// Toplam bütçe: üyeye ait kabul edilmiş projelerin toplam_butce toplamı
	queryButce := `
		SELECT COALESCE(SUM(p.toplam_butce), 0) FROM proje p
		INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id
		WHERE pt.uye_id = $1 AND pt.davet_durumu = 'kabul'
	`
	err = r.DB.QueryRow(queryButce, uyeID).Scan(&stats.ToplamButce)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetRecentProjectsByUyeID fonksiyonu, belirli bir üyenin son 2 kabul edilmiş proje başvurusunu getirir.
// Tarih sırasına göre en yeniden en eskiye doğru sıralanır.
func (r *ProjeRepository) GetRecentProjectsByUyeID(uyeID int) ([]models.ProjeOzet, error) {
	query := `
		SELECT p.proje_id,
		       COALESCE(p.proje_kodu, ''),
		       COALESCE(p.baslik_tr, 'Başlıksız Proje'),
		       COALESCE(pbt.bap_turu, 'Münferit'),
		       TO_CHAR(p.olusturma_tarihi, 'DD.MM.YYYY'),
		       COALESCE(pd.durum_adi, 'taslak')
		FROM proje p
		INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pt.uye_id = $1 AND pt.davet_durumu = 'kabul'
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
		if err := rows.Scan(&p.ProjeID, &p.ProjeKodu, &p.BaslikTr, &p.BapTuru, &p.Tarih, &p.DurumAdi); err != nil {
			return nil, err
		}
		projeler = append(projeler, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projeler, nil
}

// GetAllProjectsByUyeID fonksiyonu, belirli bir üyenin kabul ettiği tüm proje başvurularını getirir.
// Tarih sırasına göre en yeniden en eskiye doğru sıralanır.
func (r *ProjeRepository) GetAllProjectsByUyeID(uyeID int) ([]models.ProjeOzet, error) {
	query := `
		SELECT p.proje_id,
		       COALESCE(p.proje_kodu, ''),
		       COALESCE(p.baslik_tr, 'Başlıksız Proje'),
		       COALESCE(pbt.bap_turu, 'Münferit'),
		       TO_CHAR(p.olusturma_tarihi, 'DD.MM.YYYY'),
		       COALESCE(pd.durum_adi, 'taslak')
		FROM proje p
		INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pt.uye_id = $1 AND pt.davet_durumu = 'kabul'
		ORDER BY p.olusturma_tarihi DESC
	`

	rows, err := r.DB.Query(query, uyeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projeler []models.ProjeOzet
	for rows.Next() {
		var p models.ProjeOzet
		if err := rows.Scan(&p.ProjeID, &p.ProjeKodu, &p.BaslikTr, &p.BapTuru, &p.Tarih, &p.DurumAdi); err != nil {
			return nil, err
		}
		projeler = append(projeler, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projeler, nil
}

// GetProjectsByUyeIDForProfil fonksiyonu, belirli bir üyenin kabul ettiği tüm projelerini profil formatında getirir.
func (r *ProjeRepository) GetProjectsByUyeIDForProfil(uyeID int) ([]models.ProfilProjeBilgisi, error) {
	query := `
		SELECT p.proje_id,
		       COALESCE(p.proje_kodu, ''),
		       COALESCE(p.baslik_tr, 'Başlıksız Proje'),
		       COALESCE(pbt.bap_turu, 'Münferit'),
		       COALESCE(pd.durum_adi, 'taslak'),
		       COALESCE(prt.proje_rol, 'Araştırmacı')
		FROM proje p
		INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		LEFT JOIN proje_rol_tanimlama prt ON pt.proje_rol_id = prt.rol_id
		WHERE pt.uye_id = $1 AND pt.davet_durumu = 'kabul'
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
		if err := rows.Scan(&p.ProjeID, &p.ProjeKodu, &p.BaslikTr, &p.BapTuru, &p.DurumAdi, &p.UyeRol); err != nil {
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
	// Türkçe Yorum: Projeyi getirirken detay tablosundan özet, anahtar kelimeler ve diğer akademik bilgileri de çekiyoruz.
	// Nullable (NULL olabilecek) alanları Go tiplerine tararken hata almamak için COALESCE ile sarmalıyoruz.
	// proje_asama JOIN ile onay akışı aşamasını (asama_adi, asama_kodu) de çekiyoruz.
	query := `
		SELECT p.proje_id, COALESCE(p.proje_kodu, ''), COALESCE(p.baslik_tr, ''), COALESCE(p.baslik_en, ''),
		       COALESCE(p.sure_ay, 0), COALESCE(p.toplam_butce, 0), COALESCE(p.etik_kurul, false),
		       p.etik_kurul_no, p.koordinator_id, p.durum_id, p.asama_id, p.bap_turu_id,
		       p.olusturma_tarihi, p.guncelleme_tarihi,
		       COALESCE(pd.durum_adi, 'taslak'), COALESCE(pbt.bap_turu, 'Münferit'),
		       COALESCE(pa.asama_adi, ''), COALESCE(pa.asama_kodu, ''),
		       COALESCE(pdet.ozet, ''), COALESCE(pdet.ozet_en, ''),
		       COALESCE(pdet.anahtar_kelimeler, ''), COALESCE(pdet.anahtar_kelimeler_en, ''),
		       COALESCE(pdet.hedefler, ''), COALESCE(pdet.ozgunluk, ''), COALESCE(pdet.metodoloji, ''), COALESCE(pdet.kaynakca, '')
		FROM proje p
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		LEFT JOIN proje_asama pa ON p.asama_id = pa.asama_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN proje_detay pdet ON p.proje_id = pdet.proje_id
		WHERE p.proje_id = $1
	`
	p := &models.Proje{}
	err := r.DB.QueryRow(query, projeID).Scan(
		&p.ProjeID, &p.ProjeKodu, &p.BaslikTr, &p.BaslikEn, &p.SureAy, &p.ToplamButce, &p.EtikKurul,
		&p.EtikKurulNo, &p.KoordinatorID, &p.DurumID, &p.AsamaID, &p.BapTuruID,
		&p.OlusturmaTarihi, &p.GuncellemeTarihi,
		&p.DurumAdi, &p.BapTuru,
		&p.AsamaAdi, &p.AsamaKodu,
		&p.Ozet, &p.OzetEn, &p.AnahtarKelimeler, &p.AnahtarKelimelerEn,
		&p.Hedefler, &p.Ozgunluk, &p.Metodoloji, &p.Kaynakca,
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

// UpdateProjectStatusWithLog projenin genel durumunu ve iş akışı aşamasını günceller,
// bu değişikliği süreç geçmişi tablosuna kaydeder. Transaction kapsamında çalışır.
// Türkçe Yorum: yeniDurum genel durum adı, yeniAsamaKodu ise proje_asama.asama_kodu'dır.
// yeniAsamaKodu boşsa ("" veya "_silindi"), asama_id NULL'a çekilir (aşama bitti).
func (r *ProjeRepository) UpdateProjectStatusWithLog(projeID int, islemYapanID int, baslangicDurum, yeniDurum, aciklama string) error {
	return r.UpdateProjectStatusAndAsamaWithLog(projeID, islemYapanID, baslangicDurum, yeniDurum, "", aciklama)
}

// UpdateProjectStatusAndAsamaWithLog hem genel durumu hem aşama kodunu günceller.
// yeniAsamaKodu boşsa asama_id NULL'a çekilir.
func (r *ProjeRepository) UpdateProjectStatusAndAsamaWithLog(projeID int, islemYapanID int, baslangicDurum, yeniDurum, yeniAsamaKodu, aciklama string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Yeni genel durumun durum_id değerini bul
	var durumID int
	err = tx.QueryRow(`SELECT durum_id FROM proje_durum WHERE durum_adi = $1`, yeniDurum).Scan(&durumID)
	if err != nil {
		return fmt.Errorf("hedef durum (%s) bulunamadı: %v", yeniDurum, err)
	}

	// 2. Yeni aşamanın asama_id değerini bul (boşsa NULL yap)
	var asamaID *int
	if yeniAsamaKodu != "" {
		var id int
		err = tx.QueryRow(`SELECT asama_id FROM proje_asama WHERE asama_kodu = $1`, yeniAsamaKodu).Scan(&id)
		if err != nil {
			return fmt.Errorf("hedef aşama (%s) bulunamadı: %v", yeniAsamaKodu, err)
		}
		asamaID = &id
	}

	// 3. Projenin durum ve aşamasını güncelle
	_, err = tx.Exec(`UPDATE proje SET durum_id = $1, asama_id = $2, guncelleme_tarihi = CURRENT_TIMESTAMP WHERE proje_id = $3`, durumID, asamaID, projeID)
	if err != nil {
		return fmt.Errorf("proje durumu güncellenemedi: %v", err)
	}

	// 4. Süreç geçmişi tablosuna log kaydı ekle
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
	// Türkçe Yorum: Hakem rolündeki kullanıcıların adları ve unvanları timeline geçmişinde gizlenir.
	query := `
		SELECT g.gecmis_id, g.proje_id, g.islem_yapan_id, g.baslangic_durum, g.hedef_durum, g.aciklama, g.olusturma_tarihi,
		       CASE 
		           WHEN u.rol = 'komisyon' OR EXISTS(
		               SELECT 1 FROM sistem_rol sr 
		               JOIN sistem_rol_tanimlama srt ON sr.sistem_rol_id = srt.rol_id 
		               WHERE sr.uye_id = u.uye_id AND srt.rol_adi = 'komisyon'
		           ) THEN '*' 
		           WHEN u.rol = 'hakem' OR EXISTS(
		               SELECT 1 FROM sistem_rol sr 
		               JOIN sistem_rol_tanimlama srt ON sr.sistem_rol_id = srt.rol_id 
		               WHERE sr.uye_id = u.uye_id AND srt.rol_adi = 'hakem'
		           ) THEN 'Hakem'
		           ELSE COALESCE(u.ad || ' ' || u.soyad, '') 
		       END as ad_tumu,
		       CASE 
		           WHEN u.rol = 'hakem' OR EXISTS(
		               SELECT 1 FROM sistem_rol sr 
		               JOIN sistem_rol_tanimlama srt ON sr.sistem_rol_id = srt.rol_id 
		               WHERE sr.uye_id = u.uye_id AND srt.rol_adi = 'hakem'
		           ) THEN ''
		           ELSE COALESCE(u.unvan, '') 
		       END as unvan
		FROM  proje_surec_gecmisi g
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

// GetWorkflowHistoryByUyeID belirli bir kullanıcının geçmişte verdiği onay/red/revizyon kararlarını listeler.
func (r *ProjeRepository) GetWorkflowHistoryByUyeID(uyeID int) ([]models.ProjeSurecGecmisi, error) {
	query := `
		SELECT g.gecmis_id, g.proje_id, g.islem_yapan_id, g.baslangic_durum, g.hedef_durum, g.aciklama, g.olusturma_tarihi,
		       COALESCE(p.proje_kodu, ''), COALESCE(p.baslik_tr, 'Başlıksız Proje')
		FROM proje_surec_gecmisi g
		INNER JOIN proje p ON g.proje_id = p.proje_id
		WHERE g.islem_yapan_id = $1
		ORDER BY g.olusturma_tarihi DESC
	`
	rows, err := r.DB.Query(query, uyeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gecmis []models.ProjeSurecGecmisi
	for rows.Next() {
		var g models.ProjeSurecGecmisi
		err := rows.Scan(
			&g.GecmisID, &g.ProjeID, &g.IslemYapanID, &g.BaslangicDurum, &g.HedefDurum, &g.Aciklama, &g.OlusturmaTarihi,
			&g.ProjeKodu, &g.ProjeBaslik,
		)
		if err != nil {
			return nil, err
		}
		gecmis = append(gecmis, g)
	}
	return gecmis, nil
}

// GetProjectsForWorkflow belirli bir aşamadaki (asama_kodu) tüm projeleri listeler.
// Türkçe Yorum: asama_kodu boşluğu veya durum_adi ile filtrelenebilir.
// Komisyon rolünde olanlar için üyenin daha önce oylamadığı (karar = 'bekliyor') projeleri filtreler.
func (r *ProjeRepository) GetProjectsForWorkflow(rol string, filtre string, uyeID int) ([]models.Proje, error) {
	// Türkçe Yorum: Komisyon üyeleri için sadece kendi oylamadığı projeler listelenir.
	query := `
		SELECT p.proje_id, COALESCE(p.proje_kodu, ''), COALESCE(p.baslik_tr, ''), COALESCE(p.baslik_en, ''), 
		       COALESCE(p.sure_ay, 0), COALESCE(p.toplam_butce, 0), COALESCE(p.etik_kurul, false),
		       p.etik_kurul_no, p.koordinator_id, p.durum_id, p.asama_id, p.bap_turu_id,
		       p.olusturma_tarihi, p.guncelleme_tarihi,
		       COALESCE(pd.durum_adi, ''), COALESCE(pbt.bap_turu, ''),
		       COALESCE(pa.asama_adi, ''), COALESCE(pa.asama_kodu, ''),
		       COALESCE(u.unvan || ' ' || u.ad || ' ' || u.soyad, u.ad || ' ' || u.soyad, '') as koordinator_ad_soyad,
		       COALESCE(pbt.hakem_gerekli, false) as hakem_gerekli
		FROM proje p
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		LEFT JOIN proje_asama pa ON p.asama_id = pa.asama_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN uye u ON p.koordinator_id = u.uye_id
		WHERE (pa.asama_kodu = $1 OR pd.durum_adi = $2)
		  AND ($3 = 0 OR NOT EXISTS (
		      SELECT 1 FROM proje_komisyon_onay pko 
		      WHERE pko.proje_id = p.proje_id AND pko.komisyon_uye_id = $4 AND pko.karar != 'bekliyor'
		  ))
		ORDER BY p.guncelleme_tarihi DESC
	`
	
	filterUyeID := 0
	if rol == "komisyon" {
		filterUyeID = uyeID
	}

	rows, err := r.DB.Query(query, filtre, filtre, filterUyeID, filterUyeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projeler []models.Proje
	for rows.Next() {
		var p models.Proje
		err := rows.Scan(
			&p.ProjeID, &p.ProjeKodu, &p.BaslikTr, &p.BaslikEn, &p.SureAy, &p.ToplamButce, &p.EtikKurul,
			&p.EtikKurulNo, &p.KoordinatorID, &p.DurumID, &p.AsamaID, &p.BapTuruID,
			&p.OlusturmaTarihi, &p.GuncellemeTarihi,
			&p.DurumAdi, &p.BapTuru,
			&p.AsamaAdi, &p.AsamaKodu,
			&p.KoordinatorAdSoyad, &p.HakemGerekli,
		)
		if err != nil {
			return nil, err
		}
		projeler = append(projeler, p)
	}
	return projeler, nil
}

// IsPaketiInput, frontend'den gelen iş paketi verisi için input yapısıdır.
type IsPaketiInput struct {
	PaketAdi    string `json:"paket_adi"`
	PaketAmaci  string `json:"paket_amaci"`
	BaslangicAy int    `json:"baslangic_ay"`
	BitisAy     int    `json:"bitis_ay"`
}

// ButceKalemiInput, frontend'den gelen bütçe kalemi verisi için input yapısıdır.
type ButceKalemiInput struct {
	KategoriAdi string  `json:"kategori_adi"`
	Aciklama    string  `json:"aciklama"`
	Miktar      int     `json:"miktar"`
	BirimFiyat  float64 `json:"birim_fiyat"`
}

// SaveIsPaketleri projeye ait iş paketlerini kaydeder.
// Mevcut paketler silinip yeniden eklenir (upsert benzeri davranış).
func (r *ProjeRepository) SaveIsPaketleri(projeID int, paketler []IsPaketiInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("iş paketleri transaction başlatılamadı: %w", err)
	}
	defer tx.Rollback()

	// Mevcut iş paketlerini temizle
	_, err = tx.Exec(`DELETE FROM is_paketi WHERE proje_id = $1`, projeID)
	if err != nil {
		return fmt.Errorf("mevcut iş paketleri silinemedi: %w", err)
	}

	// Yeni iş paketlerini ekle
	for _, p := range paketler {
		_, err = tx.Exec(`
			INSERT INTO is_paketi (proje_id, paket_adi, paket_amaci, baslangic_ay, bitis_ay)
			VALUES ($1, $2, $3, $4, $5)
		`, projeID, p.PaketAdi, p.PaketAmaci, p.BaslangicAy, p.BitisAy)
		if err != nil {
			return fmt.Errorf("iş paketi eklenemedi: %w", err)
		}
	}

	return tx.Commit()
}

// SaveButceKalemleri projeye ait bütçe kalemlerini kaydeder.
// Mevcut kalemler silinip yeniden eklenir.
func (r *ProjeRepository) SaveButceKalemleri(projeID int, kalemler []ButceKalemiInput) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("bütçe kalemleri transaction başlatılamadı: %w", err)
	}
	defer tx.Rollback()

	// Mevcut bütçe kalemlerini temizle
	_, err = tx.Exec(`DELETE FROM butce WHERE proje_id = $1`, projeID)
	if err != nil {
		return fmt.Errorf("mevcut bütçe kalemleri silinemedi: %w", err)
	}

	// Yeni bütçe kalemlerini ekle
	for _, k := range kalemler {
		toplamFiyat := float64(k.Miktar) * k.BirimFiyat

		// Kategori adına göre kategori_id bul (bulunamazsa NULL olarak ekle)
		var kategoriID *int
		_ = tx.QueryRow(`SELECT kategori_id FROM butce_kategori WHERE kategori_adi = $1`, k.KategoriAdi).Scan(&kategoriID)

		_, err = tx.Exec(`
			INSERT INTO butce (proje_id, kategori_id, aciklama, birim_ozelligi, birim_fiyat, toplam_fiyat)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, projeID, kategoriID, k.Aciklama, k.Miktar, k.BirimFiyat, toplamFiyat)
		if err != nil {
			return fmt.Errorf("bütçe kalemi eklenemedi: %w", err)
		}
	}

	return tx.Commit()
}

// SaveProjeDetay projenin akademik detay bilgilerini kaydeder (upsert).
// Özet, anahtar kelimeler, hedefler, özgünlük, metodoloji, kaynakça alanlarını günceller.
func (r *ProjeRepository) SaveProjeDetay(detay *models.ProjeDetay) error {
	query := `
		INSERT INTO proje_detay (proje_id, ozet, ozet_en, anahtar_kelimeler, anahtar_kelimeler_en, hedefler, ozgunluk, metodoloji, kaynakca)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (proje_id) DO UPDATE SET
			ozet = EXCLUDED.ozet,
			ozet_en = EXCLUDED.ozet_en,
			anahtar_kelimeler = EXCLUDED.anahtar_kelimeler,
			anahtar_kelimeler_en = EXCLUDED.anahtar_kelimeler_en,
			hedefler = EXCLUDED.hedefler,
			ozgunluk = EXCLUDED.ozgunluk,
			metodoloji = EXCLUDED.metodoloji,
			kaynakca = EXCLUDED.kaynakca
	`
	_, err := r.DB.Exec(query, detay.ProjeID, detay.Ozet, detay.OzetEn, detay.AnahtarKelimeler, detay.AnahtarKelimelerEn, detay.Hedefler, detay.Ozgunluk, detay.Metodoloji, detay.Kaynakca)
	if err != nil {
		return fmt.Errorf("proje detay kaydedilemedi: %w", err)
	}
	return nil
}

// RiskInput, risk yönetimi kaydı oluşturmak için kullanılan girdi yapısıdır.
type RiskInput struct {
	RiskAciklamasi string `json:"risk_aciklamasi"`
	CozumPlani     string `json:"cozum_plani"`
}

// GetProjeDetaylar, projenin ek verilerini (bütçe, iş paketleri, risk, araştırma) döndürür.
func (r *ProjeRepository) GetProjeDetaylar(projeID int) ([]models.Butce, []models.IsPaketi, []models.RiskYonetimi, *models.Arastirma, error) {
	// Bütçe kalemlerini sorgula
	var butceler []models.Butce
	rowsButce, err := r.DB.Query(`
		SELECT b.kalem_id, b.proje_id, b.kategori_id, b.aciklama, COALESCE(b.birim_ozelligi, 0), b.birim_fiyat, b.toplam_fiyat, COALESCE(bk.kategori_adi, '')
		FROM butce b
		LEFT JOIN butce_kategori bk ON b.kategori_id = bk.kategori_id
		WHERE b.proje_id = $1
	`, projeID)
	if err == nil {
		defer rowsButce.Close()
		for rowsButce.Next() {
			var b models.Butce
			var katID *int
			if err := rowsButce.Scan(&b.KalemID, &b.ProjeID, &katID, &b.Aciklama, &b.BirimOzelligi, &b.BirimFiyat, &b.ToplamFiyat, &b.KategoriAdi); err == nil {
				b.KategoriID = katID
				butceler = append(butceler, b)
			}
		}
	}

	// İş paketlerini sorgula
	var isPaketleri []models.IsPaketi
	rowsWp, err := r.DB.Query(`
		SELECT paket_id, proje_id, paket_adi, COALESCE(paket_amaci, ''),
		       COALESCE(baslangic_ay, 1), COALESCE(bitis_ay, 1)
		FROM is_paketi
		WHERE proje_id = $1
	`, projeID)
	if err == nil {
		defer rowsWp.Close()
		for rowsWp.Next() {
			var ip models.IsPaketi
			if err := rowsWp.Scan(&ip.PaketID, &ip.ProjeID, &ip.PaketAdi, &ip.PaketAmaci, &ip.BaslangicAy, &ip.BitisAy); err == nil {
				isPaketleri = append(isPaketleri, ip)
			}
		}
	}

	// Risk yönetimi kayıtlarını sorgula
	var riskler []models.RiskYonetimi
	rowsRisk, err := r.DB.Query(`
		SELECT risk_id, proje_id, COALESCE(risk_aciklamasi, ''), COALESCE(cozum_plani, '')
		FROM risk_yonetimi WHERE proje_id = $1 ORDER BY risk_id ASC
	`, projeID)
	if err == nil {
		defer rowsRisk.Close()
		for rowsRisk.Next() {
			var risk models.RiskYonetimi
			if err := rowsRisk.Scan(&risk.RiskID, &risk.ProjeID, &risk.RiskAciklamasi, &risk.CozumPlani); err == nil {
				riskler = append(riskler, risk)
			}
		}
	}

	// Araştırma olanakları bilgisini sorgula
	var arastirma *models.Arastirma
	var a models.Arastirma
	err = r.DB.QueryRow(`
		SELECT proje_id, COALESCE(arastirma_amaci, '') FROM arastirma WHERE proje_id = $1
	`, projeID).Scan(&a.ProjeID, &a.ArastirmaAmaci)
	if err == nil {
		arastirma = &a
	}

	return butceler, isPaketleri, riskler, arastirma, nil
}

// SaveRiskYonetimi projeye ait risk yönetimi kayıtlarını günceller.
func (r *ProjeRepository) SaveRiskYonetimi(projeID int, riskler []RiskInput) error {
	_, err := r.DB.Exec(`DELETE FROM risk_yonetimi WHERE proje_id = $1`, projeID)
	if err != nil {
		return fmt.Errorf("risk kayıtları temizlenemedi: %w", err)
	}
	for _, risk := range riskler {
		if risk.RiskAciklamasi == "" {
			continue
		}
		_, err = r.DB.Exec(
			`INSERT INTO risk_yonetimi (proje_id, risk_aciklamasi, cozum_plani) VALUES ($1, $2, $3)`,
			projeID, risk.RiskAciklamasi, risk.CozumPlani,
		)
		if err != nil {
			return fmt.Errorf("risk kaydedilemedi: %w", err)
		}
	}
	return nil
}

// SaveArastirma projenin araştırma olanakları bilgisini kaydeder (upsert).
func (r *ProjeRepository) SaveArastirma(projeID int, olusturanID int, arastirmaAmaci string) error {
	query := `
		INSERT INTO arastirma (proje_id, olusturan_id, arastirma_amaci)
		VALUES ($1, $2, $3)
		ON CONFLICT (proje_id) DO UPDATE SET arastirma_amaci = EXCLUDED.arastirma_amaci
	`
	_, err := r.DB.Exec(query, projeID, olusturanID, arastirmaAmaci)
	if err != nil {
		return fmt.Errorf("araştırma olanakları kaydedilemedi: %w", err)
	}
	return nil
}

// YayinEtkiInput, yaygın etki çıktısı oluşturmak için kullanılan girdi yapısıdır.
type YayinEtkiInput struct {
	CiktiTuru    string `json:"cikti_turu"`
	OngorulCikti string `json:"ongorul_cikti"`
	ZamanAraligi string `json:"zaman_araligi"`
}

// YayginlastirmaEtkinlikInput, yaygınlaştırma etkinliği oluşturmak için kullanılan girdi yapısıdır.
type YayginlastirmaEtkinlikInput struct {
	EtkinlikTuru string `json:"etkinlik_turu"`
	Paydas       string `json:"paydas"`
	ZamanSure    string `json:"zaman_sure"`
	SiraNo       int    `json:"sira_no"`
}

// SaveYayinEtki projeye ait yaygın etki çıktılarını günceller (önce siler, sonra yeniden ekler).
func (r *ProjeRepository) SaveYayinEtki(projeID int, rows []YayinEtkiInput) error {
	_, err := r.DB.Exec(`DELETE FROM proje_yayin_etki WHERE proje_id = $1`, projeID)
	if err != nil {
		return fmt.Errorf("yaygın etki kayıtları temizlenemedi: %w", err)
	}
	for _, row := range rows {
		if row.CiktiTuru == "" {
			continue
		}
		_, err = r.DB.Exec(
			`INSERT INTO proje_yayin_etki (proje_id, cikti_turu, ongorul_cikti, zaman_araligi) VALUES ($1, $2, $3, $4)`,
			projeID, row.CiktiTuru, row.OngorulCikti, row.ZamanAraligi,
		)
		if err != nil {
			return fmt.Errorf("yaygın etki kaydedilemedi: %w", err)
		}
	}
	return nil
}

// SaveYayginlastirmaEtkinlik projeye ait yaygınlaştırma etkinliklerini günceller.
func (r *ProjeRepository) SaveYayginlastirmaEtkinlik(projeID int, etkinlikler []YayginlastirmaEtkinlikInput) error {
	_, err := r.DB.Exec(`DELETE FROM proje_yayginlastirma_etkinlik WHERE proje_id = $1`, projeID)
	if err != nil {
		return fmt.Errorf("yaygınlaştırma etkinlikleri temizlenemedi: %w", err)
	}
	for i, e := range etkinlikler {
		siraNo := e.SiraNo
		if siraNo == 0 {
			siraNo = i + 1
		}
		_, err = r.DB.Exec(
			`INSERT INTO proje_yayginlastirma_etkinlik (proje_id, etkinlik_turu, paydas, zaman_sure, sira_no) VALUES ($1, $2, $3, $4, $5)`,
			projeID, e.EtkinlikTuru, e.Paydas, e.ZamanSure, siraNo,
		)
		if err != nil {
			return fmt.Errorf("yaygınlaştırma etkinliği kaydedilemedi: %w", err)
		}
	}
	return nil
}

// GetProjeYayinEtkiBilgiler projeye ait yaygın etki ve yaygınlaştırma etkinliklerini döndürür.
func (r *ProjeRepository) GetProjeYayinEtkiBilgiler(projeID int) ([]models.ProjeYayinEtki, []models.ProjeYayginlastirmaEtkinlik, error) {
	var yayinEtki []models.ProjeYayinEtki
	rowsYE, err := r.DB.Query(
		`SELECT id, proje_id, cikti_turu, COALESCE(ongorul_cikti,''), COALESCE(zaman_araligi,'') FROM proje_yayin_etki WHERE proje_id = $1 ORDER BY id ASC`,
		projeID,
	)
	if err == nil {
		defer rowsYE.Close()
		for rowsYE.Next() {
			var ye models.ProjeYayinEtki
			if scanErr := rowsYE.Scan(&ye.ID, &ye.ProjeID, &ye.CiktiTuru, &ye.OngorulCikti, &ye.ZamanAraligi); scanErr == nil {
				yayinEtki = append(yayinEtki, ye)
			}
		}
	}

	var etkinlikler []models.ProjeYayginlastirmaEtkinlik
	rowsEtk, err2 := r.DB.Query(
		`SELECT id, proje_id, COALESCE(etkinlik_turu,''), COALESCE(paydas,''), COALESCE(zaman_sure,''), sira_no FROM proje_yayginlastirma_etkinlik WHERE proje_id = $1 ORDER BY sira_no ASC`,
		projeID,
	)
	if err2 == nil {
		defer rowsEtk.Close()
		for rowsEtk.Next() {
			var e models.ProjeYayginlastirmaEtkinlik
			if scanErr := rowsEtk.Scan(&e.ID, &e.ProjeID, &e.EtkinlikTuru, &e.Paydas, &e.ZamanSure, &e.SiraNo); scanErr == nil {
				etkinlikler = append(etkinlikler, e)
			}
		}
	}

	return yayinEtki, etkinlikler, nil
}

// IsProjeUyesi kullanıcının projenin kabul edilmiş bir ekip üyesi olup olmadığını kontrol eder.
// Türkçe Yorum: Belirtilen üye ID'sinin, ilgili projenin takım tablosunda 'kabul' durumunda olup olmadığını sorgular.
func (r *ProjeRepository) IsProjeUyesi(projeID int, uyeID int) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM proje_takim 
		WHERE proje_id = $1 AND uye_id = $2 AND davet_durumu = 'kabul'
	`
	err := r.DB.QueryRow(query, projeID, uyeID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateKomisyonOnayRecords projeyi oylayacak komisyon üyeleri için onay kayıtlarını oluşturur.
// Türkçe Yorum: Proje komisyona sevk edildiğinde aktif komisyon üyeleri için oylama kaydı açar. Eğer daha önce oylama kaydı açılmışsa oyları 'bekliyor' durumuna sıfırlar.
func (r *ProjeRepository) CreateKomisyonOnayRecords(projeID int) error {
	query := `
		INSERT INTO proje_komisyon_onay (proje_id, komisyon_uye_id, karar, aciklama)
		SELECT $1, u.uye_id, 'bekliyor', NULL
		FROM uye u
		JOIN sistem_rol sr ON u.uye_id = sr.uye_id
		JOIN sistem_rol_tanimlama srt ON sr.sistem_rol_id = srt.rol_id
		WHERE srt.rol_adi = 'komisyon' AND u.aktif_mi = true AND u.rol = 'komisyon'
		ON CONFLICT (proje_id, komisyon_uye_id) DO UPDATE
		SET karar = 'bekliyor', aciklama = NULL, guncelleme_tarihi = CURRENT_TIMESTAMP
	`
	_, err := r.DB.Exec(query, projeID)
	return err
}

// UpdateKomisyonKarar komisyon üyesinin bireysel oylama kararını günceller.
// Türkçe Yorum: Belirli bir komisyon üyesinin ilgili proje hakkındaki kabul, red veya revizyon kararını kaydeder.
func (r *ProjeRepository) UpdateKomisyonKarar(projeID int, komisyonUyeID int, karar string, aciklama string) error {
	query := `
		INSERT INTO proje_komisyon_onay (proje_id, komisyon_uye_id, karar, aciklama, guncelleme_tarihi)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
		ON CONFLICT (proje_id, komisyon_uye_id) DO UPDATE
		SET karar = EXCLUDED.karar, aciklama = EXCLUDED.aciklama, guncelleme_tarihi = CURRENT_TIMESTAMP
	`
	_, err := r.DB.Exec(query, projeID, komisyonUyeID, karar, aciklama)
	return err
}

// CheckAllKomisyonApproved komisyon üyelerinin oylama sonuçlarını kontrol eder.
// Türkçe Yorum: Atanmış tüm komisyon üyelerinin onay verip vermediğini sorgular. Herhangi biri red veya revizyon istemişse bunu döner.
func (r *ProjeRepository) CheckAllKomisyonApproved(projeID int) (bool, string, error) {
	rows, err := r.DB.Query(`SELECT karar FROM proje_komisyon_onay WHERE proje_id = $1`, projeID)
	if err != nil {
		return false, "", err
	}
	defer rows.Close()

	total := 0
	approved := 0
	hasRejected := false
	hasRevision := false

	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return false, "", err
		}
		total++
		if k == "onayla" {
			approved++
		} else if k == "reddet" {
			hasRejected = true
		} else if k == "revizyon" {
			hasRevision = true
		}
	}

	if hasRejected {
		return false, "reddet", nil
	}
	if hasRevision {
		return false, "revizyon", nil
	}

	return total > 0 && approved == total, "", nil
}

