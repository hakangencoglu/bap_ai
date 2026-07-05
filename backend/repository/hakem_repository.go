package repository

import (
	"database/sql"
	"fmt"

	"bap_ai/backend/models"
)

// HakemRepository, hakem değerlendirme işlemlerini gerçekleştirir.
type HakemRepository struct {
	DB *sql.DB
}

// NewHakemRepository fonksiyonu yeni bir HakemRepository döner.
func NewHakemRepository(db *sql.DB) *HakemRepository {
	return &HakemRepository{DB: db}
}

// AssignRandomHakem, belirtilen sayıda rastgele hakemi projeye atar.
func (r *HakemRepository) AssignRandomHakem(projeID int, count int) error {
	// Çoklu rol desteği için hem doğrudan rol alanına hem de sistem_rol tablosuna bakılarak rastgele hakemler seçilir
	queryRandomHakem := `
		SELECT DISTINCT u.uye_id FROM uye u
		LEFT JOIN sistem_rol sr ON u.uye_id = sr.uye_id
		LEFT JOIN sistem_rol_tanimlama srt ON sr.sistem_rol_id = srt.rol_id
		WHERE (u.rol = 'hakem' OR srt.rol_adi = 'hakem') AND u.aktif_mi = true
		ORDER BY RANDOM() LIMIT $1
	`
	rows, err := r.DB.Query(queryRandomHakem, count)
	if err != nil {
		return err
	}
	defer rows.Close()

	var hakemIDs []int
	for rows.Next() {
		var uid int
		if err := rows.Scan(&uid); err != nil {
			return err
		}
		hakemIDs = append(hakemIDs, uid)
	}

	if len(hakemIDs) == 0 {
		return fmt.Errorf("atanacak aktif hakem bulunamadı")
	}

	for _, hid := range hakemIDs {
		// Atama durumu 'Atandı' olarak başlatılır (hakem henüz kabul etmedi)
		insertQuery := `
			INSERT INTO proje_degerlendirmeleri (proje_id, hakem_id, durum, atama_durumu)
			VALUES ($1, $2, 'Bekliyor', 'Atandı')
			ON CONFLICT (proje_id, hakem_id) DO NOTHING
		`
		r.DB.Exec(insertQuery, projeID, hid)
	}
	return nil
}

