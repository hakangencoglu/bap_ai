package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// ChatService yapısı, yapay zeka entegrasyonu için gerekli yapılandırmayı tutar.
type ChatService struct {
	LLMProvider  string
	LLMEndpoint  string
	LLMModel     string
	GeminiAPIKey string
}

// NewChatService yeni bir ChatService nesnesi oluşturur
func NewChatService(provider, endpoint, model, apiKey string) *ChatService {
	// Türkçe Yorum: LLM servis yapılandırması ayarlanır ve servis nesnesi döndürülür.
	return &ChatService{
		LLMProvider:  provider,
		LLMEndpoint:  endpoint,
		LLMModel:     model,
		GeminiAPIKey: apiKey,
	}
}

// SendChatMessage yapay zeka modeline veya yerel motoruna mesaj gönderir
func (s *ChatService) SendChatMessage(userRole, userName, message string) (string, error) {
	// Türkçe Yorum: LLM sağlayıcısına göre istek yönlendirilir.
	log.Printf("ChatBot: Kullanıcı %s (%s) için mesaj alındı: %s\n", userName, userRole, message)

	// Sistem Talimatı (System Prompt) - Yapay zekaya kişiliğini ve BAP kurallarını öğretir
	systemPrompt := fmt.Sprintf(`Sen İstanbul Sabahattin Zaim Üniversitesi (İZÜ) BAP (Bilimsel Araştırma Projeleri) Yapay Zeka Asistanısın. 
Şu an sisteme giriş yapmış olan kullanıcı: %s (Rolü: %s). Ona bu rol doğrultusunda yardımcı ol.

İZÜ BAP Sistemi Kuralları ve Limitleri:
1. BAP-100 (Lisans Tez Projesi): Bütçe limiti 50.000,00 TL, Süre limiti 12 ay.
2. BAP-200 (Yüksek Lisans Tez Projesi): Bütçe limiti 100.000,00 TL, Süre limiti 24 ay.
3. BAP-300 (Doktora Tez Projesi): Bütçe limiti 150.000,00 TL, Süre limiti 36 ay.
4. BAP-400 (Akademisyen Araştırma Projesi): Bütçe limiti 30.000,00 TL, Süre limiti 6 ay.
5. BAP-500 (Bilimsel Etkinlik Destek Projesi): Bütçe limiti 250.000,00 TL, Süre limiti 36 ay.

Süreçler:
- Proje Başvurusu: Sol menüdeki 'Yeni Başvuru' sekmesinden 5 adımlı form doldurularak yapılır. Form doldurulurken otomatik kaydetme etkindir.
- Hakem Süreci: Admin hakem atar. Atanan hakemler kabul ederse projeyi 0-100 arası puanlar ve yorum bildirir.
- Revizyon Süreci: Hakemler veya Admin revizyon isteyebilir. Bu durumda proje düzenlemeye yeniden açılır.
- Satın Alma Süreci: Proje 'tamamlandi' yani aktif/sözleşme imzalanmış durumdayken akademisyen bütçe kalemlerinden satın alma talebi açar, TTO onaylar veya reddeder.
- E-İmza Süreci: Onaylanan projeler e-imza aşamasına geçer.

Sorulara kısa, net, markdown formatında ve profesyonel bir Türkçe ile yanıt ver. BAP dışı konularda nazikçe sadece BAP AI hakkında bilgi verebileceğini söyle.`, userName, userRole)

	// Sağlayıcıya göre işlem yap
	switch strings.ToLower(s.LLMProvider) {
	case "ollama":
		response, err := s.callOllamaAPI(systemPrompt, message)
		if err == nil {
			return response + "\n\n*(Yapay Zeka - Yerel Ollama)*", nil
		}
		log.Printf("ChatBot: Ollama API çağrısı başarısız oldu, yerel fallback devrede: %v\n", err)
	case "gemini":
		if s.GeminiAPIKey != "" {
			response, err := s.callGeminiAPI(systemPrompt, message)
			if err == nil {
				return response + "\n\n*(Yapay Zeka - Google Gemini)*", nil
			}
			log.Printf("ChatBot: Gemini API çağrısı başarısız oldu, yerel fallback devrede: %v\n", err)
		}
	case "openai_compatible":
		response, err := s.callOpenAICompatibleAPI(systemPrompt, message)
		if err == nil {
			return response + "\n\n*(Yapay Zeka - Bulut API)*", nil
		}
		log.Printf("ChatBot: OpenAI Uyumlu API çağrısı başarısız oldu, yerel fallback devrede: %v\n", err)
	}

	// Eğer LLM servisleri kapalıysa veya hata alındıysa, lokal akıllı kurallar devreye girer
	return s.getLocalFallbackResponse(userRole, message) + "\n\n*(Çevrimdışı / Yerel Asistan Modu)*", nil
}

