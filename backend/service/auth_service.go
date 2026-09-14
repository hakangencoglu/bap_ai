package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
	"bap_ai/configs"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService yapısı, kimlik doğrulama iş mantığını barındırır.
type AuthService struct {
	UyeRepo *repository.UyeRepository
}

// NewAuthService fonksiyonu, yeni bir AuthService nesnesi döner.
func NewAuthService(uyeRepo *repository.UyeRepository) *AuthService {
	return &AuthService{UyeRepo: uyeRepo}
}

// Register fonksiyonu, yeni bir kullanıcıyı sisteme kaydeder.
// Sadece temel bilgiler (ad, soyad, e-posta, şifre) alınır.
// Detay bilgiler giriş sonrası profil tamamlama adımında alınır.
func (s *AuthService) Register(req *models.RegisterRequest) (*models.Uye, error) {
	// Aynı e-posta ile daha önce kayıt olunmuş mu kontrol edilir
	existingUye, _ := s.UyeRepo.GetUyeByEmail(req.Eposta)
	if existingUye != nil {
		return nil, errors.New("bu e-posta adresi zaten kayıtlı")
	}

	// Şifre bcrypt ile hashlenir
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Sifre), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("şifre hashlenemedi: %w", err)
	}

	// Yeni üye nesnesi oluşturulur (sadece temel bilgiler)
	uye := &models.Uye{
		Ad:        req.Ad,
		Soyad:     req.Soyad,
		Eposta:    req.Eposta,
		SifreHash: string(hashedPassword),
	}

	// Üye veritabanına kaydedilir
	if err := s.UyeRepo.CreateUye(uye); err != nil {
		return nil, fmt.Errorf("kullanıcı kaydedilemedi: %w", err)
	}

	// Boş bir detay kaydı oluşturulur (profil tamamlama için hazırlık)
	detay := &models.UyeDetay{
		UyeID:            uye.UyeID,
		ProfilTamamlandi: false,
	}
	if err := s.UyeRepo.CreateUyeDetay(detay); err != nil {
		return nil, fmt.Errorf("kullanıcı detay kaydı oluşturulamadı: %w", err)
	}

	return uye, nil
}

// Login fonksiyonu, kullanıcının e-posta ve şifresiyle giriş yapmasını sağlar.
// Türkçe Yorum: Öncelikle veritabanı kontrol edilir. Kullanıcı veritabanında bulunamazsa (ve LDAP aktifse) LDAP sorgulanarak otomatik kaydı oluşturulur.
func (s *AuthService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	// E-posta adresini normalize et (boşlukları sil ve küçük harfe çevir)
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Eposta))
	req.Eposta = cleanEmail

	// 1. E-posta adresine göre veritabanında üye aranır (Öncelikli kontrol)
	uye, err := s.UyeRepo.GetUyeByEmail(req.Eposta)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Türkçe Yorum: Kullanıcı veritabanında yoksa ve LDAP aktifse, LDAP üzerinden sorgulama yapıyoruz.
			if configs.AppConfig.LDAPEnabled {
				ldapService := NewLDAPService()
				ldapUser, ldapErr := ldapService.AuthenticateUser(req.Eposta, req.Sifre)
				if ldapErr != nil {
					log.Printf("❌ LDAP Giriş Hatası (%s): %v\n", req.Eposta, ldapErr)
					if ldapErr.Error() == "Kullanıcı bilgileri yanlış" || ldapErr.Error() == "şifre en az 6 karakter olmalıdır" {
						return nil, errors.New("Kullanıcı bilgileri yanlış")
					}
					return nil, errors.New("Kullanıcı veritabanında veya LDAP sisteminde bulunamadı. Lütfen TTO yetkilisi ile irtibata geçiniz.")
				}

				// Türkçe Yorum: LDAP doğrulaması başarılı oldu. Kullanıcıyı sisteme otomatik "akademisyen" rolüyle kaydediyoruz.
				hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(req.Sifre), bcrypt.DefaultCost)
				if hashErr != nil {
					return nil, fmt.Errorf("şifre hashlenemedi: %w", hashErr)
				}

				newUye := &models.Uye{
					Ad:        ldapUser.Ad,
					Soyad:     ldapUser.Soyad,
					Eposta:    ldapUser.Eposta,
					SifreHash: string(hashedPassword),
				}

				if createErr := s.UyeRepo.CreateUye(newUye); createErr != nil {
					return nil, fmt.Errorf("LDAP kullanıcısı veritabanına kaydedilemedi: %w", createErr)
				}

				newDetay := &models.UyeDetay{
					UyeID:            newUye.UyeID,
					Rol:              "akademisyen",
					IzuUyesi:         true,
					ProfilTamamlandi: false,
				}

				if detayErr := s.UyeRepo.CreateUyeDetay(newDetay); detayErr != nil {
					return nil, fmt.Errorf("LDAP kullanıcı detay kaydı oluşturulamadı: %w", detayErr)
				}

				_ = s.UyeRepo.UpdateUyeRol(newUye.UyeID, "akademisyen")
				_ = s.UyeRepo.UpsertSistemRol(newUye.UyeID, "akademisyen")

				uye, err = s.UyeRepo.GetUyeByEmail(req.Eposta)
				if err != nil {
					return nil, fmt.Errorf("yeni oluşturulan LDAP kullanıcısı sorgulanamadı: %w", err)
				}
			} else {
				return nil, errors.New("Kullanıcı veritabanında veya LDAP sisteminde bulunamadı. Lütfen TTO yetkilisi ile irtibata geçiniz.")
			}
		} else {
			return nil, fmt.Errorf("kullanıcı sorgulanamadı: %w", err)
		}
	}

	// Hesabın aktif olup olmadığı kontrol edilir
	if !uye.AktifMi {
		return nil, errors.New("hesap devre dışı")
	}

	// Şifresi henüz tanımlanmamış admin tarafından eklenen kullanıcı kontrolü
	if uye.SifreHash == "pending" {
		return nil, errors.New("sifre_olusturulmali")
	}

	// Girilen şifre, veritabanındaki hash ile karşılaştırılır
	if err := bcrypt.CompareHashAndPassword([]byte(uye.SifreHash), []byte(req.Sifre)); err != nil {
		// Türkçe Yorum: Veritabanındaki şifre uyuşmadıysa ve LDAP aktifse, LDAP üzerinden şifre güncellenmiş olabilir
		if configs.AppConfig.LDAPEnabled && (uye.IzuUyesi || strings.HasSuffix(cleanEmail, "@izu.edu.tr")) {
			ldapService := NewLDAPService()
			if _, ldapErr := ldapService.AuthenticateUser(cleanEmail, req.Sifre); ldapErr == nil {
				// LDAP doğrulaması başarılı! Yerel şifreyi yeni şifre ile güncelle
				newHash, hashErr := bcrypt.GenerateFromPassword([]byte(req.Sifre), bcrypt.DefaultCost)
				if hashErr == nil {
					_ = s.UyeRepo.UpdateUyePassword(uye.UyeID, string(newHash))
					uye.SifreHash = string(newHash)
				}
				goto PasswordVerified
			}
		}
		return nil, errors.New("Kullanıcı bilgileri yanlış")
	}
