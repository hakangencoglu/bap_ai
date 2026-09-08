package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"bap_ai/backend/models"
)

// FeedbackRepository geri bildirim veri tabanı işlemlerinden sorumludur.
// Türkçe Yorum: Geri bildirim verilerini kaydetme ve gönderen/alıcı bilgilerini sorgulama katmanı.
type FeedbackRepository struct {
	DB *sql.DB
}

// NewFeedbackRepository yeni bir FeedbackRepository örneği oluşturur.
// Türkçe Yorum: FeedbackRepository yapısını ilklendiren kurucu fonksiyon.
func NewFeedbackRepository(db *sql.DB) *FeedbackRepository {
	return &FeedbackRepository{DB: db}
}

// CreateFeedback geri bildirim kaydını veritabanına ekler.
// Türkçe Yorum: Geri bildirim mesajını 'geri_bildirim' tablosuna kaydeder.
func (r *FeedbackRepository) CreateFeedback(fb *models.Feedback) error {
	query := `
		INSERT INTO geri_bildirim (uye_id, konu, mesaj, sayfa_url, durum)
		VALUES ($1, $2, $3, $4, 'yeni')
		RETURNING id, olusturma_tarihi
	`
	err := r.DB.QueryRow(query, fb.UyeID, fb.Konu, fb.Mesaj, fb.SayfaURL).Scan(&fb.ID, &fb.OlusturmaTarihi)
	if err != nil {
		return fmt.Errorf("geri bildirim kaydedilemedi: %w", err)
	}
	return nil
}

// GetAdminEmails sistemdeki aktif admin kullanıcılarının e-posta adreslerini döner.
// Türkçe Yorum: 'admin' rolüne sahip ve hesabı aktif olan üyelerin e-posta listesini çeker.
func (r *FeedbackRepository) GetAdminEmails() ([]string, error) {
	query := `
		SELECT DISTINCT u.eposta
		FROM uye u
		INNER JOIN sistem_rol sr ON u.uye_id = sr.uye_id
		INNER JOIN sistem_rol_tanimlama srt ON sr.sistem_rol_id = srt.rol_id
		WHERE srt.rol_adi = 'admin' AND u.aktif_mi = true
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("admin e-postaları sorgulanamadı: %w", err)
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err == nil && strings.TrimSpace(email) != "" {
			emails = append(emails, strings.TrimSpace(email))
		}
	}
	return emails, nil
}

// GetSenderFullDetails belirtilen üyenin profil bilgilerini ve tüm sistem rollerini getirir.
// Türkçe Yorum: E-posta imza kartı için gönderenin ad, unvan, iletişim ve tüm rol yetkilerini derler.
func (r *FeedbackRepository) GetSenderFullDetails(uyeID int) (*models.FeedbackSenderInfo, error) {
	info := &models.FeedbackSenderInfo{
		UyeID:      uyeID,
		Roles:      []string{},
		RoleLabels: []string{},
	}

	// 1. Üye bilgilerini al
	var mainRole string
	uyeQuery := `
		SELECT ad, soyad, COALESCE(unvan, ''), COALESCE(bolum, ''), eposta, COALESCE(telefon, ''), COALESCE(rol, '')
		FROM uye
		WHERE uye_id = $1
	`
	err := r.DB.QueryRow(uyeQuery, uyeID).Scan(
		&info.Ad,
		&info.Soyad,
		&info.Unvan,
		&info.Bolum,
		&info.Eposta,
		&info.Telefon,
		&mainRole,
	)
	if err != nil {
		return nil, fmt.Errorf("gönderen üye bilgileri alınamadı: %w", err)
	}

	// 2. Üyenin tüm sistem rollerini ve etiketlerini al
	rolQuery := `
		SELECT srt.rol_adi, COALESCE(srt.rol_etiketi, srt.rol_adi)
		FROM sistem_rol sr
		INNER JOIN sistem_rol_tanimlama srt ON sr.sistem_rol_id = srt.rol_id
		WHERE sr.uye_id = $1
	`
	rows, err := r.DB.Query(rolQuery, uyeID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var roleName, roleLabel string
			if err := rows.Scan(&roleName, &roleLabel); err == nil {
				info.Roles = append(info.Roles, roleName)
				info.RoleLabels = append(info.RoleLabels, roleLabel)
			}
		}
	}

	// 3. Eğer sistem_rol tablosunda kayıt yoksa uye.rol alanını fallback olarak kullan
	if len(info.Roles) == 0 && strings.TrimSpace(mainRole) != "" {
		mainRoleTrimmed := strings.TrimSpace(mainRole)
		var label string
		errLabel := r.DB.QueryRow("SELECT COALESCE(rol_etiketi, rol_adi) FROM sistem_rol_tanimlama WHERE rol_adi = $1", mainRoleTrimmed).Scan(&label)
		if errLabel != nil || label == "" {
			label = mainRoleTrimmed
		}
		info.Roles = append(info.Roles, mainRoleTrimmed)
		info.RoleLabels = append(info.RoleLabels, label)
	}

	return info, nil
}
