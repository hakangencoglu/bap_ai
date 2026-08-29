package models

import "time"

// SistemRolTanimlama yapısı, sistem rollerinin tanımlarını tutar.
type SistemRolTanimlama struct {
	RolID      int    `json:"rol_id"`
	RolAdi     string `json:"rol_adi"`
	RolEtiketi string `json:"rol_etiketi"`
}

// SistemRol yapısı, kullanıcı-rol atama ilişkisini temsil eder.
type SistemRol struct {
	RolID       int `json:"rol_id"`
	UyeID       int `json:"uye_id"`
	SistemRolID int `json:"sistem_rol_id"`
}

// ProjeRolTanimlama yapısı, proje içindeki rol tanımlarını tutar.
type ProjeRolTanimlama struct {
	RolID    int    `json:"rol_id"`
	ProjeRol string `json:"proje_rol"`
}

// ProjeDurumTanim yapısı, proje durumlarının tanımlarını tutar.
type ProjeDurumTanim struct {
	DurumID      int    `json:"durum_id"`
	DurumAdi     string `json:"durum_adi"`
	DurumEtiketi string `json:"durum_etiketi"`
}

// BAP türü versiyon durum sabitleri
const (
	BapVersiyonTaslak  = "taslak"
	BapVersiyonYayinda = "yayinda"
	BapVersiyonArsiv   = "arsiv"
)

// ProjeBapTuru yapısı, BAP proje türü kimliği + düzenlenen/görünen kural setini tutar.
// Türkçe Yorum: Kural alanları taslak veya yayındaki versiyondan gelir; projeler bap_turu_versiyon_id ile kilitlenir.
type ProjeBapTuru struct {
	BapTuruID       int     `json:"bap_turu_id"`
	BapTuru         string  `json:"bap_turu"`
	ButceLimiti     float64 `json:"butce_limiti"`
	SureLimitiAy    int     `json:"sure_limiti_ay"`
	AktifMi         bool    `json:"aktif_mi"`
	Aciklama        string  `json:"aciklama"`
	HakemGerekli    bool    `json:"hakem_gerekli"`
	HakemSayisi     int     `json:"hakem_sayisi"`
	BursiyerGerekli bool    `json:"bursiyer_gerekli"`
	BursiyerSayisi  int     `json:"bursiyer_sayisi"`
	AraRaporGerekli bool    `json:"ara_rapor_gerekli"`
	AraRaporSayisi  int     `json:"ara_rapor_sayisi"`
	AsamaIDs        []int   `json:"asama_ids"`

	// İZÜ BAP Yönergesi Madde 8 & 9 Yönetici Dinamik Kural Ayarları
	HakemTuruKisitlama       string  `json:"hakem_turu_kisitlama"`        // herhangi | kurum_ici | kurum_disi
	HakemSureGun             int     `json:"hakem_sure_gun"`              // Hakem değerlendirme süresi (gün) - varsayılan 15
	HakemUcretOraniYuzde     float64 `json:"hakem_ucret_orani_yuzde"`     // Net asgari ücret yüzdesi (%3.00 veya %5.00)
	BursiyerIzinliMi         bool    `json:"bursiyer_izinli_mi"`          // Bursiyer görevlendirilebilir mi?
	MaxAktifProjeSayisi      int     `json:"max_aktif_proje_sayisi"`       // Yürütücü max aktif proje (0 = limitsiz)
	TezOgrencisiSarti        bool    `json:"tez_ogrencisi_sarti"`         // Tez öğrencisi şartı var mı?
	YayinGecmisSarti         bool    `json:"yayin_gecmis_sarti"`          // 2 yıl içinde yayın/başvuru geçmiş şartı
	IntihalCezasiAktif       bool    `json:"intihal_cezasi_aktif"`        // İntihale 3 yıl başvuru engeli cezası aktif mi
	YurutucuGecmisProjeSarti bool    `json:"yurutucu_gecmis_proje_sarti"` // Geçmiş ulusal/uluslararası proje yürütücülük şartı
	IzinSeyahatBeyaniZorunlu bool    `json:"izin_seyahat_beyani_zorunlu"` // İzin/Seyahat durumu beyanı zorunlu mu
	FirmaOrtaklikBeyaniZorunlu bool  `json:"firma_ortaklik_beyani_zorunlu"`// Firma sahipliği/ortaklığı olmama beyanı
	MinKurumHissesiOrani     float64 `json:"min_kurum_hissesi_orani"`     // Minimum Kurum Hissesi Oranı (%)

	// Versiyon meta (admin / proje bağlama)
	VersiyonID              *int                    `json:"versiyon_id,omitempty"`
	VersiyonNo              *int                    `json:"versiyon_no,omitempty"`
	VersiyonDurum           string                  `json:"versiyon_durum,omitempty"` // taslak | yayinda | arsiv
	YayinVersiyonNo         *int                    `json:"yayin_versiyon_no,omitempty"`
	YayinVersiyonID         *int                    `json:"yayin_versiyon_id,omitempty"`
	TaslakVarMi             bool                    `json:"taslak_var_mi"`
	YayinlanmamisDegisiklik bool                    `json:"yayinlanmamis_degisiklik"`
	FormAlanlari            []ProjeBapTuruFormAlani `json:"form_alanlari,omitempty"`
}

