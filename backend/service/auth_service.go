package service

import (
	"database/sql"
	"errors"
	"fmt"
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

	// Varsayılan rol: ogrenci. Eğer istekte belirtilmişse o kullanılır.
	rol := "ogrenci"
	if req.Rol != "" {
		rol = req.Rol
	}

	// Yeni üye nesnesi oluşturulur
	uye := &models.Uye{
		Rol:       rol,
		Unvan:     req.Unvan,
		Ad:        req.Ad,
		Soyad:     req.Soyad,
		Bolum:     req.Bolum,
		Telefon:   req.Telefon,
		Eposta:    req.Eposta,
		SifreHash: string(hashedPassword),
	}

	// Üye veritabanına kaydedilir
	if err := s.UyeRepo.CreateUye(uye); err != nil {
		return nil, fmt.Errorf("kullanıcı kaydedilemedi: %w", err)
	}

	return uye, nil
}

// Login fonksiyonu, kullanıcının e-posta ve şifresiyle giriş yapmasını sağlar.
func (s *AuthService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	// E-posta adresine göre üye aranır
	uye, err := s.UyeRepo.GetUyeByEmail(req.Eposta)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("Kayıtlı böyle bir kullanıcı bulunamadı")
		}
		return nil, fmt.Errorf("kullanıcı sorgulanamadı: %w", err)
	}

	// Hesabın aktif olup olmadığı kontrol edilir
	if !uye.AktifMi {
		return nil, errors.New("hesap devre dışı")
	}

	// Girilen şifre, veritabanındaki hash ile karşılaştırılır
	if err := bcrypt.CompareHashAndPassword([]byte(uye.SifreHash), []byte(req.Sifre)); err != nil {
		return nil, errors.New("Kullanıcı bilgileri yanlış")
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
func generateToken(uye *models.Uye) (string, error) {
	// Token için claim bilgileri belirlenir
	claims := jwt.MapClaims{
		"uye_id":  uye.UyeID,
		"email":   uye.Eposta,
		"role":    uye.Rol,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	// Token oluşturulur ve imzalanır
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(configs.AppConfig.JWTSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
