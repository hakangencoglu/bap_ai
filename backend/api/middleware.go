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
