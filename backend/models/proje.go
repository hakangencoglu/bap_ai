package models

import "time"

// Proje yapısı, veritabanındaki proje tablosunun Go karşılığıdır.
type Proje struct {
	ProjeID          int       `json:"proje_id"`
	ProjeKodu        string    `json:"proje_kodu"`
	BaslikTr         string    `json:"baslik_tr"`
	BaslikEn         string    `json:"baslik_en"`
	SureAy           int       `json:"sure_ay"`
	ToplamButce      float64   `json:"toplam_butce"`
	EtikKurul        bool      `json:"etik_kurul"`
	EtikKurulNo      *int      `json:"etik_kurul_no"`       // Nullable
	KoordinatorID    *int      `json:"koordinator_id"`      // Nullable FK → uye
	DurumID          *int      `json:"durum_id"`            // Nullable FK → proje_durum (genel durum)
	AsamaID          *int      `json:"asama_id"`            // Nullable FK → proje_asama (onay akışı aşaması)
	BapTuruID        *int      `json:"bap_turu_id"`         // Nullable FK → proje_bap_turu
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	GuncellemeTarihi time.Time `json:"guncelleme_tarihi"`
	PdfDosyaYolu *string   `json:"pdf_dosya_yolu"` // Onaylanan PDF'in sunucu dosya yolu
	// Aşağıdaki alanlar JOIN ile doldurulabilir, DB'de ayrı tablolarda tutulur
	DurumAdi           string `json:"durum_adi,omitempty"`           // proje_durum tablosundan gelir
	AsamaAdi           string `json:"asama_adi,omitempty"`           // proje_asama tablosundan gelir (Dekan Onayına Sun vb.)
	AsamaKodu          string `json:"asama_kodu,omitempty"`          // proje_asama.asama_kodu (dahili kod)
	BapTuru            string `json:"bap_turu,omitempty"`            // proje_bap_turu tablosundan gelir
	KoordinatorAdSoyad string `json:"koordinator_ad_soyad,omitempty"` // uye tablosundan koordinator "Unvan Ad Soyad" veya "Ad Soyad"
	
	// Akademik Detaylar (proje_detay tablosundan JOIN ile gelir)
	Ozet                string `json:"ozet,omitempty"`
	OzetEn              string `json:"ozet_en,omitempty"`
	AnahtarKelimeler    string `json:"anahtar_kelimeler,omitempty"`
	AnahtarKelimelerEn  string `json:"anahtar_kelimeler_en,omitempty"`
	Hedefler            string `json:"hedefler,omitempty"`
	Ozgunluk            string `json:"ozgunluk,omitempty"`
	Metodoloji          string `json:"metodoloji,omitempty"`
	Kaynakca            string `json:"kaynakca,omitempty"`
}


// ProjeDetay yapısı, projenin akademik detay bilgilerini tutar.
type ProjeDetay struct {
	ProjeID              int    `json:"proje_id"`
	Ozet                 string `json:"ozet"`
	OzetEn               string `json:"ozet_en"`
	AnahtarKelimeler     string `json:"anahtar_kelimeler"`
	AnahtarKelimelerEn   string `json:"anahtar_kelimeler_en"`
	Hedefler             string `json:"hedefler"`
	Ozgunluk             string `json:"ozgunluk"`
	Metodoloji           string `json:"metodoloji"`
	Kaynakca             string `json:"kaynakca"`
}

// ProjeYayinEtki, projeden elde edilmesi öngörülen çıktıları tutar (Yaygın Etki bölümü 6a).
type ProjeYayinEtki struct {
	ID           int    `json:"id"`
	ProjeID      int    `json:"proje_id"`
	CiktiTuru    string `json:"cikti_turu"`
	OngorulCikti string `json:"ongorul_cikti"`
	ZamanAraligi string `json:"zaman_araligi"`
}

// ProjeYayginlastirmaEtkinlik, çıktıların paylaşımı ve yaygınlaştırılması etkinliklerini tutar (bölüm 6b).
type ProjeYayginlastirmaEtkinlik struct {
	ID           int    `json:"id"`
	ProjeID      int    `json:"proje_id"`
	EtkinlikTuru string `json:"etkinlik_turu"`
	Paydas       string `json:"paydas"`
	ZamanSure    string `json:"zaman_sure"`
	SiraNo       int    `json:"sira_no"`
}

