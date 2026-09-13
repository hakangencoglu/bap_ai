package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"bap_ai/backend/models"
)

// DegisiklikRepository, proje önce/sonra audit kayıtlarını yönetir.
// Türkçe Yorum: Talep ve diğer olaylarda varlık snapshot'ı alıp proje_degisiklik tablolarına yazar.
type DegisiklikRepository struct {
	DB *sql.DB
}

// NewDegisiklikRepository yeni repository oluşturur.
func NewDegisiklikRepository(db *sql.DB) *DegisiklikRepository {
	return &DegisiklikRepository{DB: db}
}

// talepTabloMap, talep tipi → tablo adı eşlemesi.
var talepTabloMap = map[string]string{
	"ek_sure":            "proje_talep_ek_sure",
	"ek_butce":           "proje_talep_ek_butce",
	"fasil_aktarimi":     "proje_talep_fasil_aktarimi",
	"arastirmaci":        "proje_talep_arastirmaci",
	"bursiyer":           "proje_talep_bursiyer",
	"proje_iptali":       "proje_talep_proje_iptali",
	"bilgi_degisimi":     "proje_talep_bilgi_degisimi",
	"proje_dondurma":     "proje_talep_proje_dondurma",
	"malzeme_guncelleme": "proje_talep_malzeme_guncelleme",
	"avans":              "proje_talep_avans",
}

