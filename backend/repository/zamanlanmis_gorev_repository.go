package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"bap_ai/backend/models"
)

// ZamanlanmisGorevRepository veritabanı kural ve log işlemlerini yönetir.
// Türkçe Yorum: Zamanlanmış görevlerin CRUD ve tetiklenme sorgularını gerçekleştirir.
type ZamanlanmisGorevRepository struct {
	DB *sql.DB
}

// NewZamanlanmisGorevRepository yeni bir ZamanlanmisGorevRepository oluşturur.
func NewZamanlanmisGorevRepository(db *sql.DB) *ZamanlanmisGorevRepository {
	return &ZamanlanmisGorevRepository{DB: db}
}

// normalizeAliciHedefleri boş alıcı listesini varsayılan yürütücü hedefiyle doldurur.
func normalizeAliciHedefleri(hedefler []string) []string {
	if len(hedefler) == 0 {
		return []string{models.AliciHedefYurutucu}
	}
	seen := map[string]bool{}
	var normalized []string
	for _, h := range hedefler {
		key := strings.TrimSpace(strings.ToLower(h))
		if key == "" || seen[key] {
			continue
		}
		if key != models.AliciHedefYurutucu && key != models.AliciHedefTTO {
			continue
		}
		seen[key] = true
		normalized = append(normalized, key)
	}
	if len(normalized) == 0 {
		return []string{models.AliciHedefYurutucu}
	}
	return normalized
}

// parseAliciHedefleri JSONB alıcı hedeflerini diziye çevirir.
func parseAliciHedefleri(raw []byte) []string {
	if len(raw) == 0 {
		return []string{models.AliciHedefYurutucu}
	}
	var hedefler []string
	if err := json.Unmarshal(raw, &hedefler); err != nil {
		return []string{models.AliciHedefYurutucu}
	}
	return normalizeAliciHedefleri(hedefler)
}

// encodeAliciHedefleri alıcı hedeflerini JSONB için hazırlar.
func encodeAliciHedefleri(hedefler []string) ([]byte, error) {
	return json.Marshal(normalizeAliciHedefleri(hedefler))
}

func scanKuralRow(scanner interface {
	Scan(dest ...interface{}) error
}, k *models.ZamanlanmisGorevKural) error {
	var aliciRaw []byte
	err := scanner.Scan(
		&k.KuralID, &k.KuralAdi, &k.BapTuruID, &k.BapTuruAdi,
		&k.TetiklemeTipi, &k.ZamanDegeri, &k.EpostaAktif, &k.SmsAktif,
		&k.EpostaKonu, &k.EpostaSablon, &k.SmsSablon, &k.AktifMi,
		&k.OlusturanID, &k.OlusturmaTarihi, &k.GuncellemeTarihi, &aliciRaw,
	)
	if err != nil {
		return err
	}
	k.AliciHedefleri = parseAliciHedefleri(aliciRaw)
	return nil
}