// ProjeBapTuruFormAlani, BAP proje türü için tanımlı başvuru form alanını temsil eder.
type ProjeBapTuruFormAlani struct {
	AlanID        int       `json:"alan_id"`
	BapTuruID     int       `json:"bap_turu_id"`
	AlanKodu      string    `json:"alan_kodu"`       // Örn: ozgun_deger, ozel_sart_1
	Etiket        string    `json:"etiket"`          // Örn: Projenin Özgün Değeri
	Bolum         string    `json:"bolum"`           // Örn: Temel Bilgiler, Akademik İçerik, Ek Belgeler
	AracTuru      string    `json:"arac_turu"`       // input, number, textarea, texteditor, dropdownlist, checkbox, radio, datepicker, file
	Secenekler    string    `json:"secenekler"`      // Dropdown/Radio için seçenekler
	ZorunluMu     bool      `json:"zorunlu_mu"`      // Zorunlu alan mı?
	AktifMi       bool      `json:"aktif_mi"`        // Formda aktif gösterilsin mi?
	Ipucu         string    `json:"ipucu"`           // Placeholder veya yardım açıklaması
	SiraNo        int       `json:"sira_no"`         // Formdaki görüntülenme sırası
	SistemAlaniMi bool      `json:"sistem_alani_mi"` // Temel sistem alanı mı
	OlusturmaTarihi time.Time `json:"olusturma_tarihi,omitempty"`
}

// ProjeBapTuruVersiyon, BAP türünün donmuş kural anlık görüntüsüdür.
type ProjeBapTuruVersiyon struct {
	VersiyonID      int        `json:"versiyon_id"`
	BapTuruID       int        `json:"bap_turu_id"`
	VersiyonNo      *int       `json:"versiyon_no,omitempty"`
	Durum           string     `json:"durum"`
	ButceLimiti     float64    `json:"butce_limiti"`
	SureLimitiAy    int        `json:"sure_limiti_ay"`
	Aciklama        string     `json:"aciklama"`
	HakemGerekli    bool       `json:"hakem_gerekli"`
	HakemSayisi     int        `json:"hakem_sayisi"`
	BursiyerGerekli bool       `json:"bursiyer_gerekli"`
	BursiyerSayisi  int        `json:"bursiyer_sayisi"`
	AraRaporGerekli bool       `json:"ara_rapor_gerekli"`
	AraRaporSayisi  int        `json:"ara_rapor_sayisi"`
	AsamaIDs        []int      `json:"asama_ids"`

	// İZÜ BAP Yönergesi Madde 8 & 9 Yönetici Dinamik Kural Ayarları
	HakemTuruKisitlama       string  `json:"hakem_turu_kisitlama"`
	HakemSureGun             int     `json:"hakem_sure_gun"`
	HakemUcretOraniYuzde     float64 `json:"hakem_ucret_orani_yuzde"`
	BursiyerIzinliMi         bool    `json:"bursiyer_izinli_mi"`
	MaxAktifProjeSayisi      int     `json:"max_aktif_proje_sayisi"`
	TezOgrencisiSarti        bool    `json:"tez_ogrencisi_sarti"`
	YayinGecmisSarti         bool    `json:"yayin_gecmis_sarti"`
	IntihalCezasiAktif       bool    `json:"intihal_cezasi_aktif"`
	YurutucuGecmisProjeSarti bool    `json:"yurutucu_gecmis_proje_sarti"`
	IzinSeyahatBeyaniZorunlu bool    `json:"izin_seyahat_beyani_zorunlu"`
	FirmaOrtaklikBeyaniZorunlu bool  `json:"firma_ortaklik_beyani_zorunlu"`
	MinKurumHissesiOrani     float64 `json:"min_kurum_hissesi_orani"`

	OlusturmaTarihi time.Time  `json:"olusturma_tarihi"`
	YayinTarihi     *time.Time `json:"yayin_tarihi,omitempty"`
}

