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

// createLDAPConnection LDAP veya LDAPS soket bağlantısını oluşturur
func (s *LDAPService) createLDAPConnection() (*ldap.Conn, error) {
	cleanHost := strings.TrimPrefix(strings.TrimPrefix(configs.AppConfig.LDAPHost, "ldaps://"), "ldap://")
	port := configs.AppConfig.LDAPPort
	if strings.Contains(cleanHost, ":") {
		parts := strings.Split(cleanHost, ":")
		cleanHost = parts[0]
		if len(parts) > 1 && parts[1] != "" {
			port = parts[1]
		}
	}
	if port == "" {
		port = "389"
	}

	isTLS := port == "636" || port == "3269" || strings.HasPrefix(configs.AppConfig.LDAPHost, "ldaps://") || strings.HasPrefix(configs.AppConfig.LDAPURL, "ldaps://")

	if isTLS {
		dialURL := fmt.Sprintf("ldaps://%s:%s", cleanHost, port)
		tlsConf := &tls.Config{InsecureSkipVerify: true}
		return ldap.DialURL(dialURL, ldap.DialWithTLSConfig(tlsConf))
	}

	dialURL := fmt.Sprintf("ldap://%s:%s", cleanHost, port)
	return ldap.DialURL(dialURL)
}

// extractUserInfo LDAP Entry kaydından kullanıcı bilgilerini ayıklar
func (s *LDAPService) extractUserInfo(entry *ldap.Entry, email, username string) *LDAPUserInfo {
	ad := entry.GetAttributeValue("givenName")
	soyad := entry.GetAttributeValue("sn")
	ldapMail := entry.GetAttributeValue("mail")
	displayName := entry.GetAttributeValue("displayName")
	cn := entry.GetAttributeValue("cn")

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
	}

	if ldapMail == "" {
		ldapMail = email
	}

	return &LDAPUserInfo{
		Eposta: ldapMail,
		Ad:     ad,
		Soyad:  soyad,
	}
}

// deriveNameFromEmail e-posta adresinden varsayılan ad ve soyad türetir
func (s *LDAPService) deriveNameFromEmail(email string) (string, string) {
	username := strings.Split(email, "@")[0]
	parts := strings.Split(username, ".")
	if len(parts) >= 2 {
		return s.capitalize(parts[0]), s.capitalize(parts[1])
	}
	return s.capitalize(username), "LDAP"
}

