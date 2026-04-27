package models

// RiskYonetimi yapısı, proje risk yönetimi bilgilerini tutar.
type RiskYonetimi struct {
	RiskID         int    `json:"risk_id"`
	ProjeID        int    `json:"proje_id"`
	PaketID        *int   `json:"paket_id"`          // Nullable FK → is_paketi
	RiskAciklamasi string `json:"risk_aciklamasi"`
	CozumPlani     string `json:"cozum_plani"`
}