PasswordVerified:

	// Giriş yapınca şifre değiştirilmesi zorlanmış mı kontrol edilir
	if uye.SifreDegistirZorla {
		return nil, errors.New("sifre_olusturulmali")
	}

	// JWT token oluşturulur
	token, err := generateToken(uye)
	if err != nil {
		return nil, fmt.Errorf("token oluşturulamadı: %w", err)
	}

	return &models.LoginResponse{
		Token: token,
		Uye:   *uye,
	}, nil
}

// generateToken fonksiyonu, üye bilgilerine göre JWT token üretir.
// Token'a profil_tamamlandi bilgisi de eklenir.
func generateToken(uye *models.UyeWithDetay) (string, error) {
	// Token için claim bilgileri belirlenir
	claims := jwt.MapClaims{
		"uye_id":             uye.UyeID,
		"email":              uye.Eposta,
		"role":               uye.Rol,
		"profil_tamamlandi":  uye.ProfilTamamlandi,
		"exp":                time.Now().Add(24 * time.Hour).Unix(),
		"iat":                time.Now().Unix(),
	}

	// Token oluşturulur ve imzalanır
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(configs.AppConfig.JWTSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// GenerateTokenForUye fonksiyonu, dışarıdan token oluşturma imkanı verir.
// Profil tamamlama sonrası yeni token üretmek için kullanılır.
func (s *AuthService) GenerateTokenForUye(uyeID int) (string, error) {
	// Güncel üye bilgileri alınır
	uye, err := s.UyeRepo.GetUyeByID(uyeID)
	if err != nil {
		return "", fmt.Errorf("üye bilgileri alınamadı: %w", err)
	}

	// Yeni token üretilir
	token, err := generateToken(uye)
	if err != nil {
		return "", fmt.Errorf("token oluşturulamadı: %w", err)
	}

	return token, nil
}

// SetPassword ilk defa şifre oluşturacak kullanıcılar için şifre belirler ve sisteme giriş yaptırır.
func (s *AuthService) SetPassword(eposta string, sifre string) (*models.LoginResponse, error) {
	// E-posta adresine göre üye aranır
	uye, err := s.UyeRepo.GetUyeByEmail(eposta)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("Kullanıcı bulunamadı")
		}
		return nil, fmt.Errorf("kullanıcı sorgulanamadı: %w", err)
	}

	// Şifrenin beklemede veya zorla şifre değiştirme kapsamında olup olmadığı kontrol edilir
	// Türkçe Yorum: Kullanıcının şifresi pending değilse ve şifre değiştirme zorunlu kılınmamışsa hata verilir
	if uye.SifreHash != "pending" && !uye.SifreDegistirZorla {
		return nil, errors.New("bu kullanıcının şifresi zaten tanımlanmış")
	}

	// Hesabın aktif olup olmadığı kontrol edilir
	if !uye.AktifMi {
		return nil, errors.New("hesap devre dışı")
	}

	// Şifre uzunluğu kontrol edilir
	if len(sifre) < 6 {
		return nil, errors.New("şifre en az 6 karakter olmalıdır")
	}

	// Şifre bcrypt ile hashlenir
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(sifre), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("şifre hashlenemedi: %w", err)
	}

	// Şifre veritabanında güncellenir
	if err := s.UyeRepo.UpdateUyePassword(uye.UyeID, string(hashedPassword)); err != nil {
		return nil, fmt.Errorf("şifre güncellenemedi: %w", err)
	}

	// Güncel bilgileriyle JWT token oluşturulur
	uye.SifreHash = string(hashedPassword)
	token, err := generateToken(uye)
	if err != nil {
		return nil, fmt.Errorf("token oluşturulamadı: %w", err)
	}

	return &models.LoginResponse{
		Token: token,
		Uye:   *uye,
	}, nil
}

