package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"bap_ai/backend/repository"
	"bap_ai/backend/service"

	"github.com/gin-gonic/gin"
)

// PdfHandler yapısı PDF ile ilgili HTTP isteklerini karşılar.
type PdfHandler struct {
	PdfService *service.PdfService
	ProjeRepo  *repository.ProjeRepository
}

// NewPdfHandler yeni bir PdfHandler oluşturur.
func NewPdfHandler(pdfService *service.PdfService, projeRepo *repository.ProjeRepository) *PdfHandler {
	return &PdfHandler{
		PdfService: pdfService,
		ProjeRepo:  projeRepo,
	}
}

// GeneratePDF projenin PDF önizlemesini oluşturur ve döner.
// GET /api/proje/:id/pdf
func (h *PdfHandler) GeneratePDF(c *gin.Context) {
	// Proje ID'sini URL parametresinden al
	idStr := c.Param("id")
	projeID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	// PDF oluştur
	pdfBytes, err := h.PdfService.GenerateProjectPDF(projeID)
	if err != nil {
		log.Printf("PDF oluşturma hatası (proje_id=%d): %v", projeID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF oluşturulamadı"})
		return
	}

	// PDF dosyasını yanıt olarak döndür
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=proje_%d_basvuru.pdf", projeID))
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// FinalizePDF projeyi onaylayıp PDF dosyasını sunucuya kaydeder.
// POST /api/proje/:id/finalize
func (h *PdfHandler) FinalizePDF(c *gin.Context) {
	// Proje ID'sini URL parametresinden al
	idStr := c.Param("id")
	projeID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	// Kullanıcı doğrulama: sadece proje sahibi onaylayabilir
	uyeIDFloat, exists := c.Get("uye_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bilgisi bulunamadı"})
		return
	}
	_ = int(uyeIDFloat.(float64)) // Üye ID (ileride yetki kontrolü için)

	// 1. PDF oluştur
	pdfBytes, err := h.PdfService.GenerateProjectPDF(projeID)
	if err != nil {
		log.Printf("Finalize PDF oluşturma hatası (proje_id=%d): %v", projeID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF oluşturulamadı"})
		return
	}

	// 2. PDF dosyasını sunucuya kaydet
	pdfDir := "uploads/pdf"
	if err := os.MkdirAll(pdfDir, 0755); err != nil {
		log.Printf("PDF dizini oluşturulamadı: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Dosya sistemi hatası"})
		return
	}

	pdfFileName := fmt.Sprintf("proje_%d_basvuru.pdf", projeID)
	pdfPath := fmt.Sprintf("%s/%s", pdfDir, pdfFileName)

	if err := os.WriteFile(pdfPath, pdfBytes, 0644); err != nil {
		log.Printf("PDF dosyası yazılamadı: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF kaydedilemedi"})
		return
	}

	// 3. PDF yolunu veritabanına kaydet
	if err := h.ProjeRepo.SavePDFPath(projeID, pdfPath); err != nil {
		log.Printf("PDF yolu DB'ye kaydedilemedi: %v", err)
		// PDF oluştu ama DB'ye yazamadık — yine de devam
	}

	// 4. Proje durumunu TTO ön incelemesi için "incelemede" olarak güncelle ve log yaz
	var baslangicDurum string = "taslak"
	if currentProje, err := h.ProjeRepo.GetProjeByID(projeID); err == nil && currentProje != nil {
		baslangicDurum = currentProje.DurumAdi
	}
	islemYapanID := int(uyeIDFloat.(float64))

	// Türkçe Yorum: Akademisyen başvurusunu kesinleştirdiğinde proje durumunu "incelemede" (TTO ön inceleme) olarak güncelliyoruz.
	if err := h.ProjeRepo.UpdateProjectStatusWithLog(projeID, islemYapanID, baslangicDurum, "incelemede", "Başvuru akademisyen tarafından tamamlandı ve TTO ön incelemesine sunuldu."); err != nil {
		log.Printf("Proje durumu güncellenemedi: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proje durumu güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Başvuru başarıyla onaylandı ve PDF oluşturuldu",
		"pdf_path": pdfPath,
	})
}
