package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"bap_ai/backend/models"
)

// AdminRepository, admin işlemleri için veritabanı erişimini sağlar.
type AdminRepository struct {
	DB *sql.DB
}

// NewAdminRepository, yeni bir AdminRepository örneği oluşturur.
func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{DB: db}
}

// GetTotalUsersCount, sistemdeki toplam aktif kullanıcı sayısını döner.
func (r *AdminRepository) GetTotalUsersCount() (int64, error) {
	var count int64
	err := r.DB.QueryRow("SELECT COUNT(*) FROM uye WHERE aktif_mi = true").Scan(&count)
	if err != nil {
		log.Printf("GetTotalUsersCount hatası: %v", err)
	}
	return count, err
}

// GetAllUsers, sistemdeki tüm kullanıcıları tüm bilgileriyle döner.
// uye ve uye_detay tabloları JOIN edilerek detay bilgiler de dahil edilir.
func (r *AdminRepository) GetAllUsers() ([]models.Uye, error) {
	var users []models.Uye
	rows, err := r.DB.Query(`
		SELECT u.uye_id, 
		       COALESCE((
		           SELECT string_agg(srt.rol_adi, ',') 
		           FROM sistem_rol sr 
		           INNER JOIN sistem_rol_tanimlama srt ON sr.sistem_rol_id = srt.rol_id 
		           WHERE sr.uye_id = u.uye_id
		       ), d.rol, ''),
		       COALESCE(d.unvan, ''), u.ad, u.soyad,
		       COALESCE(d.bolum, ''), COALESCE(d.telefon, ''), u.eposta,
		       COALESCE(d.izu_uyesi, FALSE), u.aktif_mi, u.olusturma_tarihi
		FROM uye u
		LEFT JOIN uye_detay d ON u.uye_id = d.uye_id
		ORDER BY u.uye_id
	`)
	if err != nil {
		log.Printf("GetAllUsers hatası: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u models.Uye
		if err := rows.Scan(&u.UyeID, &u.Rol, &u.Unvan, &u.Ad, &u.Soyad, &u.Bolum, &u.Telefon, &u.Eposta, &u.IzuUyesi, &u.AktifMi, &u.OlusturmaTarihi); err == nil {
			users = append(users, u)
		}
	}
	return users, nil
}

// GetAllProjects, sistemdeki tüm projeleri döner.
func (r *AdminRepository) GetAllProjects() ([]models.Proje, error) {
	var projes []models.Proje
	rows, err := r.DB.Query(`
		SELECT p.proje_id, COALESCE(p.proje_kodu, ''), (SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'), COALESCE(pd.durum_adi, 'taslak'),
		       COALESCE(pbt.bap_turu, 'Münferit'), p.olusturma_tarihi
		FROM proje p
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
	`)
	if err != nil {
		log.Printf("GetAllProjects hatası: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Proje
		if err := rows.Scan(&p.ProjeID, &p.ProjeKodu, &p.BaslikTr, &p.DurumAdi, &p.BapTuru, &p.OlusturmaTarihi); err == nil {
			projes = append(projes, p)
		}
	}
	return projes, nil
}

// syncRolAlanlari rol değişikliklerini uye, uye_detay ve sistem_rol tablolarına
// eş zamanlı olarak yansıtır (transaction içinde çağrılmak üzere tasarlanmıştır).
// Türkçe Yorum: Tek kaynak sistem_rol tablosudur; uye.rol ve uye_detay.rol geriye dönük
// uyumluluk için senkronize tutulmaktadır.
func syncRolAlanlari(tx interface {
	Exec(query string, args ...interface{}) (interface{ RowsAffected() (int64, error) }, error)
}, uyeID int, rolAdi string, tumRoller []string) error {
	return nil // tx interface helper — asıl senkronizasyon aşağıdaki fonksiyonlarda yapılır
}

// syncRolTx, transaction içinde uye, uye_detay ve sistem_rol tablolarını senkronize eder.
// Türkçe Yorum: Rol bilgisi bu 3 tabloda tutulmaktadır. İlk rol legacy alanlar için kullanılır,
// tüm roller ise sistem_rol tablosuna yazılır.
func (r *AdminRepository) syncRolTx(tx interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}, uyeID int, tumRoller []string) error {
	ilkRol := ""
	if len(tumRoller) > 0 {
		ilkRol = strings.TrimSpace(tumRoller[0])
	}

	// 1. uye.rol güncelle (legacy uyumluluk)
	if _, err := tx.Exec(`UPDATE uye SET rol = $1 WHERE uye_id = $2`, ilkRol, uyeID); err != nil {
		log.Printf("syncRolTx uye.rol hatası (uye_id=%d): %v", uyeID, err)
		return err
	}

	// 2. uye_detay.rol güncelle (legacy uyumluluk)
	if _, err := tx.Exec(`UPDATE uye_detay SET rol = $1, guncelleme_tarihi = CURRENT_TIMESTAMP WHERE uye_id = $2`, ilkRol, uyeID); err != nil {
		log.Printf("syncRolTx uye_detay.rol hatası (uye_id=%d): %v", uyeID, err)
		return err
	}

	// 3. sistem_rol tablosunu temizle ve tüm rolleri yeniden ata (asıl kaynak)
	if _, err := tx.Exec(`DELETE FROM sistem_rol WHERE uye_id = $1`, uyeID); err != nil {
		log.Printf("syncRolTx sistem_rol temizleme hatası (uye_id=%d): %v", uyeID, err)
		return err
	}
	for _, rName := range tumRoller {
		rName = strings.TrimSpace(rName)
		if rName == "" {
			continue
		}
		_, err := tx.Exec(`
			INSERT INTO sistem_rol (uye_id, sistem_rol_id)
			SELECT $1, rol_id FROM sistem_rol_tanimlama WHERE rol_adi = $2
			ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING
		`, uyeID, rName)
		if err != nil {
			log.Printf("syncRolTx sistem_rol atama hatası (uye_id=%d, rol=%s): %v", uyeID, rName, err)
			return err
		}
	}
	return nil
}

// UpdateUserRole, bir kullanıcının rolünü günceller.
// Türkçe Yorum: Rol senkronizasyonu syncRolTx üzerinden yapılır; 3 tabloyu tek bir transaction ile günceller.
func (r *AdminRepository) UpdateUserRole(uyeID int, rolAdi string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		log.Printf("UpdateUserRole transaction başlatma hatası: %v", err)
		return err
	}
	defer tx.Rollback()

	if err := r.syncRolTx(tx, uyeID, strings.Split(rolAdi, ",")); err != nil {
		return err
	}
	return tx.Commit()
}

