package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	dsn := "host=127.0.0.1 port=5432 user=bap password=bap_admin_1 dbname=bap_app sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("sql.Open error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("DB Ping error: %v", err)
	}
	fmt.Println("=== DB Connection OK ===")

	// 1. Check all projects in DB
	rowsAll, err := db.Query(`
		SELECT p.proje_id, COALESCE(p.proje_kodu, 'yok'), p.durum_id, COALESCE(pd.durum_adi, 'NULL'), COALESCE(pd.durum_etiketi, 'NULL')
		FROM proje p
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
	`)
	if err != nil {
		fmt.Printf("Query all projects error: %v\n", err)
	} else {
		fmt.Println("--- All Projects in DB ---")
		for rowsAll.Next() {
			var id int
			var kodu, dAdi, dEtiketi string
			var dID sql.NullInt64
			rowsAll.Scan(&id, &kodu, &dID, &dAdi, &dEtiketi)
			fmt.Printf("ProjeID: %d | Kodu: %s | DurumID: %v | DurumAdi: '%s' | DurumEtiketi: '%s'\n", id, kodu, dID.Int64, dAdi, dEtiketi)
		}
		rowsAll.Close()
	}

	// 2. Test GetProjectsForWorkflow Query
	query2 := `
		SELECT p.proje_id, COALESCE(p.proje_kodu, ''),
		       COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'tr'), '') AS baslik_tr,
		       COALESCE((SELECT baslik FROM proje_baslik WHERE proje_id = p.proje_id AND dil_kodu = 'en'), '') AS baslik_en, 
		       COALESCE(p.sure_ay, 0)::INTEGER AS sure_ay,
		       (COALESCE((SELECT SUM(toplam_fiyat) FROM proje_butce WHERE proje_id = p.proje_id), p.toplam_butce, 0) + COALESCE((SELECT SUM(tutar_tl) FROM proje_talep_ek_butce WHERE proje_id = p.proje_id AND durum = 'onaylandi'), 0))::FLOAT AS toplam_butce,
		       EXISTS(SELECT 1 FROM proje_etik_kurul WHERE proje_id = p.proje_id) AS etik_kurul,
		       (SELECT CAST(NULLIF(REGEXP_REPLACE(kurul_karar_no, '\D', '', 'g'), '') AS INTEGER) FROM proje_etik_kurul WHERE proje_id = p.proje_id LIMIT 1) AS etik_kurul_no,
		       p.koordinator_id, p.durum_id, p.asama_id, p.bap_turu_id,
		       p.olusturma_tarihi, p.guncelleme_tarihi,
		       COALESCE(pd.durum_adi, 'taslak'), COALESCE(pbt.bap_turu, ''),
		       COALESCE(pa.asama_adi, ''), COALESCE(pa.asama_kodu, ''),
		       COALESCE(u.ad || ' ' || u.soyad, '') as koordinator_ad_soyad,
		       COALESCE(u.unvan, '') as koordinator_unvan,
		       COALESCE(pbv.hakem_gerekli, COALESCE(pbt.hakem_gerekli, false)) as hakem_gerekli,
		       COALESCE(TO_CHAR(ps.baslangic_tarihi, 'DD.MM.YYYY'), '') AS baslangic_tarihi,
		       COALESCE(TO_CHAR(ps.bitis_tarihi, 'DD.MM.YYYY'), '') AS bitis_tarihi,
		       COALESCE((SELECT SUM(ek_sure_ay) FROM proje_talep_ek_sure WHERE proje_id = p.proje_id AND durum = 'onaylandi'), 0)::INTEGER AS ek_sure_toplam_ay,
		       COALESCE((SELECT SUM(tutar_tl) FROM proje_talep_ek_butce WHERE proje_id = p.proje_id AND durum = 'onaylandi'), 0)::FLOAT AS ek_butce_toplam_tutar
		FROM proje p
		LEFT JOIN proje_durum pd ON p.durum_id = pd.durum_id
		LEFT JOIN proje_asama pa ON p.asama_id = pa.asama_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		LEFT JOIN proje_bap_turu_versiyon pbv ON p.bap_turu_versiyon_id = pbv.versiyon_id
		LEFT JOIN uye u ON p.koordinator_id = u.uye_id
		LEFT JOIN proje_sozlesme ps ON p.proje_id = ps.proje_id
		WHERE ($1 = '' OR pa.asama_kodu = $1 OR pd.durum_adi = $2)
		  AND (pd.durum_adi IS NULL OR pd.durum_adi != 'taslak')
		  AND ($3 = 0 OR NOT EXISTS (
		      SELECT 1 FROM proje_komisyon_onay pko 
		      WHERE pko.proje_id = p.proje_id AND pko.komisyon_uye_id = $4 AND pko.karar != 'bekliyor'
		  ))
		ORDER BY p.guncelleme_tarihi DESC
	`

	rows2, err := db.Query(query2, "", "", 0, 0)
	if err != nil {
		fmt.Printf("GetProjectsForWorkflow Query ERROR: %v\n", err)
		return
	}
	defer rows2.Close()

	fmt.Println("--- GetProjectsForWorkflow Results ---")
	count := 0
	for rows2.Next() {
		count++
		var id int
		var kodu, baslik, dAdi string
		var sure, ekSure int
		var btc, ekBtc float64
		var etik bool
		var eNo, kId, dId, aId, bId *int
		var oTar, gTar interface{}
		var bTuru, aAdi, aKodu, kAd, kUnvan string
		var hGer bool
		var bTar, bitTar string
		err := rows2.Scan(
			&id, &kodu, &baslik, &baslik, &sure, &btc, &etik,
			&eNo, &kId, &dId, &aId, &bId,
			&oTar, &gTar,
			&dAdi, &bTuru,
			&aAdi, &aKodu,
			&kAd, &kUnvan, &hGer,
			&bTar, &bitTar, &ekSure, &ekBtc,
		)
		if err != nil {
			fmt.Printf("Scan ERROR on row %d: %v\n", count, err)
		} else {
			fmt.Printf("[%d] ID: %d | Kodu: %s | Durum: '%s' | Baslik: %s\n", count, id, kodu, dAdi, baslik)
		}
	}
	fmt.Printf("Total workflow projects returned: %d\n", count)
}
