package service

import (
	"testing"
	"time"

	"bap_ai/backend/models"
)

// TestCalculateElapsedTime geçen süre hesaplama mantığını test eder.
// Türkçe Yorum: Başlangıç tarihinden referans tarihe kadar geçen süre metninin doğruluğunu kontrol eder.
func TestCalculateElapsedTime(t *testing.T) {
	baslangic := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	refDate := time.Date(2025, 4, 20, 0, 0, 0, 0, time.UTC)

	result := CalculateElapsedTime(baslangic, refDate)
	expected := "3 Ay 5 Gün"

	if result != expected {
		t.Errorf("Beklenen geçen süre '%s', ancak alınan: '%s'", expected, result)
	}
}

// TestCalculateRemainingTime kalan süre hesaplama mantığını test eder.
// Türkçe Yorum: Referans tarihten bitiş tarihine kadar kalan süre metninin doğruluğunu kontrol eder.
func TestCalculateRemainingTime(t *testing.T) {
	bitis := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	refDate := time.Date(2025, 7, 10, 0, 0, 0, 0, time.UTC)

	result := CalculateRemainingTime(bitis, refDate)
	expected := "6 Ay 5 Gün"

	if result != expected {
		t.Errorf("Beklenen kalan süre '%s', ancak alınan: '%s'", expected, result)
	}
}

// TestIsMonthlyReminderDue hatırlatma zamanı kontrol fonksiyonunu test eder.
// Türkçe Yorum: 1 ay geçtikten sonra ve daha önce e-posta atılmadığında true dönmesini test eder.
func TestIsMonthlyReminderDue(t *testing.T) {
	baslangic := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	bitis := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	info := models.ProjeSozlesmeHatirlatmaInfo{
		SozlesmeID:      1,
		ProjeID:         100,
		BaslangicTarihi: baslangic,
		BitisTarihi:     bitis,
		SureAy:          12,
	}

	// 15 gün sonra (Henüz 1 ay dolmadı -> false olmalı)
	ref1 := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	if IsMonthlyReminderDue(info, ref1) {
		t.Errorf("15 gün geçmişken hatırlatma zamanı gelmiş sayılmamalıydı")
	}

	// 1 ay 5 gün sonra (1. ay hatırlatması atılmamış -> true olmalı)
	ref2 := time.Date(2025, 2, 5, 0, 0, 0, 0, time.UTC)
	if !IsMonthlyReminderDue(info, ref2) {
		t.Errorf("1 ay geçmişken ve daha önce atılmamışken hatırlatma zamanı true dönmeliydi")
	}
}
