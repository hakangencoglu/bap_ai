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

	// Versiyon meta (admin / proje bağlama)
	VersiyonID           *int       `json:"versiyon_id,omitempty"`
	VersiyonNo           *int       `json:"versiyon_no,omitempty"`
	VersiyonDurum        string     `json:"versiyon_durum,omitempty"` // taslak | yayinda | arsiv
	YayinVersiyonNo      *int       `json:"yayin_versiyon_no,omitempty"`
	YayinVersiyonID      *int       `json:"yayin_versiyon_id,omitempty"`
	TaslakVarMi          bool       `json:"taslak_var_mi"`
	YayinlanmamisDegisiklik bool    `json:"yayinlanmamis_degisiklik"`
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
	OlusturmaTarihi time.Time  `json:"olusturma_tarihi"`
	YayinTarihi     *time.Time `json:"yayin_tarihi,omitempty"`
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
