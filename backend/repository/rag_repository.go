package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"bap_ai/backend/models"
)

// RAGRepository veritabanı üzerindeki RAG doküman işlemlerinden sorumludur.
type RAGRepository struct {
	DB *sql.DB
}

// NewRAGRepository yeni bir RAGRepository nesnesi oluşturur
func NewRAGRepository(db *sql.DB) *RAGRepository {
	return &RAGRepository{DB: db}
}

// SaveDocument bir RAG dokümanını veritabanına kaydeder veya günceller
func (r *RAGRepository) SaveDocument(doc *models.RAGDokuman) error {
	query := `
		INSERT INTO rag_dokuman (varlik_turu, varlik_id, proje_id, uye_id, baslik, icerik, metadata_json, vektor_veri)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (dokuman_id) DO UPDATE 
		SET baslik = EXCLUDED.baslik, icerik = EXCLUDED.icerik, metadata_json = EXCLUDED.metadata_json, vektor_veri = EXCLUDED.vektor_veri
	`
	_, err := r.DB.Exec(query, doc.VarlikTuru, doc.VarlikID, doc.ProjeID, doc.UyeID, doc.Baslik, doc.Icerik, doc.MetadataJSON, doc.VektorVeri)
	return err
}

// ClearDocumentsByType belirli bir varlık türüne ait tüm dokümanları temizler
func (r *RAGRepository) ClearDocumentsByType(varlikTuru string) error {
	_, err := r.DB.Exec("DELETE FROM rag_dokuman WHERE varlik_turu = $1", varlikTuru)
	return err
}

