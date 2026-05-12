package models

// ProjeDavet, kullanıcıya gelen proje davet bilgisini temsil eder
type ProjeDavet struct {
	ProjeID      int    `json:"proje_id"`
	BaslikTr     string `json:"baslik_tr"`
	BapTuru      string `json:"bap_turu"`
	DavetEdenAd  string `json:"davet_eden_ad"`
	ProjeRol     string `json:"proje_rol"`
	DavetDurumu  string `json:"davet_durumu"`
	DavetTarihi  string `json:"davet_tarihi"`
}

// DavetYanit, kullanıcının davet yanıtını temsil eder
type DavetYanit struct {
	ProjeID int  `json:"proje_id" binding:"required"`
	Kabul   bool `json:"kabul"`
}
