package repository

import (
	"database/sql"
	"fmt"

	"bap_ai/backend/models"
)

// RaporRepository yapısı veritabanı rapor işlemlerini yönetir.
type RaporRepository struct {
	DB *sql.DB
}

// NewRaporRepository yeni bir RaporRepository oluşturur.
func NewRaporRepository(db *sql.DB) *RaporRepository {
	return &RaporRepository{DB: db}
}

// CreateAraRapor yeni bir ara rapor veya kesin sonuç raporu kaydeder.
// Türkçe Yorum: Yürütücünün yüklediği rapor dosya bilgilerini veritabanına yazar.
func (r *RaporRepository) CreateAraRapor(ar *models.ProjeAraRapor) error {
	query := `
		INSERT INTO proje_ara_rapor (
			proje_id, yukleyen_uye_id, rapor_turu, rapor_donemi,
			baslik, aciklama, dosya_url, orijinal_dosya_adi, durum
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'bekliyor')
		RETURNING rapor_id, olusturma_tarihi
	`
	err := r.DB.QueryRow(
		query,
		ar.ProjeID, ar.YukleyenUyeID, ar.RaporTuru, ar.RaporDonemi,
		ar.Baslik, ar.Aciklama, ar.DosyaURL, ar.OrijinalDosyaAdi,
	).Scan(&ar.RaporID, &ar.OlusturmaTarihi)
	return err
}

