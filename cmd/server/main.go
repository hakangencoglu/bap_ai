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

	// Veritabanı migration'ları çalıştırılır
	if err := database.RunMigrations(database.DB, "migrations"); err != nil {
		log.Fatalf("Migration hatası: %v", err)
	}

	// Repository, Service ve Handler katmanları oluşturulur (Dependency Injection)
	uyeRepo := repository.NewUyeRepository(database.DB)
	projeRepo := repository.NewProjeRepository(database.DB)
	authService := service.NewAuthService(uyeRepo)
	dashboardService := service.NewDashboardService(projeRepo)
	authHandler := api.NewAuthHandler(authService)
	dashboardHandler := api.NewDashboardHandler(dashboardService)


	// Gin router oluşturulur
	router := gin.Default()

	// Statik dosyalar ve HTML şablonları sunulur
	router.Static("/static", "./frontend/static")
	router.LoadHTMLGlob("frontend/templates/*")

	// Ana sayfa için rota (Artık giriş sayfası)
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "login.html", gin.H{})
	})

	// Kayıt olma sayfası route'u eklendi
	router.GET("/register", func(c *gin.Context) {
		c.HTML(200, "register.html", gin.H{})
	})

	// Giriş yaptıktan sonraki dashboard
	router.GET("/anasayfa", func(c *gin.Context) {
		c.HTML(200, "anasayfa.html", gin.H{})
	})

	// Yeni Başvuru sayfası route'u
	router.GET("/basvuru", func(c *gin.Context) {
		c.HTML(200, "application_form.html", gin.H{})
	})

	// Projelerim sayfası route'u (şimdilik anasayfaya yönlendirilir)
	router.GET("/projelerim", func(c *gin.Context) {
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
		// Profil endpoint'i - kullanıcının tam bilgilerini döner
		protectedRoutes.GET("/profil", func(c *gin.Context) {
			uyeIDFloat, _ := c.Get("uye_id")
			uyeID := int(uyeIDFloat.(float64))

			uye, err := uyeRepo.GetUyeByID(uyeID)
			if err != nil {
				c.JSON(401, gin.H{"error": "Kullanıcı bulunamadı"})
				return
			}

			c.JSON(200, gin.H{
				"uye_id":  uye.UyeID,
				"ad":      uye.Ad,
				"soyad":   uye.Soyad,
				"unvan":   uye.Unvan,
				"email":   uye.IletisimMail,
				"bolum":   uye.Bolum,
				"role_id": uye.RoleID,
			})
		})

		// Dashboard istatistikleri endpoint'i
		protectedRoutes.GET("/dashboard/stats", dashboardHandler.GetStats)

		// Son başvurular endpoint'i
		protectedRoutes.GET("/dashboard/recent-projects", dashboardHandler.GetRecentProjects)
	}

	// Sunucu başlatılır
	port := configs.AppConfig.Port
	log.Printf("Sunucu :%s portunda başlatılıyor...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}
