package main

import (
	"log"

	"legal-riu-portal/internal/config"
	"legal-riu-portal/internal/entity"
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

	if err := db.Migrator().DropTable(&entity.CaseChronology{}, &entity.LegalCase{}); err != nil {
		log.Fatalf("Failed to drop legal_case tables: %v", err)
	}
	log.Println("Legal case tables recreated to accommodate PIC -> uuid migration")
	if err := db.AutoMigrate(
		&entity.User{},
		&entity.RefreshToken{},
		&entity.LegalOpinion{},
		&entity.LegalOpinionAttachment{},
		&entity.LegalOpinionResult{},
		&entity.DocumentReview{},
		&entity.DocumentReviewAttachment{},
		&entity.DocumentReviewResult{},
		&entity.Regency{},
		&entity.Cedant{},
		&entity.Division{},
		&entity.LegalCase{},
		&entity.CaseChronology{},
		&entity.AuditLog{},
	); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Database migrated successfully")
	if err := seed.SeedRegencies(db); err != nil {
		log.Fatalf("Regency seed failed: %v", err)
	}
	log.Println("Regencies seeded successfully")
	if err := seed.SeedDivisions(db); err != nil {
		log.Fatalf("Division seed failed: %v", err)
	}
	log.Println("Divisions seeded successfully")

	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	loRepo := repository.NewLegalOpinionRepository(db)
	drRepo := repository.NewDocumentReviewRepository(db)
	dashRepo := repository.NewDashboardRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	legalCaseRepo := repository.NewLegalCaseRepository(db)
	divisionRepo := repository.NewDivisionRepository(db)

	// ── Services ─────────────────────────────────────────────────────────────
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, cfg)
	userSvc := service.NewUserService(userRepo)
	loSvc := service.NewLegalOpinionService(loRepo, store)
	drSvc := service.NewDocumentReviewService(drRepo, store)
	dashSvc := service.NewDashboardService(dashRepo)
	auditLogSvc := service.NewAuditLogService(auditLogRepo)
	legalCaseSvc := service.NewLegalCaseService(legalCaseRepo, store)
	divisionSvc := service.NewDivisionService(divisionRepo)

	// ── Handlers ─────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authSvc, cfg, auditLogSvc)
	userHandler := handler.NewUserHandler(userSvc, auditLogSvc)
	loHandler := handler.NewLegalOpinionHandler(loSvc, auditLogSvc)
	drHandler := handler.NewDocumentReviewHandler(drSvc, auditLogSvc)
	dashHandler := handler.NewDashboardHandler(dashSvc)
	auditLogHandler := handler.NewAuditLogHandler(auditLogSvc, auditLogRepo)
	legalCaseHandler := handler.NewLegalCaseHandler(legalCaseSvc, auditLogSvc)
	divisionHandler := handler.NewDivisionHandler(divisionSvc)

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
	protected.GET("/dashboard/stats", dashHandler.UserStats)
	protected.GET("/dashboard/recent", dashHandler.UserRecent)

	// Legal opinions — presign
	protected.GET("/legal-opinions/presign", loHandler.GetPresignedURL)
	protected.GET("/legal-opinions/download", loHandler.Download)
	protected.GET("/legal-opinions/:id/pdf", loHandler.GeneratePDF)
	protected.GET("/legal-opinions", loHandler.GetAll)
	protected.POST("/legal-opinions", loHandler.Create)
	protected.GET("/legal-opinions/:id", loHandler.GetByID)
	protected.PUT("/legal-opinions/:id", loHandler.Update)
	protected.DELETE("/legal-opinions/:id", loHandler.Delete)
	protected.POST("/legal-opinions/:id/resubmit", loHandler.Resubmit)

	// Review documents
	protected.GET("/review-documents/presign", drHandler.GetPresignedURL)
	protected.GET("/review-documents/download", drHandler.Download)
	protected.GET("/review-documents", drHandler.GetAll)
	protected.POST("/review-documents", drHandler.Create)
	protected.GET("/review-documents/:id", drHandler.GetByID)
	protected.PUT("/review-documents/:id", drHandler.Update)
	protected.DELETE("/review-documents/:id", drHandler.Delete)
	protected.POST("/review-documents/:id/resubmit", drHandler.Resubmit)

	// ── Admin only ────────────────────────────────────────────────────────────
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(cfg), middleware.RoleMiddleware("ADMIN"), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	admin.GET("/dashboard/stats", dashHandler.AdminStats)
	admin.GET("/dashboard/recent", dashHandler.AdminRecent)

	admin.GET("/legal-cases/regencies", legalCaseHandler.ListRegencies)
	admin.GET("/legal-cases/cedants", legalCaseHandler.ListCedants)
	admin.POST("/legal-cases/cedants", legalCaseHandler.CreateCedant)
	admin.PUT("/legal-cases/cedants/:id", legalCaseHandler.UpdateCedant)
	admin.DELETE("/legal-cases/cedants/:id", legalCaseHandler.DeleteCedant)
	admin.GET("/divisions", divisionHandler.GetAll)
	admin.POST("/divisions", divisionHandler.Create)
	admin.PUT("/divisions/:id", divisionHandler.Update)
	admin.DELETE("/divisions/:id", divisionHandler.Delete)
	admin.GET("/legal-cases/download", legalCaseHandler.Download)
	admin.GET("/legal-cases/latest", legalCaseHandler.GetLatest)
	admin.GET("/legal-cases", legalCaseHandler.GetAll)
	admin.POST("/legal-cases", legalCaseHandler.Create)
	admin.GET("/legal-cases/:id", legalCaseHandler.GetByID)
	admin.PUT("/legal-cases/:id", legalCaseHandler.Update)
	admin.DELETE("/legal-cases/:id", legalCaseHandler.Delete)
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

	// ── Legal ────────────────────────────────────────────────────────────────
	legal := api.Group("/legal")
	legal.Use(middleware.AuthMiddleware(cfg), middleware.RoleMiddleware("LEGAL"), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	legal.GET("/dashboard/stats", dashHandler.LegalStats)
	legal.GET("/dashboard/recent", dashHandler.LegalRecent)

	legal.PATCH("/legal-opinions/:id/status", loHandler.AdminUpdateStatus)
	legal.POST("/legal-opinions/:id/result", loHandler.AdminUploadResult)
	legal.GET("/legal-opinions/:id/pdf", loHandler.GeneratePDF)
	legal.PATCH("/review-documents/:id/status", drHandler.AdminUpdateStatus)
	legal.POST("/review-documents/:id/result", drHandler.AdminUploadResult)

	legal.GET("/legal-opinions", loHandler.GetAll)
	legal.GET("/legal-opinions/:id", loHandler.GetByID)
	legal.GET("/review-documents", drHandler.GetAll)
	legal.GET("/review-documents/:id", drHandler.GetByID)

	// ─── External ─────────────────────────────────────────────────────────────
	external := api.Group("/external")
	external.Use(middleware.AuthMiddleware(cfg), middleware.RoleMiddleware("EXTERNAL"), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	external.GET("/dashboard/stats", dashHandler.ExternalStats)
	external.GET("/dashboard/recent", dashHandler.ExternalRecent)

	external.GET("/legal-opinions", loHandler.GetAll)
	external.GET("/legal-opinions/:id", loHandler.GetByID)
	external.GET("/review-documents", drHandler.GetAll)
	external.GET("/review-documents/:id", drHandler.GetByID)

	log.Printf("Server running on port %s", cfg.App.Port)
	if err := r.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
