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

func main() {
	baseURL := "http://localhost:8080"
	dbConnStr := "host=localhost port=5432 user=bap password=bap_admin_1 dbname=bap_app sslmode=disable"

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

	var loginResp LoginResponse
	json.NewDecoder(resp.Body).Decode(&loginResp)
	token := loginResp.Token
	fmt.Println("[1] Login successful.")

	// 2. Create Project
	createPayload := map[string]interface{}{
		"baslik_tr":    "Coordinator Direct Registration Test",
		"bap_turu_id":  1,
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
	var createResp ProjectResponse
	json.Unmarshal(respBody, &createResp)
	projeID := createResp.ProjeID
	fmt.Printf("[2] Created project with ID: %d\n", projeID)

	defer func() {
		// Cleanup test project
		_, _ = db.Exec("DELETE FROM proje WHERE proje_id = $1", projeID)
		fmt.Printf("[Cleanup] Deleted test project %d\n", projeID)
	}()

	// 3. Add regular researcher (role_id = 2) via API -> Should be 'beklemede'
	// Find another academician / researcher ID to test with
	var testUyeID int
	err = db.QueryRow("SELECT uye_id FROM uye WHERE eposta = 'ogrenci@izu.edu.tr'").Scan(&testUyeID)
	if err != nil {
		log.Fatalf("Failed to fetch test student ID: %v", err)
	}

	invitePayload := map[string]interface{}{
		"uye_id": testUyeID,
		"rol_id": 2, // Araştırmacı
	}
	inviteBytes, _ := json.Marshal(invitePayload)
	inviteReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/proje/%d/takim", baseURL, projeID), bytes.NewBuffer(inviteBytes))
	inviteReq.Header.Set("Content-Type", "application/json")
	inviteReq.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(inviteReq)
	if err != nil {
		log.Fatalf("POST member invite failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Fatalf("Add researcher member failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Verify in DB that student is 'beklemede'
	var testDavetDurumu string
	err = db.QueryRow("SELECT davet_durumu FROM proje_takim WHERE proje_id = $1 AND uye_id = $2", projeID, testUyeID).Scan(&testDavetDurumu)
	if err != nil {
		log.Fatalf("Failed to query student member in DB: %v", err)
	}
	fmt.Printf("[3] Student member added. DB davet_durumu: %s (Expected: beklemede)\n", testDavetDurumu)
	if testDavetDurumu != "beklemede" {
		log.Fatalf("Student should be 'beklemede' but is %s", testDavetDurumu)
	}

	// 4. Add coordinator (role_id = 1) via API -> Should be 'kabul' directly
	// Find another academician ID to assign as coordinator
	var testHocaID int
	err = db.QueryRow("SELECT uye_id FROM uye WHERE eposta = 'test.ldap@izu.edu.tr'").Scan(&testHocaID)
	if err != nil {
		log.Fatalf("Failed to fetch test academician ID: %v", err)
	}

	yurutucuPayload := map[string]interface{}{
		"uye_id": testHocaID,
		"rol_id": 1, // Yürütücü
	}
	yurutucuBytes, _ := json.Marshal(yurutucuPayload)
	yurutucuReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/proje/%d/takim", baseURL, projeID), bytes.NewBuffer(yurutucuBytes))
	yurutucuReq.Header.Set("Content-Type", "application/json")
	yurutucuReq.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(yurutucuReq)
	if err != nil {
		log.Fatalf("POST yurutucu invite failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Fatalf("Add yurutucu failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Verify in DB that yurutucu is 'kabul' directly and koordinator_id is updated
	var yurutucuDavetDurumu string
	err = db.QueryRow("SELECT davet_durumu FROM proje_takim WHERE proje_id = $1 AND uye_id = $2", projeID, testHocaID).Scan(&yurutucuDavetDurumu)
	if err != nil {
		log.Fatalf("Failed to query yurutucu member in DB: %v", err)
	}
	fmt.Printf("[4] Coordinator member added. DB davet_durumu: %s (Expected: kabul)\n", yurutucuDavetDurumu)
	if yurutucuDavetDurumu != "kabul" {
		log.Fatalf("Coordinator should be 'kabul' directly but is %s", yurutucuDavetDurumu)
	}

	var dbKoordinatorID sql.NullInt64
	err = db.QueryRow("SELECT koordinator_id FROM proje WHERE proje_id = $1", projeID).Scan(&dbKoordinatorID)
	if err != nil {
		log.Fatalf("Failed to query koordinator_id in DB: %v", err)
	}
	fmt.Printf("[5] DB proje koordinator_id: %v (Expected: %d)\n", dbKoordinatorID.Int64, testHocaID)
	if !dbKoordinatorID.Valid || dbKoordinatorID.Int64 != int64(testHocaID) {
		log.Fatalf("Project koordinator_id should be %d, got %v", testHocaID, dbKoordinatorID.Int64)
	}

	fmt.Println("All coordinator direct registration verification tests PASSED successfully!")
}
