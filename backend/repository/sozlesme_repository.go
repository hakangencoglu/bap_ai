package repository

import (
	"database/sql"
	"fmt"

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
// Türkçe Yorum: Yürürlük tarihleri PDF indirme anında hesaplandığı için burada kaydedilmez.
// Ayrıca daha önce indirilmiş bir sözleşme tekrar kaydedilerek indirme kilidi sıfırlanamaz.
func (r *SozlesmeRepository) SaveSozlesme(s *models.ProjeSozlesme) error {
	query := `
		INSERT INTO proje_sozlesme (
			proje_id, uye_id, tc_kimlik, yurutucu_adres, yurutucu_telefon, yurutucu_eposta, durum, guncelleme_tarihi
		) VALUES ($1, $2, $3, $4, $5, $6, 'dolduruldu', NOW())
		ON CONFLICT (proje_id) DO UPDATE SET
			tc_kimlik = EXCLUDED.tc_kimlik,
			yurutucu_adres = EXCLUDED.yurutucu_adres,
			yurutucu_telefon = EXCLUDED.yurutucu_telefon,
			yurutucu_eposta = EXCLUDED.yurutucu_eposta,
			durum = 'dolduruldu',
			guncelleme_tarihi = NOW()
		RETURNING id, olusturma_tarihi, guncelleme_tarihi;
	`
	err := r.DB.QueryRow(query,
		s.ProjeID, s.UyeID, s.TCKimlik, s.YurutucuAdres, s.YurutucuTelefon, s.YurutucuEposta,
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
		       p.proje_kodu, p.baslik_tr,
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
