package models

import (
	"encoding/json"
	"time"
)

// ProjeDegisiklik, proje üzerinde yapılan anlamlı bir işlemin audit başlığıdır.
// Türkçe Yorum: Talep oluşturma/onay, durum değişimi vb. olayları önce/sonra ile izler.
type ProjeDegisiklik struct {
	DegisiklikID   int                    `json:"degisiklik_id"`
	ProjeID        int                    `json:"proje_id"`
	IcerikVersiyon int                    `json:"icerik_versiyon"`
	OlayTipi       string                 `json:"olay_tipi"`
	KaynakTip      string                 `json:"kaynak_tip,omitempty"`
	KaynakID       int                    `json:"kaynak_id,omitempty"`
	TalepNo        string                 `json:"talep_no,omitempty"`
	Ozet           string                 `json:"ozet"`
	IslemiYapanID  int                    `json:"islemi_yapan_id,omitempty"`
	IslemiYapanAd  string                 `json:"islemi_yapan_ad,omitempty"`
	IslemTarihi    time.Time              `json:"islem_tarihi"`
	Detaylar       []ProjeDegisiklikDetay `json:"detaylar,omitempty"`
}

// ProjeDegisiklikDetay, bir olayın etkilenen varlık için önce/sonra snapshot'ıdır.
type ProjeDegisiklikDetay struct {
	DetayID      int             `json:"detay_id"`
	DegisiklikID int             `json:"degisiklik_id"`
	VarlikTip    string          `json:"varlik_tip"`
	VarlikID     int             `json:"varlik_id"`
	OncekiJSON   json.RawMessage `json:"onceki_json,omitempty"`
	SonrakiJSON  json.RawMessage `json:"sonraki_json,omitempty"`
	AlanDiff     json.RawMessage `json:"alan_diff,omitempty"`
}

// DegisiklikKayitIstek, audit kaydı oluşturmak için iç servis modelidir.
type DegisiklikKayitIstek struct {
	ProjeID       int
	OlayTipi      string
	KaynakTip     string
	KaynakID      int
	TalepNo       string
	Ozet          string
	IslemiYapanID int
	Detaylar      []DegisiklikDetayIstek
	VersiyonArtir bool // true ise proje.icerik_versiyon artırılır
}

// DegisiklikDetayIstek, kayıt anında gönderilen varlık snapshot'ıdır.
type DegisiklikDetayIstek struct {
	VarlikTip string
	VarlikID  int
	Onceki    interface{}
	Sonraki   interface{}
}