// authenticateLiveUser gerçek Active Directory / LDAP sunucusuna bağlanır
func (s *LDAPService) authenticateLiveUser(email, password string) (*LDAPUserInfo, error) {
	ldapHost := fmt.Sprintf("%s:%s", configs.AppConfig.LDAPHost, configs.AppConfig.LDAPPort)
	log.Printf("🌐 [CANLI LDAP BAĞLANTISI] Sunucu: %s | Hedef Kullanıcı: %s\n", ldapHost, email)
	username := strings.Split(email, "@")[0]

	// -------------------------------------------------------------
	// 1. STRATEJİ: Doğrudan Kullanıcı Bağlantısı (Direct User Bind)
	// Active Directory genellikle kullanıcının doğrudan UPN (email) veya username ile bağlanmasına izin verir.
	// Bu sayede Admin/Servis hesabı şifresi hatalı veya yetkisiz olsa dahi kullanıcı başarıyla doğrulanır!
	// -------------------------------------------------------------
	log.Printf("🔐 [CANLI LDAP] 1. Aşama: Doğrudan kullanıcı yetkilendirmesi deneniyor (%s)...\n", email)
	directConn, err := s.createLDAPConnection()
	if err != nil {
		log.Printf("❌ [CANLI LDAP SOKET HATA] %s sunucusuna bağlanılamadı: %v\n", ldapHost, err)
		return nil, fmt.Errorf("LDAP sunucusuna bağlanılamadı: %w", err)
	}
	defer directConn.Close()

	// Sırasıyla email (UPN formatı) ve kullanıcı adı (sAMAccountName) ile doğrudan bind denenir
	directOk := directConn.Bind(email, password) == nil
	if !directOk {
		directOk = directConn.Bind(username, password) == nil
	}

	if directOk {
		log.Printf("🎉 [CANLI LDAP DIRECT BIND SUCCESS] Kullanıcı (%s) şifresi doğrudan LDAP tarafından doğrulandı!\n", email)

		// Kullanıcı başarıyla bağlandı. Şimdi kendi bilgilerini (ad, soyad vb.) okumaya çalışalım
		filter := fmt.Sprintf("(|(mail=%s)(userPrincipalName=%s)(sAMAccountName=%s)(uid=%s))",
			ldap.EscapeFilter(email), ldap.EscapeFilter(email), ldap.EscapeFilter(username), ldap.EscapeFilter(username))
		searchRequest := ldap.NewSearchRequest(
			configs.AppConfig.LDAPBaseDN,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			filter,
			[]string{"dn", "givenName", "sn", "mail", "cn", "displayName", "userPrincipalName", "sAMAccountName"},
			nil,
		)
		sr, sErr := directConn.Search(searchRequest)
		if sErr == nil && len(sr.Entries) > 0 {
			return s.extractUserInfo(sr.Entries[0], email, username), nil
		}

		// Eğer kullanıcı yetkisi nedeniyle arama başarısız olursa, şifre doğru olduğu için e-postadan ad-soyad türetilerek giriş onaylanır
		ad, soyad := s.deriveNameFromEmail(email)
		return &LDAPUserInfo{
			Eposta: email,
			Ad:     ad,
			Soyad:  soyad,
		}, nil
	}

	log.Printf("ℹ️  [CANLI LDAP DIRECT BIND BAŞARISIZ] 2. Aşama: Servis hesabı (%s) ile arama yöntemine geçiliyor...\n", configs.AppConfig.LDAPBindDN)

	// -------------------------------------------------------------
	// 2. STRATEJİ: Servis Hesabı ile Arama + Kullanıcı Şifre Doğrulama
	// -------------------------------------------------------------
	adminConn, err := s.createLDAPConnection()
	if err != nil {
		return nil, fmt.Errorf("LDAP sunucusuna bağlanılamadı: %w", err)
	}
	defer adminConn.Close()

	if err := adminConn.Bind(configs.AppConfig.LDAPBindDN, configs.AppConfig.LDAPBindPassword); err != nil {
		log.Printf("❌ [CANLI LDAP ADMIN BIND HATA] %v\n", err)
		return nil, fmt.Errorf("LDAP servis hesabı (BindDN) doğrulanamadı: %w", err)
	}
	log.Printf("✅ [CANLI LDAP ADMIN BIND] Servis hesabı başarıyla yetkilendirildi.\n")

	filter := fmt.Sprintf("(|(mail=%s)(userPrincipalName=%s)(sAMAccountName=%s)(sAMAccountName=%s)(uid=%s))",
		ldap.EscapeFilter(email), ldap.EscapeFilter(email), ldap.EscapeFilter(username), ldap.EscapeFilter(email), ldap.EscapeFilter(username))

	customFilter := strings.TrimSpace(configs.AppConfig.LDAPUserFilter)
	if customFilter != "" &&
		customFilter != "(&(objectClass=user)(sAMAccountName=%s))" &&
		customFilter != "(&(objectClass=person)(mail=%s))" {
		if strings.Contains(customFilter, "{{username}}") {
			filter = strings.ReplaceAll(customFilter, "{{username}}", ldap.EscapeFilter(username))
		} else if strings.Contains(customFilter, "{{email}}") {
			filter = strings.ReplaceAll(customFilter, "{{email}}", ldap.EscapeFilter(email))
		} else if strings.Contains(customFilter, "%s") {
			if strings.Contains(customFilter, "sAMAccountName") && !strings.Contains(customFilter, "mail") && !strings.Contains(customFilter, "userPrincipalName") {
				filter = fmt.Sprintf(customFilter, ldap.EscapeFilter(username))
			} else {
				filter = fmt.Sprintf(customFilter, ldap.EscapeFilter(email))
			}
		} else {
			filter = customFilter
		}
	}

	searchRequest := ldap.NewSearchRequest(
		configs.AppConfig.LDAPBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"dn", "givenName", "sn", "mail", "cn", "displayName", "userPrincipalName", "sAMAccountName"},
		nil,
	)

	sr, err := adminConn.Search(searchRequest)
	if err != nil {
		log.Printf("❌ [CANLI LDAP SEARCH HATA] %v\n", err)
		return nil, fmt.Errorf("LDAP arama hatası: %w", err)
	}

	if len(sr.Entries) == 0 {
		log.Printf("⚠️ [CANLI LDAP KULLANICI YOK] '%s' kriteriyle kullanıcı bulunamadı.\n", email)
		return nil, errors.New("Kullanıcı bulunamadı. Lütfen TTO yetkilisi ile irtibata geçiniz.")
	}

	entry := sr.Entries[0]
	userDN := entry.DN
	upn := entry.GetAttributeValue("userPrincipalName")
	samAccount := entry.GetAttributeValue("sAMAccountName")

	// 3. Kullanıcı Şifresini Doğrulama (Temiz bir bağlantı üzerinde denenir)
	userConn, err := s.createLDAPConnection()
	if err != nil {
		return nil, fmt.Errorf("LDAP bağlantı hatası: %w", err)
	}
	defer userConn.Close()

	// Sırasıyla userDN, upn, email ve sAMAccountName ile denenir
	bindOk := userConn.Bind(userDN, password) == nil
	if !bindOk && upn != "" {
		bindOk = userConn.Bind(upn, password) == nil
	}
	if !bindOk {
		bindOk = userConn.Bind(email, password) == nil
	}
	if !bindOk && samAccount != "" {
		bindOk = userConn.Bind(samAccount, password) == nil
	}

	if !bindOk {
		log.Printf("❌ [CANLI LDAP USER BIND HATA] Kullanıcı şifresi doğrulanamadı: %s\n", email)
		return nil, errors.New("Kullanıcı bilgileri yanlış")
	}

	log.Printf("🎉 [CANLI LDAP SUCCESS] Kullanıcı şifresi başarıyla doğrulandı: %s\n", email)
	return s.extractUserInfo(entry, email, username), nil
}

// capitalize metnin ilk harfini büyük, diğerlerini küçük yapar
func (s *LDAPService) capitalize(val string) string {
	if len(val) == 0 {
		return ""
	}
	val = strings.ToLower(val)
	return strings.ToUpper(string(val[0])) + val[1:]
}
