package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
)

// FeedbackService geri bildirim iş mantığını yürütür.
// Türkçe Yorum: Geri bildirim mesajını veritabanına işleyip admin kullanıcılarına e-posta gönderen servis katmanı.
type FeedbackService struct {
	FeedbackRepo *repository.FeedbackRepository
	EpostaService *EpostaService
}

// NewFeedbackService yeni bir FeedbackService örneği oluşturur.
// Türkçe Yorum: FeedbackService yapısını ilklendiren kurucu fonksiyon.
func NewFeedbackService(feedbackRepo *repository.FeedbackRepository, epostaService *EpostaService) *FeedbackService {
	return &FeedbackService{
		FeedbackRepo:  feedbackRepo,
		EpostaService: epostaService,
	}
}

// SendFeedback geri bildirim kaydını oluşturur ve admin kullanıcılarına imza detaylarıyla e-posta gönderir.
// Türkçe Yorum: Kullanıcının mesajını kaydeder, gönderenin yetki/profil bilgilerini derleyip admin grubuna mail iletir.
func (s *FeedbackService) SendFeedback(uyeID int, req *models.CreateFeedbackRequest) error {
	if strings.TrimSpace(req.Mesaj) == "" {
		return fmt.Errorf("geri bildirim mesajı boş olamaz")
	}

	if strings.TrimSpace(req.Konu) == "" {
		req.Konu = "Genel Geri Bildirim"
	}

	// 1. Veritabanına kaydet
	fb := &models.Feedback{
		UyeID:    uyeID,
		Konu:     strings.TrimSpace(req.Konu),
		Mesaj:    strings.TrimSpace(req.Mesaj),
		SayfaURL: strings.TrimSpace(req.SayfaURL),
	}

	if err := s.FeedbackRepo.CreateFeedback(fb); err != nil {
		log.Printf("[FEEDBACK] Veritabanı kayıt hatası: %v", err)
		return err
	}

	// 2. Gönderenin detaylı bilgilerini (ad, unvan, bolum, eposta, tel, yetkiler/roller) al
	senderInfo, err := s.FeedbackRepo.GetSenderFullDetails(uyeID)
	if err != nil {
		log.Printf("[FEEDBACK] Gönderen bilgileri alınırken uyarı: %v", err)
		senderInfo = &models.FeedbackSenderInfo{
			UyeID: uyeID,
		}
	}

	// 3. Admin e-postalarını getir
	adminEmails, err := s.FeedbackRepo.GetAdminEmails()
	if err != nil || len(adminEmails) == 0 {
		log.Printf("[FEEDBACK] Admin e-postası bulunamadı veya sorgu hatası: %v", err)
		// E-posta gönderilecek admin bulunamasa bile DB'ye kaydedildiği için işlemi tamamla
		return nil
	}

	// 4. HTML E-Posta Şablonunu ve Gönderen İmza Kartını oluştur
	roleText := "Belirtilmemiş"
	if len(senderInfo.RoleLabels) > 0 {
		roleText = strings.Join(senderInfo.RoleLabels, ", ")
	} else if len(senderInfo.Roles) > 0 {
		roleText = strings.Join(senderInfo.Roles, ", ")
	}

	unvanAdSoyad := strings.TrimSpace(fmt.Sprintf("%s %s %s", senderInfo.Unvan, senderInfo.Ad, senderInfo.Soyad))
	if unvanAdSoyad == "" {
		unvanAdSoyad = fmt.Sprintf("Kullanıcı ID: %d", uyeID)
	}

	bolumText := senderInfo.Bolum
	if bolumText == "" {
		bolumText = "Belirtilmemiş"
	}

	telefonText := senderInfo.Telefon
	if telefonText == "" {
		telefonText = "Belirtilmemiş"
	}

	epostaText := senderInfo.Eposta
	if epostaText == "" {
		epostaText = "Belirtilmemiş"
	}

	tarihStr := time.Now().Format("02.01.2006 15:04")

	sayfaURLText := req.SayfaURL
	if sayfaURLText == "" {
		sayfaURLText = "Belirtilmemiş"
	}

	subject := fmt.Sprintf("[BAP Geri Bildirim] %s - %s", req.Konu, unvanAdSoyad)

	bodyHTML := fmt.Sprintf(`
<div style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; max-width: 650px; margin: 0 auto; border: 1px solid #e2e8f0; border-radius: 12px; overflow: hidden; background: #ffffff; box-shadow: 0 4px 12px rgba(0,0,0,0.05);">
    <div style="background: linear-gradient(135deg, #1e3a8a 0%, #2563eb 100%); color: #ffffff; padding: 24px; text-align: center;">
        <h2 style="margin: 0; font-size: 20px; font-weight: 700; letter-spacing: 0.5px;">BAP Sistemi - Yeni Geri Bildirim</h2>
        <p style="margin: 6px 0 0 0; font-size: 13px; opacity: 0.9;">Kullanıcı Geri Bildirim Bildirimi</p>
    </div>
    
    <div style="padding: 28px; color: #334155; line-height: 1.6;">
        <table style="width: 100%%; margin-bottom: 20px; border-collapse: collapse;">
            <tr>
                <td style="padding: 6px 0; font-weight: bold; width: 150px; color: #475569;">Geri Bildirim Konusu:</td>
                <td style="padding: 6px 0; color: #0f172a; font-weight: 600;">%s</td>
            </tr>
            <tr>
                <td style="padding: 6px 0; font-weight: bold; color: #475569;">Gönderildiği Sayfa:</td>
                <td style="padding: 6px 0; color: #2563eb;"><a href="%s" style="color: #2563eb; text-decoration: none;">%s</a></td>
            </tr>
        </table>

        <div style="margin-bottom: 24px;">
            <p style="font-weight: bold; margin-bottom: 8px; color: #1e293b;">Mesaj İçeriği:</p>
            <div style="background: #f8fafc; border-left: 4px solid #3b82f6; border-radius: 6px; padding: 18px; color: #1e293b; font-size: 14px; white-space: pre-wrap; word-break: break-word;">%s</div>
        </div>

        <hr style="border: none; border-top: 1px dashed #cbd5e1; margin: 30px 0;" />

        <!-- GÖNDEREN İMZA KARTI (EMAIL SIGNATURE) -->
        <div style="background: #f1f5f9; border: 1px solid #e2e8f0; border-radius: 10px; padding: 20px; font-size: 13px; color: #334155;">
            <div style="margin-bottom: 12px; padding-bottom: 10px; border-bottom: 1px solid #cbd5e1;">
                <div style="font-size: 14px; font-weight: 700; color: #1e3a8a;">📋 Gönderen Profil & Yetki Bilgileri (İmza Kartı)</div>
            </div>
            <table style="width: 100%%; border-collapse: collapse;">
                <tr>
                    <td style="padding: 5px 0; font-weight: bold; width: 160px; color: #64748b;">Ad Soyad / Unvan:</td>
                    <td style="padding: 5px 0; font-weight: 600; color: #0f172a;">%s</td>
                </tr>
                <tr>
                    <td style="padding: 5px 0; font-weight: bold; color: #64748b;">Bölüm / Birim:</td>
                    <td style="padding: 5px 0; color: #334155;">%s</td>
                </tr>
                <tr>
                    <td style="padding: 5px 0; font-weight: bold; color: #64748b;">E-posta Adresi:</td>
                    <td style="padding: 5px 0; color: #2563eb;"><a href="mailto:%s" style="color: #2563eb;">%s</a></td>
                </tr>
                <tr>
                    <td style="padding: 5px 0; font-weight: bold; color: #64748b;">Telefon:</td>
                    <td style="padding: 5px 0; color: #334155;">%s</td>
                </tr>
                <tr>
                    <td style="padding: 5px 0; font-weight: bold; color: #64748b;">Sistem Rolleri / Yetkileri:</td>
                    <td style="padding: 5px 0;">
                        <span style="background: #dbeafe; color: #1e40af; padding: 3px 10px; border-radius: 12px; font-weight: 600; font-size: 12px;">%s</span>
                    </td>
                </tr>
                <tr>
                    <td style="padding: 5px 0; font-weight: bold; color: #64748b;">Gönderim Tarihi:</td>
                    <td style="padding: 5px 0; color: #64748b;">%s</td>
                </tr>
            </table>
        </div>
    </div>
    
    <div style="background: #f8fafc; padding: 14px; text-align: center; font-size: 12px; color: #94a3b8; border-top: 1px solid #e2e8f0;">
        Bu e-posta İZÜ BAP Otomasyonu Geri Bildirim Modülü tarafından otomatik üretilmiştir.
    </div>
</div>
`,
		req.Konu,
		sayfaURLText, sayfaURLText,
		req.Mesaj,
		unvanAdSoyad,
		bolumText,
		epostaText, epostaText,
		telefonText,
		roleText,
		tarihStr,
	)

	// 5. E-postayı tüm Admin kullanıcılarına ilet
	if err := s.EpostaService.SendEmailSMTP(adminEmails, subject, bodyHTML); err != nil {
		log.Printf("[FEEDBACK] E-posta gönderimi uyarısı: %v", err)
	}

	return nil
}
