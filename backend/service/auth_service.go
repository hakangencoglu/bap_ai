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
	existingUye, _ := s.UyeRepo.GetUyeByEmail(req.IletisimMail)
	if existingUye != nil {
		return nil, errors.New("bu e-posta adresi zaten kayıtlı")
	}

	// Şifre bcrypt ile hashlenir
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("şifre hashlenemedi: %w", err)
	}

	// Varsayılan rol: ogrenci (role_id=3). Eğer istekte belirtilmişse o kullanılır.
	roleID := 3
	if req.RoleID > 0 {
		roleID = req.RoleID
	}

	// Yeni üye nesnesi oluşturulur
	uye := &models.Uye{
		RoleID:       roleID,
		Unvan:        req.Unvan,
		Ad:           req.Ad,
		Soyad:        req.Soyad,
		Bolum:        req.Bolum,
		IletisimTel:  req.IletisimTel,
		IletisimMail: req.IletisimMail,
		PasswordHash: string(hashedPassword),
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
	uye, err := s.UyeRepo.GetUyeByEmail(req.IletisimMail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("Kayıtlı böyle bir kullanıcı bulunamadı")
		}
		return nil, fmt.Errorf("kullanıcı sorgulanamadı: %w", err)
	}

	// Hesabın aktif olup olmadığı kontrol edilir
	if !uye.IsActive {
		return nil, errors.New("hesap devre dışı")
	}

	// Girilen şifre, veritabanındaki hash ile karşılaştırılır
	if err := bcrypt.CompareHashAndPassword([]byte(uye.PasswordHash), []byte(req.Password)); err != nil {
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
		"email":   uye.IletisimMail,
		"role_id": uye.RoleID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // 24 saat geçerli
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
