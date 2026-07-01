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

	// ─── PROJE BİLGİLERİ (Resmi Form Şeması) ───
	addProjeBilgileriSection(pdf, tr, detail)


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
			addSubTitle(pdf, tr, "Amaç ve Hedefler")
			addMultiLineText(pdf, tr, detail.ProjeDetay.Hedefler)
		}
		if detail.ProjeDetay.Ozgunluk != "" {
			addSubTitle(pdf, tr, "Özgün Değer")
			addMultiLineText(pdf, tr, detail.ProjeDetay.Ozgunluk)
		}
		if detail.ProjeDetay.Metodoloji != "" {
			addSubTitle(pdf, tr, "Yöntem")
			addMultiLineText(pdf, tr, detail.ProjeDetay.Metodoloji)
		}

		pdf.Ln(4)
	}

	// ─── ARAŞTIRMA BİLGİSİ ───
	if detail.ArastirmaBilgi != "" {
		addSectionTitle(pdf, tr, "3. ARAŞTIRMA OLANAKLARI")
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

			// HTML etiketlerini temizle ve uzunsa kısalt
			aciklama := stripHTML(b.Aciklama)
			if len([]rune(aciklama)) > 35 {
				aciklama = string([]rune(aciklama)[:35]) + "..."
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
			pdf.CellFormat(ipColWidths[3], 7, tr(fmt.Sprintf("%d. Ay", ip.BaslangicAy)), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(ipColWidths[4], 7, tr(fmt.Sprintf("%d. Ay", ip.BitisAy)), "1", 0, "C", fill, 0, "")
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
			pdf.CellFormat(cColWidths[2], 7, tr(truncateStr(ck.OngorulCikti, 45)), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(cColWidths[3], 7, tr(truncateStr(ck.ZamanAraligi, 30)), "1", 0, "L", fill, 0, "")
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

	// ─── YAYGIN ETKİ — ÖNGÖRÜLEN ÇIKTILAR ───
	if len(detail.YayinEtki) > 0 {
		addSectionTitle(pdf, tr, "10. YAYGIN ETKİ — PROJEDEN ELDE EDİLMESİ ÖNGÖRÜLEN ÇIKTILAR")

		// Başlık satırı
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		yeCols := []float64{65, 75, 40}
		yeHeaders := []string{"Çıktı Türü", "Öngörülen Çıktı(lar)", "Zaman Aralığı"}
		for i, h := range yeHeaders {
			pdf.CellFormat(yeCols[i], 8, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		ciktiTuruLabel := map[string]string{
			"bilimsel_akademik":        "Bilimsel/Akademik Çıktılar",
			"ekonomik_ticari_sosyal":   "Ekonomik/Ticari/Sosyal Çıktılar",
			"arastirmaci_yetistirme":   "Araştırmacı Yetiştirilmesi ve Yeni Proje(ler)",
			"olusturulmasina_yonelik":  "Oluşturulmasına Yönelik Çıktılar",
		}

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(50, 50, 50)
		for idx, ye := range detail.YayinEtki {
			fill := idx%2 == 0
			pdf.SetFillColor(245, 247, 250)
			label := ciktiTuruLabel[ye.CiktiTuru]
			if label == "" {
				label = ye.CiktiTuru
			}
			// Her satır için MultiCell kullan (uzun metin için)
			x := pdf.GetX()
			y := pdf.GetY()
			pdf.MultiCell(yeCols[0], 7, tr(label), "1", "L", fill)
			h0 := pdf.GetY() - y
			pdf.SetXY(x+yeCols[0], y)
			pdf.MultiCell(yeCols[1], 7, tr(truncateStr(ye.OngorulCikti, 80)), "1", "L", fill)
			h1 := pdf.GetY() - y
			maxH := h0
			if h1 > maxH {
				maxH = h1
			}
			_ = maxH
			pdf.SetXY(x+yeCols[0]+yeCols[1], y)
			pdf.CellFormat(yeCols[2], 7, tr(truncateStr(ye.ZamanAraligi, 25)), "1", 0, "C", fill, 0, "")
			pdf.Ln(-1)
		}
		pdf.Ln(4)
	}

	// ─── YAYGIN ETKİ — YAYGINLAŞTIRMA ETKİNLİKLERİ ───
	if len(detail.YayginlastirmaEtkinlikleri) > 0 {
		addSectionTitle(pdf, tr, "11. ÇIKTILARIN PAYLAŞIMI VE YAYGINLAŞTIRILMASI")

		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		eCols := []float64{15, 65, 60, 40}
		eHeaders := []string{"#", "Etkinlik Türü", "Paydaş / Olası Kullanıcılar", "Zaman ve Süre"}
		for i, h := range eHeaders {
			pdf.CellFormat(eCols[i], 8, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(50, 50, 50)
		for idx, e := range detail.YayginlastirmaEtkinlikleri {
			fill := idx%2 == 0
			pdf.SetFillColor(245, 247, 250)
			pdf.CellFormat(eCols[0], 7, tr(fmt.Sprintf("%d", e.SiraNo)), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(eCols[1], 7, tr(truncateStr(e.EtkinlikTuru, 40)), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(eCols[2], 7, tr(truncateStr(e.Paydas, 38)), "1", 0, "L", fill, 0, "")
			pdf.CellFormat(eCols[3], 7, tr(truncateStr(e.ZamanSure, 25)), "1", 0, "C", fill, 0, "")
			pdf.Ln(-1)
		}
		pdf.Ln(4)
	}

	// ─── KAYNAKÇA ───
	if detail.ProjeDetay != nil && detail.ProjeDetay.Kaynakca != "" {
		addSectionTitle(pdf, tr, "12. KAYNAKÇA")
		addMultiLineText(pdf, tr, detail.ProjeDetay.Kaynakca)
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

// addHeader PDF başlığına İZÜ logosunu, kurumsal başlığı ve doküman meta bilgi tablosunu ekler.
// Eklenen düzen: Sol tarafta logo + başlık, sağ tarafta doküman bilgileri tablosu
func addHeader(pdf *gofpdf.Fpdf, tr func(string) string) {
	// Başlık alanı yüksekliği
	headerTop := 10.0

	// ─── SOL TARAF: Logo ───
	logoPath := "frontend/static/images/izu_logo.png"
	pdf.ImageOptions(logoPath, 15, headerTop, 22, 0, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")

	// ─── ORTA KISIM: Doküman Başlığı ───
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(38, 74, 150)
	titleX := 40.0
	titleW := 90.0
	// İlk satır: BAP türü ve destek programı adı
	pdf.SetXY(titleX, headerTop+3)
	pdf.MultiCell(titleW, 6, tr("İZÜ BAP DESTEK PROGRAMI\nPROJE BAŞVURU FORMU"), "", "C", false)

	// ─── SAĞ TARAF: Doküman Meta Bilgi Tablosu ───
	metaX := 135.0
	metaLabelW := 30.0
	metaValueW := 30.0
	metaH := 5.5
	metaY := headerTop

	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetDrawColor(150, 150, 150)
	pdf.SetLineWidth(0.2)

	// Satır 1: Doküman No
	pdf.SetXY(metaX, metaY)
	pdf.CellFormat(metaLabelW, metaH, tr("Doküman No"), "1", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 7)
	pdf.CellFormat(metaValueW, metaH, tr("TTO-FR-695"), "1", 0, "C", false, 0, "")
	metaY += metaH

	// Satır 2: İlk Yayın Tarihi
	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetXY(metaX, metaY)
	pdf.CellFormat(metaLabelW, metaH, tr("İlk Yayın Tarihi"), "1", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 7)
	pdf.CellFormat(metaValueW, metaH, tr("26.02.2024"), "1", 0, "C", false, 0, "")
	metaY += metaH

	// Satır 3: Revizyon Tarihi
	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetXY(metaX, metaY)
	pdf.CellFormat(metaLabelW, metaH, tr("Revizyon Tarihi"), "1", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 7)
	pdf.CellFormat(metaValueW, metaH, tr("24.02.2026"), "1", 0, "C", false, 0, "")
	metaY += metaH

	// Satır 4: Revizyon No
	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetXY(metaX, metaY)
	pdf.CellFormat(metaLabelW, metaH, tr("Revizyon No"), "1", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 7)
	pdf.CellFormat(metaValueW, metaH, tr("02"), "1", 0, "C", false, 0, "")
	metaY += metaH

	// Satır 5: Sayfa (dinamik olarak ayarlanacak — şimdilik placeholder)
	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetXY(metaX, metaY)
	pdf.CellFormat(metaLabelW, metaH, tr("Sayfa"), "1", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 7)
	pdf.SetTextColor(38, 74, 150)
	pageStr := fmt.Sprintf("%d", pdf.PageNo())
	pdf.CellFormat(metaValueW, metaH, tr(pageStr), "1", 0, "C", false, 0, "")
	pdf.SetTextColor(50, 50, 50)

	// ─── Başlık altı ayırıcı çizgi ───
	separatorY := metaY + metaH + 3
	pdf.SetDrawColor(38, 74, 150)
	pdf.SetLineWidth(0.6)
	pdf.Line(15, separatorY, 195, separatorY)

	pdf.SetY(separatorY + 4)
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
	// HTML etiketlerini temizle
	cleanText := stripHTML(text)
	// Uzun metni birden fazla satıra böl
	pdf.MultiCell(180, 5, tr(cleanText), "", "L", false)
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

// stripHTML metin içindeki temel HTML etiketlerini temizler
func stripHTML(h string) string {
	var builder strings.Builder
	inTag := false
	for _, char := range h {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		} else if !inTag {
			builder.WriteRune(char)
		}
	}
	res := builder.String()
	res = strings.ReplaceAll(res, "&nbsp;", " ")
	res = strings.ReplaceAll(res, "&quot;", "\"")
	res = strings.ReplaceAll(res, "&lt;", "<")
	res = strings.ReplaceAll(res, "&gt;", ">")
	res = strings.ReplaceAll(res, "&amp;", "&")
	return strings.TrimSpace(res)
}

// truncateStr metni belirli bir uzunlukta keser
func truncateStr(s string, maxLen int) string {
	// HTML etiketlerini temizle
	cleanStr := stripHTML(s)
	// UTF-8 güvenli kırpma
	runes := []rune(cleanStr)
	if len(runes) > maxLen {
		return strings.TrimSpace(string(runes[:maxLen])) + "..."
	}
	return cleanStr
}

// addProjeBilgileriSection ekteki resmi form şemasındaki "PROJE BİLGİLERİ" bölümünü oluşturur.
// Düzen: Başlık satırı + Proje Başlığı, Proje Yürütücüsü*, Proje Süresi, Proje Toplam Bütçesi tablosu
func addProjeBilgileriSection(pdf *gofpdf.Fpdf, tr func(string) string, detail *repository.ProjectDetail) {
	labelW := 55.0  // Sol sütun genişliği (etiketler)
	valueW := 125.0 // Sağ sütun genişliği (değerler)
	rowH := 10.0    // Satır yüksekliği

	// ─── Bölüm Başlığı: PROJE BİLGİLERİ ───
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetFillColor(245, 247, 250)
	pdf.SetTextColor(38, 74, 150)
	pdf.SetDrawColor(150, 150, 150)
	pdf.SetLineWidth(0.2)
	pdf.CellFormat(labelW+valueW, rowH, "  "+tr("PROJE BİLGİLERİ"), "1", 1, "C", true, 0, "")

	// ─── Satır 0: Proje Numarası ───
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetFillColor(255, 255, 255)
	pdf.CellFormat(labelW, rowH, "  "+tr("Proje Numarası"), "1", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	projeKodu := detail.Proje.ProjeKodu
	if projeKodu == "" {
		projeKodu = "-"
	}
	pdf.CellFormat(valueW, rowH, "  "+tr(projeKodu), "1", 1, "L", false, 0, "")

	// ─── Satır 1: Proje Başlığı ───
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(labelW, rowH, "  "+tr("Proje Başlığı"), "1", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	baslik := detail.Proje.BaslikTr
	if baslik == "" {
		baslik = "-"
	}
	pdf.CellFormat(valueW, rowH, "  "+tr(baslik), "1", 1, "L", false, 0, "")

	// ─── Satır 1b: BAP Türü ───
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(labelW, rowH, "  "+tr("BAP Türü"), "1", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	bapTuru := detail.Proje.BapTuru
	if bapTuru == "" {
		bapTuru = "-"
	}
	pdf.CellFormat(valueW, rowH, "  "+tr(bapTuru), "1", 1, "L", false, 0, "")

	// ─── Satır 2: Proje Yürütücüsü* ───
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(labelW, rowH, "  "+tr("Proje Yürütücüsü*"), "1", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	yurutucu := detail.YurutucuAd
	if yurutucu == "" {
		yurutucu = "-"
	}
	pdf.CellFormat(valueW, rowH, "  "+tr(yurutucu), "1", 1, "L", false, 0, "")

	// ─── Satır 3: Proje Süresi (Ay) ───
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(labelW, rowH, "  "+tr("Proje Süresi (Ay)"), "1", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	sureTxt := fmt.Sprintf("%d", detail.Proje.SureAy)
	pdf.CellFormat(valueW, rowH, "  "+tr(sureTxt), "1", 1, "L", false, 0, "")

	// ─── Satır 4: Proje Toplam Bütçesi (TL) ───
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(labelW, rowH, "  "+tr("Proje Toplam Bütçesi (TL)"), "1", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	butceTxt := fmt.Sprintf("%.2f", detail.Proje.ToplamButce)
	pdf.CellFormat(valueW, rowH, "  "+tr(butceTxt), "1", 1, "L", false, 0, "")

	// ─── Dipnot: *İZÜ öğretim üyesi ───
	pdf.SetFont("Helvetica", "I", 7)
	pdf.SetTextColor(180, 50, 50)
	pdf.CellFormat(0, 5, tr("*İZÜ öğretim üyesi"), "", 1, "L", false, 0, "")
	pdf.SetTextColor(50, 50, 50)

	pdf.Ln(4)
}
