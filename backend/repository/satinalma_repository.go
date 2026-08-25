package repository

import (
	"database/sql"
	"fmt"
	"time"

	"bap_ai/backend/models"
)

// SatinalmaRepository yapısı veritabanı işlemlerini barındırır.
// Türkçe Yorum: Satın alma talepleri tablosu için CRUD ve bütçe limit kontrollerini gerçekleştiren veri erişim katmanıdır.
type SatinalmaRepository struct {
	DB *sql.DB
}

// NewSatinalmaRepository yeni bir SatinalmaRepository nesnesi döner.
// Türkçe Yorum: SatinalmaRepository için dependency injection kurucusu.
func NewSatinalmaRepository(db *sql.DB) *SatinalmaRepository {
	return &SatinalmaRepository{DB: db}
}

// CreatePurchaseRequests veritabanına toplu satın alma talebi ekler.
// Türkçe Yorum: Toplu satın alma kalemlerine sistem genelinde benzersiz bir talep_no (SA-XXX) atar.
func (r *SatinalmaRepository) CreatePurchaseRequests(reqs []*models.SatinalmaTalebi) error {
	if len(reqs) == 0 {
		return nil
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	projeID := reqs[0].ProjeID
	for _, req := range reqs {
		if req.ProjeID != projeID {
			return fmt.Errorf("toplu satın alma talebinde tüm kalemler aynı projeye ait olmalıdır")
		}
	}

	// 1. Sistem genelinde en yüksek SA numarasını bul (proje bazlı tekrarlanmayı önler)
	var maxNo int
	err = tx.QueryRow(`
		SELECT COALESCE(MAX(
			CASE WHEN talep_no ~ '^SA-[0-9]+$'
				THEN CAST(SUBSTRING(talep_no FROM 4) AS INTEGER)
				ELSE 0
			END
		), 0)
		FROM proje_satinalma_talebi
	`).Scan(&maxNo)
	if err != nil {
		return fmt.Errorf("mevcut satın alma talep numaraları okunamadı: %w", err)
	}

	// 2. Global benzersiz satın alma talep numarası oluştur (Örn: SA-001, SA-002, ...)
	talepNo := fmt.Sprintf("SA-%03d", maxNo+1)

	// 3. Tüm kalemleri aynı talep_no ile ekle
	query := `
		INSERT INTO proje_satinalma_talebi (proje_id, uye_id, kalem_id, malzeme_adi, miktar, birim_fiyat, toplam_fiyat, durum, gerekce, talep_no)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'Beklemede', $8, $9)
		RETURNING talep_id, olusturma_tarihi, guncelleme_tarihi
	`

	for _, req := range reqs {
		toplamTutar := float64(req.Miktar) * req.BirimFiyat
		err = tx.QueryRow(query, req.ProjeID, req.UyeID, req.KalemID, req.MalzemeAdi, req.Miktar, req.BirimFiyat, toplamTutar, req.Gerekce, talepNo).
			Scan(&req.TalepID, &req.OlusturmaTarihi, &req.GuncellemeTarihi)
		if err != nil {
			return fmt.Errorf("satın alma talebi eklenirken veritabanı hatası: %w", err)
		}
		req.ToplamFiyat = toplamTutar
		req.Durum = "Beklemede"
		req.TalepNo = talepNo
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// CreatePurchaseRequest veritabanına yeni bir satın alma talebi ekler.
// Türkçe Yorum: Geriye dönük uyumluluk için tekli satın alma ekleme isteklerini toplu ekleme metoduna yönlendirir.
func (r *SatinalmaRepository) CreatePurchaseRequest(req *models.SatinalmaTalebi) error {
	return r.CreatePurchaseRequests([]*models.SatinalmaTalebi{req})
}

// GetPurchaseRequestsByProject belirli bir projeye ait tüm satın alma taleplerini listeler.
// Türkçe Yorum: Akademisyenin kendi projesine ait geçmiş ve bekleyen tüm satın alma taleplerini detaylarıyla getirir.
func (r *SatinalmaRepository) GetPurchaseRequestsByProject(projeID int) ([]models.SatinalmaTalebi, error) {
	query := `
		SELECT 
			st.talep_id, COALESCE(st.talep_no, '') AS talep_no, st.proje_id, st.uye_id, st.kalem_id, st.malzeme_adi, st.miktar, st.birim_fiyat, st.toplam_fiyat, st.durum, st.gerekce, st.red_nedeni, st.olusturma_tarihi, st.guncelleme_tarihi,
			st.revize_birim_fiyat, st.revize_toplam_fiyat, st.revizyon_gerekcesi, st.revize_eden_id, st.revizyon_tarihi,
			u.ad || ' ' || u.soyad AS uye_ad_soyad,
			COALESCE(u2.ad || ' ' || u2.soyad, '') AS revize_eden_ad_soyad,
			COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'), '') AS proje_baslik,
			COALESCE(p.proje_kodu, '') AS proje_kodu,
			b.aciklama AS kalem_aciklama,
			COALESCE(bk.kategori_adi, 'Belirtilmemiş') AS butce_kategori_adi,
			b.toplam_fiyat AS mevcut_butce
		FROM proje_satinalma_talebi st
		INNER JOIN uye u ON st.uye_id = u.uye_id
		LEFT JOIN uye u2 ON st.revize_eden_id = u2.uye_id
		INNER JOIN proje p ON st.proje_id = p.proje_id
		INNER JOIN proje_butce b ON st.kalem_id = b.kalem_id
		LEFT JOIN proje_butce_kategori bk ON b.kategori_id = bk.kategori_id
		WHERE st.proje_id = $1
		ORDER BY st.olusturma_tarihi DESC
	`
	rows, err := r.DB.Query(query, projeID)
	if err != nil {
		return nil, fmt.Errorf("proje satın alma talepleri listelenirken hata: %w", err)
	}
	defer rows.Close()

	var talepler []models.SatinalmaTalebi
	for rows.Next() {
		var t models.SatinalmaTalebi
		var redNedeni sql.NullString
		err := rows.Scan(
			&t.TalepID, &t.TalepNo, &t.ProjeID, &t.UyeID, &t.KalemID, &t.MalzemeAdi, &t.Miktar, &t.BirimFiyat, &t.ToplamFiyat, &t.Durum, &t.Gerekce, &redNedeni, &t.OlusturmaTarihi, &t.GuncellemeTarihi,
			&t.RevizeBirimFiyat, &t.RevizeToplamFiyat, &t.RevizyonGerekcesi, &t.RevizeEdenID, &t.RevizyonTarihi,
			&t.UyeAdSoyad, &t.RevizeEdenAdSoyad, &t.ProjeBaslik, &t.ProjeKodu, &t.KalemAciklama, &t.ButceKategoriAdi, &t.MevcutButce,
		)
		if err != nil {
			return nil, fmt.Errorf("satın alma satırı okunurken hata: %w", err)
		}
		if redNedeni.Valid {
			val := redNedeni.String
			t.RedNedeni = &val
		}
		talepler = append(talepler, t)
	}

	return talepler, nil
}

// GetAllPurchaseRequests tüm projelerdeki satın alma taleplerini listeler.
// Türkçe Yorum: TTO yetkilisinin tüm sistemdeki satın alma taleplerini (bekleyen ve onaylanmış) onay ekranı için çekmesini sağlar.
func (r *SatinalmaRepository) GetAllPurchaseRequests() ([]models.SatinalmaTalebi, error) {
	query := `
		SELECT 
			st.talep_id, COALESCE(st.talep_no, '') AS talep_no, st.proje_id, st.uye_id, st.kalem_id, st.malzeme_adi, st.miktar, st.birim_fiyat, st.toplam_fiyat, st.durum, st.gerekce, st.red_nedeni, st.olusturma_tarihi, st.guncelleme_tarihi,
			st.revize_birim_fiyat, st.revize_toplam_fiyat, st.revizyon_gerekcesi, st.revize_eden_id, st.revizyon_tarihi,
			u.ad || ' ' || u.soyad AS uye_ad_soyad,
			COALESCE(u2.ad || ' ' || u2.soyad, '') AS revize_eden_ad_soyad,
			COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'), '') AS proje_baslik,
			COALESCE(p.proje_kodu, '') AS proje_kodu,
			b.aciklama AS kalem_aciklama,
			COALESCE(bk.kategori_adi, 'Belirtilmemiş') AS butce_kategori_adi,
			b.toplam_fiyat AS mevcut_butce
		FROM proje_satinalma_talebi st
		INNER JOIN uye u ON st.uye_id = u.uye_id
		LEFT JOIN uye u2 ON st.revize_eden_id = u2.uye_id
		INNER JOIN proje p ON st.proje_id = p.proje_id
		INNER JOIN proje_butce b ON st.kalem_id = b.kalem_id
		LEFT JOIN proje_butce_kategori bk ON b.kategori_id = bk.kategori_id
		ORDER BY st.durum DESC, st.olusturma_tarihi DESC
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("tüm satın alma talepleri listelenirken hata: %w", err)
	}
	defer rows.Close()

	var talepler []models.SatinalmaTalebi
	for rows.Next() {
		var t models.SatinalmaTalebi
		var redNedeni sql.NullString
		err := rows.Scan(
			&t.TalepID, &t.TalepNo, &t.ProjeID, &t.UyeID, &t.KalemID, &t.MalzemeAdi, &t.Miktar, &t.BirimFiyat, &t.ToplamFiyat, &t.Durum, &t.Gerekce, &redNedeni, &t.OlusturmaTarihi, &t.GuncellemeTarihi,
			&t.RevizeBirimFiyat, &t.RevizeToplamFiyat, &t.RevizyonGerekcesi, &t.RevizeEdenID, &t.RevizyonTarihi,
			&t.UyeAdSoyad, &t.RevizeEdenAdSoyad, &t.ProjeBaslik, &t.ProjeKodu, &t.KalemAciklama, &t.ButceKategoriAdi, &t.MevcutButce,
		)
		if err != nil {
			return nil, fmt.Errorf("satın alma satırı okunurken hata: %w", err)
		}
		if redNedeni.Valid {
			val := redNedeni.String
			t.RedNedeni = &val
		}
		talepler = append(talepler, t)
	}

	return talepler, nil
}

// UpdatePurchaseStatus satın alma talebinin durumunu günceller.
// Türkçe Yorum: Aynı talep_no + proje_id grubundaki tüm kalemleri onaylar/reddeder; diğer projelere sızmaz.
func (r *SatinalmaRepository) UpdatePurchaseStatus(talepID int, status string, redNedeni string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. İlgili talebin talep_no ve proje_id bilgisini bul
	var talepNo string
	var projeID int
	err = tx.QueryRow(`
		SELECT COALESCE(talep_no, ''), proje_id
		FROM proje_satinalma_talebi WHERE talep_id = $1
	`, talepID).Scan(&talepNo, &projeID)
	if err != nil {
		return fmt.Errorf("talep numarası bulunamadı: %w", err)
	}

	var redVal interface{} = nil
	if redNedeni != "" {
		redVal = redNedeni
	}

	// 2. Eğer talep_no boşsa sadece o satırı güncelle; aksi halde aynı proje + talep_no grubunu güncelle
	if talepNo == "" || talepNo == "-" {
		query := `
			UPDATE proje_satinalma_talebi
			SET durum = $1, red_nedeni = $2, guncelleme_tarihi = $3
			WHERE talep_id = $4
		`
		_, err = tx.Exec(query, status, redVal, time.Now(), talepID)
	} else {
		query := `
			UPDATE proje_satinalma_talebi
			SET durum = $1, red_nedeni = $2, guncelleme_tarihi = $3
			WHERE talep_no = $4 AND proje_id = $5
		`
		_, err = tx.Exec(query, status, redVal, time.Now(), talepNo, projeID)
	}

	if err != nil {
		return fmt.Errorf("satın alma talebi/talepleri güncellenirken hata: %w", err)
	}

	return tx.Commit()
}

// GetRemainingBudget onay için kullanılabilir kalanı döner (bekleyen hariç).
// Türkçe Yorum: Planlanan − açık taahhüt − fiili harcama. Onay anında bekleyen bu talebe dönüşeceği için bekleyen düşülmez.
func (r *SatinalmaRepository) GetRemainingBudget(projeID int, kalemID int) (float64, error) {
	breakdown, err := r.GetBudgetBreakdown(projeID, kalemID)
	if err != nil {
		return 0, err
	}
	return breakdown.Planlanan - breakdown.Taahhut - breakdown.Fiili, nil
}

// BudgetBreakdown bir kalemin 3 katmanlı bütçe özetidir.
type BudgetBreakdown struct {
	Planlanan float64
	Taahhut   float64
	Fiili     float64
	Bekleyen  float64
}

// GetBudgetBreakdown bütçe kalemi havuzunun planlanan / taahhüt / fiili / bekleyen tutarlarını hesaplar.
// Türkçe Yorum: Aynı kategoriye bağlı proje_butce satırları tek bütçe kalemi havuzudur.
// Seçilen malzeme satırının kategori kimliği bulunur ve tüm hesaplar kategori toplamı üzerinden yapılır.
func (r *SatinalmaRepository) GetBudgetBreakdown(projeID int, kalemID int) (*BudgetBreakdown, error) {
	var b BudgetBreakdown
	err := r.DB.QueryRow(`
		WITH secili AS (
			SELECT kategori_id
			FROM proje_butce
			WHERE proje_id = $1 AND kalem_id = $2
		)
		SELECT COALESCE(SUM(pb.toplam_fiyat), 0)
		FROM proje_butce pb
		CROSS JOIN secili s
		WHERE pb.proje_id = $1
		  AND (
			(s.kategori_id IS NOT NULL AND pb.kategori_id = s.kategori_id)
			OR (s.kategori_id IS NULL AND pb.kalem_id = $2)
		  )
	`, projeID, kalemID).Scan(&b.Planlanan)
	if err != nil {
		return nil, fmt.Errorf("bütçe kalemi havuzu bulunamadı: %w", err)
	}

	err = r.DB.QueryRow(`
		WITH secili AS (
			SELECT kategori_id
			FROM proje_butce
			WHERE proje_id = $1 AND kalem_id = $2
		)
		SELECT COALESCE(SUM(COALESCE(st.revize_toplam_fiyat, st.toplam_fiyat)), 0)
		FROM proje_satinalma_talebi st
		JOIN proje_butce pb ON pb.kalem_id = st.kalem_id
		CROSS JOIN secili s
		WHERE st.proje_id = $1
		  AND (
			(s.kategori_id IS NOT NULL AND pb.kategori_id = s.kategori_id)
			OR (s.kategori_id IS NULL AND st.kalem_id = $2)
		  )
		  AND st.durum = 'Onaylandı'
		  AND NOT EXISTS (
			SELECT 1 FROM proje_satinalma_odeme o
			WHERE o.talep_id = st.talep_id AND o.durum = 'onaylandi'
		  )
	`, projeID, kalemID).Scan(&b.Taahhut)
	if err != nil {
		return nil, fmt.Errorf("taahhüt tutarı hesaplanamadı: %w", err)
	}

	err = r.DB.QueryRow(`
		WITH secili AS (
			SELECT kategori_id
			FROM proje_butce
			WHERE proje_id = $1 AND kalem_id = $2
		)
		SELECT COALESCE(SUM(o.fiili_tutar), 0)
		FROM proje_satinalma_odeme o
		JOIN proje_butce pb ON pb.kalem_id = o.kalem_id
		CROSS JOIN secili s
		WHERE o.proje_id = $1
		  AND (
			(s.kategori_id IS NOT NULL AND pb.kategori_id = s.kategori_id)
			OR (s.kategori_id IS NULL AND o.kalem_id = $2)
		  )
		  AND o.durum = 'onaylandi'
	`, projeID, kalemID).Scan(&b.Fiili)
	if err != nil {
		return nil, fmt.Errorf("fiili tutar hesaplanamadı: %w", err)
	}

	err = r.DB.QueryRow(`
		WITH secili AS (
			SELECT kategori_id
			FROM proje_butce
			WHERE proje_id = $1 AND kalem_id = $2
		)
		SELECT COALESCE(SUM(COALESCE(st.revize_toplam_fiyat, st.toplam_fiyat)), 0)
		FROM proje_satinalma_talebi st
		JOIN proje_butce pb ON pb.kalem_id = st.kalem_id
		CROSS JOIN secili s
		WHERE st.proje_id = $1
		  AND (
			(s.kategori_id IS NOT NULL AND pb.kategori_id = s.kategori_id)
			OR (s.kategori_id IS NULL AND st.kalem_id = $2)
		  )
		  AND st.durum = 'Beklemede'
	`, projeID, kalemID).Scan(&b.Bekleyen)
	if err != nil {
		return nil, fmt.Errorf("bekleyen tutar hesaplanamadı: %w", err)
	}

	return &b, nil
}

// GetPurchaseRequestByID satın alma talebini getirir.
// Türkçe Yorum: Satın alma talebini ID bazında getirmek için kullanılır.
func (r *SatinalmaRepository) GetPurchaseRequestByID(talepID int) (*models.SatinalmaTalebi, error) {
	query := `
		SELECT 
			st.talep_id, COALESCE(st.talep_no, '') AS talep_no, st.proje_id, st.uye_id, st.kalem_id, st.malzeme_adi, st.miktar, st.birim_fiyat, st.toplam_fiyat, st.durum, st.gerekce, st.red_nedeni, st.olusturma_tarihi, st.guncelleme_tarihi,
			st.revize_birim_fiyat, st.revize_toplam_fiyat, st.revizyon_gerekcesi, st.revize_eden_id, st.revizyon_tarihi,
			u.ad || ' ' || u.soyad AS uye_ad_soyad,
			COALESCE(u2.ad || ' ' || u2.soyad, '') AS revize_eden_ad_soyad
		FROM proje_satinalma_talebi st
		INNER JOIN uye u ON st.uye_id = u.uye_id
		LEFT JOIN uye u2 ON st.revize_eden_id = u2.uye_id
		WHERE st.talep_id = $1
	`
	var t models.SatinalmaTalebi
	var redNedeni sql.NullString
	err := r.DB.QueryRow(query, talepID).Scan(
		&t.TalepID, &t.TalepNo, &t.ProjeID, &t.UyeID, &t.KalemID, &t.MalzemeAdi, &t.Miktar, &t.BirimFiyat, &t.ToplamFiyat, &t.Durum, &t.Gerekce, &redNedeni, &t.OlusturmaTarihi, &t.GuncellemeTarihi,
		&t.RevizeBirimFiyat, &t.RevizeToplamFiyat, &t.RevizyonGerekcesi, &t.RevizeEdenID, &t.RevizyonTarihi,
		&t.UyeAdSoyad, &t.RevizeEdenAdSoyad,
	)
	if err != nil {
		return nil, err
	}
	if redNedeni.Valid {
		val := redNedeni.String
		t.RedNedeni = &val
	}
	return &t, nil
}

// GetReservedBudget yeni talep için kullanılabilir kalanı döner.
// Türkçe Yorum: Planlanan − açık taahhüt − fiili − bekleyen.
func (r *SatinalmaRepository) GetReservedBudget(projeID int, kalemID int) (float64, error) {
	breakdown, err := r.GetBudgetBreakdown(projeID, kalemID)
	if err != nil {
		return 0, err
	}
	return breakdown.Planlanan - breakdown.Taahhut - breakdown.Fiili - breakdown.Bekleyen, nil
}

// RevisePurchaseRequest satın alma talebinin fiyatını günceller ve revizyon gerekçesini kaydeder.
// Türkçe Yorum: Admin veya yetkilendirilmiş personel (TTO) tarafından bütçe kalemi fiyatının güncellenmesini ve revizyon logs kaydı olarak gerekçesiyle tutulmasını sağlar.
func (r *SatinalmaRepository) RevisePurchaseRequest(talepID int, yeniBirimFiyat float64, gerekce string, yetkiliID int) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Talebe ait miktarı öğrenerek yeni toplam fiyatı hesapla
	var miktar int
	err = tx.QueryRow(`SELECT miktar FROM proje_satinalma_talebi WHERE talep_id = $1`, talepID).Scan(&miktar)
	if err != nil {
		return fmt.Errorf("talep bulunamadı: %w", err)
	}

	yeniToplamFiyat := yeniBirimFiyat * float64(miktar)

	// 2. Revizyon alanlarını güncelle
	query := `
		UPDATE proje_satinalma_talebi
		SET revize_birim_fiyat = $1,
			revize_toplam_fiyat = $2,
			revizyon_gerekcesi = $3,
			revize_eden_id = $4,
			revizyon_tarihi = $5,
			guncelleme_tarihi = $5
		WHERE talep_id = $6
	`
	_, err = tx.Exec(query, yeniBirimFiyat, yeniToplamFiyat, gerekce, yetkiliID, time.Now(), talepID)
	if err != nil {
		return fmt.Errorf("satın alma talebi revize edilirken veritabanı hatası: %w", err)
	}

	return tx.Commit()
}

// GetProjectBudgetReport projenin bütçe kalemi bazlı harcama raporunu üretir.
// Türkçe Yorum: Planlanan, açık taahhüt, fiili, bekleyen ve kullanılabilir tutarları hesaplar.
func (r *SatinalmaRepository) GetProjectBudgetReport(projeID int) (*models.ProjeButceHarcamaRaporu, error) {
	rapor := &models.ProjeButceHarcamaRaporu{ProjeID: projeID}

	err := r.DB.QueryRow(`
		SELECT COALESCE(p.proje_kodu, ''),
		       COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), '')
		FROM proje p WHERE p.proje_id = $1
	`, projeID).Scan(&rapor.ProjeKodu, &rapor.ProjeBaslik)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("proje bulunamadı")
	}
	if err != nil {
		return nil, fmt.Errorf("proje bilgisi alınamadı: %w", err)
	}

	rows, err := r.DB.Query(`
		WITH havuz AS (
			SELECT
				b.proje_id,
				b.kategori_id,
				MIN(b.kalem_id) AS kalem_id,
				COALESCE(bk.kategori_adi, 'Belirtilmemiş') AS kategori_adi,
				STRING_AGG(COALESCE(NULLIF(b.aciklama, ''), 'Açıklamasız'), ' | ' ORDER BY b.kalem_id) AS aciklama,
				SUM(COALESCE(b.toplam_fiyat, 0)) AS planlanan
			FROM proje_butce b
			LEFT JOIN proje_butce_kategori bk ON b.kategori_id = bk.kategori_id
			WHERE b.proje_id = $1
			GROUP BY b.proje_id, b.kategori_id, bk.kategori_adi
		)
		SELECT
			h.kalem_id,
			h.kategori_adi,
			h.aciklama,
			h.planlanan,
			COALESCE((
				SELECT SUM(COALESCE(st.revize_toplam_fiyat, st.toplam_fiyat))
				FROM proje_satinalma_talebi st
				JOIN proje_butce sb ON sb.kalem_id = st.kalem_id
				WHERE st.proje_id = h.proje_id
				  AND sb.kategori_id IS NOT DISTINCT FROM h.kategori_id
				  AND st.durum = 'Onaylandı'
				  AND NOT EXISTS (
					SELECT 1 FROM proje_satinalma_odeme o
					WHERE o.talep_id = st.talep_id AND o.durum = 'onaylandi'
				  )
			), 0) AS taahhut,
			COALESCE((
				SELECT SUM(o.fiili_tutar)
				FROM proje_satinalma_odeme o
				JOIN proje_butce ob ON ob.kalem_id = o.kalem_id
				WHERE o.proje_id = h.proje_id
				  AND ob.kategori_id IS NOT DISTINCT FROM h.kategori_id
				  AND o.durum = 'onaylandi'
			), 0) AS fiili,
			COALESCE((
				SELECT SUM(COALESCE(st.revize_toplam_fiyat, st.toplam_fiyat))
				FROM proje_satinalma_talebi st
				JOIN proje_butce sb ON sb.kalem_id = st.kalem_id
				WHERE st.proje_id = h.proje_id
				  AND sb.kategori_id IS NOT DISTINCT FROM h.kategori_id
				  AND st.durum = 'Beklemede'
			), 0) AS bekleyen
		FROM havuz h
		ORDER BY h.kategori_adi
	`, projeID)
	if err != nil {
		return nil, fmt.Errorf("bütçe raporu sorgulanamadı: %w", err)
	}
	defer rows.Close()

	rapor.Kalemler = []models.ButceHarcamaRaporKalemi{}
	for rows.Next() {
		var k models.ButceHarcamaRaporKalemi
		if err := rows.Scan(&k.KalemID, &k.KategoriAdi, &k.Aciklama, &k.Planlanan, &k.Taahhut, &k.Fiili, &k.Bekleyen); err != nil {
			return nil, err
		}
		k.Kullanilabilir = k.Planlanan - k.Taahhut - k.Fiili - k.Bekleyen
		k.Harcanan = k.Taahhut + k.Fiili + k.Bekleyen
		k.Odenen = k.Fiili
		k.Kalan = k.Kullanilabilir
		rapor.Kalemler = append(rapor.Kalemler, k)
		rapor.ToplamPlanlanan += k.Planlanan
		rapor.ToplamTaahhut += k.Taahhut
		rapor.ToplamFiili += k.Fiili
		rapor.ToplamBekleyen += k.Bekleyen
		rapor.ToplamKullanilabilir += k.Kullanilabilir
		rapor.ToplamHarcanan += k.Harcanan
		rapor.ToplamOdenen += k.Odenen
		rapor.ToplamKalan += k.Kalan
	}
	return rapor, nil
}

// HasApprovedOdeme talebin onaylı mutabakat kaydı olup olmadığını kontrol eder.
func (r *SatinalmaRepository) HasApprovedOdeme(talepID int) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM proje_satinalma_odeme WHERE talep_id = $1 AND durum = 'onaylandi'
		)
	`, talepID).Scan(&exists)
	return exists, err
}