// GetAraRaporlarByProjeID belirli bir projeye ait tüm rapor teslimlerini listeler.
// Türkçe Yorum: Projenin ara rapor geçmişini ve onay durumlarını getirir.
func (r *RaporRepository) GetAraRaporlarByProjeID(projeID int) ([]models.ProjeAraRapor, error) {
	query := `
		SELECT r.rapor_id, r.proje_id, r.yukleyen_uye_id, r.rapor_turu, r.rapor_donemi,
		       r.baslik, COALESCE(r.aciklama, ''), r.dosya_url, COALESCE(r.orijinal_dosya_adi, ''),
		       r.durum, r.onaylayan_uye_id, COALESCE(r.onay_notu, ''), r.onay_tarihi,
		       r.olusturma_tarihi, r.guncelleme_tarihi,
		       COALESCE(u1.unvan || ' ' || u1.ad || ' ' || u1.soyad, '') AS yukleyen_ad,
		       COALESCE(u2.unvan || ' ' || u2.ad || ' ' || u2.soyad, '') AS onaylayan_ad
		FROM proje_ara_rapor r
		LEFT JOIN uye u1 ON r.yukleyen_uye_id = u1.uye_id
		LEFT JOIN uye u2 ON r.onaylayan_uye_id = u2.uye_id
		WHERE r.proje_id = $1
		ORDER BY r.rapor_donemi ASC, r.olusturma_tarihi DESC
	`
	rows, err := r.DB.Query(query, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.ProjeAraRapor
	for rows.Next() {
		var item models.ProjeAraRapor
		var onaylayanID sql.NullInt64
		var onayTarihi sql.NullTime

		err := rows.Scan(
			&item.RaporID, &item.ProjeID, &item.YukleyenUyeID, &item.RaporTuru, &item.RaporDonemi,
			&item.Baslik, &item.Aciklama, &item.DosyaURL, &item.OrijinalDosyaAdi,
			&item.Durum, &onaylayanID, &item.OnayNotu, &onayTarihi,
			&item.OlusturmaTarihi, &item.GuncellemeTarihi,
			&item.YukleyenAdSoyad, &item.OnaylayanAdSoyad,
		)
		if err != nil {
			return nil, err
		}
		if onaylayanID.Valid {
			id := int(onaylayanID.Int64)
			item.OnaylayanUyeID = &id
		}
		if onayTarihi.Valid {
			t := onayTarihi.Time
			item.OnayTarihi = &t
		}
		list = append(list, item)
	}
	return list, nil
}

// GetPendingRaporlar TTO veya Admin incelemesi bekleyen raporları getirir.
// Türkçe Yorum: Durumu 'bekliyor' olan raporları proje başlık bilgisiyle listeler.
func (r *RaporRepository) GetPendingRaporlar() ([]models.ProjeAraRapor, error) {
	query := `
		SELECT r.rapor_id, r.proje_id, r.yukleyen_uye_id, r.rapor_turu, r.rapor_donemi,
		       r.baslik, COALESCE(r.aciklama, ''), r.dosya_url, COALESCE(r.orijinal_dosya_adi, ''),
		       r.durum, r.olusturma_tarihi,
		       COALESCE(u1.unvan || ' ' || u1.ad || ' ' || u1.soyad, '') AS yukleyen_ad,
		       COALESCE(p.proje_kodu, '') AS proje_kodu,
		       COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'), '') AS proje_baslik,
		       COALESCE(pbt.bap_turu, '') AS bap_turu
		FROM proje_ara_rapor r
		INNER JOIN proje p ON r.proje_id = p.proje_id
		LEFT JOIN uye u1 ON r.yukleyen_uye_id = u1.uye_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		WHERE r.durum = 'bekliyor'
		ORDER BY r.olusturma_tarihi ASC
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.ProjeAraRapor
	for rows.Next() {
		var item models.ProjeAraRapor
		err := rows.Scan(
			&item.RaporID, &item.ProjeID, &item.YukleyenUyeID, &item.RaporTuru, &item.RaporDonemi,
			&item.Baslik, &item.Aciklama, &item.DosyaURL, &item.OrijinalDosyaAdi,
			&item.Durum, &item.OlusturmaTarihi,
			&item.YukleyenAdSoyad, &item.ProjeKodu, &item.ProjeBaslik, &item.BapTuru,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

// GetAraRaporByID tek bir rapor kaydını ID ile getirir.
func (r *RaporRepository) GetAraRaporByID(raporID int) (*models.ProjeAraRapor, error) {
	query := `
		SELECT r.rapor_id, r.proje_id, r.yukleyen_uye_id, r.rapor_turu, r.rapor_donemi,
		       r.baslik, COALESCE(r.aciklama, ''), r.dosya_url, COALESCE(r.orijinal_dosya_adi, ''),
		       r.durum, r.onaylayan_uye_id, COALESCE(r.onay_notu, ''), r.onay_tarihi,
		       r.olusturma_tarihi, r.guncelleme_tarihi
		FROM proje_ara_rapor r
		WHERE r.rapor_id = $1
	`
	var item models.ProjeAraRapor
	var onaylayanID sql.NullInt64
	var onayTarihi sql.NullTime

	err := r.DB.QueryRow(query, raporID).Scan(
		&item.RaporID, &item.ProjeID, &item.YukleyenUyeID, &item.RaporTuru, &item.RaporDonemi,
		&item.Baslik, &item.Aciklama, &item.DosyaURL, &item.OrijinalDosyaAdi,
		&item.Durum, &onaylayanID, &item.OnayNotu, &onayTarihi,
		&item.OlusturmaTarihi, &item.GuncellemeTarihi,
	)
	if err != nil {
		return nil, err
	}
	if onaylayanID.Valid {
		id := int(onaylayanID.Int64)
		item.OnaylayanUyeID = &id
	}
	if onayTarihi.Valid {
		t := onayTarihi.Time
		item.OnayTarihi = &t
	}
	return &item, nil
}

// UpdateRaporStatus rapor değerlendirme kararını kaydeder.
// Türkçe Yorum: TTO veya Komisyon tarafından verilen onay/revizyon/red kararlarını yazar.
func (r *RaporRepository) UpdateRaporStatus(raporID int, onaylayanUyeID int, durum string, onayNotu string) error {
	query := `
		UPDATE proje_ara_rapor
		SET durum = $1, onaylayan_uye_id = $2, onay_notu = $3, onay_tarihi = CURRENT_TIMESTAMP, guncelleme_tarihi = CURRENT_TIMESTAMP
		WHERE rapor_id = $4
	`
	res, err := r.DB.Exec(query, durum, onaylayanUyeID, onayNotu, raporID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("rapor bulunamadı")
	}
	return nil
}
