package service

import (
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"bap_ai/configs"
)

// EpostaService, e-posta bildirimlerinin SMTP üzerinden gönderilmesinden sorumludur.
// Türkçe Yorum: Bu servis veritabanı bağlantısı (DB) ve genel uygulama ayarlarını (Config) barındırır.
type EpostaService struct {
	DB     *sql.DB
	Config *configs.Config
}

// NewEpostaService yeni bir EpostaService nesnesi oluşturur.
// Türkçe Yorum: EpostaService yapısını başlatmak için kurucu fonksiyon.
func NewEpostaService(db *sql.DB, config *configs.Config) *EpostaService {
	return &EpostaService{
		DB:     db,
		Config: config,
	}
}

// SendEmailSMTP belirtilen alıcılara SMTP protokolü üzerinden e-posta gönderir.
// Türkçe Yorum: SMTP bağlantısını kurup TLS veya STARTTLS durumunu yöneterek base64 formatında HTML e-posta iletir.
func (s *EpostaService) SendEmailSMTP(to []string, subject string, body string) error {
	// Türkçe Yorum: Gönderilen e-postanın bir kopyası sistem içi bildirim olarak veritabanına yazılır.
	s.saveNotificationDB(to, subject, body)

	// E-posta gönderimi pasif ise konsola mock log basılır ve işlem başarılı kabul edilir.
	if !s.Config.SMTPEnabled {
		log.Printf("[E-POSTA MOCK - AKTİF DEĞİL] Alıcılar: %v\nKonu: %s\nİçerik: %s\n", to, subject, body)
		return nil
	}

	host := s.Config.SMTPHost
	port := s.Config.SMTPPort
	user := s.Config.SMTPUser
	pass := s.Config.SMTPPassword
	from := s.Config.SMTPFrom

	if host == "" || port == "" || from == "" {
		return fmt.Errorf("SMTP ayarları eksik. Lütfen .env dosyasını kontrol edin")
	}

	// MIME header bilgilerini Türkçe karakter desteğiyle yapılandır
	header := make(map[string]string)
	header["From"] = from
	header["To"] = strings.Join(to, ",")
	header["Subject"] = "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(subject)) + "?="
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = `text/html; charset="utf-8"`
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	addr := host + ":" + port
	auth := smtp.PlainAuth("", user, pass, host)

	// Port 465 ise doğrudan SSL/TLS bağlantısı kurulur.
	if port == "465" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         host,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS bağlantısı kurulamadı: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("SMTP istemcisi oluşturulamadı: %w", err)
		}
		defer client.Close()

		if pass != "" {
			if err = client.Auth(auth); err != nil {
				return fmt.Errorf("SMTP kimlik doğrulama hatası: %w", err)
			}
		}

		if err = client.Mail(from); err != nil {
			return fmt.Errorf("SMTP MAIL komutu hatası: %w", err)
		}

		for _, addrTo := range to {
			if err = client.Rcpt(addrTo); err != nil {
				return fmt.Errorf("SMTP RCPT komutu hatası: %w", err)
			}
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("SMTP DATA komutu hatası: %w", err)
		}

		_, err = w.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("E-posta gövdesi yazılamadı: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("SMTP veri akışı kapatılamadı: %w", err)
		}

		return client.Quit()
	}

	// Port 587 veya 25 için standart SendMail kullanılır (içinde STARTTLS barındırır).
	var smtpAuth smtp.Auth
	if pass != "" {
		smtpAuth = auth
	}
	err := smtp.SendMail(addr, smtpAuth, from, to, []byte(message))
	if err != nil {
		return fmt.Errorf("SMTP e-posta gönderimi başarısız oldu: %w", err)
	}

	return nil
}

