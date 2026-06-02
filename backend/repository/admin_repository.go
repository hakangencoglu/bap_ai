package repository

import (
	"database/sql"
	"log"

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
		SELECT u.uye_id, COALESCE(d.rol, ''), COALESCE(d.unvan, ''), u.ad, u.soyad,
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
		SELECT p.proje_id, p.baslik_tr, COALESCE(pd.durum_adi, 'taslak'),
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
		if err := rows.Scan(&p.ProjeID, &p.BaslikTr, &p.DurumAdi, &p.BapTuru, &p.OlusturmaTarihi); err == nil {
			projes = append(projes, p)
		}
	}
	return projes, nil
}

// UpdateUserRole, bir kullanıcının rolünü günceller.
// Hem uye, hem uye_detay, hem de sistem_rol tabloları güncellenir.
func (r *AdminRepository) UpdateUserRole(uyeID int, rolAdi string) error {
	// 1. Uye tablosundaki rol alanını güncelle (geriye dönük uyumluluk)
	_, err := r.DB.Exec(`UPDATE uye SET rol = $1 WHERE uye_id = $2`, rolAdi, uyeID)
	if err != nil {
		log.Printf("UpdateUserRole uye hatası: %v", err)
		return err
	}

	// 2. Uye_detay tablosundaki rol alanını da güncelle
	_, err = r.DB.Exec(`UPDATE uye_detay SET rol = $1, guncelleme_tarihi = CURRENT_TIMESTAMP WHERE uye_id = $2`, rolAdi, uyeID)
	if err != nil {
		log.Printf("UpdateUserRole uye_detay hatası: %v", err)
		return err
	}

	// 3. Sistem_rol tablosunu güncelle (mevcut rolleri temizle, yenisini ata)
	_, err = r.DB.Exec(`DELETE FROM sistem_rol WHERE uye_id = $1`, uyeID)
	if err != nil {
		log.Printf("UpdateUserRole sistem_rol temizleme hatası: %v", err)
		return err
	}

	_, err = r.DB.Exec(`
		INSERT INTO sistem_rol (uye_id, sistem_rol_id)
		SELECT $1, rol_id FROM sistem_rol_tanimlama WHERE rol_adi = $2
		ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING
	`, uyeID, rolAdi)
	if err != nil {
		log.Printf("UpdateUserRole sistem_rol atama hatası: %v", err)
	}
	return err
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
		WHERE pd.durum_adi = 'onaylandi'
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

	err = r.DB.QueryRow("SELECT COALESCE(SUM(toplam_butce), 0) FROM proje").Scan(&stats.ToplamButce)
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
	PaketID         int    `json:"paket_id"`
	PaketAdi        string `json:"paket_adi"`
	PaketAmaci      string `json:"paket_amaci"`
	BaslangicTarihi string `json:"baslangic_tarihi"`
	BitisTarihi     string `json:"bitis_tarihi"`
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
	Aciklama      string `json:"aciklama"`
	CiktiPeriyodu string `json:"cikti_periyodu"`
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
	Proje          models.Proje      `json:"proje"`
	ProjeDetay     *models.ProjeDetay `json:"proje_detay"`
	Butceler       []models.Butce    `json:"butceler"`
	Reviews        []ReviewDetail    `json:"reviews"`
	YurutucuAd     string            `json:"yurutucu_ad"`
	TakimUyeleri   []TakimUyeDetail  `json:"takim_uyeleri"`
	IsPaketleri    []IsPaketiDetail  `json:"is_paketleri"`
	Riskler        []RiskDetail      `json:"riskler"`
	ArastirmaBilgi string            `json:"arastirma_bilgi"`
	Ciktilar       []CiktiDetail     `json:"ciktilar"`
	Yayinlar       []YayinDetail     `json:"yayinlar"`
	Revizyonlar    []RevizyonDetail  `json:"revizyonlar"`
}

