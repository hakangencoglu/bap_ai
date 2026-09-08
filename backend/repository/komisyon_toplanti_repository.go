package repository

import (
	"database/sql"
	"fmt"

	"bap_ai/backend/models"
)

// KomisyonToplantiRepository toplantı ↔ proje köprü tablo işlemlerini yürütür.
// Türkçe Yorum: Hangi projenin hangi toplantıda görüşüldüğünü ve kararlarını yönetir.
type KomisyonToplantiRepository struct {
	DB *sql.DB
}

// NewKomisyonToplantiRepository yeni bir KomisyonToplantiRepository oluşturur.
func NewKomisyonToplantiRepository(db *sql.DB) *KomisyonToplantiRepository {
	return &KomisyonToplantiRepository{DB: db}
}

// GetToplantiDurum toplantının mevcut durumunu (planli|tamamlandi|iptal) döner.
// Türkçe Yorum: Tamamlanmış toplantıya yeni gündem maddesi eklenmesini engellemek için kullanılır.
func (r *KomisyonToplantiRepository) GetToplantiDurum(toplantiID int) (string, error) {
	var durum string
	err := r.DB.QueryRow(`
		SELECT COALESCE(durum, 'planli') FROM komisyon_toplantisi WHERE toplanti_id = $1
	`, toplantiID).Scan(&durum)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("toplantı bulunamadı")
	}
	if err != nil {
		return "", fmt.Errorf("toplantı durumu sorgulanamadı: %w", err)
	}
	return durum, nil
}

// AddProjeToToplanti bir projeyi veya talebi belirtilen toplantıya gündem maddesi olarak ekler.
// Türkçe Yorum: Raportör, komisyon_bekliyor durumundaki projeleri veya beklemedeki talepleri aktif toplantıya ekler.
func (r *KomisyonToplantiRepository) AddProjeToToplanti(toplantiID, projeID, talepID int, gundemTipi, talepTipi string, gundemSirasi, ekleyenID int) error {
	if gundemTipi == "" {
		if talepID > 0 {
			gundemTipi = "talep"
		} else {
			gundemTipi = "basvuru"
		}
	}
	query := `
		INSERT INTO komisyon_toplanti_proje (toplanti_id, proje_id, talep_id, gundem_tipi, talep_tipi, gundem_sirasi, ekleyen_id)
		VALUES ($1, $2, NULLIF($3, 0), $4, $5, NULLIF($6, 0), $7)
	`
	_, err := r.DB.Exec(query, toplantiID, projeID, talepID, gundemTipi, talepTipi, gundemSirasi, ekleyenID)
	if err != nil {
		return fmt.Errorf("gündem maddesi toplantıya eklenemedi: %w", err)
	}
	return nil
}