// GetAllRules tüm zamanlanmış görev kurallarını getirir.
func (r *ZamanlanmisGorevRepository) GetAllRules() ([]*models.ZamanlanmisGorevKural, error) {
	query := `
		SELECT 
			zgk.kural_id, zgk.kural_adi, zgk.bap_turu_id, 
			COALESCE(pbt.bap_turu, 'Tüm BAP Türleri') AS bap_turu_adi,
			zgk.tetikleme_tipi, zgk.zaman_degeri, zgk.eposta_aktif, zgk.sms_aktif,
			zgk.eposta_konu, zgk.eposta_sablon, zgk.sms_sablon, zgk.aktif_mi,
			zgk.olusturan_id, zgk.olusturma_tarihi, zgk.guncelleme_tarihi,
			COALESCE(zgk.alici_hedefleri, '["yurutucu"]'::jsonb)
		FROM zamanlanmis_gorev_kural zgk
		LEFT JOIN proje_bap_turu pbt ON pbt.bap_turu_id = zgk.bap_turu_id
		ORDER BY zgk.kural_id DESC
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("kurallar sorgulanamadı: %w", err)
	}
	defer rows.Close()

	var kurallar []*models.ZamanlanmisGorevKural
	for rows.Next() {
		var k models.ZamanlanmisGorevKural
		if err := scanKuralRow(rows, &k); err != nil {
			return nil, err
		}
		kurallar = append(kurallar, &k)
	}
	return kurallar, nil
}

// GetRuleByID ID'ye göre kuralı getirir.
func (r *ZamanlanmisGorevRepository) GetRuleByID(kuralID int) (*models.ZamanlanmisGorevKural, error) {
	query := `
		SELECT 
			zgk.kural_id, zgk.kural_adi, zgk.bap_turu_id, 
			COALESCE(pbt.bap_turu, 'Tüm BAP Türleri') AS bap_turu_adi,
			zgk.tetikleme_tipi, zgk.zaman_degeri, zgk.eposta_aktif, zgk.sms_aktif,
			zgk.eposta_konu, zgk.eposta_sablon, zgk.sms_sablon, zgk.aktif_mi,
			zgk.olusturan_id, zgk.olusturma_tarihi, zgk.guncelleme_tarihi,
			COALESCE(zgk.alici_hedefleri, '["yurutucu"]'::jsonb)
		FROM zamanlanmis_gorev_kural zgk
		LEFT JOIN proje_bap_turu pbt ON pbt.bap_turu_id = zgk.bap_turu_id
		WHERE zgk.kural_id = $1
	`
	var k models.ZamanlanmisGorevKural
	err := scanKuralRow(r.DB.QueryRow(query, kuralID), &k)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &k, nil
}

// CreateRule yeni bir zamanlanmış kural ekler.
func (r *ZamanlanmisGorevRepository) CreateRule(k *models.ZamanlanmisGorevKural) error {
	k.AliciHedefleri = normalizeAliciHedefleri(k.AliciHedefleri)
	aliciJSON, err := encodeAliciHedefleri(k.AliciHedefleri)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO zamanlanmis_gorev_kural (
			kural_adi, bap_turu_id, tetikleme_tipi, zaman_degeri,
			eposta_aktif, sms_aktif, eposta_konu, eposta_sablon, sms_sablon,
			aktif_mi, olusturan_id, alici_hedefleri
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING kural_id, olusturma_tarihi, guncelleme_tarihi
	`
	return r.DB.QueryRow(
		query,
		k.KuralAdi, k.BapTuruID, k.TetiklemeTipi, k.ZamanDegeri,
		k.EpostaAktif, k.SmsAktif, k.EpostaKonu, k.EpostaSablon, k.SmsSablon,
		k.AktifMi, k.OlusturanID, aliciJSON,
	).Scan(&k.KuralID, &k.OlusturmaTarihi, &k.GuncellemeTarihi)
}

// UpdateRule zamanlanmış bir kuralı günceller.
func (r *ZamanlanmisGorevRepository) UpdateRule(k *models.ZamanlanmisGorevKural) error {
	k.AliciHedefleri = normalizeAliciHedefleri(k.AliciHedefleri)
	aliciJSON, err := encodeAliciHedefleri(k.AliciHedefleri)
	if err != nil {
		return err
	}

	query := `
		UPDATE zamanlanmis_gorev_kural SET
			kural_adi = $1,
			bap_turu_id = $2,
			tetikleme_tipi = $3,
			zaman_degeri = $4,
			eposta_aktif = $5,
			sms_aktif = $6,
			eposta_konu = $7,
			eposta_sablon = $8,
			sms_sablon = $9,
			aktif_mi = $10,
			alici_hedefleri = $11,
			guncelleme_tarihi = CURRENT_TIMESTAMP
		WHERE kural_id = $12
	`
	_, err = r.DB.Exec(
		query,
		k.KuralAdi, k.BapTuruID, k.TetiklemeTipi, k.ZamanDegeri,
		k.EpostaAktif, k.SmsAktif, k.EpostaKonu, k.EpostaSablon, k.SmsSablon,
		k.AktifMi, aliciJSON, k.KuralID,
	)
	return err
}

