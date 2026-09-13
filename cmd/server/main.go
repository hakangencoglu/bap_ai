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
	satinalmaService.PageAccess = adminService
	bildirimService := service.NewBildirimService(bildirimRepo)

	// Türkçe Yorum: Geri bildirim modülü katmanları ilklendirilir.
	feedbackRepo := repository.NewFeedbackRepository(database.DB)
	feedbackService := service.NewFeedbackService(feedbackRepo, epostaService)
	feedbackHandler := api.NewFeedbackHandler(feedbackService, adminService)

	auditRepo := repository.NewAuditRepository(database.DB)

	authHandler := api.NewAuthHandler(authService)
	dashboardHandler := api.NewDashboardHandler(dashboardService)
	profilHandler := api.NewProfilHandler(profilService, authService)
	projeHandler := api.NewProjeHandler(projeService, uyeRepo, davetRepo, epostaService)
	hakemHandler := api.NewHakemHandler(hakemService)
	adminHandler := api.NewAdminHandler(adminService, auditRepo)
	revizyonHandler := api.NewRevizyonHandler(revizyonService)
	pdfHandler := api.NewPdfHandler(pdfService, projeRepo)
	davetHandler := api.NewDavetHandler(davetService)
	eimzaHandler := api.NewEimzaHandler(eimzaService, uyeRepo)
	satinalmaHandler := api.NewSatinalmaHandler(satinalmaService)
	bildirimHandler := api.NewBildirimHandler(bildirimService)
	// Türkçe Yorum: RAG (Vektör & Full-Text Search) altyapısı ilklendirilir.
	ragRepo := repository.NewRAGRepository(database.DB)
	embeddingService := service.NewEmbeddingService(configs.AppConfig.LLMProvider, configs.AppConfig.LLMEndpoint)
	ragService := service.NewRAGService(ragRepo, embeddingService)

	// Türkçe Yorum: Sunucu başladığında veritabanındaki tüm projeler ve detaylar otomatik olarak RAG indeksine taşınır.
	go func() {
		if count, err := ragService.SyncDatabase(); err != nil {
			log.Printf("RAG Başlangıç İndeksleme Uyarısı: %v", err)
		} else {
			log.Printf("RAG Başlangıç İndeksleme: Veritabanından %d adet kayıt RAG indeksine aktarıldı.", count)
		}
	}()

	chatService := service.NewChatService(configs.AppConfig.LLMProvider, configs.AppConfig.LLMEndpoint, configs.AppConfig.LLMModel, configs.AppConfig.GeminiAPIKey)
	chatHandler := api.NewChatHandler(chatService, adminService, projeRepo, ragService)

	komisyonRepo := repository.NewKomisyonRepository(database.DB)
	komisyonService := service.NewKomisyonService(komisyonRepo)

	// Türkçe Yorum: Talep sistemi için repository, service ve handler oluşturulur.
	talepRepo := repository.NewTalepRepository(database.DB)
	if n, err := talepRepo.ReconcileApprovedFasilAktarimlari(); err != nil {
		log.Printf("Fasıl aktarım mutabakat uyarısı: %v", err)
	} else if n > 0 {
		log.Printf("Fasıl aktarım mutabakat: %d onaylı talep bütçeye uygulandı.", n)
	}
	degisiklikRepo := repository.NewDegisiklikRepository(database.DB)
	degisiklikService := service.NewDegisiklikService(degisiklikRepo)
	degisiklikHandler := api.NewDegisiklikHandler(degisiklikService)
	talepService := service.NewTalepService(talepRepo, degisiklikRepo)
	satinalmaService.Audit = degisiklikRepo
	talepHandler := api.NewTalepHandler(talepService)

	// Türkçe Yorum: Toplantı ↔ proje köprü tablo katmanı ayrı modül olarak ilklendirilir.
	komisyonToplantiRepo := repository.NewKomisyonToplantiRepository(database.DB)
	komisyonToplantiService := service.NewKomisyonToplantiService(komisyonToplantiRepo, projeService, talepService)
	komisyonToplantiHandler := api.NewKomisyonToplantiHandler(komisyonToplantiService, pdfService)
	komisyonHandler := api.NewKomisyonHandler(komisyonService, komisyonToplantiService, pdfService)

	// Türkçe Yorum: Proje Sözleşmesi sistemi ve aylık e-posta hatırlatıcı servisi ilklendirilir.
	sozlesmeRepo := repository.NewSozlesmeRepository(database.DB)
	sozlesmeService := service.NewSozlesmeService(sozlesmeRepo, adminRepo, pdfService, epostaService)
	sozlesmeHatirlatmaService := service.NewSozlesmeHatirlatmaService(sozlesmeRepo, epostaService)
	sozlesmeHandler := api.NewSozlesmeHandler(sozlesmeService, sozlesmeHatirlatmaService)

	// Türkçe Yorum: Sözleşme aylık e-posta hatırlatma zamanlayıcısı arka planda başlatılır (24 saatlik periyot).
	sozlesmeHatirlatmaService.StartHatirlatmaScheduler(24 * time.Hour)

	// Türkçe Yorum: Ara Rapor (Gelişme Raporu) ve Kesin Sonuç Raporu katmanları ilklendirilir.
	raporRepo := repository.NewRaporRepository(database.DB)
	raporService := service.NewRaporService(raporRepo, projeRepo, epostaService)
	raporHandler := api.NewRaporHandler(raporService)

	// Türkçe Yorum: Zamanlanmış Görev & Otomatik Bildirim (E-Posta ve SMS) Mimarisi ilklendirilir.
	zamanlanmisGorevRepo := repository.NewZamanlanmisGorevRepository(database.DB)
	smsService := service.NewSmsService(database.DB, configs.AppConfig)
	zamanlanmisGorevService := service.NewZamanlanmisGorevService(zamanlanmisGorevRepo, epostaService, smsService)
	zamanlanmisGorevScheduler := service.NewZamanlanmisGorevScheduler(zamanlanmisGorevService)
	zamanlanmisGorevHandler := api.NewZamanlanmisGorevHandler(zamanlanmisGorevService)

	// Zamanlanmış görev motoru 12 saatlik periyotlarla arka planda çalışması için başlatılır
	zamanlanmisGorevScheduler.StartScheduler(12 * time.Hour)
	
	// Gin router oluşturulur
	router := gin.Default()
	router.Use(api.AuditLogMiddleware(auditRepo))

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

	// Admin Zamanlanmış Görevler Sayfası
	router.GET("/admin/zamanlanmis-gorevler", func(c *gin.Context) {
		c.HTML(200, "admin_zamanlanmis_gorevler.html", gin.H{})
	})

	// Admin Teslim Edilen Raporlar & Takip Modülü Sayfası
	router.GET("/admin/raporlar", func(c *gin.Context) {
		c.HTML(200, "admin_raporlar.html", gin.H{})
	})
	router.GET("/tto/raporlar", func(c *gin.Context) {
		c.HTML(200, "admin_raporlar.html", gin.H{})
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

		// Geri bildirim modülü endpoint'i
		protectedRoutes.POST("/feedback", feedbackHandler.SubmitFeedback)

		// Dashboard istatistikleri endpoint'i
		protectedRoutes.GET("/dashboard/stats", dashboardHandler.GetStats)

		// Son başvurular endpoint'i
		protectedRoutes.GET("/dashboard/recent-projects", dashboardHandler.GetRecentProjects)
		// Tüm başvurular endpoint'i
		protectedRoutes.GET("/dashboard/all-projects", dashboardHandler.GetAllProjects)

		// Yeni proje başvurusu endpoint'i
		protectedRoutes.POST("/proje", projeHandler.CreateProje)
		protectedRoutes.POST("/proje/upload-ek-dosya", projeHandler.UploadEkDosya)

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
		protectedRoutes.DELETE("/proje/:id/takim/:uye_id", projeHandler.RemoveTeamMember)
		protectedRoutes.POST("/proje/:id/takim/:uye_id/belge", projeHandler.UploadTakimBelgesi)
		protectedRoutes.GET("/proje/:id/takim-belgeler", projeHandler.GetTakimBelgeleri)

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
		protectedRoutes.GET("/workflow/projects", api.RequireRoles("dekan", "komisyon", "komisyon_baskani", "komisyon_raportoru", "tto", "admin"), projeHandler.GetWorkflowProjects)
		protectedRoutes.POST("/workflow/action", api.RequireRoles("dekan", "komisyon", "komisyon_baskani", "komisyon_raportoru", "tto", "admin"), projeHandler.HandleWorkflowAction)
		protectedRoutes.GET("/workflow/history", api.RequireRoles("dekan", "komisyon", "komisyon_baskani", "komisyon_raportoru", "tto", "admin"), projeHandler.GetWorkflowHistory)
		// Türkçe Yorum: TTO ve Admin rollerinin projelerin durumunu doğrudan güncelleyebilmesi için endpoint tanımlandı.
		protectedRoutes.PUT("/workflow/project/status", api.RequireRoles("admin", "tto"), adminHandler.UpdateProjectStatus)
		protectedRoutes.GET("/proje/:id/surec-gecmisi", projeHandler.GetSurecGecmisi)

		// E-İmza API endpoint'leri
		protectedRoutes.GET("/eimza/pending", eimzaHandler.GetPendingSignatures)
		protectedRoutes.GET("/eimza/signed", eimzaHandler.GetSignedDocuments)
		protectedRoutes.POST("/eimza/sign", eimzaHandler.SignDocument)

		// Satın Alma API endpoint'leri
		protectedRoutes.POST("/satinalma/talep", api.RequireRoles("akademisyen", "admin"), satinalmaHandler.CreatePurchaseRequest)
		protectedRoutes.DELETE("/satinalma/talep/:id", api.SuperDeleteMiddleware(), satinalmaHandler.DeletePurchaseRequest)
		protectedRoutes.GET("/satinalma/proje/:id", satinalmaHandler.GetPurchaseRequestsByProject)
		protectedRoutes.GET("/satinalma/proje/:id/butce-raporu", api.RequireRoles("tto", "admin", "akademisyen"), satinalmaHandler.GetProjectBudgetReport)
		protectedRoutes.GET("/satinalma/proje/:id/odemeler", api.RequireRoles("tto", "admin", "akademisyen"), satinalmaHandler.ListOdemelerByProje)
		protectedRoutes.GET("/satinalma/tum", api.RequireRoles("tto", "admin"), satinalmaHandler.GetAllPurchaseRequests)
		protectedRoutes.POST("/satinalma/onay", api.RequireRoles("tto", "admin"), satinalmaHandler.HandlePurchaseApproval)
		protectedRoutes.POST("/satinalma/revize", api.RequireRoles("tto", "admin"), satinalmaHandler.RevisePurchaseRequest)
		protectedRoutes.POST("/satinalma/mutabakat", api.RequireRoles("tto", "admin"), satinalmaHandler.HandleMutabakat)
		protectedRoutes.GET("/satinalma/mutabakat/bekleyen", api.RequireRoles("tto", "admin"), satinalmaHandler.ListPendingMutabakat)

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
		protectedRoutes.GET("/komisyon/toplantilar", api.RequireRoles("komisyon_baskani", "komisyon_raportoru", "admin"), komisyonHandler.GetMeetingsList)
		// Türkçe Yorum: Toplantı silme yalnızca admin1@izu.edu.tr (e-posta; rol ile verilemez).
		protectedRoutes.DELETE("/komisyon/toplanti/:id", api.SuperDeleteMiddleware(), komisyonHandler.DeleteMeeting)

		// Komisyon Toplantı ↔ Proje Köprü Tablo endpoint'leri
		// Türkçe Yorum: Hangi projenin hangi toplantıda görüşüldüğünü yöneten route'lar.
		protectedRoutes.GET("/komisyon/bekleyen-projeler", api.RequireRoles("komisyon_baskani", "komisyon_raportoru", "admin"), komisyonToplantiHandler.GetBekleyenProjeler)
		protectedRoutes.POST("/komisyon/toplanti/:id/projeler", api.RequireRoles("komisyon_baskani", "komisyon_raportoru", "admin"), komisyonToplantiHandler.AddProjeToToplanti)
		protectedRoutes.DELETE("/komisyon/toplanti/:id/projeler/:proje_id", api.RequireRoles("komisyon_baskani", "komisyon_raportoru", "admin"), komisyonToplantiHandler.RemoveProjeFromToplanti)
		protectedRoutes.GET("/komisyon/toplanti/:id/projeler", api.RequireRoles("komisyon_baskani", "komisyon_raportoru", "komisyon", "admin"), komisyonToplantiHandler.GetProjectsByToplanti)
		protectedRoutes.PUT("/komisyon/toplanti/:id/projeler/:proje_id/karar", api.RequireRoles("komisyon_baskani", "komisyon_raportoru", "admin"), komisyonToplantiHandler.SetProjeKarar)
		protectedRoutes.GET("/komisyon/toplanti/:id/tutanak", api.RequireRoles("komisyon_baskani", "komisyon_raportoru", "admin"), komisyonToplantiHandler.GetToplantiBelgePDF)

		// BAP Türleri endpoint'i (Başvuru dolduranlar için)
		protectedRoutes.GET("/bap-turleri", adminHandler.GetBapTurleriPublic)
		protectedRoutes.GET("/bap-turu/:id/form-alanlari", adminHandler.GetBapTuruFormAlanlari)
		protectedRoutes.GET("/proje/:id/dinamik-alanlar", projeHandler.GetProjeDinamikAlanlar)

		// Yapay Zeka Sohbet API endpoint'i
		protectedRoutes.POST("/chat", chatHandler.SendMessage)
		protectedRoutes.GET("/chat/status", chatHandler.GetStatus)
		protectedRoutes.GET("/chat/history", chatHandler.GetHistory)
		protectedRoutes.DELETE("/chat/history", chatHandler.ClearHistory)
		protectedRoutes.POST("/chat/sync-rag", api.RequireRoles("admin"), chatHandler.SyncRAG)

		// Proje Talep API endpoint'leri (Akademisyen gönderir, Admin/TTO yönetir)
		// Türkçe Yorum: :tip param ile tek handler tüm talep tiplerini karşılar.
		protectedRoutes.POST("/talep/:tip", api.RequireRoles("akademisyen", "admin"), talepHandler.SubmitTalep)
		protectedRoutes.GET("/talepler", talepHandler.GetAllTalepler)
		protectedRoutes.POST("/talep/onay", api.RequireRoles("admin", "tto"), talepHandler.OnayTalep)
		protectedRoutes.GET("/proje/:id/talepler", talepHandler.GetTaleplerByProje)
		protectedRoutes.GET("/proje/:id/degisiklikler", degisiklikHandler.ListDegisiklikler)
		protectedRoutes.GET("/proje/:id/degisiklikler/:degisiklik_id", degisiklikHandler.GetDegisiklik)

		// Proje Sözleşmesi API endpoint'leri
		protectedRoutes.POST("/proje/:id/sozlesme", sozlesmeHandler.SaveSozlesme)
		protectedRoutes.GET("/proje/:id/sozlesme", sozlesmeHandler.GetSozlesme)
		protectedRoutes.GET("/proje/:id/sozlesme/pdf", sozlesmeHandler.DownloadSozlesmePDF)
		protectedRoutes.POST("/sozlesme/hatirlatmalari-calistir", api.RequireRoles("admin", "komisyon_baskani", "komisyon"), sozlesmeHandler.TriggerMonthlyReminders)

		// Ara Rapor (Gelişme Raporu) ve Kesin Sonuç Raporu API endpoint'leri
		protectedRoutes.POST("/proje/:id/ara-rapor", raporHandler.SubmitAraRapor)
		protectedRoutes.POST("/proje/:id/ara-rapor/upload", raporHandler.UploadRaporDosya)
		protectedRoutes.GET("/proje/:id/ara-raporlar", raporHandler.GetProjeRaporlari)
		protectedRoutes.GET("/admin/ara-raporlar/bekleyen", api.RequireRoles("admin", "tto", "komisyon_baskani", "komisyon"), raporHandler.GetBekleyenRaporlar)
		protectedRoutes.GET("/admin/ara-raporlar/takip-matrisi", api.RequireRoles("admin", "tto", "komisyon_baskani", "komisyon"), raporHandler.GetTTORaporTakipMatrisi)
		protectedRoutes.POST("/admin/ara-rapor/degerlendir", api.RequireRoles("admin", "tto", "komisyon_baskani", "komisyon"), raporHandler.DegerlendirRapor)


		// Admin API endpoint'leri
		adminRoutes := protectedRoutes.Group("/admin")
		adminRoutes.Use(api.AdminMiddleware())
		{
			adminRoutes.GET("/stats", adminHandler.GetStats)
			adminRoutes.GET("/audit-logs", adminHandler.GetAuditLogs)
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
			adminRoutes.PUT("/bap-turu/:id/aktiflik", adminHandler.UpdateBapTuruAktiflik)
			adminRoutes.POST("/bap-turu/:id/yayinla", adminHandler.PublishBapTuru)
			adminRoutes.GET("/bap-turu/:id/form-alanlari", adminHandler.GetBapTuruFormAlanlari)
			adminRoutes.POST("/bap-turu/:id/form-alanlari", adminHandler.SaveBapTuruFormAlanlari)
			adminRoutes.DELETE("/bap-turu/:id/form-alanlari/:alan_id", adminHandler.DeleteBapTuruFormAlani)

			// Admin Yetki Yönetimi endpoints
			adminRoutes.GET("/sayfa-yetkileri", adminHandler.GetSayfaYetkiMatrix)
			adminRoutes.PUT("/sayfa-yetkileri", adminHandler.UpdateSayfaYetki)
			adminRoutes.POST("/role", adminHandler.CreateRole)
			adminRoutes.PUT("/role/:id", adminHandler.UpdateRole)

			// Admin Süreç Aşamaları endpoints
			adminRoutes.GET("/surec-asamalari", adminHandler.GetProjeAsamalari)
			adminRoutes.POST("/surec-asamasi", adminHandler.CreateProjeAsamasi)
			adminRoutes.PUT("/surec-asamasi/:id", adminHandler.UpdateProjeAsamasi)

			// Admin Zamanlanmış Görev & Otomatik Bildirim (E-Posta & SMS) endpoints
			adminRoutes.GET("/zamanlanmis-gorev/kurallar", zamanlanmisGorevHandler.GetAllRules)
			adminRoutes.POST("/zamanlanmis-gorev/kural", zamanlanmisGorevHandler.CreateRule)
			adminRoutes.PUT("/zamanlanmis-gorev/kural/:id", zamanlanmisGorevHandler.UpdateRule)
			adminRoutes.DELETE("/zamanlanmis-gorev/kural/:id", zamanlanmisGorevHandler.DeleteRule)
			adminRoutes.POST("/zamanlanmis-gorev/kurallar/toplu-sil", zamanlanmisGorevHandler.DeleteRulesBulk)
			adminRoutes.POST("/zamanlanmis-gorev/calistir", zamanlanmisGorevHandler.TriggerManually)
			adminRoutes.GET("/zamanlanmis-gorev/loglar", zamanlanmisGorevHandler.GetLogs)

			// Türkçe Yorum: Kalıcı silme yalnızca admin1@izu.edu.tr — rol ile verilemez.
			adminRoutes.GET("/super-delete-yetki", adminHandler.GetSuperDeleteYetki)
			superDelete := adminRoutes.Group("")
			superDelete.Use(api.SuperDeleteMiddleware())
			{
				superDelete.DELETE("/role/:id", adminHandler.DeleteRole)
				superDelete.DELETE("/surec-asamasi/:id", adminHandler.DeleteProjeAsamasi)
				superDelete.DELETE("/user/:id", adminHandler.DeleteUser)
				superDelete.DELETE("/project/:id", adminHandler.DeleteProject)
				superDelete.DELETE("/bap-turu/:id", adminHandler.DeleteBapTuru)
			}
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