// FormatEmailTemplate modern ve şık bir HTML şablonu oluşturur.
// Türkçe Yorum: E-postaların kurumsal ve şık görünmesi için inline CSS içeren bir şablon doldurulur.
func (s *EpostaService) FormatEmailTemplate(greeting, message, projectCode, projectTitle, coordinator, newStatus, description string) string {
	descHTML := ""
	if description != "" {
		descHTML = fmt.Sprintf(`<tr><td class="label">Açıklama:</td><td>%s</td></tr>`, description)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  body { font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; background-color: #f4f6f9; margin: 0; padding: 20px; }
  .card { max-width: 600px; margin: 0 auto; background: #ffffff; border-radius: 8px; border: 1px solid #e1e8ed; overflow: hidden; box-shadow: 0 4px 6px rgba(0,0,0,0.05); }
  .header { background: linear-gradient(135deg, #1e3c72 0%%, #2a5298 100%%); color: #ffffff; padding: 25px 20px; text-align: center; }
  .header h1 { margin: 0; font-size: 22px; font-weight: 600; }
  .content { padding: 30px 20px; color: #333333; line-height: 1.6; }
  .project-info { background: #f8f9fa; border-left: 4px solid #1e3c72; padding: 15px; margin: 20px 0; border-radius: 0 4px 4px 0; }
  .project-info table { width: 100%%; border-collapse: collapse; }
  .project-info td { padding: 5px 0; vertical-align: top; font-size: 14px; }
  .project-info td.label { font-weight: bold; width: 120px; color: #555555; }
  .footer { background: #f4f6f9; text-align: center; padding: 15px; font-size: 12px; color: #777777; border-top: 1px solid #e1e8ed; }
</style>
</head>
<body>
  <div class="card">
    <div class="header">
      <h1>BAP Otomasyon Sistemi Bildirimi</h1>
    </div>
    <div class="content">
      <p>%s</p>
      <p>%s</p>
      <div class="project-info">
        <table>
          <tr><td class="label">Proje Kodu:</td><td>%s</td></tr>
          <tr><td class="label">Proje Başlığı:</td><td>%s</td></tr>
          <tr><td class="label">Yürütücü:</td><td>%s</td></tr>
          <tr><td class="label">Yeni Durum:</td><td><strong>%s</strong></td></tr>
          %s
        </table>
      </div>
      <p>Detayları incelemek ve işlem yapmak için BAP otomasyon sistemine giriş yapabilirsiniz.</p>
    </div>
    <div class="footer">
      Bu e-posta BAP Otomasyon Sistemi tarafından otomatik olarak üretilmiştir. Lütfen doğrudan yanıtlamayınız.
    </div>
  </div>
</body>
</html>`, greeting, message, projectCode, projectTitle, coordinator, newStatus, descHTML)
}

// GetStatusLabel durum anahtarını Türkçe etiket ismine dönüştürür.
// Türkçe Yorum: Veritabanındaki teknik durum isimlerini kullanıcı dostu Türkçe karşılıklarına çevirir.
func (s *EpostaService) GetStatusLabel(status string) string {
	switch status {
	case "taslak":
		return "Taslak"
	case "incelemede":
		return "TTO Ön İnceleme"
	case "dekan_onayi_bekliyor":
		return "Dekan Onayı Bekliyor"
	case "dekan_onayladi":
		return "Dekan Onayladı"
	case "komisyon_bekliyor":
		return "Komisyon Onayı Bekliyor"
	case "komisyon_onayladi":
		return "Komisyon Onayladı"
	case "hakem_atama_bekliyor":
		return "Hakem Ataması Bekleniyor"
	case "hakem_bekliyor":
		return "Hakem Değerlendirmesinde"
	case "hakem_onayladi":
		return "Hakem Onayladı"
	case "sozlesme_imza":
		return "Sözleşme / İmza Aşaması"
	case "tto_aktif":
		return "TTO Onayı Bekliyor"
	case "onaylandi":
		return "Onaylandı"
	case "reddedildi":
		return "Reddedildi"
	case "tamamlandi":
		return "Tamamlandı"
	case "revizyon":
		return "Revizyon Talebi"
	case "yururlukte":
		return "Yürürlükte (Aktif)"
	default:
		return status
	}
}

// SendStatusNotificationEmail projenin durum değişikliklerinde ilgili muhataplara e-posta gönderir.
// Türkçe Yorum: Bu fonksiyon, proje durum değişikliklerinde (taslak -> incelemede, onay, ret, revizyon vb.) alıcıları tespit eder ve e-postayı tetikler.
func (s *EpostaService) SendStatusNotificationEmail(projeID int, islemYapanID int, baslangicDurum, yeniDurum, aciklama string) {
	// Proje ve Yürütücü bilgilerini çek
	var projeKodu, baslikTr, coordName, coordEposta, coordBolum string
	query := `
		SELECT COALESCE(p.proje_kodu, 'KODSUZ'), COALESCE(p.baslik_tr, 'Başlıksız Proje'),
		       COALESCE(u.unvan || ' ' || u.ad || ' ' || u.soyad, u.ad || ' ' || u.soyad, 'Bilinmiyor'),
		       COALESCE(u.eposta, ''), COALESCE(u.bolum, '')
		FROM proje p
		LEFT JOIN uye u ON p.koordinator_id = u.uye_id
		WHERE p.proje_id = $1
	`
	err := s.DB.QueryRow(query, projeID).Scan(&projeKodu, &baslikTr, &coordName, &coordEposta, &coordBolum)
	if err != nil {
		log.Printf("E-posta gönderimi için proje bilgileri alınamadı (ProjeID: %d): %v", projeID, err)
		return
	}

	var recipients []string
	var greeting, message, subject string

	statusLabel := s.GetStatusLabel(yeniDurum)

	// Duruma göre alıcıları ve e-posta içeriğini belirle
	switch yeniDurum {
	case "incelemede":
		// Türkçe Yorum: Akademisyen projeyi girdiğinde TTO temsilcilerine gider.
		subject = fmt.Sprintf("Yeni Proje Ön İnceleme Talebi - %s", projeKodu)
		greeting = "Sayın TTO Temsilcisi,"
		message = fmt.Sprintf("Sisteme yeni bir proje başvurusu yapılmıştır. Lütfen ön inceleme işlemlerini gerçekleştirmek üzere sisteme giriş yapınız.")
		
		// TTO e-postalarını çek
		recipients = s.getEmailsByRole("tto")

	case "dekan_onayi_bekliyor":
		// Türkçe Yorum: TTO onaylayıp dekan onayına gönderdiğinde ilgili dekan(lar)a gider.
		subject = fmt.Sprintf("Dekan Onayı Bekleyen Proje Başvurusu - %s", projeKodu)
		greeting = "Sayın Dekan,"
		message = fmt.Sprintf("Fakülteniz/Bölümünüz öğretim üyesi tarafından sunulan proje başvurusu TTO ön incelemesinden geçerek onayınıza sunulmuştur.")
		
		// Bölüme göre Dekan e-postasını bul, yoksa genel dekanları al
		recipients = s.getEmailsByRoleAndDepartment("dekan", coordBolum)

	case "dekan_onayladi":
		// Türkçe Yorum: Dekan onayladığında TTO'ya gider.
		subject = fmt.Sprintf("Dekan Onay Bildirimi - %s", projeKodu)
		greeting = "Sayın TTO Temsilcisi,"
		message = fmt.Sprintf("İlgili dekanlık tarafından onaylanan proje başvurusu, komisyona sevk edilmek üzere TTO ekranına düşmüştür.")
		recipients = s.getEmailsByRole("tto")

	case "komisyon_bekliyor":
		// Türkçe Yorum: TTO komisyon onayına gönderdiğinde komisyon üyelerine gider.
		subject = fmt.Sprintf("Komisyon Onayı Bekleyen Proje Başvurusu - %s", projeKodu)
		greeting = "Sayın BAP Komisyon Üyesi,"
		message = fmt.Sprintf("Değerlendirmeniz için komisyon onayına sunulan yeni bir proje başvurusu bulunmaktadır.")
		recipients = s.getEmailsByRole("komisyon")

	case "komisyon_onayladi":
		// Türkçe Yorum: Komisyon onayladığında TTO'ya gider.
		subject = fmt.Sprintf("Komisyon Onay Bildirimi - %s", projeKodu)
		greeting = "Sayın TTO Temsilcisi,"
		message = fmt.Sprintf("Komisyon değerlendirmesi başarıyla tamamlanan proje, hakem atama veya doğrudan sözleşme işlemleri için sevk edilmiştir.")
		recipients = s.getEmailsByRole("tto")

	case "hakem_bekliyor":
		// Türkçe Yorum: Hakem atandığında veya hakem daveti gönderildiğinde atanan hakemlere gider.
		subject = fmt.Sprintf("BAP Projesi Hakem Değerlendirme Talebi - %s", projeKodu)
		greeting = "Sayın Hakem,"
		message = fmt.Sprintf("BAP Otomasyon Sistemi üzerinden değerlendirmeniz için tarafınıza bir proje atanmıştır. Detayları incelemek ve değerlendirme formunu doldurmak üzere lütfen sisteme giriş yapınız.")
		
		// Atanan ve henüz karar vermemiş hakemleri çek
		recipients = s.getAssignedHakems(projeID)

	case "hakem_onayladi":
		// Türkçe Yorum: Tüm hakem değerlendirmeleri bittiğinde TTO'ya gider.
		subject = fmt.Sprintf("Hakem Değerlendirmeleri Tamamlandı - %s", projeKodu)
		greeting = "Sayın TTO Temsilcisi,"
		message = fmt.Sprintf("Projeye ait hakem değerlendirmeleri başarıyla tamamlanmıştır. Projeyi sözleşme aşamasına taşımak üzere onayınızı beklemektedir.")
		recipients = s.getEmailsByRole("tto")

	case "sozlesme_imza":
		// Türkçe Yorum: Sözleşme imza aşamasına geldiğinde proje koordinatörüne gider.
		subject = fmt.Sprintf("Proje Sözleşme İmza Aşaması - %s", projeKodu)
		greeting = fmt.Sprintf("Sayın %s,", coordName)
		message = fmt.Sprintf("Başvurmuş olduğunuz '%s' başlıklı projeniz tüm onay aşamalarından geçerek Sözleşme İmza aşamasına gelmiştir. Lütfen e-imza işlemlerinizi tamamlamak üzere sisteme giriş yapınız.", baslikTr)
		if coordEposta != "" {
			recipients = append(recipients, coordEposta)
		}

	case "yururlukte":
		// Türkçe Yorum: Proje yürürlüğe girdiğinde koordinatöre gider.
		subject = fmt.Sprintf("Projeniz Yürürlüğe Alınmıştır - %s", projeKodu)
		greeting = fmt.Sprintf("Sayın %s,", coordName)
		message = fmt.Sprintf("Tebrikler! '%s' başlıklı BAP projeniz onaylanmış ve yürürlüğe (aktif duruma) alınmıştır. Bütçe harcamalarınızı ve satın alma taleplerinizi sistem üzerinden başlatabilirsiniz.", baslikTr)
		if coordEposta != "" {
			recipients = append(recipients, coordEposta)
		}

	case "revizyon":
		// Türkçe Yorum: TTO, dekan, komisyon veya hakem revizyon istediğinde proje koordinatörüne gider.
		subject = fmt.Sprintf("Proje Revizyon Talebi - %s", projeKodu)
		greeting = fmt.Sprintf("Sayın %s,", coordName)
		message = fmt.Sprintf("BAP projeniz için inceleme mercileri tarafından revizyon (düzeltme) talep edilmiştir. Lütfen belirtilen açıklamalar doğrultusunda düzeltmeleri tamamlayıp tekrar gönderiniz.")
		if coordEposta != "" {
			recipients = append(recipients, coordEposta)
		}

	case "reddedildi":
		// Türkçe Yorum: Proje reddedildiğinde koordinatöre gider.
		subject = fmt.Sprintf("Proje Başvurusu Reddedildi - %s", projeKodu)
		greeting = fmt.Sprintf("Sayın %s,", coordName)
		message = fmt.Sprintf("Bap projeniz yapılan değerlendirmeler neticesinde maalesef reddedilmiştir. Detaylar ve red gerekçesi aşağıda belirtilmiştir.")
		if coordEposta != "" {
			recipients = append(recipients, coordEposta)
		}

	default:
		// Türkçe Yorum: Tanımlanmamış diğer tüm durum geçişlerinde yürütücüye bilgi gider.
		subject = fmt.Sprintf("Proje Durum Güncellemesi - %s", projeKodu)
		greeting = fmt.Sprintf("Sayın %s,", coordName)
		message = fmt.Sprintf("BAP projenizin durumu güncellenmiştir.")
		if coordEposta != "" {
			recipients = append(recipients, coordEposta)
		}
	}

	// Alıcı listesi boş ise gönderme
	if len(recipients) == 0 {
		log.Printf("E-posta gönderimi iptal edildi: Alıcı bulunamadı (ProjeID: %d, Durum: %s)", projeID, yeniDurum)
		return
	}

	// HTML Şablonunu oluştur
	htmlBody := s.FormatEmailTemplate(greeting, message, projeKodu, baslikTr, coordName, statusLabel, aciklama)

	// E-postayı gönder
	err = s.SendEmailSMTP(recipients, subject, htmlBody)
	if err != nil {
		log.Printf("E-posta gönderiminde hata oluştu (Proje: %s, Durum: %s): %v", projeKodu, yeniDurum, err)
	} else {
		log.Printf("E-posta başarıyla gönderildi/simüle edildi (Proje: %s, Alıcılar: %v)", projeKodu, recipients)
	}
}

// getEmailsByRole belirtilen role sahip kullanıcıların e-postalarını getirir.
// Türkçe Yorum: Belirtilen sistemsel roldeki (tto, komisyon vb.) tüm kullanıcıların aktif e-posta adreslerini sorgular.
func (s *EpostaService) getEmailsByRole(role string) []string {
	query := `
		SELECT DISTINCT u.eposta 
		FROM uye u
		LEFT JOIN sistem_rol sr ON u.uye_id = sr.uye_id
		LEFT JOIN sistem_rol_tanimlama srt ON sr.sistem_rol_id = srt.rol_id
		WHERE (u.rol = $1 OR srt.rol_adi = $1) AND u.aktif_mi = true AND u.eposta IS NOT NULL AND u.eposta != ''
	`
	rows, err := s.DB.Query(query, role)
	if err != nil {
		log.Printf("Role göre e-postalar çekilirken hata: %v", err)
		return nil
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err == nil {
			emails = append(emails, email)
		}
	}
	return emails
}

// getEmailsByRoleAndDepartment belirtilen role ve bölüm/fakülte eşleşmesine sahip kullanıcıların e-postalarını getirir.
// Türkçe Yorum: Bölüm bazlı eşleştirme yaparak (örn: Dekan) doğru hedef kişileri bulur; bulunamazsa tüm roldekilere döner.
func (s *EpostaService) getEmailsByRoleAndDepartment(role, department string) []string {
	var emails []string
	if department != "" {
		query := `
			SELECT DISTINCT u.eposta 
			FROM uye u
			LEFT JOIN sistem_rol sr ON u.uye_id = sr.uye_id
			LEFT JOIN sistem_rol_tanimlama srt ON sr.sistem_rol_id = srt.rol_id
			WHERE (u.rol = $1 OR srt.rol_adi = $1) AND u.aktif_mi = true 
			  AND u.bolum = $2 AND u.eposta IS NOT NULL AND u.eposta != ''
		`
		rows, err := s.DB.Query(query, role, department)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var email string
				if err := rows.Scan(&email); err == nil {
					emails = append(emails, email)
				}
			}
		}
	}

	// Eğer bölüme ait özel dekan bulunamadıysa, sistemdeki tüm dekanlara gönder.
	if len(emails) == 0 {
		return s.getEmailsByRole(role)
	}

	return emails
}

// getAssignedHakems projeye atanmış hakemlerin e-posta adreslerini çeker.
// Türkçe Yorum: Değerlendirmesi 'Bekliyor' olan hakemleri tespit edip e-postalarını listeler.
func (s *EpostaService) getAssignedHakems(projeID int) []string {
	query := `
		SELECT DISTINCT u.eposta 
		FROM proje_degerlendirmeleri pd
		JOIN uye u ON pd.hakem_id = u.uye_id
		WHERE pd.proje_id = $1 AND pd.durum = 'Bekliyor' AND u.aktif_mi = true 
		  AND u.eposta IS NOT NULL AND u.eposta != ''
	`
	rows, err := s.DB.Query(query, projeID)
	if err != nil {
		log.Printf("Hakem e-postaları çekilirken hata: %v", err)
		return nil
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err == nil {
			emails = append(emails, email)
		}
	}
	return emails
}

// saveNotificationDB e-posta alıcıları için veritabanına bildirim kaydı ekler.
// Türkçe Yorum: E-posta gönderilen kullanıcıların sistem içi bildirim kutusunda da bu mesajı görebilmesi için veritabanına yazar.
func (s *EpostaService) saveNotificationDB(to []string, subject string, body string) {
	if len(to) == 0 {
		return
	}

	// Alıcı e-postalarına karşılık gelen üye ID'lerini sorgula
	query := `
		SELECT uye_id FROM uye 
		WHERE eposta = ANY(string_to_array($1, ',')) AND aktif_mi = true
	`
	emailsStr := strings.Join(to, ",")
	rows, err := s.DB.Query(query, emailsStr)
	if err != nil {
		log.Printf("Bildirim kaydedilirken kullanıcı sorgulama hatası: %v", err)
		return
	}
	defer rows.Close()

	var uyeIDs []int
	for rows.Next() {
		var uid int
		if err := rows.Scan(&uid); err == nil {
			uyeIDs = append(uyeIDs, uid)
		}
	}

	if len(uyeIDs) == 0 {
		return
	}

	// Her bir alıcı için veritabanına bildirim kaydı ekle
	insertQuery := `
		INSERT INTO bildirim (uye_id, baslik, icerik)
		VALUES ($1, $2, $3)
	`
	for _, uid := range uyeIDs {
		_, err := s.DB.Exec(insertQuery, uid, subject, body)
		if err != nil {
			log.Printf("Kullanıcıya (UyeID: %d) bildirim kaydı eklenemedi: %v", uid, err)
		}
	}
}

// SendPurchaseNotificationEmail satın alma taleplerinde alıcılara e-posta ve sistem içi bildirim gönderir.
// Türkçe Yorum: Satın alma talebi oluşturulduğunda veya onaylandığında/reddedildiğinde e-posta ve db bildirimi tetikler.
func (s *EpostaService) SendPurchaseNotificationEmail(talepID int, eventType string, islemYapanID int) {
	// Talep detaylarını sorgula
	var (
		projeKodu, projeBaslik, akademisyenAd, akademisyenEposta, malzemeAdi, durum, kalemAdi, redNedeni string
		miktar                                                                                          int
		birimFiyat                                                                                      float64
	)

	query := `
		SELECT COALESCE(p.proje_kodu, 'KODSUZ'), COALESCE(p.baslik_tr, 'Başlıksız Proje'),
		       COALESCE(u.ad || ' ' || u.soyad, 'Akademisyen'), COALESCE(u.eposta, ''),
		       COALESCE(sat.malzeme_adi, ''), COALESCE(sat.durum, 'Beklemede'),
		       COALESCE(bk.kategori_adi, ''), COALESCE(sat.red_nedeni, ''),
		       COALESCE(sat.miktar, 0), COALESCE(sat.birim_fiyat, 0.0)
		FROM satinalma_talebi sat
		JOIN proje p ON sat.proje_id = p.proje_id
		JOIN uye u ON sat.uye_id = u.uye_id
		LEFT JOIN butce b ON sat.kalem_id = b.kalem_id
		LEFT JOIN butce_kategori bk ON b.kategori_id = bk.kategori_id
		WHERE sat.talep_id = $1
	`
	err := s.DB.QueryRow(query, talepID).Scan(
		&projeKodu, &projeBaslik, &akademisyenAd, &akademisyenEposta,
		&malzemeAdi, &durum, &kalemAdi, &redNedeni, &miktar, &birimFiyat,
	)
	if err != nil {
		log.Printf("Satın alma e-posta bildirimi için talep bulunamadı (TalepID: %d): %v", talepID, err)
		return
	}

	var recipients []string
	var greeting, message, subject string

	toplamTutar := float64(miktar) * birimFiyat

	switch eventType {
	case "create":
		subject = fmt.Sprintf("Yeni Satın Alma Talebi Oluşturuldu - %s", projeKodu)
		greeting = "Sayın Yetkili / Akademisyen,"
		message = fmt.Sprintf("%s tarafından '%s' başlıklı proje için yeni bir satın alma talebi oluşturulmuştur.", akademisyenAd, projeBaslik)
		
		// Alıcılar: Talebi oluşturan akademisyen ve tüm TTO üyeleri
		if akademisyenEposta != "" {
			recipients = append(recipients, akademisyenEposta)
		}
		ttoEmails := s.getEmailsByRole("tto")
		recipients = append(recipients, ttoEmails...)

	case "update":
		subject = fmt.Sprintf("Satın Alma Talebi Sonucu - %s", projeKodu)
		greeting = fmt.Sprintf("Sayın %s,", akademisyenAd)
		if durum == "Onaylandı" {
			message = fmt.Sprintf("Yaptığınız satın alma talebi TTO tarafından onaylanmıştır.")
		} else {
			message = fmt.Sprintf("Yaptığınız satın alma talebi TTO tarafından reddedilmiştir.")
		}
		
		// Alıcılar: Sadece talebi oluşturan akademisyen
		if akademisyenEposta != "" {
			recipients = append(recipients, akademisyenEposta)
		}
	}

	if len(recipients) == 0 {
		return
	}

	// Satın alma detay tablosu HTML'i
	detayHTML := fmt.Sprintf(`
		<tr><td class="label">Malzeme/Hizmet:</td><td>%s</td></tr>
		<tr><td class="label">Bütçe Kalemi:</td><td>%s</td></tr>
		<tr><td class="label">Miktar:</td><td>%d</td></tr>
		<tr><td class="label">Birim Fiyat:</td><td>%.2f ₺</td></tr>
		<tr><td class="label">Toplam Tutar:</td><td>%.2f ₺</td></tr>
	`, malzemeAdi, kalemAdi, miktar, birimFiyat, toplamTutar)

	if eventType == "update" && durum == "Reddedildi" && redNedeni != "" {
		detayHTML += fmt.Sprintf(`<tr><td class="label" style="color:red;">Red Gerekçesi:</td><td style="color:red; font-weight:bold;">%s</td></tr>`, redNedeni)
	}

	htmlBody := s.FormatEmailTemplate(greeting, message, projeKodu, projeBaslik, akademisyenAd, durum, "")
	// HTML tablosunun içine detayları enjekte et (FormatEmailTemplate'deki durum hücresinin ardına)
	htmlBody = strings.Replace(htmlBody, "<strong>"+durum+"</strong></td></tr>", "<strong>"+durum+"</strong></td></tr>"+detayHTML, 1)

	// E-posta gönder
	err = s.SendEmailSMTP(recipients, subject, htmlBody)
	if err != nil {
		log.Printf("Satın alma e-posta gönderim hatası: %v", err)
	}
}
