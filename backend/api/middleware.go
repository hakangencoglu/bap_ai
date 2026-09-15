package api

import (
	"net/http"
	"strings"

	"bap_ai/backend/models"
	"bap_ai/backend/repository"
	"bap_ai/configs"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware fonksiyonu, JWT token doğrulaması yapan middleware döner.
// Korumalı endpointler bu middleware arkasına konulur.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authorization başlığından token alınır
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkilendirme başlığı bulunamadı"})
			c.Abort()
			return
		}

		// "Bearer <token>" formatı kontrol edilir
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçersiz token formatı"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Token doğrulanır ve parse edilir
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(configs.AppConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçersiz veya süresi dolmuş token"})
			c.Abort()
			return
		}

		// Token claim bilgileri context'e eklenir (sonraki handler'larda kullanılabilir)
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("uye_id", claims["uye_id"])
			c.Set("email", claims["email"])
			c.Set("role", claims["role"]) // Artık string rol adı kullanıyoruz

			// Profil tamamlanma durumu context'e eklenir
			if profilTamamlandi, exists := claims["profil_tamamlandi"]; exists {
				c.Set("profil_tamamlandi", profilTamamlandi)
			} else {
				c.Set("profil_tamamlandi", false)
			}
		}

		// Sonraki handler'a geçilir
		c.Next()
	}
}

// ProfilZorunluMiddleware fonksiyonu, profil tamamlanmadan erişimi engelleyen middleware döner.
// Profil tamamlama ve profil bilgi endpoint'leri hariç tüm korumalı endpointlerde kullanılır.
func ProfilZorunluMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Profil tamamlama, fakülte-bölüm ve profil bilgi endpoint'leri muaf tutulur
		path := c.Request.URL.Path
		if path == "/api/profil/tamamla" || path == "/api/profil/bilgiler" || path == "/api/auth/check-page-access" || path == "/api/fakulteler" || path == "/api/bolumler" {
			c.Next()
			return
		}

		// Context'ten profil tamamlanma durumu alınır
		profilTamamlandi, exists := c.Get("profil_tamamlandi")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Profil bilgisi bulunamadı"})
			c.Abort()
			return
		}

		// Boolean kontrolü yapılır
		tamamlandi, ok := profilTamamlandi.(bool)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Geçersiz profil bilgisi"})
			c.Abort()
			return
		}

		// Profil tamamlanmamışsa erişim engellenir
		if !tamamlandi {
			c.JSON(http.StatusForbidden, gin.H{
				"error":              "Profil bilgilerinizi tamamlamanız gerekmektedir",
				"profil_tamamlandi":  false,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminMiddleware fonksiyonu, gelen isteğin admin yetkisine sahip olup olmadığını kontrol eder.
// Bu middleware, AuthMiddleware'den sonra çalıştırılmalıdır.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Context'ten role bilgisini al (artık string)
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Rol bilgisi bulunamadı"})
			c.Abort()
			return
		}

		// String rol kontrolü
		roleStr, ok := role.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Geçersiz rol tipi"})
			c.Abort()
			return
		}

		// Çoklu rol kontrolü (virgülle ayrılmış rolleri split edip kontrol ediyoruz)
		isAdmin := false
		roles := strings.Split(roleStr, ",")
		for _, r := range roles {
			if strings.TrimSpace(r) == "admin" {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bu işlem için admin yetkisi gerekmektedir"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SuperDeleteEmail, kalıcı silme yetkisi yalnızca bu e-postaya aittir (rol ile verilemez).
const SuperDeleteEmail = "admin1@izu.edu.tr"

// IsSuperDeleteEmail, e-postanın özel silme yetkisine sahip olup olmadığını kontrol eder.
func IsSuperDeleteEmail(email string) bool {
	return strings.EqualFold(strings.TrimSpace(email), SuperDeleteEmail)
}

// SuperDeleteMiddleware, kalıcı silme endpoint'lerini yalnızca admin1@izu.edu.tr için açar.
// Türkçe Yorum: Admin rolü olan diğer kullanıcılar (admin@izu.edu.tr dahil) bu middleware'den geçemez.
func SuperDeleteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		emailVal, exists := c.Get("email")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "E-posta bilgisi bulunamadı"})
			c.Abort()
			return
		}
		email, ok := emailVal.(string)
		if !ok || !IsSuperDeleteEmail(email) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bu silme işlemi için yetkiniz yok"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRoles fonksiyonu, gelen isteğin parametre olarak verilen rollerden biri olup olmadığını kontrol eder.
// Bu middleware, AuthMiddleware'den sonra çalıştırılmalıdır.
func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Rol bilgisi bulunamadı"})
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Geçersiz rol tipi"})
			c.Abort()
			return
		}

		// Çoklu rol kontrolü (virgülle ayrılmış rolleri split edip kontrol ediyoruz)
		isAllowed := false
		userRoles := strings.Split(roleStr, ",")
		for _, allowedRole := range allowedRoles {
			for _, userRole := range userRoles {
				if strings.TrimSpace(userRole) == allowedRole {
					isAllowed = true
					break
				}
			}
			if isAllowed {
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bu işlemi yapmak için yetkiniz bulunmamaktadır"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuditLogMiddleware, tüm kullanıcı API işlemlerini otomatik olarak veritabanındaki sistem_islem_log tablosuna kaydeder.
// Türkçe Yorum: Asenkron çalışır ve yanıt süresini etkilemeden tam denetim (Audit Trail) günlüğü tutar.
func AuditLogMiddleware(auditRepo *repository.AuditRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/static/") || path == "/favicon.ico" {
			return
		}

		var uyeID *int
		if val, exists := c.Get("uye_id"); exists {
			if idFloat, ok := val.(float64); ok {
				idInt := int(idFloat)
				uyeID = &idInt
			} else if idInt, ok := val.(int); ok {
				uyeID = &idInt
			}
		}

		email, _ := c.Get("email")
		emailStr, _ := email.(string)

		role, _ := c.Get("role")
		roleStr, _ := role.(string)

		logItem := &models.AuditLog{
			UyeID:     uyeID,
			Email:     emailStr,
			Rol:       roleStr,
			IslemTuru: c.Request.Method,
			Endpoint:  path,
			IPAdresi:  c.ClientIP(),
			DurumKodu: c.Writer.Status(),
			Aciklama:  c.Request.Method + " " + path,
		}

		go func(l *models.AuditLog) {
			if auditRepo != nil {
				_ = auditRepo.LogKaydet(l)
			}
		}(logItem)
	}
}
