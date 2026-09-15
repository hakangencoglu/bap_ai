package repository

import (
	"database/sql"
	"fmt"
	"bap_ai/backend/models"
)

// FakulteRepository veritabanı fakülte ve bölüm işlemlerini yönetir.
type FakulteRepository struct {
	DB *sql.DB
}

// NewFakulteRepository yeni bir FakulteRepository örneği oluşturur.
func NewFakulteRepository(db *sql.DB) *FakulteRepository {
	return &FakulteRepository{DB: db}
}

// GetFakulteler sistemdeki tüm aktif fakülteleri alfabetik olarak listeler.
// Türkçe Yorum: Kullanıcılara sunulacak açılır fakülte listesi için veri çeker.
func (r *FakulteRepository) GetFakulteler() ([]models.Fakulte, error) {
	query := `
		SELECT fakulte_id, fakulte_adi, COALESCE(fakulte_kodu, ''), COALESCE(kisa_ad, ''), aktif, olusturma_tarihi, guncelleme_tarihi
		FROM fakulte
		WHERE aktif = true
		ORDER BY fakulte_adi ASC
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("fakülteler listelenemedi: %w", err)
	}
	defer rows.Close()

	var list []models.Fakulte
	for rows.Next() {
		var f models.Fakulte
		if err := rows.Scan(&f.FakulteID, &f.FakulteAdi, &f.FakulteKodu, &f.KisaAd, &f.Aktif, &f.OlusturmaTarihi, &f.GuncellemeTarihi); err != nil {
			return nil, fmt.Errorf("fakülte satırı okunamadı: %w", err)
		}
		list = append(list, f)
	}
	if list == nil {
		list = []models.Fakulte{}
	}
	return list, nil
}

// GetBolumler belirli bir fakülteye bağlı veya tüm aktif bölümleri listeler.
// Türkçe Yorum: fakulteID verilirse fakulte_bolum ilişki tablosu üzerinden sadece o fakültenin bölümlerini döner.
func (r *FakulteRepository) GetBolumler(fakulteID ...int) ([]models.Bolum, error) {
	var query string
	var rows *sql.Rows
	var err error

	if len(fakulteID) > 0 && fakulteID[0] > 0 {
		query = `
			SELECT b.bolum_id, b.bolum_adi, COALESCE(b.bolum_kodu, ''), COALESCE(b.kisa_ad, ''), b.aktif, b.olusturma_tarihi, b.guncelleme_tarihi
			FROM bolum b
			JOIN fakulte_bolum fb ON fb.bolum_id = b.bolum_id
			WHERE fb.fakulte_id = $1 AND fb.aktif = true AND b.aktif = true
			ORDER BY b.bolum_adi ASC
		`
		rows, err = r.DB.Query(query, fakulteID[0])
	} else {
		query = `
			SELECT bolum_id, bolum_adi, COALESCE(bolum_kodu, ''), COALESCE(kisa_ad, ''), aktif, olusturma_tarihi, guncelleme_tarihi
			FROM bolum
			WHERE aktif = true
			ORDER BY bolum_adi ASC
		`
		rows, err = r.DB.Query(query)
	}

	if err != nil {
		return nil, fmt.Errorf("bölümler listelenemedi: %w", err)
	}
	defer rows.Close()

	var list []models.Bolum
	for rows.Next() {
		var b models.Bolum
		if err := rows.Scan(&b.BolumID, &b.BolumAdi, &b.BolumKodu, &b.KisaAd, &b.Aktif, &b.OlusturmaTarihi, &b.GuncellemeTarihi); err != nil {
			return nil, fmt.Errorf("bölüm satırı okunamadı: %w", err)
		}
		list = append(list, b)
	}
	if list == nil {
		list = []models.Bolum{}
	}
	return list, nil
}

// GetFakulteByID ID'ye göre fakülte detayını getirir.
func (r *FakulteRepository) GetFakulteByID(fakulteID int) (*models.Fakulte, error) {
	query := `
		SELECT fakulte_id, fakulte_adi, COALESCE(fakulte_kodu, ''), COALESCE(kisa_ad, ''), aktif, olusturma_tarihi, guncelleme_tarihi
		FROM fakulte
		WHERE fakulte_id = $1
	`
	var f models.Fakulte
	err := r.DB.QueryRow(query, fakulteID).Scan(&f.FakulteID, &f.FakulteAdi, &f.FakulteKodu, &f.KisaAd, &f.Aktif, &f.OlusturmaTarihi, &f.GuncellemeTarihi)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// GetBolumByID ID'ye göre bölüm detayını getirir.
func (r *FakulteRepository) GetBolumByID(bolumID int) (*models.Bolum, error) {
	query := `
		SELECT bolum_id, bolum_adi, COALESCE(bolum_kodu, ''), COALESCE(kisa_ad, ''), aktif, olusturma_tarihi, guncelleme_tarihi
		FROM bolum
		WHERE bolum_id = $1
	`
	var b models.Bolum
	err := r.DB.QueryRow(query, bolumID).Scan(&b.BolumID, &b.BolumAdi, &b.BolumKodu, &b.KisaAd, &b.Aktif, &b.OlusturmaTarihi, &b.GuncellemeTarihi)
	if err != nil {
		return nil, err
	}
	return &b, nil
}
