package repository

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"bap_ai/backend/models"
)

// EimzaRepository yapısı, e-imza işlemlerinin veritabanı katmanıdır.
// Türkçe Yorum: E-İmza veritabanı sorgularından sorumlu yapı.
type EimzaRepository struct {
	DB *sql.DB
}

// NewEimzaRepository yeni bir EimzaRepository oluşturur.
// Türkçe Yorum: Yeni EimzaRepository nesnesini başlatır.
func NewEimzaRepository(db *sql.DB) *EimzaRepository {
	return &EimzaRepository{DB: db}
}

// GetPendingSignatures kullanıcı rolüne göre imza bekleyen projeleri listeler.
// Türkçe Yorum: Kullanıcının sahip olduğu tüm rollere (virgülle ayrılmış olabilir) göre imza bekleyen projeleri çeker ve tekil olarak birleştirir.
func (r *EimzaRepository) GetPendingSignatures(uyeID int, rol string) ([]models.Proje, error) {
	roles := strings.Split(rol, ",")
	var allProjects []models.Proje
	seen := make(map[int]bool)

	for _, singleRol := range roles {
		singleRol = strings.TrimSpace(singleRol)
		var query string
		var args []interface{}

		switch singleRol {
		case "akademisyen":
			// Akademisyenin (yürütücü) taslak durumundaki kendi projeleri imza bekler
			query = `
				SELECT p.proje_id, COALESCE(p.baslik_tr, ''), COALESCE(p.baslik_en, ''), COALESCE(p.sure_ay, 0), COALESCE(p.toplam_butce, 0), p.olusturma_tarihi,
				       COALESCE(pd.durum_adi, 'taslak'), COALESCE(pbt.bap_turu, 'Münferit'),
				       COALESCE(u.ad || ' ' || u.soyad, '') as koordinator_ad_soyad
				FROM proje p
				LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
				LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
				LEFT JOIN uye u ON p.koordinator_id = u.uye_id
				WHERE p.koordinator_id = $1 AND pd.durum_adi IN ('taslak', 'revizyon')
				  AND NOT EXISTS (SELECT 1 FROM proje_imza pi WHERE pi.proje_id = p.proje_id AND pi.uye_id = $1)
				ORDER BY p.guncelleme_tarihi DESC
			`
			args = append(args, uyeID)

		case "dekan":
			// Dekan onayı bekleyen tüm projeler
			query = `
				SELECT p.proje_id, COALESCE(p.baslik_tr, ''), COALESCE(p.baslik_en, ''), COALESCE(p.sure_ay, 0), COALESCE(p.toplam_butce, 0), p.olusturma_tarihi,
				       COALESCE(pd.durum_adi, 'dekan_onayi_bekliyor'), COALESCE(pbt.bap_turu, 'Münferit'),
				       COALESCE(u.ad || ' ' || u.soyad, '') as koordinator_ad_soyad
				FROM proje p
				LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
				LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
				LEFT JOIN uye u ON p.koordinator_id = u.uye_id
				WHERE pd.durum_adi = 'dekan_onayi_bekliyor'
				  AND NOT EXISTS (SELECT 1 FROM proje_imza pi WHERE pi.proje_id = p.proje_id AND pi.uye_id = $1)
				ORDER BY p.guncelleme_tarihi DESC
			`
			args = append(args, uyeID)

		case "komisyon":
			// Komisyon onayı bekleyen tüm projeler
			query = `
				SELECT p.proje_id, COALESCE(p.baslik_tr, ''), COALESCE(p.baslik_en, ''), COALESCE(p.sure_ay, 0), COALESCE(p.toplam_butce, 0), p.olusturma_tarihi,
				       COALESCE(pd.durum_adi, 'komisyon_bekliyor'), COALESCE(pbt.bap_turu, 'Münferit'),
				       COALESCE(u.ad || ' ' || u.soyad, '') as koordinator_ad_soyad
				FROM proje p
				LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
				LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
				LEFT JOIN uye u ON p.koordinator_id = u.uye_id
				WHERE pd.durum_adi = 'komisyon_bekliyor'
				  AND NOT EXISTS (SELECT 1 FROM proje_imza pi WHERE pi.proje_id = p.proje_id AND pi.uye_id = $1)
				ORDER BY p.guncelleme_tarihi DESC
			`
			args = append(args, uyeID)

		case "tto":
			// TTO onayı bekleyen (aktif) tüm projeler
			query = `
				SELECT p.proje_id, COALESCE(p.baslik_tr, ''), COALESCE(p.baslik_en, ''), COALESCE(p.sure_ay, 0), COALESCE(p.toplam_butce, 0), p.olusturma_tarihi,
				       COALESCE(pd.durum_adi, 'tto_aktif'), COALESCE(pbt.bap_turu, 'Münferit'),
				       COALESCE(u.ad || ' ' || u.soyad, '') as koordinator_ad_soyad
				FROM proje p
				LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
				LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
				LEFT JOIN uye u ON p.koordinator_id = u.uye_id
				WHERE pd.durum_adi = 'tto_aktif'
				  AND NOT EXISTS (SELECT 1 FROM proje_imza pi WHERE pi.proje_id = p.proje_id AND pi.uye_id = $1)
				ORDER BY p.guncelleme_tarihi DESC
			`
			args = append(args, uyeID)

		case "admin":
			// Admin tüm süreç aşamalarındaki imzalanmamış projeleri görebilir
			query = `
				SELECT p.proje_id, COALESCE(p.baslik_tr, ''), COALESCE(p.baslik_en, ''), COALESCE(p.sure_ay, 0), COALESCE(p.toplam_butce, 0), p.olusturma_tarihi,
				       COALESCE(pd.durum_adi, 'taslak'), COALESCE(pbt.bap_turu, 'Münferit'),
				       COALESCE(u.ad || ' ' || u.soyad, '') as koordinator_ad_soyad
				FROM proje p
				LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
				LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
				LEFT JOIN uye u ON p.koordinator_id = u.uye_id
				WHERE pd.durum_adi IN ('taslak', 'revizyon', 'dekan_onayi_bekliyor', 'komisyon_bekliyor', 'tto_aktif')
				  AND NOT EXISTS (SELECT 1 FROM proje_imza pi WHERE pi.proje_id = p.proje_id AND pi.uye_id = $1)
				ORDER BY p.guncelleme_tarihi DESC
			`
			args = append(args, uyeID)

		default:
			continue
		}

		rows, err := r.DB.Query(query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var p models.Proje
			err := rows.Scan(
				&p.ProjeID, &p.BaslikTr, &p.BaslikEn, &p.SureAy, &p.ToplamButce, &p.OlusturmaTarihi,
				&p.DurumAdi, &p.BapTuru, &p.KoordinatorAdSoyad,
			)
			if err != nil {
				return nil, err
			}
			if !seen[p.ProjeID] {
				seen[p.ProjeID] = true
				allProjects = append(allProjects, p)
			}
		}
	}
	return allProjects, nil
}

