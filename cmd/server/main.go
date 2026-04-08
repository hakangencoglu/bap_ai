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
	hakemRepo := repository.NewHakemRepository(database.DB)
	adminRepo := repository.NewAdminRepository(database.DB)
	
	authService := service.NewAuthService(uyeRepo)
	dashboardService := service.NewDashboardService(projeRepo)
	profilService := service.NewProfilService(uyeRepo, projeRepo)
	projeService := service.NewProjeService(projeRepo, hakemRepo)
	hakemService := service.NewHakemService(hakemRepo, projeRepo)
	adminService := service.NewAdminService(adminRepo, projeRepo)

	authHandler := api.NewAuthHandler(authService)
	dashboardHandler := api.NewDashboardHandler(dashboardService)
	profilHandler := api.NewProfilHandler(profilService)
	projeHandler := api.NewProjeHandler(projeService)
	hakemHandler := api.NewHakemHandler(hakemService)
	adminHandler := api.NewAdminHandler(adminService)
	
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

	// Profil sayfası route'u
	router.GET("/profil", func(c *gin.Context) {
		c.HTML(200, "profil.html", gin.H{})
	})

	// Hakem Dashboard sayfası
	router.GET("/hakem/dashboard", func(c *gin.Context) {
		c.HTML(200, "hakem_dashboard.html", gin.H{})
	})

	// Hakem Değerlendirme sayfası
	router.GET("/hakem/degerlendirme", func(c *gin.Context) {
		c.HTML(200, "hakem_degerlendirme.html", gin.H{})
	})

	// Admin Dashboard sayfası
	router.GET("/admin/dashboard", func(c *gin.Context) {
		c.HTML(200, "admin_dashboard.html", gin.H{})
	})

	// Admin Proje Durum ve Takip Sayfası
	router.GET("/admin/projects/status", func(c *gin.Context) {
		c.HTML(200, "admin_project_status.html", gin.H{})
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
		// Profil endpoint'leri - kullanıcının bilgilerini ve projelerini döner
		protectedRoutes.GET("/profil/bilgiler", profilHandler.GetProfilBilgileri)
		protectedRoutes.GET("/profil/projeler", profilHandler.GetProfilProjeleri)

		// Dashboard istatistikleri endpoint'i
		protectedRoutes.GET("/dashboard/stats", dashboardHandler.GetStats)

		// Son başvurular endpoint'i
		protectedRoutes.GET("/dashboard/recent-projects", dashboardHandler.GetRecentProjects)

		// Yeni proje başvurusu endpoint'i
		protectedRoutes.POST("/proje", projeHandler.CreateProje)

		// Hakem API endpoint'leri
		protectedRoutes.GET("/hakem/projeler", hakemHandler.GetAtananProjeler)
		protectedRoutes.POST("/hakem/degerlendir", hakemHandler.SubmitDegerlendirme)

		// Admin API endpoint'leri
		adminRoutes := protectedRoutes.Group("/admin")
		adminRoutes.Use(api.AdminMiddleware())
		{
			adminRoutes.GET("/stats", adminHandler.GetStats)
			adminRoutes.GET("/users", adminHandler.GetAllUsers)
			adminRoutes.GET("/projects", adminHandler.GetAllProjects)
			adminRoutes.PUT("/user/role", adminHandler.UpdateUserRole)
			adminRoutes.PUT("/user/status", adminHandler.UpdateUserStatus)
			adminRoutes.PUT("/project/status", adminHandler.UpdateProjectStatus)
			adminRoutes.GET("/project/:id/details", adminHandler.GetProjectDetails)
		}
	}

	// Sunucu başlatılır
	port := configs.AppConfig.Port
	log.Printf("Sunucu :%s portunda başlatılıyor...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}