// InsertButceHareketTx ledger satırı ekler (transaction içinde).
func (r *SatinalmaRepository) InsertButceHareketTx(tx *sql.Tx, h *models.ButceHareket) error {
	_, err := tx.Exec(`
		INSERT INTO proje_butce_hareket
			(proje_id, kalem_id, kaynak_tip, kaynak_id, hareket_tip, tutar, aciklama, islemi_yapan_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, h.ProjeID, h.KalemID, h.KaynakTip, h.KaynakID, h.HareketTip, h.Tutar, h.Aciklama, h.IslemiYapanID)
	if err != nil {
		return fmt.Errorf("bütçe hareketi yazılamadı: %w", err)
	}
	return nil
}

// CreateMutabakatTx fiili ödeme kaydını oluşturur ve talebi kapatır/iptal eder.
// Türkçe Yorum: Transaction içinde odeme + ledger + talep durum güncellemesi yapar.
func (r *SatinalmaRepository) CreateMutabakatTx(
	tx *sql.Tx,
	odeme *models.SatinalmaOdeme,
	yeniTalepDurum string,
	islemiYapanID int,
) error {
	now := time.Now()
	err := tx.QueryRow(`
		INSERT INTO proje_satinalma_odeme (
			talep_id, talep_no, proje_id, kalem_id,
			taahhut_tutari, fiili_tutar, fark_tutari, fark_yonu,
			fatura_no, fatura_tarihi, odeme_tarihi, para_birimi,
			durum, tto_uye_id, tto_gerekce, karar_tarihi
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10, $11, COALESCE(NULLIF($12, ''), 'TRY'),
			$13, $14, $15, $16
		)
		RETURNING odeme_id, olusturma_tarihi, guncelleme_tarihi
	`,
		odeme.TalepID, odeme.TalepNo, odeme.ProjeID, odeme.KalemID,
		odeme.TaahhutTutari, odeme.FiiliTutar, odeme.FarkTutari, odeme.FarkYonu,
		odeme.FaturaNo, odeme.FaturaTarihi, odeme.OdemeTarihi, odeme.ParaBirimi,
		odeme.Durum, odeme.TtoUyeID, odeme.TtoGerekce, now,
	).Scan(&odeme.OdemeID, &odeme.OlusturmaTarihi, &odeme.GuncellemeTarihi)
	if err != nil {
		return fmt.Errorf("ödeme/mutabakat kaydı oluşturulamadı: %w", err)
	}
	odeme.KararTarihi = &now

	_, err = tx.Exec(`
		UPDATE proje_satinalma_talebi
		SET durum = $1, guncelleme_tarihi = $2
		WHERE talep_id = $3
	`, yeniTalepDurum, now, odeme.TalepID)
	if err != nil {
		return fmt.Errorf("satın alma durumu güncellenemedi: %w", err)
	}

	// Ledger: taahhüt iptali (serbest bırakma + işaretli)
	aciklamaRez := "Mutabakat: taahhüt serbest bırakıldı"
	if err = r.InsertButceHareketTx(tx, &models.ButceHareket{
		ProjeID: odeme.ProjeID, KalemID: odeme.KalemID,
		KaynakTip: "satinalma_mutabakat", KaynakID: odeme.OdemeID,
		HareketTip: "rezervasyon_iptal", Tutar: odeme.TaahhutTutari,
		Aciklama: &aciklamaRez, IslemiYapanID: &islemiYapanID,
	}); err != nil {
		return err
	}

	if odeme.FiiliTutar > 0 {
		aciklamaFiili := "Mutabakat: fiili harcama"
		if err = r.InsertButceHareketTx(tx, &models.ButceHareket{
			ProjeID: odeme.ProjeID, KalemID: odeme.KalemID,
			KaynakTip: "satinalma_mutabakat", KaynakID: odeme.OdemeID,
			HareketTip: "fiili_harcama", Tutar: -odeme.FiiliTutar,
			Aciklama: &aciklamaFiili, IslemiYapanID: &islemiYapanID,
		}); err != nil {
			return err
		}
	}

	return nil
}

// BeginTx yeni bir transaction başlatır.
func (r *SatinalmaRepository) BeginTx() (*sql.Tx, error) {
	return r.DB.Begin()
}

// ListPendingMutabakat mutabakat bekleyen (Onaylandı, henüz kapatılmamış) talepleri listeler.
func (r *SatinalmaRepository) ListPendingMutabakat() ([]models.SatinalmaTalebi, error) {
	query := `
		SELECT 
			st.talep_id, COALESCE(st.talep_no, '') AS talep_no, st.proje_id, st.uye_id, st.kalem_id, st.malzeme_adi, st.miktar, st.birim_fiyat, st.toplam_fiyat, st.durum, st.gerekce, st.red_nedeni, st.olusturma_tarihi, st.guncelleme_tarihi,
			st.revize_birim_fiyat, st.revize_toplam_fiyat, st.revizyon_gerekcesi, st.revize_eden_id, st.revizyon_tarihi,
			u.ad || ' ' || u.soyad AS uye_ad_soyad,
			COALESCE(u2.ad || ' ' || u2.soyad, '') AS revize_eden_ad_soyad,
			COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'), '') AS proje_baslik,
			COALESCE(p.proje_kodu, '') AS proje_kodu,
			b.aciklama AS kalem_aciklama,
			COALESCE(bk.kategori_adi, 'Belirtilmemiş') AS butce_kategori_adi,
			b.toplam_fiyat AS mevcut_butce
		FROM proje_satinalma_talebi st
		INNER JOIN uye u ON st.uye_id = u.uye_id
		LEFT JOIN uye u2 ON st.revize_eden_id = u2.uye_id
		INNER JOIN proje p ON st.proje_id = p.proje_id
		INNER JOIN proje_butce b ON st.kalem_id = b.kalem_id
		LEFT JOIN proje_butce_kategori bk ON b.kategori_id = bk.kategori_id
		WHERE st.durum = 'Onaylandı'
		  AND NOT EXISTS (
			SELECT 1 FROM proje_satinalma_odeme o
			WHERE o.talep_id = st.talep_id AND o.durum = 'onaylandi'
		  )
		ORDER BY st.olusturma_tarihi DESC
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("mutabakat bekleyen talepler listelenemedi: %w", err)
	}
	defer rows.Close()

	var talepler []models.SatinalmaTalebi
	for rows.Next() {
		var t models.SatinalmaTalebi
		var redNedeni sql.NullString
		err := rows.Scan(
			&t.TalepID, &t.TalepNo, &t.ProjeID, &t.UyeID, &t.KalemID, &t.MalzemeAdi, &t.Miktar, &t.BirimFiyat, &t.ToplamFiyat, &t.Durum, &t.Gerekce, &redNedeni, &t.OlusturmaTarihi, &t.GuncellemeTarihi,
			&t.RevizeBirimFiyat, &t.RevizeToplamFiyat, &t.RevizyonGerekcesi, &t.RevizeEdenID, &t.RevizyonTarihi,
			&t.UyeAdSoyad, &t.RevizeEdenAdSoyad, &t.ProjeBaslik, &t.ProjeKodu, &t.KalemAciklama, &t.ButceKategoriAdi, &t.MevcutButce,
		)
		if err != nil {
			return nil, fmt.Errorf("mutabakat satırı okunamadı: %w", err)
		}
		if redNedeni.Valid {
			val := redNedeni.String
			t.RedNedeni = &val
		}
		talepler = append(talepler, t)
	}
	return talepler, nil
}