// GetSignedDocuments kullanıcının imzaladığı projeleri ve imza detaylarını listeler.
// Türkçe Yorum: Kullanıcının e-imza attığı projelerin geçmişini listeler.
func (r *EimzaRepository) GetSignedDocuments(uyeID int) ([]models.Proje, error) {
	query := `
		SELECT p.proje_id, COALESCE(p.baslik_tr, ''), COALESCE(p.baslik_en, ''), COALESCE(p.sure_ay, 0), COALESCE(p.toplam_butce, 0), pi.imza_tarihi,
		       COALESCE(pd.durum_adi, 'taslak'), COALESCE(pbt.bap_turu, 'Münferit'),
		       COALESCE(u.ad || ' ' || u.soyad, '') as koordinator_ad_soyad
		FROM proje p
		INNER JOIN proje_imza pi ON p.proje_id = pi.proje_id
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN uye u ON p.koordinator_id = u.uye_id
		WHERE pi.uye_id = $1
		ORDER BY pi.imza_tarihi DESC
	`
	rows, err := r.DB.Query(query, uyeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projeler []models.Proje
	for rows.Next() {
		var p models.Proje
		err := rows.Scan(
			&p.ProjeID, &p.BaslikTr, &p.BaslikEn, &p.SureAy, &p.ToplamButce, &p.OlusturmaTarihi, // İmza tarihini OlusturmaTarihi yerine set ediyoruz
			&p.DurumAdi, &p.BapTuru, &p.KoordinatorAdSoyad,
		)
		if err != nil {
			return nil, err
		}
		projeler = append(projeler, p)
	}
	return projeler, nil
}

// SignDocument e-imza işlemini gerçekleştirir ve proje durumunu günceller.
// Türkçe Yorum: E-imza atma kaydını oluşturur ve projeyi bir sonraki iş akışı adımına taşır.
func (r *EimzaRepository) SignDocument(projeID, uyeID int, rol, imzaciAdSoyad, yontem, pin, ip, userAgent string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Projenin mevcut durumunu bul
	var mevcutDurum string
	var durumID int
	queryProje := `
		SELECT p.durum_id, pd.durum_adi 
		FROM proje p
		INNER JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE p.proje_id = $1
	`
	err = tx.QueryRow(queryProje, projeID).Scan(&durumID, &mevcutDurum)
	if err != nil {
		return fmt.Errorf("proje durumu alınamadı: %v", err)
	}

	// 2. Bir sonraki durumu belirle (Switch-case akışı)
	var yeniDurum string
	var logAciklama string
	var imzaRol string

	switch mevcutDurum {
	case "taslak", "revizyon":
		yeniDurum = "dekan_onayi_bekliyor"
		logAciklama = fmt.Sprintf("Proje yürütücüsü %s tarafından e-imza ile imzalandı. Başvuru dekan onayına sunuldu. (%s)", imzaciAdSoyad, yontem)
		imzaRol = "akademisyen"
	case "dekan_onayi_bekliyor":
		yeniDurum = "komisyon_bekliyor"
		logAciklama = fmt.Sprintf("Dekan %s tarafından e-imza ile imzalandı. Komisyon onayına sunuldu. (%s)", imzaciAdSoyad, yontem)
		imzaRol = "dekan"
	case "komisyon_bekliyor":
		yeniDurum = "tto_aktif"
		logAciklama = fmt.Sprintf("Komisyon üyesi %s tarafından e-imza ile imzalandı. TTO onayına sunuldu. (%s)", imzaciAdSoyad, yontem)
		imzaRol = "komisyon"
	case "tto_aktif":
		yeniDurum = "tamamlandi"
		logAciklama = fmt.Sprintf("TTO Yetkilisi %s tarafından e-imza ile onaylandı ve imzalandı. Proje başarıyla tamamlandı. (%s)", imzaciAdSoyad, yontem)
		imzaRol = "tto"
	default:
		// Admin veya diğer durumlar için
		yeniDurum = mevcutDurum
		logAciklama = fmt.Sprintf("Proje %s tarafından e-imza ile imzalandı. (%s)", imzaciAdSoyad, yontem)
		imzaRol = rol
	}

	// 3. E-İmza hash token üret (simüle edilmiş)
	hashData := fmt.Sprintf("%d-%d-%s-%s-%d", projeID, uyeID, pin, yontem, time.Now().UnixNano())
	tokenHash := sha256.Sum256([]byte(hashData))
	imzaToken := fmt.Sprintf("E-IMZA-%x", tokenHash[:16])

	// 4. Imza kaydını tabloya ekle
	insertQuery := `
		INSERT INTO proje_imza (proje_id, uye_id, imzaci_ad_soyad, rol, imza_token, ip_adresi, tarayici_bilgisi)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (proje_id, uye_id) DO NOTHING
	`
	_, err = tx.Exec(insertQuery, projeID, uyeID, imzaciAdSoyad, imzaRol, imzaToken, ip, userAgent)
	if err != nil {
		return fmt.Errorf("e-imza kaydı veritabanına eklenemedi: %v", err)
	}

	// 5. Projenin yeni durum ID'sini al
	var yeniDurumID int
	err = tx.QueryRow(`SELECT durum_id FROM proje_durum WHERE durum_adi = $1`, yeniDurum).Scan(&yeniDurumID)
	if err != nil {
		return fmt.Errorf("yeni durum (%s) bulunamadı: %v", yeniDurum, err)
	}

	// 6. Proje durumunu güncelle
	_, err = tx.Exec(`UPDATE proje SET durum_id = $1, guncelleme_tarihi = CURRENT_TIMESTAMP WHERE proje_id = $2`, yeniDurumID, projeID)
	if err != nil {
		return fmt.Errorf("proje durumu güncellenemedi: %v", err)
	}

	// 7. Süreç geçmişi kaydı (Log) ekle
	logQuery := `
		INSERT INTO proje_surec_gecmisi (proje_id, islem_yapan_id, baslangic_durum, hedef_durum, aciklama)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.Exec(logQuery, projeID, uyeID, mevcutDurum, yeniDurum, logAciklama)
	if err != nil {
		return fmt.Errorf("süreç geçmişi log kaydı eklenemedi: %v", err)
	}

	return tx.Commit()
}
