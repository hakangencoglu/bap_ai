package models

import "time"

// RAGDokuman yapısı, veritabanından RAG indeksi için çıkarılan içerik parçalarını temsil eder.
type RAGDokuman struct {
	DokumanID       int       `json:"dokuman_id"`
	VarlikTuru      string    `json:"varlik_turu"` // 'proje', 'is_paketi', 'risk', 'butce', 'satinalma', 'uye', 'mevzuat'
	VarlikID        int       `json:"varlik_id"`
	ProjeID         *int      `json:"proje_id,omitempty"`
	UyeID           *int      `json:"uye_id,omitempty"`
	Baslik          string    `json:"baslik"`
	Icerik          string    `json:"icerik"`
	MetadataJSON    string    `json:"metadata_json,omitempty"`
	VektorVeri      string    `json:"vektor_veri,omitempty"`
	OlusturmaTarihi time.Time `json:"olusturma_tarihi"`
}

// RAGSearchResult yapısı, RAG arama sorgusu sonucunda dönen doküman parçasını ifade eder.
type RAGSearchResult struct {
	DokumanID  int     `json:"dokuman_id"`
	VarlikTuru string  `json:"varlik_turu"`
	VarlikID   int     `json:"varlik_id"`
	Baslik     string  `json:"baslik"`
	Icerik     string  `json:"icerik"`
	Skor       float64 `json:"skor"`
}

// RAGSearchQuery yapısı, Chatbot tarafından gelen RAG arama parametrelerini tutar.
type RAGSearchQuery struct {
	SorguMetni  string   `json:"sorgu_metni"`
	KullaniciID int      `json:"kullanici_id"`
	Roller      []string `json:"roller"`
	Limit       int      `json:"limit"`
}

// RAGQueryType yapısı, sorgunun türünü belirler (Sayısal vs Semantik).
type RAGQueryType string

const (
	QueryTypeStructured RAGQueryType = "structured" // Sayısal, bütçe, durum sorguları
	QueryTypeSemantic   RAGQueryType = "semantic"   // Metin, özet, risk, amaç sorguları
	QueryTypeHybrid     RAGQueryType = "hybrid"     // Hem sayısal hem metinsel
)

// ChatGecmisiItem yapısı veritabanındaki kullanıcı sohbet geçmişi kaydını temsil eder.
type ChatGecmisiItem struct {
	MesajID         int       `json:"mesaj_id"`
	UyeID           int       `json:"uye_id"`
	Rol             string    `json:"rol"` // 'user' veya 'assistant'
	Icerik          string    `json:"icerik"`
	OlusturmaTarihi time.Time `json:"olusturma_tarihi"`
}
