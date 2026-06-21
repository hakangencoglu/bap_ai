package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"bap_ai/configs"

	"github.com/go-ldap/ldap/v3"
)

// LDAPUserInfo LDAP sunucusundan dönen kullanıcı detaylarını tutar
type LDAPUserInfo struct {
	Eposta string
	Ad     string
	Soyad  string
}

// LDAPService LDAP kimlik doğrulama işlemlerini yönetir
type LDAPService struct{}

// NewLDAPService yeni bir LDAPService nesnesi oluşturur
func NewLDAPService() *LDAPService {
	return &LDAPService{}
}

// AuthenticateUser kullanıcının e-posta ve şifresini LDAP üzerinden doğrular
func (s *LDAPService) AuthenticateUser(email, password string) (*LDAPUserInfo, error) {
	// Türkçe Yorum: LDAP yapılandırması aktif değilse hata döndür
	if !configs.AppConfig.LDAPEnabled {
		return nil, errors.New("LDAP kimlik doğrulama sistemi devre dışı")
	}

	// Türkçe Yorum: Şifre uzunluğu kontrol edilir
	if len(password) < 6 {
		return nil, errors.New("şifre en az 6 karakter olmalıdır")
	}

	// Türkçe Yorum: MOCK modu aktifse, simülasyon adımlarına geç
	if configs.AppConfig.LDAPMock {
		return s.authenticateMockUser(email, password)
	}

	// Türkçe Yorum: Canlı LDAP sunucusuna bağlanıp sorgula
	return s.authenticateLiveUser(email, password)
}

// authenticateMockUser test ortamı için LDAP simülasyonu yapar
func (s *LDAPService) authenticateMockUser(email, password string) (*LDAPUserInfo, error) {
	log.Printf("LDAP Mock Giriş Denemesi: %s\n", email)

	// Türkçe Yorum: E-posta formatı kontrol edilir
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[1] != "izu.edu.tr" {
		return nil, errors.New("LDAP doğrulaması başarısız: Sadece @izu.edu.tr uzantılı e-postalar kabul edilir")
	}

	// Türkçe Yorum: E-posta ön ekinden Ad ve Soyad tahmin edilir (örn: ahmet.yilmaz -> Ahmet Yılmaz)
	prefix := parts[0]
	nameParts := strings.Split(prefix, ".")
	ad := ""
	soyad := ""

	if len(nameParts) >= 2 {
		ad = s.capitalize(nameParts[0])
		soyad = s.capitalize(nameParts[1])
	} else {
		ad = s.capitalize(prefix)
		soyad = "LDAP"
	}

	// Türkçe Yorum: Mock doğrulama başarılı kabul edilerek kullanıcı bilgileri dönülür
	return &LDAPUserInfo{
		Eposta: email,
		Ad:     ad,
		Soyad:  soyad,
	}, nil
}

// authenticateLiveUser gerçek LDAP sunucusu üzerinden doğrulama yapar
func (s *LDAPService) authenticateLiveUser(email, password string) (*LDAPUserInfo, error) {
	ldapHost := fmt.Sprintf("%s:%s", configs.AppConfig.LDAPHost, configs.AppConfig.LDAPPort)
	log.Printf("LDAP Sunucusuna Bağlanılıyor: %s\n", ldapHost)

	// 1. LDAP sunucusuna bağlan
	l, err := ldap.DialURL("ldap://" + ldapHost)
	if err != nil {
		return nil, fmt.Errorf("LDAP sunucusuna bağlanılamadı: %w", err)
	}
	defer l.Close()

	// 2. Arama yapabilmek için admin kullanıcısı ile bind (bağlanma) işlemi gerçekleştir
	err = l.Bind(configs.AppConfig.LDAPBindDN, configs.AppConfig.LDAPBindPassword)
	if err != nil {
		return nil, fmt.Errorf("LDAP admin bind başarısız: %w", err)
	}

	// 3. Kullanıcıyı filtreye göre ara
	filter := fmt.Sprintf(configs.AppConfig.LDAPUserFilter, email)
	searchRequest := ldap.NewSearchRequest(
		configs.AppConfig.LDAPBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"dn", "givenName", "sn", "mail"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("LDAP kullanıcı araması başarısız: %w", err)
	}

	if len(sr.Entries) == 0 {
		return nil, sql.ErrNoRows // Kullanıcı bulunamadı
	}

	entry := sr.Entries[0]
	userDN := entry.DN
	ad := entry.GetAttributeValue("givenName")
	soyad := entry.GetAttributeValue("sn")
	ldapMail := entry.GetAttributeValue("mail")

	if ldapMail == "" {
		ldapMail = email
	}

	// 4. Kullanıcının kendi şifresi ile bind (bağlanma) işlemini doğrula
	err = l.Bind(userDN, password)
	if err != nil {
		return nil, errors.New("LDAP doğrulaması başarısız: Kullanıcı adı veya şifre yanlış")
	}

	return &LDAPUserInfo{
		Eposta: ldapMail,
		Ad:     ad,
		Soyad:  soyad,
	}, nil
}

// capitalize metnin ilk harfini büyük, diğerlerini küçük yapar
func (s *LDAPService) capitalize(val string) string {
	if len(val) == 0 {
		return ""
	}
	// Türkçe karakter uyumluluğu için basit büyük/küçük harf dönüşümü
	val = strings.ToLower(val)
	return strings.ToUpper(string(val[0])) + val[1:]
}
