package models

// ProjeAsama yapısı, iş akışındaki onay masalarını (aşamalarını) temsil eder.
// Örnek: "Dekan Onayına Sun", "Komisyona Sun", "Hakeme Sun"
type ProjeAsama struct {
	AsamaID          int    `json:"asama_id"`
	AsamaKodu        string `json:"asama_kodu"` // Dahili kod: dekan_onayina_sun
	AsamaAdi         string `json:"asama_adi"`  // Görünen ad: Dekan Onayına Sun
	SiraNo           int    `json:"sira_no"`                      // İş akışındaki sıra
	DurumAdi         string `json:"durum_adi"`                     // İlişkili bekleme durumu (Örn: dekan_onayi_bekliyor)
	OnayDurumAdi     string `json:"onay_durum_adi"`                // İlişkili onaylandı durumu (Örn: dekan_onayladi)
	AktifProjeSayisi int    `json:"aktif_proje_sayisi,omitempty"` // Aşamadaki aktif proje sayısı
}
