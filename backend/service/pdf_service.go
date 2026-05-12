package service

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"bap_ai/backend/repository"

	"github.com/jung-kurt/gofpdf"
)

// PdfService yapısı, PDF oluşturma iş mantığını barındırır.
type PdfService struct {
	ProjeRepo *repository.ProjeRepository
	AdminRepo *repository.AdminRepository
}

// NewPdfService yeni bir PdfService oluşturur.
func NewPdfService(projeRepo *repository.ProjeRepository, adminRepo *repository.AdminRepository) *PdfService {
	return &PdfService{
		ProjeRepo: projeRepo,
		AdminRepo: adminRepo,
	}
}

// GenerateProjectPDF proje verilerini alıp İZÜ kurumsal temalı PDF olarak üretir.
func (s *PdfService) GenerateProjectPDF(projeID int) ([]byte, error) {
	// Proje detaylarını admin repository'den çek (tüm ilişkili veriler dahil)
	detail, err := s.AdminRepo.GetProjectDetailsForAdmin(projeID)
	if err != nil {
		return nil, fmt.Errorf("proje detayları alınamadı: %w", err)
	}

	// PDF belgesi oluştur (A4, dikey, milimetre)
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 20)

	// Türkçe karakter desteği için UTF-8 çeviri haritası
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// İlk sayfayı ekle
	pdf.AddPage()

	// ─── LOGO ve BAŞLIK ───
	addHeader(pdf, tr)

	// ─── PROJE GENEL BİLGİLERİ ───
	addSectionTitle(pdf, tr, "1. PROJE GENEL BİLGİLERİ")

	addTableRow(pdf, tr, "Proje Başlığı (TR)", detail.Proje.BaslikTr)
	addTableRow(pdf, tr, "Proje Başlığı (EN)", detail.Proje.BaslikEn)
	addTableRow(pdf, tr, "BAP Türü", detail.Proje.BapTuru)
	addTableRow(pdf, tr, "Durum", detail.Proje.DurumAdi)
	addTableRow(pdf, tr, "Proje Süresi", fmt.Sprintf("%d Ay", detail.Proje.SureAy))
	addTableRow(pdf, tr, "Toplam Bütçe", fmt.Sprintf("%.2f ₺", detail.Proje.ToplamButce))
	addTableRow(pdf, tr, "Etik Kurul Onayı", boolToStr(detail.Proje.EtikKurul))
	addTableRow(pdf, tr, "Yürütücü", detail.YurutucuAd)
	addTableRow(pdf, tr, "Oluşturma Tarihi", detail.Proje.OlusturmaTarihi.Format("02.01.2006"))

	pdf.Ln(4)

	// ─── AKADEMİK DETAYLAR ───
	if detail.ProjeDetay != nil {
		addSectionTitle(pdf, tr, "2. AKADEMİK DETAYLAR")

		if detail.ProjeDetay.Ozet != "" {
			addSubTitle(pdf, tr, "Proje Özeti")
			addMultiLineText(pdf, tr, detail.ProjeDetay.Ozet)
		}
		if detail.ProjeDetay.AnahtarKelimeler != "" {
			addSubTitle(pdf, tr, "Anahtar Kelimeler")
			addMultiLineText(pdf, tr, detail.ProjeDetay.AnahtarKelimeler)
		}
		if detail.ProjeDetay.Hedefler != "" {
			addSubTitle(pdf, tr, "Hedefler")
			addMultiLineText(pdf, tr, detail.ProjeDetay.Hedefler)
		}
		if detail.ProjeDetay.Ozgunluk != "" {
			addSubTitle(pdf, tr, "Özgünlük")
			addMultiLineText(pdf, tr, detail.ProjeDetay.Ozgunluk)
		}
		if detail.ProjeDetay.Metodoloji != "" {
			addSubTitle(pdf, tr, "Metodoloji")
			addMultiLineText(pdf, tr, detail.ProjeDetay.Metodoloji)
		}

		pdf.Ln(4)
	}

	// ─── ARAŞTIRMA BİLGİSİ ───
	if detail.ArastirmaBilgi != "" {
		addSectionTitle(pdf, tr, "3. ARAŞTIRMA BİLGİSİ")
		addMultiLineText(pdf, tr, detail.ArastirmaBilgi)
		pdf.Ln(4)
	}

	// ─── TAKIM ÜYELERİ ───
	if len(detail.TakimUyeleri) > 0 {
		addSectionTitle(pdf, tr, "4. PROJE EKİBİ")

		// Tablo başlığı
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		colWidths := []float64{15, 60, 50, 50}
		headers := []string{"#", "Ad Soyad", "Sistem Rolü", "Proje Rolü"}
		for i, h := range headers {
			pdf.CellFormat(colWidths[i], 8, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		// Tablo satırları
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(50, 50, 50)
		for idx, uye := range detail.TakimUyeleri {
			pdf.SetFillColor(245, 247, 250)
			fill := idx%2 == 0
			pdf.CellFormat(colWidths[0], 7, tr(fmt.Sprintf("%d", idx+1)), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(colWidths[1], 7, tr(uye.AdSoyad), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(colWidths[2], 7, tr(uye.Rol), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(colWidths[3], 7, tr(uye.ProjeRol), "1", 0, "C", fill, 0, "")
			pdf.Ln(-1)
		}
		pdf.Ln(4)
	}

	// ─── BÜTÇE KALEMLERİ ───
	if len(detail.Butceler) > 0 {
		addSectionTitle(pdf, tr, "5. BÜTÇE KALEMLERİ")

		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		bColWidths := []float64{15, 55, 50, 30, 30}
		bHeaders := []string{"#", "Kategori", "Açıklama", "Birim Fiyat", "Toplam"}
		for i, h := range bHeaders {
			pdf.CellFormat(bColWidths[i], 8, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(50, 50, 50)
		var genelToplam float64
		for idx, b := range detail.Butceler {
			fill := idx%2 == 0
			pdf.SetFillColor(245, 247, 250)
			pdf.CellFormat(bColWidths[0], 7, tr(fmt.Sprintf("%d", idx+1)), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(bColWidths[1], 7, tr(b.KategoriAdi), "1", 0, "L", fill, 0, "")

			// Açıklama uzunsa kısalt
			aciklama := b.Aciklama
			if len(aciklama) > 35 {
				aciklama = aciklama[:35] + "..."
			}
			pdf.CellFormat(bColWidths[2], 7, tr(aciklama), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(bColWidths[3], 7, tr(fmt.Sprintf("%.2f", b.BirimFiyat)), "1", 0, "R", fill, 0, "")
			pdf.CellFormat(bColWidths[4], 7, tr(fmt.Sprintf("%.2f", b.ToplamFiyat)), "1", 0, "R", fill, 0, "")
			pdf.Ln(-1)
			genelToplam += b.ToplamFiyat
		}

		// Toplam satırı
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		pdf.CellFormat(bColWidths[0]+bColWidths[1]+bColWidths[2]+bColWidths[3], 8, tr("GENEL TOPLAM"), "1", 0, "R", true, 0, "")
		pdf.CellFormat(bColWidths[4], 8, tr(fmt.Sprintf("%.2f ₺", genelToplam)), "1", 0, "R", true, 0, "")
		pdf.Ln(-1)
		pdf.Ln(4)
	}

	// ─── İŞ PAKETLERİ ───
	if len(detail.IsPaketleri) > 0 {
		addSectionTitle(pdf, tr, "6. İŞ PAKETLERİ")

		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		ipColWidths := []float64{15, 55, 50, 30, 30}
		ipHeaders := []string{"#", "Paket Adı", "Amacı", "Başlangıç", "Bitiş"}
		for i, h := range ipHeaders {
			pdf.CellFormat(ipColWidths[i], 8, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(50, 50, 50)
		for idx, ip := range detail.IsPaketleri {
			fill := idx%2 == 0
			pdf.SetFillColor(245, 247, 250)
			pdf.CellFormat(ipColWidths[0], 7, tr(fmt.Sprintf("%d", idx+1)), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(ipColWidths[1], 7, tr(truncateStr(ip.PaketAdi, 35)), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(ipColWidths[2], 7, tr(truncateStr(ip.PaketAmaci, 35)), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(ipColWidths[3], 7, tr(ip.BaslangicTarihi), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(ipColWidths[4], 7, tr(ip.BitisTarihi), "1", 0, "C", fill, 0, "")
			pdf.Ln(-1)
		}
		pdf.Ln(4)
	}

	// ─── RİSK YÖNETİMİ ───
	if len(detail.Riskler) > 0 {
		addSectionTitle(pdf, tr, "7. RİSK YÖNETİMİ")

		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		rColWidths := []float64{15, 80, 85}
		rHeaders := []string{"#", "Risk Açıklaması", "Çözüm Planı"}
		for i, h := range rHeaders {
			pdf.CellFormat(rColWidths[i], 8, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(50, 50, 50)
		for idx, rk := range detail.Riskler {
			fill := idx%2 == 0
			pdf.SetFillColor(245, 247, 250)
			pdf.CellFormat(rColWidths[0], 7, tr(fmt.Sprintf("%d", idx+1)), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(rColWidths[1], 7, tr(truncateStr(rk.RiskAciklamasi, 55)), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(rColWidths[2], 7, tr(truncateStr(rk.CozumPlani, 55)), "1", 0, "L", fill, 0, "")
			pdf.Ln(-1)
		}
		pdf.Ln(4)
	}

	// ─── PROJE ÇIKTILARI ───
	if len(detail.Ciktilar) > 0 {
		addSectionTitle(pdf, tr, "8. PROJE ÇIKTILARI")

		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		cColWidths := []float64{15, 45, 70, 50}
		cHeaders := []string{"#", "Çıktı Türü", "Açıklama", "Periyot"}
		for i, h := range cHeaders {
			pdf.CellFormat(cColWidths[i], 8, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(50, 50, 50)
		for idx, ck := range detail.Ciktilar {
			fill := idx%2 == 0
			pdf.SetFillColor(245, 247, 250)
			pdf.CellFormat(cColWidths[0], 7, tr(fmt.Sprintf("%d", idx+1)), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(cColWidths[1], 7, tr(truncateStr(ck.CiktiTuru, 30)), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(cColWidths[2], 7, tr(truncateStr(ck.Aciklama, 45)), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(cColWidths[3], 7, tr(truncateStr(ck.CiktiPeriyodu, 30)), "1", 0, "L", fill, 0, "")
			pdf.Ln(-1)
		}
		pdf.Ln(4)
	}

	// ─── YAYIN BİLGİLERİ ───
	if len(detail.Yayinlar) > 0 {
		addSectionTitle(pdf, tr, "9. YAYIN BİLGİLERİ")

		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		yColWidths := []float64{15, 55, 65, 45}
		yHeaders := []string{"#", "Yayın Türü", "Yayın Çıktısı", "Tahmini Tarih"}
		for i, h := range yHeaders {
			pdf.CellFormat(yColWidths[i], 8, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(50, 50, 50)
		for idx, y := range detail.Yayinlar {
			fill := idx%2 == 0
			pdf.SetFillColor(245, 247, 250)
			pdf.CellFormat(yColWidths[0], 7, tr(fmt.Sprintf("%d", idx+1)), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(yColWidths[1], 7, tr(truncateStr(y.YayinTuru, 35)), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(yColWidths[2], 7, tr(truncateStr(y.YayinCiktisi, 40)), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(yColWidths[3], 7, tr(truncateStr(y.TahminiYayinTarihi, 30)), "1", 0, "C", fill, 0, "")
			pdf.Ln(-1)
		}
		pdf.Ln(4)
	}

	// ─── ALT BİLGİ (FOOTER) ───
	addFooter(pdf, tr)

	// PDF'i byte array olarak döndür
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("PDF oluşturulurken hata: %w", err)
	}

	return buf.Bytes(), nil
}

// ─────────── YARDIMCI FONKSİYONLAR ───────────

// addHeader PDF başlığına İZÜ logosunu ve kurumsal bilgiyi ekler
func addHeader(pdf *gofpdf.Fpdf, tr func(string) string) {
	// İZÜ logosu ekleme (dosya mevcutsa)
	logoPath := "frontend/static/images/izu_logo.png"
	pdf.ImageOptions(logoPath, 15, 10, 25, 0, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")

	// Üniversite adı
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(38, 74, 150)
	pdf.SetXY(45, 12)
	pdf.CellFormat(0, 7, tr("İstanbul Sabahattin Zaim Üniversitesi"), "", 1, "L", false, 0, "")

	// Alt başlık
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(193, 158, 103) // Gold renk
	pdf.SetXY(45, 20)
	pdf.CellFormat(0, 6, tr("Bilimsel Araştırma Projeleri (BAP) Koordinatörlüğü"), "", 1, "L", false, 0, "")

	// Belge başlığı
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetXY(45, 28)
	pdf.CellFormat(0, 6, tr("PROJE BAŞVURU FORMU"), "", 1, "L", false, 0, "")

	// Ayırıcı çizgi
	pdf.SetDrawColor(38, 74, 150)
	pdf.SetLineWidth(0.8)
	pdf.Line(15, 38, 195, 38)

	pdf.Ln(8)
	pdf.SetY(42)
}

// addSectionTitle bölüm başlığı ekler (mavi arka planlı)
func addSectionTitle(pdf *gofpdf.Fpdf, tr func(string) string, title string) {
	// Sayfa sonu kontrolü: başlık için yeterli alan yoksa yeni sayfa ekle
	if pdf.GetY() > 260 {
		pdf.AddPage()
		addHeaderSmall(pdf, tr)
	}

	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetFillColor(38, 74, 150)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(180, 8, "  "+tr(title), "", 1, "L", true, 0, "")
	pdf.SetTextColor(50, 50, 50)
	pdf.Ln(3)
}

// addHeaderSmall sonraki sayfalarda küçük header ekler
func addHeaderSmall(pdf *gofpdf.Fpdf, tr func(string) string) {
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(0, 5, tr("İZÜ BAP - Proje Başvuru Formu"), "", 1, "R", false, 0, "")
	pdf.SetDrawColor(200, 200, 200)
	pdf.SetLineWidth(0.3)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(4)
}

// addSubTitle alt başlık ekler
func addSubTitle(pdf *gofpdf.Fpdf, tr func(string) string, title string) {
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(38, 74, 150)
	pdf.CellFormat(0, 6, tr(title), "", 1, "L", false, 0, "")
	pdf.SetTextColor(50, 50, 50)
}

// addTableRow anahtar-değer formatında tablo satırı ekler
func addTableRow(pdf *gofpdf.Fpdf, tr func(string) string, key string, value string) {
	if value == "" {
		value = "-"
	}
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(245, 247, 250)
	pdf.CellFormat(60, 7, "  "+tr(key), "1", 0, "L", true, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(120, 7, "  "+tr(value), "1", 1, "L", false, 0, "")
}

// addMultiLineText çok satırlı metin ekler
func addMultiLineText(pdf *gofpdf.Fpdf, tr func(string) string, text string) {
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(60, 60, 60)
	// Uzun metni birden fazla satıra böl
	pdf.MultiCell(180, 5, tr(text), "", "L", false)
	pdf.Ln(2)
	pdf.SetTextColor(50, 50, 50)
}

// addFooter sayfa sonuna tarih ve imza alanı ekler
func addFooter(pdf *gofpdf.Fpdf, tr func(string) string) {
	pdf.Ln(10)

	// Ayırıcı çizgi
	pdf.SetDrawColor(38, 74, 150)
	pdf.SetLineWidth(0.5)
	y := pdf.GetY()
	if y > 250 {
		pdf.AddPage()
		y = pdf.GetY()
	}
	pdf.Line(15, y, 195, y)
	pdf.Ln(5)

	// Oluşturma tarihi
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(0, 5, tr(fmt.Sprintf("Bu belge %s tarihinde BAP Sistemi tarafından otomatik olarak oluşturulmuştur.", time.Now().Format("02.01.2006 15:04"))), "", 1, "C", false, 0, "")

	// İmza alanları
	pdf.Ln(15)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(90, 5, tr("Proje Yürütücüsü"), "T", 0, "C", false, 0, "")
	pdf.CellFormat(90, 5, tr("BAP Koordinatörü"), "T", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(90, 5, tr("İmza / Tarih"), "", 0, "C", false, 0, "")
	pdf.CellFormat(90, 5, tr("İmza / Tarih"), "", 1, "C", false, 0, "")
}

// boolToStr boolean değeri Türkçe stringe dönüştürür
func boolToStr(val bool) string {
	if val {
		return "Evet"
	}
	return "Hayır"
}

// truncateStr metni belirli bir uzunlukta keser
func truncateStr(s string, maxLen int) string {
	// UTF-8 güvenli kırpma
	runes := []rune(s)
	if len(runes) > maxLen {
		return strings.TrimSpace(string(runes[:maxLen])) + "..."
	}
	return s
}
