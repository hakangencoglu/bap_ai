package service

import (
	"database/sql"
	"fmt"
	"strings"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// ProjeService yapısı proje iş mantığını barındırır.
type ProjeService struct {
	ProjeRepo    *repository.ProjeRepository
	HakemRepo    *repository.HakemRepository
	RevizyonRepo *repository.RevizyonRepository
}

// NewProjeService fonksiyonu yeni bir ProjeService oluşturur.
func NewProjeService(projeRepo *repository.ProjeRepository, hakemRepo *repository.HakemRepository, revizyonRepo *repository.RevizyonRepository) *ProjeService {
	return &ProjeService{
		ProjeRepo:    projeRepo,
		HakemRepo:    hakemRepo,
		RevizyonRepo: revizyonRepo,
	}
}

// ValidateNewApplication, kullanıcının aktif kısıtlamalarını ve seçilen BAP türünün yönetici tarafından belirlenmiş dinamik kurallarını denetler.
func (s *ProjeService) ValidateNewApplication(uyeID int, p *models.Proje) (*models.ProjeBapTuru, error) {
	// 1. Üye Aktif Cezai Bloke / Kısıtlama Kontrolü (uye_kisitlama)
	var kisitlamaAciklama string
	err := s.ProjeRepo.DB.QueryRow(`
		SELECT aciklama FROM uye_kisitlama
		WHERE uye_id = $1 AND aktif_mi = true
		  AND (bitis_tarihi IS NULL OR bitis_tarihi > CURRENT_TIMESTAMP)
		ORDER BY kisitlama_id DESC LIMIT 1
	`, uyeID).Scan(&kisitlamaAciklama)
	if err == nil && kisitlamaAciklama != "" {
		return nil, fmt.Errorf("İZÜ BAP Yönergesi gereği aktif kısıtlamanız bulunmaktadır (%s). Yeni BAP başvurusu yapamazsınız", kisitlamaAciklama)
	}

	if p.BapTuruID == nil {
		return nil, nil
	}

	// 2. Seçilen BAP Türünün Aktif / Donmuş Kurallarını Yükle
	var rule models.ProjeBapTuru
	query := `
		SELECT pbt.bap_turu_id, pbt.bap_turu,
		       COALESCE(v.sure_limiti_ay, COALESCE(pbt.sure_limiti_ay, 24)),
		       COALESCE(v.hakem_gerekli, COALESCE(pbt.hakem_gerekli, true)),
		       COALESCE(v.hakem_sayisi, COALESCE(pbt.hakem_sayisi, 2)),
		       COALESCE(v.hakem_turu_kisitlama, COALESCE(pbt.hakem_turu_kisitlama, 'herhangi')),
		       COALESCE(v.hakem_sure_gun, COALESCE(pbt.hakem_sure_gun, 15)),
		       COALESCE(v.hakem_ucret_orani_yuzde, COALESCE(pbt.hakem_ucret_orani_yuzde, 3.00)),
		       COALESCE(v.bursiyer_izinli_mi, COALESCE(pbt.bursiyer_izinli_mi, true)),
		       COALESCE(v.max_aktif_proje_sayisi, COALESCE(pbt.max_aktif_proje_sayisi, 0)),
		       COALESCE(v.tez_ogrencisi_sarti, COALESCE(pbt.tez_ogrencisi_sarti, false)),
		       COALESCE(v.yayin_gecmis_sarti, COALESCE(pbt.yayin_gecmis_sarti, false)),
		       COALESCE(v.intihal_cezasi_aktif, COALESCE(pbt.intihal_cezasi_aktif, true)),
		       COALESCE(v.yurutucu_gecmis_proje_sarti, COALESCE(pbt.yurutucu_gecmis_proje_sarti, false)),
		       COALESCE(v.izin_seyahat_beyani_zorunlu, COALESCE(pbt.izin_seyahat_beyani_zorunlu, false)),
		       COALESCE(v.firma_ortaklik_beyani_zorunlu, COALESCE(pbt.firma_ortaklik_beyani_zorunlu, false)),
		       COALESCE(v.min_kurum_hissesi_orani, COALESCE(pbt.min_kurum_hissesi_orani, 0.00))
		FROM proje_bap_turu pbt
		LEFT JOIN LATERAL (
			SELECT * FROM proje_bap_turu_versiyon
			WHERE bap_turu_id = pbt.bap_turu_id AND durum = 'yayinda'
			ORDER BY versiyon_no DESC NULLS LAST, versiyon_id DESC LIMIT 1
		) v ON true
		WHERE pbt.bap_turu_id = $1
	`
	err = s.ProjeRepo.DB.QueryRow(query, *p.BapTuruID).Scan(
		&rule.BapTuruID, &rule.BapTuru, &rule.SureLimitiAy,
		&rule.HakemGerekli, &rule.HakemSayisi, &rule.HakemTuruKisitlama,
		&rule.HakemSureGun, &rule.HakemUcretOraniYuzde, &rule.BursiyerIzinliMi,
		&rule.MaxAktifProjeSayisi, &rule.TezOgrencisiSarti, &rule.YayinGecmisSarti,
		&rule.IntihalCezasiAktif, &rule.YurutucuGecmisProjeSarti,
		&rule.IzinSeyahatBeyaniZorunlu, &rule.FirmaOrtaklikBeyaniZorunlu, &rule.MinKurumHissesiOrani,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// 3. Maksimum Süre Denetimi
	if rule.SureLimitiAy > 0 && p.SureAy > rule.SureLimitiAy {
		return nil, fmt.Errorf("%s başvurusunda maksimum proje süresi %d aydır", rule.BapTuru, rule.SureLimitiAy)
	}

	// 4. Aktif Proje Sayısı Kısıtlaması
	if rule.MaxAktifProjeSayisi > 0 {
		var activeCount int
		s.ProjeRepo.DB.QueryRow(`
			SELECT COUNT(*) FROM proje p
			JOIN proje_durum pd ON p.durum_id = pd.durum_id
			WHERE p.koordinator_id = $1 AND p.bap_turu_id = $2
			  AND pd.durum_adi NOT IN ('reddedildi', 'tamamlandi')
		`, uyeID, *p.BapTuruID).Scan(&activeCount)
		if activeCount >= rule.MaxAktifProjeSayisi {
			return nil, fmt.Errorf("%s türünde aynı anda en fazla %d aktif proje yürütebilirsiniz. Aktif proje sayınız: %d", rule.BapTuru, rule.MaxAktifProjeSayisi, activeCount)
		}
	}

	// 5. İzin / Seyahat Durumu Beyan Denetimi
	if rule.IzinSeyahatBeyaniZorunlu && p.IzinSeyahatBeyani == "" {
		return nil, fmt.Errorf("İZÜ BAP Yönergesi Madde 8.8 gereğince sonraki döneme ait İzin / Kurum Dışında Bulunma beyanı yapılması zorunludur")
	}

	// 6. Firma Ortaklığı Olmama Beyanı (B Tipi Projeler)
	if rule.FirmaOrtaklikBeyaniZorunlu && !p.FirmaOrtaklikBeyani {
		return nil, fmt.Errorf("İZÜ BAP Yönergesi Madde 8.2 gereğince proje ekibindeki öğretim üyelerinin iş birliği yapılan özel sektör kuruluşunda ortaklığı/sahipliği bulunmadığına dair beyan zorunludur")
	}

	// 7. Minimum Kurum Hissesi Denetimi (Akademik Danışmanlık Projeleri)
	if rule.MinKurumHissesiOrani > 0 && p.KurumHissesiOrani < rule.MinKurumHissesiOrani {
		return nil, fmt.Errorf("%s projelerinde bütçe teklifi en az %%%.2f oranında Kurum Hissesi içermek zorundadır", rule.BapTuru, rule.MinKurumHissesiOrani)
	}

	// 8. Geçmiş Ulusal/Uluslararası Proje Yürütücülüğü Şartı (BAP500)
	if rule.YurutucuGecmisProjeSarti && p.YurutucuGecmisProjeBeyani == "" {
		return nil, fmt.Errorf("%s programına başvurabilmek için daha önce başarıyla tamamlanmış en az bir ulusal/uluslararası projede yürütücülük yapılmış olma şartının beyan edilmesi gerekmektedir", rule.BapTuru)
	}

	return &rule, nil
}

// CreateProje veritabanına bir proje ekler ve üye-proje bağlantısını sağlar.
// Türkçe Yorum: Dinamik yönetici kural kontrolü ve BAP türüne uygun hakem ataması yapılır.
func (s *ProjeService) CreateProje(uyeID int, p *models.Proje, uyeRol string) error {
	// Başvuru öncesi dinamik yönetici kurallarını denetle
	rule, err := s.ValidateNewApplication(uyeID, p)
	if err != nil {
		return err
	}

	err = s.ProjeRepo.CreateProje(uyeID, p, uyeRol)
	if err != nil {
		return fmt.Errorf("proje oluşturulurken bir hata meydana geldi: %w", err)
	}

	// BAP türünün hakem kuralına göre dinamik atama yap
	if s.HakemRepo != nil && rule != nil {
		if rule.HakemGerekli && rule.HakemSayisi > 0 {
			s.HakemRepo.AssignRandomHakemWithRules(p.ProjeID, rule.HakemSayisi, rule.HakemTuruKisitlama, rule.HakemSureGun, rule.HakemUcretOraniYuzde)
		} else if !rule.HakemGerekli || rule.HakemSayisi == 0 {
			// Hakemsiz proje türü (Örn: BAP100 veya Akademik Danışmanlık): doğrudan Komisyon Onayına sevk et
			s.ProjeRepo.UpdateProjectStatusWithLog(p.ProjeID, uyeID, "incelemede", "komisyon_bekliyor", "Hakemsiz proje türü: Başvuru doğrudan BAP Komisyon değerlendirmesine sevk edildi.")
		}
	} else if s.HakemRepo != nil {
		s.HakemRepo.AssignRandomHakem(p.ProjeID, 2)
	}

	return nil
}

// GetProjeByID proje ID'sine göre projeyi döner.
// GetProjeByID ID'ye göre tek bir projeyi getirir.
// Türkçe Yorum: Sözleşme formu için komisyon karar tarihi de eklenir.
func (s *ProjeService) GetProjeByID(projeID int) (*models.Proje, error) {
	p, err := s.ProjeRepo.GetProjeByID(projeID)
	if err != nil || p == nil {
		return p, err
	}
	var kararTarihi sql.NullTime
	qErr := s.ProjeRepo.DB.QueryRow(`
		SELECT COALESCE(kt.tarih::timestamp, ktp.karar_tarihi)
		FROM komisyon_toplanti_proje ktp
		JOIN komisyon_toplantisi kt ON kt.toplanti_id = ktp.toplanti_id
		WHERE ktp.proje_id = $1
		  AND ktp.karar = 'onaylandi'
		ORDER BY COALESCE(kt.tarih::timestamp, ktp.karar_tarihi) DESC NULLS LAST, ktp.id DESC
		LIMIT 1
	`, projeID).Scan(&kararTarihi)
	if qErr == nil && kararTarihi.Valid {
		p.KomisyonKararTarihi = kararTarihi.Time.Format("2006-01-02")
	}
	return p, nil
}

// UpdateProje mevcut bir projeyi günceller ve varsa bekleyen revizyonu kapatır.
func (s *ProjeService) UpdateProje(p *models.Proje) error {
	err := s.ProjeRepo.UpdateProje(p)
	if err == nil && s.RevizyonRepo != nil {
		s.RevizyonRepo.MarkRevizyonAsDone(p.ProjeID)
	}
	return err
}

// DeleteTaslakProje taslak durumundaki bir projeyi siler.
// Yetki ve durum kontrolü repository katmanında yapılır.
func (s *ProjeService) DeleteTaslakProje(projeID int, uyeID int) error {
	return s.ProjeRepo.DeleteTaslakProje(projeID, uyeID)
}

// isHakemGerekli projenin bağlı BAP türü versiyonunda hakem değerlendirmesi gerekli mi kontrol eder.
// Türkçe Yorum: Sorgu başarısız olursa güvenli varsayılan olarak hakem gerekli kabul edilir.
func (s *ProjeService) isHakemGerekli(projeID int) bool {
	var hakemGerekli bool
	err := s.ProjeRepo.DB.QueryRow(`
		SELECT COALESCE(pbv.hakem_gerekli, COALESCE(pbt.hakem_gerekli, false))
		FROM proje p
		LEFT JOIN proje_bap_turu_versiyon pbv ON p.bap_turu_versiyon_id = pbv.versiyon_id
		LEFT JOIN proje_bap_turu pbt ON p.bap_turu_id = pbt.bap_turu_id
		WHERE p.proje_id = $1
	`, projeID).Scan(&hakemGerekli)
	if err != nil {
		return true
	}
	return hakemGerekli
}

// GetSonrakiAsama projenin iş akışında mevcut durumdan sonraki aşamayı döner.
// Türkçe Yorum: Aşama sırası projenin bağlı BAP türü versiyonundan okunur; hakem gerekmiyorsa
// hakem aşaması akıştan düşülür. Süreç sonundaysa nil döner.
func (s *ProjeService) GetSonrakiAsama(projeID int, currentDurum string) (*models.ProjeAsama, error) {
	akis, err := s.ProjeRepo.GetWorkflowStages(projeID)
	if err != nil {
		return nil, err
	}
	if len(akis) == 0 {
		return nil, fmt.Errorf("projenin iş akışı tanımlı değil")
	}

	// Hakem değerlendirmesi gerekmiyorsa hakem aşaması akıştan çıkarılır
	if !s.isHakemGerekli(projeID) {
		var filtreli []models.ProjeAsama
		for _, a := range akis {
			if a.AsamaKodu != models.AsamaHakemeSun {
				filtreli = append(filtreli, a)
			}
		}
		akis = filtreli
	}
	if len(akis) == 0 {
		return nil, nil
	}

	// Mevcut durumun karşılık geldiği aşamayı bul (bekleme veya onay durumu üzerinden)
	mevcutIdx := -1
	for i, a := range akis {
		if a.DurumAdi == currentDurum || a.OnayDurumAdi == currentDurum {
			mevcutIdx = i
			break
		}
	}

	// Akışta karşılığı olmayan durumlar (ör. taslak) için ilk aşama sonraki aşamadır
	if mevcutIdx == -1 {
		return &akis[0], nil
	}
	if mevcutIdx+1 < len(akis) {
		return &akis[mevcutIdx+1], nil
	}
	return nil, nil
}

// GetNextWorkflowStatus bir projenin BAP türü versiyonuna göre bir sonraki aşama durumunu bulur.
// Türkçe Yorum: Sonraki aşamanın bekleme durumunu döner; aşama kalmadıysa süreç 'yururlukte' olur.
func (s *ProjeService) GetNextWorkflowStatus(projeID int, currentDurum string) (string, error) {
	sonraki, err := s.GetSonrakiAsama(projeID, currentDurum)
	if err != nil {
		return "", err
	}
	if sonraki == nil || sonraki.DurumAdi == "" {
		return models.DurumYururlukte, nil
	}
	return sonraki.DurumAdi, nil
}

// resolveKomisyonDurum komisyon_bekliyor aşamasındaki özel oylama mantığını işler.
// Türkçe Yorum: Admin/TTO/Başkan tek karar verebilir; normal komisyon üyesi oy kullanır
// ve tüm oylar tamamlanınca sonuç hesaplanır.
func (s *ProjeService) resolveKomisyonDurum(projeID int, islemYapanID int, action string, aciklama string) (string, error) {
	if action != models.AksiyonOnayla && action != models.AksiyonReddet && action != models.AksiyonRevizyon {
		return "", fmt.Errorf("geçersiz komisyon aksiyonu: %s", action)
	}

	// Kullanıcı rollerini sorgula (sistem_rol + uye_detay + uye)
	var userRoles string
	err := s.ProjeRepo.DB.QueryRow(`
		SELECT COALESCE((
		    SELECT string_agg(srt.rol_adi, ',') 
		    FROM sistem_rol sr 
		    INNER JOIN sistem_rol_tanimlama srt ON sr.sistem_rol_id = srt.rol_id 
		    WHERE sr.uye_id = u.uye_id
		), d.rol, u.rol, '')
		FROM uye u
		LEFT JOIN uye_detay d ON u.uye_id = d.uye_id
		WHERE u.uye_id = $1
	`, islemYapanID).Scan(&userRoles)
	if err != nil {
		return "", fmt.Errorf("kullanıcı bilgisi alınamadı: %w", err)
	}

	// Admin, TTO, Komisyon Üyesi, Komisyon Başkanı veya Komisyon Raportörü nihai kararı tek başına verir
	isYonetici := false
	for _, r := range strings.Split(userRoles, ",") {
		r = strings.TrimSpace(r)
		if r == models.RolAdmin || r == models.RolTTO || r == models.RolKomisyonBaskani || r == models.RolKomisyonRaportoru || r == models.RolKomisyon {
			isYonetici = true
			break
		}
	}

	if isYonetici {
		// Türkçe Yorum: Komisyon yetkilisi kararı verdiğinde proje durumu direkt güncellenir
		switch action {
		case models.AksiyonReddet:
			return models.DurumReddedildi, nil
		case models.AksiyonRevizyon:
			return models.DurumRevizyon, nil
		default:
			return models.DurumKomisyonOnayladi, nil
		}
	}

	// Normal komisyon üyesi: bireysel oyunu kaydet, sonra tüm oyları kontrol et
	if err := s.ProjeRepo.UpdateKomisyonKarar(projeID, islemYapanID, action, aciklama); err != nil {
		return "", fmt.Errorf("komisyon kararı kaydedilemedi: %w", err)
	}

	allApproved, criticalKarar, err := s.ProjeRepo.CheckAllKomisyonApproved(projeID)
	if err != nil {
		return "", fmt.Errorf("komisyon onayları kontrol edilemedi: %w", err)
	}

	switch criticalKarar {
	case models.AksiyonReddet:
		return models.DurumReddedildi, nil
	case models.AksiyonRevizyon:
		return models.DurumRevizyon, nil
	}
	if allApproved {
		return models.DurumKomisyonOnayladi, nil
	}
	// Türkçe Yorum: Henüz tüm üyeler oy kullanmadı, durum değişmez
	return models.DurumKomisyonBekliyor, nil
}

// resolveKomisyonOnayladiDurum komisyon onayı sonrası sonraki durumu iş akışından çözer.
// Türkçe Yorum: Sıra BAP türü versiyonundaki aşama sırasına göre belirlenir; hakem
// aşaması akışta yoksa (ya da hakem gerekmiyorsa) doğrudan sonraki aşamaya geçilir.
func (s *ProjeService) resolveKomisyonOnayladiDurum(projeID int, action string) (string, error) {
	switch action {
	case models.AksiyonReddet:
		return models.DurumReddedildi, nil
	case models.AksiyonRevizyon:
		return models.DurumRevizyon, nil
	case models.AksiyonOnayla, models.AksiyonOnaylaHakemsiz:
		return s.GetNextWorkflowStatus(projeID, models.DurumKomisyonOnayladi)
	default:
		return "", fmt.Errorf("geçersiz işlem: %s", action)
	}
}

// ProcessWorkflowAction onay mekanizmasındaki kararları işler (TTO, Dekan, Komisyon, Hakem kararları).
// Türkçe Yorum: Durum geçişleri workflowGecisTablo üzerinden çözülür. Komisyon oylaması ve
// hakem zorunluluğu gibi özel mantık ayrı yardımcı fonksiyonlara taşınmıştır.
func (s *ProjeService) ProcessWorkflowAction(projeID int, islemYapanID int, action string, aciklama string) error {
	p, err := s.ProjeRepo.GetProjeByID(projeID)
	if err != nil {
		return fmt.Errorf("proje bulunamadı: %w", err)
	}

	var yeniDurum string

	switch p.DurumAdi {
	case models.DurumKomisyonBekliyor:
		// Türkçe Yorum: Komisyon oylaması özel mantık içerdiği için ayrı fonksiyonda işleniyor
		yeniDurum, err = s.resolveKomisyonDurum(projeID, islemYapanID, action, aciklama)
		if err != nil {
			return err
		}

	case models.DurumKomisyonOnayladi:
		// Türkçe Yorum: Hakem gereksinimi kontrolü için ayrı fonksiyon çağrılıyor
		yeniDurum, err = s.resolveKomisyonOnayladiDurum(projeID, action)
		if err != nil {
			return err
		}

	default:
		if action == models.AksiyonReddet {
			yeniDurum = models.DurumReddedildi
		} else if action == models.AksiyonRevizyon {
			yeniDurum = models.DurumRevizyon
		} else if action == models.AksiyonTamamla {
			// Türkçe Yorum: Yürürlükteki proje TTO tarafından başarıyla tamamlanır.
			if p.DurumAdi != models.DurumYururlukte && p.DurumAdi != models.DurumTTOAktif {
				return fmt.Errorf("yalnızca yürürlükteki projeler tamamlanabilir (mevcut: %s)", p.DurumAdi)
			}
			yeniDurum = models.DurumTamamlandi
		} else if action == models.AksiyonOnayla {
			// Türkçe Yorum: Dinamik iş akışı geçişlerini durum bazlı belirler
			switch p.DurumAdi {
			case models.DurumIncelemede:
				yeniDurum, err = s.GetNextWorkflowStatus(projeID, p.DurumAdi)
			case models.DurumDekanOnayiBekliyor:
				yeniDurum = models.DurumDekanOnayladi
			case models.DurumDekanOnayladi:
				yeniDurum, err = s.GetNextWorkflowStatus(projeID, p.DurumAdi)
			case models.DurumKomisyonOnayladi:
				yeniDurum, err = s.GetNextWorkflowStatus(projeID, p.DurumAdi)
			case models.DurumHakemAtamaBekliyor:
				yeniDurum, err = s.GetNextWorkflowStatus(projeID, p.DurumAdi)
			case models.DurumHakemBekliyor:
				yeniDurum = models.DurumHakemOnayladi
			case models.DurumHakemOnayladi:
				yeniDurum, err = s.GetNextWorkflowStatus(projeID, p.DurumAdi)
			case models.DurumSozlesmeDolduruldu:
				// Türkçe Yorum: Akademisyen sözleşmeyi doldurduktan sonra TTO projeyi yürürlüğe alır.
				yeniDurum = models.DurumYururlukte
			case models.DurumSozlesmeImza:
				// Türkçe Yorum: Akademisyen henüz sözleşme formunu doldurmadıysa proje yürürlüğe alınamaz.
				var count int
				errSoz := s.ProjeRepo.DB.QueryRow(`SELECT COUNT(*) FROM proje_sozlesme WHERE proje_id = $1`, projeID).Scan(&count)
				if errSoz != nil || count == 0 {
					return fmt.Errorf("akademisyen henüz proje sözleşmesini doldurmadığı için proje yürürlüğe alınamaz")
				}
				yeniDurum = models.DurumYururlukte
			default:
				return fmt.Errorf("bu durum için onay süreci işletilemez: %s", p.DurumAdi)
			}
			if err != nil {
				return fmt.Errorf("sonraki süreç durumu hesaplanamadı: %w", err)
			}
		} else {
			return fmt.Errorf("geçersiz aksiyon: %s", action)
		}
	}

	// Türkçe Yorum: Durum değişmeyecekse (komisyon henüz tamamlanmadı) sadece güncelleme loglamadan çık
	if yeniDurum == p.DurumAdi {
		return nil
	}

	// Türkçe Yorum: Komisyon_bekliyor durumuna ilk kez geçiliyorsa komisyon oylama kayıtları açılır
	if yeniDurum == models.DurumKomisyonBekliyor && p.DurumAdi != models.DurumKomisyonBekliyor {
		if err := s.ProjeRepo.CreateKomisyonOnayRecords(projeID); err != nil {
			return fmt.Errorf("komisyon oylama kayıtları oluşturulamadı: %w", err)
		}
	}

	err = s.ProjeRepo.UpdateProjectStatusWithLog(projeID, islemYapanID, p.DurumAdi, yeniDurum, aciklama)
	if err != nil {
		return fmt.Errorf("proje durum güncellenirken hata oluştu: %w", err)
	}

	return nil
}

// GetProjeSurecGecmisi projenin geçmiş durum değişikliklerini listeler.
// Türkçe Bilgilendirme: İstek yapan rolünün yetkisine göre maskeleme parametresini repository'ye iletir.
func (s *ProjeService) GetProjeSurecGecmisi(projeID int, isAdminOrTTO bool) ([]models.ProjeSurecGecmisi, error) {
	return s.ProjeRepo.GetProjeSurecGecmisi(projeID, isAdminOrTTO)
}

// GetWorkflowHistoryByUyeID belirli bir kullanıcının geçmiş onay kararlarını çeker.
func (s *ProjeService) GetWorkflowHistoryByUyeID(uyeID int) ([]models.ProjeSurecGecmisi, error) {
	return s.ProjeRepo.GetWorkflowHistoryByUyeID(uyeID)
}

// ttoPanelWorkflowDurumlari TTO yönetim panelinde listelenmesi gereken proje durumlarını döner.
// Türkçe Yorum: Admin ve TTO aynı paneli kullandığı için her iki rol de aynı durum kümesini görmelidir.
func ttoPanelWorkflowDurumlari() []string {
	return []string{
		models.DurumIncelemede, models.DurumDekanOnayiBekliyor,
		models.DurumDekanOnayladi, models.DurumKomisyonBekliyor,
		models.DurumKomisyonOnayladi, models.DurumHakemAtamaBekliyor,
		models.DurumHakemBekliyor, models.DurumHakemOnayladi,
		models.DurumSozlesmeImza, models.DurumSozlesmeDolduruldu, models.DurumTTOAktif,
		models.DurumYururlukte, models.DurumReddedildi,
		models.DurumRevizyon, models.DurumTamamlandi,
	}
}

// GetProjectsForWorkflow rol bazında onay bekleyen projeleri listeler.
// Türkçe Yorum: Kullanıcının sahip olduğu tüm rollere göre onay bekleyen projeleri çeker ve tekil olarak birleştirir.
func (s *ProjeService) GetProjectsForWorkflow(rol string, uyeID int) ([]models.Proje, error) {
	roles := strings.Split(rol, ",")
	var allProjects []models.Proje
	seen := make(map[int]bool)

	// appendUniq: tekrar eden proje ID'lerini filtreleyen yardımcı fonksiyon
	appendUniq := func(projeler []models.Proje) {
		for _, p := range projeler {
			if !seen[p.ProjeID] {
				seen[p.ProjeID] = true
				allProjects = append(allProjects, p)
			}
		}
	}

	for _, r := range roles {
		r = strings.ToLower(strings.TrimSpace(r))

		// Admin ve TTO rolleri TTO panelinde taslak harici tüm süreçteki projeleri tek sorguda görebilir.
		if r == models.RolAdmin || r == models.RolTTO || strings.HasPrefix(r, "tto") || strings.HasPrefix(r, "admin") {
			projeler, err := s.ProjeRepo.GetProjectsForWorkflow(r, "", 0)
			if err != nil {
				fmt.Printf("GetProjectsForWorkflow hatası (%s rolü): %v\n", r, err)
			}
			appendUniq(projeler)
			continue
		}

		// Türkçe Yorum: Hakem rolü için hakem_bekliyor durumundaki projeleri gösterir
		if r == models.RolHakem {
			projeler, err := s.ProjeRepo.GetProjectsForWorkflow(r, models.DurumHakemBekliyor, 0)
			if err != nil {
				fmt.Printf("GetProjectsForWorkflow hatası (hakem rolü): %v\n", err)
			}
			appendUniq(projeler)
			continue
		}

		// Türkçe Yorum: Diğer roller için rol→durum eşlemesi
		rolDurumMap := map[string]string{
			models.RolDekan:             models.DurumDekanOnayiBekliyor,
			models.RolKomisyon:          models.DurumKomisyonBekliyor,
			models.RolKomisyonBaskani:   models.DurumKomisyonBekliyor,
			models.RolKomisyonRaportoru: models.DurumKomisyonBekliyor,
		}

		durum, ok := rolDurumMap[r]
		if !ok {
			continue
		}

		projeler, err := s.ProjeRepo.GetProjectsForWorkflow(r, durum, uyeID)
		if err != nil {
			fmt.Printf("GetProjectsForWorkflow hatası (%s rolü, %s durumu): %v\n", r, durum, err)
		}
		appendUniq(projeler)
	}

	// Eğer özel bir filtreyle eşleşen proje bulunamadıysa (veya genel erişimde ise) fallback olarak süreçteki tüm projeler çekilir
	if len(allProjects) == 0 {
		projeler, err := s.ProjeRepo.GetProjectsForWorkflow("tto", "", 0)
		if err == nil && len(projeler) > 0 {
			appendUniq(projeler)
		}
	}

	// Türkçe Yorum: Ön yüzün sevk butonlarını sabit sıraya göre değil iş akışına göre çizebilmesi için
	// her projenin sonraki aşaması hesaplanıp listeye eklenir.
	s.enrichSonrakiAsama(allProjects)
	s.enrichKomisyonToplantiBilgi(allProjects)

	return allProjects, nil
}

// enrichKomisyonToplantiBilgi komisyon_bekliyor projelerin gündemdeki toplantı bilgisini ekler.
// Türkçe Yorum: Alt listedeki "toplantı kararı bekleniyor" rozetinin hangi toplantıya ait olduğunu gösterir.
func (s *ProjeService) enrichKomisyonToplantiBilgi(projeler []models.Proje) {
	for i := range projeler {
		if projeler[i].DurumAdi != models.DurumKomisyonBekliyor {
			continue
		}
		var toplantiNo, karar string
		err := s.ProjeRepo.DB.QueryRow(`
			SELECT kt.toplanti_no, ktp.karar
			FROM komisyon_toplanti_proje ktp
			JOIN komisyon_toplantisi kt ON kt.toplanti_id = ktp.toplanti_id
			WHERE ktp.proje_id = $1
			  AND ktp.karar IN ('bekliyor', 'ertelendi')
			ORDER BY ktp.id DESC
			LIMIT 1
		`, projeler[i].ProjeID).Scan(&toplantiNo, &karar)
		if err != nil {
			continue
		}
		projeler[i].KomisyonToplantiNo = toplantiNo
		projeler[i].KomisyonToplantiKarar = karar
	}
}

// enrichSonrakiAsama listedeki projelere iş akışındaki sonraki aşama bilgisini ekler.
// Türkçe Yorum: Karar beklenen veya süreci bitmiş durumlar için sonraki aşama hesaplanmaz.
func (s *ProjeService) enrichSonrakiAsama(projeler []models.Proje) {
	// Sevk kararı verilemeyen (bekleyen ya da nihai) durumlar
	pasifDurumlar := map[string]bool{
		models.DurumTaslak:              true,
		models.DurumDekanOnayiBekliyor:  true,
		models.DurumKomisyonBekliyor:    true,
		models.DurumHakemBekliyor:       true,
		models.DurumYururlukte:          true,
		models.DurumTamamlandi:          true,
		models.DurumReddedildi:          true,
		models.DurumRevizyon:            true,
	}

	for i := range projeler {
		if pasifDurumlar[projeler[i].DurumAdi] {
			continue
		}
		sonraki, err := s.GetSonrakiAsama(projeler[i].ProjeID, projeler[i].DurumAdi)
		if err != nil {
			continue
		}
		if sonraki == nil {
			// Süreç sonu: proje yürürlüğe alınır
			projeler[i].SonrakiDurumAdi = models.DurumYururlukte
			continue
		}
		projeler[i].SonrakiAsamaKodu = sonraki.AsamaKodu
		projeler[i].SonrakiAsamaAdi = sonraki.AsamaAdi
		projeler[i].SonrakiDurumAdi = sonraki.DurumAdi
	}
}

// GetButceKategorileri bütçe kategorilerini döner.
// Türkçe Bilgilendirme: ProjeRepository'den bütçe kategorilerini alıp handler'a iletir.
func (s *ProjeService) GetButceKategorileri() ([]models.ButceKategori, error) {
	return s.ProjeRepo.GetButceKategorileri()
}

// GetSistemRolleri sistemdeki tüm rolleri döner.
// Türkçe Bilgilendirme: ProjeRepository'den sistem rollerini alıp handler'a iletir.
func (s *ProjeService) GetSistemRolleri() ([]models.SistemRolTanimlama, error) {
	return s.ProjeRepo.GetSistemRolleri()
}

// GetProjeDurumlari proje durum tanımlarını döner.
// Türkçe Bilgilendirme: ProjeRepository'den proje durum tanımlarını alıp handler'a iletir.
func (s *ProjeService) GetProjeDurumlari() ([]models.ProjeDurumTanim, error) {
	return s.ProjeRepo.GetProjeDurumlari()
}
