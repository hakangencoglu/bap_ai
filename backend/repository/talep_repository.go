package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
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

// UpdateFasilAktarimiDurum, fasıl aktarımı talebinin durumunu günceller (red vb.).
func (r *TalepRepository) UpdateFasilAktarimiDurum(id int, durum, redNotu string) error {
	_, err := r.DB.Exec(`UPDATE proje_talep_fasil_aktarimi SET durum=$1, red_notu=$2, guncelleme_tarihi=NOW() WHERE id=$3`, durum, redNotu, id)
	return err
}

// ApproveFasilAktarimi, fasıl aktarımını onaylar ve bütçe havuzlarını günceller.
func (r *TalepRepository) ApproveFasilAktarimi(id int, durum, redNotu string) error {
	if durum != models.TalepOnaylandi {
		return r.UpdateFasilAktarimiDurum(id, durum, redNotu)
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	talep, err := r.getFasilAktarimiByIDTx(tx, id)
	if err != nil {
		return err
	}
	if talep.Durum != models.TalepBeklemede {
		return fmt.Errorf("fasıl aktarım talebi zaten işlenmiş")
	}

	if err := r.applyFasilAktarimiBudgetTx(tx, talep); err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE proje_talep_fasil_aktarimi
		SET durum=$1, red_notu=$2, guncelleme_tarihi=NOW()
		WHERE id=$3 AND durum=$4
	`, durum, redNotu, id, models.TalepBeklemede)
	if err != nil {
		return fmt.Errorf("fasıl aktarım durumu güncellenemedi: %w", err)
	}

	return tx.Commit()
}

// ReconcileApprovedFasilAktarimlari, onaylanmış fakat bütçeye yansımamış fasıl taleplerini uygular.
func (r *TalepRepository) ReconcileApprovedFasilAktarimlari() (int, error) {
	rows, err := r.DB.Query(`
		SELECT t.id
		FROM proje_talep_fasil_aktarimi t
		WHERE t.durum = $1
		  AND NOT EXISTS (
			SELECT 1 FROM proje_butce_hareket h
			WHERE h.kaynak_tip = 'fasil_aktarimi' AND h.kaynak_id = t.id
		  )
		ORDER BY t.olusturma_tarihi ASC
	`, models.TalepOnaylandi)
	if err != nil {
		return 0, fmt.Errorf("fasıl aktarım mutabakat listesi alınamadı: %w", err)
	}
	defer rows.Close()

	uygulanan := 0
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return uygulanan, err
		}

		tx, err := r.DB.Begin()
		if err != nil {
			return uygulanan, err
		}

		talep, err := r.getFasilAktarimiByIDTx(tx, id)
		if err != nil {
			tx.Rollback()
			return uygulanan, err
		}
		if err := r.applyFasilAktarimiBudgetTx(tx, talep); err != nil {
			tx.Rollback()
			return uygulanan, fmt.Errorf("fasıl aktarım #%d uygulanamadı: %w", id, err)
		}
		if err := tx.Commit(); err != nil {
			return uygulanan, err
		}
		uygulanan++
	}
	return uygulanan, rows.Err()
}

// getFasilAktarimiByIDTx, transaction içinde tek fasıl aktarım talebini getirir.
func (r *TalepRepository) getFasilAktarimiByIDTx(tx *sql.Tx, id int) (*models.TalepFasilAktarimi, error) {
	var t models.TalepFasilAktarimi
	err := tx.QueryRow(`
		SELECT id, proje_id, uye_id, talep_no, kaynak_kalem, hedef_kalem, tutar_tl,
		       gerekce, durum, COALESCE(red_notu, ''), olusturma_tarihi
		FROM proje_talep_fasil_aktarimi
		WHERE id = $1
	`, id).Scan(
		&t.ID, &t.ProjeID, &t.UyeID, &t.TalepNo, &t.KaynakKalem, &t.HedefKalem, &t.TutarTL,
		&t.Gerekce, &t.Durum, &t.RedNotu, &t.OlusturmaTarihi,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("fasıl aktarım talebi bulunamadı")
	}
	if err != nil {
		return nil, fmt.Errorf("fasıl aktarım talebi okunamadı: %w", err)
	}
	return &t, nil
}

// applyFasilAktarimiBudgetTx, kaynak/hedef kategori havuzları arasında tutarı taşır.
func (r *TalepRepository) applyFasilAktarimiBudgetTx(tx *sql.Tx, talep *models.TalepFasilAktarimi) error {
	var uygulandi bool
	if err := tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM proje_butce_hareket
			WHERE kaynak_tip = 'fasil_aktarimi' AND kaynak_id = $1
		)
	`, talep.ID).Scan(&uygulandi); err != nil {
		return fmt.Errorf("fasıl aktarım ledger kontrolü yapılamadı: %w", err)
	}
	if uygulandi {
		return nil
	}

	if talep.KaynakKalem == talep.HedefKalem {
		return fmt.Errorf("kaynak ve hedef bütçe kalemi aynı olamaz")
	}
	if talep.TutarTL <= 0 {
		return fmt.Errorf("aktarım tutarı geçersiz")
	}

	kaynakKategoriID, err := r.resolveKategoriIDTx(tx, talep.ProjeID, talep.KaynakKalem)
	if err != nil {
		return fmt.Errorf("kaynak kalem çözümlenemedi (%s): %w", talep.KaynakKalem, err)
	}
	hedefKategoriID, err := r.resolveKategoriIDTx(tx, talep.ProjeID, talep.HedefKalem)
	if err != nil {
		return fmt.Errorf("hedef kalem çözümlenemedi (%s): %w", talep.HedefKalem, err)
	}
	if kaynakKategoriID == hedefKategoriID {
		return fmt.Errorf("kaynak ve hedef kategori aynı")
	}

	kullanilabilir, err := r.getKategoriKullanilabilirTx(tx, talep.ProjeID, kaynakKategoriID)
	if err != nil {
		return err
	}
	const eps = 0.009
	if kullanilabilir+eps < talep.TutarTL {
		return fmt.Errorf("kaynak kalemde yeterli kullanılabilir bütçe yok (kalan: %.2f TL, talep: %.2f TL)", kullanilabilir, talep.TutarTL)
	}

	kaynakKalemID, err := r.deductCategoryBudgetTx(tx, talep.ProjeID, kaynakKategoriID, talep.TutarTL)
	if err != nil {
		return err
	}

	hedefAciklama := fmt.Sprintf("Fasıl aktarımı: %s → %s (%s)", talep.KaynakKalem, talep.HedefKalem, talep.TalepNo)
	hedefKalemID, err := r.addCategoryBudgetTx(tx, talep.ProjeID, hedefKategoriID, talep.TutarTL, hedefAciklama)
	if err != nil {
		return err
	}

	kaynakAciklama := fmt.Sprintf("Fasıl aktarımı çıkışı: %s → %s (%s)", talep.KaynakKalem, talep.HedefKalem, talep.TalepNo)
	if err := r.insertButceHareketTx(tx, &models.ButceHareket{
		ProjeID: talep.ProjeID, KalemID: kaynakKalemID,
		KaynakTip: "fasil_aktarimi", KaynakID: talep.ID,
		HareketTip: "manuel_duzeltme", Tutar: -talep.TutarTL,
		Aciklama: &kaynakAciklama,
	}); err != nil {
		return err
	}

	hedefLedgerAciklama := fmt.Sprintf("Fasıl aktarımı girişi: %s → %s (%s)", talep.KaynakKalem, talep.HedefKalem, talep.TalepNo)
	return r.insertButceHareketTx(tx, &models.ButceHareket{
		ProjeID: talep.ProjeID, KalemID: hedefKalemID,
		KaynakTip: "fasil_aktarimi", KaynakID: talep.ID,
		HareketTip: "artirim", Tutar: talep.TutarTL,
		Aciklama: &hedefLedgerAciklama,
	})
}

