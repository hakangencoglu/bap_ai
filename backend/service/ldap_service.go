package service

import (
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
		return nil, errors.New("Kullanıcı bulunamadı. Lütfen TTO yetkilisi ile irtibata geçiniz.")
	}

	emailLower := strings.ToLower(strings.TrimSpace(email))
	prefix := parts[0]

	// Türkçe Yorum: Test/Mock ortamında kayıtsız / olmayan hesabı simüle etmek için
	// 'tanimsiz', 'taninmayan', 'bulunamadi', 'invalid', 'yok', 'hata', 'nonexistent' e-postaları reddedilir.
	invalidKeywords := []string{"tanimsiz", "taninmayan", "bulunamadi", "invalid", "yok", "hata", "nonexistent"}
	for _, kw := range invalidKeywords {
		if strings.Contains(prefix, kw) {
			return nil, errors.New("Kullanıcı bulunamadı. Lütfen TTO yetkilisi ile irtibata geçiniz.")
		}
	}

	// Türkçe Yorum: E-posta ön ekinden Ad ve Soyad üretilir (örn: furkan.cakir -> Furkan Cakir)
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

	return &LDAPUserInfo{
		Eposta: emailLower,
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
	username := strings.Split(email, "@")[0]
	filter := fmt.Sprintf("(|(mail=%s)(userPrincipalName=%s)(sAMAccountName=%s))", email, email, username)
	if configs.AppConfig.LDAPUserFilter != "" {
		filter = fmt.Sprintf(configs.AppConfig.LDAPUserFilter, email)
	}

	searchRequest := ldap.NewSearchRequest(
		configs.AppConfig.LDAPBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"dn", "givenName", "sn", "mail", "cn", "displayName"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil || len(sr.Entries) == 0 {
		return nil, errors.New("Kullanıcı bulunamadı. Lütfen TTO yetkilisi ile irtibata geçiniz.")
	}

	entry := sr.Entries[0]
	userDN := entry.DN
	ad := entry.GetAttributeValue("givenName")
	soyad := entry.GetAttributeValue("sn")
	ldapMail := entry.GetAttributeValue("mail")

	if ad == "" && soyad == "" {
		displayName := entry.GetAttributeValue("displayName")
		if displayName == "" {
			displayName = entry.GetAttributeValue("cn")
		}
		if displayName != "" {
			parts := strings.Split(displayName, " ")
			if len(parts) >= 2 {
				ad = parts[0]
				soyad = strings.Join(parts[1:], " ")
			} else {
				ad = displayName
				soyad = "LDAP"
			}
		} else {
			ad = s.capitalize(username)
			soyad = "LDAP"
		}
	}

	if ldapMail == "" {
		ldapMail = email
	}

	// 4. Kullanıcının kendi şifresi ile bind (bağlanma) işlemini doğrula
	err = l.Bind(userDN, password)
	if err != nil {
		return nil, errors.New("Kullanıcı bilgileri yanlış")
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