// UpdateUserStatus, bir kullanıcının aktiflik durumunu günceller.
func (r *AdminRepository) UpdateUserStatus(uyeID int, isActive bool) error {
	_, err := r.DB.Exec("UPDATE uye SET aktif_mi = $1 WHERE uye_id = $2", isActive, uyeID)
	if err != nil {
		log.Printf("UpdateUserStatus hatası: %v", err)
	}
	return err
}

// UpdateProjectStatus, projenin genel statüsünü günceller.
func (r *AdminRepository) UpdateProjectStatus(projeID int, durum string) error {
	// Durum adına göre durum_id bul ve güncelle
	_, err := r.DB.Exec(`
		UPDATE proje SET durum_id = (SELECT durum_id FROM proje_durum WHERE durum_adi = $1)
		WHERE proje_id = $2
	`, durum, projeID)
	if err != nil {
		log.Printf("UpdateProjectStatus hatası: %v", err)
	}
	return err
}

// GetProjectStats, genel sistemdeki tüm proje istatistiklerini hesaplayıp döner.
func (r *AdminRepository) GetProjectStats() (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	err := r.DB.QueryRow(`
		SELECT COUNT(*) FROM proje p
		INNER JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pd.durum_adi IN ('onaylandi', 'tto_aktif', 'yururlukte')
	`).Scan(&stats.AktifProje)
	if err != nil {
		return nil, err
	}

	err = r.DB.QueryRow(`
		SELECT COUNT(*) FROM proje p
		INNER JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pd.durum_adi = 'incelemede'
	`).Scan(&stats.OnayBekleyen)
	if err != nil {
		return nil, err
	}

	err = r.DB.QueryRow(`
		SELECT COUNT(*) FROM proje p
		INNER JOIN proje_durum pd ON p.durum_id = pd.durum_id
		WHERE pd.durum_adi = 'tamamlandi'
	`).Scan(&stats.Tamamlanan)
	if err != nil {
		return nil, err
	}

	err = r.DB.QueryRow("SELECT COALESCE(SUM((SELECT SUM(toplam_fiyat) FROM proje_butce WHERE proje_id = p.proje_id)), 0) FROM proje p").Scan(&stats.ToplamButce)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// ReviewDetail, Admin sayfasında hakem yorumlarını göstermek için veri yapısı.
type ReviewDetail struct {
	DegerlendirmeID int    `json:"degerlendirme_id"`
	HakemAdSoyad    string `json:"hakem_ad_soyad"`
	Puan            int    `json:"puan"`
	Yorum           string `json:"yorum"`
	Durum           string `json:"durum"`
}

// TakimUyeDetail, Admin detay sayfasında takım üyelerini göstermek için veri yapısı.
type TakimUyeDetail struct {
	UyeID    int    `json:"uye_id"`
	AdSoyad  string `json:"ad_soyad"`
	Rol      string `json:"rol"`
	ProjeRol string `json:"proje_rol"`
}

// IsPaketiDetail, Admin detay sayfasında iş paketlerini göstermek için veri yapısı.
type IsPaketiDetail struct {
	PaketID    int    `json:"paket_id"`
	PaketAdi   string `json:"paket_adi"`
	PaketAmaci string `json:"paket_amaci"`
	BaslangicAy int    `json:"baslangic_ay"`
	BitisAy    int    `json:"bitis_ay"`
}

// RiskDetail, Admin detay sayfasında risk yönetimini göstermek için veri yapısı.
type RiskDetail struct {
	RiskID         int    `json:"risk_id"`
	RiskAciklamasi string `json:"risk_aciklamasi"`
	CozumPlani     string `json:"cozum_plani"`
}

// CiktiDetail, Admin detay sayfasında proje çıktılarını göstermek için veri yapısı.
type CiktiDetail struct {
	CiktiID       int    `json:"cikti_id"`
	CiktiTuru     string `json:"cikti_turu"`
	OngorulCikti  string `json:"ongorul_cikti"`
	ZamanAraligi  string `json:"zaman_araligi"`
}

// YayinDetail, Admin detay sayfasında yayınlaştırma bilgilerini göstermek için veri yapısı.
type YayinDetail struct {
	YayinTuru          string `json:"yayin_turu"`
	YayinCiktisi       string `json:"yayin_ciktisi"`
	TahminiYayinTarihi string `json:"tahmini_yayin_tarihi"`
}

// RevizyonDetail, Admin detay sayfasında revizyon geçmişini göstermek için veri yapısı.
type RevizyonDetail struct {
	RevizyonID      int    `json:"revizyon_id"`
	Aciklama        string `json:"aciklama"`
	Durum           string `json:"durum"`
	OlusturanAdSoyad string `json:"olusturan_ad_soyad"`
	OlusturmaTarihi string `json:"olusturma_tarihi"`
}

// ProjectDetail, Admin'in göreceği proje detay haritası (tüm proje içeriğini barındırır)
type ProjectDetail struct {
	Proje                       models.Proje                          `json:"proje"`
	ProjeDetay                  *models.ProjeDetay                    `json:"proje_detay"`
	Butceler                    []models.Butce                        `json:"butceler"`
	Reviews                     []ReviewDetail                        `json:"reviews"`
	YurutucuAd                  string                                `json:"yurutucu_ad"`
	TakimUyeleri                []TakimUyeDetail                      `json:"takim_uyeleri"`
	IsPaketleri                 []IsPaketiDetail                      `json:"is_paketleri"`
	Riskler                     []RiskDetail                          `json:"riskler"`
	ArastirmaBilgi              string                                `json:"arastirma_bilgi"`
	Ciktilar                    []CiktiDetail                         `json:"ciktilar"`
	Yayinlar                    []YayinDetail                         `json:"yayinlar"`
	Revizyonlar                 []RevizyonDetail                      `json:"revizyonlar"`
	YayinEtki                   []models.ProjeYayinEtki               `json:"yayin_etki"`
	YayginlastirmaEtkinlikleri  []models.ProjeYayginlastirmaEtkinlik  `json:"yayginlastirma_etkinlikleri"`
}

// GetProjectDetailsForAdmin, bir projenin tüm içeriğini admin veya yetkili kullanıcı için detaylı şekilde döner.
// Türkçe Bilgilendirme: Admin veya TTO yetkilisi ise hakemlerin gerçek ad-soyad bilgilerini döner, aksi halde "Hakem" olarak maskeler.
// Türkçe Yorum: Proje temel bilgisi, yürütücü ve akademik detay tek sorguda çekilerek DB round-trip sayısı azaltılmıştır.
func (r *AdminRepository) GetProjectDetailsForAdmin(projeID int, isAdminOrTTO bool) (*ProjectDetail, error) {
	detail := &ProjectDetail{
		ProjeDetay: &models.ProjeDetay{},
	}

	// 1. Proje Temel Bilgisi + Yürütücü + Akademik Detay — tek sorguda
	// Türkçe Yorum: Önceden 3 ayrı DB çağrısı gerektiren bilgiler tek bir LEFT JOIN sorgusu ile alınıyor.
	errTemel := r.DB.QueryRow(`
		SELECT
		    p.proje_id,
		    COALESCE(p.proje_kodu, ''),
		    COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'), ''),
		    COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'en'), ''),
		    COALESCE(pbt.bap_turu, 'Münferit'),
		    COALESCE(pd.durum_adi, 'taslak'),
		    COALESCE((SELECT COALESCE(SUM(toplam_fiyat), 0) FROM proje_butce WHERE proje_id = p.proje_id), 0),
		    p.olusturma_tarihi,
		    COALESCE(p.sure_ay, 0),
		    COALESCE(EXISTS(SELECT 1 FROM proje_etik_kurul WHERE proje_id = p.proje_id), false),
		    COALESCE((
		        SELECT u.ad || ' ' || u.soyad
		        FROM proje_takim pt2
		        INNER JOIN uye u ON u.uye_id = pt2.uye_id
		        INNER JOIN proje_rol_tanimlama prt ON pt2.proje_rol_id = prt.rol_id
		        WHERE pt2.proje_id = p.proje_id AND prt.proje_rol = 'Yürütücü'
		        LIMIT 1
		    ), 'Bilinmiyor') AS yurutucu_ad,
		    COALESCE(pdet.ozet, ''),
		    COALESCE(pdet.anahtar_kelimeler, ''),
		    COALESCE(pdet.hedefler, ''),
		    COALESCE(pdet.ozgunluk, ''),
		    COALESCE(pdet.metodoloji, ''),
		    COALESCE(pdet.kaynakca, '')
		FROM proje p
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN proje_detay pdet ON p.proje_id = pdet.proje_id
		WHERE p.proje_id = $1
	`, projeID).Scan(
		&detail.Proje.ProjeID, &detail.Proje.ProjeKodu, &detail.Proje.BaslikTr, &detail.Proje.BaslikEn,
		&detail.Proje.BapTuru, &detail.Proje.DurumAdi, &detail.Proje.ToplamButce,
		&detail.Proje.OlusturmaTarihi, &detail.Proje.SureAy, &detail.Proje.EtikKurul,
		&detail.YurutucuAd,
		&detail.ProjeDetay.Ozet, &detail.ProjeDetay.AnahtarKelimeler,
		&detail.ProjeDetay.Hedefler, &detail.ProjeDetay.Ozgunluk,
		&detail.ProjeDetay.Metodoloji, &detail.ProjeDetay.Kaynakca,
	)
	if errTemel != nil {
		return nil, errTemel
	}
	detail.ProjeDetay.ProjeID = projeID

	// 2. Takım Üyeleri
	rowsTakim, err := r.DB.Query(`
		SELECT u.uye_id, COALESCE(u.ad || ' ' || u.soyad, 'Bilinmiyor'),
		       COALESCE(d.rol, 'belirsiz'), COALESCE(prt.proje_rol, 'Araştırmacı')
		FROM proje_takim pt
		INNER JOIN uye u ON pt.uye_id = u.uye_id
		LEFT JOIN uye_detay d ON u.uye_id = d.uye_id
		LEFT JOIN proje_rol_tanimlama prt ON pt.proje_rol_id = prt.rol_id
		WHERE pt.proje_id = $1
	`, projeID)
	if err != nil {
		log.Printf("GetProjectDetailsForAdmin takım sorgusu hatası (proje_id=%d): %v", projeID, err)
	} else {
		defer rowsTakim.Close()
		for rowsTakim.Next() {
			var t TakimUyeDetail
			if scanErr := rowsTakim.Scan(&t.UyeID, &t.AdSoyad, &t.Rol, &t.ProjeRol); scanErr != nil {
				log.Printf("GetProjectDetailsForAdmin takım satırı okunamadı (proje_id=%d): %v", projeID, scanErr)
				continue
			}
			detail.TakimUyeleri = append(detail.TakimUyeleri, t)
		}
	}

	// 3. Bütçe Bilgileri
	rowsButce, err := r.DB.Query(`
		SELECT kalem_id, COALESCE(bk.kategori_adi, ''), aciklama, COALESCE(birim_ozelligi, 0), birim_fiyat, toplam_fiyat
		FROM proje_butce b
		LEFT JOIN proje_butce_kategori bk ON b.kategori_id = bk.kategori_id
		WHERE b.proje_id = $1
	`, projeID)
	if err != nil {
		log.Printf("GetProjectDetailsForAdmin bütçe sorgusu hatası (proje_id=%d): %v", projeID, err)
	} else {
		defer rowsButce.Close()
		for rowsButce.Next() {
			var b models.Butce
			if scanErr := rowsButce.Scan(&b.KalemID, &b.KategoriAdi, &b.Aciklama, &b.BirimOzelligi, &b.BirimFiyat, &b.ToplamFiyat); scanErr != nil {
				log.Printf("GetProjectDetailsForAdmin bütçe satırı okunamadı (proje_id=%d): %v", projeID, scanErr)
				continue
			}
			detail.Butceler = append(detail.Butceler, b)
		}
	}

	// 4. Hakem Değerlendirmeleri (koşullu maskeleme)
	var queryReviews string
	if isAdminOrTTO {
		queryReviews = `
			SELECT d.degerlendirme_id, COALESCE(u.ad || ' ' || u.soyad, 'Silinmiş Kullanıcı'),
			       COALESCE(d.puan, 0), COALESCE(d.yorum, ''), d.durum
			FROM proje_degerlendirmeleri d
			JOIN uye u ON u.uye_id = d.hakem_id
			WHERE d.proje_id = $1
		`
	} else {
		queryReviews = `
			SELECT d.degerlendirme_id, 'Hakem',
			       COALESCE(d.puan, 0), COALESCE(d.yorum, ''), d.durum
			FROM proje_degerlendirmeleri d
			JOIN uye u ON u.uye_id = d.hakem_id
			WHERE d.proje_id = $1
		`
	}
	rowsR, err := r.DB.Query(queryReviews, projeID)
	if err != nil {
		log.Printf("GetProjectDetailsForAdmin hakem sorgusu hatası (proje_id=%d): %v", projeID, err)
	} else {
		defer rowsR.Close()
		for rowsR.Next() {
			var rd ReviewDetail
			if scanErr := rowsR.Scan(&rd.DegerlendirmeID, &rd.HakemAdSoyad, &rd.Puan, &rd.Yorum, &rd.Durum); scanErr != nil {
				log.Printf("GetProjectDetailsForAdmin hakem satırı okunamadı (proje_id=%d): %v", projeID, scanErr)
				continue
			}
			detail.Reviews = append(detail.Reviews, rd)
		}
	}

	// 5. İş Paketleri
	rowsPaket, err := r.DB.Query(`
		SELECT paket_id, COALESCE(paket_adi, ''), COALESCE(paket_amaci, ''),
		       COALESCE(baslangic_ay, 1), COALESCE(bitis_ay, 1)
		FROM proje_is_paketi WHERE proje_id = $1 ORDER BY paket_id
	`, projeID)
	if err != nil {
		log.Printf("GetProjectDetailsForAdmin iş paketleri sorgusu hatası (proje_id=%d): %v", projeID, err)
	} else {
		defer rowsPaket.Close()
		for rowsPaket.Next() {
			var ip IsPaketiDetail
			if scanErr := rowsPaket.Scan(&ip.PaketID, &ip.PaketAdi, &ip.PaketAmaci, &ip.BaslangicAy, &ip.BitisAy); scanErr != nil {
				log.Printf("GetProjectDetailsForAdmin iş paketi satırı okunamadı (proje_id=%d): %v", projeID, scanErr)
				continue
			}
			detail.IsPaketleri = append(detail.IsPaketleri, ip)
		}
	}

	// 6. Risk Yönetimi
	rowsRisk, err := r.DB.Query(`
		SELECT risk_id, COALESCE(risk_aciklamasi, ''), COALESCE(cozum_plani, '')
		FROM proje_risk_yonetimi WHERE proje_id = $1
	`, projeID)
	if err != nil {
		log.Printf("GetProjectDetailsForAdmin risk sorgusu hatası (proje_id=%d): %v", projeID, err)
	} else {
		defer rowsRisk.Close()
		for rowsRisk.Next() {
			var rk RiskDetail
			if scanErr := rowsRisk.Scan(&rk.RiskID, &rk.RiskAciklamasi, &rk.CozumPlani); scanErr != nil {
				log.Printf("GetProjectDetailsForAdmin risk satırı okunamadı (proje_id=%d): %v", projeID, scanErr)
				continue
			}
			detail.Riskler = append(detail.Riskler, rk)
		}
	}

	// 7. Araştırma Bilgisi
	if scanErr := r.DB.QueryRow(`
		SELECT COALESCE(arastirma_amaci, '') FROM proje_arastirma WHERE proje_id = $1
	`, projeID).Scan(&detail.ArastirmaBilgi); scanErr != nil && scanErr.Error() != "sql: no rows in result set" {
		log.Printf("GetProjectDetailsForAdmin araştırma sorgusu hatası (proje_id=%d): %v", projeID, scanErr)
	}

	// 8. Yaygın Etki Çıktıları
	rowsYE, err := r.DB.Query(`
		SELECT id, proje_id, cikti_turu, COALESCE(ongorul_cikti,''), COALESCE(zaman_araligi,'')
		FROM proje_yayin_etki WHERE proje_id = $1 ORDER BY id ASC
	`, projeID)
	if err != nil {
		log.Printf("GetProjectDetailsForAdmin yayin_etki sorgusu hatası (proje_id=%d): %v", projeID, err)
	} else {
		defer rowsYE.Close()
		for rowsYE.Next() {
			var ye models.ProjeYayinEtki
			if scanErr := rowsYE.Scan(&ye.ID, &ye.ProjeID, &ye.CiktiTuru, &ye.OngorulCikti, &ye.ZamanAraligi); scanErr != nil {
				log.Printf("GetProjectDetailsForAdmin yayin_etki satırı okunamadı (proje_id=%d): %v", projeID, scanErr)
				continue
			}
			detail.YayinEtki = append(detail.YayinEtki, ye)
			// Türkçe Yorum: Çıktı bilgisi hem YayinEtki hem de Ciktilar listesine ekleniyor (eski uyumluluk)
			detail.Ciktilar = append(detail.Ciktilar, CiktiDetail{
				CiktiID:      ye.ID,
				CiktiTuru:    ye.CiktiTuru,
				OngorulCikti: ye.OngorulCikti,
				ZamanAraligi: ye.ZamanAraligi,
			})
		}
	}

	// 9. Yaygınlaştırma Etkinlikleri
	rowsEtk, err := r.DB.Query(`
		SELECT id, proje_id, COALESCE(etkinlik_turu,''), COALESCE(paydas,''), COALESCE(zaman_sure,''), sira_no
		FROM proje_yayginlastirma_etkinlik WHERE proje_id = $1 ORDER BY sira_no ASC
	`, projeID)
	if err != nil {
		log.Printf("GetProjectDetailsForAdmin etkinlik sorgusu hatası (proje_id=%d): %v", projeID, err)
	} else {
		defer rowsEtk.Close()
		for rowsEtk.Next() {
			var e models.ProjeYayginlastirmaEtkinlik
			if scanErr := rowsEtk.Scan(&e.ID, &e.ProjeID, &e.EtkinlikTuru, &e.Paydas, &e.ZamanSure, &e.SiraNo); scanErr != nil {
				log.Printf("GetProjectDetailsForAdmin etkinlik satırı okunamadı (proje_id=%d): %v", projeID, scanErr)
				continue
			}
			detail.YayginlastirmaEtkinlikleri = append(detail.YayginlastirmaEtkinlikleri, e)
		}
	}

	// 10. Yayınlaştırma Bilgileri
	rowsYayin, err := r.DB.Query(`
		SELECT COALESCE(yayin_turu, ''), COALESCE(yayin_ciktisi, ''),
		       COALESCE(tahmini_yayin_tarihi, '')
		FROM proje_yayinlastirma WHERE proje_id = $1
	`, projeID)
	if err != nil {
		log.Printf("GetProjectDetailsForAdmin yayinlastirma sorgusu hatası (proje_id=%d): %v", projeID, err)
	} else {
		defer rowsYayin.Close()
		for rowsYayin.Next() {
			var y YayinDetail
			if scanErr := rowsYayin.Scan(&y.YayinTuru, &y.YayinCiktisi, &y.TahminiYayinTarihi); scanErr != nil {
				log.Printf("GetProjectDetailsForAdmin yayin satırı okunamadı (proje_id=%d): %v", projeID, scanErr)
				continue
			}
			detail.Yayinlar = append(detail.Yayinlar, y)
		}
	}

	// 11. Revizyon Geçmişi
	rowsRev, err := r.DB.Query(`
		SELECT r.revizyon_id, COALESCE(r.aciklama, ''), COALESCE(r.durum, 'bekliyor'),
		       COALESCE(u.ad || ' ' || u.soyad, 'Bilinmiyor'),
		       TO_CHAR(r.olusturma_tarihi, 'DD.MM.YYYY')
		FROM proje_revizyon r
		LEFT JOIN uye u ON r.olusturan_kisi_id = u.uye_id
		WHERE r.proje_id = $1 ORDER BY r.olusturma_tarihi DESC
	`, projeID)
	if err != nil {
		log.Printf("GetProjectDetailsForAdmin revizyon sorgusu hatası (proje_id=%d): %v", projeID, err)
	} else {
		defer rowsRev.Close()
		for rowsRev.Next() {
			var rv RevizyonDetail
			if scanErr := rowsRev.Scan(&rv.RevizyonID, &rv.Aciklama, &rv.Durum, &rv.OlusturanAdSoyad, &rv.OlusturmaTarihi); scanErr != nil {
				log.Printf("GetProjectDetailsForAdmin revizyon satırı okunamadı (proje_id=%d): %v", projeID, scanErr)
				continue
			}
			detail.Revizyonlar = append(detail.Revizyonlar, rv)
		}
	}

	return detail, nil
}

// AtananHakemDetay, bir projeye atanan hakemin bilgilerini ve atama durumunu tutar.
type AtananHakemDetay struct {
	DegerlendirmeID int    `json:"degerlendirme_id"`
	HakemID         int    `json:"hakem_id"`
	HakemAdSoyad    string `json:"hakem_ad_soyad"`
	HakemBolum      string `json:"hakem_bolum"`
	AtamaDurumu     string `json:"atama_durumu"`
	Durum           string `json:"durum"`
	Puan            *int   `json:"puan"`
	RedNedeni       string `json:"red_nedeni"`
}

// AssignHakemToProje, admin veya TTO tarafından belirli bir hakemi projeye atar.
func (r *AdminRepository) AssignHakemToProje(projeID, hakemID int) error {
	// Revizyon sonrası yeniden atama yapıldığında eski değerlendirme kaydı sıfırlanır
	query := `
		INSERT INTO proje_degerlendirmeleri (proje_id, hakem_id, durum, atama_durumu, puan, yorum, red_nedeni)
		VALUES ($1, $2, 'Bekliyor', 'Atandı', NULL, NULL, NULL)
		ON CONFLICT (proje_id, hakem_id) DO UPDATE
		SET durum = 'Bekliyor',
		    atama_durumu = 'Atandı',
		    puan = NULL,
		    yorum = NULL,
		    red_nedeni = NULL,
		    olusturma_tarihi = CURRENT_TIMESTAMP;
	`
	_, err := r.DB.Exec(query, projeID, hakemID)
	if err != nil {
		log.Printf("AssignHakemToProje hatası: %v", err)
		return err
	}
	return nil
}

// GetDegerlendirilmemisProjeleri, hakem ataması gereken projeleri getirir.
// Durumu 'incelemede' veya 'komisyon_bekliyor' olan tüm projeler listelenir.
func (r *AdminRepository) GetDegerlendirilmemisProjeleri() ([]models.Proje, error) {
	query := `
		SELECT p.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'), ''), COALESCE(pd.durum_adi, 'taslak'),
		       COALESCE(pbt.bap_turu, 'Münferit'), p.olusturma_tarihi
		FROM proje p
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		WHERE pd.durum_adi IN ('incelemede', 'komisyon_bekliyor')
		ORDER BY p.olusturma_tarihi DESC
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		log.Printf("GetDegerlendirilmemisProjeleri hatası: %v", err)
		return nil, err
	}
	defer rows.Close()

	var projeler []models.Proje
	for rows.Next() {
		var p models.Proje
		if err := rows.Scan(&p.ProjeID, &p.ProjeKodu, &p.BaslikTr, &p.DurumAdi, &p.BapTuru, &p.OlusturmaTarihi); err == nil {
			projeler = append(projeler, p)
		}
	}
	return projeler, nil
}

// GetHakemListesi, sistemdeki aktif hakem kullanıcılarını getirir.
func (r *AdminRepository) GetHakemListesi() ([]models.Uye, error) {
	// Çoklu rol desteği için hem doğrudan rol alanına hem de sistem_rol tablosuna bakılır
	query := `
		SELECT DISTINCT u.uye_id, u.ad, u.soyad, u.eposta, COALESCE(d.bolum, ''), COALESCE(d.unvan, '')
		FROM uye u
		LEFT JOIN uye_detay d ON u.uye_id = d.uye_id
		LEFT JOIN sistem_rol sr ON u.uye_id = sr.uye_id
		LEFT JOIN sistem_rol_tanimlama srt ON sr.sistem_rol_id = srt.rol_id
		WHERE (u.rol = 'hakem' OR srt.rol_adi = 'hakem') AND u.aktif_mi = true
		ORDER BY u.ad, u.soyad
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		log.Printf("GetHakemListesi hatası: %v", err)
		return nil, err
	}
	defer rows.Close()

	var hakemler []models.Uye
	for rows.Next() {
		var h models.Uye
		if err := rows.Scan(&h.UyeID, &h.Ad, &h.Soyad, &h.Eposta, &h.Bolum, &h.Unvan); err == nil {
			hakemler = append(hakemler, h)
		}
	}
	return hakemler, nil
}

