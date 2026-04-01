package main

import (
	"log"

	"bap_ai/backend/api"
	"bap_ai/backend/database"
	"bap_ai/backend/repository"
	"bap_ai/backend/service"
	"bap_ai/configs"

	"github.com/gin-gonic/gin"
)

func main() {
	// Uygulama konfigürasyonu yüklenir
	configs.LoadConfig()

	// Veritabanı bağlantısı başlatılır
	database.Connect()

	// Repository, Service ve Handler katmanları oluşturulur (Dependency Injection)
	uyeRepo := repository.NewUyeRepository(database.DB)
	authService := service.NewAuthService(uyeRepo)
	authHandler := api.NewAuthHandler(authService)

	// Gin router oluşturulur
	router := gin.Default()

	// Statik dosyalar ve HTML şablonları sunulur
	router.Static("/static", "./frontend/static")
	router.LoadHTMLGlob("frontend/templates/*")

	// Ana sayfa için rota (Artık giriş sayfası)
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "login.html", gin.H{})
	})

	// Giriş yaptıktan sonraki dashboard
	router.GET("/anasayfa", func(c *gin.Context) {
		c.HTML(200, "anasayfa.html", gin.H{})
	})

	// API rotaları tanımlanır
	authRoutes := router.Group("/api/auth")
	{
		authRoutes.POST("/register", authHandler.Register) // Kayıt olma endpoint'i
		authRoutes.POST("/login", authHandler.Login)       // Giriş yapma endpoint'i
	}

	// Korumalı rotalar (JWT doğrulaması gerektirir)
	protectedRoutes := router.Group("/api")
	protectedRoutes.Use(api.AuthMiddleware())
	{
		// İleride eklenecek korumalı endpoint'ler buraya yazılacak
		protectedRoutes.GET("/profil", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"mesaj":  "Korumalı alana erişim sağlandı",
				"uye_id": c.GetFloat64("uye_id"),
				"email":  c.GetString("email"),
			})
		})
	}

	// Sunucu başlatılır
	port := configs.AppConfig.Port
	log.Printf("Sunucu :%s portunda başlatılıyor...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}
