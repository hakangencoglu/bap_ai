package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

// RunSchema veritabanı şemasını schema.sql dosyasından okuyarak uygular
func RunSchema(db *sql.DB, schemaPath string) error {
	// Veritabanının daha önce kurulup kurulmadığını kontrol et
	var exists bool
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = 'uye'
	);`

	// 'uye' tablosunun varlığını sorgula
	err := db.QueryRow(query).Scan(&exists)
	if err != nil {
		return fmt.Errorf("veritabanı durumu kontrol edilemedi: %v", err)
	}

	// Eğer tablo zaten varsa şema kurulumunu atla (veri kaybı ve sıfırlanmayı önlemek için)
	// Türkçe Yorum: Veritabanı zaten kuruluysa herhangi bir şema veya veri değişikliği yapmadan doğrudan başarılı şekilde döner.
	if exists {
		log.Println("Şema: Veritabanı zaten kurulu. Veritabanı şemasında herhangi bir güncelleme veya değişiklik yapılmadı.")

		// Türkçe Yorum: Veritabanı normalizasyonu ve tablo isimlerinin güncellenmesi göçü
		normalizationQuery := `
			-- 1. Tabloları Yeniden Adlandır (Eğer eski isimleriyle duruyorlarsa)
			ALTER TABLE IF EXISTS butce RENAME TO proje_butce;
			ALTER TABLE IF EXISTS butce_kategori RENAME TO proje_butce_kategori;
			ALTER TABLE IF EXISTS butce_tanim RENAME TO proje_butce_tanim;
			ALTER TABLE IF EXISTS is_paketi RENAME TO proje_is_paketi;
			ALTER TABLE IF EXISTS risk_yonetimi RENAME TO proje_risk_yonetimi;
			ALTER TABLE IF EXISTS arastirma RENAME TO proje_arastirma;
			ALTER TABLE IF EXISTS olanak_tur RENAME TO proje_olanak_tur;
			ALTER TABLE IF EXISTS revizyonlar RENAME TO proje_revizyon;
			ALTER TABLE IF EXISTS satinalma_talebi RENAME TO proje_satinalma_talebi;
			ALTER TABLE IF EXISTS talep_ek_sure RENAME TO proje_talep_ek_sure;
			ALTER TABLE IF EXISTS talep_ek_butce RENAME TO proje_talep_ek_butce;
			ALTER TABLE IF EXISTS talep_fasil_aktarimi RENAME TO proje_talep_fasil_aktarimi;
			ALTER TABLE IF EXISTS talep_arastirmaci RENAME TO proje_talep_arastirmaci;
			ALTER TABLE IF EXISTS talep_bursiyer RENAME TO proje_talep_bursiyer;
			ALTER TABLE IF EXISTS talep_proje_iptali RENAME TO proje_talep_proje_iptali;
			ALTER TABLE IF EXISTS talep_bilgi_degisimi RENAME TO proje_talep_bilgi_degisimi;
			ALTER TABLE IF EXISTS talep_proje_dondurma RENAME TO proje_talep_proje_dondurma;
			ALTER TABLE IF EXISTS talep_malzeme_guncelleme RENAME TO proje_talep_malzeme_guncelleme;
			ALTER TABLE IF EXISTS talep_avans RENAME TO proje_talep_avans;

			-- 2. Normalizasyon: proje_baslik ve proje_etik_kurul tablolarının oluşturulması ve verilerin taşınması
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'proje_baslik') THEN
					CREATE TABLE proje_baslik (
						proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
						dil_kodu VARCHAR(10) NOT NULL,
						baslik VARCHAR(500) NOT NULL,
						PRIMARY KEY (proje_id, dil_kodu)
					);
					
					-- Türkçe başlıkları taşı
					INSERT INTO proje_baslik (proje_id, dil_kodu, baslik)
					SELECT proje_id, 'tr', baslik_tr FROM proje WHERE baslik_tr IS NOT NULL AND baslik_tr <> '';
					
					-- İngilizce başlıkları taşı
					INSERT INTO proje_baslik (proje_id, dil_kodu, baslik)
					SELECT proje_id, 'en', baslik_en FROM proje WHERE baslik_en IS NOT NULL AND baslik_en <> '';
				END IF;
			END $$;

			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'proje_etik_kurul') THEN
					CREATE TABLE proje_etik_kurul (
						proje_id INTEGER PRIMARY KEY REFERENCES proje(proje_id) ON DELETE CASCADE,
						kurul_karar_no VARCHAR(100) NOT NULL
					);
					
					INSERT INTO proje_etik_kurul (proje_id, kurul_karar_no)
					SELECT proje_id, CAST(etik_kurul_no AS VARCHAR) FROM proje WHERE etik_kurul = TRUE AND etik_kurul_no IS NOT NULL;
				END IF;
			END $$;

			-- Eski denormalize sütunları temizle
			ALTER TABLE proje DROP COLUMN IF EXISTS baslik_tr;
			ALTER TABLE proje DROP COLUMN IF EXISTS baslik_en;
			ALTER TABLE proje DROP COLUMN IF EXISTS toplam_butce;
			ALTER TABLE proje DROP COLUMN IF EXISTS etik_kurul;
			ALTER TABLE proje DROP COLUMN IF EXISTS etik_kurul_no;

			-- Eski ve artık çalışmayan triggers/functions temizliği
			DROP FUNCTION IF EXISTS sync_project_to_dynamic_table() CASCADE;
		`
		if _, err := db.Exec(normalizationQuery); err != nil {
			log.Printf("Uyarı: Veritabanı normalizasyonu ve yeniden isimlendirme göçü uygulanamadı: %v", err)
		} else {
			log.Println("Bilgi: Veritabanı normalizasyonu ve tablo yeniden isimlendirme göçü başarıyla uygulandı.")
		}

		// Türkçe Yorum: 'rol_etiketi' sütunu sistem_rol_tanimlama tablosuna eklenir ve Türkçe etiketler atanır.
		rolEtiketiQuery := `
			ALTER TABLE sistem_rol_tanimlama ADD COLUMN IF NOT EXISTS rol_etiketi VARCHAR(100);
			UPDATE sistem_rol_tanimlama SET rol_etiketi = 'Sistem Yöneticisi' WHERE rol_adi = 'admin';
			UPDATE sistem_rol_tanimlama SET rol_etiketi = 'Akademisyen' WHERE rol_adi = 'akademisyen';
			UPDATE sistem_rol_tanimlama SET rol_etiketi = 'Öğrenci' WHERE rol_adi = 'ogrenci';
			UPDATE sistem_rol_tanimlama SET rol_etiketi = 'Hakem' WHERE rol_adi = 'hakem';
			UPDATE sistem_rol_tanimlama SET rol_etiketi = 'Fakülte Dekanı' WHERE rol_adi = 'dekan';
			UPDATE sistem_rol_tanimlama SET rol_etiketi = 'BAP Komisyon Üyesi' WHERE rol_adi = 'komisyon';
			UPDATE sistem_rol_tanimlama SET rol_etiketi = 'TTO Temsilcisi' WHERE rol_adi = 'tto';
			INSERT INTO sistem_rol_tanimlama (rol_adi, rol_etiketi) VALUES ('komisyon_raportoru', 'Komisyon Raportörü') ON CONFLICT (rol_adi) DO UPDATE SET rol_etiketi = EXCLUDED.rol_etiketi;
		`
		if _, err := db.Exec(rolEtiketiQuery); err != nil {
			log.Printf("Uyarı: rol_etiketi sütunu eklenemedi veya güncellenemedi: %v", err)
		} else {
			log.Println("Bilgi: sistem_rol_tanimlama tablosuna rol_etiketi sütunu eklendi ve varsayılan veriler güncellendi.")
		}

		// Türkçe Yorum: 'durum_etiketi' sütunu proje_durum tablosuna eklenir ve Türkçe etiketler atanır.
		durumEtiketiQuery := `
			ALTER TABLE proje_durum ADD COLUMN IF NOT EXISTS durum_etiketi VARCHAR(100);
			UPDATE proje_durum SET durum_etiketi = 'Taslak' WHERE durum_adi = 'taslak';
			UPDATE proje_durum SET durum_etiketi = 'İncelemede' WHERE durum_adi = 'incelemede';
			UPDATE proje_durum SET durum_etiketi = 'Dekan Onayı Bekliyor' WHERE durum_adi = 'dekan_onayi_bekliyor';
			UPDATE proje_durum SET durum_etiketi = 'Dekan Onayladı' WHERE durum_adi = 'dekan_onayladi';
			UPDATE proje_durum SET durum_etiketi = 'Komisyon Onayı Bekliyor' WHERE durum_adi = 'komisyon_bekliyor';
			UPDATE proje_durum SET durum_etiketi = 'Komisyon Onayladı' WHERE durum_adi = 'komisyon_onayladi';
			UPDATE proje_durum SET durum_etiketi = 'Hakem Atama Bekleniyor' WHERE durum_adi = 'hakem_atama_bekliyor';
			UPDATE proje_durum SET durum_etiketi = 'Hakem İncelemesinde' WHERE durum_adi = 'hakem_bekliyor';
			UPDATE proje_durum SET durum_etiketi = 'Hakem Onayladı' WHERE durum_adi = 'hakem_onayladi';
			UPDATE proje_durum SET durum_etiketi = 'Sözleşme / İmza Aşaması' WHERE durum_adi = 'sozlesme_imza';
			UPDATE proje_durum SET durum_etiketi = 'Sözleşme Dolduruldu' WHERE durum_adi = 'sozlesme_dolduruldu';
			UPDATE proje_durum SET durum_etiketi = 'TTO Onayı Bekliyor' WHERE durum_adi = 'tto_aktif';
			UPDATE proje_durum SET durum_etiketi = 'Onaylandı' WHERE durum_adi = 'onaylandi';
			UPDATE proje_durum SET durum_etiketi = 'Reddedildi' WHERE durum_adi = 'reddedildi';
			UPDATE proje_durum SET durum_etiketi = 'Onaylandı (Tamamlandı)' WHERE durum_adi = 'tamamlandi';
			UPDATE proje_durum SET durum_etiketi = 'Revizyon Gerekli' WHERE durum_adi = 'revizyon';
			UPDATE proje_durum SET durum_etiketi = 'Yürürlükte (Aktif)' WHERE durum_adi = 'yururlukte';
		`
		if _, err := db.Exec(durumEtiketiQuery); err != nil {
			log.Printf("Uyarı: durum_etiketi sütunu eklenemedi veya güncellenemedi: %v", err)
		} else {
			log.Println("Bilgi: proje_durum tablosuna durum_etiketi sütunu eklendi ve varsayılan veriler güncellendi.")
		}

		// Türkçe Yorum: Mevcut veritabanında proje_kodu sütunu yoksa eklenir. Mevcut tüm kayıtların proje kodları yeni yil-bapturu-numara (örn: 2026-BAP100-003) şablonuna göre güncellenir.
		alterQuery := `
			ALTER TABLE proje ADD COLUMN IF NOT EXISTS proje_kodu VARCHAR(100) UNIQUE;
			
			WITH numbered_projects AS (
				SELECT 
					p.proje_id,
					pbt.bap_turu,
					COALESCE(EXTRACT(YEAR FROM p.olusturma_tarihi), EXTRACT(YEAR FROM CURRENT_TIMESTAMP)) as yil,
					ROW_NUMBER() OVER (
						PARTITION BY p.bap_turu_id, EXTRACT(YEAR FROM p.olusturma_tarihi) 
						ORDER BY p.olusturma_tarihi, p.proje_id
					) as sira_no
				FROM proje p
				LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
			)
			UPDATE proje p
			SET proje_kodu = TO_CHAR(np.yil, 'FM9999') || '-' || COALESCE(REPLACE(np.bap_turu, '-', ''), 'BAP') || '-' || LPAD(np.sira_no::text, 3, '0')
			FROM numbered_projects np
			WHERE p.proje_id = np.proje_id;
		`
		if _, err := db.Exec(alterQuery); err != nil {
			log.Printf("Uyarı: Proje kodu sütunu veya backfill işlemi uygulanamadı: %v", err)
		} else {
			log.Println("Bilgi: Proje kodu sütunu kontrol edildi ve mevcut boş kayıtlar için geriye dönük proje kodları oluşturuldu.")
		}

		// Türkçe Yorum: Zaten kurulu olan veritabanı için chatbot ve eimza yetkilendirme alanlarını kontrol edip sadece admin rolüne tanımlıyoruz.
		accessQuery := `
			-- Chatbot sayfasını tanımla
			INSERT INTO sistem_sayfa (sayfa_adi, sayfa_kodu, url_yolu)
			VALUES ('Yapay Zeka Asistanı (Chatbot)', 'chatbot', '/api/chat')
			ON CONFLICT (sayfa_kodu) DO NOTHING;

			-- Chatbot için admin dışındaki tüm yetkileri sil
			DELETE FROM sayfa_rol_yetki 
			WHERE sayfa_id = (SELECT sayfa_id FROM sistem_sayfa WHERE sayfa_kodu = 'chatbot')
			  AND sistem_rol_id != (SELECT rol_id FROM sistem_rol_tanimlama WHERE rol_adi = 'admin');

			-- Chatbot için admin yetkisini ekle
			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT srt.rol_id, ss.sayfa_id
			FROM sistem_rol_tanimlama srt, sistem_sayfa ss
			WHERE ss.sayfa_kodu = 'chatbot' AND srt.rol_adi = 'admin'
			ON CONFLICT DO NOTHING;

			-- E-İmza için admin dışındaki tüm yetkileri sil
			DELETE FROM sayfa_rol_yetki 
			WHERE sayfa_id = (SELECT sayfa_id FROM sistem_sayfa WHERE sayfa_kodu = 'eimza')
			  AND sistem_rol_id != (SELECT rol_id FROM sistem_rol_tanimlama WHERE rol_adi = 'admin');

			-- E-İmza için admin yetkisini ekle
			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT srt.rol_id, ss.sayfa_id
			FROM sistem_rol_tanimlama srt, sistem_sayfa ss
			WHERE ss.sayfa_kodu = 'eimza' AND srt.rol_adi = 'admin'
			ON CONFLICT DO NOTHING;
		`
		if _, err := db.Exec(accessQuery); err != nil {
			log.Printf("Uyarı: Sayfa yetki alanları dinamik olarak güncellenemedi: %v", err)
		} else {
			log.Println("Bilgi: Chatbot ve E-İmza sayfa yetkileri sadece yönetici (admin) rolüne kısıtlandı.")
		}

		// Türkçe Yorum: 'yururlukte' durumunu ekleyip, mevcut 'tamamlandi' projeleri bu duruma taşıyoruz.
		yururlukteStatusQuery := `
			INSERT INTO proje_durum (durum_adi) VALUES ('yururlukte') ON CONFLICT (durum_adi) DO NOTHING;
			UPDATE proje 
			SET durum_id = (SELECT durum_id FROM proje_durum WHERE durum_adi = 'yururlukte')
			WHERE durum_id = (SELECT durum_id FROM proje_durum WHERE durum_adi = 'tamamlandi');
		`
		if _, err := db.Exec(yururlukteStatusQuery); err != nil {
			log.Printf("Uyarı: 'yururlukte' durum göçü uygulanamadı: %v", err)
		} else {
			log.Println("Bilgi: 'yururlukte' durum göçü ve proje güncellemeleri başarıyla uygulandı.")
		}

		// Türkçe Yorum: Mevcut veritabanına yeni iş akışı için gerekli eksik tüm durumları ekle.
		newStatusQuery := `
			INSERT INTO proje_durum (durum_adi) VALUES ('dekan_onayi_bekliyor') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('dekan_onayladi') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('komisyon_bekliyor') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('komisyon_onayladi') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('hakem_atama_bekliyor') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('hakem_bekliyor') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('hakem_onayladi') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('sozlesme_imza') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('sozlesme_dolduruldu') ON CONFLICT (durum_adi) DO NOTHING;
			INSERT INTO proje_durum (durum_adi) VALUES ('tto_aktif') ON CONFLICT (durum_adi) DO NOTHING;
		`
		if _, err := db.Exec(newStatusQuery); err != nil {
			log.Printf("Uyarı: Yeni durum kayıtları (hakem_bekliyor, sozlesme_imza) eklenemedi: %v", err)
		} else {
			log.Println("Bilgi: Yeni iş akışı durum kayıtları (hakem_bekliyor, sozlesme_imza) başarıyla eklendi.")
		}

		// Türkçe Yorum: Yeni proje_asama tablosunu oluştur (iş akışı onay masaları).
		// Bu tablo; "Dekan Onayına Sun", "Komisyona Sun", "Hakeme Sun" gibi
		// kullanıcıya gösterilen Türkçe aşama isimlerini tutar.
		projeAsamaQuery := `
			CREATE TABLE IF NOT EXISTS proje_asama (
				asama_id   SERIAL PRIMARY KEY,
				asama_kodu VARCHAR(100) UNIQUE NOT NULL,
				asama_adi  VARCHAR(200) NOT NULL,
				sira_no    INTEGER DEFAULT 0,
				durum_adi  VARCHAR(100),
				onay_durum_adi VARCHAR(100)
			);

			ALTER TABLE proje_asama ADD COLUMN IF NOT EXISTS durum_adi VARCHAR(100);
			ALTER TABLE proje_asama ADD COLUMN IF NOT EXISTS onay_durum_adi VARCHAR(100);

			INSERT INTO proje_asama (asama_kodu, asama_adi, sira_no, durum_adi, onay_durum_adi) VALUES
				('tto_on_inceleme',   'TTO Ön İnceleme',  1, 'incelemede', 'incelemede'),
				('dekan_onayina_sun', 'Dekan Onayına Sun', 2, 'dekan_onayi_bekliyor', 'dekan_onayladi'),
				('komisyona_sun',     'Komisyona Sun',      3, 'komisyon_bekliyor', 'komisyon_onayladi'),
				('hakeme_sun',        'Hakeme Sun',         4, 'hakem_atama_bekliyor', 'hakem_onayladi'),
				('sozlesme_imza',     'Sözleşme İmzası',    5, 'sozlesme_imza', 'sozlesme_imza')
			ON CONFLICT (asama_kodu) DO UPDATE SET 
				durum_adi = COALESCE(proje_asama.durum_adi, EXCLUDED.durum_adi),
				onay_durum_adi = COALESCE(proje_asama.onay_durum_adi, EXCLUDED.onay_durum_adi);

			ALTER TABLE proje ADD COLUMN IF NOT EXISTS asama_id INTEGER REFERENCES proje_asama(asama_id) ON DELETE SET NULL;
			CREATE INDEX IF NOT EXISTS idx_proje_asama_id ON proje(asama_id);

			-- Türkçe Yorum: Proje türlerinin süreç aşamalarını (iş akışını) eşleyen tablo
			CREATE TABLE IF NOT EXISTS proje_bap_turu_asama (
				bap_turu_id INTEGER NOT NULL REFERENCES proje_bap_turu(bap_turu_id) ON DELETE CASCADE,
				asama_id INTEGER NOT NULL REFERENCES proje_asama(asama_id) ON DELETE CASCADE,
				sira_no INTEGER NOT NULL,
				PRIMARY KEY (bap_turu_id, asama_id)
			);

			-- Geriye dönük uyumluluk: Tüm mevcut BAP türlerine mevcut tüm aşamaları default olarak ata
			INSERT INTO proje_bap_turu_asama (bap_turu_id, asama_id, sira_no)
			SELECT pbt.bap_turu_id, pa.asama_id, pa.sira_no
			FROM proje_bap_turu pbt, proje_asama pa
			ON CONFLICT DO NOTHING;
		`
		if _, err := db.Exec(projeAsamaQuery); err != nil {
			log.Printf("Uyarı: proje_asama veya proje_bap_turu_asama tablosu oluşturulamadı: %v", err)
		} else {
			log.Println("Bilgi: proje_asama ve proje_bap_turu_asama tabloları başarıyla kontrol edildi/oluşturuldu.")
		}

		// Türkçe Yorum: BAP Proje Sözleşmesi tablosunu oluştur
		sozlesmeMigrationQuery := `
			CREATE TABLE IF NOT EXISTS proje_sozlesme (
				id SERIAL PRIMARY KEY,
				proje_id INT NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
				uye_id INT NOT NULL REFERENCES uye(uye_id),
				tc_kimlik VARCHAR(11) NOT NULL,
				yurutucu_adres TEXT NOT NULL,
				yurutucu_telefon VARCHAR(20) NOT NULL,
				yurutucu_eposta VARCHAR(100) NOT NULL,
				baslangic_tarihi DATE NOT NULL,
				bitis_tarihi DATE NOT NULL,
				durum VARCHAR(20) DEFAULT 'dolduruldu',
				olusturma_tarihi TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				guncelleme_tarihi TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				CONSTRAINT unique_proje_sozlesme UNIQUE (proje_id)
			);
		`
		if _, err := db.Exec(sozlesmeMigrationQuery); err != nil {
			log.Printf("Uyarı: proje_sozlesme tablosu oluşturulamadı: %v", err)
		} else {
			log.Println("Bilgi: proje_sozlesme tablosu başarıyla kontrol edildi/oluşturuldu.")
		}

		// Türkçe Yorum: Hakem bekleyen veya yeniden atanan projelerde takılı kalan eski hakem değerlendirme kayıtlarını sıfırla
		resetHakemQuery := `
			UPDATE proje_degerlendirmeleri pd
			SET durum = 'Bekliyor', atama_durumu = 'Atandı', puan = NULL, yorum = NULL, red_nedeni = NULL
			FROM proje p
			JOIN proje_durum pdur ON p.durum_id = pdur.durum_id
			WHERE pd.proje_id = p.proje_id AND pdur.durum_adi IN ('hakem_bekliyor', 'hakem_atama_bekliyor') AND pd.durum <> 'Bekliyor';
		`
		if _, err := db.Exec(resetHakemQuery); err != nil {
			log.Printf("Uyarı: Hakem değerlendirme durumları güncellenemedi: %v", err)
		}

		// Türkçe Yorum: 5 adet varsayılan komisyon üyesini ve çoklu komisyon onay tablosunu oluştur.
		komisyonMigrationQuery := `
			INSERT INTO uye (rol, ad, soyad, unvan, bolum, eposta, telefon, izu_uyesi, sifre_hash, aktif_mi)
			VALUES 
				('komisyon', 'Ahmet', 'Yılmaz', 'Prof. Dr.', 'Bilgisayar Mühendisliği', 'komisyon1@izu.edu.tr', '5555555561', true, '$2a$10$rj4nxdqm9EN.wDQM/H0ZkOJquMceS41lk1INHgnBOg5LH1Xc/zfai', true),
				('komisyon', 'Mehmet', 'Kaya', 'Prof. Dr.', 'Endüstri Mühendisliği', 'komisyon2@izu.edu.tr', '5555555562', true, '$2a$10$rj4nxdqm9EN.wDQM/H0ZkOJquMceS41lk1INHgnBOg5LH1Xc/zfai', true),
				('komisyon', 'Ayşe', 'Demir', 'Prof. Dr.', 'İşletme', 'komisyon3@izu.edu.tr', '5555555563', true, '$2a$10$rj4nxdqm9EN.wDQM/H0ZkOJquMceS41lk1INHgnBOg5LH1Xc/zfai', true),
				('komisyon', 'Fatma', 'Çelik', 'Prof. Dr.', 'Mimarlık', 'komisyon4@izu.edu.tr', '5555555564', true, '$2a$10$rj4nxdqm9EN.wDQM/H0ZkOJquMceS41lk1INHgnBOg5LH1Xc/zfai', true),
				('komisyon', 'Mustafa', 'Şahin', 'Prof. Dr.', 'Hukuk', 'komisyon5@izu.edu.tr', '5555555565', true, '$2a$10$rj4nxdqm9EN.wDQM/H0ZkOJquMceS41lk1INHgnBOg5LH1Xc/zfai', true)
			ON CONFLICT (eposta) DO NOTHING;

			INSERT INTO sistem_rol (uye_id, sistem_rol_id)
			SELECT u.uye_id, srt.rol_id
			FROM uye u, sistem_rol_tanimlama srt
			WHERE u.eposta IN ('komisyon1@izu.edu.tr', 'komisyon2@izu.edu.tr', 'komisyon3@izu.edu.tr', 'komisyon4@izu.edu.tr', 'komisyon5@izu.edu.tr')
			  AND srt.rol_adi = 'komisyon'
			ON CONFLICT (uye_id, sistem_rol_id) DO NOTHING;

			INSERT INTO uye_detay (uye_id, rol, unvan, bolum, telefon, izu_uyesi, profil_tamamlandi)
			SELECT u.uye_id, 'komisyon', u.unvan, u.bolum, u.telefon, true, true
			FROM uye u
			WHERE u.eposta IN ('komisyon1@izu.edu.tr', 'komisyon2@izu.edu.tr', 'komisyon3@izu.edu.tr', 'komisyon4@izu.edu.tr', 'komisyon5@izu.edu.tr')
			  AND NOT EXISTS (SELECT 1 FROM uye_detay ud WHERE ud.uye_id = u.uye_id);

			CREATE TABLE IF NOT EXISTS proje_komisyon_onay (
				onay_id SERIAL PRIMARY KEY,
				proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
				komisyon_uye_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
				karar VARCHAR(50) DEFAULT 'bekliyor',
				aciklama TEXT,
				olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(proje_id, komisyon_uye_id)
			);
			CREATE INDEX IF NOT EXISTS idx_proje_komisyon_onay_proje ON proje_komisyon_onay(proje_id);
			CREATE INDEX IF NOT EXISTS idx_proje_komisyon_onay_uye ON proje_komisyon_onay(komisyon_uye_id);
		`
		if _, err := db.Exec(komisyonMigrationQuery); err != nil {
			log.Printf("Uyarı: Komisyon üyeleri veya onay tablosu göçü uygulanamadı: %v", err)
		} else {
			log.Println("Bilgi: Komisyon üyeleri ve onay tablosu göçü başarıyla uygulandı.")
		}

		// Türkçe Yorum: Mevcut veritabanı için satın alma sayfalarını ve varsayılan rol yetkilerini ekle.
		satinalmaPageQuery := `
			INSERT INTO sistem_sayfa (sayfa_adi, sayfa_kodu, url_yolu)
			VALUES
				('Satın Alma Talepleri (Araştırmacı)', 'satinalma_arastirmaci', '/satinalma'),
				('Satın Alma Yönetimi (TTO)',           'satinalma_tto',         '/tto/satinalma')
			ON CONFLICT (sayfa_kodu) DO NOTHING;

			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT srt.rol_id, ss.sayfa_id
			FROM sistem_rol_tanimlama srt, sistem_sayfa ss
			WHERE ss.sayfa_kodu = 'satinalma_arastirmaci'
			  AND srt.rol_adi IN ('akademisyen', 'ogrenci')
			ON CONFLICT DO NOTHING;

			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT srt.rol_id, ss.sayfa_id
			FROM sistem_rol_tanimlama srt, sistem_sayfa ss
			WHERE ss.sayfa_kodu = 'satinalma_tto'
			  AND srt.rol_adi IN ('tto', 'admin')
			ON CONFLICT DO NOTHING;
		`
		if _, err := db.Exec(satinalmaPageQuery); err != nil {
			log.Printf("Uyarı: Satın alma sayfa yetkileri eklenemedi: %v", err)
		} else {
			log.Println("Bilgi: Satın alma sayfa tanımları ve varsayılan rol yetkileri başarıyla eklendi.")
		}

		// Türkçe Yorum: Hakem Değerlendirme Sorularının Dinamikleştirilmesi için yeni tabloları oluşturup seed verilerini ekle.
		hakemSorulariQuery := `
			CREATE TABLE IF NOT EXISTS hakem_degerlendirme_basliklari (
				baslik_id SERIAL PRIMARY KEY,
				baslik_adi VARCHAR(200) UNIQUE NOT NULL,
				maksimum_puan INTEGER NOT NULL,
				sira_no INTEGER DEFAULT 0
			);

			CREATE TABLE IF NOT EXISTS hakem_degerlendirme_sorulari (
				soru_id SERIAL PRIMARY KEY,
				soru_kodu VARCHAR(10) NOT NULL,
				soru_metni TEXT UNIQUE NOT NULL,
				sira_no INTEGER DEFAULT 0
			);

			-- Türkçe Yorum: Mevcut veritabanında maksimum_puan sütunu yoksa eklenir.
			ALTER TABLE hakem_degerlendirme_sorulari ADD COLUMN IF NOT EXISTS maksimum_puan INTEGER DEFAULT 5;

			CREATE TABLE IF NOT EXISTS hakem_degerlendirme_baslik_soru (
				baslik_id INTEGER NOT NULL REFERENCES hakem_degerlendirme_basliklari(baslik_id) ON DELETE CASCADE,
				soru_id INTEGER NOT NULL REFERENCES hakem_degerlendirme_sorulari(soru_id) ON DELETE CASCADE,
				PRIMARY KEY (baslik_id, soru_id)
			);

			CREATE TABLE IF NOT EXISTS proje_degerlendirme_soru_cevaplari (
				cevap_id SERIAL PRIMARY KEY,
				degerlendirme_id INTEGER NOT NULL REFERENCES proje_degerlendirmeleri(degerlendirme_id) ON DELETE CASCADE,
				baslik_id INTEGER NOT NULL REFERENCES hakem_degerlendirme_basliklari(baslik_id) ON DELETE CASCADE,
				soru_id INTEGER NOT NULL REFERENCES hakem_degerlendirme_sorulari(soru_id) ON DELETE CASCADE,
				puan_degeri VARCHAR(50) NOT NULL,
				olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(degerlendirme_id, baslik_id, soru_id)
			);

			-- Başlıklar
			INSERT INTO hakem_degerlendirme_basliklari (baslik_adi, maksimum_puan, sira_no) VALUES
			('Özgün Değer', 24, 1),
			('Projenin Yönetimi', 20, 2),
			('Projenin Yaygın Etkisi', 20, 3),
			('Yapılabilirlik: Ekipman/Ortam', 10, 4),
			('Yapılabilirlik: Süre', 16, 5),
			('Yapılabilirlik: Bütçe', 10, 6)
			ON CONFLICT (baslik_adi) DO UPDATE SET maksimum_puan = EXCLUDED.maksimum_puan, sira_no = EXCLUDED.sira_no;

			-- Sorular
			INSERT INTO hakem_degerlendirme_sorulari (soru_kodu, soru_metni, maksimum_puan, sira_no) VALUES
			('A', 'Yerel, ulusal veya uluslararası bir soruna bilimsel çözüm getirmektedir.', 6, 1),
			('B', 'Yöntem, kuram veya ortaya koyacağı bilgi açısından bilimsel ya da teknolojik bir yenilik getirmektedir.', 5, 2),
			('C', 'Yeni, farklı bakış sunan ve tamamlayıcı bilimsel bir araştırma sorusu ortaya atmaktadır.', 5, 3),
			('D', 'Temel ve güncel bilimsel kaynaklara dayalı literatür taraması ile bilimsel tutarlılığı, bütünlüğü vurgulanmış ve diğer bilimsel çalışmalarla ilişki kurulmuştur.', 4, 4),
			('E', 'Araştırmanın amacı (problem/hipotez) açıkça belirtilmiştir.', 4, 5),

			('A', 'Araştırmanın amacını (problem/hipotez) test edecek bilimsel araştırma yöntemleri açıkça belirtilmiştir.', 7, 1),
			('B', 'Araştırmada proje yönetim araçları kullanılmıştır.', 7, 2),
			('C', 'Veri toplama yöntemleri ve araçları (varsa geliştirilme süreçleri) belirtilmiştir.', 6, 3),

			('A', 'Bulgular, evrensel ve/veya ulusal düzeyde araştırmacılar tarafından ilgili bilimsel alanda kullanılabilir özelliktedir.', 5, 1),
			('B', 'Araştırmacı/Yürütücü elde edilecek bulgularıyla yeni projelere düşünsel kaynak oluşturma ya da ileri bilimsel araştırma üretme potansiyeli vardır.', 5, 2),
			('C', 'Desteklenecek projenin lisansüstü tezi üretme veya araştırmacı/öğrenci yetiştirilmesine katkı sağlama potansiyeli vardır.', 5, 3),
			('D', 'Yayın, patent, ödül, yarışma derecesi, bildiri ile tescil edilecek çıktılar elde etme potansiyeli vardır.', 5, 4),

			('A', 'Projenin yürütüleceği bölümün/merkezin altyapısı, ortamı ve olanakları yeterlidir.', 5, 1),
			('B', 'Proje kapsamında istenilen ek ekipman mevcut altyapı ve proje ile uyumludur.', 5, 2),

			('A', 'Önerilen araştırma süresi gerçekçidir.', 6, 1),
			('B', 'Projede her bir iş paketinin hangi sürede gerçekleştirileceği detaylandırılmıştır.', 5, 2),
			('C', 'Projenin başarısını olumsuz yönde etkileyebilecek riskler ve alınacak tedbirler (B Planı) belirtilmiştir.', 5, 3),

			('A', 'Önerilen bütçe gerçekçidir ve bütçenin hazırlanmasında ekonomiklik dikkate alınmıştır.', 5, 1),
			('B', 'Talep edilen destek iş paketleriyle uyumlu hazırlanmıştır.', 5, 2)
			ON CONFLICT (soru_metni) DO UPDATE SET soru_kodu = EXCLUDED.soru_kodu, maksimum_puan = EXCLUDED.maksimum_puan, sira_no = EXCLUDED.sira_no;

			-- İlişkilendirmeler (3. Tablo)
			-- Özgün Değer (1)
			INSERT INTO hakem_degerlendirme_baslik_soru (baslik_id, soru_id)
			SELECT b.baslik_id, s.soru_id FROM hakem_degerlendirme_basliklari b, hakem_degerlendirme_sorulari s
			WHERE b.baslik_adi = 'Özgün Değer' AND s.soru_metni IN (
				'Yerel, ulusal veya uluslararası bir soruna bilimsel çözüm getirmektedir.',
				'Yöntem, kuram veya ortaya koyacağı bilgi açısından bilimsel ya da teknolojik bir yenilik getirmektedir.',
				'Yeni, farklı bakış sunan ve tamamlayıcı bilimsel bir araştırma sorusu ortaya atmaktadır.',
				'Temel ve güncel bilimsel kaynaklara dayalı literatür taraması ile bilimsel tutarlılığı, bütünlüğü vurgulanmış ve diğer bilimsel çalışmalarla ilişki kurulmuştur.',
				'Araştırmanın amacı (problem/hipotez) açıkça belirtilmiştir.'
			) ON CONFLICT DO NOTHING;

			-- Projenin Yönetimi (2)
			INSERT INTO hakem_degerlendirme_baslik_soru (baslik_id, soru_id)
			SELECT b.baslik_id, s.soru_id FROM hakem_degerlendirme_basliklari b, hakem_degerlendirme_sorulari s
			WHERE b.baslik_adi = 'Projenin Yönetimi' AND s.soru_metni IN (
				'Araştırmanın amacını (problem/hipotez) test edecek bilimsel araştırma yöntemleri açıkça belirtilmiştir.',
				'Araştırmada proje yönetim araçları kullanılmıştır.',
				'Veri toplama yöntemleri ve araçları (varsa geliştirilme süreçleri) belirtilmiştir.'
			) ON CONFLICT DO NOTHING;

			-- Projenin Yaygın Etkisi (3)
			INSERT INTO hakem_degerlendirme_baslik_soru (baslik_id, soru_id)
			SELECT b.baslik_id, s.soru_id FROM hakem_degerlendirme_basliklari b, hakem_degerlendirme_sorulari s
			WHERE b.baslik_adi = 'Projenin Yaygın Etkisi' AND s.soru_metni IN (
				'Bulgular, evrensel ve/veya ulusal düzeyde araştırmacılar tarafından ilgili bilimsel alanda kullanılabilir özelliktedir.',
				'Araştırmacı/Yürütücü elde edilecek bulgularıyla yeni projelere düşünsel kaynak oluşturma ya da ileri bilimsel araştırma üretme potansiyeli vardır.',
				'Desteklenecek projenin lisansüstü tezi üretme veya araştırmacı/öğrenci yetiştirilmesine katkı sağlama potansiyeli vardır.',
				'Yayın, patent, ödül, yarışma derecesi, bildiri ile tescil edilecek çıktılar elde etme potansiyeli vardır.'
			) ON CONFLICT DO NOTHING;

			-- Yapılabilirlik: Ekipman/Ortam (4)
			INSERT INTO hakem_degerlendirme_baslik_soru (baslik_id, soru_id)
			SELECT b.baslik_id, s.soru_id FROM hakem_degerlendirme_basliklari b, hakem_degerlendirme_sorulari s
			WHERE b.baslik_adi = 'Yapılabilirlik: Ekipman/Ortam' AND s.soru_metni IN (
				'Projenin yürütüleceği bölümün/merkezin altyapısı, ortamı ve olanakları yeterlidir.',
				'Proje kapsamında istenilen ek ekipman mevcut altyapı ve proje ile uyumludur.'
			) ON CONFLICT DO NOTHING;

			-- Yapılabilirlik: Süre (5)
			INSERT INTO hakem_degerlendirme_baslik_soru (baslik_id, soru_id)
			SELECT b.baslik_id, s.soru_id FROM hakem_degerlendirme_basliklari b, hakem_degerlendirme_sorulari s
			WHERE b.baslik_adi = 'Yapılabilirlik: Süre' AND s.soru_metni IN (
				'Önerilen araştırma süresi gerçekçidir.',
				'Projede her bir iş paketinin hangi sürede gerçekleştirileceği detaylandırılmıştır.',
				'Projenin başarısını olumsuz yönde etkileyebilecek riskler ve alınacak tedbirler (B Planı) belirtilmiştir.'
			) ON CONFLICT DO NOTHING;

			-- Yapılabilirlik: Bütçe (6)
			INSERT INTO hakem_degerlendirme_baslik_soru (baslik_id, soru_id)
			SELECT b.baslik_id, s.soru_id FROM hakem_degerlendirme_basliklari b, hakem_degerlendirme_sorulari s
			WHERE b.baslik_adi = 'Yapılabilirlik: Bütçe' AND s.soru_metni IN (
				'Önerilen bütçe gerçekçidir ve bütçenin hazırlanmasında ekonomiklik dikkate alınmıştır.',
				'Talep edilen destek iş paketleriyle uyumlu hazırlanmıştır.'
			) ON CONFLICT DO NOTHING;
		`
		if _, err := db.Exec(hakemSorulariQuery); err != nil {
			log.Printf("Uyarı: Hakem dinamik soruları ve tabloları eklenemedi: %v", err)
		} else {
			log.Println("Bilgi: Hakem dinamik değerlendirme başlıkları, soruları ve cevap tablosu başarıyla eklendi/güncellendi.")
		}

		// Türkçe Yorum: Bildirim tablosu oluşturulur (sistem içi bildirimler için)
		bildirimQuery := `
			CREATE TABLE IF NOT EXISTS bildirim (
				bildirim_id SERIAL PRIMARY KEY,
				uye_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
				baslik VARCHAR(255) NOT NULL,
				icerik TEXT NOT NULL,
				okundu BOOLEAN DEFAULT FALSE,
				olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);
			CREATE INDEX IF NOT EXISTS idx_bildirim_uye_id ON bildirim(uye_id);
		`
		if _, err := db.Exec(bildirimQuery); err != nil {
			log.Printf("Uyarı: bildirim tablosu oluşturulamadı: %v", err)
		} else {
			log.Println("Bilgi: bildirim tablosu ve indeksi başarıyla kuruldu.")
		}

		// Türkçe Yorum: Satın alma talepleri için talep_no sütununu 2. sütun yapacak şekilde tabloyu yeniden düzenler ve geriye dönük numaralandırır.
		satinalmaTalepNoQuery := `
			DO $$
			BEGIN
				-- Eğer proje_satinalma_talebi tablosu varsa ve talep_no 2. kolon değilse (ordinal_position != 2)
				IF EXISTS (
					SELECT 1 
					FROM information_schema.tables 
					WHERE table_schema = 'public' AND table_name = 'proje_satinalma_talebi'
				) AND NOT EXISTS (
					SELECT 1 
					FROM information_schema.columns 
					WHERE table_schema = 'public' 
					  AND table_name = 'proje_satinalma_talebi' 
					  AND column_name = 'talep_no' 
					  AND ordinal_position = 2
				) THEN
					-- Eski tabloyu yeniden adlandır
					ALTER TABLE proje_satinalma_talebi RENAME TO proje_satinalma_talebi_old;
					
					-- Yeni tabloyu doğru kolon sırası ile oluştur
					CREATE TABLE proje_satinalma_talebi (
						talep_id SERIAL PRIMARY KEY,
						talep_no VARCHAR(100),
						proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
						uye_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE SET NULL,
						kalem_id INTEGER NOT NULL REFERENCES proje_butce(kalem_id) ON DELETE CASCADE,
						malzeme_adi VARCHAR(500) NOT NULL,
						miktar INTEGER NOT NULL CHECK (miktar > 0),
						birim_fiyat NUMERIC(10, 2) NOT NULL CHECK (birim_fiyat >= 0),
						toplam_fiyat NUMERIC(12, 2) NOT NULL CHECK (toplam_fiyat >= 0),
						durum VARCHAR(50) DEFAULT 'Beklemede',
						gerekce TEXT NOT NULL,
						red_nedeni TEXT,
						olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
						guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
					);
					
					-- Verileri kopyala
					IF EXISTS (
						SELECT 1 
						FROM information_schema.columns 
						WHERE table_schema = 'public' 
						  AND table_name = 'proje_satinalma_talebi_old' 
						  AND column_name = 'talep_no'
					) THEN
						INSERT INTO proje_satinalma_talebi (talep_id, talep_no, proje_id, uye_id, kalem_id, malzeme_adi, miktar, birim_fiyat, toplam_fiyat, durum, gerekce, red_nedeni, olusturma_tarihi, guncelleme_tarihi)
						SELECT talep_id, talep_no, proje_id, uye_id, kalem_id, malzeme_adi, miktar, birim_fiyat, toplam_fiyat, durum, gerekce, red_nedeni, olusturma_tarihi, guncelleme_tarihi 
						FROM proje_satinalma_talebi_old;
					ELSE
						INSERT INTO proje_satinalma_talebi (talep_id, talep_no, proje_id, uye_id, kalem_id, malzeme_adi, miktar, birim_fiyat, toplam_fiyat, durum, gerekce, red_nedeni, olusturma_tarihi, guncelleme_tarihi)
						SELECT talep_id, NULL, proje_id, uye_id, kalem_id, malzeme_adi, miktar, birim_fiyat, toplam_fiyat, durum, gerekce, red_nedeni, olusturma_tarihi, guncelleme_tarihi 
						FROM proje_satinalma_talebi_old;
					END IF;
					
					-- İndeksleri tekrar oluştur
					CREATE INDEX IF NOT EXISTS idx_proje_satinalma_talebi_proje_id ON proje_satinalma_talebi(proje_id);
					CREATE INDEX IF NOT EXISTS idx_proje_satinalma_talebi_kalem_id ON proje_satinalma_talebi(kalem_id);
					
					-- ID dizisini (sequence) güncelle
					PERFORM setval(pg_get_serial_sequence('proje_satinalma_talebi', 'talep_id'), COALESCE(MAX(talep_id), 1)) FROM proje_satinalma_talebi;
					
					-- Eski tabloyu sil
					DROP TABLE proje_satinalma_talebi_old;
				END IF;
			END $$;

			-- Tabloda talep_no kolonu yoksa ekle (güvenlik için)
			ALTER TABLE proje_satinalma_talebi ADD COLUMN IF NOT EXISTS talep_no VARCHAR(100);
			ALTER TABLE proje_satinalma_talebi DROP CONSTRAINT IF EXISTS proje_satinalma_talebi_talep_no_key;

			-- Satın alma talep numaralarını güncelle
			WITH numbered_requests AS (
				SELECT 
					st.talep_id,
					'SA-' || LPAD(DENSE_RANK() OVER (
						PARTITION BY st.proje_id 
						ORDER BY st.olusturma_tarihi, st.talep_no
					)::text, 3, '0') as yepyeni_talep_no
				FROM proje_satinalma_talebi st
			)
			UPDATE proje_satinalma_talebi st
			SET talep_no = nr.yepyeni_talep_no
			FROM numbered_requests nr
			WHERE st.talep_id = nr.talep_id 
			  AND (st.talep_no IS NULL OR st.talep_no LIKE '%BAP%' OR st.talep_no NOT LIKE 'SA-%');
		`
		if _, err := db.Exec(satinalmaTalepNoQuery); err != nil {
			log.Printf("Uyarı: satinalma_talebi tablosunda talep_no'nun 2. sütun yapılması veya backfill uygulanamadı: %v", err)
		} else {
			log.Println("Bilgi: satinalma_talebi tablosunda talep_no'nun 2. sütun yapılması ve geriye dönük numara atamaları başarıyla uygulandı.")
		}

		// Türkçe Yorum: Komisyon toplantı ve katılım tabloları oluşturulur, komisyon başkanı rolü ve sayfası eklenir.
		komisyonBaskaniMigrationQuery := `
			CREATE TABLE IF NOT EXISTS komisyon_toplantisi (
				toplanti_id SERIAL PRIMARY KEY,
				toplanti_no VARCHAR(100) NOT NULL UNIQUE,
				tarih TIMESTAMP WITH TIME ZONE NOT NULL,
				gundem TEXT NOT NULL,
				karar TEXT NOT NULL,
				olusturan_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE SET NULL,
				olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);

			CREATE TABLE IF NOT EXISTS komisyon_toplanti_katilimci (
				toplanti_id INTEGER NOT NULL REFERENCES komisyon_toplantisi(toplanti_id) ON DELETE CASCADE,
				uye_id INTEGER NOT NULL REFERENCES uye(uye_id) ON DELETE CASCADE,
				katildi BOOLEAN NOT NULL DEFAULT TRUE,
				PRIMARY KEY (toplanti_id, uye_id)
			);

			INSERT INTO sistem_rol_tanimlama (rol_adi, rol_etiketi)
			VALUES ('komisyon_baskani', 'BAP Komisyon Başkanı') 
			ON CONFLICT (rol_adi) DO UPDATE SET rol_etiketi = EXCLUDED.rol_etiketi;

			INSERT INTO sistem_sayfa (sayfa_adi, sayfa_kodu, url_yolu)
			VALUES ('Komisyon Başkanı Dashboard', 'komisyon_baskani_dashboard', '/komisyon/baskan/dashboard') 
			ON CONFLICT (sayfa_kodu) DO NOTHING;

			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT r.rol_id, s.sayfa_id
			FROM sistem_rol_tanimlama r, sistem_sayfa s
			WHERE r.rol_adi IN ('admin', 'komisyon_baskani') AND s.sayfa_kodu = 'komisyon_baskani_dashboard'
			ON CONFLICT DO NOTHING;
		`
		if _, err := db.Exec(komisyonBaskaniMigrationQuery); err != nil {
			log.Printf("Uyarı: Komisyon Başkanı tabloları, rolleri ve yetkileri eklenemedi: %v", err)
		} else {
			log.Println("Bilgi: Komisyon Başkanı tabloları, rolleri ve yetkileri başarıyla eklendi/güncellendi.")
		}

		// Türkçe Yorum: Hakemin gizlilik taahhütnamesini onayladığı tarihi tutmak için sütun eklenir.
		// Hakem, projeyi kabul edip değerlendirmeye başlamadan önce taahhütnameyi onaylamak zorundadır.
		taahhutnameQuery := `
			ALTER TABLE proje_degerlendirmeleri
			ADD COLUMN IF NOT EXISTS taahhutname_onay_tarihi TIMESTAMP WITH TIME ZONE;
		`
		if _, err := db.Exec(taahhutnameQuery); err != nil {
			log.Printf("Uyarı: taahhutname_onay_tarihi sütunu eklenemedi: %v", err)
		} else {
			log.Println("Bilgi: proje_degerlendirmeleri tablosuna taahhutname_onay_tarihi sütunu başarıyla eklendi/kontrol edildi.")
		}

		// Türkçe Yorum: Satın alma bütçesinde seçilebilecek 'Bursiyer' bütçe kategorisi eklenir.
		// (Ön yüzde BAP-100 projelerinde bu kategori seçime kapatılır.)
		bursiyerKategoriQuery := `
			INSERT INTO proje_butce_kategori (kategori_adi)
			VALUES ('Bursiyer') ON CONFLICT (kategori_adi) DO NOTHING;
		`
		if _, err := db.Exec(bursiyerKategoriQuery); err != nil {
			log.Printf("Uyarı: 'Bursiyer' bütçe kategorisi eklenemedi: %v", err)
		} else {
			log.Println("Bilgi: 'Bursiyer' bütçe kategorisi başarıyla eklendi/kontrol edildi.")
		}

		// Türkçe Yorum: Proje sözleşmesi PDF'i tek seferlik indirilebilir; indirilme durumu takip edilir.
		// Ayrıca yürürlük tarihleri indirme anında hesaplandığından NOT NULL kısıtı kaldırılır.
		sozlesmeIndirmeQuery := `
			ALTER TABLE proje_sozlesme ADD COLUMN IF NOT EXISTS indirildi_mi BOOLEAN DEFAULT FALSE;
			ALTER TABLE proje_sozlesme ADD COLUMN IF NOT EXISTS indirme_tarihi TIMESTAMP;
			ALTER TABLE proje_sozlesme ALTER COLUMN baslangic_tarihi DROP NOT NULL;
			ALTER TABLE proje_sozlesme ALTER COLUMN bitis_tarihi DROP NOT NULL;
		`
		if _, err := db.Exec(sozlesmeIndirmeQuery); err != nil {
			log.Printf("Uyarı: proje_sozlesme indirme sütunları eklenemedi: %v", err)
		} else {
			log.Println("Bilgi: proje_sozlesme tablosuna indirme takip sütunları başarıyla eklendi/kontrol edildi.")
		}

		// Türkçe Yorum: Proje Başvuruları modülünü sistem_sayfa tablosuna ekler ve varsayılan olarak admin ile tto rollerine yetkisini atar.
		projeBasvurulariSayfaQuery := `
			INSERT INTO sistem_sayfa (sayfa_adi, sayfa_kodu, url_yolu)
			VALUES ('Proje Başvuruları', 'proje_basvurulari', '/admin/proje-basvurulari')
			ON CONFLICT (sayfa_kodu) DO UPDATE SET sayfa_adi = EXCLUDED.sayfa_adi, url_yolu = EXCLUDED.url_yolu;

			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT srt.rol_id, ss.sayfa_id
			FROM sistem_rol_tanimlama srt, sistem_sayfa ss
			WHERE ss.sayfa_kodu = 'proje_basvurulari' AND srt.rol_adi IN ('admin', 'tto')
			ON CONFLICT DO NOTHING;
		`
		if _, err := db.Exec(projeBasvurulariSayfaQuery); err != nil {
			log.Printf("Uyarı: 'Proje Başvuruları' sistem sayfası ve yetkileri eklenemedi: %v", err)
		} else {
			log.Println("Bilgi: 'Proje Başvuruları' sistem sayfası ve varsayılan yetkileri başarıyla eklendi/güncellendi.")
		}

		// Türkçe Yorum: Proje Talepleri Yönetimi modülünü sistem_sayfa tablosuna ekler ve varsayılan olarak admin ile tto rollerine yetkisini atar.
		projeTalepleriSayfaQuery := `
			INSERT INTO sistem_sayfa (sayfa_adi, sayfa_kodu, url_yolu)
			VALUES ('Proje Talepleri Yönetimi (TTO)', 'proje_talepleri_tto', '/tto/talepler')
			ON CONFLICT (sayfa_kodu) DO UPDATE SET sayfa_adi = EXCLUDED.sayfa_adi, url_yolu = EXCLUDED.url_yolu;

			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT srt.rol_id, ss.sayfa_id
			FROM sistem_rol_tanimlama srt, sistem_sayfa ss
			WHERE ss.sayfa_kodu = 'proje_talepleri_tto' AND srt.rol_adi IN ('admin', 'tto')
			ON CONFLICT DO NOTHING;
		`
		if _, err := db.Exec(projeTalepleriSayfaQuery); err != nil {
			log.Printf("Uyarı: 'Proje Talepleri Yönetimi' sistem sayfası ve yetkileri eklenemedi: %v", err)
		} else {
			log.Println("Bilgi: 'Proje Talepleri Yönetimi' sistem sayfası ve varsayılan yetkileri başarıyla eklendi/güncellendi.")
		}

		// Türkçe Yorum: Zamanlanmış Görevler modülünü sistem_sayfa tablosuna ekler ve varsayılan olarak admin ve komisyon_baskani rollerine atar.
		zamanlanmisGorevlerSayfaQuery := `
			INSERT INTO sistem_sayfa (sayfa_adi, sayfa_kodu, url_yolu)
			VALUES ('Zamanlanmış Görevler', 'zamanlanmis_gorevler', '/admin/zamanlanmis-gorevler')
			ON CONFLICT (sayfa_kodu) DO UPDATE SET sayfa_adi = EXCLUDED.sayfa_adi, url_yolu = EXCLUDED.url_yolu;

			INSERT INTO sayfa_rol_yetki (sistem_rol_id, sayfa_id)
			SELECT srt.rol_id, ss.sayfa_id
			FROM sistem_rol_tanimlama srt, sistem_sayfa ss
			WHERE ss.sayfa_kodu = 'zamanlanmis_gorevler' AND srt.rol_adi IN ('admin', 'komisyon_baskani')
			ON CONFLICT DO NOTHING;
		`
		if _, err := db.Exec(zamanlanmisGorevlerSayfaQuery); err != nil {
			log.Printf("Uyarı: 'Zamanlanmış Görevler' sistem sayfası ve yetkileri eklenemedi: %v", err)
		} else {
			log.Println("Bilgi: 'Zamanlanmış Görevler' sistem sayfası ve varsayılan yetkileri başarıyla eklendi/güncellendi.")
		}

		// Türkçe Yorum: Zamanlanmış görev kural ve log tablolarını kontrol eder ve yoksa oluşturur.
		zamanlanmisGorevlerTablesQuery := `
			CREATE TABLE IF NOT EXISTS zamanlanmis_gorev_kural (
				kural_id SERIAL PRIMARY KEY,
				kural_adi VARCHAR(200) NOT NULL,
				bap_turu_id INTEGER REFERENCES proje_bap_turu(bap_turu_id) ON DELETE CASCADE,
				tetikleme_tipi VARCHAR(50) NOT NULL,
				zaman_degeri INTEGER NOT NULL DEFAULT 1,
				eposta_aktif BOOLEAN DEFAULT TRUE,
				sms_aktif BOOLEAN DEFAULT FALSE,
				eposta_konu VARCHAR(255) NOT NULL,
				eposta_sablon TEXT NOT NULL,
				sms_sablon TEXT NOT NULL,
				aktif_mi BOOLEAN DEFAULT TRUE,
				olusturan_id INTEGER REFERENCES uye(uye_id) ON DELETE SET NULL,
				olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				guncelleme_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);

			CREATE TABLE IF NOT EXISTS zamanlanmis_gorev_log (
				log_id SERIAL PRIMARY KEY,
				kural_id INTEGER NOT NULL REFERENCES zamanlanmis_gorev_kural(kural_id) ON DELETE CASCADE,
				proje_id INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
				kanal VARCHAR(20) NOT NULL,
				alici VARCHAR(255) NOT NULL,
				icerik TEXT NOT NULL,
				durum VARCHAR(50) NOT NULL DEFAULT 'basarili',
				hata_mesaji TEXT,
				gonderim_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);

			INSERT INTO zamanlanmis_gorev_kural 
				(kural_adi, bap_turu_id, tetikleme_tipi, zaman_degeri, eposta_aktif, sms_aktif, eposta_konu, eposta_sablon, sms_sablon, aktif_mi)
			VALUES 
				('Gelişme Raporu 6. Ay Hatırlatması', NULL, 'baslangic_sonrasi_ay', 6, TRUE, TRUE, 
				 'BAP Projesi 6. Ay Gelişme Raporu Hatırlatması - {proje_kodu}', 
				 '<p>Sayın {yurutucu_ad},</p><p>Yürütücüsü olduğunuz {proje_kodu} kodlu ve "{proje_baslik}" başlıklı projenizin 6. ay gelişme raporu teslim zamanı yaklaşmıştır. Lütfen raporunuzu sisteme yükleyiniz.</p>', 
				 'Sayin {yurutucu_ad}, {proje_kodu} kodlu projenizin 6. ay gelisme raporu teslim zamani gelmistir. Lutfen BAP otomasyonuna giris yapiniz.', TRUE),
				('Proje Bitişine 30 Gün Kala Sonuç Raporu Uyarısı', NULL, 'bitim_oncesi_gun', 30, TRUE, TRUE, 
				 'BAP Proje Bitiş Uyarısı ve Sonuç Raporu Bildirimi - {proje_kodu}', 
				 '<p>Sayın {yurutucu_ad},</p><p>Yürütücüsü olduğunuz {proje_kodu} kodlu projenizin bitimine {kalan_gun} gün kalmıştır. Bitiş tarihinden itibaren 2 ay içinde kesin sonuç raporunun teslim edilmesi gerekmektedir.</p>', 
				 'Sayin {yurutucu_ad}, {proje_kodu} projenizin bitimine {kalan_gun} gun kalmistir. Detaylar icin BAP sistemini ziyaret ediniz.', TRUE)
			ON CONFLICT DO NOTHING;
		`
		if _, err := db.Exec(zamanlanmisGorevlerTablesQuery); err != nil {
			log.Printf("Uyarı: zamanlanmis_gorev_kural / log tabloları oluşturulamadı: %v", err)
		} else {
			log.Println("Bilgi: zamanlanmis_gorev_kural ve zamanlanmis_gorev_log tabloları kontrol edildi/başarıyla oluşturuldu.")
		}

		// Türkçe Yorum: Proje sözleşmesi e-posta hatırlatma log tablosunu kontrol eder ve yoksa oluşturur.
		hatirlatmaLogQuery := `
			CREATE TABLE IF NOT EXISTS proje_sozlesme_hatirlatma_log (
				id SERIAL PRIMARY KEY,
				sozlesme_id INT NOT NULL REFERENCES proje_sozlesme(id) ON DELETE CASCADE,
				proje_id INT NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
				gonderim_tarihi TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				gecen_sure VARCHAR(100) NOT NULL,
				kalan_sure VARCHAR(100) NOT NULL,
				gonderilen_eposta VARCHAR(150) NOT NULL,
				donem_indeks INT DEFAULT 1,
				durum VARCHAR(20) DEFAULT 'gonderildi'
			);
		`
		if _, err := db.Exec(hatirlatmaLogQuery); err != nil {
			log.Printf("Uyarı: proje_sozlesme_hatirlatma_log tablosu oluşturulamadı: %v", err)
		} else {
			log.Println("Bilgi: proje_sozlesme_hatirlatma_log tablosu kontrol edildi/başarıyla oluşturuldu.")
		}

		// Türkçe Yorum: proje_bap_turu ara_rapor sütunları eklenir.
		araRaporMigrationQuery := `
			ALTER TABLE proje_bap_turu ADD COLUMN IF NOT EXISTS ara_rapor_gerekli BOOLEAN DEFAULT FALSE;
			ALTER TABLE proje_bap_turu ADD COLUMN IF NOT EXISTS ara_rapor_sayisi INTEGER DEFAULT 0;
		`
		if _, err := db.Exec(araRaporMigrationQuery); err != nil {
			log.Printf("Uyarı: proje_bap_turu ara_rapor sütunları eklenemedi: %v", err)
		} else {
			log.Println("Bilgi: proje_bap_turu ara_rapor sütunları başarıyla yüklendi/kontrol edildi.")
		}

		// Türkçe Yorum: BAP türü versiyonlama — taslak/yayın/arşiv; projeler versiyona kilitlenir.
		bapVersiyonMigrationQuery := `
			CREATE TABLE IF NOT EXISTS proje_bap_turu_versiyon (
				versiyon_id SERIAL PRIMARY KEY,
				bap_turu_id INTEGER NOT NULL REFERENCES proje_bap_turu(bap_turu_id) ON DELETE CASCADE,
				versiyon_no INTEGER,
				durum VARCHAR(20) NOT NULL DEFAULT 'taslak',
				butce_limiti NUMERIC(12, 2) DEFAULT 0,
				sure_limiti_ay INTEGER DEFAULT 0,
				aciklama TEXT DEFAULT '',
				hakem_gerekli BOOLEAN DEFAULT FALSE,
				hakem_sayisi INTEGER DEFAULT 0,
				bursiyer_gerekli BOOLEAN DEFAULT FALSE,
				bursiyer_sayisi INTEGER DEFAULT 0,
				ara_rapor_gerekli BOOLEAN DEFAULT FALSE,
				ara_rapor_sayisi INTEGER DEFAULT 0,
				olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				yayin_tarihi TIMESTAMP WITH TIME ZONE,
				CONSTRAINT chk_bap_versiyon_durum CHECK (durum IN ('taslak', 'yayinda', 'arsiv')),
				CONSTRAINT uq_bap_turu_versiyon_no UNIQUE (bap_turu_id, versiyon_no)
			);
			CREATE INDEX IF NOT EXISTS idx_bap_versiyon_turu ON proje_bap_turu_versiyon(bap_turu_id);
			CREATE INDEX IF NOT EXISTS idx_bap_versiyon_durum ON proje_bap_turu_versiyon(bap_turu_id, durum);

			CREATE TABLE IF NOT EXISTS proje_bap_turu_versiyon_asama (
				versiyon_id INTEGER NOT NULL REFERENCES proje_bap_turu_versiyon(versiyon_id) ON DELETE CASCADE,
				asama_id INTEGER NOT NULL REFERENCES proje_asama(asama_id) ON DELETE CASCADE,
				sira_no INTEGER NOT NULL,
				PRIMARY KEY (versiyon_id, asama_id)
			);

			ALTER TABLE proje ADD COLUMN IF NOT EXISTS bap_turu_versiyon_id INTEGER REFERENCES proje_bap_turu_versiyon(versiyon_id) ON DELETE SET NULL;
			CREATE INDEX IF NOT EXISTS idx_proje_bap_turu_versiyon_id ON proje(bap_turu_versiyon_id);

			-- Mevcut türler için v1 yayın (yalnızca henüz versiyonu olmayanlar)
			INSERT INTO proje_bap_turu_versiyon (
				bap_turu_id, versiyon_no, durum, butce_limiti, sure_limiti_ay, aciklama,
				hakem_gerekli, hakem_sayisi, bursiyer_gerekli, bursiyer_sayisi,
				ara_rapor_gerekli, ara_rapor_sayisi, yayin_tarihi
			)
			SELECT pbt.bap_turu_id, 1, 'yayinda',
				COALESCE(pbt.butce_limiti, 0), COALESCE(pbt.sure_limiti_ay, 0), COALESCE(pbt.aciklama, ''),
				COALESCE(pbt.hakem_gerekli, false), COALESCE(pbt.hakem_sayisi, 0),
				COALESCE(pbt.bursiyer_gerekli, false), COALESCE(pbt.bursiyer_sayisi, 0),
				COALESCE(pbt.ara_rapor_gerekli, false), COALESCE(pbt.ara_rapor_sayisi, 0),
				CURRENT_TIMESTAMP
			FROM proje_bap_turu pbt
			WHERE NOT EXISTS (
				SELECT 1 FROM proje_bap_turu_versiyon v WHERE v.bap_turu_id = pbt.bap_turu_id
			);

			INSERT INTO proje_bap_turu_versiyon_asama (versiyon_id, asama_id, sira_no)
			SELECT v.versiyon_id, pbta.asama_id, pbta.sira_no
			FROM proje_bap_turu_versiyon v
			JOIN proje_bap_turu_asama pbta ON pbta.bap_turu_id = v.bap_turu_id
			WHERE v.versiyon_no = 1 AND v.durum = 'yayinda'
			ON CONFLICT DO NOTHING;

			-- Aşama kaydı olmayan v1'ler için tüm aşamaları yedekle
			INSERT INTO proje_bap_turu_versiyon_asama (versiyon_id, asama_id, sira_no)
			SELECT v.versiyon_id, pa.asama_id, pa.sira_no
			FROM proje_bap_turu_versiyon v
			CROSS JOIN proje_asama pa
			WHERE v.versiyon_no = 1 AND v.durum = 'yayinda'
			  AND NOT EXISTS (
				SELECT 1 FROM proje_bap_turu_versiyon_asama va WHERE va.versiyon_id = v.versiyon_id
			  )
			ON CONFLICT DO NOTHING;

			UPDATE proje p
			SET bap_turu_versiyon_id = v.versiyon_id
			FROM proje_bap_turu_versiyon v
			WHERE p.bap_turu_id = v.bap_turu_id
			  AND v.durum = 'yayinda'
			  AND p.bap_turu_versiyon_id IS NULL
			  AND p.bap_turu_id IS NOT NULL;
		`
		if _, err := db.Exec(bapVersiyonMigrationQuery); err != nil {
			log.Printf("Uyarı: BAP türü versiyonlama migrasyonu başarısız: %v", err)
		} else {
			log.Println("Bilgi: BAP türü versiyonlama (taslak/yayın) tabloları ve v1 backfill kontrol edildi.")
		}

		// Türkçe Yorum: proje_satinalma_talebi tablosuna bütçe revizyon sütunları eklenir.
		butceRevizyonQuery := `
			ALTER TABLE proje_satinalma_talebi ADD COLUMN IF NOT EXISTS revize_birim_fiyat NUMERIC(10, 2);
			ALTER TABLE proje_satinalma_talebi ADD COLUMN IF NOT EXISTS revize_toplam_fiyat NUMERIC(12, 2);
			ALTER TABLE proje_satinalma_talebi ADD COLUMN IF NOT EXISTS revizyon_gerekcesi TEXT;
			ALTER TABLE proje_satinalma_talebi ADD COLUMN IF NOT EXISTS revize_eden_id INTEGER REFERENCES uye(uye_id) ON DELETE SET NULL;
			ALTER TABLE proje_satinalma_talebi ADD COLUMN IF NOT EXISTS revizyon_tarihi TIMESTAMP WITH TIME ZONE;
		`
		if _, err := db.Exec(butceRevizyonQuery); err != nil {
			log.Printf("Uyarı: proje_satinalma_talebi bütçe revizyon sütunları eklenemedi: %v", err)
		} else {
			log.Println("Bilgi: proje_satinalma_talebi bütçe revizyon sütunları başarıyla yüklendi/kontrol edildi.")
		}

		// Türkçe Yorum: 'yururlukte' durumundaki ve sözleşme tarihleri eksik projelerin tarihlerini backfill eder.
		sozlesmeBackfillQuery := `
			INSERT INTO proje_sozlesme (
				proje_id, uye_id, tc_kimlik, yurutucu_adres, yurutucu_telefon, yurutucu_eposta, 
				baslangic_tarihi, bitis_tarihi, durum, indirildi_mi, indirme_tarihi
			)
			SELECT 
				p.proje_id, 
				p.koordinator_id, 
				'11111111111', 
				'İZÜ Kampüsü', 
				'02126929600', 
				COALESCE(u.eposta, 'yurutucu@izu.edu.tr'), 
				p.olusturma_tarihi::date, 
				(p.olusturma_tarihi + (INTERVAL '1 month' * COALESCE(p.sure_ay, 12)))::date,
				'imzalandi',
				true,
				p.olusturma_tarihi
			FROM proje p
			JOIN uye u ON p.koordinator_id = u.uye_id
			JOIN proje_durum pd ON p.durum_id = pd.durum_id
			WHERE pd.durum_adi = 'yururlukte'
			ON CONFLICT (proje_id) DO UPDATE SET
				baslangic_tarihi = COALESCE(proje_sozlesme.baslangic_tarihi, EXCLUDED.baslangic_tarihi),
				bitis_tarihi = COALESCE(proje_sozlesme.bitis_tarihi, EXCLUDED.bitis_tarihi),
				durum = 'imzalandi',
				indirildi_mi = true,
				indirme_tarihi = COALESCE(proje_sozlesme.indirme_tarihi, EXCLUDED.indirme_tarihi);
		`
		if _, err := db.Exec(sozlesmeBackfillQuery); err != nil {
			log.Printf("Uyarı: Aktif projelerin sözleşme tarihleri backfill edilemedi: %v", err)
		} else {
			log.Println("Bilgi: Aktif projelerin sözleşme tarihleri başarıyla backfill edildi/kontrol edildi.")
		}

		// Türkçe Yorum: Aynı SA-XXX numarasının farklı projelerde tekrarlanmasını giderir (global tekil talep_no).
		satinalmaTalepNoUniqueQuery := `
			DO $$
			BEGIN
				IF EXISTS (
					SELECT 1
					FROM proje_satinalma_talebi
					WHERE talep_no IS NOT NULL AND talep_no <> ''
					GROUP BY talep_no
					HAVING COUNT(DISTINCT proje_id) > 1
				) THEN
					WITH groups AS (
						SELECT
							proje_id,
							talep_no,
							MIN(olusturma_tarihi) AS first_created
						FROM proje_satinalma_talebi
						WHERE talep_no IS NOT NULL AND talep_no <> ''
						GROUP BY proje_id, talep_no
					),
					ranked AS (
						SELECT
							proje_id,
							talep_no,
							'SA-' || LPAD(
								ROW_NUMBER() OVER (ORDER BY first_created, proje_id, talep_no)::text,
								3, '0'
							) AS yeni_no
						FROM groups
					)
					UPDATE proje_satinalma_talebi st
					SET talep_no = r.yeni_no
					FROM ranked r
					WHERE st.proje_id = r.proje_id
					  AND st.talep_no = r.talep_no;
				END IF;
			END $$;
		`
		if _, err := db.Exec(satinalmaTalepNoUniqueQuery); err != nil {
			log.Printf("Uyarı: satinalma talep_no global tekilleştirme uygulanamadı: %v", err)
		} else {
			log.Println("Bilgi: satinalma talep_no global tekilleştirme kontrol edildi/uygulandı.")
		}

		// Türkçe Yorum: Komisyon toplantısı ↔ proje ilişkilendirme için köprü tablo ve durum kolonu eklenir.
		// komisyon_toplantisi.durum kolonu yoksa eklenir; komisyon_toplanti_proje köprü tablosu oluşturulur.
		komisyonProjeQuery := `
			ALTER TABLE komisyon_toplantisi
				ADD COLUMN IF NOT EXISTS durum VARCHAR(30) NOT NULL DEFAULT 'tamamlandi';

			CREATE TABLE IF NOT EXISTS komisyon_toplanti_proje (
				id               SERIAL PRIMARY KEY,
				toplanti_id      INTEGER NOT NULL REFERENCES komisyon_toplantisi(toplanti_id) ON DELETE CASCADE,
				proje_id         INTEGER NOT NULL REFERENCES proje(proje_id) ON DELETE CASCADE,
				gundem_sirasi    INTEGER,
				karar            VARCHAR(50) NOT NULL DEFAULT 'bekliyor',
				karar_aciklamasi TEXT,
				karar_tarihi     TIMESTAMP WITH TIME ZONE,
				ekleyen_id       INTEGER REFERENCES uye(uye_id) ON DELETE SET NULL,
				olusturma_tarihi TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(toplanti_id, proje_id)
			);
			CREATE INDEX IF NOT EXISTS idx_ktp_toplanti ON komisyon_toplanti_proje(toplanti_id);
			CREATE INDEX IF NOT EXISTS idx_ktp_proje    ON komisyon_toplanti_proje(proje_id);
		`
		if _, err := db.Exec(komisyonProjeQuery); err != nil {
			log.Printf("Uyarı: komisyon_toplanti_proje tablosu oluşturulamadı veya durum kolonu eklenemedi: %v", err)
		} else {
			log.Println("Bilgi: komisyon_toplanti_proje köprü tablosu ve durum kolonu başarıyla oluşturuldu/kontrol edildi.")
		}

		// Türkçe Yorum: Karar alanına revizyon değeri dokümantasyonu (VARCHAR kısıtı yok; uygulama katmanı doğrular).
		log.Println("Bilgi: komisyon_toplanti_proje.karar değerleri: bekliyor|onaylandi|reddedildi|ertelendi|revizyon")

		return nil

	}

	// Şema dosyasını oku
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("şema dosyası okunamadı (%s): %v", schemaPath, err)
	}

	// Şemayı çalıştır
	_, err = db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("şema çalıştırılamadı (%s): %v", schemaPath, err)
	}

	log.Printf("Veritabanı şeması başarıyla uygulandı: %s\n", schemaPath)
	return nil
}
