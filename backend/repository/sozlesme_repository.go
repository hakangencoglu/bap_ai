package repository

import (
	"database/sql"
	"fmt"
	"time"

	"bap_ai/backend/models"
)

// SozlesmeRepository, proje sözleşmesi veri erişim katmanıdır.
// Türkçe Yorum: Sözleşme kaydetme, güncelleme ve proje ID'sine göre getirme sorgularını içerir.
type SozlesmeRepository struct {
	DB *sql.DB
}

// NewSozlesmeRepository yeni bir SozlesmeRepository örneği oluşturur.
func NewSozlesmeRepository(db *sql.DB) *SozlesmeRepository {
	return &SozlesmeRepository{DB: db}
}

// SaveSozlesme, projeye ait sözleşme verisini ekler veya varsa günceller.
// Türkçe Yorum: Otomatik hesaplanan veya seçilen yürürlük tarihleri veritabanına kaydedilir.
// Ayrıca daha önce indirilmiş bir sözleşme tekrar kaydedilerek indirme kilidi sıfırlanamaz.
func (r *SozlesmeRepository) SaveSozlesme(s *models.ProjeSozlesme) error {
	// Türkçe Yorum: EXCLUDED tarihleri zaten DATE; '' ile NULLIF DATE cast hatası verir (22007).
	query := `
		INSERT INTO proje_sozlesme (
			proje_id, uye_id, tc_kimlik, yurutucu_adres, yurutucu_telefon, yurutucu_eposta, baslangic_tarihi, bitis_tarihi, durum, guncelleme_tarihi
		) VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, '')::DATE, NULLIF($8, '')::DATE, 'dolduruldu', NOW())
		ON CONFLICT (proje_id) DO UPDATE SET
			tc_kimlik = EXCLUDED.tc_kimlik,
			yurutucu_adres = EXCLUDED.yurutucu_adres,
			yurutucu_telefon = EXCLUDED.yurutucu_telefon,
			yurutucu_eposta = EXCLUDED.yurutucu_eposta,
			baslangic_tarihi = COALESCE(EXCLUDED.baslangic_tarihi, proje_sozlesme.baslangic_tarihi),
			bitis_tarihi = COALESCE(EXCLUDED.bitis_tarihi, proje_sozlesme.bitis_tarihi),
			durum = 'dolduruldu',
			guncelleme_tarihi = NOW()
		RETURNING id, olusturma_tarihi, guncelleme_tarihi;
	`
	err := r.DB.QueryRow(query,
		s.ProjeID, s.UyeID, s.TCKimlik, s.YurutucuAdres, s.YurutucuTelefon, s.YurutucuEposta, s.BaslangicTarihi, s.BitisTarihi,
	).Scan(&s.ID, &s.OlusturmaTarihi, &s.GuncellemeTarihi)

	if err != nil {
		return fmt.Errorf("sözleşme kaydedilemedi: %w", err)
	}
	return nil
}