// GetProjeyeAtananHakemler, bir projeye atanan hakemlerin listesini atama durumlarıyla birlikte döner.
func (r *AdminRepository) GetProjeyeAtananHakemler(projeID int) ([]AtananHakemDetay, error) {
	query := `
		SELECT deg.degerlendirme_id, deg.hakem_id,
		       COALESCE(u.ad || ' ' || u.soyad, 'Bilinmiyor'),
		       COALESCE(d.bolum, ''),
		       COALESCE(deg.atama_durumu, 'Kabul Edildi'),
		       deg.durum, deg.puan, COALESCE(deg.red_nedeni, '')
		FROM proje_degerlendirmeleri deg
		JOIN uye u ON u.uye_id = deg.hakem_id
		LEFT JOIN uye_detay d ON u.uye_id = d.uye_id
		WHERE deg.proje_id = $1
		ORDER BY deg.olusturma_tarihi DESC
	`
	rows, err := r.DB.Query(query, projeID)
	if err != nil {
		log.Printf("GetProjeyeAtananHakemler hatası: %v", err)
		return nil, err
	}
	defer rows.Close()

	var hakemler []AtananHakemDetay
	for rows.Next() {
		var h AtananHakemDetay
		if err := rows.Scan(&h.DegerlendirmeID, &h.HakemID, &h.HakemAdSoyad, &h.HakemBolum, &h.AtamaDurumu, &h.Durum, &h.Puan, &h.RedNedeni); err == nil {
			hakemler = append(hakemler, h)
		}
	}
	return hakemler, nil
}

