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

	// Veritabanı şeması çalıştırılır
	if err := database.RunSchema(database.DB, "backend/database/schema.sql"); err != nil {
		log.Fatalf("Şema yükleme hatası: %v", err)
	}

	// Repository, Service ve Handler katmanları oluşturulur (Dependency Injection)
	uyeRepo := repository.NewUyeRepository(database.DB)
	projeRepo := repository.NewProjeRepository(database.DB)
	hakemRepo := repository.NewHakemRepository(database.DB)
	adminRepo := repository.NewAdminRepository(database.DB)
	revizyonRepo := repository.NewRevizyonRepository(database.DB)
	davetRepo := repository.NewDavetRepository(database.DB)
	
	authService := service.NewAuthService(uyeRepo)
	dashboardService := service.NewDashboardService(projeRepo)
	profilService := service.NewProfilService(uyeRepo, projeRepo)
	projeService := service.NewProjeService(projeRepo, hakemRepo, revizyonRepo)
	hakemService := service.NewHakemService(hakemRepo, projeRepo, adminRepo)
	adminService := service.NewAdminService(adminRepo, projeRepo)
	revizyonService := service.NewRevizyonService(revizyonRepo)
	pdfService := service.NewPdfService(projeRepo, adminRepo)
	davetService := service.NewDavetService(davetRepo)

	authHandler := api.NewAuthHandler(authService)
	dashboardHandler := api.NewDashboardHandler(dashboardService)
	profilHandler := api.NewProfilHandler(profilService, authService)
	projeHandler := api.NewProjeHandler(projeService, uyeRepo, davetRepo)
	hakemHandler := api.NewHakemHandler(hakemService)
	adminHandler := api.NewAdminHandler(adminService)
	revizyonHandler := api.NewRevizyonHandler(revizyonService)
	pdfHandler := api.NewPdfHandler(pdfService, projeRepo)
	davetHandler := api.NewDavetHandler(davetService)
	
	// Gin router oluşturulur
	router := gin.Default()

	// Statik dosyalar ve HTML şablonları sunulur
	router.Static("/static", "./frontend/static")
	router.Static("/uploads", "./uploads")
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

	// Profil tamamlama sayfası route'u (giriş sonrası zorunlu profil tamamlama)
	router.GET("/profil-tamamla", func(c *gin.Context) {
		c.HTML(200, "profil_tamamla.html", gin.H{})
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

	// Admin Hakem Atama Sayfası
	router.GET("/admin/hakem-atama", func(c *gin.Context) {
		c.HTML(200, "admin_hakem_atama.html", gin.H{})
	})

	// Dekan Dashboard sayfası
	router.GET("/dekan/dashboard", func(c *gin.Context) {
		c.HTML(200, "dekan_dashboard.html", gin.H{})
	})

	// Komisyon Dashboard sayfası
	router.GET("/komisyon/dashboard", func(c *gin.Context) {
		c.HTML(200, "komisyon_dashboard.html", gin.H{})
	})

	// TTO Dashboard sayfası
	router.GET("/tto/dashboard", func(c *gin.Context) {
		c.HTML(200, "tto_dashboard.html", gin.H{})
	})


	// API rotaları tanımlanır
	authRoutes := router.Group("/api/auth")
	{
		authRoutes.POST("/register", authHandler.Register) // Kayıt olma endpoint'i
		authRoutes.POST("/login", authHandler.Login)       // Giriş yapma endpoint'i
		authRoutes.POST("/set-password", authHandler.SetPassword) // İlk girişte şifre belirleme endpoint'i
	}

	// Korumalı rotalar (JWT doğrulaması gerektirir)
	protectedRoutes := router.Group("/api")
	protectedRoutes.Use(api.AuthMiddleware())
	{
		// Profil endpoint'leri - profil zorunlu middleware'den muaf
		protectedRoutes.GET("/profil/bilgiler", profilHandler.GetProfilBilgileri)
		protectedRoutes.POST("/profil/tamamla", profilHandler.TamamlaProfil)

		// Profil zorunlu middleware'i eklenir (profil tamamlanmadan diğer endpointlere erişim engellenir)
		protectedRoutes.Use(api.ProfilZorunluMiddleware())

		// Profil projeleri endpoint'i
		protectedRoutes.GET("/profil/projeler", profilHandler.GetProfilProjeleri)

		// Dashboard istatistikleri endpoint'i
		protectedRoutes.GET("/dashboard/stats", dashboardHandler.GetStats)

		// Son başvurular endpoint'i
		protectedRoutes.GET("/dashboard/recent-projects", dashboardHandler.GetRecentProjects)
		// Tüm başvurular endpoint'i
		protectedRoutes.GET("/dashboard/all-projects", dashboardHandler.GetAllProjects)

		// Yeni proje başvurusu endpoint'i
		protectedRoutes.POST("/proje", projeHandler.CreateProje)

		// Akademisyen listesi endpoint'i (Yürütücü seçimi için)
		protectedRoutes.GET("/akademisyenler", projeHandler.GetAkademisyenler)

		// Projeni getirme, güncelleme ve silme
		protectedRoutes.GET("/proje/:id", projeHandler.GetProje)
		protectedRoutes.GET("/proje/:id/detaylar", projeHandler.GetProjeDetaylar)
		protectedRoutes.PUT("/proje/:id", projeHandler.UpdateProje)
		protectedRoutes.DELETE("/proje/:id", projeHandler.DeleteTaslakProje)

		// Proje ek verileri (iş paketleri, bütçe, detaylar) kaydetme endpoint'i
		protectedRoutes.POST("/proje/:id/extras", projeHandler.SaveProjectExtras)

		// Proje üyeleri
		protectedRoutes.GET("/proje/:id/uyeler", projeHandler.GetUyeler)
		protectedRoutes.POST("/proje/:id/takim", projeHandler.AddTeamMember)

		// Kullanıcı arama (ekip üyesi ekleme için)
		protectedRoutes.GET("/uyeler/ara", projeHandler.SearchUyeler)

		// PDF oluşturma ve onaylama endpoint'leri
		protectedRoutes.GET("/proje/:id/pdf", pdfHandler.GeneratePDF)
		protectedRoutes.POST("/proje/:id/finalize", pdfHandler.FinalizePDF)

		// Davet endpoint'leri
		protectedRoutes.GET("/davetler", davetHandler.GetBekleyenDavetler)
		protectedRoutes.POST("/davet/yanit", davetHandler.RespondDavet)

		// Revizyon oluşturma ve getirme
		protectedRoutes.POST("/revizyon", api.RequireRoles("admin", "akademisyen", "hakem"), revizyonHandler.CreateRevizyon)
		protectedRoutes.GET("/proje/:id/revizyon", revizyonHandler.GetAktifRevizyon)

		// Hakem API endpoint'leri
		protectedRoutes.GET("/hakem/projeler", hakemHandler.GetAtananProjeler)
		protectedRoutes.POST("/hakem/degerlendir", hakemHandler.SubmitDegerlendirme)
		protectedRoutes.POST("/hakem/karar", hakemHandler.KabulRedKarar)
		protectedRoutes.GET("/hakem/proje/:id", hakemHandler.GetProjeDetay)

		// Onay Süreci (Workflow) API endpoint'leri
		protectedRoutes.GET("/workflow/projects", api.RequireRoles("dekan", "komisyon", "tto", "admin"), projeHandler.GetWorkflowProjects)
		protectedRoutes.POST("/workflow/action", api.RequireRoles("dekan", "komisyon", "tto", "admin"), projeHandler.HandleWorkflowAction)
		protectedRoutes.GET("/proje/:id/surec-gecmisi", projeHandler.GetSurecGecmisi)


		// BAP Türleri endpoint'i (Başvuru dolduranlar için)
		protectedRoutes.GET("/bap-turleri", adminHandler.GetBapTurleriPublic)

		// Admin API endpoint'leri
		adminRoutes := protectedRoutes.Group("/admin")
		adminRoutes.Use(api.AdminMiddleware())
		{
			adminRoutes.GET("/stats", adminHandler.GetStats)
			adminRoutes.GET("/users", adminHandler.GetAllUsers)
			adminRoutes.GET("/projects", adminHandler.GetAllProjects)
			adminRoutes.PUT("/user/role", adminHandler.UpdateUserRole)
			adminRoutes.PUT("/user/status", adminHandler.UpdateUserStatus)
			adminRoutes.POST("/user", adminHandler.CreateUser)
			adminRoutes.POST("/users/bulk", adminHandler.BulkCreateUsers) // Türkçe Yorum: Toplu kullanıcı ekleme endpoint'i
			adminRoutes.PUT("/user/:id", adminHandler.UpdateUser)
			adminRoutes.PUT("/project/status", adminHandler.UpdateProjectStatus)
			adminRoutes.GET("/project/:id/details", adminHandler.GetProjectDetails)
			adminRoutes.POST("/hakem-ata", adminHandler.AssignHakem)
			adminRoutes.GET("/projeler/degerlendirme-bekleyen", adminHandler.GetDegerlendirilmemisProjeleri)
			adminRoutes.GET("/hakemler", adminHandler.GetHakemListesi)
			adminRoutes.GET("/project/:id/hakemler", adminHandler.GetProjeyeAtananHakemler)
			
			// Admin BAP Türü Tanımlama endpoints
			adminRoutes.GET("/bap-turleri", adminHandler.GetBapTurleri)
			adminRoutes.POST("/bap-turu", adminHandler.CreateBapTuru)
			adminRoutes.PUT("/bap-turu/:id", adminHandler.UpdateBapTuru)
		}
	}

	// Sunucu başlatılır
	port := configs.AppConfig.Port
	log.Printf("Sunucu :%s portunda başlatılıyor...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}
