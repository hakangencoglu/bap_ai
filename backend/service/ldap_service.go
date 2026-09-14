package service

import (
	"crypto/tls"
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

// AuthenticateUser kullanıcının e-posta ve şifresiyle LDAP üzerinden doğrular ve gelen tüm verileri detaylıca loglar.
func (s *LDAPService) AuthenticateUser(email, password string) (*LDAPUserInfo, error) {
	log.Println("============================================================")
	log.Printf("🔑 [LDAP DOĞRULAMA İSTEĞİ GELDİ] E-posta: %s\n", email)
	log.Printf("ℹ️  [LDAP AYARLARI] Host: %s:%s | BaseDN: %s | Filter: %s | Mock: %v\n",
		configs.AppConfig.LDAPHost, configs.AppConfig.LDAPPort,
		configs.AppConfig.LDAPBaseDN, configs.AppConfig.LDAPUserFilter,
		configs.AppConfig.LDAPMock,
	)
	log.Println("------------------------------------------------------------")

	// Türkçe Yorum: LDAP yapılandırması aktif değilse hata döndür
	if !configs.AppConfig.LDAPEnabled {
		log.Println("❌ [LDAP HATA] LDAP kimlik doğrulama sistemi .env üzerinde devre dışı (LDAP_ENABLED=false)")
		log.Println("============================================================")
		return nil, errors.New("LDAP kimlik doğrulama sistemi devre dışı")
	}

	// Türkçe Yorum: Şifre uzunluğu kontrol edilir
	if len(password) < 6 {
		log.Println("❌ [LDAP HATA] Girilen şifre 6 karakterden kısa.")
		log.Println("============================================================")
		return nil, errors.New("şifre en az 6 karakter olmalıdır")
	}

	// Türkçe Yorum: MOCK modu aktifse simülasyon loglama ve doğrulama çalıştırılır
	if configs.AppConfig.LDAPMock {
		user, err := s.authenticateMockUser(email, password)
		if err != nil {
			log.Printf("❌ [MOCK LDAP SONUÇ] Başarısız: %v\n", err)
		} else {
			log.Printf("✅ [MOCK LDAP SONUÇ] Başarılı -> E-posta: %s | Ad: %s | Soyad: %s\n", user.Eposta, user.Ad, user.Soyad)
		}
		log.Println("============================================================")
		return user, err
	}

	// Türkçe Yorum: Canlı LDAP sunucusuna bağlanıp gerçek sorgu yapılır ve dönen veriler loglanır
	user, err := s.authenticateLiveUser(email, password)
	if err != nil {
		log.Printf("❌ [CANLI LDAP SONUÇ] Başarısız: %v\n", err)
	} else {
		log.Printf("🎉 [CANLI LDAP SONUÇ] Başarılı -> E-posta: %s | Ad: %s | Soyad: %s\n", user.Eposta, user.Ad, user.Soyad)
	}
	log.Println("============================================================")
	return user, err
}

// authenticateMockUser test ortamı için LDAP simülasyonu yapar ve log basar
func (s *LDAPService) authenticateMockUser(email, password string) (*LDAPUserInfo, error) {
	log.Printf("🔍 [MOCK LDAP SORGUSU] E-posta: %s sorgulanıyor...\n", email)

	// Türkçe Yorum: E-posta formatı kontrol edilir
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[1] != "izu.edu.tr" {
		log.Printf("⚠️ [MOCK LDAP REJD] %s hesabı @izu.edu.tr alan adına sahip değil.\n", email)
		return nil, errors.New("Kullanıcı bulunamadı. Lütfen TTO yetkilisi ile irtibata geçiniz.")
	}

	emailLower := strings.ToLower(strings.TrimSpace(email))

	// Türkçe Yorum: Mock LDAP rehber veritabanı
	mockDirectory := map[string]LDAPUserInfo{
		"ldap.akademisyen@izu.edu.tr": {Eposta: "ldap.akademisyen@izu.edu.tr", Ad: "LDAP", Soyad: "Akademisyen"},
		"yeni.akademisyen@izu.edu.tr": {Eposta: "yeni.akademisyen@izu.edu.tr", Ad: "Yeni", Soyad: "Akademisyen"},
		"mehmet.ak@izu.edu.tr":        {Eposta: "mehmet.ak@izu.edu.tr", Ad: "Mehmet", Soyad: "Ak"},
		"ahmet.yilmaz@izu.edu.tr":     {Eposta: "ahmet.yilmaz@izu.edu.tr", Ad: "Ahmet", Soyad: "Yılmaz"},
		"bap.akademisyen@izu.edu.tr":  {Eposta: "bap.akademisyen@izu.edu.tr", Ad: "BAP", Soyad: "Akademisyen"},
		"ldap_user@izu.edu.tr":        {Eposta: "ldap_user@izu.edu.tr", Ad: "LDAP", Soyad: "User"},
		"test.ldap@izu.edu.tr":        {Eposta: "test.ldap@izu.edu.tr", Ad: "Test", Soyad: "LDAP"},
	}

	if userInfo, found := mockDirectory[emailLower]; found {
		log.Printf("📋 [MOCK LDAP VERİSİ GELDİ] DN: cn=%s %s,dc=izu,dc=edu,dc=tr | givenName: %s | sn: %s | mail: %s\n",
			userInfo.Ad, userInfo.Soyad, userInfo.Ad, userInfo.Soyad, userInfo.Eposta)
		return &userInfo, nil
	}

	prefix := parts[0]
	if strings.HasPrefix(prefix, "ldap.") || strings.HasPrefix(prefix, "mock.") {
		nameParts := strings.Split(prefix, ".")
		ad := s.capitalize(nameParts[0])
		soyad := "LDAP"
		if len(nameParts) >= 2 {
			soyad = s.capitalize(nameParts[1])
		}
		log.Printf("📋 [MOCK LDAP DİNAMİK VERİ GELDİ] givenName: %s | sn: %s | mail: %s\n", ad, soyad, emailLower)
		return &LDAPUserInfo{
			Eposta: emailLower,
			Ad:     ad,
			Soyad:  soyad,
		}, nil
	}

	log.Printf("⚠️ [MOCK LDAP REJD] %s hesabı Mock LDAP rehberinde kayıtlı değil.\n", email)
	return nil, errors.New("Kullanıcı bulunamadı. Lütfen TTO yetkilisi ile irtibata geçiniz.")
}

// authenticateLiveUser gerçek Active Directory / LDAP sunucusuna bağlanır, dönen tüm nesne ve öznitelikleri loglar.
func (s *LDAPService) authenticateLiveUser(email, password string) (*LDAPUserInfo, error) {
	ldapHost := fmt.Sprintf("%s:%s", configs.AppConfig.LDAPHost, configs.AppConfig.LDAPPort)
	log.Printf("🌐 [CANLI LDAP BAĞLANTISI] Sunucuya bağlanılıyor: %s ...\n", ldapHost)

	// 1. LDAP sunucusuna bağlan (Port 636 veya ldaps durumunda TLS kullanılır)
	var l *ldap.Conn
	var err error
	cleanHost := strings.TrimPrefix(strings.TrimPrefix(configs.AppConfig.LDAPHost, "ldaps://"), "ldap://")
	isTLS := configs.AppConfig.LDAPPort == "636" || strings.HasPrefix(configs.AppConfig.LDAPHost, "ldaps://")

	if isTLS {
		dialURL := fmt.Sprintf("ldaps://%s:%s", cleanHost, configs.AppConfig.LDAPPort)
		tlsConf := &tls.Config{InsecureSkipVerify: true}
		l, err = ldap.DialURL(dialURL, ldap.DialWithTLSConfig(tlsConf))
	} else {
		dialURL := fmt.Sprintf("ldap://%s:%s", cleanHost, configs.AppConfig.LDAPPort)
		l, err = ldap.DialURL(dialURL)
	}

	if err != nil {
		log.Printf("❌ [CANLI LDAP DIAL HATA] %s sunucusuna bağlantı kurulamadı: %v\n", ldapHost, err)
		return nil, fmt.Errorf("LDAP sunucusuna bağlanılamadı: %w", err)
	}
	defer l.Close()
	log.Printf("✅ [CANLI LDAP DIAL] %s sunucusuna bağlantı başarılı.\n", ldapHost)

	// 2. Admin BIND işlemi
	log.Printf("🔑 [CANLI LDAP BIND] Admin DN (%s) ile bağlanılıyor...\n", configs.AppConfig.LDAPBindDN)
	err = l.Bind(configs.AppConfig.LDAPBindDN, configs.AppConfig.LDAPBindPassword)
	if err != nil {
		log.Printf("❌ [CANLI LDAP ADMIN BIND HATA] %v\n", err)
		return nil, fmt.Errorf("LDAP admin bind başarısız: %w", err)
	}
	log.Printf("✅ [CANLI LDAP ADMIN BIND] Admin yetkilendirmesi başarılı.\n")

	// 3. Kullanıcı Arama
	username := strings.Split(email, "@")[0]
	// Active Directory ve standart LDAP için en kapsamlı ve hataya dayanıklı arama filtresi
	filter := fmt.Sprintf("(|(mail=%s)(userPrincipalName=%s)(sAMAccountName=%s)(sAMAccountName=%s)(uid=%s))",
		ldap.EscapeFilter(email), ldap.EscapeFilter(email), ldap.EscapeFilter(username), ldap.EscapeFilter(email), ldap.EscapeFilter(username))

	customFilter := strings.TrimSpace(configs.AppConfig.LDAPUserFilter)
	if customFilter != "" &&
		customFilter != "(&(objectClass=user)(sAMAccountName=%s))" &&
		customFilter != "(&(objectClass=person)(mail=%s))" {
		if strings.Contains(customFilter, "sAMAccountName") && !strings.Contains(customFilter, "mail") && !strings.Contains(customFilter, "userPrincipalName") {
			filter = fmt.Sprintf(customFilter, ldap.EscapeFilter(username))
		} else {
			filter = fmt.Sprintf(customFilter, ldap.EscapeFilter(email))
		}
	}

	log.Printf("🔍 [CANLI LDAP SEARCH] BaseDN: %s | Filtre: %s\n", configs.AppConfig.LDAPBaseDN, filter)

	searchRequest := ldap.NewSearchRequest(
		configs.AppConfig.LDAPBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"dn", "givenName", "sn", "mail", "cn", "displayName", "userPrincipalName", "sAMAccountName"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Printf("❌ [CANLI LDAP SEARCH HATA] %v\n", err)
		return nil, errors.New("Kullanıcı bulunamadı. Lütfen TTO yetkilisi ile irtibata geçiniz.")
	}

	log.Printf("📊 [CANLI LDAP REHBER SONUCU] Bulunan kayıt sayısı: %d\n", len(sr.Entries))

	if len(sr.Entries) == 0 {
		log.Printf("⚠️ [CANLI LDAP KULLANICI YOK] '%s' kriteriyle LDAP rehberinde eşleşen kayıt bulunamadı.\n", email)
		return nil, errors.New("Kullanıcı bulunamadı. Lütfen TTO yetkilisi ile irtibata geçiniz.")
	}

	entry := sr.Entries[0]
	userDN := entry.DN
	ad := entry.GetAttributeValue("givenName")
	soyad := entry.GetAttributeValue("sn")
	ldapMail := entry.GetAttributeValue("mail")
	displayName := entry.GetAttributeValue("displayName")
	cn := entry.GetAttributeValue("cn")
	upn := entry.GetAttributeValue("userPrincipalName")
	samAccount := entry.GetAttributeValue("sAMAccountName")

	log.Println("📋 [CANLI LDAP GELEN ÖZNİTELİKLER (ATTRIBUTES)]")
	log.Printf("   ├─ DN: %s\n", userDN)
	log.Printf("   ├─ givenName (Ad): %s\n", ad)
	log.Printf("   ├─ sn (Soyad): %s\n", soyad)
	log.Printf("   ├─ mail (E-Posta): %s\n", ldapMail)
	log.Printf("   ├─ displayName: %s\n", displayName)
	log.Printf("   ├─ cn: %s\n", cn)
	log.Printf("   ├─ userPrincipalName: %s\n", upn)
	log.Printf("   └─ sAMAccountName: %s\n", samAccount)

	if ad == "" && soyad == "" {
		if displayName == "" {
			displayName = cn
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
		log.Printf("ℹ️  [LDAP AD/SOYAD TAMAMLAMA] givenName/sn boş olduğu için türetildi -> Ad: %s | Soyad: %s\n", ad, soyad)
	}

	if ldapMail == "" {
		ldapMail = email
	}

	// 4. Kullanıcı şifresi doğrulama (USER BIND)
	log.Printf("🔐 [CANLI LDAP USER BIND] Kullanıcı DN (%s) ve şifre ile doğrulanıyor...\n", userDN)
	err = l.Bind(userDN, password)
	if err != nil {
		log.Printf("❌ [CANLI LDAP USER BIND HATA] Şifre doğrulama başarısız: %v\n", err)
		return nil, errors.New("Kullanıcı bilgileri yanlış")
	}

	log.Printf("🎉 [CANLI LDAP USER BIND SUCCESS] Kullanıcı şifresi LDAP tarafından onaylandı!\n")

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
	val = strings.ToLower(val)
	return strings.ToUpper(string(val[0])) + val[1:]
}
