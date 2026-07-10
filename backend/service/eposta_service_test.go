package service

import (
	"strings"
	"testing"

	"bap_ai/configs"
)

// TestGetStatusLabel durum etiketlerinin doğru Türkçe isimlere çevrildiğini doğrular.
// Türkçe Yorum: GetStatusLabel fonksiyonunun beklenen tüm durum anahtarlarını doğru Türkçe karşılıklarına çevirip çevirmediğini test eder.
func TestGetStatusLabel(t *testing.T) {
	s := NewEpostaService(nil, &configs.Config{})

	tests := []struct {
		status string
		want   string
	}{
		{"taslak", "Taslak"},
		{"incelemede", "TTO Ön İnceleme"},
		{"dekan_onayi_bekliyor", "Dekan Onayı Bekliyor"},
		{"dekan_onayladi", "Dekan Onayladı"},
		{"komisyon_bekliyor", "Komisyon Onayı Bekliyor"},
		{"komisyon_onayladi", "Komisyon Onayladı"},
		{"hakem_atama_bekliyor", "Hakem Ataması Bekleniyor"},
		{"hakem_bekliyor", "Hakem Değerlendirmesinde"},
		{"hakem_onayladi", "Hakem Onayladı"},
		{"sozlesme_imza", "Sözleşme / İmza Aşaması"},
		{"tto_aktif", "TTO Onayı Bekliyor"},
		{"onaylandi", "Onaylandı"},
		{"reddedildi", "Reddedildi"},
		{"tamamlandi", "Tamamlandı"},
		{"revizyon", "Revizyon Talebi"},
		{"yururlukte", "Yürürlükte (Aktif)"},
		{"bilinmeyen_durum", "bilinmeyen_durum"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := s.GetStatusLabel(tt.status)
			if got != tt.want {
				t.Errorf("GetStatusLabel(%s) = %s; want %s", tt.status, got, tt.want)
			}
		})
	}
}

// TestFormatEmailTemplate e-posta HTML şablonunun oluşturulmasını doğrular.
// Türkçe Yorum: FormatEmailTemplate fonksiyonunun parametreleri HTML gövdesi içinde doğru şekilde birleştirdiğini test eder.
func TestFormatEmailTemplate(t *testing.T) {
	s := NewEpostaService(nil, &configs.Config{})

	greeting := "Sayın Test Kullanıcısı"
	message := "Projeniz onaylandı."
	projectCode := "2026-BAP-001"
	projectTitle := "Yapay Zeka Destekli BAP Projesi"
	coordinator := "Prof. Dr. Ahmet Yılmaz"
	newStatus := "Onaylandı"
	description := "Bu bir test açıklamasıdır."

	html := s.FormatEmailTemplate(greeting, message, projectCode, projectTitle, coordinator, newStatus, description)

	if !strings.Contains(html, greeting) {
		t.Errorf("HTML şablonu selamlama metnini içermiyor")
	}
	if !strings.Contains(html, message) {
		t.Errorf("HTML şablonu ana mesajı içermiyor")
	}
	if !strings.Contains(html, projectCode) {
		t.Errorf("HTML şablonu proje kodunu içermiyor")
	}
	if !strings.Contains(html, projectTitle) {
		t.Errorf("HTML şablonu proje başlığını içermiyor")
	}
	if !strings.Contains(html, coordinator) {
		t.Errorf("HTML şablonu koordinatör adını içermiyor")
	}
	if !strings.Contains(html, newStatus) {
		t.Errorf("HTML şablonu yeni durum etiketini içermiyor")
	}
	if !strings.Contains(html, description) {
		t.Errorf("HTML şablonu açıklama metnini içermiyor")
	}
}