// UyeKisitlama yapısı, akademisyenlerin cezai BAP başvuru engellerini temsil eder.
type UyeKisitlama struct {
	KisitlamaID     int        `json:"kisitlama_id"`
	UyeID           int        `json:"uye_id"`
	KisitlamaTuru   string     `json:"kisitlama_turu"` // intihal | yayin_eksikligi
	BaslangicTarihi time.Time  `json:"baslangic_tarihi"`
	BitisTarihi     *time.Time `json:"bitis_tarihi,omitempty"`
	Aciklama        string     `json:"aciklama"`
	AktifMi         bool       `json:"aktif_mi"`
	OlusturmaTarihi time.Time  `json:"olusturma_tarihi"`
	// JOIN ile doldurulacak alanlar
	AdTumu          string     `json:"ad_tumu,omitempty"`
	Unvan           string     `json:"unvan,omitempty"`
	Eposta          string     `json:"eposta,omitempty"`
}

// HakemHakedis yapısı, hakemlerin değerlendirme ücreti hakediş ve süre takibini temsil eder.
type HakemHakedis struct {
	HakedisID        int        `json:"hakedis_id"`
	ProjeID          int        `json:"proje_id"`
	HakemUyeID       int        `json:"hakem_uye_id"`
	SonTeslimTarihi  time.Time  `json:"son_teslim_tarihi"`
	TamamlanmaTarihi *time.Time `json:"tamamlanma_tarihi,omitempty"`
	TutarTL          float64    `json:"tutar_tl"`
	UcretOraniYuzde  float64    `json:"ucret_orani_yuzde"`
	OdemeDurumu      string     `json:"odeme_durumu"` // bekliyor | onaylandi | odendi
	OlusturmaTarihi  time.Time  `json:"olusturma_tarihi"`
	// JOIN ile doldurulacak alanlar
	HakemAdTumu      string     `json:"hakem_ad_tumu,omitempty"`
	HakemUnvan       string     `json:"hakem_unvan,omitempty"`
	ProjeKodu        string     `json:"proje_kodu,omitempty"`
	ProjeBaslik      string     `json:"proje_baslik,omitempty"`
	BapTuru          string     `json:"bap_turu,omitempty"`
	GecikmeGun       int        `json:"gecikme_gun,omitempty"`
}

// ProjeCiktiTuru yapısı, proje çıktı türlerini tutar.
type ProjeCiktiTuru struct {
	CiktiTuruID int    `json:"cikti_turu_id"`
	CiktiTuru   string `json:"cikti_turu"`
}

// ProjeEtik yapısı, etik durum tanımlarını tutar.
type ProjeEtik struct {
	EtikID     int    `json:"etik_id"`
	EtikDurumu string `json:"etik_durumu"`
}

// ButceTanim yapısı, bütçe türü tanımlarını tutar.
type ButceTanim struct {
	TanimTipID int    `json:"tanim_tip_id"`
	TanimAdi   string `json:"tanim_adi"`
}

// ButceKategori yapısı, bütçe kategorilerini tutar.
type ButceKategori struct {
	KategoriID  int    `json:"kategori_id"`
	KategoriAdi string `json:"kategori_adi"`
}

// OlanakTur yapısı, olanak türlerini tutar.
type OlanakTur struct {
	OlanakTurID int    `json:"olanak_tur_id"`
	TurAdi      string `json:"tur_adi"`
}