// DeleteRule zamanlanmış kuralı siler.
func (r *ZamanlanmisGorevRepository) DeleteRule(kuralID int) error {
	query := `DELETE FROM zamanlanmis_gorev_kural WHERE kural_id = $1`
	_, err := r.DB.Exec(query, kuralID)
	return err
}

// DeleteRulesBulk seçili kuralları toplu olarak siler.
// Türkçe Yorum: Admin panelinde çoklu seçim ile kural silme işlemini gerçekleştirir.
func (r *ZamanlanmisGorevRepository) DeleteRulesBulk(kuralIDs []int) (int64, error) {
	if len(kuralIDs) == 0 {
		return 0, fmt.Errorf("silinecek kural seçilmedi")
	}

	placeholders := make([]string, len(kuralIDs))
	args := make([]interface{}, len(kuralIDs))
	for i, id := range kuralIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`DELETE FROM zamanlanmis_gorev_kural WHERE kural_id IN (%s)`, strings.Join(placeholders, ","))
	result, err := r.DB.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("kurallar silinemedi: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

// GetActiveProjectsForRules kuralla eşleşen aktif projeleri ve yürütücü bilgilerini sorgular.
// Türkçe Yorum: Yürürlükte veya onaylı projelerin başlangıç/bitiş tarihlerini ve sürelerini alarak filtreler.
func (r *ZamanlanmisGorevRepository) GetActiveProjectsForRules(bapTuruID *int) ([]*models.KuralEslesenProje, error) {
	query := `
		SELECT 
			p.proje_id,
			COALESCE(p.proje_kodu, 'PROJE-' || p.proje_id),
			COALESCE((SELECT pb.baslik FROM proje_baslik pb WHERE pb.proje_id = p.proje_id AND pb.dil_kodu = 'tr' LIMIT 1), 'Başlıksız Proje'),
			COALESCE(pbt.bap_turu, 'Genel'),
			COALESCE(p.sure_ay, 12),
			COALESCE(ps.baslangic_tarihi, p.olusturma_tarihi::date),
			COALESCE(ps.bitis_tarihi, (COALESCE(ps.baslangic_tarihi, p.olusturma_tarihi::date) + (COALESCE(p.sure_ay, 12) || ' month')::interval)::date),
			u.uye_id,
			COALESCE(NULLIF(TRIM(COALESCE(u.unvan,'') || ' ' || u.ad || ' ' || u.soyad), ''), 'Bilinmiyor'),
			COALESCE(u.eposta, ''),
			COALESCE(u.telefon, '')
		FROM proje p
		JOIN proje_durum pd ON pd.durum_id = p.durum_id
		JOIN uye u ON u.uye_id = p.koordinator_id
		LEFT JOIN proje_bap_turu pbt ON pbt.bap_turu_id = p.bap_turu_id
		LEFT JOIN proje_sozlesme ps ON ps.proje_id = p.proje_id
		WHERE pd.durum_adi IN ('yururlukte', 'onaylandi', 'tamamlandi')
	`
	args := []interface{}{}
	if bapTuruID != nil && *bapTuruID > 0 {
		query += ` AND p.bap_turu_id = $1`
		args = append(args, *bapTuruID)
	}

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("aktif projeler sorgulanamadı: %w", err)
	}
	defer rows.Close()

	var projeler []*models.KuralEslesenProje
	for rows.Next() {
		var ep models.KuralEslesenProje
		var bas, bit time.Time
		err := rows.Scan(
			&ep.ProjeID, &ep.ProjeKodu, &ep.ProjeBaslik, &ep.BapTuru, &ep.SureAy,
			&bas, &bit, &ep.YurutucuID, &ep.YurutucuAd, &ep.YurutucuEposta, &ep.YurutucuTelefon,
		)
		if err != nil {
			return nil, err
		}
		ep.BaslangicTarihi = bas
		ep.BitisTarihi = bit
		projeler = append(projeler, &ep)
	}
	return projeler, nil
}

