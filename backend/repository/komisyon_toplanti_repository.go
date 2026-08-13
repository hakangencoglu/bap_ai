package repository

import (
	"database/sql"
	"fmt"

	"bap_ai/backend/models"
)

// KomisyonToplantiRepository toplantı ↔ proje köprü tablo işlemlerini yürütür.
// Türkçe Yorum: Hangi projenin hangi toplantıda görüşüldüğünü ve kararlarını yönetir.
type KomisyonToplantiRepository struct {
	DB *sql.DB
}

// NewKomisyonToplantiRepository yeni bir KomisyonToplantiRepository oluşturur.
func NewKomisyonToplantiRepository(db *sql.DB) *KomisyonToplantiRepository {
	return &KomisyonToplantiRepository{DB: db}
}

// AddProjeToToplanti bir projeyi belirtilen toplantıya gündem maddesi olarak ekler.
// Türkçe Yorum: Raportör, komisyon_bekliyor durumundaki projeleri aktif toplantıya ekler.
func (r *KomisyonToplantiRepository) AddProjeToToplanti(toplantiID, projeID, gundemSirasi, ekleyenID int) error {
	query := `
		INSERT INTO komisyon_toplanti_proje (toplanti_id, proje_id, gundem_sirasi, ekleyen_id)
		VALUES ($1, $2, NULLIF($3, 0), $4)
		ON CONFLICT (toplanti_id, proje_id) DO UPDATE SET
			gundem_sirasi = EXCLUDED.gundem_sirasi,
			ekleyen_id   = EXCLUDED.ekleyen_id
	`
	_, err := r.DB.Exec(query, toplantiID, projeID, gundemSirasi, ekleyenID)
	if err != nil {
		return fmt.Errorf("proje toplantıya eklenemedi: %w", err)
	}
	return nil
}

