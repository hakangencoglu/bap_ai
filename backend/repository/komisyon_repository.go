package repository

import (
	"database/sql"
	"fmt"

	"bap_ai/backend/models"
)

// KomisyonRepository veritabanı komisyon işlemlerini yürütür.
// Türkçe Yorum: Komisyon toplantı kayıtları ve üyelerinin veritabanı etkileşimleri bu katmanda gerçekleşir.
type KomisyonRepository struct {
	DB *sql.DB
}

// NewKomisyonRepository yeni bir KomisyonRepository oluşturur.
func NewKomisyonRepository(db *sql.DB) *KomisyonRepository {
	return &KomisyonRepository{DB: db}
}

// GetCommissionMembers sistemdeki aktif komisyon üyelerini listeler.
// Türkçe Yorum: Üye tablosundan rolü komisyon veya komisyon_baskani olan aktif kullanıcıları detaylarıyla getirir.
func (r *KomisyonRepository) GetCommissionMembers() ([]*models.UyeWithDetay, error) {
	query := `
		SELECT u.uye_id, u.ad, u.soyad, u.eposta, u.rol, 
		       COALESCE(d.unvan, ''), COALESCE(d.bolum, ''), COALESCE(d.telefon, ''), COALESCE(d.izu_uyesi, false)
		FROM uye u
		LEFT JOIN uye_detay d ON u.uye_id = d.uye_id
		WHERE (u.rol LIKE '%komisyon%' OR u.rol LIKE '%komisyon_baskani%') AND u.aktif_mi = true
		ORDER BY u.ad, u.soyad
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("komisyon üyeleri sorgulanamadı: %w", err)
	}
	defer rows.Close()

	var uyeler []*models.UyeWithDetay
	for rows.Next() {
		var u models.UyeWithDetay
		err := rows.Scan(&u.UyeID, &u.Ad, &u.Soyad, &u.Eposta, &u.Rol, &u.Unvan, &u.Bolum, &u.Telefon, &u.IzuUyesi)
		if err != nil {
			return nil, err
		}
		uyeler = append(uyeler, &u)
	}
	return uyeler, nil
}

// GetMeetingsCountByYear belirtilen yıldaki toplantı adedini döner.
// Türkçe Yorum: Otomatik toplantı numarası oluştururken o yıl kaçıncı toplantı olduğunu belirlemek için kullanılır.
func (r *KomisyonRepository) GetMeetingsCountByYear(year int) (int, error) {
	query := `SELECT COUNT(*) FROM komisyon_toplantisi WHERE EXTRACT(YEAR FROM tarih) = $1`
	var count int
	err := r.DB.QueryRow(query, year).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("yıllık toplantı sayısı sorgulanamadı: %w", err)
	}
	return count, nil
}

// CreateMeeting yeni bir toplantı kaydeder ve katılımcı listesini eşleştirir.
// Türkçe Yorum: Transaction kapsamında toplantı detayını ve katılımcı yoklama durumunu veritabanına yazar.
func (r *KomisyonRepository) CreateMeeting(meeting *models.KomisyonToplantisi) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("transaction başlatılamadı: %w", err)
	}
	defer tx.Rollback()

	// 1. Toplantıyı kaydet
	meetingQuery := `
		INSERT INTO komisyon_toplantisi (toplanti_no, tarih, gundem, karar, olusturan_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING toplanti_id, olusturma_tarihi
	`
	err = tx.QueryRow(meetingQuery, meeting.ToplantiNo, meeting.Tarih, meeting.Gundem, meeting.Karar, meeting.OlusturanID).
		Scan(&meeting.ToplantiID, &meeting.OlusturmaTarihi)
	if err != nil {
		return fmt.Errorf("toplantı kaydedilemedi: %w", err)
	}

	// 2. Katılımcıları kaydet
	participantQuery := `
		INSERT INTO komisyon_toplanti_katilimci (toplanti_id, uye_id, katildi)
		VALUES ($1, $2, $3)
	`
	for _, k := range meeting.Katilimcilar {
		_, err := tx.Exec(participantQuery, meeting.ToplantiID, k.UyeID, k.Katildi)
		if err != nil {
			return fmt.Errorf("toplantı katılımcısı kaydedilemedi (UyeID: %d): %w", k.UyeID, err)
		}
	}

	return tx.Commit()
}

// GetMeetingByID belirtilen ID'li toplantı detaylarını katılımcılarıyla getirir.
// Türkçe Yorum: Toplantı genel verilerini ve katılımcı yoklama listesini tek bir modelde birleştirir.
func (r *KomisyonRepository) GetMeetingByID(id int) (*models.KomisyonToplantisi, error) {
	query := `
		SELECT toplanti_id, toplanti_no, tarih, gundem, karar, olusturan_id, olusturma_tarihi
		FROM komisyon_toplantisi
		WHERE toplanti_id = $1
	`
	var m models.KomisyonToplantisi
	err := r.DB.QueryRow(query, id).Scan(&m.ToplantiID, &m.ToplantiNo, &m.Tarih, &m.Gundem, &m.Karar, &m.OlusturanID, &m.OlusturmaTarihi)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("toplantı sorgulanamadı: %w", err)
	}

	// Katılımcıları çek
	katilimciQuery := `
		SELECT k.uye_id, u.ad, u.soyad, COALESCE(d.unvan, ''), COALESCE(d.bolum, ''), k.katildi
		FROM komisyon_toplanti_katilimci k
		INNER JOIN uye u ON k.uye_id = u.uye_id
		LEFT JOIN uye_detay d ON u.uye_id = d.uye_id
		WHERE k.toplanti_id = $1
		ORDER BY u.ad, u.soyad
	`
	rows, err := r.DB.Query(katilimciQuery, id)
	if err != nil {
		return nil, fmt.Errorf("toplantı katılımcıları sorgulanamadı: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var k models.KomisyonToplantiKatilim
		err := rows.Scan(&k.UyeID, &k.Ad, &k.Soyad, &k.Unvan, &k.Bolum, &k.Katildi)
		if err != nil {
			return nil, err
		}
		m.Katilimcilar = append(m.Katilimcilar, &k)
	}

	return &m, nil
}

// ListMeetings tüm komisyon toplantılarını tarihe göre tersten listeler.
// Türkçe Yorum: Toplantı geçmişini ana panoda listelemek için tüm verileri çeker.
func (r *KomisyonRepository) ListMeetings() ([]*models.KomisyonToplantisi, error) {
	query := `
		SELECT toplanti_id, toplanti_no, tarih, gundem, karar, olusturan_id, olusturma_tarihi
		FROM komisyon_toplantisi
		ORDER BY tarih DESC, toplanti_id DESC
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("toplantılar listelenemedi: %w", err)
	}
	defer rows.Close()

	var toplantilar []*models.KomisyonToplantisi
	for rows.Next() {
		var m models.KomisyonToplantisi
		err := rows.Scan(&m.ToplantiID, &m.ToplantiNo, &m.Tarih, &m.Gundem, &m.Karar, &m.OlusturanID, &m.OlusturmaTarihi)
		if err != nil {
			return nil, err
		}
		toplantilar = append(toplantilar, &m)
	}
	return toplantilar, nil
}