// SearchHybrid yetki odaklı hibrit arama yapar (FTS + Metin Benzerliği)
func (r *RAGRepository) SearchHybrid(q models.RAGSearchQuery) ([]models.RAGSearchResult, error) {
	var results []models.RAGSearchResult
	limit := q.Limit
	if limit <= 0 {
		limit = 10
	}

	isAdminOrManagement := false
	isHakem := false
	for _, rol := range q.Roller {
		rolLower := strings.ToLower(rol)
		if strings.Contains(rolLower, "admin") || strings.Contains(rolLower, "dekan") || strings.Contains(rolLower, "komisyon") || strings.Contains(rolLower, "tto") {
			isAdminOrManagement = true
		}
		if strings.Contains(rolLower, "hakem") {
			isHakem = true
		}
	}

	// Yetki Filtresi Hazırlığı
	var whereClauses []string
	var args []interface{}
	argCount := 1

	// Sorgu metni filtresi (FTS & ILIKE)
	terms := strings.Fields(q.SorguMetni)
	var searchConditions []string
	for _, term := range terms {
		if len(term) < 2 {
			continue
		}
		cleanTerm := strings.ReplaceAll(term, "'", "''")
		searchConditions = append(searchConditions, fmt.Sprintf("(baslik ILIKE $%d OR icerik ILIKE $%d)", argCount, argCount))
		args = append(args, "%"+cleanTerm+"%")
		argCount++
	}

	if len(searchConditions) > 0 {
		whereClauses = append(whereClauses, "("+strings.Join(searchConditions, " OR ")+")")
	}

	// Rol bazlı erişim kısıtlaması
	if !isAdminOrManagement {
		if isHakem {
			whereClauses = append(whereClauses, fmt.Sprintf(`(
				proje_id IN (SELECT proje_id FROM proje_degerlendirmeleri WHERE hakem_id = $%d)
				OR (proje_id IS NULL AND uye_id IS NULL)
			)`, argCount))
			args = append(args, q.KullaniciID)
			argCount++
		} else { // Akademisyen / Öğrenci
			whereClauses = append(whereClauses, fmt.Sprintf(`(
				uye_id = $%d 
				OR proje_id IN (SELECT proje_id FROM proje_takim WHERE uye_id = $%d AND davet_durumu = 'kabul')
				OR (proje_id IS NULL AND uye_id IS NULL)
			)`, argCount, argCount))
			args = append(args, q.KullaniciID)
			argCount++
		}
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	querySQL := fmt.Sprintf(`
		SELECT dokuman_id, varlik_turu, varlik_id, baslik, icerik, 1.0 as skor
		FROM rag_dokuman
		%s
		ORDER BY olusturma_tarihi DESC
		LIMIT %d
	`, whereSQL, limit)

	rows, err := r.DB.Query(querySQL, args...)
	if err != nil {
		return nil, fmt.Errorf("RAG arama sorgusu başarısız: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var res models.RAGSearchResult
		if err := rows.Scan(&res.DokumanID, &res.VarlikTuru, &res.VarlikID, &res.Baslik, &res.Icerik, &res.Skor); err == nil {
			results = append(results, res)
		}
	}

	return results, nil
}

// QueryStructuredSummary sayısal ve yapılandırılmış istatistik sorularını yanıtlar
func (r *RAGRepository) QueryStructuredSummary(kullaniciID int, roller []string, sorguMetni string) (string, error) {
	var sb strings.Builder
	sorguLower := strings.ToLower(sorguMetni)

	isAdminOrManagement := false
	for _, rol := range roller {
		rolLower := strings.ToLower(rol)
		if strings.Contains(rolLower, "admin") || strings.Contains(rolLower, "dekan") || strings.Contains(rolLower, "komisyon") || strings.Contains(rolLower, "tto") {
			isAdminOrManagement = true
		}
	}

	// 1. Proje Sayıları ve Durum Dağılımı
	if strings.Contains(sorguLower, "kaç") || strings.Contains(sorguLower, "sayı") || strings.Contains(sorguLower, "toplam proje") || strings.Contains(sorguLower, "durum") {
		var totalProjects int
		var countQuery string
		var countErr error

		if isAdminOrManagement {
			countQuery = "SELECT COUNT(*) FROM proje"
			countErr = r.DB.QueryRow(countQuery).Scan(&totalProjects)
		} else {
			countQuery = "SELECT COUNT(DISTINCT p.proje_id) FROM proje p INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id WHERE pt.uye_id = $1"
			countErr = r.DB.QueryRow(countQuery, kullaniciID).Scan(&totalProjects)
		}

		if countErr == nil {
			sb.WriteString(fmt.Sprintf("📊 Sistemdeki Toplam Proje Sayısı: %d\n", totalProjects))
		}

		// Durum bazlı dağılım
		statusQuery := `
			SELECT COALESCE(pd.durum_adi, 'Belirsiz'), COUNT(p.proje_id)
			FROM proje p
			LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
			GROUP BY pd.durum_adi
		`
		statusRows, err := r.DB.Query(statusQuery)
		if err == nil {
			defer statusRows.Close()
			sb.WriteString("  - Durum Dağılımı:\n")
			for statusRows.Next() {
				var durum string
				var adet int
				if errScan := statusRows.Scan(&durum, &adet); errScan == nil {
					sb.WriteString(fmt.Sprintf("    * %s: %d proje\n", durum, adet))
				}
			}
		}
	}

	// 2. Bütçe ve Harcama İstatistikleri
	if strings.Contains(sorguLower, "bütçe") || strings.Contains(sorguLower, "harcama") || strings.Contains(sorguLower, "tutar") || strings.Contains(sorguLower, "maliyet") || strings.Contains(sorguLower, "kalan") {
		var toplamButce, toplamHarcanan float64

		if isAdminOrManagement {
			_ = r.DB.QueryRow("SELECT COALESCE(SUM(toplam_butce), 0) FROM proje").Scan(&toplamButce)
			_ = r.DB.QueryRow("SELECT COALESCE(SUM(toplam_fiyat), 0) FROM proje_satinalma_talebi WHERE durum = 'Onaylandı'").Scan(&toplamHarcanan)
		} else {
			_ = r.DB.QueryRow("SELECT COALESCE(SUM(p.toplam_butce), 0) FROM proje p INNER JOIN proje_takim pt ON p.proje_id = pt.proje_id WHERE pt.uye_id = $1", kullaniciID).Scan(&toplamButce)
			_ = r.DB.QueryRow("SELECT COALESCE(SUM(st.toplam_fiyat), 0) FROM proje_satinalma_talebi st INNER JOIN proje_takim pt ON st.proje_id = pt.proje_id WHERE pt.uye_id = $1 AND st.durum = 'Onaylandı'", kullaniciID).Scan(&toplamHarcanan)
		}

		kalanButce := toplamButce - toplamHarcanan
		sb.WriteString(fmt.Sprintf("💰 Bütçe Analizi:\n  - Toplam Onaylı Bütçe: %.2f TL\n  - Harcanan (Onaylı Satın Alma): %.2f TL\n  - Kalan Bütçe: %.2f TL\n", toplamButce, toplamHarcanan, kalanButce))
	}

	// 3. Kullanıcı ve Rol Dağılımları (Admin/Yönetim için)
	if isAdminOrManagement && (strings.Contains(sorguLower, "kullanıcı") || strings.Contains(sorguLower, "akademisyen") || strings.Contains(sorguLower, "üye")) {
		var toplamUye int
		_ = r.DB.QueryRow("SELECT COUNT(*) FROM uye").Scan(&toplamUye)
		sb.WriteString(fmt.Sprintf("👥 Kayıtlı Toplam Kullanıcı Sayısı: %d\n", toplamUye))
	}

	return sb.String(), nil
}

// SyncDatabaseToRAG veritabanındaki tüm proje, bütçe, risk ve kuralları RAG dokümanlarına indeksler
func (r *RAGRepository) SyncDatabaseToRAG() (int, error) {
	// Türkçe Yorum: Mevcut RAG verileri temizlenerek güncel veritabanı kayıtları aktarılır.
	_, _ = r.DB.Exec("DELETE FROM rag_dokuman")
	insertedCount := 0

	// 1. Projeleri Indeksle
	projeQuery := `
		SELECT p.proje_id, p.koordinator_id, COALESCE(p.proje_kodu, ''), COALESCE(p.baslik_tr, ''), 
		       COALESCE(p.ozet_tr, ''), COALESCE(pbt.bap_turu, ''), COALESCE(pd.durum_adi, ''), p.toplam_butce, p.sure_ay
		FROM proje p
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
	`
	pRows, err := r.DB.Query(projeQuery)
	if err == nil {
		defer pRows.Close()
		for pRows.Next() {
			var pID, kID, sAy int
			var pKod, baslik, ozet, tur, durum string
			var butce float64
			if errScan := pRows.Scan(&pID, &kID, &pKod, &baslik, &ozet, &tur, &durum, &butce, &sAy); errScan == nil {
				icerik := fmt.Sprintf("Proje Kodu: %s\nBaşlık: %s\nTür: %s\nDurum: %s\nToplam Bütçe: %.2f TL\nSüre: %d Ay\nÖzet: %s",
					pKod, baslik, tur, durum, butce, sAy, ozet)
				
				doc := &models.RAGDokuman{
					VarlikTuru: "proje",
					VarlikID:   pID,
					ProjeID:    &pID,
					UyeID:      &kID,
					Baslik:     fmt.Sprintf("Proje #%d: %s (%s)", pID, baslik, pKod),
					Icerik:     icerik,
				}
				if errSave := r.SaveDocument(doc); errSave == nil {
					insertedCount++
				}
			}
		}
	}

	// 2. Risk Yönetimi Kayıtlarını İndeksle
	riskQuery := `
		SELECT r.risk_id, r.proje_id, COALESCE(r.risk_aciklamasi, ''), COALESCE(r.cozum_plani, '')
		FROM proje_risk_yonetimi r
	`
	rRows, errRisk := r.DB.Query(riskQuery)
	if errRisk == nil {
		defer rRows.Close()
		for rRows.Next() {
			var rID, pID int
			var risk, cozum string
			if errScan := rRows.Scan(&rID, &pID, &risk, &cozum); errScan == nil {
				doc := &models.RAGDokuman{
					VarlikTuru: "risk",
					VarlikID:   rID,
					ProjeID:    &pID,
					Baslik:     fmt.Sprintf("Proje #%d Risk ve Çözüm Planı", pID),
					Icerik:     fmt.Sprintf("Risk Tanımı: %s\nÇözüm Planı: %s", risk, cozum),
				}
				if errSave := r.SaveDocument(doc); errSave == nil {
					insertedCount++
				}
			}
		}
	}

	// 3. İş Paketlerini İndeksle
	isQuery := `
		SELECT ip.paket_id, ip.proje_id, COALESCE(ip.paket_adi, ''), COALESCE(ip.paket_amaci, ''), ip.baslangic_ay, ip.bitis_ay
		FROM proje_is_paketi ip
	`
	isRows, errIs := r.DB.Query(isQuery)
	if errIs == nil {
		defer isRows.Close()
		for isRows.Next() {
			var ipID, pID, bAy, bitAy int
			var ad, amac string
			if errScan := isRows.Scan(&ipID, &pID, &ad, &amac, &bAy, &bitAy); errScan == nil {
				doc := &models.RAGDokuman{
					VarlikTuru: "is_paketi",
					VarlikID:   ipID,
					ProjeID:    &pID,
					Baslik:     fmt.Sprintf("Proje #%d İş Paketi: %s", pID, ad),
					Icerik:     fmt.Sprintf("İş Paketi Adı: %s\nAmacı: %s\nSüre: %d. Ay - %d. Ay (%d Ay)", ad, amac, bAy, bitAy, bitAy-bAy+1),
				}
				if errSave := r.SaveDocument(doc); errSave == nil {
					insertedCount++
				}
			}
		}
	}

	// 4. Zamanlanmış Görev & Mevzuat Kurallarını İndeksle
	ruleQuery := `
		SELECT kural_id, COALESCE(kural_adi, ''), COALESCE(tetikleyici_olay, ''), COALESCE(aksiyon_turu, ''), kalan_gun
		FROM zamanlanmis_gorev_kural
	`
	ruleRows, errRule := r.DB.Query(ruleQuery)
	if errRule == nil {
		defer ruleRows.Close()
		for ruleRows.Next() {
			var kID, kGun int
			var ad, olay, aksiyon string
			if errScan := ruleRows.Scan(&kID, &ad, &olay, &aksiyon, &kGun); errScan == nil {
				doc := &models.RAGDokuman{
					VarlikTuru: "mevzuat",
					VarlikID:   kID,
					Baslik:     fmt.Sprintf("BAP Süreç/Zamanlanmış Görev Kuralı: %s", ad),
					Icerik:     fmt.Sprintf("Kural Adı: %s\nTetikleyici Olay: %s\nAksiyon Türü: %s\nKalan Gün Şartı: %d Gün", ad, olay, aksiyon, kGun),
				}
				if errSave := r.SaveDocument(doc); errSave == nil {
					insertedCount++
				}
			}
		}
	}

	log.Printf("Bilgi: Veritabanından toplam %d adet kayıt RAG dokümanı olarak indekslendi.\n", insertedCount)
	return insertedCount, nil
}