// RemoveProjeFromToplanti bir projeyi toplantıdan çıkarır.
// Türkçe Yorum: Sadece 'bekliyor' durumundaki proje gündem maddelerini kaldırır.
func (r *KomisyonToplantiRepository) RemoveProjeFromToplanti(toplantiID, projeID int) error {
	query := `DELETE FROM komisyon_toplanti_proje WHERE toplanti_id = $1 AND proje_id = $2 AND karar = 'bekliyor'`
	res, err := r.DB.Exec(query, toplantiID, projeID)
	if err != nil {
		return fmt.Errorf("proje toplantıdan çıkarılamadı: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("proje bulunamadı veya karar verilmiş projeler kaldırılamaz")
	}
	return nil
}

// GetProjectsByToplanti bir toplantıdaki tüm projeleri detaylarıyla getirir.
// Türkçe Yorum: Toplantı detay modalı ve PDF tutanağı için kullanılır.
func (r *KomisyonToplantiRepository) GetProjectsByToplanti(toplantiID int) ([]*models.KomisyonToplantisiProje, error) {
	query := `
		SELECT
			ktp.id, ktp.toplanti_id, ktp.proje_id,
			ktp.gundem_sirasi, ktp.karar, COALESCE(ktp.karar_aciklamasi,''),
			ktp.karar_tarihi, ktp.ekleyen_id, ktp.olusturma_tarihi,
			COALESCE(p.proje_kodu,''),
			COALESCE((SELECT pb.baslik FROM proje_baslik pb WHERE pb.proje_id=p.proje_id AND pb.dil_kodu='tr' LIMIT 1),'Başlıksız'),
			COALESCE(u.ad||' '||u.soyad,'Bilinmiyor'),
			COALESCE(NULLIF(TRIM(COALESCE(ud.unvan,'')), ''), COALESCE(NULLIF(TRIM(COALESCE(u.unvan,'')), ''), '')),
			COALESCE(pd.durum_adi,'')
		FROM komisyon_toplanti_proje ktp
		JOIN proje p    ON p.proje_id = ktp.proje_id
		JOIN uye u      ON u.uye_id  = p.koordinator_id
		LEFT JOIN uye_detay ud ON ud.uye_id = u.uye_id
		LEFT JOIN proje_durum pd ON pd.durum_id = p.durum_id
		WHERE ktp.toplanti_id = $1
		ORDER BY COALESCE(ktp.gundem_sirasi, 9999), ktp.olusturma_tarihi
	`
	rows, err := r.DB.Query(query, toplantiID)
	if err != nil {
		return nil, fmt.Errorf("toplantı projeleri sorgulanamadı: %w", err)
	}
	defer rows.Close()

	var projeler []*models.KomisyonToplantisiProje
	for rows.Next() {
		var kp models.KomisyonToplantisiProje
		err := rows.Scan(
			&kp.ID, &kp.ToplantiID, &kp.ProjeID,
			&kp.GundemSirasi, &kp.Karar, &kp.KararAciklamasi,
			&kp.KararTarihi, &kp.EkleyenID, &kp.OlusturmaTarihi,
			&kp.ProjeKodu, &kp.ProjeBaslik, &kp.YurutucuAd, &kp.YurutucuUnvan, &kp.MevcutDurum,
		)
		if err != nil {
			return nil, err
		}
		projeler = append(projeler, &kp)
	}
	return projeler, nil
}

// GetToplantilerByProje bir projenin görüşüldüğü tüm toplantıları listeler.
func (r *KomisyonToplantiRepository) GetToplantilerByProje(projeID int) ([]*models.KomisyonToplantisiProje, error) {
	query := `
		SELECT
			ktp.id, ktp.toplanti_id, ktp.proje_id,
			ktp.gundem_sirasi, ktp.karar, COALESCE(ktp.karar_aciklamasi,''),
			ktp.karar_tarihi, ktp.ekleyen_id, ktp.olusturma_tarihi,
			COALESCE(kt.toplanti_no,''), '', '', ''
		FROM komisyon_toplanti_proje ktp
		JOIN komisyon_toplantisi kt ON kt.toplanti_id = ktp.toplanti_id
		WHERE ktp.proje_id = $1
		ORDER BY kt.tarih DESC
	`
	rows, err := r.DB.Query(query, projeID)
	if err != nil {
		return nil, fmt.Errorf("projenin toplantıları sorgulanamadı: %w", err)
	}
	defer rows.Close()

	var liste []*models.KomisyonToplantisiProje
	for rows.Next() {
		var kp models.KomisyonToplantisiProje
		err := rows.Scan(
			&kp.ID, &kp.ToplantiID, &kp.ProjeID,
			&kp.GundemSirasi, &kp.Karar, &kp.KararAciklamasi,
			&kp.KararTarihi, &kp.EkleyenID, &kp.OlusturmaTarihi,
			&kp.ProjeKodu, &kp.ProjeBaslik, &kp.YurutucuAd, &kp.MevcutDurum,
		)
		if err != nil {
			return nil, err
		}
		liste = append(liste, &kp)
	}
	return liste, nil
}

// SetProjeKarar toplantıdaki bir proje için karar kaydeder.
// Türkçe Yorum: Raportör onay/red/erteleme kararını bu fonksiyon aracılığıyla kayıt altına alır.
func (r *KomisyonToplantiRepository) SetProjeKarar(toplantiID, projeID int, karar, aciklama string) error {
	query := `
		UPDATE komisyon_toplanti_proje
		SET karar = $3, karar_aciklamasi = $4, karar_tarihi = CURRENT_TIMESTAMP
		WHERE toplanti_id = $1 AND proje_id = $2
	`
	res, err := r.DB.Exec(query, toplantiID, projeID, karar, aciklama)
	if err != nil {
		return fmt.Errorf("proje kararı kaydedilemedi: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("proje bu toplantıda bulunamadı")
	}
	return nil
}

// SyncToplantiDurumFromProjeler gündem projelerinin karar durumuna göre toplantı durumunu günceller.
// Türkçe Yorum: Bekleyen/ertelenmiş proje varsa planli; tüm nihai kararlar verildiyse tamamlandi.
func (r *KomisyonToplantiRepository) SyncToplantiDurumFromProjeler(toplantiID int) (string, error) {
	var toplam, bekleyen int
	err := r.DB.QueryRow(`
		SELECT
			COUNT(*)::int,
			COUNT(*) FILTER (WHERE karar IN ('bekliyor', 'ertelendi'))::int
		FROM komisyon_toplanti_proje
		WHERE toplanti_id = $1
	`, toplantiID).Scan(&toplam, &bekleyen)
	if err != nil {
		return "", fmt.Errorf("toplantı proje durumları okunamadı: %w", err)
	}
	if toplam == 0 {
		return "", nil
	}

	yeniDurum := "tamamlandi"
	if bekleyen > 0 {
		yeniDurum = "planli"
	}

	_, err = r.DB.Exec(`
		UPDATE komisyon_toplantisi
		SET durum = $2
		WHERE toplanti_id = $1
	`, toplantiID, yeniDurum)
	if err != nil {
		return "", fmt.Errorf("toplantı durumu güncellenemedi: %w", err)
	}
	return yeniDurum, nil
}

// GetToplantiBelgeDetay PDF tutanağı için toplantı + katılımcılar + projeler bütününü döner.
func (r *KomisyonToplantiRepository) GetToplantiBelgeDetay(toplantiID int) (*models.KomisyonToplantiBelge, error) {
	// Türkçe Yorum: Toplantı genel bilgisi mevcut KomisyonRepository üzerinden alınır.
	komisyonRepo := &KomisyonRepository{DB: r.DB}
	toplanti, err := komisyonRepo.GetMeetingByID(toplantiID)
	if err != nil {
		return nil, err
	}
	if toplanti == nil {
		return nil, fmt.Errorf("toplantı bulunamadı")
	}
	projeler, err := r.GetProjectsByToplanti(toplantiID)
	if err != nil {
		return nil, err
	}
	return &models.KomisyonToplantiBelge{
		Toplanti:     toplanti,
		Katilimcilar: toplanti.Katilimcilar,
		Projeler:     projeler,
	}, nil
}

// GetBekleyenProjeler komisyon_bekliyor durumundaki projeleri listeler.
// Türkçe Yorum: Toplantı gündemine eklenecek aday projeleri döner.
func (r *KomisyonToplantiRepository) GetBekleyenProjeler() ([]*models.KomisyonBekleyenProje, error) {
	query := `
		SELECT
			p.proje_id,
			COALESCE(p.proje_kodu, ''),
			COALESCE((SELECT pb.baslik FROM proje_baslik pb WHERE pb.proje_id = p.proje_id AND pb.dil_kodu = 'tr' LIMIT 1), 'Başlıksız'),
			COALESCE(u.ad||' '||u.soyad, 'Bilinmiyor'),
			COALESCE(NULLIF(TRIM(COALESCE(ud.unvan,'')), ''), COALESCE(NULLIF(TRIM(COALESCE(u.unvan,'')), ''), '')),
			COALESCE(pbt.bap_turu, ''),
			COALESCE((SELECT SUM(b.toplam_fiyat) FROM proje_butce b WHERE b.proje_id = p.proje_id), 0)
		FROM proje p
		JOIN proje_durum pd ON pd.durum_id = p.durum_id
		JOIN uye u ON u.uye_id = p.koordinator_id
		LEFT JOIN uye_detay ud ON ud.uye_id = u.uye_id
		LEFT JOIN proje_bap_turu pbt ON pbt.bap_turu_id = p.bap_turu_id
		WHERE pd.durum_adi = 'komisyon_bekliyor'
		ORDER BY p.proje_id DESC
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("bekleyen projeler sorgulanamadı: %w", err)
	}
	defer rows.Close()

	var liste []*models.KomisyonBekleyenProje
	for rows.Next() {
		var p models.KomisyonBekleyenProje
		if err := rows.Scan(&p.ProjeID, &p.ProjeKodu, &p.ProjeBaslik, &p.YurutucuAd, &p.YurutucuUnvan, &p.BapTuru, &p.ToplamButce); err != nil {
			return nil, err
		}
		liste = append(liste, &p)
	}
	return liste, nil
}
