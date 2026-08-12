package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

// EmbeddingService vektör üretimi ve kelime vektörleştirme servisidir.
type EmbeddingService struct {
	Endpoint   string
	Provider   string // 'ollama', 'openai', 'gemini', 'fallback'
	HTTPClient *http.Client
}

// NewEmbeddingService yeni bir EmbeddingService nesnesi oluşturur
func NewEmbeddingService(provider, endpoint string) *EmbeddingService {
	return &EmbeddingService{
		Endpoint: endpoint,
		Provider: strings.ToLower(provider),
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GenerateEmbedding metin içeriğini vektör dizisine (float64 slice) dönüştürür
func (e *EmbeddingService) GenerateEmbedding(text string) ([]float64, error) {
	if strings.TrimSpace(text) == "" {
		return make([]float64, 128), nil
	}

	if e.Provider == "ollama" && e.Endpoint != "" {
		vec, err := e.callOllamaEmbedding(text)
		if err == nil && len(vec) > 0 {
			return vec, nil
		}
	}

	// Fallback: Yerel Frekans Tabanlı Vektör Üreteci (TF-IDF / Hash Vectorization)
	return e.generateLocalTFIDFVector(text, 128), nil
}

// callOllamaEmbedding Ollama /api/embeddings servisini çağırır
func (e *EmbeddingService) callOllamaEmbedding(text string) ([]float64, error) {
	url := fmt.Sprintf("%s/api/embeddings", e.Endpoint)
	reqBody := map[string]interface{}{
		"model":  "nomic-embed-text",
		"prompt": text,
	}
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := e.HTTPClient.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama embedding hatası HTTP: %d", resp.StatusCode)
	}

	var res struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res.Embedding, nil
}

// generateLocalTFIDFVector dış servis erişilemediğinde 128-boyutlu deterministik hash vektörü üretir
func (e *EmbeddingService) generateLocalTFIDFVector(text string, dim int) []float64 {
	vec := make([]float64, dim)
	words := strings.Fields(strings.ToLower(text))
	if len(words) == 0 {
		return vec
	}

	for _, word := range words {
		var hash uint32 = 5381
		for i := 0; i < len(word); i++ {
			hash = ((hash << 5) + hash) + uint32(word[i])
		}
		idx := int(hash % uint32(dim))
		vec[idx] += 1.0
	}

	// L2 Normalizasyonu
	var sumSq float64
	for _, val := range vec {
		sumSq += val * val
	}
	if sumSq > 0 {
		norm := math.Sqrt(sumSq)
		for i := 0; i < dim; i++ {
			vec[i] = vec[i] / norm
		}
	}

	return vec
}