// resolveKategoriIDTx, kategori adını proje_butce_kategori kimliğine çevirir.
func (r *TalepRepository) resolveKategoriIDTx(tx *sql.Tx, projeID int, kategoriAdi string) (int, error) {
	var kategoriID int

	err := tx.QueryRow(`
		SELECT kategori_id FROM proje_butce_kategori WHERE kategori_adi = $1
	`, kategoriAdi).Scan(&kategoriID)
	if err == nil {
		return kategoriID, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	err = tx.QueryRow(`
		SELECT DISTINCT b.kategori_id
		FROM proje_butce b
		JOIN proje_butce_kategori bk ON bk.kategori_id = b.kategori_id
		WHERE b.proje_id = $1 AND bk.kategori_adi = $2
		LIMIT 1
	`, projeID, kategoriAdi).Scan(&kategoriID)
	if err == nil {
		return kategoriID, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	err = tx.QueryRow(`
		SELECT kategori_id
		FROM proje_butce_kategori
		WHERE kategori_adi ILIKE '%' || $1 || '%' OR $1 ILIKE '%' || kategori_adi || '%'
		ORDER BY length(kategori_adi) ASC
		LIMIT 1
	`, kategoriAdi).Scan(&kategoriID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("kategori bulunamadı: %s", kategoriAdi)
	}
	if err != nil {
		return 0, err
	}
	return kategoriID, nil
}

// getKategoriKullanilabilirTx, kategori havuzundaki kullanılabilir tutarı hesaplar.
func (r *TalepRepository) getKategoriKullanilabilirTx(tx *sql.Tx, projeID, kategoriID int) (float64, error) {
	var planlanan, taahhut, fiili, bekleyen float64

	err := tx.QueryRow(`
		SELECT COALESCE(SUM(toplam_fiyat), 0)
		FROM proje_butce
		WHERE proje_id = $1 AND kategori_id = $2
	`, projeID, kategoriID).Scan(&planlanan)
	if err != nil {
		return 0, fmt.Errorf("planlanan bütçe hesaplanamadı: %w", err)
	}

	err = tx.QueryRow(`
		SELECT COALESCE(SUM(COALESCE(st.revize_toplam_fiyat, st.toplam_fiyat)), 0)
		FROM proje_satinalma_talebi st
		JOIN proje_butce pb ON pb.kalem_id = st.kalem_id
		WHERE st.proje_id = $1
		  AND pb.kategori_id = $2
		  AND st.durum = 'Onaylandı'
		  AND NOT EXISTS (
			SELECT 1 FROM proje_satinalma_odeme o
			WHERE o.talep_id = st.talep_id AND o.durum = 'onaylandi'
		  )
	`, projeID, kategoriID).Scan(&taahhut)
	if err != nil {
		return 0, fmt.Errorf("taahhüt tutarı hesaplanamadı: %w", err)
	}

	err = tx.QueryRow(`
		SELECT COALESCE(SUM(o.fiili_tutar), 0)
		FROM proje_satinalma_odeme o
		JOIN proje_butce pb ON pb.kalem_id = o.kalem_id
		WHERE o.proje_id = $1
		  AND pb.kategori_id = $2
		  AND o.durum = 'onaylandi'
	`, projeID, kategoriID).Scan(&fiili)
	if err != nil {
		return 0, fmt.Errorf("fiili tutar hesaplanamadı: %w", err)
	}

	err = tx.QueryRow(`
		SELECT COALESCE(SUM(COALESCE(st.revize_toplam_fiyat, st.toplam_fiyat)), 0)
		FROM proje_satinalma_talebi st
		JOIN proje_butce pb ON pb.kalem_id = st.kalem_id
		WHERE st.proje_id = $1
		  AND pb.kategori_id = $2
		  AND st.durum = 'Beklemede'
	`, projeID, kategoriID).Scan(&bekleyen)
	if err != nil {
		return 0, fmt.Errorf("bekleyen tutar hesaplanamadı: %w", err)
	}

	return planlanan - taahhut - fiili - bekleyen, nil
}

// deductCategoryBudgetTx, kategori havuzundan tutarı düşer ve işlem yapılan kalem kimliğini döner.
func (r *TalepRepository) deductCategoryBudgetTx(tx *sql.Tx, projeID, kategoriID int, tutar float64) (int, error) {
	rows, err := tx.Query(`
		SELECT kalem_id, COALESCE(toplam_fiyat, 0)
		FROM proje_butce
		WHERE proje_id = $1 AND kategori_id = $2 AND COALESCE(toplam_fiyat, 0) > 0
		ORDER BY kalem_id ASC
	`, projeID, kategoriID)
	if err != nil {
		return 0, fmt.Errorf("kaynak bütçe satırları okunamadı: %w", err)
	}
	defer rows.Close()

	kalan := tutar
	var ilkKalemID int
	for rows.Next() {
		var kalemID int
		var satirTutar float64
		if err := rows.Scan(&kalemID, &satirTutar); err != nil {
			return 0, err
		}
		if ilkKalemID == 0 {
			ilkKalemID = kalemID
		}
		if kalan <= 0 {
			break
		}

		dusulecek := math.Min(satirTutar, kalan)
		if dusulecek <= 0 {
			continue
		}

		_, err = tx.Exec(`
			UPDATE proje_butce
			SET toplam_fiyat = toplam_fiyat - $1, guncelleme_tarihi = NOW()
			WHERE kalem_id = $2 AND proje_id = $3
		`, dusulecek, kalemID, projeID)
		if err != nil {
			return 0, fmt.Errorf("kaynak bütçe düşülemedi: %w", err)
		}
		kalan -= dusulecek
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if kalan > 0.009 {
		return 0, fmt.Errorf("kaynak kategoride yeterli planlanan bütçe yok")
	}
	if ilkKalemID == 0 {
		return 0, fmt.Errorf("kaynak kategori için bütçe satırı bulunamadı")
	}
	return ilkKalemID, nil
}

// addCategoryBudgetTx, hedef kategori havuzuna tutar ekler ve işlem yapılan kalem kimliğini döner.
func (r *TalepRepository) addCategoryBudgetTx(tx *sql.Tx, projeID, kategoriID int, tutar float64, aciklama string) (int, error) {
	var hedefKalemID int
	err := tx.QueryRow(`
		SELECT kalem_id
		FROM proje_butce
		WHERE proje_id = $1 AND kategori_id = $2
		ORDER BY kalem_id ASC
		LIMIT 1
	`, projeID, kategoriID).Scan(&hedefKalemID)
	if err == nil {
		_, err = tx.Exec(`
			UPDATE proje_butce
			SET toplam_fiyat = COALESCE(toplam_fiyat, 0) + $1,
			    birim_fiyat = COALESCE(birim_fiyat, 0) + $1,
			    guncelleme_tarihi = NOW()
			WHERE kalem_id = $2
		`, tutar, hedefKalemID)
		if err != nil {
			return 0, fmt.Errorf("hedef bütçe artırılamadı: %w", err)
		}
		return hedefKalemID, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("hedef bütçe satırı aranırken hata: %w", err)
	}

	err = tx.QueryRow(`
		INSERT INTO proje_butce (proje_id, kategori_id, aciklama, birim_ozelligi, birim_fiyat, toplam_fiyat)
		VALUES ($1, $2, $3, 1, $4, $4)
		RETURNING kalem_id
	`, projeID, kategoriID, aciklama, tutar).Scan(&hedefKalemID)
	if err != nil {
		return 0, fmt.Errorf("hedef bütçe satırı oluşturulamadı: %w", err)
	}
	return hedefKalemID, nil
}

// insertButceHareketTx, fasıl aktarımı için bütçe ledger satırı yazar.
func (r *TalepRepository) insertButceHareketTx(tx *sql.Tx, h *models.ButceHareket) error {
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

const talepBaslikAltSorgu = `COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr' LIMIT 1), '')`

// buildTalepUnionQuery tüm talep türlerini tip bazlı detay JSON alanıyla birleştirir.
// Türkçe Yorum: Listeleme ekranlarında fasıl aktarımı vb. tür özel alanların modalda gösterilmesi için detay üretir.
func buildTalepUnionQuery(extraWhere string) string {
	return fmt.Sprintf(`
		SELECT t.id, t.talep_no, 'ek_sure' AS tip, 'Ek Süre' AS tip_etiket,
		       t.proje_id, COALESCE(p.proje_kodu, ''), %s, t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi,
		       json_build_object('ek_sure_ay', t.ek_sure_ay) AS detay
		FROM proje_talep_ek_sure t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'ek_butce', 'Ek Bütçe',
		       t.proje_id, COALESCE(p.proje_kodu, ''), %s, t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi,
		       json_build_object('butce_kalemi', t.butce_kalemi, 'tutar_tl', t.tutar_tl) AS detay
		FROM proje_talep_ek_butce t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'fasil_aktarimi', 'Fasıl Aktarımı',
		       t.proje_id, COALESCE(p.proje_kodu, ''), %s, t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi,
		       json_build_object('kaynak_kalem', t.kaynak_kalem, 'hedef_kalem', t.hedef_kalem, 'tutar_tl', t.tutar_tl) AS detay
		FROM proje_talep_fasil_aktarimi t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'arastirmaci', 'Araştırmacı Değişikliği',
		       t.proje_id, COALESCE(p.proje_kodu, ''), %s, t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi,
		       json_build_object('islem_turu', t.islem_turu, 'arastirmaci_adi', t.arastirmaci_adi) AS detay
		FROM proje_talep_arastirmaci t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'bursiyer', 'Bursiyer İşlemi',
		       t.proje_id, COALESCE(p.proje_kodu, ''), %s, t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi,
		       json_build_object('bursiyer_kimlik', t.bursiyer_kimlik, 'bursiyer_adi', t.bursiyer_adi, 'islem_turu', t.islem_turu) AS detay
		FROM proje_talep_bursiyer t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'proje_iptali', 'Proje İptali',
		       t.proje_id, COALESCE(p.proje_kodu, ''), %s, t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi,
		       '{}'::json AS detay
		FROM proje_talep_proje_iptali t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'bilgi_degisimi', 'Bilgi Değişimi',
		       t.proje_id, COALESCE(p.proje_kodu, ''), %s, t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi,
		       json_build_object('degisiklik_tanimi', t.degisiklik_tanimi) AS detay
		FROM proje_talep_bilgi_degisimi t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'proje_dondurma', 'Proje Dondurma',
		       t.proje_id, COALESCE(p.proje_kodu, ''), %s, t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi,
		       json_build_object('dondurma_sure_ay', t.dondurma_sure_ay) AS detay
		FROM proje_talep_proje_dondurma t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'malzeme_guncelleme', 'Malzeme Güncelleme',
		       t.proje_id, COALESCE(p.proje_kodu, ''), %s, t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi,
		       json_build_object('guncelleme_tanimi', t.guncelleme_tanimi) AS detay
		FROM proje_talep_malzeme_guncelleme t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		UNION ALL
		SELECT t.id, t.talep_no, 'avans', 'Avans',
		       t.proje_id, COALESCE(p.proje_kodu, ''), %s, t.uye_id,
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad,
		       t.durum, COALESCE(t.gerekce, ''), t.olusturma_tarihi,
		       json_build_object('butce_kalemi', t.butce_kalemi, 'tutar_tl', t.tutar_tl) AS detay
		FROM proje_talep_avans t JOIN proje p ON p.proje_id=t.proje_id JOIN uye u ON u.uye_id=t.uye_id
		WHERE 1=1 %s
		ORDER BY olusturma_tarihi DESC`,
		talepBaslikAltSorgu, extraWhere,
		talepBaslikAltSorgu, extraWhere,
		talepBaslikAltSorgu, extraWhere,
		talepBaslikAltSorgu, extraWhere,
		talepBaslikAltSorgu, extraWhere,
		talepBaslikAltSorgu, extraWhere,
		talepBaslikAltSorgu, extraWhere,
		talepBaslikAltSorgu, extraWhere,
		talepBaslikAltSorgu, extraWhere,
		talepBaslikAltSorgu, extraWhere,
	)
}

// parseTalepDetayJSON birleşik sorgudan gelen detay JSON'unu map'e çevirir.
func parseTalepDetayJSON(raw []byte) map[string]interface{} {
	if len(raw) == 0 {
		return nil
	}
	var detay map[string]interface{}
	if err := json.Unmarshal(raw, &detay); err != nil || len(detay) == 0 {
		return nil
	}
	return detay
}

// scanTalepOzetList sorgu sonucunu TalepOzet dizisine dönüştürür.
func scanTalepOzetList(rows *sql.Rows) ([]models.TalepOzet, error) {
	var list []models.TalepOzet
	for rows.Next() {
		var item models.TalepOzet
		var detayRaw []byte
		if err := rows.Scan(&item.ID, &item.TalepNo, &item.TalepTipi, &item.TalepTipiEtiketi,
			&item.ProjeID, &item.ProjeKodu, &item.ProjeBaslik, &item.UyeID,
			&item.TalepEdenAd, &item.Durum, &item.Gerekce, &item.OlusturmaTarihi, &detayRaw); err != nil {
			continue
		}
		item.Detay = parseTalepDetayJSON(detayRaw)
		list = append(list, item)
	}
	return list, nil
}

// GetAllTalepler, tüm talep türlerini birleşik olarak getirir (Admin/TTO için).
// Türkçe Yorum: Her tablodaki beklemede/onaylandi/reddedildi durumlu kayıtları tip bilgisiyle birleştirir.
func (r *TalepRepository) GetAllTalepler(sadeceBekleyen bool) ([]models.TalepOzet, error) {
	extraWhere := ""
	if sadeceBekleyen {
		extraWhere = "AND t.durum = 'beklemede'"
	}

	rows, err := r.DB.Query(buildTalepUnionQuery(extraWhere))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTalepOzetList(rows)
}

// GetTaleplerByUye, sadece belirli bir üyenin (akademisyenin) tüm talep türlerini birleştirerek getirir.
// Türkçe Yorum: Akademisyenin kendi taleplerini görüntülemesi için kullanılır.
func (r *TalepRepository) GetTaleplerByUye(uyeID int, sadeceBekleyen bool) ([]models.TalepOzet, error) {
	extraCond := fmt.Sprintf("AND t.uye_id = %d", uyeID)
	if sadeceBekleyen {
		extraCond += " AND t.durum = 'beklemede'"
	}

	rows, err := r.DB.Query(buildTalepUnionQuery(extraCond))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTalepOzetList(rows)
}
