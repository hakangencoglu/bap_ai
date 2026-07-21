package service

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"

	"github.com/jung-kurt/gofpdf"
)

// pdfFontFamily PDF üretiminde kullanılan gömülü Türkçe uyumlu font ailesidir.
const pdfFontFamily = "DejaVu"

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
// Türkçe Bilgilendirme: PDF çıktısı için yetki parametresi false olarak iletilir (hakem adları maskelenir).
func (s *PdfService) GenerateProjectPDF(projeID int) ([]byte, error) {
	// Proje detaylarını admin repository'den çek (tüm ilişkili veriler dahil)
	detail, err := s.AdminRepo.GetProjectDetailsForAdmin(projeID, false)
	if err != nil {
		return nil, fmt.Errorf("proje detayları alınamadı: %w", err)
	}

	// PDF belgesi oluştur (A4, dikey, milimetre)
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 20)

	// Türkçe karakter desteği için UTF-8 gömülü font (DejaVuSans) kaydı.
	// Standart pdfFontFamily core fontu yalnızca Latin-1 (cp1252) destekler;
	// Türkçe'ye özgü ğ, ş, ı, İ, Ğ, Ş karakterleri bu kodlamada bulunmadığından
	// hem önizlemede hem PDF çıktısında bozuk/eksik görünür. DejaVuSans ise tam
	// Unicode desteği sağlar. İtalik için ayrı dosya olmadığından "I"/"BI" stilleri
	// regular/bold dosyalarına eşlenir.
	const fontDir = "frontend/static/fonts"
	pdf.AddUTF8Font(pdfFontFamily, "", fontDir+"/DejaVuSans.ttf")
	pdf.AddUTF8Font(pdfFontFamily, "B", fontDir+"/DejaVuSans-Bold.ttf")
	pdf.AddUTF8Font(pdfFontFamily, "I", fontDir+"/DejaVuSans.ttf")
	pdf.AddUTF8Font(pdfFontFamily, "BI", fontDir+"/DejaVuSans-Bold.ttf")

	// UTF-8 gömülü font kullanıldığında çeviri haritasına gerek yoktur; metinler
	// doğrudan UTF-8 olarak verilir. tr birim (identity) fonksiyon olarak kalır ki
	// mevcut tr(...) çağrıları değişmeden çalışmaya devam etsin.
	tr := func(s string) string { return s }

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
		pdf.SetFont(pdfFontFamily, "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		colWidths := []float64{15, 60, 50, 50}
		headers := []string{"#", "Ad Soyad", "Sistem Rolü", "Proje Rolü"}
		for i, h := range headers {
			pdf.CellFormat(colWidths[i], 8, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		// Tablo satırları
		pdf.SetFont(pdfFontFamily, "", 9)
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

		pdf.SetFont(pdfFontFamily, "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		bColWidths := []float64{15, 55, 50, 30, 30}
		bHeaders := []string{"#", "Kategori", "Açıklama", "Birim Fiyat", "Toplam"}
		for i, h := range bHeaders {
			pdf.CellFormat(bColWidths[i], 8, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont(pdfFontFamily, "", 9)
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
		pdf.SetFont(pdfFontFamily, "B", 9)
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

		pdf.SetFont(pdfFontFamily, "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		ipColWidths := []float64{15, 55, 50, 30, 30}
		ipHeaders := []string{"#", "Paket Adı", "Amacı", "Başlangıç", "Bitiş"}
		for i, h := range ipHeaders {
			pdf.CellFormat(ipColWidths[i], 8, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont(pdfFontFamily, "", 9)
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

		pdf.SetFont(pdfFontFamily, "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		rColWidths := []float64{15, 80, 85}
		rHeaders := []string{"#", "Risk Açıklaması", "Çözüm Planı"}
		for i, h := range rHeaders {
			pdf.CellFormat(rColWidths[i], 8, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont(pdfFontFamily, "", 9)
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

		pdf.SetFont(pdfFontFamily, "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		cColWidths := []float64{15, 45, 70, 50}
		cHeaders := []string{"#", "Çıktı Türü", "Açıklama", "Periyot"}
		for i, h := range cHeaders {
			pdf.CellFormat(cColWidths[i], 8, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont(pdfFontFamily, "", 9)
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

		pdf.SetFont(pdfFontFamily, "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		yColWidths := []float64{15, 55, 65, 45}
		yHeaders := []string{"#", "Yayın Türü", "Yayın Çıktısı", "Tahmini Tarih"}
		for i, h := range yHeaders {
			pdf.CellFormat(yColWidths[i], 8, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont(pdfFontFamily, "", 9)
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
		pdf.SetFont(pdfFontFamily, "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		yeCols := []float64{65, 75, 40}
		yeHeaders := []string{"Çıktı Türü", "Öngörülen Çıktı(lar)", "Zaman Aralığı"}
		for i, h := range yeHeaders {
			pdf.CellFormat(yeCols[i], 8, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		ciktiTuruLabel := map[string]string{
			"bilimsel_akademik":       "Bilimsel/Akademik Çıktılar",
			"ekonomik_ticari_sosyal":  "Ekonomik/Ticari/Sosyal Çıktılar",
			"arastirmaci_yetistirme":  "Araştırmacı Yetiştirilmesi ve Yeni Proje(ler)",
			"olusturulmasina_yonelik": "Oluşturulmasına Yönelik Çıktılar",
		}

		pdf.SetFont(pdfFontFamily, "", 9)
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

		pdf.SetFont(pdfFontFamily, "B", 9)
		pdf.SetFillColor(38, 74, 150)
		pdf.SetTextColor(255, 255, 255)
		eCols := []float64{15, 65, 60, 40}
		eHeaders := []string{"#", "Etkinlik Türü", "Paydaş / Olası Kullanıcılar", "Zaman ve Süre"}
		for i, h := range eHeaders {
			pdf.CellFormat(eCols[i], 8, tr(h), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont(pdfFontFamily, "", 9)
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
	pdf.SetFont(pdfFontFamily, "B", 11)
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

	pdf.SetFont(pdfFontFamily, "B", 7)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetDrawColor(150, 150, 150)
	pdf.SetLineWidth(0.2)

	// Satır 1: Doküman No
	pdf.SetXY(metaX, metaY)
	pdf.CellFormat(metaLabelW, metaH, tr("Doküman No"), "1", 0, "L", false, 0, "")
	pdf.SetFont(pdfFontFamily, "", 7)
	pdf.CellFormat(metaValueW, metaH, tr("TTO-FR-695"), "1", 0, "C", false, 0, "")
	metaY += metaH

	// Satır 2: İlk Yayın Tarihi
	pdf.SetFont(pdfFontFamily, "B", 7)
	pdf.SetXY(metaX, metaY)
	pdf.CellFormat(metaLabelW, metaH, tr("İlk Yayın Tarihi"), "1", 0, "L", false, 0, "")
	pdf.SetFont(pdfFontFamily, "", 7)
	pdf.CellFormat(metaValueW, metaH, tr("26.02.2024"), "1", 0, "C", false, 0, "")
	metaY += metaH

	// Satır 3: Revizyon Tarihi
	pdf.SetFont(pdfFontFamily, "B", 7)
	pdf.SetXY(metaX, metaY)
	pdf.CellFormat(metaLabelW, metaH, tr("Revizyon Tarihi"), "1", 0, "L", false, 0, "")
	pdf.SetFont(pdfFontFamily, "", 7)
	pdf.CellFormat(metaValueW, metaH, tr("24.02.2026"), "1", 0, "C", false, 0, "")
	metaY += metaH

	// Satır 4: Revizyon No
	pdf.SetFont(pdfFontFamily, "B", 7)
	pdf.SetXY(metaX, metaY)
	pdf.CellFormat(metaLabelW, metaH, tr("Revizyon No"), "1", 0, "L", false, 0, "")
	pdf.SetFont(pdfFontFamily, "", 7)
	pdf.CellFormat(metaValueW, metaH, tr("02"), "1", 0, "C", false, 0, "")
	metaY += metaH

	// Satır 5: Sayfa (dinamik olarak ayarlanacak — şimdilik placeholder)
	pdf.SetFont(pdfFontFamily, "B", 7)
	pdf.SetXY(metaX, metaY)
	pdf.CellFormat(metaLabelW, metaH, tr("Sayfa"), "1", 0, "L", false, 0, "")
	pdf.SetFont(pdfFontFamily, "B", 7)
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

	pdf.SetFont(pdfFontFamily, "B", 11)
	pdf.SetFillColor(38, 74, 150)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(180, 8, "  "+tr(title), "", 1, "L", true, 0, "")
	pdf.SetTextColor(50, 50, 50)
	pdf.Ln(3)
}

// addHeaderSmall sonraki sayfalarda küçük header ekler
func addHeaderSmall(pdf *gofpdf.Fpdf, tr func(string) string) {
	pdf.SetFont(pdfFontFamily, "I", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(0, 5, tr("İZÜ BAP - Proje Başvuru Formu"), "", 1, "R", false, 0, "")
	pdf.SetDrawColor(200, 200, 200)
	pdf.SetLineWidth(0.3)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(4)
}

// addSubTitle alt başlık ekler
func addSubTitle(pdf *gofpdf.Fpdf, tr func(string) string, title string) {
	pdf.SetFont(pdfFontFamily, "B", 10)
	pdf.SetTextColor(38, 74, 150)
	pdf.CellFormat(0, 6, tr(title), "", 1, "L", false, 0, "")
	pdf.SetTextColor(50, 50, 50)
}

// addTableRow anahtar-değer formatında tablo satırı ekler
func addTableRow(pdf *gofpdf.Fpdf, tr func(string) string, key string, value string) {
	if value == "" {
		value = "-"
	}
	pdf.SetFont(pdfFontFamily, "B", 9)
	pdf.SetFillColor(245, 247, 250)
	pdf.CellFormat(60, 7, "  "+tr(key), "1", 0, "L", true, 0, "")
	pdf.SetFont(pdfFontFamily, "", 9)
	pdf.CellFormat(120, 7, "  "+tr(value), "1", 1, "L", false, 0, "")
}

// addMultiLineText çok satırlı metin ekler
func addMultiLineText(pdf *gofpdf.Fpdf, tr func(string) string, text string) {
	pdf.SetFont(pdfFontFamily, "", 9)
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
	pdf.SetFont(pdfFontFamily, "I", 8)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(0, 5, tr(fmt.Sprintf("Bu belge %s tarihinde BAP Sistemi tarafından otomatik olarak oluşturulmuştur.", time.Now().Format("02.01.2006 15:04"))), "", 1, "C", false, 0, "")

	// İmza alanları
	pdf.Ln(15)
	pdf.SetFont(pdfFontFamily, "", 9)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(90, 5, tr("Proje Yürütücüsü"), "T", 0, "C", false, 0, "")
	pdf.CellFormat(90, 5, tr("BAP Koordinatörü"), "T", 1, "C", false, 0, "")
	pdf.SetFont(pdfFontFamily, "I", 8)
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
	// Türkçe Yorum: HTML etiketlerini temizlemeden önce satır kesmelerini ve listeleri düzgün biçimlendirilmiş metne dönüştürür.
	r := h
	r = strings.ReplaceAll(r, "<br>", "\n")
	r = strings.ReplaceAll(r, "<br/>", "\n")
	r = strings.ReplaceAll(r, "<br />", "\n")
	r = strings.ReplaceAll(r, "</p>", "\n")
	r = strings.ReplaceAll(r, "</div>", "\n")
	r = strings.ReplaceAll(r, "<li>", "\n • ")
	r = strings.ReplaceAll(r, "</li>", "")

	var builder strings.Builder
	inTag := false
	for _, char := range r {
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

	// Ardışık yeni satır karakterlerini temizle
	lines := strings.Split(res, "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanLines = append(cleanLines, line)
		}
	}
	return strings.TrimSpace(strings.Join(cleanLines, "\n"))
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
	pdf.SetFont(pdfFontFamily, "B", 11)
	pdf.SetFillColor(245, 247, 250)
	pdf.SetTextColor(38, 74, 150)
	pdf.SetDrawColor(150, 150, 150)
	pdf.SetLineWidth(0.2)
	pdf.CellFormat(labelW+valueW, rowH, "  "+tr("PROJE BİLGİLERİ"), "1", 1, "C", true, 0, "")

	// ─── Satır 0: Proje Numarası ───
	pdf.SetFont(pdfFontFamily, "B", 9)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetFillColor(255, 255, 255)
	pdf.CellFormat(labelW, rowH, "  "+tr("Proje Numarası"), "1", 0, "L", false, 0, "")
	pdf.SetFont(pdfFontFamily, "", 9)
	projeKodu := detail.Proje.ProjeKodu
	if projeKodu == "" {
		projeKodu = "-"
	}
	pdf.CellFormat(valueW, rowH, "  "+tr(projeKodu), "1", 1, "L", false, 0, "")

	// ─── Satır 1: Proje Başlığı ───
	pdf.SetFont(pdfFontFamily, "B", 9)
	pdf.CellFormat(labelW, rowH, "  "+tr("Proje Başlığı"), "1", 0, "L", false, 0, "")
	pdf.SetFont(pdfFontFamily, "", 9)
	baslik := detail.Proje.BaslikTr
	if baslik == "" {
		baslik = "-"
	}
	pdf.CellFormat(valueW, rowH, "  "+tr(baslik), "1", 1, "L", false, 0, "")

	// ─── Satır 1b: BAP Türü ───
	pdf.SetFont(pdfFontFamily, "B", 9)
	pdf.CellFormat(labelW, rowH, "  "+tr("BAP Türü"), "1", 0, "L", false, 0, "")
	pdf.SetFont(pdfFontFamily, "", 9)
	bapTuru := detail.Proje.BapTuru
	if bapTuru == "" {
		bapTuru = "-"
	}
	pdf.CellFormat(valueW, rowH, "  "+tr(bapTuru), "1", 1, "L", false, 0, "")

	// ─── Satır 2: Proje Yürütücüsü* ───
	pdf.SetFont(pdfFontFamily, "B", 9)
	pdf.CellFormat(labelW, rowH, "  "+tr("Proje Yürütücüsü*"), "1", 0, "L", false, 0, "")
	pdf.SetFont(pdfFontFamily, "", 9)
	yurutucu := detail.YurutucuAd
	if yurutucu == "" {
		yurutucu = "-"
	}
	pdf.CellFormat(valueW, rowH, "  "+tr(yurutucu), "1", 1, "L", false, 0, "")

	// ─── Satır 3: Proje Süresi (Ay) ───
	pdf.SetFont(pdfFontFamily, "B", 9)
	pdf.CellFormat(labelW, rowH, "  "+tr("Proje Süresi (Ay)"), "1", 0, "L", false, 0, "")
	pdf.SetFont(pdfFontFamily, "", 9)
	sureTxt := fmt.Sprintf("%d", detail.Proje.SureAy)
	pdf.CellFormat(valueW, rowH, "  "+tr(sureTxt), "1", 1, "L", false, 0, "")

	// ─── Satır 4: Proje Toplam Bütçesi (TL) ───
	pdf.SetFont(pdfFontFamily, "B", 9)
	pdf.CellFormat(labelW, rowH, "  "+tr("Proje Toplam Bütçesi (TL)"), "1", 0, "L", false, 0, "")
	pdf.SetFont(pdfFontFamily, "", 9)
	butceTxt := fmt.Sprintf("%.2f", detail.Proje.ToplamButce)
	pdf.CellFormat(valueW, rowH, "  "+tr(butceTxt), "1", 1, "L", false, 0, "")

	// ─── Dipnot: *İZÜ öğretim üyesi ───
	pdf.SetFont(pdfFontFamily, "I", 7)
	pdf.SetTextColor(180, 50, 50)
	pdf.CellFormat(0, 5, tr("*İZÜ öğretim üyesi"), "", 1, "L", false, 0, "")
	pdf.SetTextColor(50, 50, 50)

	pdf.Ln(4)
}

// GenerateCommissionMeetingPDF komisyon toplantı tutanağını PDF olarak oluşturur.
// Türkçe Yorum: Toplantı genel bilgilerini, gündemini, kararlarını ve katılımcı listesini İZÜ kurumsal şablonunda PDF dosyasına dönüştürür.
func (s *PdfService) GenerateCommissionMeetingPDF(meeting *models.KomisyonToplantisi) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 20)

	const fontDir = "frontend/static/fonts"
	pdf.AddUTF8Font(pdfFontFamily, "", fontDir+"/DejaVuSans.ttf")
	pdf.AddUTF8Font(pdfFontFamily, "B", fontDir+"/DejaVuSans-Bold.ttf")
	pdf.AddUTF8Font(pdfFontFamily, "I", fontDir+"/DejaVuSans.ttf")
	pdf.AddUTF8Font(pdfFontFamily, "BI", fontDir+"/DejaVuSans-Bold.ttf")

	tr := func(s string) string { return s }

	pdf.AddPage()

	// Başlık ve Logo Alanı
	pdf.SetFont(pdfFontFamily, "B", 13)
	pdf.SetTextColor(38, 74, 150) // Kurumsal mavi
	pdf.CellFormat(0, 10, tr("T.C. İSTANBUL SABAHATTİN ZAİM ÜNİVERSİTESİ"), "", 1, "C", false, 0, "")
	pdf.SetFont(pdfFontFamily, "B", 11)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 6, tr("BİLİMSEL ARAŞTIRMA PROJELERİ (BAP) KOMİSYONU"), "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 6, tr("TOPLANTI KARAR TUTANAĞI"), "", 1, "C", false, 0, "")

	pdf.Ln(4)
	pdf.SetDrawColor(38, 74, 150)
	pdf.SetLineWidth(0.8)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(6)

	// Toplantı Temel Bilgileri
	pdf.SetFont(pdfFontFamily, "B", 10)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(35, 6, tr("Toplantı No:"), "", 0, "L", false, 0, "")
	pdf.SetFont(pdfFontFamily, "", 10)
	pdf.CellFormat(65, 6, tr(meeting.ToplantiNo), "", 0, "L", false, 0, "")

	pdf.SetFont(pdfFontFamily, "B", 10)
	pdf.CellFormat(35, 6, tr("Toplantı Tarihi:"), "", 0, "L", false, 0, "")
	pdf.SetFont(pdfFontFamily, "", 10)
	tarihStr := meeting.Tarih.Format("02.01.2006")
	pdf.CellFormat(55, 6, tr(tarihStr), "", 1, "L", false, 0, "")

	pdf.Ln(6)

	// ─── Gündem Başlığı ve İçeriği ───
	pdf.SetFont(pdfFontFamily, "B", 11)
	pdf.SetFillColor(240, 243, 248)
	pdf.SetTextColor(38, 74, 150)
	pdf.CellFormat(0, 8, "  "+tr("TOPLANTI GÜNDEMİ"), "B", 1, "L", true, 0, "")
	pdf.Ln(2)
	pdf.SetFont(pdfFontFamily, "", 10)
	pdf.SetTextColor(50, 50, 50)
	pdf.MultiCell(0, 5, tr(stripHTML(meeting.Gundem)), "", "L", false)
	pdf.Ln(6)

	// ─── Karar Başlığı ve İçeriği ───
	pdf.SetFont(pdfFontFamily, "B", 11)
	pdf.SetFillColor(240, 243, 248)
	pdf.SetTextColor(38, 74, 150)
	pdf.CellFormat(0, 8, "  "+tr("TOPLANTI KARARLARI"), "B", 1, "L", true, 0, "")
	pdf.Ln(2)
	pdf.SetFont(pdfFontFamily, "", 10)
	pdf.SetTextColor(50, 50, 50)
	pdf.MultiCell(0, 5, tr(stripHTML(meeting.Karar)), "", "L", false)
	pdf.Ln(8)

	// ─── Katılımcılar Başlığı ve İmzalar ───
	pdf.SetFont(pdfFontFamily, "B", 11)
	pdf.SetFillColor(240, 243, 248)
	pdf.SetTextColor(38, 74, 150)
	pdf.CellFormat(0, 8, "  "+tr("KATILIMCI YOKLAMA VE İMZA LİSTESİ"), "B", 1, "L", true, 0, "")
	pdf.Ln(4)

	// Katılımcı Tablosu Başlıkları
	pdf.SetFont(pdfFontFamily, "B", 9)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFillColor(38, 74, 150)
	pdf.CellFormat(60, 8, tr("Adı Soyadı"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(55, 8, tr("Unvan / Bölüm"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, tr("Katılım Durumu"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, tr("İmza"), "1", 1, "C", true, 0, "")

	// Katılımcı Verileri
	pdf.SetFont(pdfFontFamily, "", 9)
	pdf.SetTextColor(50, 50, 50)
	for i, k := range meeting.Katilimcilar {
		// Satır arka planı alternatif renklendirme
		fill := false
		if i%2 == 1 {
			pdf.SetFillColor(245, 247, 250)
			fill = true
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		fullName := fmt.Sprintf("%s %s", k.Ad, k.Soyad)
		unvanBolum := k.Unvan
		if k.Bolum != "" {
			if unvanBolum != "" {
				unvanBolum += " - "
			}
			unvanBolum += k.Bolum
		}

		katilimDurum := "KATILMADI"
		if k.Katildi {
			katilimDurum = "KATILDI"
		}

		pdf.CellFormat(60, 10, "  "+tr(fullName), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(55, 10, "  "+tr(unvanBolum), "1", 0, "L", fill, 0, "")

		if k.Katildi {
			pdf.SetTextColor(34, 139, 34) // Yeşil
			pdf.SetFont(pdfFontFamily, "B", 9)
		} else {
			pdf.SetTextColor(178, 34, 34) // Kırmızı
			pdf.SetFont(pdfFontFamily, "B", 9)
		}
		pdf.CellFormat(35, 10, tr(katilimDurum), "1", 0, "C", fill, 0, "")

		pdf.SetTextColor(50, 50, 50)
		pdf.SetFont(pdfFontFamily, "", 9)
		pdf.CellFormat(40, 10, "", "1", 1, "C", fill, 0, "")
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("pdf çıktısı oluşturulamadı: %w", err)
	}

	return buf.Bytes(), nil
}

// pdfVal boş değer yerine varsayılan (placeholder) döner.
func pdfVal(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

// formatTL bir tutarı Türkçe biçimde (binlik nokta, ondalık virgül) döner. Örn: 12500 -> "12.500,00"
func formatTL(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	parts := strings.SplitN(s, ".", 2)
	intPart := parts[0]
	neg := strings.HasPrefix(intPart, "-")
	if neg {
		intPart = intPart[1:]
	}
	n := len(intPart)
	var b strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 && (n-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteByte(intPart[i])
	}
	res := b.String() + "," + parts[1]
	if neg {
		res = "-" + res
	}
	return res
}

// drawSozlesmeHeader her sayfanın üstüne TTO-FR-366 doküman başlığını ve meta tablosunu çizer.
func drawSozlesmeHeader(pdf *gofpdf.Fpdf) {
	top := 8.0
	logoPath := "frontend/static/images/izu_logo.png"
	pdf.ImageOptions(logoPath, 15, top, 20, 0, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")

	// Orta: doküman başlığı
	pdf.SetFont(pdfFontFamily, "B", 9)
	pdf.SetTextColor(20, 20, 20)
	pdf.SetXY(40, top+1)
	pdf.MultiCell(95, 4.5, "T.C. İSTANBUL SABAHATTİN ZAİM ÜNİVERSİTESİ\nBİLİMSEL ARAŞTIRMA PROJELERİ\nPROJE SÖZLEŞMESİ", "", "C", false)

	// Sağ: meta tablo
	metaX := 140.0
	lw := 26.0
	vw := 29.0
	h := 5.0
	y := top
	rows := [][2]string{
		{"Doküman No", "TTO-FR-366"},
		{"İlk Yayın Tarihi", "15.02.2019"},
		{"Revizyon Tarihi", "17.04.2026"},
		{"Revizyon No", "03"},
		{"Sayfa", fmt.Sprintf("%d / {nb}", pdf.PageNo())},
	}
	pdf.SetDrawColor(120, 120, 120)
	pdf.SetLineWidth(0.2)
	for _, r := range rows {
		pdf.SetXY(metaX, y)
		pdf.SetFont(pdfFontFamily, "B", 7)
		pdf.SetTextColor(40, 40, 40)
		pdf.CellFormat(lw, h, r[0], "1", 0, "L", false, 0, "")
		pdf.SetFont(pdfFontFamily, "", 7)
		pdf.CellFormat(vw, h, r[1], "1", 0, "C", false, 0, "")
		y += h
	}

	// Ayırıcı çizgi
	sepY := 35.0
	pdf.SetDrawColor(20, 20, 20)
	pdf.SetLineWidth(0.4)
	pdf.Line(15, sepY, 195, sepY)

	// İçerik başlangıç noktası
	pdf.SetXY(15, 39)
}

// GenerateSozlesmePDF, BAP Proje Sözleşmesini (TTO-FR-366) resmi formatta PDF olarak üretir.
// Türkçe Yorum: 1.2 (yürütücü bilgileri), 2.1 (proje no/başlığı), 12.1 (bütçe tablosu) ve 16 (yürürlük
// tarihleri + imza) alanları proje ve sözleşme verileriyle otomatik doldurulur. Yürürlük tarihleri,
// indirme anı (baslangic) baz alınarak ve BAP türü süresine göre (bitis) hesaplanmış olarak verilir.
func (s *PdfService) GenerateSozlesmePDF(detail *repository.ProjectDetail, sz *models.ProjeSozlesme, baslangic, bitis time.Time) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 39, 15)
	pdf.SetAutoPageBreak(true, 18)

	const fontDir = "frontend/static/fonts"
	pdf.AddUTF8Font(pdfFontFamily, "", fontDir+"/DejaVuSans.ttf")
	pdf.AddUTF8Font(pdfFontFamily, "B", fontDir+"/DejaVuSans-Bold.ttf")
	pdf.AddUTF8Font(pdfFontFamily, "I", fontDir+"/DejaVuSans.ttf")
	pdf.AddUTF8Font(pdfFontFamily, "BI", fontDir+"/DejaVuSans-Bold.ttf")

	pdf.AliasNbPages("{nb}")
	pdf.SetHeaderFunc(func() { drawSozlesmeHeader(pdf) })
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont(pdfFontFamily, "", 8)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(0, 10, fmt.Sprintf("Sayfa %d / {nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	// Yerel yardımcılar: madde başlığı ve gövde metni
	heading := func(t string) {
		pdf.Ln(1.5)
		pdf.SetFont(pdfFontFamily, "B", 10)
		pdf.SetTextColor(15, 15, 15)
		pdf.MultiCell(0, 5.2, t, "", "L", false)
	}
	body := func(t string) {
		pdf.SetFont(pdfFontFamily, "", 9.5)
		pdf.SetTextColor(45, 45, 45)
		pdf.MultiCell(0, 4.8, t, "", "J", false)
		pdf.Ln(1)
	}

	// ─── Dinamik alanlar ───
	yurutucu := detail.YurutucuAd
	if strings.TrimSpace(yurutucu) == "" || yurutucu == "Bilinmiyor" {
		if strings.TrimSpace(sz.YurutucuAd) != "" {
			yurutucu = sz.YurutucuAd
		} else {
			yurutucu = "PROJE YÜRÜTÜCÜSÜ"
		}
	}
	projeKodu := detail.Proje.ProjeKodu
	if projeKodu == "" {
		projeKodu = fmt.Sprintf("PROJE-%d", detail.Proje.ProjeID)
	}
	projeBaslik := pdfVal(detail.Proje.BaslikTr, "-")
	tc := pdfVal(sz.TCKimlik, "....................")
	adres := pdfVal(sz.YurutucuAdres, "....................")
	tel := pdfVal(sz.YurutucuTelefon, "....................")
	eposta := pdfVal(sz.YurutucuEposta, "....................")
	basStr := baslangic.Format("02.01.2006")
	bitStr := bitis.Format("02.01.2006")

	sureAy := detail.Proje.SureAy
	if sureAy <= 0 {
		// Proje süresi tanımlı değilse indirme ve bitiş tarihinden ay farkını hesapla
		sureAy = int(bitis.Sub(baslangic).Hours()/24/30 + 0.5)
	}

	pdf.AddPage()

	// ─── 1. TARAFLAR ───
	heading("1. TARAFLAR")
	body("1.1. İSTANBUL SABAHATTİN ZAİM ÜNİVERSİTESİ (Sözleşmede İZÜ olarak anılacaktır.)\nAdres: Halkalı Merkez Mah. Halkalı Cad. No:281 Küçükçekmece / İSTANBUL\nTelefon: 0212 692 96 00     E-Posta: bap@izu.edu.tr")
	body(fmt.Sprintf("1.2. %s (Sözleşmede PROJE YÜRÜTÜCÜSÜ olarak anılacaktır.)\nT.C. Kimlik No: %s\nAdres: %s\nTelefon: %s\nE-Posta: %s", yurutucu, tc, adres, tel, eposta))
	body("Her iki taraf, 1.1. ve 1.2. maddelerinde belirtilen adreslerini tebligat adresi olarak kabul etmişlerdir. Adres değişiklikleri usulüne uygun şekilde karşı tarafa tebliğ edilmedikçe, en son belirtilen adreslere yapılacak tebliğ, ilgili tarafa yapılmış sayılır.")

	// ─── 2. SÖZLEŞMENİN KONUSU ───
	heading("2. SÖZLEŞMENİN KONUSU")
	body(fmt.Sprintf("2.1. İşbu Sözleşmenin konusu; taraflarca üzerinde mutabık kalınan ve ekinde yer alan Bilimsel Araştırma Projesi Başvuru Formu'nda kapsam ve içeriği ayrıntılı olarak belirtilen \"%s\" no'lu, \"%s\" başlıklı araştırma projesinin, İZÜ tarafından desteklenmesine ilişkin usul ve esasların belirlenmesidir.", projeKodu, projeBaslik))

	// ─── 3. PROJE YÜRÜTÜCÜSÜNÜN GÖREVLERİ ───
	heading("3. PROJE YÜRÜTÜCÜSÜNÜN GÖREVLERİ")
	body("3.1. Projenin, ekli araştırma projesi başvuru formunda belirtilen program içinde, İZÜ Bilimsel Araştırma Projeleri Yönergesi ve sözleşmedeki süre, amaç ve şartlara uygun olarak yürütülmesi, geliştirilmesi ve sonuçlandırılmasından PROJE YÜRÜTÜCÜSÜ sorumludur. Desteklenmesi kabul edilmiş projenin amaç, kapsam, süre ve program bütçesinde Bilimsel Araştırma Projesi Komisyonunun yazılı izni alınmadan hiçbir değişiklik yapılamaz.")
	body("3.2. PROJE YÜRÜTÜCÜSÜ'nün herhangi bir nedenle görevinden ayrılması veya İZÜ ile ilişiğinin kesilmesi durumunda, proje yürütücülüğü Bilimsel Araştırma Projesi Komisyonu kararıyla uygun görülen bir personele devredilebilir.")
	body("3.3. Bilimsel Araştırma Projeleri kapsamındaki harcamalar İZÜ Bilimsel Araştırma Projesi Yönergesi'ne uygun olarak yapılır.")
	body(fmt.Sprintf("3.4. PROJE YÜRÜTÜCÜSÜ, proje ile ilgili verileri ve bulgularını, yayınladığı her türlü yazı, makale ve sunduğu bildirilerde \"İstanbul Sabahattin Zaim Üniversitesi tarafından desteklenmiştir. (Proje No: %s)\" ibaresini belirtmek zorundadır.", projeKodu))
	body("3.5. İnsanlar ve hayvanlar üzerinde gerçekleştirilecek çalışmalar için zorunlu olan etik kurulu onayının alınması zorunludur ve PROJE YÜRÜTÜCÜSÜ'nün sorumluluğundadır.")
	body("3.6. Proje ekibi, İZÜ bilimsel araştırma projeleri Yönergesi, bilim etiği normları, etik kurulu ve çalışma esaslarına uymakla yükümlüdür.")

	// ─── 4. PROJE YÜRÜTÜCÜSÜNÜN SORUMLULUĞU ───
	heading("4. PROJE YÜRÜTÜCÜSÜNÜN SORUMLULUĞU")
	body("4.1. PROJE YÜRÜTÜCÜSÜ, proje kapsamında yapılan harcamalar, raporlamalar ve bilimsel çıktılar açısından kişisel sorumluluk taşır. Bilerek veya ağır ihmal sonucu sözleşme hükümlerine aykırılık teşkil eden fiiller sebebiyle üniversitenin maddi zarara uğraması hâlinde, yürütücü bu zararı tazmin etmekle yükümlüdür.")

	// ─── 5. ARAÇ, GEREÇ VE DONANIM ───
	heading("5. ARAÇ, GEREÇ VE DONANIM")
	body("5.1. Proje bütçesi gereğince Bilimsel Araştırma Projesi Komisyonu tarafından yurt içinden veya yurt dışından temin edilerek projeye tahsis edilen, sarf malzemesi dışındaki demirbaş niteliğindeki her türlü teçhizat, ilgili Akademik Birim Yönetimi harcama birimi adına kaydedilir. Kaydedilen taşınır, zimmet fişi düzenlenerek PROJE YÜRÜTÜCÜSÜ'nün kullanımına tahsis edilir.")
	body("5.2. Sonuç raporu verilen projelerin makine ve teçhizatı, Bilimsel Araştırma Projesi Komisyonu tarafından gerekli görüldüğü takdirde, daha yaygın yararlanma sağlanması açısından, İZÜ içindeki ilgili bir laboratuvara veya ihtiyaç duyulan başka bir proje yürütücüsüne, bu maddenin ilk fıkra hükmü saklı kalmak kaydı ile verilebilir.")

	// ─── 6. GELİŞME RAPORLARI ───
	heading("6. GELİŞME RAPORLARI")
	body("6.1. PROJE YÜRÜTÜCÜSÜ, projenin devamı süresince her 6 (altı) ayda bir, proje kapsamındaki bilimsel ve mali gelişmeleri içeren ayrıntılı raporları Bilimsel Araştırma Projesi Komisyonu'na sunmakla yükümlüdür. Raporun zamanında sunulmaması veya eksik/veri içermeyen rapor sunulması durumunda, proje ödemeleri durdurulur.")

	// ─── 7. KESİN RAPOR ───
	heading("7. KESİN RAPOR")
	body("7.1. PROJE YÜRÜTÜCÜSÜ, sözleşmede belirtilen proje bitim tarihini izleyen 2 (iki) ay içinde araştırma sonuçlarını içeren kesin raporu sunmakla yükümlüdür.")

	// ─── 8. GÜVENLİK ÖNLEMLERİ ───
	heading("8. GÜVENLİK ÖNLEMLERİ")
	body("8.1. PROJE YÜRÜTÜCÜSÜ proje yerinde kazaları önlemeden ve sağlık şartları bakımından gerekli her türlü güvenlik önlemlerinin alınmasından sorumludur.")

	// ─── 9. KİŞİSEL VERİLERİN KORUNMASI VE GİZLİLİK ───
	heading("9. KİŞİSEL VERİLERİN KORUNMASI VE GİZLİLİK")
	body("9.1. PROJE YÜRÜTÜCÜSÜ, işbu sözleşme kapsamında İZÜ'ye ait öğrendiği tüm bilgileri gizli tutacağını, saklayacağını ve koruyacağını; tüm bilgileri doğrudan ya da dolaylı olarak aralarındaki ilişki amacı dışında kullanmayacağını, İZÜ'nün rızası olmadan üçüncü kişiler ile paylaşmayacağını beyan ve taahhüt eder.")
	body("9.2. PROJE YÜRÜTÜCÜSÜ işbu sözleşme kapsamında öğrendiği tüm kişisel bilgileri 6698 sayılı Kişisel Verilerin Korunması Kanunu ve ilgili mevzuat uyarınca korumak amacıyla gerekli tüm teknik ve idari tedbirleri alacak ve kişisel verilerin korunması hususunda yazılı taahhütname imzalatır.")

	// ─── 10. PATENT HAKLARI ───
	heading("10. PATENT VE FİKRİ MÜLKİYET HAKLARI")
	body("10.1. İZÜ çalışanları tarafından, İZÜ bünyesinde yürütülen bilimsel araştırma ve çalışmalar sonucunda ya da çalışanların üniversitede edindikleri bilgi, deneyim ve birikimlere dayanarak veya üniversitenin altyapı, ekipman, araç ve gereçlerinden yararlanmak suretiyle geliştirilen buluşlara ilişkin tüm fikri ve sınai mülkiyet hakları münhasıran İZÜ'ye aittir. Ancak, bir buluştan gelir elde edilmesi hâlinde, elde edilen gelirin en az üçte biri (1/3'ü) buluşu gerçekleştiren kişiye ödenir.")
	body("10.2. İZÜ tarafından desteklenen projeler neticesinde ortaya çıkan bilimsel sonuçlara ilişkin telif hakları İZÜ'ye aittir. Ancak bilimsel yayın, kitap ve benzeri eserlerin telif hakları, Üniversite Yönetim Kurulu kararı ile kısmen veya tamamen eser sahibine devredilebilir.")

	// ─── 11. HARCAMALARIN DENETİMİ ───
	heading("11. HARCAMALARIN DENETİMİ")
	body("11.1. Proje kapsamında yapılan tüm harcamalar, İZÜ'nün iç denetim birimleri tarafından denetlenebilir. Proje yürütücüsü, istenildiği takdirde harcamalara ilişkin tüm fatura, ödeme belgesi ve diğer destekleyici evrakları ibraz etmekle yükümlüdür.")

	// ─── 12. DESTEK MİKTARI VE BÜTÇE ───
	heading("12. DESTEK MİKTARI")
	body("12.1. Bu sözleşme ekinde yer alan destek kalemlerine ilişkin tutarların tamamı, sözleşme imzalanmadan önce İZÜ Bilimsel Araştırma Projeleri Komisyonu tarafından onaylanmış ve sözleşme ekinde belirtilmiş olacaktır. Sözleşmenin geçerlilik kazanabilmesi için tüm bütçe kalemlerinin net, imzalı ve tarihli şekilde belirtilmesi zorunludur.")

	// 12.1 Bütçe tablosu (7 sabit kalem + toplam)
	addSozlesmeBudgetTable(pdf, detail)

	// ─── 13. CEZAİ SORUMLULUKLAR ───
	heading("13. CEZAİ SORUMLULUKLAR")
	body("13.1. Projenin; gelişme veya kesin raporlarını zamanında sunmaması, proje amacına aykırı faaliyetlerde bulunulması, etik ilkelere veya sözleşme hükümlerine aykırı davranışta bulunulması hâlinde, proje BAP Komisyonu kararıyla durdurulabilir veya iptal edilebilir. Bu durumda yürütücüye yapılan harcamalar, güncel rayiç bedeller üzerinden üniversiteye iade ettirilir.")
	body("13.2. Proje yürütülmekte iken proje çalışmalarında bilimsel etiğe aykırılık saptandığında, Bilimsel Araştırma Projesi Komisyonu kararı ile iptal edilir. Bu suretle projenin iptaline yol açan kişi veya kişiler 3 (üç) yıl süreyle proje desteğinden yararlanamaz.")
	body("13.3. Projede onaylanan bütçenin, üniversitenin tabi olduğu \"Vakıf Yükseköğretim Kurumları İhale Yönetmeliği\"ne uygun harcanması önem arz etmektedir. Yürütücü; hizmet, sarf, cihaz vb. alımlarında üniversite satın alma birimi üzerinden yönetmeliğe uygun alım yapmak zorundadır. Yönetmeliğe uygun olmayan ve sözleşme bütçesini aşan alımlar yürütücünün sorumluluğundadır.")

	// ─── 14. RAPOR TESLİM TARİHLERİ ───
	heading("14. RAPOR TESLİM TARİHLERİ")
	body("14.1. Gelişme ve sonuç raporları, işbu sözleşmede belirtilen tarihlerde İZÜ Teknoloji Transfer Ofisi Koordinatörlüğü'ne iletilecektir.")

	// ─── 15. ANLAŞMAZLIKLARIN ÇÖZÜMÜ ───
	heading("15. ANLAŞMAZLIKLARIN ÇÖZÜMÜ")
	body("15.1. İşbu sözleşmenin uygulanmasından doğabilecek her türlü anlaşmazlığın çözümünde İstanbul (Küçükçekmece) Mahkemeleri ve İcra Daireleri yetkilidir.")

	// ─── 16. YÜRÜRLÜK ───
	heading("16. YÜRÜRLÜK")
	body(fmt.Sprintf("16.1. İşbu Sözleşme 16 (on altı) maddeden ve aşağıdaki yürürlük tablosundan oluşmakta olup, taraflarca imzalandığı %s tarihinde yürürlüğe girer. Proje süresi %d (ay) olup sözleşme %s tarihinde sona erer.", basStr, sureAy, bitStr))

	addSozlesmeTarihTablosu(pdf, basStr, bitStr, sureAy)
	addSozlesmeImzaBlok(pdf, yurutucu)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("sözleşme pdf çıktısı oluşturulamadı: %w", err)
	}
	return buf.Bytes(), nil
}

// addSozlesmeBudgetTable, 12.1 maddesinin bütçe destek tablosunu projenin bütçe kalemlerinden
// 7 sabit kalem halinde kategorize ederek çizer ve toplamı hesaplar.
func addSozlesmeBudgetTable(pdf *gofpdf.Fpdf, detail *repository.ProjectDetail) {
	type katSatir struct {
		etiket string
		adlar  []string
		tutar  float64
	}
	katlar := []katSatir{
		{"1. Sarf Malzeme", []string{"sarf malzeme"}, 0},
		{"2. Seyahat", []string{"seyahat (yolluk)", "seyahat", "yolluk"}, 0},
		{"3. Hizmet Alımı", []string{"hizmet alımı", "hizmet alimi"}, 0},
		{"4. Makine/Teçhizat", []string{"makine-teçhizat", "makine/teçhizat", "makine teçhizat"}, 0},
		{"5. Bursiyer", []string{"bursiyer"}, 0},
		{"6. Basılı-Yayın Alımı", []string{"yayın/basım", "basılı-yayın alımı", "yayın", "basım"}, 0},
		{"7. Yazılım Alımı", []string{"yazılım", "yazılım alımı"}, 0},
	}
	var toplam float64
	for _, b := range detail.Butceler {
		adi := strings.ToLower(strings.TrimSpace(b.KategoriAdi))
		tut := b.ToplamFiyat
		if tut == 0 {
			tut = float64(b.BirimOzelligi) * b.BirimFiyat
		}
		for i := range katlar {
			matched := false
			for _, a := range katlar[i].adlar {
				if adi == a {
					katlar[i].tutar += tut
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		toplam += tut
	}

	pdf.Ln(1)
	// Başlık satırı
	pdf.SetFont(pdfFontFamily, "B", 9)
	pdf.SetFillColor(38, 74, 150)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(120, 8, "  BÜTÇE DESTEK ADI", "1", 0, "L", true, 0, "")
	pdf.CellFormat(60, 8, "TUTAR (TL)  ", "1", 1, "R", true, 0, "")

	pdf.SetFont(pdfFontFamily, "", 9)
	pdf.SetTextColor(40, 40, 40)
	for i, k := range katlar {
		fill := i%2 == 0
		pdf.SetFillColor(245, 247, 250)
		tutStr := "-"
		if k.tutar > 0 {
			tutStr = formatTL(k.tutar)
		}
		pdf.CellFormat(120, 7, "  "+k.etiket, "1", 0, "L", fill, 0, "")
		pdf.CellFormat(60, 7, tutStr+"  ", "1", 1, "R", fill, 0, "")
	}
	// Toplam satırı
	pdf.SetFont(pdfFontFamily, "B", 9)
	pdf.SetFillColor(224, 231, 245)
	pdf.CellFormat(120, 8, "  TOPLAM", "1", 0, "R", true, 0, "")
	pdf.CellFormat(60, 8, formatTL(toplam)+" TL  ", "1", 1, "R", true, 0, "")
	pdf.SetTextColor(45, 45, 45)
}

// addSozlesmeTarihTablosu, 16. maddedeki yürürlük tarihleri tablosunu çizer.
// Türkçe Yorum: Başlangıç, indirme tarihi; bitiş ise indirme tarihine BAP türü süresi eklenerek hesaplanmıştır.
func addSozlesmeTarihTablosu(pdf *gofpdf.Fpdf, basStr, bitStr string, sureAy int) {
	pdf.Ln(2)
	rows := [][2]string{
		{"Sözleşme / Proje Başlangıç Tarihi", basStr},
		{"Proje Süresi", fmt.Sprintf("%d Ay", sureAy)},
		{"Sözleşme / Proje Bitiş Tarihi", bitStr},
	}
	for _, r := range rows {
		pdf.SetFont(pdfFontFamily, "B", 9)
		pdf.SetFillColor(245, 247, 250)
		pdf.SetTextColor(40, 40, 40)
		pdf.CellFormat(90, 7, "  "+r[0], "1", 0, "L", true, 0, "")
		pdf.SetFont(pdfFontFamily, "", 9)
		pdf.CellFormat(90, 7, "  "+r[1], "1", 1, "L", false, 0, "")
	}
}

// addSozlesmeImzaBlok, sözleşme sonundaki taraf imza alanlarını çizer.
// Türkçe Yorum: "Proje Yürütücüsü" başlığının altına yürütücünün adı soyadı otomatik yazdırılır.
func addSozlesmeImzaBlok(pdf *gofpdf.Fpdf, yurutucu string) {
	// İmza bloğu için yeterli alan yoksa yeni sayfa
	if pdf.GetY() > 235 {
		pdf.AddPage()
	}
	pdf.Ln(12)
	y := pdf.GetY()

	// Sol: İZÜ
	pdf.SetFont(pdfFontFamily, "B", 9)
	pdf.SetTextColor(20, 20, 20)
	pdf.SetXY(20, y)
	pdf.CellFormat(80, 6, "İSTANBUL SABAHATTİN ZAİM ÜNİVERSİTESİ", "", 2, "C", false, 0, "")
	pdf.Ln(14)
	pdf.SetX(20)
	pdf.SetFont(pdfFontFamily, "", 9)
	pdf.CellFormat(80, 5, "Yetkili İmza", "T", 2, "C", false, 0, "")

	// Sağ: Proje Yürütücüsü
	pdf.SetFont(pdfFontFamily, "B", 9)
	pdf.SetTextColor(20, 20, 20)
	pdf.SetXY(110, y)
	pdf.CellFormat(80, 6, "PROJE YÜRÜTÜCÜSÜ", "", 2, "C", false, 0, "")
	pdf.SetXY(110, y+7)
	pdf.SetFont(pdfFontFamily, "B", 9)
	pdf.SetTextColor(38, 74, 150)
	pdf.CellFormat(80, 6, yurutucu, "", 2, "C", false, 0, "")
	pdf.SetXY(110, y+20)
	pdf.SetFont(pdfFontFamily, "", 9)
	pdf.SetTextColor(20, 20, 20)
	pdf.CellFormat(80, 5, "İmza & Tarih", "T", 2, "C", false, 0, "")
}