// ProjeTakim yapısı, proje takım üyesi bilgisini tutar.
type ProjeTakim struct {
	ProjeID    int    `json:"proje_id"`
	UyeID      int    `json:"uye_id"`
	ProjeRolID *int   `json:"proje_rol_id"`
	// JOIN ile doldurulacak alanlar
	AdTumu   string `json:"ad_tumu,omitempty"`   // "Ad Soyad"
	ProjeRol string `json:"proje_rol,omitempty"` // proje_rol_tanimlama tablosundan
}

// ProjeYayinlastirma yapısı, proje yayınlaştırma bilgilerini tutar.
type ProjeYayinlastirma struct {
	ProjeID             int    `json:"proje_id"`
	YayinTuru           string `json:"yayin_turu"`
	YayinCiktisi        string `json:"yayin_ciktisi"`
	TahminiYayinTarihi  string `json:"tahmini_yayin_tarihi"`
}

// ProjeCikti yapısı, projenin beklenen çıktılarını tutar.
type ProjeCikti struct {
	CiktiID      int    `json:"cikti_id"`
	ProjeID      int    `json:"proje_id"`
	CiktiTuruID  *int   `json:"cikti_turu_id"`
	Aciklama     string `json:"aciklama"`
	CiktiPeriyodu string `json:"cikti_periyodu"`
	// JOIN ile doldurulacak alan
	CiktiTuru string `json:"cikti_turu,omitempty"` // proje_cikti_turu tablosundan
}

// DashboardStats yapısı, dashboard sayfasında gösterilecek istatistikleri tutar.
type DashboardStats struct {
	AktifProje   int     `json:"aktif_proje"`
	OnayBekleyen int     `json:"onay_bekleyen"`
	Tamamlanan   int     `json:"tamamlanan"`
	ToplamButce  float64 `json:"toplam_butce"`
}


// ProjeOzet yapısı, dashboard'daki son başvurular tablosu için özet proje bilgisi tutar.
type ProjeOzet struct {
	ProjeID   int    `json:"proje_id"`
	ProjeKodu string `json:"proje_kodu"`
	BaslikTr  string `json:"baslik_tr"`
	BapTuru  string `json:"bap_turu"`
	Tarih    string `json:"tarih"`
	DurumAdi string `json:"durum_adi"`
}

// ProfilProjeBilgisi yapısı, profil sayfasındaki proje kartları için bilgi tutar.
type ProfilProjeBilgisi struct {
	ProjeID   int    `json:"proje_id"`
	ProjeKodu string `json:"proje_kodu"`
	BaslikTr  string `json:"baslik_tr"`
	BapTuru  string `json:"bap_turu"`
	DurumAdi string `json:"durum_adi"`
	UyeRol   string `json:"uye_rol"`
}

// ProjeUye yapısı, projeye kayıtlı üyelerin modal vs işlemlerde listelenmesi için oluşturuldu.
type ProjeUye struct {
	UyeID    int    `json:"uye_id"`
	AdTumu   string `json:"ad_tumu"`    // "Ad Soyad"
	Rol      string `json:"rol"`        // Sistemdeki rolü
	ProjeRol string `json:"proje_rol"`  // Projedeki rolü (Yürütücü, Araştırmacı vb.)
}

// ProjeSurecGecmisi projenin durum değişikliklerini ve onay geçmişini tutar.
type ProjeSurecGecmisi struct {
	GecmisID         int       `json:"gecmis_id"`
	ProjeID          int       `json:"proje_id"`
	IslemYapanID     int       `json:"islem_yapan_id"`
	BaslangicDurum   string    `json:"baslangic_durum"`
	HedefDurum       string    `json:"hedef_durum"`
	Aciklama         string    `json:"aciklama"`
	OlusturmaTarihi  time.Time `json:"olusturma_tarihi"`
	// JOIN ile doldurulacak alanlar
	IslemYapanAdTumu string    `json:"islem_yapan_ad_tumu,omitempty"` // İşlemi yapan üyenin "Ad Soyad" bilgisi
	IslemYapanUnvan  string    `json:"islem_yapan_unvan,omitempty"`   // İşlemi yapan üyenin unvanı (Prof. Dr. vb.)
	ProjeKodu        string    `json:"proje_kodu,omitempty"`          // İşlemin ait olduğu projenin kodu
	ProjeBaslik      string    `json:"proje_baslik,omitempty"`        // İşlemin ait olduğu projenin başlığı
}

