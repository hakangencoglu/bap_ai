package repository

import (
	"database/sql"
	"fmt"
	"time"

	"bap_ai/backend/models"
)

// TalepRepository, proje talep tabloları için veri erişim katmanıdır.
// Türkçe Yorum: Her talep türü için oluşturma, listeleme ve onay/red fonksiyonlarını barındırır.
type TalepRepository struct {
	DB *sql.DB
}

// NewTalepRepository yeni bir TalepRepository nesnesi döner.
func NewTalepRepository(db *sql.DB) *TalepRepository {
	return &TalepRepository{DB: db}
}

// talepNoUret, talep türüne göre benzersiz talep numarası üretir.
// Örn: BAP-2025-001-EKSURE-3
func (r *TalepRepository) talepNoUret(tx *sql.Tx, projeID int, tipKodu string) (string, error) {
	var projeKodu sql.NullString
	err := tx.QueryRow(`SELECT proje_kodu FROM proje WHERE proje_id = $1`, projeID).Scan(&projeKodu)
	if err != nil {
		return "", fmt.Errorf("proje kodu alınamadı: %w", err)
	}
	kodu := fmt.Sprintf("BAP-PROJE-%d", projeID)
	if projeKodu.Valid && projeKodu.String != "" {
		kodu = projeKodu.String
	}
	yil := time.Now().Year()
	return fmt.Sprintf("%s-%d-%s-%d", kodu, yil, tipKodu, time.Now().UnixNano()%10000), nil
}

// ---- 1) Ek Süre ----

// CreateEkSure, yeni bir ek süre talebi oluşturur.
func (r *TalepRepository) CreateEkSure(t *models.TalepEkSure) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Türkçe Yorum: Benzersiz talep numarası üretilir.
	no, err := r.talepNoUret(tx, t.ProjeID, "EKSURE")
	if err != nil {
		return err
	}
	t.TalepNo = no

	err = tx.QueryRow(`
		INSERT INTO proje_talep_ek_sure (proje_id, uye_id, talep_no, ek_sure_ay, gerekce)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, olusturma_tarihi`,
		t.ProjeID, t.UyeID, t.TalepNo, t.EkSureAy, t.Gerekce,
	).Scan(&t.ID, &t.OlusturmaTarihi)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetEkSureByProje, projeye ait ek süre taleplerini listeler.
