package models

// SistemRolTanimlama yapısı, sistem rollerinin tanımlarını tutar.
type SistemRolTanimlama struct {
	RolID  int    `json:"rol_id"`
	RolAdi string `json:"rol_adi"`
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
	DurumID  int    `json:"durum_id"`
	DurumAdi string `json:"durum_adi"`
}

// ProjeBapTuru yapısı, BAP proje türlerini tutar.
type ProjeBapTuru struct {
	BapTuruID       int     `json:"bap_turu_id"`
	BapTuru         string  `json:"bap_turu"`
	ButceLimiti     float64 `json:"butce_limiti"`
	SureLimitiAy    int     `json:"sure_limiti_ay"`
	AktifMi         bool    `json:"aktif_mi"`
	Aciklama        string  `json:"aciklama"`
	HakemGerekli    bool    `json:"hakem_gerekli"`    // Hakem değerlendirmesi gerekli mi?
	HakemSayisi     int     `json:"hakem_sayisi"`     // Gerekli hakem sayısı
	BursiyerGerekli bool    `json:"bursiyer_gerekli"` // Bursiyer desteği gerekli mi?
	BursiyerSayisi  int     `json:"bursiyer_sayisi"`  // Gerekli bursiyer sayısı
}

// ProjeCiktiTuru yapısı, proje çıktı türlerini tutar.
type ProjeCiktiTuru struct {
	CiktiTuruID int    `json:"cikti_turu_id"`
	CiktiTuru   string `json:"cikti_turu"`
}

// ProjeEtik yapısı, etik durum tanımlarını tutar.
type ProjeEtik struct {
	EtikID      int    `json:"etik_id"`
	EtikDurumu  string `json:"etik_durumu"`
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
