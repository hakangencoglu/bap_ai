package repository

import (
	"database/sql"
	"fmt"

	"bap_ai/backend/models"
)

// BildirimRepository veritabanındaki bildirim işlemlerini yönetir.
// Türkçe Yorum: Kullanıcılara özel sistem içi bildirimlerin çekilmesi ve okundu durumunun güncellenmesi işlemlerini yürüten katmandır.
type BildirimRepository struct {
	DB *sql.DB
}

// NewBildirimRepository yeni bir BildirimRepository oluşturur.
// Türkçe Yorum: BildirimRepository için dependency injection kurucusu.
func NewBildirimRepository(db *sql.DB) *BildirimRepository {
	return &BildirimRepository{DB: db}
}

// GetBildirimlerByUyeID kullanıcının bildirimlerini tersten kronolojik olarak getirir.
// Türkçe Yorum: Giriş yapan kullanıcının tüm bildirimlerini en yeni tarihten en eski tarihe doğru listeler.
func (r *BildirimRepository) GetBildirimlerByUyeID(uyeID int) ([]models.Bildirim, error) {
	query := `
		SELECT bildirim_id, uye_id, baslik, icerik, okundu, olusturma_tarihi
		FROM bildirim
		WHERE uye_id = $1
		ORDER BY olusturma_tarihi DESC
	`
	rows, err := r.DB.Query(query, uyeID)
	if err != nil {
		return nil, fmt.Errorf("bildirimler sorgulanırken hata: %w", err)
	}
	defer rows.Close()

	var list []models.Bildirim
	for rows.Next() {
		var b models.Bildirim
		err := rows.Scan(&b.BildirimID, &b.UyeID, &b.Baslik, &b.Icerik, &b.Okundu, &b.OlusturmaTarihi)
		if err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, nil
}

// MarkAsRead belirtilen bildirimi okundu olarak işaretler.
// Türkçe Yorum: Kullanıcı belirli bir bildirimi okuduğunda okundu bayrağını veritabanında günceller.
func (r *BildirimRepository) MarkAsRead(bildirimID int, uyeID int) error {
	query := `
		UPDATE bildirim
		SET okundu = true
		WHERE bildirim_id = $1 AND uye_id = $2
	`
	_, err := r.DB.Exec(query, bildirimID, uyeID)
	return err
}

// MarkAllAsRead kullanıcının tüm bildirimlerini okundu olarak işaretler.
// Türkçe Yorum: Kullanıcı "Tümünü Okundu İşaretle" seçeneğini tıkladığında çalıştırılan sorgudur.
func (r *BildirimRepository) MarkAllAsRead(uyeID int) error {
	query := `
		UPDATE bildirim
		SET okundu = true
		WHERE uye_id = $1
	`
	_, err := r.DB.Exec(query, uyeID)
	return err
}
