package api

import (
	"net/http"
	"strings"

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
		// Profil tamamlama ve profil bilgi endpoint'leri muaf tutulur
		path := c.Request.URL.Path
		if path == "/api/profil/tamamla" || path == "/api/profil/bilgiler" {
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

		if roleStr != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bu işlem için admin yetkisi gerekmektedir"})
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

		isAllowed := false
		for _, allowedRole := range allowedRoles {
			if roleStr == allowedRole {
				isAllowed = true
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