// RemoveProjeFromToplanti bir projeyi veya talebi toplantıdan çıkarır.
// Türkçe Yorum: Sadece 'bekliyor' durumundaki gündem maddelerini kaldırır.
func (r *KomisyonToplantiRepository) RemoveProjeFromToplanti(toplantiID, projeID, talepID int) error {
	var err error
	var res sql.Result
	if talepID > 0 {
		query := `DELETE FROM komisyon_toplanti_proje WHERE toplanti_id = $1 AND proje_id = $2 AND talep_id = $3 AND karar = 'bekliyor'`
		res, err = r.DB.Exec(query, toplantiID, projeID, talepID)
	} else {
		query := `DELETE FROM komisyon_toplanti_proje WHERE toplanti_id = $1 AND proje_id = $2 AND (talep_id IS NULL OR talep_id = 0) AND karar = 'bekliyor'`
		res, err = r.DB.Exec(query, toplantiID, projeID)
	}
	if err != nil {
		return fmt.Errorf("gündem maddesi toplantıdan çıkarılamadı: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("gündem maddesi bulunamadı veya karar verilmiş maddeler kaldırılamaz")
	}
	return nil
}

// GetProjectsByToplanti bir toplantıdaki tüm proje ve talepleri detaylarıyla getirir.
// Türkçe Yorum: Toplantı detay modalı ve PDF tutanağı için kullanılır.
func (r *KomisyonToplantiRepository) GetProjectsByToplanti(toplantiID int) ([]*models.KomisyonToplantisiProje, error) {
	query := `
		SELECT
			ktp.id, ktp.toplanti_id, ktp.proje_id, COALESCE(ktp.talep_id, 0),
			COALESCE(ktp.gundem_tipi, 'basvuru'), COALESCE(ktp.talep_tipi, ''),
			ktp.gundem_sirasi, ktp.karar, COALESCE(ktp.karar_aciklamasi,''),
			ktp.karar_tarihi, ktp.ekleyen_id, ktp.olusturma_tarihi,
			COALESCE(p.proje_kodu,''),
			COALESCE((SELECT pb.baslik FROM proje_baslik pb WHERE pb.proje_id=p.proje_id AND pb.dil_kodu='tr' LIMIT 1),'Başlıksız'),
			COALESCE(u.ad||' '||u.soyad,'Bilinmiyor'),
			COALESCE(NULLIF(TRIM(COALESCE(ud.unvan,'')), ''), COALESCE(NULLIF(TRIM(COALESCE(u.unvan,'')), ''), '')),
			COALESCE(pd.durum_adi,'')
		FROM komisyon_toplanti_proje ktp
		JOIN proje p    ON p.proje_id = ktp.proje_id
		JOIN uye u      ON u.uye_id  = p.koordinator_id
		LEFT JOIN uye_detay ud ON ud.uye_id = u.uye_id
		LEFT JOIN proje_durum pd ON pd.durum_id = p.durum_id
		WHERE ktp.toplanti_id = $1
		ORDER BY COALESCE(ktp.gundem_sirasi, 9999), ktp.olusturma_tarihi
	`
	rows, err := r.DB.Query(query, toplantiID)
	if err != nil {
		return nil, fmt.Errorf("toplantı projeleri sorgulanamadı: %w", err)
	}
	defer rows.Close()

	var projeler []*models.KomisyonToplantisiProje
	for rows.Next() {
		var kp models.KomisyonToplantisiProje
		err := rows.Scan(
			&kp.ID, &kp.ToplantiID, &kp.ProjeID, &kp.TalepID,
			&kp.GundemTipi, &kp.TalepTipi,
			&kp.GundemSirasi, &kp.Karar, &kp.KararAciklamasi,
			&kp.KararTarihi, &kp.EkleyenID, &kp.OlusturmaTarihi,
			&kp.ProjeKodu, &kp.ProjeBaslik, &kp.YurutucuAd, &kp.YurutucuUnvan, &kp.MevcutDurum,
		)
		if err != nil {
			return nil, err
		}
		if kp.TalepID > 0 && kp.TalepTipi != "" {
			r.fillTalepDetayForToplanti(&kp)
		} else {
			kp.DetayMetin = "Yeni Proje Başvurusu"
		}
		projeler = append(projeler, &kp)
	}
	return projeler, nil
}

// fillTalepDetayForToplanti talep gündem maddesinin etiket ve detay bilgilerini doldurur.
func (r *KomisyonToplantiRepository) fillTalepDetayForToplanti(kp *models.KomisyonToplantisiProje) {
	query := buildTalepUnionQuery(fmt.Sprintf("AND t.id = %d AND t.proje_id = %d", kp.TalepID, kp.ProjeID))
	rows, err := r.DB.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()

	list, err := scanTalepOzetList(rows)
	if err == nil && len(list) > 0 {
		t := list[0]
		kp.TalepTipiEtiketi = t.TalepTipiEtiketi
		kp.DetayMetin = formatTalepDetayMetin(t.TalepTipiEtiketi, t.Detay, t.Gerekce)
	}
}

// GetToplantilerByProje bir projenin görüşüldüğü tüm toplantıları listeler.
func (r *KomisyonToplantiRepository) GetToplantilerByProje(projeID int) ([]*models.KomisyonToplantisiProje, error) {
	query := `
		SELECT
			ktp.id, ktp.toplanti_id, ktp.proje_id, COALESCE(ktp.talep_id, 0),
			COALESCE(ktp.gundem_tipi, 'basvuru'), COALESCE(ktp.talep_tipi, ''),
			ktp.gundem_sirasi, ktp.karar, COALESCE(ktp.karar_aciklamasi,''),
			ktp.karar_tarihi, ktp.ekleyen_id, ktp.olusturma_tarihi,
			COALESCE(kt.toplanti_no,''), '', '', ''
		FROM komisyon_toplanti_proje ktp
		JOIN komisyon_toplantisi kt ON kt.toplanti_id = ktp.toplanti_id
		WHERE ktp.proje_id = $1
		ORDER BY kt.tarih DESC
	`
	rows, err := r.DB.Query(query, projeID)
	if err != nil {
		return nil, fmt.Errorf("projenin toplantıları sorgulanamadı: %w", err)
	}
	defer rows.Close()

	var liste []*models.KomisyonToplantisiProje
	for rows.Next() {
		var kp models.KomisyonToplantisiProje
		err := rows.Scan(
			&kp.ID, &kp.ToplantiID, &kp.ProjeID, &kp.TalepID,
			&kp.GundemTipi, &kp.TalepTipi,
			&kp.GundemSirasi, &kp.Karar, &kp.KararAciklamasi,
			&kp.KararTarihi, &kp.EkleyenID, &kp.OlusturmaTarihi,
			&kp.ProjeKodu, &kp.ProjeBaslik, &kp.YurutucuAd, &kp.MevcutDurum,
		)
		if err != nil {
			return nil, err
		}
		liste = append(liste, &kp)
	}
	return liste, nil
}

// SetProjeKarar toplantıdaki bir proje/talep için karar kaydeder.
// Türkçe Yorum: Raportör onay/red/erteleme kararını bu fonksiyon aracılığıyla kayıt altına alır.
func (r *KomisyonToplantiRepository) SetProjeKarar(toplantiID, projeID, talepID int, karar, aciklama string) error {
	var err error
	var res sql.Result
	if talepID > 0 {
		query := `
			UPDATE komisyon_toplanti_proje
			SET karar = $4, karar_aciklamasi = $5, karar_tarihi = CURRENT_TIMESTAMP
			WHERE toplanti_id = $1 AND proje_id = $2 AND talep_id = $3
		`
		res, err = r.DB.Exec(query, toplantiID, projeID, talepID, karar, aciklama)
	} else {
		query := `
			UPDATE komisyon_toplanti_proje
			SET karar = $3, karar_aciklamasi = $4, karar_tarihi = CURRENT_TIMESTAMP
			WHERE toplanti_id = $1 AND proje_id = $2 AND (talep_id IS NULL OR talep_id = 0)
		`
		res, err = r.DB.Exec(query, toplantiID, projeID, karar, aciklama)
	}
	if err != nil {
		return fmt.Errorf("gündem kararı kaydedilemedi: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("gündem maddesi bu toplantıda bulunamadı")
	}
	return nil
}

// SyncToplantiDurumFromProjeler gündem projelerinin karar durumuna göre toplantı durumunu günceller.
// Türkçe Yorum: Bekleyen/ertelenmiş proje varsa planli; tüm nihai kararlar verildiyse tamamlandi.
func (r *KomisyonToplantiRepository) SyncToplantiDurumFromProjeler(toplantiID int) (string, error) {
	var toplam, bekleyen int
	err := r.DB.QueryRow(`
		SELECT
			COUNT(*)::int,
			COUNT(*) FILTER (WHERE karar IN ('bekliyor', 'ertelendi'))::int
		FROM komisyon_toplanti_proje
		WHERE toplanti_id = $1
	`, toplantiID).Scan(&toplam, &bekleyen)
	if err != nil {
		return "", fmt.Errorf("toplantı proje durumları okunamadı: %w", err)
	}
	if toplam == 0 {
		return "", nil
	}

	yeniDurum := "tamamlandi"
	if bekleyen > 0 {
		yeniDurum = "planli"
	}

	_, err = r.DB.Exec(`
		UPDATE komisyon_toplantisi
		SET durum = $2
		WHERE toplanti_id = $1
	`, toplantiID, yeniDurum)
	if err != nil {
		return "", fmt.Errorf("toplantı durumu güncellenemedi: %w", err)
	}
	return yeniDurum, nil
}

// GetToplantiBelgeDetay PDF tutanağı için toplantı + katılımcılar + projeler bütününü döner.
func (r *KomisyonToplantiRepository) GetToplantiBelgeDetay(toplantiID int) (*models.KomisyonToplantiBelge, error) {
	komisyonRepo := &KomisyonRepository{DB: r.DB}
	toplanti, err := komisyonRepo.GetMeetingByID(toplantiID)
	if err != nil {
		return nil, err
	}
	if toplanti == nil {
		return nil, fmt.Errorf("toplantı bulunamadı")
	}
	projeler, err := r.GetProjectsByToplanti(toplantiID)
	if err != nil {
		return nil, err
	}
	return &models.KomisyonToplantiBelge{
		Toplanti:     toplanti,
		Katilimcilar: toplanti.Katilimcilar,
		Projeler:     projeler,
	}, nil
}

// GetBekleyenProjeler komisyon_bekliyor durumundaki projeleri ve beklemedeki tüm talepleri listeler.
// Türkçe Yorum: Toplantı gündemine eklenecek aday proje ve talepleri döner.
func (r *KomisyonToplantiRepository) GetBekleyenProjeler() ([]*models.KomisyonBekleyenProje, error) {
	var liste []*models.KomisyonBekleyenProje

	// 1. Yeni Proje Başvuruları (komisyon_bekliyor)
	basvuruQuery := `
		SELECT
			p.proje_id,
			COALESCE(p.proje_kodu, ''),
			COALESCE((SELECT pb.baslik FROM proje_baslik pb WHERE pb.proje_id = p.proje_id AND pb.dil_kodu = 'tr' LIMIT 1), 'Başlıksız'),
			COALESCE(u.ad||' '||u.soyad, 'Bilinmiyor'),
			COALESCE(NULLIF(TRIM(COALESCE(ud.unvan,'')), ''), COALESCE(NULLIF(TRIM(COALESCE(u.unvan,'')), ''), '')),
			COALESCE(pbt.bap_turu, ''),
			COALESCE((SELECT SUM(b.toplam_fiyat) FROM proje_butce b WHERE b.proje_id = p.proje_id), 0)
		FROM proje p
		JOIN proje_durum pd ON pd.durum_id = p.durum_id
		JOIN uye u ON u.uye_id = p.koordinator_id
		LEFT JOIN uye_detay ud ON ud.uye_id = u.uye_id
		LEFT JOIN proje_bap_turu pbt ON pbt.bap_turu_id = p.bap_turu_id
		WHERE pd.durum_adi = 'komisyon_bekliyor'
		ORDER BY p.proje_id DESC
	`
	rows, err := r.DB.Query(basvuruQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p models.KomisyonBekleyenProje
			p.GundemTipi = "basvuru"
			if err := rows.Scan(&p.ProjeID, &p.ProjeKodu, &p.ProjeBaslik, &p.YurutucuAd, &p.YurutucuUnvan, &p.BapTuru, &p.ToplamButce); err == nil {
				p.DetayMetin = "Yeni Proje Başvurusu (Komisyon Onayı Bekliyor)"
				liste = append(liste, &p)
			}
		}
	}

	// 2. Proje Talepleri (Fasıl Aktarımı, Ek Süre, Ek Bütçe, vb.)
	talepRepo := &TalepRepository{DB: r.DB}
	talepler, err := talepRepo.GetAllTalepler(true) // sadeceBekleyen = true ('beklemede')
	if err == nil {
		for _, t := range talepler {
			p := &models.KomisyonBekleyenProje{
				ProjeID:          t.ProjeID,
				TalepID:          t.ID,
				GundemTipi:       "talep",
				TalepTipi:        t.TalepTipi,
				TalepTipiEtiketi: t.TalepTipiEtiketi,
				ProjeKodu:        t.ProjeKodu,
				ProjeBaslik:      t.ProjeBaslik,
				YurutucuAd:       t.TalepEdenAd,
				Gerekce:          t.Gerekce,
				Detay:            t.Detay,
			}
			p.DetayMetin = formatTalepDetayMetin(t.TalepTipiEtiketi, t.Detay, t.Gerekce)
			liste = append(liste, p)
		}
	}

	return liste, nil
}

// formatTalepDetayMetin talep türü ve detay map'inden kısa ve öz özet metin üretir.
func formatTalepDetayMetin(etiket string, detay map[string]interface{}, gerekce string) string {
	metin := etiket
	if len(detay) > 0 {
		switch etiket {
		case "Fasıl Aktarımı":
			kaynak, _ := detay["kaynak_kalem"].(string)
			hedef, _ := detay["hedef_kalem"].(string)
			tutar, _ := detay["tutar_tl"].(float64)
			metin = fmt.Sprintf("Fasıl Aktarımı: %s → %s (%.2f TL)", kaynak, hedef, tutar)
		case "Ek Süre":
			ay, _ := detay["ek_sure_ay"].(float64)
			metin = fmt.Sprintf("Ek Süre Talebi: %.0f Ay", ay)
		case "Ek Bütçe":
			kalem, _ := detay["butce_kalemi"].(string)
			tutar, _ := detay["tutar_tl"].(float64)
			metin = fmt.Sprintf("Ek Bütçe Talebi: %s (%.2f TL)", kalem, tutar)
		case "Araştırmacı Değişikliği":
			islem, _ := detay["islem_turu"].(string)
			kisi, _ := detay["arastirmaci_adi"].(string)
			metin = fmt.Sprintf("Araştırmacı (%s): %s", islem, kisi)
		case "Bursiyer İşlemi":
			islem, _ := detay["islem_turu"].(string)
			kisi, _ := detay["bursiyer_adi"].(string)
			metin = fmt.Sprintf("Bursiyer (%s): %s", islem, kisi)
		case "Proje Dondurma":
			ay, _ := detay["dondurma_sure_ay"].(float64)
			metin = fmt.Sprintf("Proje Dondurma Talebi: %.0f Ay", ay)
		case "Malzeme Güncelleme":
			tanim, _ := detay["guncelleme_tanimi"].(string)
			metin = fmt.Sprintf("Malzeme Güncelleme: %s", tanim)
		case "Avans":
			kalem, _ := detay["butce_kalemi"].(string)
			tutar, _ := detay["tutar_tl"].(float64)
			metin = fmt.Sprintf("Avans Talebi: %s (%.2f TL)", kalem, tutar)
		case "Bilgi Değişimi":
			tanim, _ := detay["degisiklik_tanimi"].(string)
			metin = fmt.Sprintf("Bilgi Değişikliği: %s", tanim)
		}
	}
	if gerekce != "" {
		metin += fmt.Sprintf(" - Gerekçe: %s", gerekce)
	}
	return metin
}
