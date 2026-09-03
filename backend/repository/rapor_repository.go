package repository

import (
	"database/sql"
	"fmt"
	"time"

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

// GetTTORaporTakipMatrisi TTO yetkilisinin tüm yürürlükteki projelerin ara rapor teslim durumlarını izlemesini sağlar.
// Türkçe Yorum: Zamanı geçen (gecikmiş), bekleyen, onaylanan ve yaklaşan ara rapor durumlarını tek bir matriste listeler.
func (r *RaporRepository) GetTTORaporTakipMatrisi() ([]models.TTORaporTakipItem, error) {
	query := `
		SELECT 
			p.proje_id,
			p.proje_kodu,
			COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), p.proje_kodu) AS proje_baslik,
			COALESCE(u.unvan || ' ' || u.ad || ' ' || u.soyad, '') AS yurutucu_ad_soyad,
			COALESCE(u.eposta, '') AS yurutucu_eposta,
			COALESCE(pbt.bap_turu, 'BAP') AS bap_turu,
			ps.olusturma_tarihi AS baslangic_tarihi,
			ar.rapor_id,
			ar.durum AS rapor_durum,
			ar.dosya_url,
			ar.olusturma_tarihi AS yuklenme_tarihi
		FROM proje p
		JOIN proje_durum pd ON p.durum_id = pd.durum_id
		LEFT JOIN uye u ON p.koordinator_id = u.uye_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN (
			SELECT proje_id, MAX(olusturma_tarihi) AS olusturma_tarihi
			FROM proje_surec_gecmisi
			WHERE hedef_durum = 'yururlukte'
			GROUP BY proje_id
		) ps ON p.proje_id = ps.proje_id
		LEFT JOIN proje_ara_rapor ar ON p.proje_id = ar.proje_id
		WHERE pd.durum_adi IN ('yururlukte', 'tamamlandi')
		ORDER BY p.proje_id DESC, ar.rapor_donemi DESC
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.TTORaporTakipItem
	for rows.Next() {
		var item models.TTORaporTakipItem
		var baslangicTarihi sql.NullTime
		var raporID sql.NullInt64
		var dbRaporDurum, dosyaURL sql.NullString
		var yuklenmeTarihi sql.NullTime

		err := rows.Scan(
			&item.ProjeID, &item.ProjeKodu, &item.ProjeBaslik,
			&item.YurutucuAdSoyad, &item.YurutucuEposta, &item.BapTuru,
			&baslangicTarihi, &raporID, &dbRaporDurum, &dosyaURL, &yuklenmeTarihi,
		)
		if err != nil {
			return nil, err
		}

		if baslangicTarihi.Valid {
			t := baslangicTarihi.Time
			item.BaslangicTarihi = &t
			// 6 aylık periyod hesabı
			monthsSinceStart := int(time.Since(t).Hours() / (24 * 30))
			donem := (monthsSinceStart / 6) + 1
			if donem < 1 {
				donem = 1
			}
			item.HesaplananDonem = donem
			sonTeslim := t.AddDate(0, donem*6, 0)
			item.SonTeslimTarihi = &sonTeslim

			// Rapor durum belirleme
			if dbRaporDurum.Valid && dbRaporDurum.String != "" {
				item.RaporDurumu = dbRaporDurum.String // bekliyor, onaylandi, revizyon, reddedildi
			} else {
				if time.Now().After(sonTeslim) {
					item.RaporDurumu = "gecikmis"
				} else if time.Until(sonTeslim).Hours() < 30*24 { // Son 30 gün
					item.RaporDurumu = "yaklasiyor"
				} else {
					item.RaporDurumu = "beklenmiyor"
				}
			}
		} else {
			item.HesaplananDonem = 1
			if dbRaporDurum.Valid {
				item.RaporDurumu = dbRaporDurum.String
			} else {
				item.RaporDurumu = "beklenmiyor"
			}
		}

		if raporID.Valid {
			id := int(raporID.Int64)
			item.YuklenenRaporID = &id
		}
		if dosyaURL.Valid {
			item.YuklenenDosyaURL = dosyaURL.String
		}
		if yuklenmeTarihi.Valid {
			yt := yuklenmeTarihi.Time
			item.YuklenmeTarihi = &yt
		}

		list = append(list, item)
	}

	return list, nil
}
