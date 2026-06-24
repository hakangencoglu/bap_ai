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

// CreatePurchaseRequest veritabanına yeni bir satın alma talebi ekler.
// Türkçe Yorum: Akademisyen tarafından gönderilen yeni satın alma talebini satinalma_talebi tablosuna yazar.
func (r *SatinalmaRepository) CreatePurchaseRequest(req *models.SatinalmaTalebi) error {
	query := `
		INSERT INTO satinalma_talebi (proje_id, uye_id, kalem_id, malzeme_adi, miktar, birim_fiyat, toplam_fiyat, durum, gerekce)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'Beklemede', $8)
		RETURNING talep_id, olusturma_tarihi, guncelleme_tarihi
	`
	toplamTutar := float64(req.Miktar) * req.BirimFiyat
	err := r.DB.QueryRow(query, req.ProjeID, req.UyeID, req.KalemID, req.MalzemeAdi, req.Miktar, req.BirimFiyat, toplamTutar, req.Gerekce).
		Scan(&req.TalepID, &req.OlusturmaTarihi, &req.GuncellemeTarihi)
	if err != nil {
		return fmt.Errorf("satın alma talebi eklenirken veritabanı hatası: %w", err)
	}
	req.ToplamFiyat = toplamTutar
	req.Durum = "Beklemede"
	return nil
}

// GetPurchaseRequestsByProject belirli bir projeye ait tüm satın alma taleplerini listeler.
// Türkçe Yorum: Akademisyenin kendi projesine ait geçmiş ve bekleyen tüm satın alma taleplerini detaylarıyla getirir.
func (r *SatinalmaRepository) GetPurchaseRequestsByProject(projeID int) ([]models.SatinalmaTalebi, error) {
	query := `
		SELECT 
			st.talep_id, st.proje_id, st.uye_id, st.kalem_id, st.malzeme_adi, st.miktar, st.birim_fiyat, st.toplam_fiyat, st.durum, st.gerekce, st.red_nedeni, st.olusturma_tarihi, st.guncelleme_tarihi,
			u.ad || ' ' || u.soyad AS uye_ad_soyad,
			p.baslik_tr AS proje_baslik,
			b.aciklama AS kalem_aciklama,
			COALESCE(bk.kategori_adi, 'Belirtilmemiş') AS butce_kategori_adi,
			b.toplam_fiyat AS mevcut_butce
		FROM satinalma_talebi st
		INNER JOIN uye u ON st.uye_id = u.uye_id
		INNER JOIN proje p ON st.proje_id = p.proje_id
		INNER JOIN butce b ON st.kalem_id = b.kalem_id
		LEFT JOIN butce_kategori bk ON b.kategori_id = bk.kategori_id
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
			&t.TalepID, &t.ProjeID, &t.UyeID, &t.KalemID, &t.MalzemeAdi, &t.Miktar, &t.BirimFiyat, &t.ToplamFiyat, &t.Durum, &t.Gerekce, &redNedeni, &t.OlusturmaTarihi, &t.GuncellemeTarihi,
			&t.UyeAdSoyad, &t.ProjeBaslik, &t.KalemAciklama, &t.ButceKategoriAdi, &t.MevcutButce,
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
			st.talep_id, st.proje_id, st.uye_id, st.kalem_id, st.malzeme_adi, st.miktar, st.birim_fiyat, st.toplam_fiyat, st.durum, st.gerekce, st.red_nedeni, st.olusturma_tarihi, st.guncelleme_tarihi,
			u.ad || ' ' || u.soyad AS uye_ad_soyad,
			p.baslik_tr AS proje_baslik,
			b.aciklama AS kalem_aciklama,
			COALESCE(bk.kategori_adi, 'Belirtilmemiş') AS butce_kategori_adi,
			b.toplam_fiyat AS mevcut_butce
		FROM satinalma_talebi st
		INNER JOIN uye u ON st.uye_id = u.uye_id
		INNER JOIN proje p ON st.proje_id = p.proje_id
		INNER JOIN butce b ON st.kalem_id = b.kalem_id
		LEFT JOIN butce_kategori bk ON b.kategori_id = bk.kategori_id
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
			&t.TalepID, &t.ProjeID, &t.UyeID, &t.KalemID, &t.MalzemeAdi, &t.Miktar, &t.BirimFiyat, &t.ToplamFiyat, &t.Durum, &t.Gerekce, &redNedeni, &t.OlusturmaTarihi, &t.GuncellemeTarihi,
			&t.UyeAdSoyad, &t.ProjeBaslik, &t.KalemAciklama, &t.ButceKategoriAdi, &t.MevcutButce,
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
// Türkçe Yorum: Talebi onaylar veya girilen gerekçe ile reddeder. Güncelleme tarihini güncel zaman yapar.
func (r *SatinalmaRepository) UpdatePurchaseStatus(talepID int, status string, redNedeni string) error {
	query := `
		UPDATE satinalma_talebi
		SET durum = $1, red_nedeni = $2, guncelleme_tarihi = $3
		WHERE talep_id = $4
	`
	var redVal interface{} = nil
	if redNedeni != "" {
		redVal = redNedeni
	}
	_, err := r.DB.Exec(query, status, redVal, time.Now(), talepID)
	if err != nil {
		return fmt.Errorf("satın alma talebi güncellenirken hata: %w", err)
	}
	return nil
}

// GetRemainingBudget bir bütçe kaleminin kalan bütçesini sorgular.
// Türkçe Yorum: Bütçe kaleminin toplam bütçe değerinden, o kalem için onaylanmış satın alma taleplerinin tutarlarını çıkartır.
func (r *SatinalmaRepository) GetRemainingBudget(projeID int, kalemID int) (float64, error) {
	var totalBudget float64
	err := r.DB.QueryRow(`
		SELECT toplam_fiyat FROM butce 
		WHERE proje_id = $1 AND kalem_id = $2
	`, projeID, kalemID).Scan(&totalBudget)
	if err != nil {
		return 0, fmt.Errorf("bütçe kalem bütçesi bulunamadı: %w", err)
	}

	var totalSpent float64
	err = r.DB.QueryRow(`
		SELECT COALESCE(SUM(toplam_fiyat), 0) FROM satinalma_talebi 
		WHERE proje_id = $1 AND kalem_id = $2 AND durum = 'Onaylandı'
	`, projeID, kalemID).Scan(&totalSpent)
	if err != nil {
		return 0, fmt.Errorf("bütçe harcama toplamı hesaplanamadı: %w", err)
	}

	return totalBudget - totalSpent, nil
}

// GetPurchaseRequestByID satın alma talebini getirir.
// Türkçe Yorum: Satın alma talebini ID bazında getirmek için kullanılır.
func (r *SatinalmaRepository) GetPurchaseRequestByID(talepID int) (*models.SatinalmaTalebi, error) {
	query := `
		SELECT talep_id, proje_id, uye_id, kalem_id, malzeme_adi, miktar, birim_fiyat, toplam_fiyat, durum, gerekce, red_nedeni, olusturma_tarihi, guncelleme_tarihi
		FROM satinalma_talebi
		WHERE talep_id = $1
	`
	var t models.SatinalmaTalebi
	var redNedeni sql.NullString
	err := r.DB.QueryRow(query, talepID).Scan(
		&t.TalepID, &t.ProjeID, &t.UyeID, &t.KalemID, &t.MalzemeAdi, &t.Miktar, &t.BirimFiyat, &t.ToplamFiyat, &t.Durum, &t.Gerekce, &redNedeni, &t.OlusturmaTarihi, &t.GuncellemeTarihi,
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