// callOllamaAPI lokal Ollama sunucusundan yanıt üretir
func (s *ChatService) callOllamaAPI(systemPrompt, userMessage string) (string, error) {
	// Türkçe Yorum: Ollama /api/chat endpoint'i için istek gövdesi hazırlanır.
	url := fmt.Sprintf("%s/api/chat", s.LLMEndpoint)

	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type Request struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
		Stream   bool      `json:"stream"`
	}

	reqBody := Request{
		Model: s.LLMModel,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama sunucusu hata kodu döndü: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	type Response struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	var res Response
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return "", err
	}

	return res.Message.Content, nil
}

// callGeminiAPI Google Gemini API'sinden yanıt üretir
func (s *ChatService) callGeminiAPI(systemPrompt, userMessage string) (string, error) {
	// Türkçe Yorum: Gemini API generateContent endpoint'ine istek atılır.
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", s.LLMModel, s.GeminiAPIKey)

	// Eski/yeni model uyumluluğu için model ismi boşsa flash atanır
	if s.LLMModel == "" || s.LLMModel == "llama3" {
		url = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s", s.GeminiAPIKey)
	}

	type Part struct {
		Text string `json:"text"`
	}
	type Content struct {
		Parts []Part `json:"parts"`
	}
	type Request struct {
		Contents []Content `json:"contents"`
	}

	combinedPrompt := fmt.Sprintf("%s\n\nKullanıcı Sorusu:\n%s", systemPrompt, userMessage)

	reqBody := Request{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: combinedPrompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini api hata kodu döndü: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	type GeminiResponse struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	var res GeminiResponse
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return "", err
	}

	if len(res.Candidates) > 0 && len(res.Candidates[0].Content.Parts) > 0 {
		return res.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("gemini api geçerli bir yanıt dönmedi")
}

// callOpenAICompatibleAPI OpenAI uyumlu bir yerel API'den yanıt üretir
func (s *ChatService) callOpenAICompatibleAPI(systemPrompt, userMessage string) (string, error) {
	// Türkçe Yorum: OpenAI uyumlu /v1/chat/completions endpoint'ine istek atılır.
	url := fmt.Sprintf("%s/v1/chat/completions", s.LLMEndpoint)

	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type Request struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
	}

	reqBody := Request{
		Model: s.LLMModel,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai uyumlu sunucu hata kodu döndü: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	type Response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	var res Response
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return "", err
	}

	if len(res.Choices) > 0 {
		return res.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("uyumlu api geçerli bir yanıt dönmedi")
}

// getLocalFallbackResponse yerel kural tabanlı motor ile akıllı yanıt üretir (LLM'e ulaşılamazsa)
func (s *ChatService) getLocalFallbackResponse(userRole, message string) string {
	// Türkçe Yorum: LLM aktif değilse, anahtar kelime eşleştirmesi ile Türkçe BAP kuralları ve yönlendirmeleri döner.
	msg := strings.ToLower(message)

	if strings.Contains(msg, "merhaba") || strings.Contains(msg, "selam") {
		return fmt.Sprintf("Merhaba! Ben **İZÜ BAP Yapay Zeka Asistanı** 🤖. \n\nSistemdeki rolünüz **%s** olarak görünüyor. Size BAP proje limitleri, başvuru süreci, hakem değerlendirmeleri veya satın alma talepleri gibi konularda rehberlik edebilirim. \n\nNasıl yardımcı olabilirim?", userRole)
	}

	if strings.Contains(msg, "bap-100") {
		return "### BAP-100 (Lisans Tez Projesi)\n- **Bütçe Limiti:** 50.000,00 TL\n- **Süre Sınırı:** En fazla 12 ay\n- **Açıklama:** İZÜ lisans öğrencilerinin araştırma kültürünü geliştirmek amacıyla tez çalışmalarına yönelik verdikleri destek projeleridir."
	}
	if strings.Contains(msg, "bap-200") {
		return "### BAP-200 (Yüksek Lisans Tez Projesi)\n- **Bütçe Limiti:** 100.000,00 TL\n- **Süre Sınırı:** En fazla 24 ay\n- **Açıklama:** Enstitü bünyesindeki tezli yüksek lisans programı öğrencilerinin tez projelerini desteklemeyi hedefler."
	}
	if strings.Contains(msg, "bap-300") {
		return "### BAP-300 (Doktora Tez Projesi)\n- **Bütçe Limiti:** 150.000,00 TL\n- **Süre Sınırı:** En fazla 36 ay\n- **Açıklama:** Doktora öğrencilerinin tez çalışmalarının desteklenmesine yöneliktir."
	}
	if strings.Contains(msg, "bap-400") {
		return "### BAP-400 (Akademisyen Araştırma Projesi)\n- **Bütçe Limiti:** 30.000,00 TL\n- **Süre Sınırı:** En fazla 6 ay\n- **Açıklama:** Üniversitemiz akademisyenlerinin bireysel araştırma veya ön fizibilite projelerine verilen destektir."
	}
	if strings.Contains(msg, "bap-500") {
		return "### BAP-500 (Bilimsel Etkinlik Destek Projesi)\n- **Bütçe Limiti:** 250.000,00 TL\n- **Süre Sınırı:** En fazla 36 ay\n- **Açıklama:** Büyük ölçekli kongre, konferans, sempozyum veya uluslararası bilimsel etkinliklerin düzenlenmesine yönelik kurumsal destektir."
	}

	if strings.Contains(msg, "bütçe") || strings.Contains(msg, "limit") || strings.Contains(msg, "tutar") || strings.Contains(msg, "para") {
		return "### İZÜ BAP Proje Bütçe Limitleri:\n\n| Proje Türü | Limit (TL) | Maks. Süre |\n| :--- | :--- | :--- |\n| **BAP-100** | 50.000,00 TL | 12 Ay |\n| **BAP-200** | 100.000,00 TL | 24 Ay |\n| **BAP-300** | 150.000,00 TL | 36 Ay |\n| **BAP-400** | 30.000,00 TL | 6 Ay |\n| **BAP-500** | 250.000,00 TL | 36 Ay |\n\n*Not: Başvurularda bütçe kalemleri Makine-Teçhizat, Sarf Malzeme, Hizmet Alımı, Yazılım ve Seyahat olarak detaylandırılmalıdır.*"
	}

	if strings.Contains(msg, "satın alma") || strings.Contains(msg, "satınalma") || strings.Contains(msg, "sipariş") || strings.Contains(msg, "harcama") {
		if userRole == "akademisyen" || userRole == "admin" {
			return "### Satın Alma Talebi Nasıl Oluşturulur? (Akademisyen)\n1. Yürütücüsü olduğunuz projenin onaylanıp **Aktif (TTO Aktif)** duruma gelmesi gerekir.\n2. **Anasayfa** panelinde projenizin yanındaki yeşil **'Satın Alma'** butonuna tıklayın.\n3. Açılan pencerede bütçe kalemini seçin, malzeme adını, miktarını ve birim fiyatını girerek gerekçenizi yazın.\n4. **'Talebi Gönder'** butonuyla TTO onayına sunun.\n\n*Talebiniz TTO tarafından onaylandığında bütçenizden otomatik olarak düşülecektir.*"
		}
		if userRole == "tto" {
			return "### Satın Alma Onay Süreci (TTO Yetkilisi)\n1. Sol menüdeki **'Satın Alma Talepleri'** sekmesine geçiş yapın.\n2. Akademisyenlerden gelen tüm satın alma istekleri burada **'Beklemede'** durumunda listelenir.\n3. **'Onayla'** veya red gerekçesi yazarak **'Reddet'** butonları ile talepleri karara bağlayabilirsiniz."
		}
		return "Satın alma işlemleri sadece yürütücü akademisyenler ve TTO yetkilileri tarafından yönetilebilir. Aktif projelerin bütçeleri kapsamında malzeme/hizmet alımı talepleri açılabilmektedir."
	}

	if strings.Contains(msg, "hakem") || strings.Contains(msg, "değerlendirme") || strings.Contains(msg, "puan") {
		return "### Hakem Değerlendirme Süreci\n- Başvuru tamamlandığında Admin paneli üzerinden ilgili alandan bağımsız **hakem ataması** yapılır.\n- Hakemler kendilerine atanan projeyi inceledikten sonra **Kabul** veya gerekçe bildirerek **Red** kararı verirler.\n- Değerlendirmeyi kabul eden hakemler; projenin özgün değerini, hedeflerini ve metodolojisini **0-100 puan** arası notlandırıp, detaylı rapor yazarlar."
	}

	if strings.Contains(msg, "revizyon") || strings.Contains(msg, "düzeltme") || strings.Contains(msg, "düzenle") {
		return "### Revizyon (Düzeltme) İşlemi\n- Hakem veya yöneticilerin başvuruda eksik gördüğü yerler için **Revizyon** talebi oluşturulabilir.\n- Revizyon kararı verildiğinde akademisyene bildirim gider ve projesi yeniden **düzenlenebilir taslak** durumuna geçer.\n- Akademisyen, revizyon notlarındaki düzeltmeleri yaparak başvuruyu günceller ve tekrar onay sürecine sunar."
	}

	if strings.Contains(msg, "imza") || strings.Contains(msg, "sözleşme") || strings.Contains(msg, "e-imza") {
		return "### E-İmza Paneli\n- Proje başvurusu tüm onay basamaklarından (Dekan ve Komisyon) geçtikten sonra ıslak imza yerine **E-İmza Paneli** üzerinden e-imzalanır.\n- Sol menüdeki **E-İmza Paneli** alanında bekleyen ve imzalanan belgelerinizi görüntüleyebilirsiniz. İmzalama işlemi tarayıcı üzerinden güvenli token ve IP/tarayıcı bilgisi doğrulamasıyla gerçekleşir."
	}

	if strings.Contains(msg, "nasıl başvuru") || strings.Contains(msg, "başvuru yap") || strings.Contains(msg, "başvuru nasıl") {
		return "### BAP Başvurusu Nasıl Yapılır?\n1. Sol menüdeki **'Yeni Başvuru'** butonuna tıklayın.\n2. **5 Adımda formu doldurun:**\n   - **1. Adım:** Başlık ve BAP Türü seçimi.\n   - **2. Adım:** Proje Ekibi (Araştırmacı, Danışman, Bursiyer davet etme).\n   - **3. Adım:** Bütçe Kalemleri ve harcama detayları.\n   - **4. Adım:** Proje Özeti ve Anahtar kelimeler (Türkçe & İngilizce).\n   - **5. Adım:** Risk planları ve Kaynakça.\n3. Form doldurulurken verileriniz **otomatik olarak arka planda kaydedilir**.\n4. Son adımda **'Başvuruyu Tamamla'** diyerek onay akışını başlatabilirsiniz."
	}

	// Genel yanıt
	return fmt.Sprintf("BAP asistanı olarak sorunuzu tam anlayamadım, ancak İZÜ BAP sistemiyle ilgili şu konularda destek sağlayabilirim:\n\n- **BAP-100/200/300/400/500** bütçe ve süre limitleri\n- **Yeni Başvuru** oluşturma ve otomatik kaydetme adımları\n- **Satın Alma** talepleri oluşturma ve onay süreçleri\n- **Hakem Değerlendirmeleri** ve puanlama sistemi\n- **Revizyon (Düzeltme)** işlemleri\n- **E-İmza** süreçleri\n\nLütfen detaylı bilgi almak istediğiniz konuyu sorunuz (Örn: *'bap-300 bütçesi nedir?'* veya *'satın alma nasıl yapılır?'*).")
}