// MarkIndirildi, sözleşme PDF'inin indirildiğini işaretler ve hesaplanan yürürlük tarihlerini kaydeder.
// Türkçe Yorum: Tek seferlik indirme kuralı için indirildi_mi bayrağı TRUE yapılır.
func (r *SozlesmeRepository) MarkIndirildi(projeID int, baslangic, bitis string) error {
	query := `
		UPDATE proje_sozlesme
		SET indirildi_mi = TRUE,
		    indirme_tarihi = NOW(),
		    baslangic_tarihi = $2,
		    bitis_tarihi = $3,
		    durum = 'indirildi',
		    guncelleme_tarihi = NOW()
		WHERE proje_id = $1 AND indirildi_mi = FALSE
	`
	res, err := r.DB.Exec(query, projeID, baslangic, bitis)
	if err != nil {
		return fmt.Errorf("sözleşme indirme durumu güncellenemedi: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("sözleşme bulunamadı veya zaten indirilmiş")
	}
	return nil
}

// GetSozlesmeByProjeID, belirtilen proje ID'sine ait sözleşme kaydını getirir.
func (r *SozlesmeRepository) GetSozlesmeByProjeID(projeID int) (*models.ProjeSozlesme, error) {
	query := `
		SELECT s.id, s.proje_id, s.uye_id, s.tc_kimlik, s.yurutucu_adres, s.yurutucu_telefon,
		       s.yurutucu_eposta, COALESCE(TO_CHAR(s.baslangic_tarihi, 'YYYY-MM-DD'), ''), COALESCE(TO_CHAR(s.bitis_tarihi, 'YYYY-MM-DD'), ''),
		       s.durum, COALESCE(s.indirildi_mi, false), s.indirme_tarihi, s.olusturma_tarihi, s.guncelleme_tarihi,
		       p.proje_kodu, (SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'),
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad
		FROM proje_sozlesme s
		JOIN proje p ON p.proje_id = s.proje_id
		JOIN uye u   ON u.uye_id   = s.uye_id
		WHERE s.proje_id = $1;
	`
	var s models.ProjeSozlesme
	err := r.DB.QueryRow(query, projeID).Scan(
		&s.ID, &s.ProjeID, &s.UyeID, &s.TCKimlik, &s.YurutucuAdres, &s.YurutucuTelefon,
		&s.YurutucuEposta, &s.BaslangicTarihi, &s.BitisTarihi,
		&s.Durum, &s.IndirildiMi, &s.IndirmeTarihi, &s.OlusturmaTarihi, &s.GuncellemeTarihi,
		&s.ProjeKodu, &s.ProjeBaslik, &s.YurutucuAd,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("sözleşme okunamadı: %w", err)
	}
	return &s, nil
}

// GetActiveSignedContractsForReminder, aylık e-posta hatırlatması gönderilecek aktif ve yürürlükte olan tüm sözleşmeleri getirir.
// Türkçe Yorum: Başlangıç tarihi girilmiş sözleşmeleri, en son gönderilen hatırlatma tarihi ve dönemi ile birlikte sorgular.
func (r *SozlesmeRepository) GetActiveSignedContractsForReminder() ([]models.ProjeSozlesmeHatirlatmaInfo, error) {
	query := `
		SELECT s.id, s.proje_id, COALESCE(p.proje_kodu, ''), COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'), ''),
		       COALESCE(u.unvan||' ','') || u.ad || ' ' || u.soyad AS yurutucu_ad,
		       COALESCE(s.yurutucu_eposta, u.eposta),
		       s.baslangic_tarihi, s.bitis_tarihi, COALESCE(p.sure_ay, 12),
		       l.gonderim_tarihi, COALESCE(l.donem_indeks, 0)
		FROM proje_sozlesme s
		JOIN proje p ON p.proje_id = s.proje_id
		JOIN uye u   ON u.uye_id   = s.uye_id
		LEFT JOIN LATERAL (
			SELECT gonderim_tarihi, donem_indeks
			FROM proje_sozlesme_hatirlatma_log
			WHERE sozlesme_id = s.id
			ORDER BY gonderim_tarihi DESC
			LIMIT 1
		) l ON true
		WHERE s.baslangic_tarihi IS NOT NULL;
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("aktif sözleşmeler sorgulanamadı: %w", err)
	}
	defer rows.Close()

	var list []models.ProjeSozlesmeHatirlatmaInfo
	for rows.Next() {
		var item models.ProjeSozlesmeHatirlatmaInfo
		var sonHatirlatma *time.Time
		var sonDonem int

		if err := rows.Scan(
			&item.SozlesmeID, &item.ProjeID, &item.ProjeKodu, &item.ProjeBaslik,
			&item.YurutucuAd, &item.YurutucuEposta,
			&item.BaslangicTarihi, &item.BitisTarihi, &item.SureAy,
			&sonHatirlatma, &sonDonem,
		); err != nil {
			return nil, fmt.Errorf("sözleşme hatırlatma verisi okunamadı: %w", err)
		}
		item.SonHatirlatmaTarihi = sonHatirlatma
		item.SonDonemIndeks = sonDonem
		list = append(list, item)
	}

	return list, nil
}

// SaveReminderLog, gönderilen aylık e-posta hatırlatma kaydını veritabanına işler.
// Türkçe Yorum: E-posta başarıyla iletildikten sonra proje_sozlesme_hatirlatma_log tablosuna log ekler.
func (r *SozlesmeRepository) SaveReminderLog(logItem *models.SozlesmeHatirlatmaLog) error {
	query := `
		INSERT INTO proje_sozlesme_hatirlatma_log (
			sozlesme_id, proje_id, gonderim_tarihi, gecen_sure, kalan_sure, gonderilen_eposta, donem_indeks, durum
		) VALUES ($1, $2, NOW(), $3, $4, $5, $6, $7)
		RETURNING id, gonderim_tarihi;
	`
	err := r.DB.QueryRow(query,
		logItem.SozlesmeID, logItem.ProjeID, logItem.GecenSure, logItem.KalanSure,
		logItem.GonderilenEposta, logItem.DonemIndeks, logItem.Durum,
	).Scan(&logItem.ID, &logItem.GonderimTarihi)

	if err != nil {
		return fmt.Errorf("hatırlatma log kaydı eklenemedi: %w", err)
	}
	return nil
}

