package models

// Arastirma yapısı, projeye bağlı araştırma bilgilerini tutar.
type Arastirma struct {
	ProjeID         int    `json:"proje_id"`
	OlusturanID     *int   `json:"olusturan_id"`      // Nullable FK → uye
	ArastirmaAmaci  string `json:"arastirma_amaci"`
}
