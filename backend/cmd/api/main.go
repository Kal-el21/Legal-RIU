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
	documentTypeRepo := repository.NewDocumentTypeRepository(db)
	materialRepo := repository.NewLegalMaterialRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)

	// ── Services ─────────────────────────────────────────────────────────────
	permissionSvc := service.NewPermissionService(permissionRepo, userRepo)
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, cfg, permissionSvc)
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
	documentTypeSvc := service.NewDocumentTypeService(documentTypeRepo)
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
	documentTypeHandler := handler.NewDocumentTypeHandler(documentTypeSvc)
	materialHandler := handler.NewLegalMaterialHandler(materialSvc)
	permissionHandler := handler.NewPermissionHandler(permissionSvc)

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
	requirePermission := func(codes ...string) gin.HandlerFunc {
		return middleware.RequirePermission(permissionSvc, codes...)
	}

	// ── Protected ─────────────────────────────────────────────────────────────
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg), middleware.PermissionContextMiddleware(permissionSvc), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	protected.GET("/auth/me", authHandler.Me)
	protected.GET("/auth/permissions", authHandler.Permissions)
	protected.POST("/auth/change-password", authHandler.ChangePassword)

	// Settings
	protected.PUT("/settings/profile", authHandler.UpdateProfile)
	protected.PUT("/settings/notifications", authHandler.UpdateNotification)
	protected.PUT("/settings/two-fa", authHandler.Toggle2FA)

	// Dashboard
	protected.GET("/divisions", divisionHandler.GetAll)
	protected.GET("/companies", requirePermission("case_management.view", "case_management.create", "master_data.view"), companyHandler.GetAll)
	protected.GET("/case-types", requirePermission("case_management.view", "case_management.create", "master_data.view"), caseTypeHandler.GetAll)
	protected.GET("/case-categories", requirePermission("case_management.view", "case_management.create", "master_data.view"), caseCategoryHandler.GetAll)
	protected.GET("/document-types", requirePermission("case_management.view", "case_management.create", "master_data.view"), documentTypeHandler.GetAll)

	userOnly := protected.Group("")
	userOnly.Use(middleware.RoleMiddleware("USER"))

	userOnly.GET("/dashboard/stats", requirePermission("dashboard.user.view"), dashHandler.UserStats)
	userOnly.GET("/dashboard/recent", requirePermission("dashboard.user.view"), dashHandler.UserRecent)
	userOnly.GET("/dashboard/reminders", requirePermission("dashboard.user.view"), dashHandler.GetReminders)
	userOnly.PATCH("/dashboard/reminders/read", requirePermission("dashboard.user.view"), notificationSettingHandler.MarkReminderRead)
	userOnly.PATCH("/dashboard/reminders/read-all", requirePermission("dashboard.user.view"), notificationSettingHandler.MarkAllRemindersRead)

	// Legal opinions — presign
	userOnly.GET("/legal-opinions/presign", requirePermission("legal_opinion.view.own", "legal_opinion.view.all"), loHandler.GetPresignedURL)
	userOnly.GET("/legal-opinions/download", requirePermission("legal_opinion.download.all", "legal_opinion.view.own", "legal_opinion.view.all"), loHandler.Download)
	userOnly.GET("/legal-opinions/:id/pdf", requirePermission("legal_opinion.download.all", "legal_opinion.view.own", "legal_opinion.view.all"), loHandler.GeneratePDF)
	userOnly.GET("/legal-opinions", requirePermission("legal_opinion.view.own", "legal_opinion.view.all"), loHandler.GetAll)
	userOnly.POST("/legal-opinions", requirePermission("legal_opinion.create.own"), loHandler.Create)
	userOnly.GET("/legal-opinions/:id", requirePermission("legal_opinion.view.own", "legal_opinion.view.all"), loHandler.GetByID)
	userOnly.PUT("/legal-opinions/:id", requirePermission("legal_opinion.update.own"), loHandler.Update)
	userOnly.DELETE("/legal-opinions/:id", requirePermission("legal_opinion.delete.own"), loHandler.Delete)
	userOnly.POST("/legal-opinions/:id/resubmit", requirePermission("legal_opinion.resubmit.own"), loHandler.Resubmit)

	// Review documents
	userOnly.GET("/review-documents/presign", requirePermission("document_review.view.own", "document_review.view.all"), drHandler.GetPresignedURL)
	userOnly.GET("/review-documents/download", requirePermission("document_review.download.all", "document_review.view.own", "document_review.view.all"), drHandler.Download)
	userOnly.GET("/review-documents", requirePermission("document_review.view.own", "document_review.view.all"), drHandler.GetAll)
	userOnly.POST("/review-documents", requirePermission("document_review.create.own"), drHandler.Create)
	userOnly.GET("/review-documents/:id", requirePermission("document_review.view.own", "document_review.view.all"), drHandler.GetByID)
	userOnly.PUT("/review-documents/:id", requirePermission("document_review.update.own"), drHandler.Update)
	userOnly.DELETE("/review-documents/:id", requirePermission("document_review.delete.own"), drHandler.Delete)
	userOnly.POST("/review-documents/:id/resubmit", requirePermission("document_review.resubmit.own"), drHandler.Resubmit)

	registerLegalCaseRoutes(userOnly, legalCaseHandler, requirePermission)

	// ── Admin only ─────────────────────────────────────────────────────────────
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(cfg), middleware.PermissionContextMiddleware(permissionSvc), middleware.RoleMiddleware("ADMIN"), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

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

	admin.GET("/document-types", documentTypeHandler.GetAll)
	admin.GET("/document-types/:id", documentTypeHandler.GetByID)
	admin.POST("/document-types", documentTypeHandler.Create)
	admin.PUT("/document-types/:id", documentTypeHandler.Update)
	admin.DELETE("/document-types/:id", documentTypeHandler.Delete)

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
	admin.GET("/permissions", requirePermission("user_management.manage_permissions"), permissionHandler.GetCatalog)
	admin.GET("/users/:id/permissions", requirePermission("user_management.manage_permissions"), permissionHandler.GetUserAccess)
	admin.PUT("/users/:id/permissions", requirePermission("user_management.manage_permissions"), permissionHandler.UpdateUserAccess)
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
	legal.Use(middleware.AuthMiddleware(cfg), middleware.PermissionContextMiddleware(permissionSvc), middleware.RoleMiddleware("LEGAL"), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	legal.GET("/dashboard/stats", requirePermission("dashboard.legal.view"), dashHandler.LegalStats)
	legal.GET("/dashboard/recent", requirePermission("dashboard.legal.view"), dashHandler.LegalRecent)
	legal.GET("/dashboard/reminders", requirePermission("dashboard.legal.view"), notificationSettingHandler.GetLegalReminders)
	legal.PATCH("/dashboard/reminders/read", notificationSettingHandler.MarkReminderRead)
	legal.PATCH("/dashboard/reminders/read-all", notificationSettingHandler.MarkAllRemindersRead)

	legal.GET("/legal-opinions/presign", requirePermission("legal_opinion.view.all"), loHandler.GetPresignedURL)
	legal.GET("/legal-opinions/download", requirePermission("legal_opinion.download.all", "legal_opinion.view.all"), loHandler.Download)
	legal.GET("/legal-opinions", requirePermission("legal_opinion.view.all"), loHandler.GetAll)
	legal.GET("/legal-opinions/:id", requirePermission("legal_opinion.view.all"), loHandler.GetByID)
	legal.PATCH("/legal-opinions/:id/status", requirePermission("legal_opinion.update_status.all"), loHandler.AdminUpdateStatus)
	legal.POST("/legal-opinions/:id/result", requirePermission("legal_opinion.upload_result.all"), loHandler.AdminUploadResult)
	legal.GET("/legal-opinions/:id/pdf", requirePermission("legal_opinion.download.all", "legal_opinion.view.all"), loHandler.GeneratePDF)
	legal.GET("/review-documents/presign", requirePermission("document_review.view.all"), drHandler.GetPresignedURL)
	legal.GET("/review-documents/download", requirePermission("document_review.download.all", "document_review.view.all"), drHandler.Download)
	legal.GET("/review-documents", requirePermission("document_review.view.all"), drHandler.GetAll)
	legal.GET("/review-documents/:id", requirePermission("document_review.view.all"), drHandler.GetByID)
	legal.PATCH("/review-documents/:id/status", requirePermission("document_review.update_status.all"), drHandler.AdminUpdateStatus)
	legal.POST("/review-documents/:id/result", requirePermission("document_review.upload_result.all"), drHandler.AdminUploadResult)

	registerLegalCaseRoutes(legal, legalCaseHandler, requirePermission)

	legal.GET("/audit-logs", requirePermission("audit_log.view"), auditLogHandler.GetAll)

	legal.GET("/materials", requirePermission("legal_material.view", "legal_material.manage"), materialHandler.GetAll)
	legal.GET("/materials/:id", requirePermission("legal_material.view", "legal_material.manage"), materialHandler.GetByID)
	legal.POST("/materials", requirePermission("legal_material.manage"), materialHandler.Create)
	legal.PUT("/materials/:id", requirePermission("legal_material.manage"), materialHandler.Update)
	legal.DELETE("/materials/:id", requirePermission("legal_material.manage"), materialHandler.Delete)

	// ─── Legal AU ─────────────────────────────────────────────────────────────
	legalAU := api.Group("/legal-au")
	legalAU.Use(middleware.AuthMiddleware(cfg), middleware.PermissionContextMiddleware(permissionSvc), middleware.RoleMiddleware("LEGAL_AU"), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	legalAU.GET("/cases", requirePermission("case_management.view"), legalCaseHandler.GetAll)
	legalAU.GET("/cases/:id", requirePermission("case_management.view"), legalCaseHandler.GetByID)
	legalAU.PATCH("/cases/:id/status", requirePermission("case_management.update_status"), legalCaseHandler.UpdateStatus)
	legalAU.POST("/cases/:id/chronology", requirePermission("case_management.manage_chronology"), legalCaseHandler.CreateChronology)
	legalAU.PUT("/cases/:id/chronology/:chronId", requirePermission("case_management.manage_chronology"), legalCaseHandler.UpdateChronology)
	legalAU.DELETE("/cases/:id/chronology/:chronId", requirePermission("case_management.manage_chronology"), legalCaseHandler.DeleteChronology)

	legalAU.GET("/materials", requirePermission("legal_material.view", "legal_material.manage"), materialHandler.GetAll)
	legalAU.GET("/materials/:id", requirePermission("legal_material.view", "legal_material.manage"), materialHandler.GetByID)
	legalAU.POST("/materials", requirePermission("legal_material.manage"), materialHandler.Create)
	legalAU.PUT("/materials/:id", requirePermission("legal_material.manage"), materialHandler.Update)
	legalAU.DELETE("/materials/:id", requirePermission("legal_material.manage"), materialHandler.Delete)

	// ─── External ─────────────────────────────────────────────────────────────
	external := api.Group("/external")
	external.Use(middleware.AuthMiddleware(cfg), middleware.PermissionContextMiddleware(permissionSvc), middleware.RoleMiddleware("EXTERNAL"), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	registerLegalCaseRoutes(external, legalCaseHandler, requirePermission)

	// ── Public materials ──────────────────────────────────────────────────────
	protectedMaterials := api.Group("")
	protectedMaterials.Use(middleware.AuthMiddleware(cfg), middleware.PermissionContextMiddleware(permissionSvc))
	protectedMaterials.GET("/materials", requirePermission("legal_material.view", "legal_material.manage"), materialHandler.GetAll)
	protectedMaterials.GET("/materials/:id", requirePermission("legal_material.view", "legal_material.manage"), materialHandler.GetByID)

	log.Printf("Server running on port %s", cfg.App.Port)
	if err := r.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func registerLegalCaseRoutes(group *gin.RouterGroup, legalCaseHandler *handler.LegalCaseHandler, requirePermission func(...string) gin.HandlerFunc) {
	group.GET("/legal-cases/regencies", requirePermission("case_management.view", "case_management.create"), legalCaseHandler.ListRegencies)
	group.GET("/legal-cases/cedants", requirePermission("case_management.view", "case_management.create"), legalCaseHandler.ListCedants)
	group.POST("/legal-cases/cedants", requirePermission("case_management.manage_reference"), legalCaseHandler.CreateCedant)
	group.PUT("/legal-cases/cedants/:id", requirePermission("case_management.manage_reference"), legalCaseHandler.UpdateCedant)
	group.DELETE("/legal-cases/cedants/:id", requirePermission("case_management.manage_reference"), legalCaseHandler.DeleteCedant)
	group.GET("/legal-cases/download", requirePermission("case_management.view", "case_management.manage_document"), legalCaseHandler.Download)
	group.GET("/legal-cases/latest", requirePermission("case_management.view"), legalCaseHandler.GetLatest)
	group.GET("/legal-cases", requirePermission("case_management.view"), legalCaseHandler.GetAll)
	group.POST("/legal-cases", requirePermission("case_management.create"), legalCaseHandler.Create)
	group.GET("/legal-cases/:id", requirePermission("case_management.view"), legalCaseHandler.GetByID)
	group.PUT("/legal-cases/:id", requirePermission("case_management.update"), legalCaseHandler.Update)
	group.DELETE("/legal-cases/:id", requirePermission("case_management.delete"), legalCaseHandler.Delete)
	group.POST("/legal-cases/:id/upload-document", requirePermission("case_management.manage_document"), legalCaseHandler.UploadDocument)
	group.DELETE("/legal-cases/:id/document", requirePermission("case_management.manage_document"), legalCaseHandler.DeleteDocument)
	group.GET("/legal-cases/:id/chronology", requirePermission("case_management.view", "case_management.manage_chronology"), legalCaseHandler.ListChronologies)
	group.POST("/legal-cases/:id/chronology", requirePermission("case_management.manage_chronology"), legalCaseHandler.CreateChronology)
	group.PUT("/legal-cases/:id/chronology/:chronId", requirePermission("case_management.manage_chronology"), legalCaseHandler.UpdateChronology)
	group.DELETE("/legal-cases/:id/chronology/:chronId", requirePermission("case_management.manage_chronology"), legalCaseHandler.DeleteChronology)
}