func (r *TalepRepository) GetEkSureByProje(projeID int) ([]models.TalepEkSure, error) {
	rows, err := r.DB.Query(`
		SELECT t.id, t.proje_id, t.uye_id, t.talep_no, t.ek_sure_ay,
		       t.gerekce, t.durum, COALESCE(t.red_notu,''), t.olusturma_tarihi,
		       p.proje_kodu, (SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'),
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad
		FROM proje_talep_ek_sure t
		JOIN proje p ON p.proje_id = t.proje_id
		JOIN uye u   ON u.uye_id   = t.uye_id
		WHERE t.proje_id = $1
		ORDER BY t.olusturma_tarihi DESC`, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.TalepEkSure
	for rows.Next() {
		var item models.TalepEkSure
		rows.Scan(&item.ID, &item.ProjeID, &item.UyeID, &item.TalepNo, &item.EkSureAy,
			&item.Gerekce, &item.Durum, &item.RedNotu, &item.OlusturmaTarihi,
			&item.ProjeKodu, &item.ProjeBaslik, &item.TalepEdenAd)
		list = append(list, item)
	}
	return list, nil
}

// UpdateEkSureDurum, ek süre talebinin durumunu günceller (onay/red).
func (r *TalepRepository) UpdateEkSureDurum(id int, durum, redNotu string) error {
	_, err := r.DB.Exec(`
		UPDATE proje_talep_ek_sure SET durum=$1, red_notu=$2, guncelleme_tarihi=NOW()
		WHERE id=$3`, durum, redNotu, id)
	return err
}

// ---- 2) Ek Bütçe ----

// CreateEkButce, yeni bir ek bütçe talebi oluşturur.
func (r *TalepRepository) CreateEkButce(t *models.TalepEkButce) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	no, err := r.talepNoUret(tx, t.ProjeID, "EKBUTCE")
	if err != nil {
		return err
	}
	t.TalepNo = no
	err = tx.QueryRow(`
		INSERT INTO proje_talep_ek_butce (proje_id, uye_id, talep_no, butce_kalemi, tutar_tl, gerekce)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, olusturma_tarihi`,
		t.ProjeID, t.UyeID, t.TalepNo, t.ButceKalemi, t.TutarTL, t.Gerekce,
	).Scan(&t.ID, &t.OlusturmaTarihi)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetEkButceByProje, projeye ait ek bütçe taleplerini listeler.
func (r *TalepRepository) GetEkButceByProje(projeID int) ([]models.TalepEkButce, error) {
	rows, err := r.DB.Query(`
		SELECT t.id, t.proje_id, t.uye_id, t.talep_no, t.butce_kalemi, t.tutar_tl,
		       t.gerekce, t.durum, COALESCE(t.red_notu,''), t.olusturma_tarihi,
		       p.proje_kodu, (SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'),
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad
		FROM proje_talep_ek_butce t
		JOIN proje p ON p.proje_id = t.proje_id
		JOIN uye u   ON u.uye_id   = t.uye_id
		WHERE t.proje_id = $1 ORDER BY t.olusturma_tarihi DESC`, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.TalepEkButce
	for rows.Next() {
		var item models.TalepEkButce
		rows.Scan(&item.ID, &item.ProjeID, &item.UyeID, &item.TalepNo, &item.ButceKalemi,
			&item.TutarTL, &item.Gerekce, &item.Durum, &item.RedNotu, &item.OlusturmaTarihi,
			&item.ProjeKodu, &item.ProjeBaslik, &item.TalepEdenAd)
		list = append(list, item)
	}
	return list, nil
}

// UpdateEkButceDurum, ek bütçe talebinin durumunu günceller.
func (r *TalepRepository) UpdateEkButceDurum(id int, durum, redNotu string) error {
	_, err := r.DB.Exec(`UPDATE proje_talep_ek_butce SET durum=$1, red_notu=$2, guncelleme_tarihi=NOW() WHERE id=$3`, durum, redNotu, id)
	return err
}

// ---- 3) Fasıl Aktarımı ----

// CreateFasilAktarimi, yeni bir fasıl aktarımı talebi oluşturur.
func (r *TalepRepository) CreateFasilAktarimi(t *models.TalepFasilAktarimi) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	no, err := r.talepNoUret(tx, t.ProjeID, "FASIL")
	if err != nil {
		return err
	}
	t.TalepNo = no
	err = tx.QueryRow(`
		INSERT INTO proje_talep_fasil_aktarimi (proje_id, uye_id, talep_no, kaynak_kalem, hedef_kalem, tutar_tl, gerekce)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, olusturma_tarihi`,
		t.ProjeID, t.UyeID, t.TalepNo, t.KaynakKalem, t.HedefKalem, t.TutarTL, t.Gerekce,
	).Scan(&t.ID, &t.OlusturmaTarihi)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetFasilAktarimiByProje, projeye ait fasıl aktarımı taleplerini listeler.
func (r *TalepRepository) GetFasilAktarimiByProje(projeID int) ([]models.TalepFasilAktarimi, error) {
	rows, err := r.DB.Query(`
		SELECT t.id, t.proje_id, t.uye_id, t.talep_no, t.kaynak_kalem, t.hedef_kalem, t.tutar_tl,
		       t.gerekce, t.durum, COALESCE(t.red_notu,''), t.olusturma_tarihi,
		       p.proje_kodu, (SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'),
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad
		FROM proje_talep_fasil_aktarimi t
		JOIN proje p ON p.proje_id = t.proje_id
		JOIN uye u   ON u.uye_id   = t.uye_id
		WHERE t.proje_id = $1 ORDER BY t.olusturma_tarihi DESC`, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.TalepFasilAktarimi
	for rows.Next() {
		var item models.TalepFasilAktarimi
		rows.Scan(&item.ID, &item.ProjeID, &item.UyeID, &item.TalepNo, &item.KaynakKalem,
			&item.HedefKalem, &item.TutarTL, &item.Gerekce, &item.Durum, &item.RedNotu,
			&item.OlusturmaTarihi, &item.ProjeKodu, &item.ProjeBaslik, &item.TalepEdenAd)
		list = append(list, item)
	}
	return list, nil
}

// UpdateFasilAktarimiDurum, fasıl aktarımı talebinin durumunu günceller.
func (r *TalepRepository) UpdateFasilAktarimiDurum(id int, durum, redNotu string) error {
	_, err := r.DB.Exec(`UPDATE proje_talep_fasil_aktarimi SET durum=$1, red_notu=$2, guncelleme_tarihi=NOW() WHERE id=$3`, durum, redNotu, id)
	return err
}

// ---- 4) Araştırmacı ----

// CreateArastirmaci, yeni bir araştırmacı ekleme/çıkarma talebi oluşturur.
func (r *TalepRepository) CreateArastirmaci(t *models.TalepArastirmaci) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	no, err := r.talepNoUret(tx, t.ProjeID, "ARAS")
	if err != nil {
		return err
	}
	t.TalepNo = no
	err = tx.QueryRow(`
		INSERT INTO proje_talep_arastirmaci (proje_id, uye_id, talep_no, islem_turu, arastirmaci_adi, gerekce)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, olusturma_tarihi`,
		t.ProjeID, t.UyeID, t.TalepNo, t.IslemTuru, t.ArastirmaciAdi, t.Gerekce,
	).Scan(&t.ID, &t.OlusturmaTarihi)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetArastirmaciByProje, projeye ait araştırmacı değişikliği taleplerini listeler.
func (r *TalepRepository) GetArastirmaciByProje(projeID int) ([]models.TalepArastirmaci, error) {
	rows, err := r.DB.Query(`
		SELECT t.id, t.proje_id, t.uye_id, t.talep_no, t.islem_turu, t.arastirmaci_adi,
		       t.gerekce, t.durum, COALESCE(t.red_notu,''), t.olusturma_tarihi,
		       p.proje_kodu, (SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'),
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad
		FROM proje_talep_arastirmaci t
		JOIN proje p ON p.proje_id = t.proje_id
		JOIN uye u   ON u.uye_id   = t.uye_id
		WHERE t.proje_id = $1 ORDER BY t.olusturma_tarihi DESC`, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.TalepArastirmaci
	for rows.Next() {
		var item models.TalepArastirmaci
		rows.Scan(&item.ID, &item.ProjeID, &item.UyeID, &item.TalepNo, &item.IslemTuru,
			&item.ArastirmaciAdi, &item.Gerekce, &item.Durum, &item.RedNotu,
			&item.OlusturmaTarihi, &item.ProjeKodu, &item.ProjeBaslik, &item.TalepEdenAd)
		list = append(list, item)
	}
	return list, nil
}

// UpdateArastirmaciDurum günceller.
func (r *TalepRepository) UpdateArastirmaciDurum(id int, durum, redNotu string) error {
	_, err := r.DB.Exec(`UPDATE proje_talep_arastirmaci SET durum=$1, red_notu=$2, guncelleme_tarihi=NOW() WHERE id=$3`, durum, redNotu, id)
	return err
}

// ---- 5) Bursiyer ----

// CreateBursiyer, yeni bir bursiyer işlem talebi oluşturur.
func (r *TalepRepository) CreateBursiyer(t *models.TalepBursiyer) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	no, err := r.talepNoUret(tx, t.ProjeID, "BURS")
	if err != nil {
		return err
	}
	t.TalepNo = no
	err = tx.QueryRow(`
		INSERT INTO proje_talep_bursiyer (proje_id, uye_id, talep_no, bursiyer_kimlik, bursiyer_adi, islem_turu, gerekce)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, olusturma_tarihi`,
		t.ProjeID, t.UyeID, t.TalepNo, t.BursiyerKimlik, t.BursiyerAdi, t.IslemTuru, t.Gerekce,
	).Scan(&t.ID, &t.OlusturmaTarihi)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetBursiyerByProje listeler.
func (r *TalepRepository) GetBursiyerByProje(projeID int) ([]models.TalepBursiyer, error) {
	rows, err := r.DB.Query(`
		SELECT t.id, t.proje_id, t.uye_id, t.talep_no, t.bursiyer_kimlik, t.bursiyer_adi, t.islem_turu,
		       t.gerekce, t.durum, COALESCE(t.red_notu,''), t.olusturma_tarihi,
		       p.proje_kodu, (SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'),
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad
		FROM proje_talep_bursiyer t
		JOIN proje p ON p.proje_id = t.proje_id
		JOIN uye u   ON u.uye_id   = t.uye_id
		WHERE t.proje_id = $1 ORDER BY t.olusturma_tarihi DESC`, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.TalepBursiyer
	for rows.Next() {
		var item models.TalepBursiyer
		rows.Scan(&item.ID, &item.ProjeID, &item.UyeID, &item.TalepNo, &item.BursiyerKimlik,
			&item.BursiyerAdi, &item.IslemTuru, &item.Gerekce, &item.Durum, &item.RedNotu,
			&item.OlusturmaTarihi, &item.ProjeKodu, &item.ProjeBaslik, &item.TalepEdenAd)
		list = append(list, item)
	}
	return list, nil
}

// UpdateBursiyerDurum günceller.
func (r *TalepRepository) UpdateBursiyerDurum(id int, durum, redNotu string) error {
	_, err := r.DB.Exec(`UPDATE proje_talep_bursiyer SET durum=$1, red_notu=$2, guncelleme_tarihi=NOW() WHERE id=$3`, durum, redNotu, id)
	return err
}

// ---- 6) Proje İptali ----

// CreateProjeIptali, proje iptali talebi oluşturur.
func (r *TalepRepository) CreateProjeIptali(t *models.TalepProjeIptali) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	no, err := r.talepNoUret(tx, t.ProjeID, "IPTAL")
	if err != nil {
		return err
	}
	t.TalepNo = no
	err = tx.QueryRow(`
		INSERT INTO proje_talep_proje_iptali (proje_id, uye_id, talep_no, gerekce)
		VALUES ($1,$2,$3,$4)
		RETURNING id, olusturma_tarihi`,
		t.ProjeID, t.UyeID, t.TalepNo, t.Gerekce,
	).Scan(&t.ID, &t.OlusturmaTarihi)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetProjeIptaliByProje listeler.
func (r *TalepRepository) GetProjeIptaliByProje(projeID int) ([]models.TalepProjeIptali, error) {
	rows, err := r.DB.Query(`
		SELECT t.id, t.proje_id, t.uye_id, t.talep_no,
		       t.gerekce, t.durum, COALESCE(t.red_notu,''), t.olusturma_tarihi,
		       p.proje_kodu, (SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'),
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad
		FROM proje_talep_proje_iptali t
		JOIN proje p ON p.proje_id = t.proje_id
		JOIN uye u   ON u.uye_id   = t.uye_id
		WHERE t.proje_id = $1 ORDER BY t.olusturma_tarihi DESC`, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.TalepProjeIptali
	for rows.Next() {
		var item models.TalepProjeIptali
		rows.Scan(&item.ID, &item.ProjeID, &item.UyeID, &item.TalepNo,
			&item.Gerekce, &item.Durum, &item.RedNotu, &item.OlusturmaTarihi,
			&item.ProjeKodu, &item.ProjeBaslik, &item.TalepEdenAd)
		list = append(list, item)
	}
	return list, nil
}

// UpdateProjeIptaliDurum günceller.
func (r *TalepRepository) UpdateProjeIptaliDurum(id int, durum, redNotu string) error {
	_, err := r.DB.Exec(`UPDATE proje_talep_proje_iptali SET durum=$1, red_notu=$2, guncelleme_tarihi=NOW() WHERE id=$3`, durum, redNotu, id)
	return err
}

// ---- 7) Bilgi Değişimi ----

// CreateBilgiDegisimi, proje bilgi değişimi talebi oluşturur.
func (r *TalepRepository) CreateBilgiDegisimi(t *models.TalepBilgiDegisimi) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	no, err := r.talepNoUret(tx, t.ProjeID, "BILGI")
	if err != nil {
		return err
	}
	t.TalepNo = no
	err = tx.QueryRow(`
		INSERT INTO proje_talep_bilgi_degisimi (proje_id, uye_id, talep_no, degisiklik_tanimi, gerekce)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, olusturma_tarihi`,
		t.ProjeID, t.UyeID, t.TalepNo, t.DegisiklikTanimi, t.Gerekce,
	).Scan(&t.ID, &t.OlusturmaTarihi)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetBilgiDegisimiByProje listeler.
func (r *TalepRepository) GetBilgiDegisimiByProje(projeID int) ([]models.TalepBilgiDegisimi, error) {
	rows, err := r.DB.Query(`
		SELECT t.id, t.proje_id, t.uye_id, t.talep_no, t.degisiklik_tanimi,
		       t.gerekce, t.durum, COALESCE(t.red_notu,''), t.olusturma_tarihi,
		       p.proje_kodu, (SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'),
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad
		FROM proje_talep_bilgi_degisimi t
		JOIN proje p ON p.proje_id = t.proje_id
		JOIN uye u   ON u.uye_id   = t.uye_id
		WHERE t.proje_id = $1 ORDER BY t.olusturma_tarihi DESC`, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.TalepBilgiDegisimi
	for rows.Next() {
		var item models.TalepBilgiDegisimi
		rows.Scan(&item.ID, &item.ProjeID, &item.UyeID, &item.TalepNo, &item.DegisiklikTanimi,
			&item.Gerekce, &item.Durum, &item.RedNotu, &item.OlusturmaTarihi,
			&item.ProjeKodu, &item.ProjeBaslik, &item.TalepEdenAd)
		list = append(list, item)
	}
	return list, nil
}

// UpdateBilgiDegisimiDurum günceller.
func (r *TalepRepository) UpdateBilgiDegisimiDurum(id int, durum, redNotu string) error {
	_, err := r.DB.Exec(`UPDATE proje_talep_bilgi_degisimi SET durum=$1, red_notu=$2, guncelleme_tarihi=NOW() WHERE id=$3`, durum, redNotu, id)
	return err
}

// ---- 8) Proje Dondurma ----

// CreateProjeDondurma, proje dondurma talebi oluşturur.
func (r *TalepRepository) CreateProjeDondurma(t *models.TalepProjeDondurma) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	no, err := r.talepNoUret(tx, t.ProjeID, "DOND")
	if err != nil {
		return err
	}
	t.TalepNo = no
	err = tx.QueryRow(`
		INSERT INTO proje_talep_proje_dondurma (proje_id, uye_id, talep_no, dondurma_sure_ay, gerekce)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, olusturma_tarihi`,
		t.ProjeID, t.UyeID, t.TalepNo, t.DondurmaAy, t.Gerekce,
	).Scan(&t.ID, &t.OlusturmaTarihi)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetProjeDondurmaByProje listeler.
func (r *TalepRepository) GetProjeDondurmaByProje(projeID int) ([]models.TalepProjeDondurma, error) {
	rows, err := r.DB.Query(`
		SELECT t.id, t.proje_id, t.uye_id, t.talep_no, t.dondurma_sure_ay,
		       t.gerekce, t.durum, COALESCE(t.red_notu,''), t.olusturma_tarihi,
		       p.proje_kodu, (SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'),
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad
		FROM proje_talep_proje_dondurma t
		JOIN proje p ON p.proje_id = t.proje_id
		JOIN uye u   ON u.uye_id   = t.uye_id
		WHERE t.proje_id = $1 ORDER BY t.olusturma_tarihi DESC`, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.TalepProjeDondurma
	for rows.Next() {
		var item models.TalepProjeDondurma
		rows.Scan(&item.ID, &item.ProjeID, &item.UyeID, &item.TalepNo, &item.DondurmaAy,
			&item.Gerekce, &item.Durum, &item.RedNotu, &item.OlusturmaTarihi,
			&item.ProjeKodu, &item.ProjeBaslik, &item.TalepEdenAd)
		list = append(list, item)
	}
	return list, nil
}

// UpdateProjeDondurmaD günceller.
func (r *TalepRepository) UpdateProjeDondurmaD(id int, durum, redNotu string) error {
	_, err := r.DB.Exec(`UPDATE proje_talep_proje_dondurma SET durum=$1, red_notu=$2, guncelleme_tarihi=NOW() WHERE id=$3`, durum, redNotu, id)
	return err
}

// ---- 9) Malzeme Güncelleme ----

// CreateMalzemeGuncelleme, malzeme güncelleme talebi oluşturur.
func (r *TalepRepository) CreateMalzemeGuncelleme(t *models.TalepMalzemeGuncelleme) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	no, err := r.talepNoUret(tx, t.ProjeID, "MALZ")
	if err != nil {
		return err
	}
	t.TalepNo = no
	err = tx.QueryRow(`
		INSERT INTO proje_talep_malzeme_guncelleme (proje_id, uye_id, talep_no, guncelleme_tanimi, gerekce)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, olusturma_tarihi`,
		t.ProjeID, t.UyeID, t.TalepNo, t.GuncellemeTanimi, t.Gerekce,
	).Scan(&t.ID, &t.OlusturmaTarihi)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetMalzemeGuncellemeByProje listeler.
func (r *TalepRepository) GetMalzemeGuncellemeByProje(projeID int) ([]models.TalepMalzemeGuncelleme, error) {
	rows, err := r.DB.Query(`
		SELECT t.id, t.proje_id, t.uye_id, t.talep_no, t.guncelleme_tanimi,
		       t.gerekce, t.durum, COALESCE(t.red_notu,''), t.olusturma_tarihi,
		       p.proje_kodu, (SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'),
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad
		FROM proje_talep_malzeme_guncelleme t
		JOIN proje p ON p.proje_id = t.proje_id
		JOIN uye u   ON u.uye_id   = t.uye_id
		WHERE t.proje_id = $1 ORDER BY t.olusturma_tarihi DESC`, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.TalepMalzemeGuncelleme
	for rows.Next() {
		var item models.TalepMalzemeGuncelleme
		rows.Scan(&item.ID, &item.ProjeID, &item.UyeID, &item.TalepNo, &item.GuncellemeTanimi,
			&item.Gerekce, &item.Durum, &item.RedNotu, &item.OlusturmaTarihi,
			&item.ProjeKodu, &item.ProjeBaslik, &item.TalepEdenAd)
		list = append(list, item)
	}
	return list, nil
}

// UpdateMalzemeGuncellemeDurum günceller.
func (r *TalepRepository) UpdateMalzemeGuncellemeDurum(id int, durum, redNotu string) error {
	_, err := r.DB.Exec(`UPDATE proje_talep_malzeme_guncelleme SET durum=$1, red_notu=$2, guncelleme_tarihi=NOW() WHERE id=$3`, durum, redNotu, id)
	return err
}

// ---- 10) Avans ----

// CreateAvans, avans talebi oluşturur.
func (r *TalepRepository) CreateAvans(t *models.TalepAvans) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	no, err := r.talepNoUret(tx, t.ProjeID, "AVANS")
	if err != nil {
		return err
	}
	t.TalepNo = no
	err = tx.QueryRow(`
		INSERT INTO proje_talep_avans (proje_id, uye_id, talep_no, butce_kalemi, tutar_tl, gerekce)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, olusturma_tarihi`,
		t.ProjeID, t.UyeID, t.TalepNo, t.ButceKalemi, t.TutarTL, t.Gerekce,
	).Scan(&t.ID, &t.OlusturmaTarihi)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetAvansByProje listeler.
func (r *TalepRepository) GetAvansByProje(projeID int) ([]models.TalepAvans, error) {
	rows, err := r.DB.Query(`
		SELECT t.id, t.proje_id, t.uye_id, t.talep_no, t.butce_kalemi, t.tutar_tl,
		       t.gerekce, t.durum, COALESCE(t.red_notu,''), t.olusturma_tarihi,
		       p.proje_kodu, (SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'),
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad
		FROM proje_talep_avans t
		JOIN proje p ON p.proje_id = t.proje_id
		JOIN uye u   ON u.uye_id   = t.uye_id
		WHERE t.proje_id = $1 ORDER BY t.olusturma_tarihi DESC`, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.TalepAvans
	for rows.Next() {
		var item models.TalepAvans
		rows.Scan(&item.ID, &item.ProjeID, &item.UyeID, &item.TalepNo, &item.ButceKalemi,
			&item.TutarTL, &item.Gerekce, &item.Durum, &item.RedNotu, &item.OlusturmaTarihi,
			&item.ProjeKodu, &item.ProjeBaslik, &item.TalepEdenAd)
		list = append(list, item)
	}
	return list, nil
}

// UpdateAvansDurum günceller.
func (r *TalepRepository) UpdateAvansDurum(id int, durum, redNotu string) error {
	_, err := r.DB.Exec(`UPDATE proje_talep_avans SET durum=$1, red_notu=$2, guncelleme_tarihi=NOW() WHERE id=$3`, durum, redNotu, id)
	return err
}

// ---- Genel Listeleme (Admin/TTO) ----

// GetAllTalepler, tüm talep türlerini birleşik olarak getirir (Admin/TTO için).
// Türkçe Yorum: Her tablodaki beklemede/onaylandi/reddedildi durumlu kayıtları tip bilgisiyle birleştirir.
func (r *TalepRepository) GetAllTalepler(sadeceBekleyen bool) ([]models.TalepOzet, error) {
	durumFiltreKosulu := ""
	if sadeceBekleyen {
		durumFiltreKosulu = "AND t.durum = 'beklemede'"
	}

	// Türkçe Yorum: UNION ALL ile 10 tabloyu birleştirip tek liste döner. COALESCE ile NULL hataları engellenir.
	query := fmt.Sprintf(`
		SELECT t.id, t.talep_no, 'ek_sure' AS tip, 'Ek Süre' AS tip_etiket,
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_ek_sure t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'ek_butce', 'Ek Bütçe',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_ek_butce t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'fasil_aktarimi', 'Fasıl Aktarımı',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_fasil_aktarimi t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'arastirmaci', 'Araştırmacı Değişikliği',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_arastirmaci t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'bursiyer', 'Bursiyer İşlemi',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_bursiyer t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'proje_iptali', 'Proje İptali',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_proje_iptali t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'bilgi_degisimi', 'Bilgi Değişimi',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_bilgi_degisimi t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'proje_dondurma', 'Proje Dondurma',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_proje_dondurma t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'malzeme_guncelleme', 'Malzeme Güncelleme',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_malzeme_guncelleme t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'avans', 'Avans',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_avans t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		ORDER BY olusturma_tarihi DESC`,
		durumFiltreKosulu, durumFiltreKosulu, durumFiltreKosulu, durumFiltreKosulu,
		durumFiltreKosulu, durumFiltreKosulu, durumFiltreKosulu, durumFiltreKosulu,
		durumFiltreKosulu, durumFiltreKosulu,
	)

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.TalepOzet
	for rows.Next() {
		var item models.TalepOzet
		if err := rows.Scan(&item.ID, &item.TalepNo, &item.TalepTipi, &item.TalepTipiEtiketi,
			&item.ProjeID, &item.ProjeKodu, &item.ProjeBaslik, &item.UyeID,
			&item.TalepEdenAd, &item.Durum, &item.Gerekce, &item.OlusturmaTarihi); err != nil {
			continue
		}
		list = append(list, item)
	}
	return list, nil
}

// GetTaleplerByUye, sadece belirli bir üyenin (akademisyenin) tüm talep türlerini birleştirerek getirir.
// Türkçe Yorum: Akademisyenin kendi taleplerini görüntülemesi için kullanılır.
func (r *TalepRepository) GetTaleplerByUye(uyeID int, sadeceBekleyen bool) ([]models.TalepOzet, error) {
	extraCond := fmt.Sprintf("AND t.uye_id = %d", uyeID)
	if sadeceBekleyen {
		extraCond += " AND t.durum = 'beklemede'"
	}

	query := fmt.Sprintf(`
		SELECT t.id, t.talep_no, 'ek_sure' AS tip, 'Ek Süre' AS tip_etiket,
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_ek_sure t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'ek_butce', 'Ek Bütçe',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_ek_butce t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'fasil_aktarimi', 'Fasıl Aktarımı',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_fasil_aktarimi t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'arastirmaci', 'Araştırmacı Değişikliği',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_arastirmaci t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'bursiyer', 'Bursiyer İşlemi',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_bursiyer t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'proje_iptali', 'Proje İptali',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_proje_iptali t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'bilgi_degisimi', 'Bilgi Değişimi',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_bilgi_degisimi t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'proje_dondurma', 'Proje Dondurma',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_proje_dondurma t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'malzeme_guncelleme', 'Malzeme Güncelleme',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_malzeme_guncelleme t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'avans', 'Avans',
		       t.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), ''), t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi
		FROM proje_talep_avans t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		ORDER BY olusturma_tarihi DESC`,
		extraCond, extraCond, extraCond, extraCond,
		extraCond, extraCond, extraCond, extraCond,
		extraCond, extraCond,
	)

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.TalepOzet
	for rows.Next() {
		var item models.TalepOzet
		if err := rows.Scan(&item.ID, &item.TalepNo, &item.TalepTipi, &item.TalepTipiEtiketi,
			&item.ProjeID, &item.ProjeKodu, &item.ProjeBaslik, &item.UyeID,
			&item.TalepEdenAd, &item.Durum, &item.Gerekce, &item.OlusturmaTarihi); err != nil {
			continue
		}
		list = append(list, item)
	}
	return list, nil
}
