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

	// 3. Proje Akademik Detay (özet, hedefler, metodoloji, kaynakça vb.)
	detay := &models.ProjeDetay{}
	errDetay := r.DB.QueryRow(`
		SELECT proje_id, COALESCE(ozet, ''), COALESCE(anahtar_kelimeler, ''),
		       COALESCE(hedefler, ''), COALESCE(ozgunluk, ''), COALESCE(metodoloji, ''),
		       COALESCE(kaynakca, '')
		FROM proje_detay WHERE proje_id = $1
	`, projeID).Scan(
		&detay.ProjeID, &detay.Ozet, &detay.AnahtarKelimeler,
		&detay.Hedefler, &detay.Ozgunluk, &detay.Metodoloji, &detay.Kaynakca,
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

	// 12. Yaygın Etki Çıktıları
	rowsYE, errYE := r.DB.Query(`
		SELECT id, proje_id, cikti_turu, COALESCE(ongorul_cikti,''), COALESCE(zaman_araligi,'')
		FROM proje_yayin_etki WHERE proje_id = $1 ORDER BY id ASC
	`, projeID)
	if errYE == nil {
		defer rowsYE.Close()
		for rowsYE.Next() {
			var ye models.ProjeYayinEtki
			if err := rowsYE.Scan(&ye.ID, &ye.ProjeID, &ye.CiktiTuru, &ye.OngorulCikti, &ye.ZamanAraligi); err == nil {
				detail.YayinEtki = append(detail.YayinEtki, ye)
			}
		}
	}

	// 13. Yaygınlaştırma Etkinlikleri
	rowsEtk, errEtk := r.DB.Query(`
		SELECT id, proje_id, COALESCE(etkinlik_turu,''), COALESCE(paydas,''), COALESCE(zaman_sure,''), sira_no
		FROM proje_yayginlastirma_etkinlik WHERE proje_id = $1 ORDER BY sira_no ASC
	`, projeID)
	if errEtk == nil {
		defer rowsEtk.Close()
		for rowsEtk.Next() {
			var e models.ProjeYayginlastirmaEtkinlik
			if err := rowsEtk.Scan(&e.ID, &e.ProjeID, &e.EtkinlikTuru, &e.Paydas, &e.ZamanSure, &e.SiraNo); err == nil {
				detail.YayginlastirmaEtkinlikleri = append(detail.YayginlastirmaEtkinlikleri, e)
			}
		}
	}

	// 14. Revizyon Geçmişi
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

// AssignHakemToProje, admin tarafından belirli bir hakemi projeye atar.
func (r *AdminRepository) AssignHakemToProje(projeID, hakemID int) error {
	// Aynı hakem-proje çifti varsa çakışma önlenir
	query := `
		INSERT INTO proje_degerlendirmeleri (proje_id, hakem_id, durum, atama_durumu)
		VALUES ($1, $2, 'Bekliyor', 'Atandı')
		ON CONFLICT (proje_id, hakem_id) DO NOTHING
	`
	res, err := r.DB.Exec(query, projeID, hakemID)
	if err != nil {
		log.Printf("AssignHakemToProje hatası: %v", err)
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil // Zaten atanmış, hata fırlatmaya gerek yok
	}
	return nil
}

// GetDegerlendirilmemisProjeleri, hakem ataması gereken projeleri getirir.
// Durumu 'incelemede' veya 'komisyon_bekliyor' olan tüm projeler listelenir.
func (r *AdminRepository) GetDegerlendirilmemisProjeleri() ([]models.Proje, error) {
	query := `
		SELECT p.proje_id, COALESCE(p.baslik_tr, ''), COALESCE(pd.durum_adi, 'taslak'),
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
		if err := rows.Scan(&p.ProjeID, &p.BaslikTr, &p.DurumAdi, &p.BapTuru, &p.OlusturmaTarihi); err == nil {
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
		       COALESCE(bursiyer_gerekli, false), COALESCE(bursiyer_sayisi, 0)
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
			&bt.HakemGerekli, &bt.HakemSayisi, &bt.BursiyerGerekli, &bt.BursiyerSayisi)
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
		                           hakem_gerekli, hakem_sayisi, bursiyer_gerekli, bursiyer_sayisi)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING bap_turu_id
	`
	err := r.DB.QueryRow(query, bt.BapTuru, bt.ButceLimiti, bt.SureLimitiAy, bt.AktifMi, bt.Aciklama,
		bt.HakemGerekli, bt.HakemSayisi, bt.BursiyerGerekli, bt.BursiyerSayisi).Scan(&bt.BapTuruID)
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
		    hakem_gerekli = $6, hakem_sayisi = $7, bursiyer_gerekli = $8, bursiyer_sayisi = $9
		WHERE bap_turu_id = $10
	`
	_, err := r.DB.Exec(query, bt.BapTuru, bt.ButceLimiti, bt.SureLimitiAy, bt.AktifMi, bt.Aciklama,
		bt.HakemGerekli, bt.HakemSayisi, bt.BursiyerGerekli, bt.BursiyerSayisi, bt.BapTuruID)
	if err != nil {
		log.Printf("UpdateBapTuru hatası: %v", err)
	}
	return err
}

// CreateUser, admin tarafından yeni bir kullanıcı ve detaylarını ekler (transaction ile)
// Türkçe Yorum: Yeni kullanıcı oluştururken sifre_degistir_zorla değerini de kaydediyoruz
func (r *AdminRepository) CreateUser(req *models.AdminCreateUserRequest, hashedPass string) error {
	// Veritabanı transaction'ı başlatılır
	tx, err := r.DB.Begin()
	if err != nil {
		log.Printf("CreateUser transaction başlatma hatası: %v", err)
		return err
	}
	defer tx.Rollback()

	// Gelen roller virgülle ayrılmış olabilir, ilkini legacy alanlara yazalım
	firstRole := ""
	roles := strings.Split(req.Rol, ",")
	if len(roles) > 0 {
		firstRole = strings.TrimSpace(roles[0])
	}

	// 1. Uye tablosuna temel verileri ekle
	var uyeID int
	queryUye := `
		INSERT INTO uye (ad, soyad, eposta, sifre_hash, rol, aktif_mi, sifre_degistir_zorla)
		VALUES ($1, $2, $3, $4, $5, true, $6)
		RETURNING uye_id
	`
	err = tx.QueryRow(queryUye, req.Ad, req.Soyad, req.Eposta, hashedPass, firstRole, req.SifreDegistirZorla).Scan(&uyeID)
	if err != nil {
		log.Printf("CreateUser uye tablosu hatası: %v", err)
		return err
	}

	// 2. UyeDetay tablosuna detayları ekle (profil_tamamlandi = true olarak işaretlenir)
	queryDetay := `
		INSERT INTO uye_detay (uye_id, rol, unvan, bolum, telefon, izu_uyesi, profil_tamamlandi)
		VALUES ($1, $2, $3, $4, $5, $6, true)
	`
	_, err = tx.Exec(queryDetay, uyeID, firstRole, req.Unvan, req.Bolum, req.Telefon, req.IzuUyesi)
	if err != nil {
		log.Printf("CreateUser uye_detay tablosu hatası: %v", err)
		return err
	}

	// 3. Sistem_rol tablosuna yetki/rol atamasını ekle (tüm roller için)
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
			log.Printf("CreateUser sistem_rol tablosu hatası: %v", err)
			return err
		}
	}

	// Tüm işlemler başarılı ise transaction commit edilir
	return tx.Commit()
}

// UpdateUser, admin tarafından bir kullanıcının temel ve detay bilgilerini günceller (transaction ile)
func (r *AdminRepository) UpdateUser(uyeID int, req *models.AdminUpdateUserRequest) error {
	// Veritabanı transaction'ı başlatılır
	tx, err := r.DB.Begin()
	if err != nil {
		log.Printf("UpdateUser transaction başlatma hatası: %v", err)
		return err
	}
	defer tx.Rollback()

	// Gelen roller virgülle ayrılmış olabilir, ilkini legacy alanlara yazalım
	firstRole := ""
	roles := strings.Split(req.Rol, ",")
	if len(roles) > 0 {
		firstRole = strings.TrimSpace(roles[0])
	}

	// 1. Uye tablosundaki verileri güncelle
	queryUye := `
		UPDATE uye 
		SET ad = $1, soyad = $2, eposta = $3, rol = $4, guncelleme_tarihi = CURRENT_TIMESTAMP
		WHERE uye_id = $5
	`
	_, err = tx.Exec(queryUye, req.Ad, req.Soyad, req.Eposta, firstRole, uyeID)
	if err != nil {
		log.Printf("UpdateUser uye tablosu güncelleme hatası: %v", err)
		return err
	}

	// 2. UyeDetay tablosundaki verileri güncelle
	queryDetay := `
		UPDATE uye_detay
		SET rol = $1, unvan = $2, bolum = $3, telefon = $4, izu_uyesi = $5, guncelleme_tarihi = CURRENT_TIMESTAMP
		WHERE uye_id = $6
	`
	_, err = tx.Exec(queryDetay, firstRole, req.Unvan, req.Bolum, req.Telefon, req.IzuUyesi, uyeID)
	if err != nil {
		log.Printf("UpdateUser uye_detay tablosu güncelleme hatası: %v", err)
		return err
	}

	// 3. Sistem_rol tablosundaki yetki/rol atamasını güncelle (mevcut rolleri temizle, yenilerini ata)
	_, err = tx.Exec(`DELETE FROM sistem_rol WHERE uye_id = $1`, uyeID)
	if err != nil {
		log.Printf("UpdateUser sistem_rol temizleme hatası: %v", err)
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
	rowsPages, err := r.DB.Query("SELECT sayfa_id, sayfa_adi, sayfa_kodu, url_yolu FROM sistem_sayfa ORDER BY sayfa_id")
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