// GetProjelerByHakemID, bir hakeme atanmış tüm projeleri getirir
func (r *HakemRepository) GetProjelerByHakemID(hakemID int) ([]models.HakemProjeOzet, error) {
	query := `
		SELECT p.proje_id, COALESCE(p.proje_kodu, ''), COALESCE(p.baslik_tr, 'Başlıksız Proje'),
		       COALESCE(pbt.bap_turu, 'Münferit'), COALESCE(pd.durum_adi, 'taslak'),
		       pdeg.durum, COALESCE(pdeg.atama_durumu, 'Kabul Edildi'),
		       pdeg.puan, TO_CHAR(pdeg.olusturma_tarihi, 'DD.MM.YYYY')
		FROM proje p
		INNER JOIN proje_degerlendirmeleri pdeg ON p.proje_id = pdeg.proje_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pdeg.hakem_id = $1
		ORDER BY pdeg.olusturma_tarihi DESC
	`
	rows, err := r.DB.Query(query, hakemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projeler []models.HakemProjeOzet
	for rows.Next() {
		var p models.HakemProjeOzet
		if err := rows.Scan(&p.ProjeID, &p.ProjeKodu, &p.BaslikTr, &p.BapTuru, &p.DurumAdi, &p.HakemDurum, &p.AtamaDurumu, &p.Puan, &p.Tarih); err != nil {
			return nil, err
		}
		projeler = append(projeler, p)
	}
	return projeler, nil
}

// UpdateAtamaKarar, hakemin atamayı kabul veya reddetmesini veritabanına yazar
func (r *HakemRepository) UpdateAtamaKarar(hakemID, projeID int, karar, redNedeni string) error {
	query := `
		UPDATE proje_degerlendirmeleri
		SET atama_durumu = $1, red_nedeni = $2, karar_tarihi = CURRENT_TIMESTAMP, guncelleme_tarihi = CURRENT_TIMESTAMP
		WHERE proje_id = $3 AND hakem_id = $4 AND atama_durumu = 'Atandı'
	`
	res, err := r.DB.Exec(query, karar, redNedeni, projeID, hakemID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("atama kaydı bulunamadı veya zaten karar verilmiş")
	}
	return nil
}

// SubmitDegerlendirme, hakemin yaptığı değerlendirmeyi DB'ye kaydeder
func (r *HakemRepository) SubmitDegerlendirme(hakemID int, req models.DegerlendirmeRequest) error {
	// Sadece atamayı kabul etmiş hakemler değerlendirme yapabilir
	// Türkçe Yorum: Veri tabanı tutarlılığını korumak için işlemi transaction (tx) ile yapıyoruz.
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Önce degerlendirme_id bilgisini alıyoruz
	var degerlendirmeID int
	getIDQuery := `
		SELECT degerlendirme_id FROM proje_degerlendirmeleri 
		WHERE proje_id = $1 AND hakem_id = $2 AND atama_durumu = 'Kabul Edildi'
	`
	err = tx.QueryRow(getIDQuery, req.ProjeID, hakemID).Scan(&degerlendirmeID)
	if err != nil {
		return fmt.Errorf("değerlendirme kaydı bulunamadı veya hakem atamayı henüz kabul etmemiş: %w", err)
	}

	// 2. Ana değerlendirme tablosunu güncelliyoruz
	updateQuery := `
		UPDATE proje_degerlendirmeleri
		SET puan = $1, yorum = $2, durum = $3, guncelleme_tarihi = CURRENT_TIMESTAMP
		WHERE degerlendirme_id = $4
	`
	res, err := tx.Exec(updateQuery, req.Puan, req.Yorum, req.Durum, degerlendirmeID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("değerlendirme kaydı güncellenemedi")
	}

	// 3. Alt soru cevaplarını kaydediyoruz
	if len(req.Cevaplar) > 0 {
		insertAnsQuery := `
			INSERT INTO proje_degerlendirme_soru_cevaplari (degerlendirme_id, baslik_id, soru_id, puan_degeri)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (degerlendirme_id, baslik_id, soru_id) 
			DO UPDATE SET puan_degeri = EXCLUDED.puan_degeri
		`
		for _, c := range req.Cevaplar {
			_, err = tx.Exec(insertAnsQuery, degerlendirmeID, c.BaslikID, c.SoruID, c.PuanDegeri)
			if err != nil {
				return fmt.Errorf("alt soru cevabı kaydedilemedi (soru_id: %d): %w", c.SoruID, err)
			}
		}
	}

	// İşlemi tamamla (commit)
	return tx.Commit()
}

// GetDegerlendirmeQuestions, değerlendirme formundaki başlıkları ve bunlara bağlı alt soruları getirir
func (r *HakemRepository) GetDegerlendirmeQuestions() ([]models.HakemDegerlendirmeBaslik, error) {
	// 1. Önce başlıkları çekiyoruz
	baslikQuery := `
		SELECT baslik_id, baslik_adi, maksimum_puan, sira_no 
		FROM hakem_degerlendirme_basliklari 
		ORDER BY sira_no
	`
	baslikRows, err := r.DB.Query(baslikQuery)
	if err != nil {
		return nil, err
	}
	defer baslikRows.Close()

	var basliklar []models.HakemDegerlendirmeBaslik
	for baslikRows.Next() {
		var b models.HakemDegerlendirmeBaslik
		if err := baslikRows.Scan(&b.BaslikID, &b.BaslikAdi, &b.MaksimumPuan, &b.SiraNo); err != nil {
			return nil, err
		}
		// Go modelleri kuralına göre slice'ı boş başlatıyoruz nil olmaması için
		b.Sorular = []models.HakemDegerlendirmeSoru{}
		basliklar = append(basliklar, b)
	}

	// 2. Her başlık için soruları çekip eşleştiriyoruz
	for i := range basliklar {
		soruQuery := `
			SELECT s.soru_id, s.soru_kodu, s.soru_metni, s.maksimum_puan, s.sira_no 
			FROM hakem_degerlendirme_sorulari s
			INNER JOIN hakem_degerlendirme_baslik_soru bs ON s.soru_id = bs.soru_id
			WHERE bs.baslik_id = $1
			ORDER BY s.sira_no
		`
		soruRows, err := r.DB.Query(soruQuery, basliklar[i].BaslikID)
		if err != nil {
			return nil, err
		}
		defer soruRows.Close()

		for soruRows.Next() {
			var s models.HakemDegerlendirmeSoru
			if err := soruRows.Scan(&s.SoruID, &s.SoruKodu, &s.SoruMetni, &s.MaksimumPuan, &s.SiraNo); err != nil {
				return nil, err
			}
			basliklar[i].Sorular = append(basliklar[i].Sorular, s)
		}
	}

	return basliklar, nil
}

// GetAllDegerlendirmeByProjeID, bir projenin tüm hakem değerlendirmelerini getirir.
func (r *HakemRepository) GetAllDegerlendirmeByProjeID(projeID int) ([]models.ProjeDegerlendirme, error) {
	query := `
		SELECT degerlendirme_id, proje_id, hakem_id, COALESCE(puan, 0), COALESCE(yorum, ''), durum, COALESCE(atama_durumu, 'Atandı')
		FROM proje_degerlendirmeleri WHERE proje_id = $1
	`
	rows, err := r.DB.Query(query, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var degs []models.ProjeDegerlendirme
	for rows.Next() {
		var d models.ProjeDegerlendirme
		if err := rows.Scan(&d.DegerlendirmeID, &d.ProjeID, &d.HakemID, &d.Puan, &d.Yorum, &d.Durum, &d.AtamaDurumu); err != nil {
			return nil, err
		}
		degs = append(degs, d)
	}
	return degs, nil
}

// IsHakemAssigned, hakemin projeye atanıp atanmadığını kontrol eder.
// Türkçe Yorum: Hakemin ilgili projeyi görüntülemeye/değerlendirmeye yetkisi olup olmadığını kontrol eder.
func (r *HakemRepository) IsHakemAssigned(hakemID int, projeID int) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM proje_degerlendirmeleri 
			WHERE proje_id = $1 AND hakem_id = $2 AND atama_durumu IN ('Kabul Edildi', 'Atandı')
		)
	`
	err := r.DB.QueryRow(query, projeID, hakemID).Scan(&exists)
	return exists, err
}