// ListOdemelerByProje bir projenin mutabakat/ödeme kayıtlarını listeler.
func (r *SatinalmaRepository) ListOdemelerByProje(projeID int) ([]models.SatinalmaOdeme, error) {
	rows, err := r.DB.Query(`
		SELECT
			o.odeme_id, o.talep_id, o.talep_no, o.proje_id, o.kalem_id,
			o.taahhut_tutari, o.fiili_tutar, o.fark_tutari, o.fark_yonu,
			o.fatura_no, o.fatura_tarihi, o.odeme_tarihi, COALESCE(o.para_birimi, 'TRY'),
			o.evrak_yolu, o.durum, o.tto_uye_id, o.tto_gerekce, o.karar_tarihi,
			o.olusturma_tarihi, o.guncelleme_tarihi,
			COALESCE(st.malzeme_adi, ''),
			COALESCE(p.proje_kodu, ''),
			COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'), ''),
			COALESCE(u.ad || ' ' || u.soyad, ''),
			COALESCE(bk.kategori_adi, '')
		FROM proje_satinalma_odeme o
		INNER JOIN proje_satinalma_talebi st ON o.talep_id = st.talep_id
		INNER JOIN proje p ON o.proje_id = p.proje_id
		LEFT JOIN uye u ON o.tto_uye_id = u.uye_id
		LEFT JOIN proje_butce b ON o.kalem_id = b.kalem_id
		LEFT JOIN proje_butce_kategori bk ON b.kategori_id = bk.kategori_id
		WHERE o.proje_id = $1
		ORDER BY o.olusturma_tarihi DESC
	`, projeID)
	if err != nil {
		return nil, fmt.Errorf("ödeme kayıtları listelenemedi: %w", err)
	}
	defer rows.Close()

	var list []models.SatinalmaOdeme
	for rows.Next() {
		var o models.SatinalmaOdeme
		var faturaNo, evrak, gerekce sql.NullString
		var faturaTarihi, odemeTarihi, kararTarihi sql.NullTime
		var ttoID sql.NullInt64
		if err := rows.Scan(
			&o.OdemeID, &o.TalepID, &o.TalepNo, &o.ProjeID, &o.KalemID,
			&o.TaahhutTutari, &o.FiiliTutar, &o.FarkTutari, &o.FarkYonu,
			&faturaNo, &faturaTarihi, &odemeTarihi, &o.ParaBirimi,
			&evrak, &o.Durum, &ttoID, &gerekce, &kararTarihi,
			&o.OlusturmaTarihi, &o.GuncellemeTarihi,
			&o.MalzemeAdi, &o.ProjeKodu, &o.ProjeBaslik, &o.TtoAdSoyad, &o.ButceKategoriAdi,
		); err != nil {
			return nil, err
		}
		if faturaNo.Valid {
			o.FaturaNo = &faturaNo.String
		}
		if faturaTarihi.Valid {
			o.FaturaTarihi = &faturaTarihi.Time
		}
		if odemeTarihi.Valid {
			o.OdemeTarihi = &odemeTarihi.Time
		}
		if evrak.Valid {
			o.EvrakYolu = &evrak.String
		}
		if ttoID.Valid {
			id := int(ttoID.Int64)
			o.TtoUyeID = &id
		}
		if gerekce.Valid {
			o.TtoGerekce = &gerekce.String
		}
		if kararTarihi.Valid {
			o.KararTarihi = &kararTarihi.Time
		}
		list = append(list, o)
	}
	return list, nil
}

// RecordApprovalReservationTx onay anında rezervasyon ledger kaydı yazar.
func (r *SatinalmaRepository) RecordApprovalReservationTx(tx *sql.Tx, talep *models.SatinalmaTalebi, islemiYapanID int) error {
	tutar := talep.EffectiveAmount()
	aciklama := "Satın alma onayı: taahhüt rezervasyonu"
	return r.InsertButceHareketTx(tx, &models.ButceHareket{
		ProjeID: talep.ProjeID, KalemID: talep.KalemID,
		KaynakTip: "satinalma_onay", KaynakID: talep.TalepID,
		HareketTip: "rezervasyon", Tutar: -tutar,
		Aciklama: &aciklama, IslemiYapanID: &islemiYapanID,
	})
}
