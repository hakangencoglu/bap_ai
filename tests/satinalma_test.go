package tests

import (
	"bap_ai/backend/database"
	"bap_ai/backend/models"
	"bap_ai/backend/repository"
	"bap_ai/backend/service"
	"bap_ai/configs"
	"database/sql"
	"testing"
)

// TestSatinalmaProcesses satın alma yetki ve rezervasyon mantığını test eder.
func TestSatinalmaProcesses(t *testing.T) {
	// 1. Setup DB
	configs.LoadConfig()
	database.Connect()
	if database.DB == nil {
		t.Fatal("Veritabanı bağlantısı kurulamadı")
	}

	err := database.RunSchema(database.DB, "../backend/database/schema.sql")
	if err != nil {
		t.Fatalf("Şema yükleme hatası: %v", err)
	}

	projeRepo := repository.NewProjeRepository(database.DB)
	satinalmaRepo := repository.NewSatinalmaRepository(database.DB)
	satinalmaSrv := service.NewSatinalmaService(satinalmaRepo, projeRepo)

	// Test verileri için temizlik ve oluşturma işlemleri
	// 2. Geçici Test Üyesi Ekle
	var testUyeID int
	err = database.DB.QueryRow(`
		INSERT INTO uye (rol, ad, soyad, unvan, bolum, eposta, telefon, izu_uyesi, sifre_hash, aktif_mi)
		VALUES ('akademisyen', 'Test', 'Hoca', 'Dr.', 'Test Bölümü', 'test.hoca.satinalma@izu.edu.tr', '12345', true, 'hash', true)
		RETURNING uye_id
	`).Scan(&testUyeID)
	if err != nil {
		t.Fatalf("Test üyesi oluşturulamadı: %v", err)
	}
	defer func() {
		_, _ = database.DB.Exec("DELETE FROM uye WHERE uye_id = $1", testUyeID)
	}()

	// 3. Geçici Test Projesi Ekle (Durum: tamamlandi - aktif)
	var durumID int
	err = database.DB.QueryRow("SELECT durum_id FROM proje_durum WHERE durum_adi = 'tamamlandi'").Scan(&durumID)
	if err != nil {
		t.Fatalf("tamamlandi durum ID bulunamadı: %v", err)
	}

	var testProjeID int
	err = database.DB.QueryRow(`
		INSERT INTO proje (baslik_tr, bap_turu_id, sure_ay, toplam_butce, koordinator_id, durum_id)
		VALUES ('Test Satın Alma Projesi', 1, 12, 10000.00, $1, $2)
		RETURNING proje_id
	`, testUyeID, durumID).Scan(&testProjeID)
	if err != nil {
		t.Fatalf("Test projesi oluşturulamadı: %v", err)
	}
	defer func() {
		_, _ = database.DB.Exec("DELETE FROM proje WHERE proje_id = $1", testProjeID)
	}()

	// Bütçe kalemi ekle
	var testKalemID int
	err = database.DB.QueryRow(`
		INSERT INTO butce (proje_id, aciklama, birim_ozelligi, birim_fiyat, toplam_fiyat)
		VALUES ($1, 'Test Cihazı', 1, 5000.00, 5000.00)
		RETURNING kalem_id
	`, testProjeID).Scan(&testKalemID)
	if err != nil {
		t.Fatalf("Bütçe kalemi oluşturulamadı: %v", err)
	}

	// 4. Yetki Testi (Başlangıçta üye olmamalı)
	isMember, err := projeRepo.IsProjeUyesi(testProjeID, testUyeID)
	if err != nil {
		t.Fatalf("IsProjeUyesi hatası: %v", err)
	}
	if isMember {
		t.Error("Kullanıcı projeye henüz eklenmemişken IsProjeUyesi true döndü")
	}

	// Proje takımına ekle (davet_durumu = 'kabul')
	_, err = database.DB.Exec(`
		INSERT INTO proje_takim (proje_id, uye_id, proje_rol_id, davet_durumu)
		VALUES ($1, $2, 1, 'kabul')
	`, testProjeID, testUyeID)
	if err != nil {
		t.Fatalf("Proje takımına ekleme yapılamadı: %v", err)
	}

	// Şimdi üye olmalı
	isMember, err = projeRepo.IsProjeUyesi(testProjeID, testUyeID)
	if err != nil {
		t.Fatalf("IsProjeUyesi hatası: %v", err)
	}
	if !isMember {
		t.Error("Kullanıcı projeye eklenmiş ve kabul etmişken IsProjeUyesi false döndü")
	}

	// 5. Yetkilendirme Servis Seviyesi Testi
	// Kayıtlı başka bir üyeyi (yetkisiz) bul
	var unauthorizedUyeID int
	err = database.DB.QueryRow("SELECT uye_id FROM uye WHERE eposta = 'ogrenci@izu.edu.tr'").Scan(&unauthorizedUyeID)
	if err != nil {
		t.Fatalf("Öğrenci üye bulunamadı: %v", err)
	}

	var adminUyeID int
	err = database.DB.QueryRow("SELECT uye_id FROM uye WHERE eposta = 'admin@izu.edu.tr'").Scan(&adminUyeID)
	if err != nil {
		t.Fatalf("Admin üye bulunamadı: %v", err)
	}

	// Yetkisiz akademisyen/öğrenci satın alma talebi göndermeyi denediğinde hata almalı
	unauthorizedTalep := &models.SatinalmaTalebi{
		ProjeID:    testProjeID,
		UyeID:      unauthorizedUyeID,
		KalemID:    testKalemID,
		MalzemeAdi: "Test Malzemesi",
		Miktar:     1,
		BirimFiyat: 1000.00,
		Gerekce:    "Gerekçe",
	}
	err = satinalmaSrv.CreatePurchaseRequest(unauthorizedTalep, "ogrenci")
	if err == nil {
		t.Error("Yetkisiz üyenin satın alma talebi göndermesi engellenmedi")
	}

	// Admin ise yetkisiz üye olmasına rağmen CreatePurchaseRequest'e izin vermeli (bütçe aşılmadığı sürece)
	adminTalep := &models.SatinalmaTalebi{
		ProjeID:    testProjeID,
		UyeID:      adminUyeID,
		KalemID:    testKalemID,
		MalzemeAdi: "Test Malzemesi (Admin)",
		Miktar:     1,
		BirimFiyat: 1000.00,
		Gerekce:    "Admin talebi",
	}
	err = satinalmaSrv.CreatePurchaseRequest(adminTalep, "admin")
	if err != nil {
		t.Errorf("Admin rolüyle satın alma talebi oluşturulamadı: %v", err)
	}

	// 6. Bütçe Rezervasyon Testleri
	// 5000 ₺ toplam bütçe var. 1000 ₺ admin talebi "Beklemede" durumunda.
	// Kalan rezerve bütçe 4000 ₺ olmalı.
	reservedBudget, err := satinalmaRepo.GetReservedBudget(testProjeID, testKalemID)
	if err != nil {
		t.Fatalf("GetReservedBudget hatası: %v", err)
	}
	if reservedBudget != 4000.00 {
		t.Errorf("Beklenen kalan rezerv bütçe 4000.00, alınan: %.2f", reservedBudget)
	}

	// Akademisyen üye olarak kalan rezerv bütçeden fazla (örneğin 4500 ₺) bir talep göndermeye çalışsın -> Hata almalı
	overBudgetTalep := &models.SatinalmaTalebi{
		ProjeID:    testProjeID,
		UyeID:      testUyeID,
		KalemID:    testKalemID,
		MalzemeAdi: "Pahalı Malzeme",
		Miktar:     1,
		BirimFiyat: 4500.00,
		Gerekce:    "Gerekçe",
	}
	err = satinalmaSrv.CreatePurchaseRequest(overBudgetTalep, "akademisyen")
	if err == nil {
		t.Error("Rezerve bütçeyi aşan satın alma talebi oluşturulabildi")
	}

	// Akademisyen bütçe sınırları dahilinde (örneğin 3000 ₺) talep göndersin -> Başarılı olmalı
	validTalep := &models.SatinalmaTalebi{
		ProjeID:    testProjeID,
		UyeID:      testUyeID,
		KalemID:    testKalemID,
		MalzemeAdi: "Uygun Malzeme",
		Miktar:     1,
		BirimFiyat: 3000.00,
		Gerekce:    "Gerekçe",
	}
	err = satinalmaSrv.CreatePurchaseRequest(validTalep, "akademisyen")
	if err != nil {
		t.Errorf("Geçerli satın alma talebi oluşturulurken hata alındı: %v", err)
	}

	// 7. GetPurchaseRequestsByProject Yetki Testi
	// Yetkisiz kullanıcı projeyi listeleyememeli
	_, err = satinalmaSrv.GetPurchaseRequestsByProject(testProjeID, unauthorizedUyeID, "ogrenci")
	if err == nil {
		t.Error("Yetkisiz üyenin satın alma taleplerini listelemesi engellenmedi")
	}

	// TTO veya Admin listeleyebilmeli
	list, err := satinalmaSrv.GetPurchaseRequestsByProject(testProjeID, adminUyeID, "tto")
	if err != nil {
		t.Errorf("TTO rolüyle listeleme yaparken hata alındı: %v", err)
	}
	if len(list) < 2 {
		t.Errorf("Beklenen en az 2 satın alma talebi listesi, bulunan: %d", len(list))
	}
}

// NullStringHelper sql.NullString'leri oluşturmak için yardımcı
func NullStringHelper(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
