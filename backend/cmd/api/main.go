package main

import (
	"log"

	"legal-riu-portal/internal/config"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/handler"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	db := config.InitDatabase(cfg)
	store := storage.InitMinIO(cfg)

	if err := db.AutoMigrate(
		&entity.User{},
		&entity.RefreshToken{},
		&entity.LegalOpinion{},
		&entity.LegalOpinionAttachment{},
		&entity.LegalOpinionResult{},
		&entity.DocumentReview{},
		&entity.DocumentReviewAttachment{},
		&entity.DocumentReviewResult{},
	); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Database migrated successfully")

	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	loRepo := repository.NewLegalOpinionRepository(db)
	drRepo := repository.NewDocumentReviewRepository(db)
	dashRepo := repository.NewDashboardRepository(db)

	// ── Services ─────────────────────────────────────────────────────────────
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, cfg)
	userSvc := service.NewUserService(userRepo)
	loSvc := service.NewLegalOpinionService(loRepo, store)
	drSvc := service.NewDocumentReviewService(drRepo, store)
	dashSvc := service.NewDashboardService(dashRepo)

	// ── Handlers ─────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	loHandler := handler.NewLegalOpinionHandler(loSvc)
	drHandler := handler.NewDocumentReviewHandler(drSvc)
	dashHandler := handler.NewDashboardHandler(dashSvc)

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
	loginLimiter := middleware.NewLoginRateLimiter(cfg.Security.LoginRateWindowMinutes)

	// ── Public auth ───────────────────────────────────────────────────────────
	api.POST("/auth/login", loginLimiter.Middleware(), authHandler.Login)
	api.POST("/auth/refresh", loginLimiter.Middleware(), authHandler.RefreshToken)
	api.POST("/auth/logout", authHandler.Logout)

	// ── Protected ─────────────────────────────────────────────────────────────
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg), middleware.CSRFProtection())

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

	// Legal opinions — user CRUD
	protected.GET("/legal-opinions", loHandler.GetAll)
	protected.POST("/legal-opinions", loHandler.Create)
	protected.GET("/legal-opinions/:id", loHandler.GetByID)
	protected.PUT("/legal-opinions/:id", loHandler.Update)
	protected.DELETE("/legal-opinions/:id", loHandler.Delete)
	protected.POST("/legal-opinions/:id/resubmit", loHandler.Resubmit)

	// Review documents
	protected.GET("/review-documents/presign", drHandler.GetPresignedURL)
	protected.GET("/review-documents", drHandler.GetAll)
	protected.POST("/review-documents", drHandler.Create)
	protected.GET("/review-documents/:id", drHandler.GetByID)
	protected.PUT("/review-documents/:id", drHandler.Update)
	protected.DELETE("/review-documents/:id", drHandler.Delete)
	protected.POST("/review-documents/:id/resubmit", drHandler.Resubmit)

	// ── Admin only ────────────────────────────────────────────────────────────
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(cfg), middleware.RoleMiddleware("ADMIN"), middleware.CSRFProtection())

	admin.GET("/dashboard/stats", dashHandler.AdminStats)
	admin.GET("/dashboard/recent", dashHandler.AdminRecent)

	admin.PATCH("/legal-opinions/:id/status", loHandler.AdminUpdateStatus)
	admin.POST("/legal-opinions/:id/result", loHandler.AdminUploadResult)
	admin.PATCH("/review-documents/:id/status", drHandler.AdminUpdateStatus)
	admin.POST("/review-documents/:id/result", drHandler.AdminUploadResult)

	admin.GET("/users", userHandler.GetAll)
	admin.POST("/users", userHandler.Create)
	admin.PUT("/users/:id", userHandler.Update)
	admin.PATCH("/users/:id/status", userHandler.UpdateStatus)
	admin.POST("/users/:id/reset-password", userHandler.ResetPassword)

	log.Printf("Server running on port %s", cfg.App.Port)
	if err := r.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}