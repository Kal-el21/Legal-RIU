package main

import (
	"log"

	"legal-riu-portal/internal/config"
	"legal-riu-portal/internal/handler"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/seed"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	db := config.InitDatabase(cfg)
	store := storage.InitMinIO(cfg)

	if err := seed.RunAllMigrationsAndSeeds(db); err != nil {
		log.Fatalf("Migration and seed failed: %v", err)
	}

	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	loRepo := repository.NewLegalOpinionRepository(db)
	drRepo := repository.NewDocumentReviewRepository(db)
	dashRepo := repository.NewDashboardRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	legalCaseRepo := repository.NewLegalCaseRepository(db)
	divisionRepo := repository.NewDivisionRepository(db)
	notificationSettingRepo := repository.NewNotificationSettingRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	purposeTypeRepo := repository.NewPurposeTypeRepository(db)
	caseTypeRepo := repository.NewCaseTypeRepository(db)
	caseCategoryRepo := repository.NewCaseCategoryRepository(db)
	materialRepo := repository.NewLegalMaterialRepository(db)

	// ── Services ─────────────────────────────────────────────────────────────
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, cfg)
	userSvc := service.NewUserService(userRepo)
	loSvc := service.NewLegalOpinionService(loRepo, store)
	drSvc := service.NewDocumentReviewService(drRepo, store)
	notificationSettingSvc := service.NewNotificationSettingService(notificationSettingRepo, dashRepo)
	dashSvc := service.NewDashboardService(dashRepo, notificationSettingSvc)
	auditLogSvc := service.NewAuditLogService(auditLogRepo)
	legalCaseSvc := service.NewLegalCaseService(legalCaseRepo, store)
	divisionSvc := service.NewDivisionService(divisionRepo)
	companySvc := service.NewCompanyService(companyRepo)
	purposeTypeSvc := service.NewPurposeTypeService(purposeTypeRepo)
	caseTypeSvc := service.NewCaseTypeService(caseTypeRepo)
	caseCategorySvc := service.NewCaseCategoryService(caseCategoryRepo)
	materialSvc := service.NewLegalMaterialService(materialRepo)

	// ── Handlers ─────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authSvc, cfg, auditLogSvc)
	userHandler := handler.NewUserHandler(userSvc, auditLogSvc)
	loHandler := handler.NewLegalOpinionHandler(loSvc, auditLogSvc)
	drHandler := handler.NewDocumentReviewHandler(drSvc, auditLogSvc)
	dashHandler := handler.NewDashboardHandler(dashSvc)
	auditLogHandler := handler.NewAuditLogHandler(auditLogSvc, auditLogRepo)
	legalCaseHandler := handler.NewLegalCaseHandler(legalCaseSvc, auditLogSvc, userRepo)
	divisionHandler := handler.NewDivisionHandler(divisionSvc)
	notificationSettingHandler := handler.NewNotificationSettingHandler(notificationSettingSvc)
	companyHandler := handler.NewCompanyHandler(companySvc)
	purposeTypeHandler := handler.NewPurposeTypeHandler(purposeTypeSvc)
	caseTypeHandler := handler.NewCaseTypeHandler(caseTypeSvc)
	caseCategoryHandler := handler.NewCaseCategoryHandler(caseCategorySvc)
	materialHandler := handler.NewLegalMaterialHandler(materialSvc)

	// ── Gin ──────────────────────────────────────────────────────────────────
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.SecurityHeaders())
	r.MaxMultipartMemory = 110 << 20 // 110 MB

	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Security.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length", "X-CSRF-Token"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "Legal RIU Portal API"})
	})

	api := r.Group("/api/v1")
	loginLimiter := middleware.NewLoginRateLimiter(cfg.Security.LoginRateLimit, cfg.Security.LoginRateWindowMinutes)

	// ── Public auth ───────────────────────────────────────────────────────────
	api.POST("/auth/login", loginLimiter.Middleware(), authHandler.Login)
	api.POST("/auth/refresh", loginLimiter.Middleware(), authHandler.RefreshToken)
	api.POST("/auth/logout", authHandler.Logout)

	// ── Protected ─────────────────────────────────────────────────────────────
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	protected.GET("/auth/me", authHandler.Me)
	protected.POST("/auth/change-password", authHandler.ChangePassword)

	// Settings
	protected.PUT("/settings/profile", authHandler.UpdateProfile)
	protected.PUT("/settings/notifications", authHandler.UpdateNotification)
	protected.PUT("/settings/two-fa", authHandler.Toggle2FA)

	// Dashboard
	protected.GET("/divisions", divisionHandler.GetAll)

	userOnly := protected.Group("")
	userOnly.Use(middleware.RoleMiddleware("USER"))

	userOnly.GET("/dashboard/stats", dashHandler.UserStats)
	userOnly.GET("/dashboard/recent", dashHandler.UserRecent)
	userOnly.GET("/dashboard/reminders", dashHandler.GetReminders)
	userOnly.PATCH("/dashboard/reminders/read", notificationSettingHandler.MarkReminderRead)
	userOnly.PATCH("/dashboard/reminders/read-all", notificationSettingHandler.MarkAllRemindersRead)

	// Legal opinions — presign
	userOnly.GET("/legal-opinions/presign", loHandler.GetPresignedURL)
	userOnly.GET("/legal-opinions/download", loHandler.Download)
	userOnly.GET("/legal-opinions/:id/pdf", loHandler.GeneratePDF)
	userOnly.GET("/legal-opinions", loHandler.GetAll)
	userOnly.POST("/legal-opinions", loHandler.Create)
	userOnly.GET("/legal-opinions/:id", loHandler.GetByID)
	userOnly.PUT("/legal-opinions/:id", loHandler.Update)
	userOnly.DELETE("/legal-opinions/:id", loHandler.Delete)
	userOnly.POST("/legal-opinions/:id/resubmit", loHandler.Resubmit)

	// Review documents
	userOnly.GET("/review-documents/presign", drHandler.GetPresignedURL)
	userOnly.GET("/review-documents/download", drHandler.Download)
	userOnly.GET("/review-documents", drHandler.GetAll)
	userOnly.POST("/review-documents", drHandler.Create)
	userOnly.GET("/review-documents/:id", drHandler.GetByID)
	userOnly.PUT("/review-documents/:id", drHandler.Update)
	userOnly.DELETE("/review-documents/:id", drHandler.Delete)
	userOnly.POST("/review-documents/:id/resubmit", drHandler.Resubmit)

	// ── Admin only ─────────────────────────────────────────────────────────────
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(cfg), middleware.RoleMiddleware("ADMIN"), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	admin.GET("/dashboard/stats", dashHandler.AdminStats)
	admin.GET("/dashboard/recent", dashHandler.AdminRecent)
	admin.GET("/dashboard/reminders", notificationSettingHandler.GetRemindersDashboard)
	admin.PATCH("/dashboard/reminders/read", notificationSettingHandler.MarkReminderRead)
	admin.PATCH("/dashboard/reminders/read-all", notificationSettingHandler.MarkAllRemindersRead)

	admin.GET("/legal-opinions/presign", loHandler.GetPresignedURL)
	admin.GET("/legal-opinions/download", loHandler.Download)
	admin.GET("/legal-opinions", loHandler.GetAll)
	admin.GET("/legal-opinions/:id", loHandler.GetByID)
	admin.GET("/review-documents/presign", drHandler.GetPresignedURL)
	admin.GET("/review-documents/download", drHandler.Download)
	admin.GET("/review-documents", drHandler.GetAll)
	admin.GET("/review-documents/:id", drHandler.GetByID)

	admin.GET("/legal-cases/regencies", legalCaseHandler.ListRegencies)
	admin.GET("/legal-cases/cedants", legalCaseHandler.ListCedants)
	admin.POST("/legal-cases/cedants", legalCaseHandler.CreateCedant)
	admin.PUT("/legal-cases/cedants/:id", legalCaseHandler.UpdateCedant)
	admin.DELETE("/legal-cases/cedants/:id", legalCaseHandler.DeleteCedant)
	admin.GET("/divisions", divisionHandler.GetAll)
	admin.POST("/divisions", divisionHandler.Create)
	admin.GET("/divisions/:id", divisionHandler.GetByID)
	admin.PUT("/divisions/:id", divisionHandler.Update)
	admin.DELETE("/divisions/:id", divisionHandler.Delete)

	admin.GET("/companies", companyHandler.GetAll)
	admin.GET("/companies/:id", companyHandler.GetByID)
	admin.POST("/companies", companyHandler.Create)
	admin.PUT("/companies/:id", companyHandler.Update)
	admin.DELETE("/companies/:id", companyHandler.Delete)

	admin.GET("/purpose-types", purposeTypeHandler.GetAll)
	admin.GET("/purpose-types/:id", purposeTypeHandler.GetByID)
	admin.POST("/purpose-types", purposeTypeHandler.Create)
	admin.PUT("/purpose-types/:id", purposeTypeHandler.Update)
	admin.DELETE("/purpose-types/:id", purposeTypeHandler.Delete)

	admin.GET("/case-types", caseTypeHandler.GetAll)
	admin.GET("/case-types/:id", caseTypeHandler.GetByID)
	admin.POST("/case-types", caseTypeHandler.Create)
	admin.PUT("/case-types/:id", caseTypeHandler.Update)
	admin.DELETE("/case-types/:id", caseTypeHandler.Delete)

	admin.GET("/case-categories", caseCategoryHandler.GetAll)
	admin.GET("/case-categories/:id", caseCategoryHandler.GetByID)
	admin.POST("/case-categories", caseCategoryHandler.Create)
	admin.PUT("/case-categories/:id", caseCategoryHandler.Update)
	admin.DELETE("/case-categories/:id", caseCategoryHandler.Delete)

	admin.GET("/regencies", legalCaseHandler.ListRegencies)
	admin.GET("/cedants", legalCaseHandler.ListCedants)
	admin.POST("/cedants", legalCaseHandler.CreateCedant)
	admin.PUT("/cedants/:id", legalCaseHandler.UpdateCedant)
	admin.DELETE("/cedants/:id", legalCaseHandler.DeleteCedant)

	admin.GET("/legal-cases/download", legalCaseHandler.Download)
	admin.GET("/legal-cases/latest", legalCaseHandler.GetLatest)
	admin.GET("/legal-cases", legalCaseHandler.GetAll)
	admin.POST("/legal-cases", legalCaseHandler.Create)
	admin.GET("/legal-cases/:id", legalCaseHandler.GetByID)
	admin.PUT("/legal-cases/:id", legalCaseHandler.Update)
	admin.DELETE("/legal-cases/:id", legalCaseHandler.Delete)
	admin.POST("/legal-cases/:id/upload-document", legalCaseHandler.UploadDocument)
	admin.DELETE("/legal-cases/:id/document", legalCaseHandler.DeleteDocument)
	admin.GET("/legal-cases/:id/chronology", legalCaseHandler.ListChronologies)
	admin.POST("/legal-cases/:id/chronology", legalCaseHandler.CreateChronology)
	admin.PUT("/legal-cases/:id/chronology/:chronId", legalCaseHandler.UpdateChronology)
	admin.DELETE("/legal-cases/:id/chronology/:chronId", legalCaseHandler.DeleteChronology)

	admin.PATCH("/legal-opinions/:id/status", loHandler.AdminUpdateStatus)
	admin.POST("/legal-opinions/:id/result", loHandler.AdminUploadResult)
	admin.GET("/legal-opinions/:id/pdf", loHandler.GeneratePDF)
	admin.PATCH("/review-documents/:id/status", drHandler.AdminUpdateStatus)
	admin.POST("/review-documents/:id/result", drHandler.AdminUploadResult)

	admin.GET("/users", userHandler.GetAll)
	admin.POST("/users", userHandler.Create)
	admin.PUT("/users/:id", userHandler.Update)
	admin.PATCH("/users/:id/status", userHandler.UpdateStatus)
	admin.POST("/users/:id/reset-password", userHandler.ResetPassword)

	admin.GET("/audit-logs", auditLogHandler.GetAll)
	admin.GET("/notification-settings", notificationSettingHandler.GetAll)
	admin.PUT("/notification-settings/:id", notificationSettingHandler.Update)

	admin.GET("/materials", materialHandler.GetAll)
	admin.GET("/materials/:id", materialHandler.GetByID)
	admin.POST("/materials", materialHandler.Create)
	admin.PUT("/materials/:id", materialHandler.Update)
	admin.DELETE("/materials/:id", materialHandler.Delete)

	// ── Legal ─────────────────────────────────────────────────────────────────
	legal := api.Group("/legal")
	legal.Use(middleware.AuthMiddleware(cfg), middleware.RoleMiddleware("LEGAL"), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	legal.GET("/dashboard/stats", dashHandler.LegalStats)
	legal.GET("/dashboard/recent", dashHandler.LegalRecent)
	legal.GET("/dashboard/reminders", notificationSettingHandler.GetLegalReminders)
	legal.PATCH("/dashboard/reminders/read", notificationSettingHandler.MarkReminderRead)
	legal.PATCH("/dashboard/reminders/read-all", notificationSettingHandler.MarkAllRemindersRead)

	legal.GET("/legal-opinions/presign", loHandler.GetPresignedURL)
	legal.GET("/legal-opinions/download", loHandler.Download)
	legal.GET("/legal-opinions", loHandler.GetAll)
	legal.GET("/legal-opinions/:id", loHandler.GetByID)
	legal.PATCH("/legal-opinions/:id/status", loHandler.AdminUpdateStatus)
	legal.POST("/legal-opinions/:id/result", loHandler.AdminUploadResult)
	legal.GET("/legal-opinions/:id/pdf", loHandler.GeneratePDF)
	legal.GET("/review-documents/presign", drHandler.GetPresignedURL)
	legal.GET("/review-documents/download", drHandler.Download)
	legal.GET("/review-documents", drHandler.GetAll)
	legal.GET("/review-documents/:id", drHandler.GetByID)
	legal.PATCH("/review-documents/:id/status", drHandler.AdminUpdateStatus)
	legal.POST("/review-documents/:id/result", drHandler.AdminUploadResult)

	registerLegalCaseRoutes(legal, legalCaseHandler)

	legal.GET("/audit-logs", auditLogHandler.GetAll)

	legal.GET("/materials", materialHandler.GetAll)
	legal.GET("/materials/:id", materialHandler.GetByID)
	legal.POST("/materials", materialHandler.Create)
	legal.PUT("/materials/:id", materialHandler.Update)
	legal.DELETE("/materials/:id", materialHandler.Delete)

	// ─── Legal AU ─────────────────────────────────────────────────────────────
	legalAU := api.Group("/legal-au")
	legalAU.Use(middleware.AuthMiddleware(cfg), middleware.RoleMiddleware("LEGAL_AU"), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	legalAU.GET("/cases", legalCaseHandler.GetAll)
	legalAU.GET("/cases/:id", legalCaseHandler.GetByID)
	legalAU.PATCH("/cases/:id/status", legalCaseHandler.UpdateStatus)
	legalAU.POST("/cases/:id/chronology", legalCaseHandler.CreateChronology)
	legalAU.PUT("/cases/:id/chronology/:chronId", legalCaseHandler.UpdateChronology)
	legalAU.DELETE("/cases/:id/chronology/:chronId", legalCaseHandler.DeleteChronology)

	legalAU.GET("/materials", materialHandler.GetAll)
	legalAU.GET("/materials/:id", materialHandler.GetByID)
	legalAU.POST("/materials", materialHandler.Create)
	legalAU.PUT("/materials/:id", materialHandler.Update)
	legalAU.DELETE("/materials/:id", materialHandler.Delete)

	// ─── External ─────────────────────────────────────────────────────────────
	external := api.Group("/external")
	external.Use(middleware.AuthMiddleware(cfg), middleware.RoleMiddleware("EXTERNAL"), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	registerLegalCaseRoutes(external, legalCaseHandler)

	// ── Public materials ──────────────────────────────────────────────────────
	protectedMaterials := api.Group("")
	protectedMaterials.Use(middleware.AuthMiddleware(cfg))
	protectedMaterials.GET("/materials", materialHandler.GetAll)
	protectedMaterials.GET("/materials/:id", materialHandler.GetByID)

	log.Printf("Server running on port %s", cfg.App.Port)
	if err := r.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func registerLegalCaseRoutes(group *gin.RouterGroup, legalCaseHandler *handler.LegalCaseHandler) {
	group.GET("/legal-cases/regencies", legalCaseHandler.ListRegencies)
	group.GET("/legal-cases/cedants", legalCaseHandler.ListCedants)
	group.POST("/legal-cases/cedants", legalCaseHandler.CreateCedant)
	group.PUT("/legal-cases/cedants/:id", legalCaseHandler.UpdateCedant)
	group.DELETE("/legal-cases/cedants/:id", legalCaseHandler.DeleteCedant)
	group.GET("/legal-cases/download", legalCaseHandler.Download)
	group.GET("/legal-cases/latest", legalCaseHandler.GetLatest)
	group.GET("/legal-cases", legalCaseHandler.GetAll)
	group.POST("/legal-cases", legalCaseHandler.Create)
	group.GET("/legal-cases/:id", legalCaseHandler.GetByID)
	group.PUT("/legal-cases/:id", legalCaseHandler.Update)
	group.DELETE("/legal-cases/:id", legalCaseHandler.Delete)
	group.POST("/legal-cases/:id/upload-document", legalCaseHandler.UploadDocument)
	group.DELETE("/legal-cases/:id/document", legalCaseHandler.DeleteDocument)
	group.GET("/legal-cases/:id/chronology", legalCaseHandler.ListChronologies)
	group.POST("/legal-cases/:id/chronology", legalCaseHandler.CreateChronology)
	group.PUT("/legal-cases/:id/chronology/:chronId", legalCaseHandler.UpdateChronology)
	group.DELETE("/legal-cases/:id/chronology/:chronId", legalCaseHandler.DeleteChronology)
}
