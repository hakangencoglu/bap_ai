package service

import (
	"fmt"
	"log"
	"strings"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// RAGService yapısı, hibrit arama ve bağlam sentezleme servisini yönetir.
type RAGService struct {
	RAGRepo          *repository.RAGRepository
	EmbeddingService *EmbeddingService
}

// NewRAGService yeni bir RAGService nesnesi oluşturur
func NewRAGService(ragRepo *repository.RAGRepository, embService *EmbeddingService) *RAGService {
	return &RAGService{
		RAGRepo:          ragRepo,
		EmbeddingService: embService,
	}
}

// ClassifyQueryIntent kullanıcı mesajının sayısal mı semantik mi olduğunu sınıflandırır
func (s *RAGService) ClassifyQueryIntent(message string) models.RAGQueryType {
	msgLower := strings.ToLower(message)

	structuredKeywords := []string{
		"kaç", "sayısı", "sayı", "toplam", "listesi", "listele", "bütçe", "harcama",
		"kalan", "durum dağılımı", "istatistik", "onaylanan", "bekleyen", "adı", "adı nedir",
		"projenin", "proje", "kodu", "numaralı", "hangi", "kimin",
	}

	for _, kw := range structuredKeywords {
		if strings.Contains(msgLower, kw) {
			return models.QueryTypeHybrid
		}
	}

	return models.QueryTypeSemantic
}

// BuildRAGContext kullanıcı mesajına uygun filtrelenmiş veritabanı bağlamını üretir
func (s *RAGService) BuildRAGContext(kullaniciID int, roller []string, userMessage string) (string, error) {
	var sb strings.Builder
	intent := s.ClassifyQueryIntent(userMessage)

	sb.WriteString("=== BAP VERİTABANI VE MEVZUAT BAĞLAMI (RAG) ===\n\n")

	// 1. Yapılandırılmış Sorgu (Bütçe, Sayı, İstatistikler)
	if intent == models.QueryTypeStructured || intent == models.QueryTypeHybrid {
		structuredData, err := s.RAGRepo.QueryStructuredSummary(kullaniciID, roller, userMessage)
		if err == nil && strings.TrimSpace(structuredData) != "" {
			sb.WriteString("--- YAPILANDIRILMIŞ İSTATİSTİKLER VE ÖZETLER ---\n")
			sb.WriteString(structuredData)
			sb.WriteString("\n")
		}
	}

	// 2. Semantik & Metinsel Doküman Araması
	searchQuery := models.RAGSearchQuery{
		SorguMetni:  userMessage,
		KullaniciID: kullaniciID,
		Roller:      roller,
		Limit:       8,
	}

	docs, err := s.RAGRepo.SearchHybrid(searchQuery)
	if err == nil && len(docs) > 0 {
		sb.WriteString("--- ALAKALI PROJE, RİSK, İŞ PAKETİ VE MEVZUAT DETAYLARI ---\n")
		for i, doc := range docs {
			sb.WriteString(fmt.Sprintf("[%d] %s (%s)\n%s\n\n", i+1, doc.Baslik, doc.VarlikTuru, doc.Icerik))
		}
	} else {
		log.Printf("RAG Bilgi: Semantik arama doğrudan eşleşme bulamadı, genel özetler sunuldu.\n")
	}

	return sb.String(), nil
}

// SyncDatabase veritabanındaki verileri RAG indeksine taşır
func (s *RAGService) SyncDatabase() (int, error) {
	return s.RAGRepo.SyncDatabaseToRAG()
}
