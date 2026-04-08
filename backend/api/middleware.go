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
			c.Set("role_id", claims["role_id"])
		}

		// Sonraki handler'a geçilir
		c.Next()
	}
}

// AdminMiddleware fonksiyonu, gelen isteğin admin yetkisine sahip olup olmadığını kontrol eder.
// Bu middleware, AuthMiddleware'den sonra çalıştırılmalıdır.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Context'ten role_id bilgisini al
		roleID, exists := c.Get("role_id")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Rol bilgisi bulunamadı"})
			c.Abort()
			return
		}

		// role_id float64 olarak gelebilir (JSON unmarshal)
		var role float64
		switch v := roleID.(type) {
		case float64:
			role = v
		case int:
			role = float64(v)
		default:
			c.JSON(http.StatusForbidden, gin.H{"error": "Geçersiz rol tipi"})
			c.Abort()
			return
		}

		// Varsayılan Admin Role_ID 1 kabul ediliyor. (migrations/001_create_roles_table.sql'deki admin).
		if role != 1 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bu işlem için admin yetkisi gerekmektedir"})
			c.Abort()
			return
		}

		c.Next()
	}
}