// GetProjectDetailsForAdmin, bir projenin tüm içeriğini admin için detaylı şekilde döner.
func (r *AdminRepository) GetProjectDetailsForAdmin(projeID int) (*ProjectDetail, error) {
	detail := &ProjectDetail{}

	// 1. Proje Temel Bilgisi
	err := r.DB.QueryRow(`
		SELECT p.proje_id, COALESCE(p.baslik_tr, ''), COALESCE(p.baslik_en, ''),
		       COALESCE(pbt.bap_turu, 'Münferit'),
		       COALESCE(pd.durum_adi, 'taslak'), COALESCE(p.toplam_butce, 0),
		       p.olusturma_tarihi, COALESCE(p.sure_ay, 0), COALESCE(p.etik_kurul, false)
		FROM proje p
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		WHERE p.proje_id = $1
	`, projeID).Scan(
		&detail.Proje.ProjeID, &detail.Proje.BaslikTr, &detail.Proje.BaslikEn,
		&detail.Proje.BapTuru, &detail.Proje.DurumAdi,
		&detail.Proje.ToplamButce, &detail.Proje.OlusturmaTarihi,
		&detail.Proje.SureAy, &detail.Proje.EtikKurul,
	)
	if err != nil {
		return nil, err
	}

	// 2. Yürütücü Bilgisi
	r.DB.QueryRow(`
		SELECT COALESCE(u.ad || ' ' || u.soyad, 'Bilinmiyor')
		FROM proje_takim pt
		INNER JOIN uye u ON u.uye_id = pt.uye_id
		INNER JOIN proje_rol_tanimlama prt ON pt.proje_rol_id = prt.rol_id
		WHERE pt.proje_id = $1 AND prt.proje_rol = 'Yürütücü'
		LIMIT 1
	`, projeID).Scan(&detail.YurutucuAd)

	// 3. Proje Akademik Detay (özet, hedefler, metodoloji vb.)
	detay := &models.ProjeDetay{}
	errDetay := r.DB.QueryRow(`
		SELECT proje_id, COALESCE(ozet, ''), COALESCE(anahtar_kelimeler, ''),
		       COALESCE(hedefler, ''), COALESCE(ozgunluk, ''), COALESCE(metodoloji, '')
		FROM proje_detay WHERE proje_id = $1
	`, projeID).Scan(
		&detay.ProjeID, &detay.Ozet, &detay.AnahtarKelimeler,
		&detay.Hedefler, &detay.Ozgunluk, &detay.Metodoloji,
	)
	if errDetay == nil {
		detail.ProjeDetay = detay
	}

	// 4. Takım Üyeleri
	rowsTakim, errTakim := r.DB.Query(`
		SELECT u.uye_id, COALESCE(u.ad || ' ' || u.soyad, 'Bilinmiyor'),
		       COALESCE(d.rol, 'belirsiz'), COALESCE(prt.proje_rol, 'Araştırmacı')
		FROM proje_takim pt
		INNER JOIN uye u ON pt.uye_id = u.uye_id
		LEFT JOIN uye_detay d ON u.uye_id = d.uye_id
		LEFT JOIN proje_rol_tanimlama prt ON pt.proje_rol_id = prt.rol_id
		WHERE pt.proje_id = $1
	`, projeID)
	if errTakim == nil {
		defer rowsTakim.Close()
		for rowsTakim.Next() {
			var t TakimUyeDetail
			if err := rowsTakim.Scan(&t.UyeID, &t.AdSoyad, &t.Rol, &t.ProjeRol); err == nil {
				detail.TakimUyeleri = append(detail.TakimUyeleri, t)
			}
		}
	}

	// 5. Bütçe Bilgileri
	var butceler []models.Butce
	rowsButce, errButce := r.DB.Query(`
		SELECT kalem_id, COALESCE(bk.kategori_adi, ''), aciklama, birim_fiyat, toplam_fiyat
		FROM butce b
		LEFT JOIN butce_kategori bk ON b.kategori_id = bk.kategori_id
		WHERE b.proje_id = $1
	`, projeID)
	if errButce == nil {
		defer rowsButce.Close()
		for rowsButce.Next() {
			var b models.Butce
			if err := rowsButce.Scan(&b.KalemID, &b.KategoriAdi, &b.Aciklama, &b.BirimFiyat, &b.ToplamFiyat); err == nil {
				butceler = append(butceler, b)
			}
		}
	}
	detail.Butceler = butceler

	// 6. Hakem Değerlendirmeleri
	var reviews []ReviewDetail
	rowsR, errR := r.DB.Query(`
		SELECT d.degerlendirme_id, COALESCE(u.ad || ' ' || u.soyad, 'Silinmiş Kullanıcı'),
		       COALESCE(d.puan, 0), COALESCE(d.yorum, ''), d.durum
		FROM proje_degerlendirmeleri d
		JOIN uye u ON u.uye_id = d.hakem_id
		WHERE d.proje_id = $1
	`, projeID)
	if errR == nil {
		defer rowsR.Close()
		for rowsR.Next() {
			var rd ReviewDetail
			if err := rowsR.Scan(&rd.DegerlendirmeID, &rd.HakemAdSoyad, &rd.Puan, &rd.Yorum, &rd.Durum); err == nil {
				reviews = append(reviews, rd)
			}
		}
	}
	detail.Reviews = reviews

	// 7. İş Paketleri
	rowsPaket, errPaket := r.DB.Query(`
		SELECT paket_id, COALESCE(paket_adi, ''), COALESCE(paket_amaci, ''),
		       COALESCE(TO_CHAR(baslangic_tarihi, 'DD.MM.YYYY'), '-'),
		       COALESCE(TO_CHAR(bitis_tarihi, 'DD.MM.YYYY'), '-')
		FROM is_paketi WHERE proje_id = $1 ORDER BY paket_id
	`, projeID)
	if errPaket == nil {
		defer rowsPaket.Close()
		for rowsPaket.Next() {
			var ip IsPaketiDetail
			if err := rowsPaket.Scan(&ip.PaketID, &ip.PaketAdi, &ip.PaketAmaci, &ip.BaslangicTarihi, &ip.BitisTarihi); err == nil {
				detail.IsPaketleri = append(detail.IsPaketleri, ip)
			}
		}
	}

	// 8. Risk Yönetimi
	rowsRisk, errRisk := r.DB.Query(`
		SELECT risk_id, COALESCE(risk_aciklamasi, ''), COALESCE(cozum_plani, '')
		FROM risk_yonetimi WHERE proje_id = $1
	`, projeID)
	if errRisk == nil {
		defer rowsRisk.Close()
		for rowsRisk.Next() {
			var rk RiskDetail
			if err := rowsRisk.Scan(&rk.RiskID, &rk.RiskAciklamasi, &rk.CozumPlani); err == nil {
				detail.Riskler = append(detail.Riskler, rk)
			}
		}
	}

	// 9. Araştırma Bilgileri
	r.DB.QueryRow(`
		SELECT COALESCE(arastirma_amaci, '') FROM arastirma WHERE proje_id = $1
	`, projeID).Scan(&detail.ArastirmaBilgi)

	// 10. Proje Çıktıları
	rowsCikti, errCikti := r.DB.Query(`
		SELECT c.cikti_id, COALESCE(ct.cikti_turu, 'Belirtilmemiş'),
		       COALESCE(c.aciklama, ''), COALESCE(c.cikti_periyodu, '')
		FROM proje_cikti c
		LEFT JOIN proje_cikti_turu ct ON c.cikti_turu_id = ct.cikti_turu_id
		WHERE c.proje_id = $1
	`, projeID)
	if errCikti == nil {
		defer rowsCikti.Close()
		for rowsCikti.Next() {
			var ck CiktiDetail
			if err := rowsCikti.Scan(&ck.CiktiID, &ck.CiktiTuru, &ck.Aciklama, &ck.CiktiPeriyodu); err == nil {
				detail.Ciktilar = append(detail.Ciktilar, ck)
			}
		}
	}

	// 11. Yayınlaştırma Bilgileri
	rowsYayin, errYayin := r.DB.Query(`
		SELECT COALESCE(yayin_turu, ''), COALESCE(yayin_ciktisi, ''),
		       COALESCE(tahmini_yayin_tarihi, '')
		FROM proje_yayinlastirma WHERE proje_id = $1
	`, projeID)
	if errYayin == nil {
		defer rowsYayin.Close()
		for rowsYayin.Next() {
			var y YayinDetail
			if err := rowsYayin.Scan(&y.YayinTuru, &y.YayinCiktisi, &y.TahminiYayinTarihi); err == nil {
				detail.Yayinlar = append(detail.Yayinlar, y)
			}
		}
	}

	// 12. Revizyon Geçmişi
	rowsRev, errRev := r.DB.Query(`
		SELECT r.revizyon_id, COALESCE(r.aciklama, ''), COALESCE(r.durum, 'bekliyor'),
		       COALESCE(u.ad || ' ' || u.soyad, 'Bilinmiyor'),
		       TO_CHAR(r.olusturma_tarihi, 'DD.MM.YYYY')
		FROM revizyonlar r
		LEFT JOIN uye u ON r.olusturan_kisi_id = u.uye_id
		WHERE r.proje_id = $1 ORDER BY r.olusturma_tarihi DESC
	`, projeID)
	if errRev == nil {
		defer rowsRev.Close()
		for rowsRev.Next() {
			var rv RevizyonDetail
			if err := rowsRev.Scan(&rv.RevizyonID, &rv.Aciklama, &rv.Durum, &rv.OlusturanAdSoyad, &rv.OlusturmaTarihi); err == nil {
				detail.Revizyonlar = append(detail.Revizyonlar, rv)
			}
		}
	}

	return detail, nil
}