// RecordChange, bir değişiklik olayını ve detaylarını kaydeder.
// Türkçe Yorum: Versiyon artırılacaksa proje.icerik_versiyon atomik olarak yükseltilir.
func (r *DegisiklikRepository) RecordChange(istek models.DegisiklikKayitIstek) (int, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	versiyon := 1
	if istek.VersiyonArtir {
		err = tx.QueryRow(`
			UPDATE proje SET icerik_versiyon = icerik_versiyon + 1, guncelleme_tarihi = NOW()
			WHERE proje_id = $1
			RETURNING icerik_versiyon
		`, istek.ProjeID).Scan(&versiyon)
		if err != nil {
			return 0, fmt.Errorf("içerik versiyonu artırılamadı: %w", err)
		}
	} else {
		err = tx.QueryRow(`SELECT COALESCE(icerik_versiyon, 1) FROM proje WHERE proje_id = $1`, istek.ProjeID).Scan(&versiyon)
		if err != nil {
			versiyon = 1
		}
	}

	var degisiklikID int
	var islemiYapan interface{}
	if istek.IslemiYapanID > 0 {
		islemiYapan = istek.IslemiYapanID
	}
	err = tx.QueryRow(`
		INSERT INTO proje_degisiklik
			(proje_id, icerik_versiyon, olay_tipi, kaynak_tip, kaynak_id, talep_no, ozet, islemi_yapan_id)
		VALUES ($1, $2, $3, $4, NULLIF($5, 0), NULLIF($6, ''), $7, $8)
		RETURNING degisiklik_id
	`, istek.ProjeID, versiyon, istek.OlayTipi, nullIfEmpty(istek.KaynakTip), istek.KaynakID,
		istek.TalepNo, istek.Ozet, islemiYapan).Scan(&degisiklikID)
	if err != nil {
		return 0, fmt.Errorf("değişiklik başlığı yazılamadı: %w", err)
	}

	for _, d := range istek.Detaylar {
		oncekiRaw, err := toJSONRaw(d.Onceki)
		if err != nil {
			return 0, err
		}
		sonrakiRaw, err := toJSONRaw(d.Sonraki)
		if err != nil {
			return 0, err
		}
		diffRaw, err := buildAlanDiff(d.Onceki, d.Sonraki)
		if err != nil {
			return 0, err
		}
		_, err = tx.Exec(`
			INSERT INTO proje_degisiklik_detay
				(degisiklik_id, varlik_tip, varlik_id, onceki_json, sonraki_json, alan_diff)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, degisiklikID, d.VarlikTip, d.VarlikID, oncekiRaw, sonrakiRaw, diffRaw)
		if err != nil {
			return 0, fmt.Errorf("değişiklik detayı yazılamadı: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return degisiklikID, nil
}

// CaptureProje, projenin özet snapshot'ını döner.
func (r *DegisiklikRepository) CaptureProje(projeID int) (map[string]interface{}, error) {
	row := map[string]interface{}{}
	var sureAy, icerikVersiyon int
	var projeKodu, baslik, durumAdi string
	var durumID sql.NullInt64
	err := r.DB.QueryRow(`
		SELECT p.proje_id, COALESCE(p.proje_kodu, ''), COALESCE(p.baslik_tr, ''),
		       COALESCE(p.sure_ay, 0), p.durum_id, COALESCE(pd.durum_adi, ''),
		       COALESCE(p.icerik_versiyon, 1)
		FROM proje p
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE p.proje_id = $1
	`, projeID).Scan(&projeID, &projeKodu, &baslik, &sureAy, &durumID, &durumAdi, &icerikVersiyon)
	if err != nil {
		return nil, err
	}
	row["proje_id"] = projeID
	row["proje_kodu"] = projeKodu
	row["baslik_tr"] = baslik
	row["sure_ay"] = sureAy
	if durumID.Valid {
		row["durum_id"] = durumID.Int64
	}
	row["durum_adi"] = durumAdi
	row["icerik_versiyon"] = icerikVersiyon

	var bitis sql.NullString
	_ = r.DB.QueryRow(`SELECT bitis_tarihi::text FROM proje_sozlesme WHERE proje_id = $1`, projeID).Scan(&bitis)
	if bitis.Valid {
		row["bitis_tarihi"] = bitis.String
	}
	return row, nil
}

// CaptureButceByKategori, kategori adına göre bütçe özetini döner.
func (r *DegisiklikRepository) CaptureButceByKategori(projeID int, kategoriAdi string) (map[string]interface{}, error) {
	var kategoriID int
	var planlanan float64
	err := r.DB.QueryRow(`
		SELECT COALESCE(bk.kategori_id, 0), COALESCE(SUM(b.toplam_fiyat), 0)
		FROM proje_butce_kategori bk
		LEFT JOIN proje_butce b ON b.kategori_id = bk.kategori_id AND b.proje_id = $1
		WHERE bk.kategori_adi = $2
		GROUP BY bk.kategori_id
	`, projeID, kategoriAdi).Scan(&kategoriID, &planlanan)
	if err == sql.ErrNoRows {
		return map[string]interface{}{
			"proje_id":     projeID,
			"kategori_adi": kategoriAdi,
			"planlanan":    0.0,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"proje_id":     projeID,
		"kategori_id":  kategoriID,
		"kategori_adi": kategoriAdi,
		"planlanan":    planlanan,
	}, nil
}

// CaptureButceOzet, projenin tüm kategori bütçe özetini döner.
func (r *DegisiklikRepository) CaptureButceOzet(projeID int) ([]map[string]interface{}, error) {
	rows, err := r.DB.Query(`
		SELECT COALESCE(bk.kategori_id, 0), COALESCE(bk.kategori_adi, 'Belirtilmemiş'),
		       COALESCE(SUM(b.toplam_fiyat), 0)
		FROM proje_butce b
		LEFT JOIN proje_butce_kategori bk ON b.kategori_id = bk.kategori_id
		WHERE b.proje_id = $1
		GROUP BY bk.kategori_id, bk.kategori_adi
		ORDER BY bk.kategori_adi
	`, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var kid int
		var adi string
		var planlanan float64
		if err := rows.Scan(&kid, &adi, &planlanan); err != nil {
			return nil, err
		}
		list = append(list, map[string]interface{}{
			"kategori_id":  kid,
			"kategori_adi": adi,
			"planlanan":    planlanan,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	return list, nil
}

// CaptureTakim, proje ekip listesini döner.
func (r *DegisiklikRepository) CaptureTakim(projeID int) ([]map[string]interface{}, error) {
	rows, err := r.DB.Query(`
		SELECT u.uye_id, u.ad || ' ' || u.soyad AS ad_tumu,
		       COALESCE(prt.proje_rol, 'Araştırmacı'),
		       COALESCE(pt.davet_durumu, 'beklemede')
		FROM proje_takim pt
		INNER JOIN uye u ON pt.uye_id = u.uye_id
		LEFT JOIN proje_rol_tanimlama prt ON pt.proje_rol_id = prt.rol_id
		WHERE pt.proje_id = $1
		ORDER BY u.ad, u.soyad
	`, projeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []map[string]interface{}
	for rows.Next() {
		var uid int
		var ad, rol, davet string
		if err := rows.Scan(&uid, &ad, &rol, &davet); err != nil {
			return nil, err
		}
		list = append(list, map[string]interface{}{
			"uye_id":       uid,
			"ad_tumu":      ad,
			"proje_rol":    rol,
			"davet_durumu": davet,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	return list, nil
}

// GetTalepMeta, talep tipine göre proje_id, talep_no ve ham satır snapshot'ı döner.
func (r *DegisiklikRepository) GetTalepMeta(tip string, id int) (projeID int, talepNo string, snapshot map[string]interface{}, err error) {
	tablo, ok := talepTabloMap[tip]
	if !ok {
		return 0, "", nil, fmt.Errorf("bilinmeyen talep tipi: %s", tip)
	}
	err = r.DB.QueryRow(
		fmt.Sprintf(`SELECT proje_id, COALESCE(talep_no, '') FROM %s WHERE id = $1`, tablo),
		id,
	).Scan(&projeID, &talepNo)
	if err != nil {
		return 0, "", nil, err
	}
	snapshot, err = r.captureTalepRow(tip, id)
	return projeID, talepNo, snapshot, err
}

// captureTalepRow, talep satırını tipine göre map olarak okur.
func (r *DegisiklikRepository) captureTalepRow(tip string, id int) (map[string]interface{}, error) {
	m := map[string]interface{}{"talep_tipi": tip, "id": id}
	switch tip {
	case "ek_sure":
		var projeID, uyeID, ekSureAy int
		var talepNo, gerekce, durum, redNotu string
		err := r.DB.QueryRow(`
			SELECT proje_id, uye_id, COALESCE(talep_no,''), ek_sure_ay, COALESCE(gerekce,''),
			       COALESCE(durum,''), COALESCE(red_notu,'')
			FROM proje_talep_ek_sure WHERE id=$1`, id).
			Scan(&projeID, &uyeID, &talepNo, &ekSureAy, &gerekce, &durum, &redNotu)
		if err != nil {
			return nil, err
		}
		m["proje_id"], m["uye_id"], m["talep_no"] = projeID, uyeID, talepNo
		m["ek_sure_ay"], m["gerekce"], m["durum"], m["red_notu"] = ekSureAy, gerekce, durum, redNotu
	case "ek_butce":
		var projeID, uyeID int
		var talepNo, kalem, gerekce, durum, redNotu string
		var tutar float64
		err := r.DB.QueryRow(`
			SELECT proje_id, uye_id, COALESCE(talep_no,''), COALESCE(butce_kalemi,''), tutar_tl,
			       COALESCE(gerekce,''), COALESCE(durum,''), COALESCE(red_notu,'')
			FROM proje_talep_ek_butce WHERE id=$1`, id).
			Scan(&projeID, &uyeID, &talepNo, &kalem, &tutar, &gerekce, &durum, &redNotu)
		if err != nil {
			return nil, err
		}
		m["proje_id"], m["uye_id"], m["talep_no"] = projeID, uyeID, talepNo
		m["butce_kalemi"], m["tutar_tl"], m["gerekce"], m["durum"], m["red_notu"] = kalem, tutar, gerekce, durum, redNotu
	case "fasil_aktarimi":
		var projeID, uyeID int
		var talepNo, kaynak, hedef, gerekce, durum, redNotu string
		var tutar float64
		err := r.DB.QueryRow(`
			SELECT proje_id, uye_id, COALESCE(talep_no,''), COALESCE(kaynak_kalem,''), COALESCE(hedef_kalem,''),
			       tutar_tl, COALESCE(gerekce,''), COALESCE(durum,''), COALESCE(red_notu,'')
			FROM proje_talep_fasil_aktarimi WHERE id=$1`, id).
			Scan(&projeID, &uyeID, &talepNo, &kaynak, &hedef, &tutar, &gerekce, &durum, &redNotu)
		if err != nil {
			return nil, err
		}
		m["proje_id"], m["uye_id"], m["talep_no"] = projeID, uyeID, talepNo
		m["kaynak_kalem"], m["hedef_kalem"], m["tutar_tl"] = kaynak, hedef, tutar
		m["gerekce"], m["durum"], m["red_notu"] = gerekce, durum, redNotu
	case "arastirmaci":
		var projeID, uyeID int
		var talepNo, islem, ad, gerekce, durum, redNotu string
		err := r.DB.QueryRow(`
			SELECT proje_id, uye_id, COALESCE(talep_no,''), COALESCE(islem_turu,''), COALESCE(arastirmaci_adi,''),
			       COALESCE(gerekce,''), COALESCE(durum,''), COALESCE(red_notu,'')
			FROM proje_talep_arastirmaci WHERE id=$1`, id).
			Scan(&projeID, &uyeID, &talepNo, &islem, &ad, &gerekce, &durum, &redNotu)
		if err != nil {
			return nil, err
		}
		m["proje_id"], m["uye_id"], m["talep_no"] = projeID, uyeID, talepNo
		m["islem_turu"], m["arastirmaci_adi"], m["gerekce"], m["durum"], m["red_notu"] = islem, ad, gerekce, durum, redNotu
	case "bursiyer":
		var projeID, uyeID int
		var talepNo, kimlik, ad, islem, gerekce, durum, redNotu string
		err := r.DB.QueryRow(`
			SELECT proje_id, uye_id, COALESCE(talep_no,''), COALESCE(bursiyer_kimlik,''), COALESCE(bursiyer_adi,''),
			       COALESCE(islem_turu,''), COALESCE(gerekce,''), COALESCE(durum,''), COALESCE(red_notu,'')
			FROM proje_talep_bursiyer WHERE id=$1`, id).
			Scan(&projeID, &uyeID, &talepNo, &kimlik, &ad, &islem, &gerekce, &durum, &redNotu)
		if err != nil {
			return nil, err
		}
		m["proje_id"], m["uye_id"], m["talep_no"] = projeID, uyeID, talepNo
		m["bursiyer_kimlik"] = maskKimlik(kimlik)
		m["bursiyer_adi"], m["islem_turu"] = ad, islem
		m["gerekce"], m["durum"], m["red_notu"] = gerekce, durum, redNotu
	case "proje_iptali":
		var projeID, uyeID int
		var talepNo, gerekce, durum, redNotu string
		err := r.DB.QueryRow(`
			SELECT proje_id, uye_id, COALESCE(talep_no,''), COALESCE(gerekce,''), COALESCE(durum,''), COALESCE(red_notu,'')
			FROM proje_talep_proje_iptali WHERE id=$1`, id).
			Scan(&projeID, &uyeID, &talepNo, &gerekce, &durum, &redNotu)
		if err != nil {
			return nil, err
		}
		m["proje_id"], m["uye_id"], m["talep_no"] = projeID, uyeID, talepNo
		m["gerekce"], m["durum"], m["red_notu"] = gerekce, durum, redNotu
	case "bilgi_degisimi":
		var projeID, uyeID int
		var talepNo, tanim, gerekce, durum, redNotu string
		err := r.DB.QueryRow(`
			SELECT proje_id, uye_id, COALESCE(talep_no,''), COALESCE(degisiklik_tanimi,''),
			       COALESCE(gerekce,''), COALESCE(durum,''), COALESCE(red_notu,'')
			FROM proje_talep_bilgi_degisimi WHERE id=$1`, id).
			Scan(&projeID, &uyeID, &talepNo, &tanim, &gerekce, &durum, &redNotu)
		if err != nil {
			return nil, err
		}
		m["proje_id"], m["uye_id"], m["talep_no"] = projeID, uyeID, talepNo
		m["degisiklik_tanimi"], m["gerekce"], m["durum"], m["red_notu"] = tanim, gerekce, durum, redNotu
	case "proje_dondurma":
		var projeID, uyeID, dondurmaAy int
		var talepNo, gerekce, durum, redNotu string
		err := r.DB.QueryRow(`
			SELECT proje_id, uye_id, COALESCE(talep_no,''), dondurma_sure_ay,
			       COALESCE(gerekce,''), COALESCE(durum,''), COALESCE(red_notu,'')
			FROM proje_talep_proje_dondurma WHERE id=$1`, id).
			Scan(&projeID, &uyeID, &talepNo, &dondurmaAy, &gerekce, &durum, &redNotu)
		if err != nil {
			return nil, err
		}
		m["proje_id"], m["uye_id"], m["talep_no"] = projeID, uyeID, talepNo
		m["dondurma_sure_ay"], m["gerekce"], m["durum"], m["red_notu"] = dondurmaAy, gerekce, durum, redNotu
	case "malzeme_guncelleme":
		var projeID, uyeID int
		var talepNo, tanim, gerekce, durum, redNotu string
		err := r.DB.QueryRow(`
			SELECT proje_id, uye_id, COALESCE(talep_no,''), COALESCE(guncelleme_tanimi,''),
			       COALESCE(gerekce,''), COALESCE(durum,''), COALESCE(red_notu,'')
			FROM proje_talep_malzeme_guncelleme WHERE id=$1`, id).
			Scan(&projeID, &uyeID, &talepNo, &tanim, &gerekce, &durum, &redNotu)
		if err != nil {
			return nil, err
		}
		m["proje_id"], m["uye_id"], m["talep_no"] = projeID, uyeID, talepNo
		m["guncelleme_tanimi"], m["gerekce"], m["durum"], m["red_notu"] = tanim, gerekce, durum, redNotu
	case "avans":
		var projeID, uyeID int
		var talepNo, kalem, gerekce, durum, redNotu string
		var tutar float64
		err := r.DB.QueryRow(`
			SELECT proje_id, uye_id, COALESCE(talep_no,''), COALESCE(butce_kalemi,''), tutar_tl,
			       COALESCE(gerekce,''), COALESCE(durum,''), COALESCE(red_notu,'')
			FROM proje_talep_avans WHERE id=$1`, id).
			Scan(&projeID, &uyeID, &talepNo, &kalem, &tutar, &gerekce, &durum, &redNotu)
		if err != nil {
			return nil, err
		}
		m["proje_id"], m["uye_id"], m["talep_no"] = projeID, uyeID, talepNo
		m["butce_kalemi"], m["tutar_tl"], m["gerekce"], m["durum"], m["red_notu"] = kalem, tutar, gerekce, durum, redNotu
	default:
		return nil, fmt.Errorf("bilinmeyen talep tipi: %s", tip)
	}
	return m, nil
}

// ListByProje, proje değişikliklerini listeler (detay JSON olmadan).
func (r *DegisiklikRepository) ListByProje(projeID int, olayTipi, kaynakTip string) ([]models.ProjeDegisiklik, error) {
	q := `
		SELECT d.degisiklik_id, d.proje_id, d.icerik_versiyon, d.olay_tipi,
		       COALESCE(d.kaynak_tip, ''), COALESCE(d.kaynak_id, 0), COALESCE(d.talep_no, ''),
		       COALESCE(d.ozet, ''), COALESCE(d.islemi_yapan_id, 0),
		       COALESCE(u.ad || ' ' || u.soyad, ''), d.islem_tarihi
		FROM proje_degisiklik d
		LEFT JOIN uye u ON d.islemi_yapan_id = u.uye_id
		WHERE d.proje_id = $1`
	args := []interface{}{projeID}
	idx := 2
	if olayTipi != "" {
		q += fmt.Sprintf(` AND d.olay_tipi = $%d`, idx)
		args = append(args, olayTipi)
		idx++
	}
	if kaynakTip != "" {
		q += fmt.Sprintf(` AND d.kaynak_tip = $%d`, idx)
		args = append(args, kaynakTip)
	}
	q += ` ORDER BY d.islem_tarihi DESC, d.degisiklik_id DESC`

	rows, err := r.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.ProjeDegisiklik
	for rows.Next() {
		var item models.ProjeDegisiklik
		if err := rows.Scan(
			&item.DegisiklikID, &item.ProjeID, &item.IcerikVersiyon, &item.OlayTipi,
			&item.KaynakTip, &item.KaynakID, &item.TalepNo, &item.Ozet,
			&item.IslemiYapanID, &item.IslemiYapanAd, &item.IslemTarihi,
		); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	if list == nil {
		list = []models.ProjeDegisiklik{}
	}
	return list, nil
}

// GetByID, değişiklik başlığı ve detaylarını döner.
func (r *DegisiklikRepository) GetByID(degisiklikID int) (*models.ProjeDegisiklik, error) {
	item := &models.ProjeDegisiklik{}
	err := r.DB.QueryRow(`
		SELECT d.degisiklik_id, d.proje_id, d.icerik_versiyon, d.olay_tipi,
		       COALESCE(d.kaynak_tip, ''), COALESCE(d.kaynak_id, 0), COALESCE(d.talep_no, ''),
		       COALESCE(d.ozet, ''), COALESCE(d.islemi_yapan_id, 0),
		       COALESCE(u.ad || ' ' || u.soyad, ''), d.islem_tarihi
		FROM proje_degisiklik d
		LEFT JOIN uye u ON d.islemi_yapan_id = u.uye_id
		WHERE d.degisiklik_id = $1
	`, degisiklikID).Scan(
		&item.DegisiklikID, &item.ProjeID, &item.IcerikVersiyon, &item.OlayTipi,
		&item.KaynakTip, &item.KaynakID, &item.TalepNo, &item.Ozet,
		&item.IslemiYapanID, &item.IslemiYapanAd, &item.IslemTarihi,
	)
	if err != nil {
		return nil, err
	}

	rows, err := r.DB.Query(`
		SELECT detay_id, degisiklik_id, varlik_tip, varlik_id, onceki_json, sonraki_json, alan_diff
		FROM proje_degisiklik_detay WHERE degisiklik_id = $1 ORDER BY detay_id
	`, degisiklikID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d models.ProjeDegisiklikDetay
		var onceki, sonraki, diff []byte
		if err := rows.Scan(&d.DetayID, &d.DegisiklikID, &d.VarlikTip, &d.VarlikID, &onceki, &sonraki, &diff); err != nil {
			return nil, err
		}
		d.OncekiJSON = onceki
		d.SonrakiJSON = sonraki
		d.AlanDiff = diff
		item.Detaylar = append(item.Detaylar, d)
	}
	if item.Detaylar == nil {
		item.Detaylar = []models.ProjeDegisiklikDetay{}
	}
	return item, nil
}

// nullIfEmpty, boş string için nil döner.
func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// toJSONRaw, değeri JSONB için hazırlar.
func toJSONRaw(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// buildAlanDiff, düz map alanları için basit önce/sonra farkı üretir.
func buildAlanDiff(onceki, sonraki interface{}) (interface{}, error) {
	om, ok1 := asMap(onceki)
	sm, ok2 := asMap(sonraki)
	if !ok1 && !ok2 {
		return nil, nil
	}
	diff := map[string]interface{}{}
	keys := map[string]bool{}
	for k := range om {
		keys[k] = true
	}
	for k := range sm {
		keys[k] = true
	}
	for k := range keys {
		ov, oHas := om[k]
		sv, sHas := sm[k]
		oStr := fmt.Sprintf("%v", ov)
		sStr := fmt.Sprintf("%v", sv)
		if !oHas {
			diff[k] = map[string]interface{}{"onceki": nil, "sonraki": sv}
		} else if !sHas {
			diff[k] = map[string]interface{}{"onceki": ov, "sonraki": nil}
		} else if oStr != sStr {
			diff[k] = map[string]interface{}{"onceki": ov, "sonraki": sv}
		}
	}
	if len(diff) == 0 {
		return nil, nil
	}
	return toJSONRaw(diff)
}

// asMap, interface{} değerini map'e çevirir.
func asMap(v interface{}) (map[string]interface{}, bool) {
	if v == nil {
		return map[string]interface{}{}, false
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m, true
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, false
	}
	return m, true
}

// maskKimlik, TC/kimlik numarasının ortasını maskeler.
func maskKimlik(kimlik string) string {
	kimlik = strings.TrimSpace(kimlik)
	if len(kimlik) < 5 {
		return "***"
	}
	return kimlik[:2] + strings.Repeat("*", len(kimlik)-4) + kimlik[len(kimlik)-2:]
}
