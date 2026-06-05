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
	loRepo := repository.NewLegalOpinionRepository(db)

	// ── Services ─────────────────────────────────────────────────────────────
	authSvc := service.NewAuthService(userRepo, cfg)
	loSvc := service.NewLegalOpinionService(loRepo, store)

	// ── Handlers ─────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authSvc)
	loHandler := handler.NewLegalOpinionHandler(loSvc)

	// ── Gin ──────────────────────────────────────────────────────────────────
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.MaxMultipartMemory = 110 << 20 // 110 MB

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "Legal RIU Portal API"})
	})

	api := r.Group("/api/v1")

	// ── Public auth ───────────────────────────────────────────────────────────
	api.POST("/auth/login", authHandler.Login)

	// ── Protected ─────────────────────────────────────────────────────────────
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))

	protected.GET("/auth/me", authHandler.Me)
	protected.POST("/auth/change-password", authHandler.ChangePassword)

	// Legal opinions — presign
	protected.GET("/legal-opinions/presign", loHandler.GetPresignedURL)

	// Legal opinions — user CRUD
	protected.GET("/legal-opinions", loHandler.GetAll)
	protected.POST("/legal-opinions", loHandler.Create)
	protected.GET("/legal-opinions/:id", loHandler.GetByID)
	protected.PUT("/legal-opinions/:id", loHandler.Update)
	protected.DELETE("/legal-opinions/:id", loHandler.Delete)
	protected.POST("/legal-opinions/:id/resubmit", loHandler.Resubmit)

	// ── Admin only ────────────────────────────────────────────────────────────
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(cfg), middleware.RoleMiddleware("ADMIN"))

	admin.PATCH("/legal-opinions/:id/status", loHandler.AdminUpdateStatus)
	admin.POST("/legal-opinions/:id/result", loHandler.AdminUploadResult)

	log.Printf("Server running on port %s", cfg.App.Port)
	if err := r.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
