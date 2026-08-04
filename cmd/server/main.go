package main

import (
	"log"
	"time"

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
	satinalmaRepo := repository.NewSatinalmaRepository(database.DB)
	bildirimRepo := repository.NewBildirimRepository(database.DB)
	
	// Türkçe Yorum: EpostaService ilklendirilir ve ProjeRepository durum değişikliklerini dinleyecek callback'e bağlanır.
	epostaService := service.NewEpostaService(database.DB, configs.AppConfig)
	projeRepo.OnStatusChange = epostaService.SendStatusNotificationEmail

	authService := service.NewAuthService(uyeRepo)
	dashboardService := service.NewDashboardService(projeRepo)
	profilService := service.NewProfilService(uyeRepo, projeRepo)
	projeService := service.NewProjeService(projeRepo, hakemRepo, revizyonRepo)
	hakemService := service.NewHakemService(hakemRepo, projeRepo, adminRepo)
	adminService := service.NewAdminService(adminRepo, projeRepo)
	revizyonService := service.NewRevizyonService(revizyonRepo)
	pdfService := service.NewPdfService(projeRepo, adminRepo)
	davetService := service.NewDavetService(davetRepo)
	eimzaRepo := repository.NewEimzaRepository(database.DB)
	eimzaService := service.NewEimzaService(eimzaRepo)
	satinalmaService := service.NewSatinalmaService(satinalmaRepo, projeRepo)
	satinalmaService.OnPurchaseAction = epostaService.SendPurchaseNotificationEmail
	bildirimService := service.NewBildirimService(bildirimRepo)

	authHandler := api.NewAuthHandler(authService)
	dashboardHandler := api.NewDashboardHandler(dashboardService)
	profilHandler := api.NewProfilHandler(profilService, authService)
	projeHandler := api.NewProjeHandler(projeService, uyeRepo, davetRepo)
	hakemHandler := api.NewHakemHandler(hakemService)
	adminHandler := api.NewAdminHandler(adminService)
	revizyonHandler := api.NewRevizyonHandler(revizyonService)
	pdfHandler := api.NewPdfHandler(pdfService, projeRepo)
	davetHandler := api.NewDavetHandler(davetService)
	eimzaHandler := api.NewEimzaHandler(eimzaService, uyeRepo)
	satinalmaHandler := api.NewSatinalmaHandler(satinalmaService)
	bildirimHandler := api.NewBildirimHandler(bildirimService)
	chatService := service.NewChatService(configs.AppConfig.LLMProvider, configs.AppConfig.LLMEndpoint, configs.AppConfig.LLMModel, configs.AppConfig.GeminiAPIKey)
	chatHandler := api.NewChatHandler(chatService, adminService, projeRepo)

	komisyonRepo := repository.NewKomisyonRepository(database.DB)
	komisyonService := service.NewKomisyonService(komisyonRepo)
	komisyonHandler := api.NewKomisyonHandler(komisyonService, pdfService)

	// Türkçe Yorum: Talep sistemi için repository, service ve handler oluşturulur.
	talepRepo := repository.NewTalepRepository(database.DB)
	talepService := service.NewTalepService(talepRepo)
	talepHandler := api.NewTalepHandler(talepService)

	// Türkçe Yorum: Proje Sözleşmesi sistemi ve aylık e-posta hatırlatıcı servisi ilklendirilir.
	sozlesmeRepo := repository.NewSozlesmeRepository(database.DB)
	sozlesmeService := service.NewSozlesmeService(sozlesmeRepo, adminRepo, pdfService)
	sozlesmeHatirlatmaService := service.NewSozlesmeHatirlatmaService(sozlesmeRepo, epostaService)
	sozlesmeHandler := api.NewSozlesmeHandler(sozlesmeService, sozlesmeHatirlatmaService)

	// Türkçe Yorum: Sözleşme aylık e-posta hatırlatma zamanlayıcısı arka planda başlatılır (24 saatlik periyot).
	sozlesmeHatirlatmaService.StartHatirlatmaScheduler(24 * time.Hour)


	
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

	// Proje Başvuruları Modülü sayfası
	router.GET("/admin/proje-basvurulari", func(c *gin.Context) {
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

	// Komisyon Başkanı Dashboard sayfası
	router.GET("/komisyon/baskan/dashboard", func(c *gin.Context) {
		c.HTML(200, "komisyon_baskani_dashboard.html", gin.H{})
	})

	// TTO Dashboard sayfası
	router.GET("/tto/dashboard", func(c *gin.Context) {
		c.HTML(200, "tto_dashboard.html", gin.H{})
	})

	// Araştırmacı Satın Alma Talepleri sayfası
	router.GET("/satinalma", func(c *gin.Context) {
		c.HTML(200, "satinalma_arastirmaci.html", gin.H{})
	})

	// TTO Satın Alma Yönetimi sayfası (tto_dashboard içindeki bölüme yönlendirme için alias)
	router.GET("/tto/satinalma", func(c *gin.Context) {
		c.Redirect(302, "/tto/dashboard?section=satinalma")
	})

	// TTO Proje Talepleri Yönetimi sayfası (tto_dashboard içindeki talepler bölümüne yönlendirme için alias)
	router.GET("/tto/talepler", func(c *gin.Context) {
		c.Redirect(302, "/tto/dashboard?section=talepler")
	})

	// E-İmza Paneli sayfası
	router.GET("/eimza", func(c *gin.Context) {
		c.HTML(200, "eimza.html", gin.H{})
	})


	// API rotaları tanımlanır
	authRoutes := router.Group("/api/auth")
	{
		authRoutes.POST("/register", authHandler.Register) // Kayıt olma endpoint'i
		authRoutes.POST("/login", authHandler.Login)       // Giriş yapma endpoint'i
		authRoutes.POST("/set-password", authHandler.SetPassword) // İlk girişte şifre belirleme endpoint'i
	}

	// Geçici DB test endpoint'i
	router.GET("/api/test/db-status", projeHandler.GetDBStatus)

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

		// Bütçe kategorileri lookup endpoint'i
		protectedRoutes.GET("/butce-kategorileri", projeHandler.GetButceKategorileri)

		// Sistem rolleri lookup endpoint'i
		protectedRoutes.GET("/sistem-rolleri", projeHandler.GetSistemRolleri)

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
		protectedRoutes.GET("/hakem/degerlendirme-sorulari", hakemHandler.GetDegerlendirmeQuestions)

		// TTO ve Admin ortak Hakem Yönetim API'leri (Süreç içi hakem atama)
		protectedRoutes.POST("/workflow/assign-hakem", api.RequireRoles("admin", "tto"), adminHandler.AssignHakem)
		protectedRoutes.GET("/workflow/hakemler", api.RequireRoles("admin", "tto"), adminHandler.GetHakemListesi)
		protectedRoutes.GET("/workflow/project/:id/hakemler", api.RequireRoles("admin", "tto"), adminHandler.GetProjeyeAtananHakemler)

		// Onay Süreci (Workflow) API endpoint'leri
		protectedRoutes.GET("/workflow/projects", api.RequireRoles("dekan", "komisyon", "komisyon_baskani", "tto", "admin"), projeHandler.GetWorkflowProjects)
		protectedRoutes.POST("/workflow/action", api.RequireRoles("dekan", "komisyon", "komisyon_baskani", "tto", "admin"), projeHandler.HandleWorkflowAction)
		protectedRoutes.GET("/workflow/history", api.RequireRoles("dekan", "komisyon", "komisyon_baskani", "tto", "admin"), projeHandler.GetWorkflowHistory)
		// Türkçe Yorum: TTO ve Admin rollerinin projelerin durumunu doğrudan güncelleyebilmesi için endpoint tanımlandı.
		protectedRoutes.PUT("/workflow/project/status", api.RequireRoles("admin", "tto"), adminHandler.UpdateProjectStatus)
		protectedRoutes.GET("/proje/:id/surec-gecmisi", projeHandler.GetSurecGecmisi)

		// E-İmza API endpoint'leri
		protectedRoutes.GET("/eimza/pending", eimzaHandler.GetPendingSignatures)
		protectedRoutes.GET("/eimza/signed", eimzaHandler.GetSignedDocuments)
		protectedRoutes.POST("/eimza/sign", eimzaHandler.SignDocument)

		// Satın Alma API endpoint'leri
		protectedRoutes.POST("/satinalma/talep", api.RequireRoles("akademisyen", "admin"), satinalmaHandler.CreatePurchaseRequest)
		protectedRoutes.GET("/satinalma/proje/:id", satinalmaHandler.GetPurchaseRequestsByProject)
		protectedRoutes.GET("/satinalma/tum", api.RequireRoles("tto", "admin"), satinalmaHandler.GetAllPurchaseRequests)
		protectedRoutes.POST("/satinalma/onay", api.RequireRoles("tto", "admin"), satinalmaHandler.HandlePurchaseApproval)
		protectedRoutes.POST("/satinalma/revize", api.RequireRoles("tto", "admin"), satinalmaHandler.RevisePurchaseRequest)

		// Bildirim API endpoint'leri
		protectedRoutes.GET("/bildirimler", bildirimHandler.GetBildirimler)
		protectedRoutes.POST("/bildirimler/:id/oku", bildirimHandler.MarkAsRead)
		protectedRoutes.POST("/bildirimler/oku-hepsi", bildirimHandler.MarkAllAsRead)

		// Komisyon Başkanı Toplantı Yönetim API endpoint'leri
		protectedRoutes.GET("/komisyon/uyeler", api.RequireRoles("komisyon_baskani", "admin"), komisyonHandler.GetCommissionMembers)
		protectedRoutes.GET("/komisyon/toplanti/next-no", api.RequireRoles("komisyon_baskani", "admin"), komisyonHandler.GetNextMeetingNumber)
		protectedRoutes.POST("/komisyon/toplanti", api.RequireRoles("komisyon_baskani", "admin"), komisyonHandler.CreateMeeting)
		protectedRoutes.POST("/komisyon/toplanti/preview-pdf", api.RequireRoles("komisyon_baskani", "admin"), komisyonHandler.PreviewMeetingPDF)
		protectedRoutes.GET("/komisyon/toplanti/:id/pdf", api.RequireRoles("komisyon_baskani", "admin"), komisyonHandler.GetMeetingPDF)
		protectedRoutes.GET("/komisyon/toplantilar", api.RequireRoles("komisyon_baskani", "admin"), komisyonHandler.GetMeetingsList)

		// BAP Türleri endpoint'i (Başvuru dolduranlar için)
		protectedRoutes.GET("/bap-turleri", adminHandler.GetBapTurleriPublic)

		// Yapay Zeka Sohbet API endpoint'i
		protectedRoutes.POST("/chat", chatHandler.SendMessage)
		protectedRoutes.GET("/chat/status", chatHandler.GetStatus)

		// Proje Talep API endpoint'leri (Akademisyen gönderir, Admin/TTO yönetir)
		// Türkçe Yorum: :tip param ile tek handler tüm talep tiplerini karşılar.
		protectedRoutes.POST("/talep/:tip", api.RequireRoles("akademisyen", "admin"), talepHandler.SubmitTalep)
		protectedRoutes.GET("/talepler", api.RequireRoles("admin", "tto"), talepHandler.GetAllTalepler)
		protectedRoutes.POST("/talep/onay", api.RequireRoles("admin", "tto"), talepHandler.OnayTalep)
		protectedRoutes.GET("/proje/:id/talepler", talepHandler.GetTaleplerByProje)

		// Proje Sözleşmesi API endpoint'leri
		protectedRoutes.POST("/proje/:id/sozlesme", sozlesmeHandler.SaveSozlesme)
		protectedRoutes.GET("/proje/:id/sozlesme", sozlesmeHandler.GetSozlesme)
		protectedRoutes.GET("/proje/:id/sozlesme/pdf", sozlesmeHandler.DownloadSozlesmePDF)
		protectedRoutes.POST("/sozlesme/hatirlatmalari-calistir", api.RequireRoles("admin", "komisyon_baskani", "komisyon"), sozlesmeHandler.TriggerMonthlyReminders)


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

			// Admin Yetki Yönetimi endpoints
			adminRoutes.GET("/sayfa-yetkileri", adminHandler.GetSayfaYetkiMatrix)
			adminRoutes.PUT("/sayfa-yetkileri", adminHandler.UpdateSayfaYetki)
			adminRoutes.POST("/role", adminHandler.CreateRole)
			adminRoutes.PUT("/role/:id", adminHandler.UpdateRole)
			adminRoutes.DELETE("/role/:id", adminHandler.DeleteRole)
		}

		// Sayfa yetki erişim kontrol endpoint'leri
		protectedRoutes.GET("/auth/check-page-access", adminHandler.CheckPageAccess)
		protectedRoutes.GET("/auth/my-allowed-pages", adminHandler.GetMyAllowedPages)
	}

	// Sunucu başlatılır
	port := configs.AppConfig.Port
	log.Printf("Sunucu :%s portunda başlatılıyor...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}