// IsAlreadySent verilen kural, proje, kanal ve alıcı için yakın zamanda bildirim gönderilip gönderilmediğini denetler.
func (r *ZamanlanmisGorevRepository) IsAlreadySent(kuralID, projeID int, kanal, alici string) bool {
	query := `
		SELECT COUNT(1) 
		FROM zamanlanmis_gorev_log 
		WHERE kural_id = $1 AND proje_id = $2 AND kanal = $3 AND alici = $4
		  AND gonderim_tarihi >= CURRENT_TIMESTAMP - INTERVAL '25 days'
	`
	var count int
	err := r.DB.QueryRow(query, kuralID, projeID, kanal, alici).Scan(&count)
	return err == nil && count > 0
}

// GetTTORecipients aktif TTO rolündeki kullanıcıların iletişim bilgilerini döner.
func (r *ZamanlanmisGorevRepository) GetTTORecipients() ([]models.BildirimAliciKisi, error) {
	query := `
		SELECT DISTINCT
			COALESCE(NULLIF(TRIM(COALESCE(d.unvan, '') || ' ' || u.ad || ' ' || u.soyad), ''), u.ad || ' ' || u.soyad),
			COALESCE(u.eposta, ''),
			COALESCE(u.telefon, '')
		FROM uye u
		LEFT JOIN uye_detay d ON d.uye_id = u.uye_id
		JOIN sistem_rol sr ON sr.uye_id = u.uye_id
		JOIN sistem_rol_tanimlama srt ON sr.sistem_rol_id = srt.rol_id
		WHERE u.aktif_mi = true AND srt.rol_adi = 'tto'
		ORDER BY 1
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("TTO alıcıları sorgulanamadı: %w", err)
	}
	defer rows.Close()

	var list []models.BildirimAliciKisi
	for rows.Next() {
		var item models.BildirimAliciKisi
		item.HedefKey = models.AliciHedefTTO
		if err := rows.Scan(&item.AdSoyad, &item.Eposta, &item.Telefon); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

// SaveLog gönderim logunu kaydeder.
func (r *ZamanlanmisGorevRepository) SaveLog(log *models.ZamanlanmisGorevLog) error {
	query := `
		INSERT INTO zamanlanmis_gorev_log (
			kural_id, proje_id, kanal, alici, icerik, durum, hata_mesaji
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING log_id, gonderim_tarihi
	`
	return r.DB.QueryRow(
		query,
		log.KuralID, log.ProjeID, log.Kanal, log.Alici, log.Icerik, log.Durum, log.HataMesaji,
	).Scan(&log.LogID, &log.GonderimTarihi)
}

// GetLogs son gönderim loglarını getirir.
func (r *ZamanlanmisGorevRepository) GetLogs(limit int) ([]*models.ZamanlanmisGorevLog, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT 
			l.log_id, l.kural_id, COALESCE(k.kural_adi, 'Silinmiş Kural'),
			l.proje_id, COALESCE(p.proje_kodu, 'PROJE-' || l.proje_id),
			l.kanal, l.alici, l.icerik, l.durum, COALESCE(l.hata_mesaji, ''), l.gonderim_tarihi
		FROM zamanlanmis_gorev_log l
		LEFT JOIN zamanlanmis_gorev_kural k ON k.kural_id = l.kural_id
		LEFT JOIN proje p ON p.proje_id = l.proje_id
		ORDER BY l.log_id DESC
		LIMIT $1
	`
	rows, err := r.DB.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("loglar sorgulanamadı: %w", err)
	}
	defer rows.Close()

	var loglar []*models.ZamanlanmisGorevLog
	for rows.Next() {
		var lg models.ZamanlanmisGorevLog
		err := rows.Scan(
			&lg.LogID, &lg.KuralID, &lg.KuralAdi,
			&lg.ProjeID, &lg.ProjeKodu,
			&lg.Kanal, &lg.Alici, &lg.Icerik, &lg.Durum, &lg.HataMesaji, &lg.GonderimTarihi,
		)
		if err != nil {
			return nil, err
		}
		loglar = append(loglar, &lg)
	}
	return loglar, nil
}