// GetBapTurleri, sistemdeki BAP proje türlerini listeler.
// onlyActive true ise sadece aktif olanlar getirilir.
func (r *AdminRepository) GetBapTurleri(onlyActive bool) ([]models.ProjeBapTuru, error) {
	var list []models.ProjeBapTuru
	query := `
		SELECT bap_turu_id, bap_turu, COALESCE(butce_limiti, 0), COALESCE(sure_limiti_ay, 0), aktif_mi, COALESCE(aciklama, ''),
		       COALESCE(hakem_gerekli, false), COALESCE(hakem_sayisi, 0),
		       COALESCE(bursiyer_gerekli, false), COALESCE(bursiyer_sayisi, 0),
		       COALESCE(ara_rapor_gerekli, false), COALESCE(ara_rapor_sayisi, 0)
		FROM proje_bap_turu
	`
	if onlyActive {
		query += " WHERE aktif_mi = true"
	}
	query += " ORDER BY bap_turu_id"

	rows, err := r.DB.Query(query)
	if err != nil {
		log.Printf("GetBapTurleri hatası: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bt models.ProjeBapTuru
		err := rows.Scan(&bt.BapTuruID, &bt.BapTuru, &bt.ButceLimiti, &bt.SureLimitiAy, &bt.AktifMi, &bt.Aciklama,
			&bt.HakemGerekli, &bt.HakemSayisi, &bt.BursiyerGerekli, &bt.BursiyerSayisi,
			&bt.AraRaporGerekli, &bt.AraRaporSayisi)
		if err == nil {
			list = append(list, bt)
		}
	}
	return list, nil
}

// CreateBapTuru, yeni bir BAP proje türü tanımlar.
func (r *AdminRepository) CreateBapTuru(bt *models.ProjeBapTuru) error {
	query := `
		INSERT INTO proje_bap_turu (bap_turu, butce_limiti, sure_limiti_ay, aktif_mi, aciklama,
		                           hakem_gerekli, hakem_sayisi, bursiyer_gerekli, bursiyer_sayisi,
		                           ara_rapor_gerekli, ara_rapor_sayisi)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING bap_turu_id
	`
	err := r.DB.QueryRow(query, bt.BapTuru, bt.ButceLimiti, bt.SureLimitiAy, bt.AktifMi, bt.Aciklama,
		bt.HakemGerekli, bt.HakemSayisi, bt.BursiyerGerekli, bt.BursiyerSayisi,
		bt.AraRaporGerekli, bt.AraRaporSayisi).Scan(&bt.BapTuruID)
	if err != nil {
		log.Printf("CreateBapTuru hatası: %v", err)
	}
	return err
}

// UpdateBapTuru, mevcut bir BAP proje türünü günceller.
func (r *AdminRepository) UpdateBapTuru(bt *models.ProjeBapTuru) error {
	query := `
		UPDATE proje_bap_turu
		SET bap_turu = $1, butce_limiti = $2, sure_limiti_ay = $3, aktif_mi = $4, aciklama = $5,
		    hakem_gerekli = $6, hakem_sayisi = $7, bursiyer_gerekli = $8, bursiyer_sayisi = $9,
		    ara_rapor_gerekli = $10, ara_rapor_sayisi = $11
		WHERE bap_turu_id = $12
	`
	_, err := r.DB.Exec(query, bt.BapTuru, bt.ButceLimiti, bt.SureLimitiAy, bt.AktifMi, bt.Aciklama,
		bt.HakemGerekli, bt.HakemSayisi, bt.BursiyerGerekli, bt.BursiyerSayisi,
		bt.AraRaporGerekli, bt.AraRaporSayisi, bt.BapTuruID)
	if err != nil {
		log.Printf("UpdateBapTuru hatası: %v", err)
	}
	return err
}

// CreateUser, admin tarafından yeni bir kullanıcı ve detaylarını ekler (transaction ile).
// Türkçe Yorum: Rol senkronizasyonu syncRolTx fonksiyonu üzerinden yapılır; tüm tablolar tek transaction'da güncellenir.
func (r *AdminRepository) CreateUser(req *models.AdminCreateUserRequest, hashedPass string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		log.Printf("CreateUser transaction başlatma hatası: %v", err)
		return err
	}
	defer tx.Rollback()

	roles := strings.Split(req.Rol, ",")
	ilkRol := ""
	if len(roles) > 0 {
		ilkRol = strings.TrimSpace(roles[0])
	}

	// 1. Uye tablosuna temel verileri ekle
	var uyeID int
	err = tx.QueryRow(`
		INSERT INTO uye (ad, soyad, eposta, sifre_hash, rol, aktif_mi, sifre_degistir_zorla)
		VALUES ($1, $2, $3, $4, $5, true, $6)
		RETURNING uye_id
	`, req.Ad, req.Soyad, req.Eposta, hashedPass, ilkRol, req.SifreDegistirZorla).Scan(&uyeID)
	if err != nil {
		log.Printf("CreateUser uye tablosu hatası: %v", err)
		return err
	}

	// 2. UyeDetay tablosuna detayları ekle (profil_tamamlandi = true olarak işaretlenir)
	_, err = tx.Exec(`
		INSERT INTO uye_detay (uye_id, rol, unvan, bolum, telefon, izu_uyesi, profil_tamamlandi)
		VALUES ($1, $2, $3, $4, $5, $6, true)
	`, uyeID, ilkRol, req.Unvan, req.Bolum, req.Telefon, req.IzuUyesi)
	if err != nil {
		log.Printf("CreateUser uye_detay tablosu hatası: %v", err)
		return err
	}

	// 3. Tüm rolleri sistem_rol tablosuna syncRolTx ile aktar
	if err := r.syncRolTx(tx, uyeID, roles); err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateUser, admin tarafından bir kullanıcının temel ve detay bilgilerini günceller (transaction ile).
// Türkçe Yorum: Rol senkronizasyonu syncRolTx fonksiyonu üzerinden yapılır; tüm tablolar tek transaction'da güncellenir.
func (r *AdminRepository) UpdateUser(uyeID int, req *models.AdminUpdateUserRequest) error {
	tx, err := r.DB.Begin()
	if err != nil {
		log.Printf("UpdateUser transaction başlatma hatası: %v", err)
		return err
	}
	defer tx.Rollback()

	roles := strings.Split(req.Rol, ",")
	ilkRol := ""
	if len(roles) > 0 {
		ilkRol = strings.TrimSpace(roles[0])
	}

	// 1. Uye tablosundaki temel verileri güncelle
	_, err = tx.Exec(`
		UPDATE uye
		SET ad = $1, soyad = $2, eposta = $3, guncelleme_tarihi = CURRENT_TIMESTAMP
		WHERE uye_id = $4
	`, req.Ad, req.Soyad, req.Eposta, uyeID)
	if err != nil {
		log.Printf("UpdateUser uye tablosu güncelleme hatası: %v", err)
		return err
	}

	// 2. UyeDetay tablosundaki verileri güncelle
	_, err = tx.Exec(`
		UPDATE uye_detay
		SET unvan = $1, bolum = $2, telefon = $3, izu_uyesi = $4, guncelleme_tarihi = CURRENT_TIMESTAMP
		WHERE uye_id = $5
	`, req.Unvan, req.Bolum, req.Telefon, req.IzuUyesi, uyeID)
	if err != nil {
		log.Printf("UpdateUser uye_detay tablosu güncelleme hatası: %v", err)
		return err
	}

	// 3. Rol senkronizasyonu: uye.rol, uye_detay.rol ve sistem_rol syncRolTx ile güncellenir
	if err := r.syncRolTx(tx, uyeID, roles); err != nil {
		return err
	}

	// 4. İlk rol uye_detay.rol legacy alanına yazılır (geriye dönük uyumluluk)
	_, err = tx.Exec(`UPDATE uye_detay SET rol = $1 WHERE uye_id = $2`, ilkRol, uyeID)
	if err != nil {
		log.Printf("UpdateUser uye_detay.rol senkronizasyon hatası: %v", err)
		return err
	}

	for _, rName := range roles {
		rName = strings.TrimSpace(rName)
		if rName == "" {
			continue
		}
		querySistemRol := `
			INSERT INTO sistem_rol (uye_id, sistem_rol_id)
			SELECT $1, rol_id FROM sistem_rol_tanimlama WHERE rol_adi = $2
			ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING
		`
		_, err = tx.Exec(querySistemRol, uyeID, rName)
		if err != nil {
			log.Printf("UpdateUser sistem_rol tablosu atama hatası: %v", err)
			return err
		}
	}

	// Tüm işlemler başarılı ise transaction commit edilir
	return tx.Commit()
}

// GetSayfaYetkiMatrix sistemdeki tüm rolleri, yetkilendirilebilir sayfaları ve aktif yetki eşleşmelerini döner.
func (r *AdminRepository) GetSayfaYetkiMatrix() (*models.SayfaYetkiMatrix, error) {
	// 1. Rolleri çek
	rowsRoles, err := r.DB.Query("SELECT rol_id, rol_adi FROM sistem_rol_tanimlama ORDER BY rol_id")
	if err != nil {
		return nil, err
	}
	defer rowsRoles.Close()

	var roles []models.SistemRolTanimlama
	for rowsRoles.Next() {
		var role models.SistemRolTanimlama
		if err := rowsRoles.Scan(&role.RolID, &role.RolAdi); err == nil {
			roles = append(roles, role)
		}
	}

	// 2. Sayfaları çek
	rowsPages, err := r.DB.Query("SELECT sayfa_id, sayfa_adi, sayfa_kodu, url_yolu FROM sistem_sayfa ORDER BY sayfa_adi ASC")
	if err != nil {
		return nil, err
	}
	defer rowsPages.Close()

	var pages []models.SistemSayfa
	for rowsPages.Next() {
		var page models.SistemSayfa
		if err := rowsPages.Scan(&page.SayfaID, &page.SayfaAdi, &page.SayfaKodu, &page.UrlYolu); err == nil {
			pages = append(pages, page)
		}
	}

	// 3. Yetkileri çek
	rowsPerms, err := r.DB.Query("SELECT yetki_id, sistem_rol_id, sayfa_id FROM sayfa_rol_yetki")
	if err != nil {
		return nil, err
	}
	defer rowsPerms.Close()

	var permissions []models.SayfaRolYetki
	for rowsPerms.Next() {
		var perm models.SayfaRolYetki
		if err := rowsPerms.Scan(&perm.YetkiID, &perm.SistemRolID, &perm.SayfaID); err == nil {
			permissions = append(permissions, perm)
		}
	}

	return &models.SayfaYetkiMatrix{
		Roles:       roles,
		Pages:       pages,
		Permissions: permissions,
	}, nil
}

// UpdateSayfaYetki adminin gönderdiği sayfa rol yetki değişikliklerini transaction ile veritabanına yansıtır.
func (r *AdminRepository) UpdateSayfaYetki(permissions []models.UpdateSayfaYetkiItem) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, perm := range permissions {
		if perm.Allowed {
			// Yetki ver
			_, err = tx.Exec(`
				INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
				VALUES ($1, $2)
				ON CONFLICT (sistem_rol_id, sayfa_id) DO NOTHING
			`, perm.SistemRolID, perm.SayfaID)
			if err != nil {
				return err
			}
		} else {
			// Yetki kaldır
			_, err = tx.Exec(`
				DELETE FROM sayfa_rol_yetki
				WHERE sistem_rol_id = $1 AND sayfa_id = $2
			`, perm.SistemRolID, perm.SayfaID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// CheckPageAccess belirtilen rollerden herhangi birinin istenen path (URL) değerine erişim yetkisi olup olmadığını kontrol eder.
func (r *AdminRepository) CheckPageAccess(roles []string, path string) (bool, error) {
	if len(roles) == 0 {
		return false, nil
	}

	// Admin rolü her zaman tüm sayfalara erişebilir (bypass kontrolü)
	for _, role := range roles {
		if strings.TrimSpace(role) == "admin" {
			return true, nil
		}
	}

	// Parametrelere göre yetki kontrolü yap
	// SQL IN parametreleri dinamik oluşturulur
	placeholders := make([]string, len(roles))
	args := make([]interface{}, len(roles)+1)
	args[0] = path

	for i, role := range roles {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = strings.TrimSpace(role)
	}

	query := fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1 FROM sayfa_rol_yetki sry
			INNER JOIN sistem_rol_tanimlama srt ON sry.sistem_rol_id = srt.rol_id
			INNER JOIN sistem_sayfa ss ON sry.sayfa_id = ss.sayfa_id
			WHERE ss.url_yolu = $1 AND srt.rol_adi IN (%s)
		)
	`, strings.Join(placeholders, ","))

	var exists bool
	err := r.DB.QueryRow(query, args...).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// CreateRole yeni bir sistem rolü oluşturur ve bu role sayfa yetkileri atar.
// Türkçe Yorum: Admin yetki yönetiminden yeni rol eklemek için transaction yapısıyla çalışır.
func (r *AdminRepository) CreateRole(rolAdi string, sayfaIDs []int) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var rolID int
	err = tx.QueryRow(`
		INSERT INTO sistem_rol_tanimlama (rol_adi)
		VALUES ($1)
		RETURNING rol_id
	`, rolAdi).Scan(&rolID)
	if err != nil {
		return err
	}

	for _, sayfaID := range sayfaIDs {
		_, err = tx.Exec(`
			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			VALUES ($1, $2)
		`, rolID, sayfaID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateRole mevcut bir sistem rolünün adını ve sayfa yetkilerini günceller.
// Türkçe Yorum: Rol ismini günceller, eski yetkileri temizler ve yeni yetkileri atar.
func (r *AdminRepository) UpdateRole(rolID int, rolAdi string, sayfaIDs []int) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE sistem_rol_tanimlama
		SET rol_adi = $1
		WHERE rol_id = $2
	`, rolAdi, rolID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		DELETE FROM sayfa_rol_yetki
		WHERE sistem_rol_id = $1
	`, rolID)
	if err != nil {
		return err
	}

	for _, sayfaID := range sayfaIDs {
		_, err = tx.Exec(`
			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			VALUES ($1, $2)
		`, rolID, sayfaID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// DeleteRole belirtilen sistem rolünü veritabanından siler.
// Türkçe Yorum: Rolü siler, yabancı anahtar kısıtlamaları (ON DELETE CASCADE) sayesinde yetkileri ve kullanıcı atamaları da silinir.
func (r *AdminRepository) DeleteRole(rolID int) error {
	_, err := r.DB.Exec(`
		DELETE FROM sistem_rol_tanimlama
		WHERE rol_id = $1
	`, rolID)
	return err
}

// GetAllowedPagesForRoles kullanıcının sahip olduğu rollere göre erişebileceği sayfaların url_yolu değerlerini döner.
// Türkçe Yorum: Verilen rollerin erişim izni olan sistem sayfalarının URL yollarını liste olarak döner.
func (r *AdminRepository) GetAllowedPagesForRoles(roles []string) ([]string, error) {
	if len(roles) == 0 {
		return []string{}, nil
	}

	// Admin rolü her zaman tüm sayfalara erişebilir (bypass kontrolü)
	isAdmin := false
	for _, role := range roles {
		if strings.TrimSpace(role) == "admin" {
			isAdmin = true
			break
		}
	}

	if isAdmin {
		rows, err := r.DB.Query("SELECT url_yolu FROM sistem_sayfa")
		if err != nil {
			log.Printf("GetAllowedPagesForRoles (admin) hatası: %v", err)
			return nil, err
		}
		defer rows.Close()

		var urls []string
		for rows.Next() {
			var url string
			if err := rows.Scan(&url); err == nil {
				urls = append(urls, url)
			}
		}
		return urls, nil
	}

	// Parametrelere göre yetki sorgulama
	placeholders := make([]string, len(roles))
	args := make([]interface{}, len(roles))
	for i, role := range roles {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = strings.TrimSpace(role)
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT ss.url_yolu
		FROM sayfa_rol_yetki sry
		INNER JOIN sistem_rol_tanimlama srt ON sry.sistem_rol_id = srt.rol_id
		INNER JOIN sistem_sayfa ss ON sry.sayfa_id = ss.sayfa_id
		WHERE srt.rol_adi IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		log.Printf("GetAllowedPagesForRoles hatası: %v", err)
		return nil, err
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err == nil {
			urls = append(urls, url)
		}
	}

	return urls, nil
}


