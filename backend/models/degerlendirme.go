package models

import "time"

// ProjeDegerlendirme, bir projenin hakemler tarafından değerlendirmesini tutar
type ProjeDegerlendirme struct {
	DegerlendirmeID int       `json:"degerlendirme_id"`
	ProjeID         int       `json:"proje_id"`
	HakemID         int       `json:"hakem_id"`
	Puan            int       `json:"puan"`
	Yorum           string    `json:"yorum"`
	Durum           string    `json:"durum"` // Bekliyor, Onaylandı, Reddedildi, Revizyon
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// HakemProjeOzet, hakem dashboard'ında gösterilecek atanmış projenin özetidir
type HakemProjeOzet struct {
	ProjeID   int    `json:"proje_id"`
	BaslikTr  string `json:"baslik_tr"`
	Tur       string `json:"tur"`
	Durum     string `json:"durum"`     // Projenin genel durumu
	HakemDurum string `json:"hakem_durum"` // Bekliyor, Onaylandı...
	Puan      *int   `json:"puan"`
	Tarih     string `json:"tarih"`
}

// DegerlendirmeRequest, hakemin projeyi puanlama/yorumlama isteğidir
type DegerlendirmeRequest struct {
	ProjeID int    `json:"proje_id" binding:"required"`
	Puan    int    `json:"puan" binding:"required"`
	Yorum   string `json:"yorum" binding:"required"`
	Durum   string `json:"durum" binding:"required"` // Onaylandı, Reddedildi, Revizyon
}
