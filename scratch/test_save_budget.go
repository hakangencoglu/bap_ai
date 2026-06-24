package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type LoginResponse struct {
	Token string `json:"token"`
}

type ProjectResponse struct {
	ProjeID int `json:"proje_id"`
}

type BudgetDetailsResponse struct {
	Butceler []struct {
		KalemID       int     `json:"kalem_id"`
		ProjeID       int     `json:"proje_id"`
		Aciklama      string  `json:"aciklama"`
		BirimOzelligi int     `json:"birim_ozelligi"`
		BirimFiyat    float64 `json:"birim_fiyat"`
		ToplamFiyat   float64 `json:"toplam_fiyat"`
		KategoriAdi   string  `json:"kategori_adi"`
	} `json:"butceler"`
}

func main() {
	baseURL := "http://localhost:8080"
	dbConnStr := "host=localhost port=5432 user=bap password=bap_admin_1 dbname=bap_app sslmode=disable"

	// Connect to DB directly for raw checks and cleanup
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("DB Open Error: %v", err)
	}
	defer db.Close()

	// 1. Login
	loginPayload := map[string]string{
		"eposta": "akademisyen@izu.edu.tr",
		"sifre":  "akademisyen123",
	}
	bodyBytes, _ := json.Marshal(loginPayload)
	resp, err := http.Post(baseURL+"/api/auth/login", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		log.Fatalf("Login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Fatalf("Login failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var loginResp LoginResponse
	json.NewDecoder(resp.Body).Decode(&loginResp)
	token := loginResp.Token
	fmt.Println("[1] Login successful.")

	// 2. Create Project
	createPayload := map[string]interface{}{
		"baslik_tr":    "Budget Test Project",
		"bap_turu_id":  1, // BAP-100
		"sure_ay":      12,
		"toplam_butce": 0.0,
	}
	createBytes, _ := json.Marshal(createPayload)
	req, _ := http.NewRequest("POST", baseURL+"/api/proje", bytes.NewBuffer(createBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("POST project request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		log.Fatalf("Create project failed: %s", string(respBody))
	}

	var createResp ProjectResponse
	json.Unmarshal(respBody, &createResp)
	projeID := createResp.ProjeID
	fmt.Printf("[2] Created project with ID: %d\n", projeID)

	defer func() {
		// Cleanup test project
		_, _ = db.Exec("DELETE FROM proje WHERE proje_id = $1", projeID)
		fmt.Printf("[Cleanup] Deleted test project %d\n", projeID)
	}()

	// 3. Save Project Extras (with budget details)
	extrasPayload := map[string]interface{}{
		"butce_kalemleri": []map[string]interface{}{
			{
				"kategori_adi": "Makine-Teçhizat",
				"aciklama":     "Bilgisayar - Çok güçlü gerekçe",
				"miktar":       3,
				"birim_fiyat":  15000.00,
			},
		},
	}
	extrasBytes, _ := json.Marshal(extrasPayload)
	extrasReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/proje/%d/extras", baseURL, projeID), bytes.NewBuffer(extrasBytes))
	extrasReq.Header.Set("Content-Type", "application/json")
	extrasReq.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(extrasReq)
	if err != nil {
		log.Fatalf("POST extras failed: %v", err)
	}
	defer resp.Body.Close()

	extrasRespBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Save extras failed with status %d: %s", resp.StatusCode, string(extrasRespBody))
	}
	fmt.Println("[3] Saved budget items successfully via API.")

	// 4. Fetch Details via API
	getDetailsReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/proje/%d/detaylar", baseURL, projeID), nil)
	getDetailsReq.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(getDetailsReq)
	if err != nil {
		log.Fatalf("GET details failed: %v", err)
	}
	defer resp.Body.Close()

	detailsBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Fetch details failed with status %d: %s", resp.StatusCode, string(detailsBody))
	}

	var detailsResp BudgetDetailsResponse
	json.Unmarshal(detailsBody, &detailsResp)

	if len(detailsResp.Butceler) != 1 {
		log.Fatalf("Expected 1 budget item, got %d", len(detailsResp.Butceler))
	}

	item := detailsResp.Butceler[0]
	fmt.Printf("[4] API Verification:\n")
	fmt.Printf("    - Kategori: %s\n", item.KategoriAdi)
	fmt.Printf("    - Açıklama: %s\n", item.Aciklama)
	fmt.Printf("    - Birim Özelliği (Miktar): %d (Expected: 3)\n", item.BirimOzelligi)
	fmt.Printf("    - Birim Fiyat: %.2f (Expected: 15000.00)\n", item.BirimFiyat)
	fmt.Printf("    - Toplam Fiyat: %.2f (Expected: 45000.00)\n", item.ToplamFiyat)

	if item.BirimOzelligi != 3 {
		log.Fatalf("API verification failed: BirimOzelligi should be 3, got %d", item.BirimOzelligi)
	}

	// 5. Database Direct Verification
	var dbBirimOzelligi int
	var dbAciklama string
	var dbBirimFiyat float64
	var dbToplamFiyat float64
	err = db.QueryRow("SELECT birim_ozelligi, aciklama, birim_fiyat, toplam_fiyat FROM butce WHERE proje_id = $1", projeID).
		Scan(&dbBirimOzelligi, &dbAciklama, &dbBirimFiyat, &dbToplamFiyat)
	if err != nil {
		log.Fatalf("Direct DB Query failed: %v", err)
	}

	fmt.Printf("[5] DB Verification:\n")
	fmt.Printf("    - DB birim_ozelligi: %d (Expected: 3)\n", dbBirimOzelligi)
	fmt.Printf("    - DB aciklama: %s (Expected: 'Bilgisayar - Çok güçlü gerekçe')\n", dbAciklama)
	fmt.Printf("    - DB birim_fiyat: %.2f (Expected: 15000.00)\n", dbBirimFiyat)
	fmt.Printf("    - DB toplam_fiyat: %.2f (Expected: 45000.00)\n", dbToplamFiyat)

	if dbBirimOzelligi != 3 || dbAciklama != "Bilgisayar - Çok güçlü gerekçe" || dbBirimFiyat != 15000.00 || dbToplamFiyat != 45000.00 {
		log.Fatalf("Database verification failed!")
	}

	fmt.Println("All budget saving and loading verification tests PASSED successfully!")
}
