package router

import (
	"legal-riu-portal/internal/config"
	"legal-riu-portal/internal/handler"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// New merakit seluruh aplikasi (repository, service, handler, middleware, route)
// dan mengembalikan router siap pakai. Dipakai oleh cmd/api dan oleh tes.
func New(cfg *config.Config, db *gorm.DB, store *storage.MinIOClient) (*gin.Engine, error) {
	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	loRepo := repository.NewLegalOpinionRepository(db)
	drRepo := repository.NewDocumentReviewRepository(db)
	agreementRepo := repository.NewAgreementDocumentRepository(db)
	agreementTemplateRepo := repository.NewAgreementTemplateRepository(db)
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
	documentRepositoryRepo := repository.NewRepositoryDocumentRepository(db)

	// ── Services ─────────────────────────────────────────────────────────────
	permissionSvc := service.NewPermissionService(permissionRepo, userRepo)
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, cfg, permissionSvc)
	userSvc := service.NewUserService(userRepo)
	documentRepositorySvc := service.NewRepositoryDocumentService(documentRepositoryRepo, store)
	loSvc := service.NewLegalOpinionService(loRepo, store, documentRepositorySvc)
	drSvc := service.NewDocumentReviewService(drRepo, store, documentRepositorySvc)
	agreementTemplateSvc := service.NewAgreementTemplateService(agreementTemplateRepo, store)
	agreementSvc := service.NewAgreementDocumentService(agreementRepo, store, agreementTemplateSvc, documentRepositorySvc)
	agreementTypeSvc := service.NewAgreementTypeService(agreementTemplateRepo)
	notificationSettingSvc := service.NewNotificationSettingService(notificationSettingRepo, dashRepo)
	dashSvc := service.NewDashboardService(dashRepo)
	auditLogSvc := service.NewAuditLogService(auditLogRepo)
	legalCaseSvc := service.NewLegalCaseService(legalCaseRepo, store, documentRepositorySvc)
	divisionSvc := service.NewDivisionService(divisionRepo)
	companySvc := service.NewCompanyService(companyRepo)
	purposeTypeSvc := service.NewPurposeTypeService(purposeTypeRepo)
	caseTypeSvc := service.NewCaseTypeService(caseTypeRepo)
	caseCategorySvc := service.NewCaseCategoryService(caseCategoryRepo)
	documentTypeSvc := service.NewDocumentTypeService(documentTypeRepo)
	materialSvc := service.NewLegalMaterialService(materialRepo)
	reportSvc := service.NewReportService(db)

	// ── Handlers ─────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authSvc, cfg, auditLogSvc)
	userHandler := handler.NewUserHandler(userSvc, auditLogSvc)
	loHandler := handler.NewLegalOpinionHandler(loSvc, auditLogSvc)
	drHandler := handler.NewDocumentReviewHandler(drSvc, auditLogSvc)
	agreementHandler := handler.NewAgreementDocumentHandler(agreementSvc)
	agreementTemplateHandler := handler.NewAgreementTemplateHandler(agreementTemplateSvc)
	agreementTypeHandler := handler.NewAgreementTypeHandler(agreementTypeSvc)
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
	reportHandler := handler.NewReportHandler(reportSvc)
	documentRepositoryHandler := handler.NewRepositoryDocumentHandler(documentRepositorySvc, userRepo)

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
	api.POST("/auth/ldap-login", loginLimiter.Middleware("username"), authHandler.LDAPLogin)
	api.POST("/auth/login", loginLimiter.Middleware("email"), authHandler.Login)
	api.POST("/auth/refresh", loginLimiter.Middleware(""), authHandler.RefreshToken)
	api.POST("/auth/logout", authHandler.Logout)
	requirePermission := func(codes ...string) gin.HandlerFunc {
		return middleware.RequirePermission(permissionSvc, codes...)
	}

	// ── Protected ─────────────────────────────────────────────────────────────
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg, userRepo), middleware.PermissionContextMiddleware(permissionSvc), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	protected.GET("/auth/me", authHandler.Me)
	protected.GET("/auth/permissions", authHandler.Permissions)
	protected.POST("/auth/change-password", authHandler.ChangePassword)

	// Settings
	protected.PUT("/settings/profile", authHandler.UpdateProfile)
	protected.PUT("/settings/notifications", authHandler.UpdateNotification)
	protected.PUT("/settings/two-fa", authHandler.Toggle2FA)

	// Document Repository (accessible by all authenticated roles)
	protected.GET("/document-repositories", requirePermission("document_repository.view"), documentRepositoryHandler.GetAll)
	protected.GET("/document-repositories/:id", requirePermission("document_repository.view"), documentRepositoryHandler.GetByID)
	protected.GET("/document-repositories/:id/download", requirePermission("document_repository.view"), documentRepositoryHandler.Download)

	// Dashboard
	protected.GET("/divisions", divisionHandler.GetAll)
	protected.GET("/companies", requirePermission("case_management.view", "case_management.create", "master_data.view"), companyHandler.GetAll)
	protected.GET("/case-types", requirePermission("case_management.view", "case_management.create", "master_data.view"), caseTypeHandler.GetAll)
	protected.GET("/case-categories", requirePermission("case_management.view", "case_management.create", "master_data.view"), caseCategoryHandler.GetAll)
	protected.GET("/document-types", documentTypeHandler.GetAll)
	protected.GET("/agreement-document-types", requirePermission("agreement_document.view.own", "agreement_document.view.all", "agreement_document.create.own"), agreementHandler.ListTypes)
	protected.GET("/agreement-document-types/:code/schema", requirePermission("agreement_document.view.own", "agreement_document.view.all", "agreement_document.create.own"), agreementHandler.GetType)

	// Lonceng notifikasi: satu alamat untuk semua peran, terbuka bagi yang boleh melihat pengajuan
	lihatPengajuan := requirePermission(
		"legal_opinion.view.own", "legal_opinion.view.all",
		"document_review.view.own", "document_review.view.all",
		"agreement_document.view.own", "agreement_document.view.all",
	)
	protected.GET("/notifications", lihatPengajuan, notificationSettingHandler.GetReminders)
	protected.PATCH("/notifications/read", lihatPengajuan, notificationSettingHandler.MarkReminderRead)
	protected.PATCH("/notifications/read-all", lihatPengajuan, notificationSettingHandler.MarkAllRemindersRead)

	userOnly := protected.Group("")
	userOnly.Use(middleware.RoleMiddleware("USER"))

	userOnly.GET("/dashboard/stats", requirePermission("dashboard.user.view"), dashHandler.UserStats)
	userOnly.GET("/dashboard/recent", requirePermission("dashboard.user.view"), dashHandler.UserRecent)

	// Legal opinions
	userOnly.GET("/legal-opinions/download", requirePermission("legal_opinion.download.all", "legal_opinion.view.own", "legal_opinion.view.all"), loHandler.Download)
	userOnly.GET("/legal-opinions/:id/pdf", requirePermission("legal_opinion.download.all", "legal_opinion.view.own", "legal_opinion.view.all"), loHandler.GeneratePDF)
	userOnly.GET("/legal-opinions", requirePermission("legal_opinion.view.own", "legal_opinion.view.all"), loHandler.GetAll)
	userOnly.POST("/legal-opinions", requirePermission("legal_opinion.create.own"), loHandler.Create)
	userOnly.GET("/legal-opinions/:id", requirePermission("legal_opinion.view.own", "legal_opinion.view.all"), loHandler.GetByID)
	userOnly.GET("/legal-opinions/:id/history", requirePermission("legal_opinion.view.own", "legal_opinion.view.all"), auditLogHandler.Riwayat("legal_opinion", loHandler.BolehBuka))
	userOnly.PUT("/legal-opinions/:id", requirePermission("legal_opinion.update.own"), loHandler.Update)
	userOnly.DELETE("/legal-opinions/:id", requirePermission("legal_opinion.delete.own"), loHandler.Delete)
	userOnly.POST("/legal-opinions/:id/resubmit", requirePermission("legal_opinion.resubmit.own"), loHandler.Resubmit)
	// Draf (PR Draf): hanya pemiliknya; memakai izin yang ada.
	userOnly.POST("/legal-opinions/drafts", requirePermission("legal_opinion.create.own"), loHandler.SaveDraft)
	userOnly.PUT("/legal-opinions/:id/draft", requirePermission("legal_opinion.update.own"), loHandler.UpdateDraft)
	userOnly.DELETE("/legal-opinions/:id/attachments/:attachmentId", requirePermission("legal_opinion.update.own"), loHandler.DeleteDraftAttachment)
	userOnly.POST("/legal-opinions/:id/submit", requirePermission("legal_opinion.create.own"), loHandler.SubmitDraft)

	// Review documents
	userOnly.GET("/review-documents/download", requirePermission("document_review.download.all", "document_review.view.own", "document_review.view.all"), drHandler.Download)
	userOnly.GET("/review-documents", requirePermission("document_review.view.own", "document_review.view.all"), drHandler.GetAll)
	userOnly.POST("/review-documents", requirePermission("document_review.create.own"), drHandler.Create)
	userOnly.GET("/review-documents/:id", requirePermission("document_review.view.own", "document_review.view.all"), drHandler.GetByID)
	userOnly.GET("/review-documents/:id/history", requirePermission("document_review.view.own", "document_review.view.all"), auditLogHandler.Riwayat("document_review", drHandler.BolehBuka))
	userOnly.PUT("/review-documents/:id", requirePermission("document_review.update.own"), drHandler.Update)
	userOnly.DELETE("/review-documents/:id", requirePermission("document_review.delete.own"), drHandler.Delete)
	userOnly.POST("/review-documents/:id/resubmit", requirePermission("document_review.resubmit.own"), drHandler.Resubmit)
	// Draf (PR Draf): hanya pemiliknya; memakai izin yang ada.
	userOnly.POST("/review-documents/drafts", requirePermission("document_review.create.own"), drHandler.SaveDraft)
	userOnly.PUT("/review-documents/:id/draft", requirePermission("document_review.update.own"), drHandler.UpdateDraft)
	userOnly.DELETE("/review-documents/:id/attachments/:attachmentId", requirePermission("document_review.update.own"), drHandler.DeleteDraftAttachment)
	userOnly.POST("/review-documents/:id/submit", requirePermission("document_review.create.own"), drHandler.SubmitDraft)
	userOnly.GET("/agreement-documents", requirePermission("agreement_document.view.own"), agreementHandler.GetAll)
	userOnly.POST("/agreement-documents", requirePermission("agreement_document.create.own"), agreementHandler.Create)
	userOnly.POST("/agreement-documents/preview-draft", requirePermission("agreement_document.create.own"), agreementHandler.PreviewDraft)
	userOnly.GET("/agreement-documents/:id", requirePermission("agreement_document.view.own"), agreementHandler.GetByID)
	userOnly.GET("/agreement-documents/:id/history", requirePermission("agreement_document.view.own"), auditLogHandler.Riwayat("agreement_document", agreementHandler.BolehBuka))
	userOnly.GET("/agreement-documents/:id/schema", requirePermission("agreement_document.view.own"), agreementHandler.DocumentSchema)
	userOnly.PUT("/agreement-documents/:id", requirePermission("agreement_document.update.own"), agreementHandler.Update)
	userOnly.DELETE("/agreement-documents/:id", requirePermission("agreement_document.delete.own"), agreementHandler.Delete)
	userOnly.POST("/agreement-documents/:id/resubmit", requirePermission("agreement_document.resubmit.own"), agreementHandler.Resubmit)
	// Draf (PR Draf): hanya pemiliknya; memakai izin yang ada.
	userOnly.POST("/agreement-documents/drafts", requirePermission("agreement_document.create.own"), agreementHandler.SaveDraft)
	userOnly.PUT("/agreement-documents/:id/draft", requirePermission("agreement_document.update.own"), agreementHandler.UpdateDraft)
	userOnly.DELETE("/agreement-documents/:id/attachments/:attachmentId", requirePermission("agreement_document.update.own"), agreementHandler.DeleteDraftAttachment)
	userOnly.POST("/agreement-documents/:id/submit", requirePermission("agreement_document.create.own"), agreementHandler.SubmitDraft)
	userOnly.GET("/agreement-documents/:id/preview", requirePermission("agreement_document.preview.own"), agreementHandler.Preview)
	userOnly.GET("/agreement-documents/:id/pdf", requirePermission("agreement_document.download_pdf.own"), agreementHandler.DownloadPDF)
	userOnly.GET("/agreement-documents/:id/attachments/:attachmentId", requirePermission("agreement_document.download_attachment.own"), agreementHandler.DownloadAttachment)

	registerLegalCaseRoutes(userOnly, legalCaseHandler, auditLogHandler, requirePermission)

	// ── Admin only ─────────────────────────────────────────────────────────────
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(cfg, userRepo), middleware.PermissionContextMiddleware(permissionSvc), middleware.RoleMiddleware("ADMIN"), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	admin.GET("/dashboard/stats", dashHandler.AdminStats)
	admin.GET("/dashboard/recent", dashHandler.AdminRecent)

	admin.GET("/legal-opinions/download", loHandler.Download)
	admin.GET("/legal-opinions", loHandler.GetAll)
	admin.GET("/legal-opinions/:id", loHandler.GetByID)
	admin.GET("/legal-opinions/:id/history", auditLogHandler.Riwayat("legal_opinion", loHandler.BolehBuka))
	admin.GET("/review-documents/download", drHandler.Download)
	admin.GET("/review-documents", drHandler.GetAll)
	admin.GET("/review-documents/:id", drHandler.GetByID)
	admin.GET("/review-documents/:id/history", auditLogHandler.Riwayat("document_review", drHandler.BolehBuka))

	registerLegalCaseRoutes(admin, legalCaseHandler, auditLogHandler, requirePermission)

	admin.GET("/divisions", divisionHandler.GetAll)
	admin.POST("/divisions", divisionHandler.Create)
	admin.GET("/divisions/:id", divisionHandler.GetByID)
	admin.PUT("/divisions/:id", divisionHandler.Update)
	admin.DELETE("/divisions/:id", divisionHandler.Delete)
	admin.POST("/divisions/import", requirePermission("master_data.view"), divisionHandler.ImportDivisions)
	admin.GET("/divisions/import/template", requirePermission("master_data.view"), divisionHandler.DownloadDivisionTemplate)

	admin.GET("/companies", companyHandler.GetAll)
	admin.GET("/companies/:id", companyHandler.GetByID)
	admin.POST("/companies", companyHandler.Create)
	admin.PUT("/companies/:id", companyHandler.Update)
	admin.DELETE("/companies/:id", companyHandler.Delete)
	admin.GET("/companies/:id/make-parent/preview", requirePermission("company.set_parent.all"), companyHandler.PreviewMakeParent)
	admin.POST("/companies/:id/make-parent", requirePermission("company.set_parent.all"), companyHandler.MakeParent)
	admin.POST("/companies/import", requirePermission("master_data.view"), companyHandler.ImportCompanies)
	admin.GET("/companies/import/template", requirePermission("master_data.view"), companyHandler.DownloadCompanyTemplate)

	admin.GET("/purpose-types", purposeTypeHandler.GetAll)
	admin.GET("/purpose-types/:id", purposeTypeHandler.GetByID)
	admin.POST("/purpose-types", purposeTypeHandler.Create)
	admin.PUT("/purpose-types/:id", purposeTypeHandler.Update)
	admin.DELETE("/purpose-types/:id", purposeTypeHandler.Delete)
	admin.POST("/purpose-types/import", requirePermission("master_data.view"), purposeTypeHandler.ImportPurposeTypes)
	admin.GET("/purpose-types/import/template", requirePermission("master_data.view"), purposeTypeHandler.DownloadPurposeTypeTemplate)

	admin.GET("/case-types", caseTypeHandler.GetAll)
	admin.GET("/case-types/:id", caseTypeHandler.GetByID)
	admin.POST("/case-types", caseTypeHandler.Create)
	admin.PUT("/case-types/:id", caseTypeHandler.Update)
	admin.DELETE("/case-types/:id", caseTypeHandler.Delete)
	admin.POST("/case-types/import", requirePermission("master_data.view"), caseTypeHandler.ImportCaseTypes)
	admin.GET("/case-types/import/template", requirePermission("master_data.view"), caseTypeHandler.DownloadCaseTypeTemplate)

	admin.GET("/case-categories", caseCategoryHandler.GetAll)
	admin.GET("/case-categories/:id", caseCategoryHandler.GetByID)
	admin.POST("/case-categories", caseCategoryHandler.Create)
	admin.PUT("/case-categories/:id", caseCategoryHandler.Update)
	admin.DELETE("/case-categories/:id", caseCategoryHandler.Delete)
	admin.POST("/case-categories/import", requirePermission("master_data.view"), caseCategoryHandler.ImportCaseCategories)
	admin.GET("/case-categories/import/template", requirePermission("master_data.view"), caseCategoryHandler.DownloadCaseCategoryTemplate)

	admin.GET("/document-types", documentTypeHandler.GetAll)
	admin.GET("/document-types/:id", documentTypeHandler.GetByID)
	admin.POST("/document-types", documentTypeHandler.Create)
	admin.PUT("/document-types/:id", documentTypeHandler.Update)
	admin.DELETE("/document-types/:id", documentTypeHandler.Delete)
	admin.POST("/document-types/import", requirePermission("master_data.view"), documentTypeHandler.ImportDocumentTypes)
	admin.GET("/document-types/import/template", requirePermission("master_data.view"), documentTypeHandler.DownloadDocumentTypeTemplate)

	admin.GET("/regencies", legalCaseHandler.ListRegencies)
	admin.POST("/regencies", requirePermission("case_management.manage_reference"), legalCaseHandler.CreateRegency)
	admin.PUT("/regencies/:id", requirePermission("case_management.manage_reference"), legalCaseHandler.UpdateRegency)
	admin.DELETE("/regencies/:id", requirePermission("case_management.manage_reference"), legalCaseHandler.DeleteRegency)
	admin.GET("/cedants", legalCaseHandler.ListCedants)
	admin.POST("/cedants", legalCaseHandler.CreateCedant)
	admin.PUT("/cedants/:id", legalCaseHandler.UpdateCedant)
	admin.DELETE("/cedants/:id", legalCaseHandler.DeleteCedant)
	admin.POST("/cedants/import", requirePermission("case_management.manage_reference"), legalCaseHandler.ImportCedants)
	admin.GET("/cedants/import/template", requirePermission("case_management.manage_reference"), legalCaseHandler.DownloadCedantTemplate)

	admin.GET("/materials", materialHandler.GetAll)
	admin.GET("/materials/:id", materialHandler.GetByID)
	admin.POST("/materials", materialHandler.Create)
	admin.PUT("/materials/:id", materialHandler.Update)
	admin.DELETE("/materials/:id", materialHandler.Delete)
	admin.POST("/materials/import", requirePermission("legal_material.manage"), materialHandler.ImportLegalMaterials)
	admin.GET("/materials/import/template", requirePermission("legal_material.manage"), materialHandler.DownloadLegalMaterialTemplate)
	admin.PATCH("/legal-opinions/:id/status", loHandler.AdminUpdateStatus)
	admin.POST("/legal-opinions/:id/result", loHandler.AdminUploadResult)
	admin.GET("/legal-opinions/:id/pdf", loHandler.GeneratePDF)
	admin.PATCH("/review-documents/:id/status", drHandler.AdminUpdateStatus)
	admin.POST("/review-documents/:id/result", drHandler.AdminUploadResult)
	admin.GET("/review-documents/:id/pdf", requirePermission("document_review.view.all"), drHandler.GeneratePDF)
	admin.GET("/agreement-documents", requirePermission("agreement_document.view.all"), agreementHandler.GetAll)
	admin.GET("/agreement-documents/:id", requirePermission("agreement_document.view.all"), agreementHandler.GetByID)
	admin.GET("/agreement-documents/:id/history", requirePermission("agreement_document.view.all"), auditLogHandler.Riwayat("agreement_document", agreementHandler.BolehBuka))
	admin.GET("/agreement-documents/:id/schema", requirePermission("agreement_document.view.all"), agreementHandler.DocumentSchema)
	admin.PATCH("/agreement-documents/:id/meta", requirePermission("agreement_document.update_meta.all"), agreementHandler.UpdateMeta)
	admin.PATCH("/agreement-documents/:id/status", requirePermission("agreement_document.update_status.all"), agreementHandler.UpdateStatus)
	admin.GET("/agreement-documents/:id/preview", requirePermission("agreement_document.preview.all"), agreementHandler.Preview)
	admin.GET("/agreement-documents/:id/pdf", requirePermission("agreement_document.download_pdf.all"), agreementHandler.DownloadPDF)
	admin.GET("/agreement-documents/:id/docx", requirePermission("agreement_document.download_docx.all"), agreementHandler.DownloadDOCX)
	admin.GET("/agreement-documents/:id/attachments/:attachmentId", requirePermission("agreement_document.download_attachment.all"), agreementHandler.DownloadAttachment)
	admin.GET("/agreement-company-master", requirePermission("agreement_company_master.view.all"), agreementHandler.GetMaster)
	admin.PUT("/agreement-company-master", requirePermission("agreement_company_master.manage.all"), agreementHandler.UpdateMaster)
	admin.GET("/agreement-signatories", requirePermission("agreement_company_master.view.all"), agreementHandler.ListSignatories)
	admin.POST("/agreement-signatories", requirePermission("agreement_company_master.manage.all"), agreementHandler.CreateSignatory)
	admin.PUT("/agreement-signatories/:id", requirePermission("agreement_company_master.manage.all"), agreementHandler.UpdateSignatory)
	admin.DELETE("/agreement-signatories/:id", requirePermission("agreement_company_master.manage.all"), agreementHandler.DeleteSignatory)
	registerAgreementTemplateRoutes(admin, requirePermission, agreementTemplateHandler)
	admin.GET("/agreement-types", requirePermission("agreement_type.manage.all"), agreementTypeHandler.List)
	admin.POST("/agreement-types", requirePermission("agreement_type.manage.all"), agreementTypeHandler.Create)
	admin.PUT("/agreement-types/:code", requirePermission("agreement_type.manage.all"), agreementTypeHandler.Update)

	admin.GET("/users", userHandler.GetAll)
	admin.POST("/users", userHandler.Create)
	admin.GET("/permissions", requirePermission("user_management.manage_permissions"), permissionHandler.GetCatalog)
	admin.GET("/users/:id/permissions", requirePermission("user_management.manage_permissions"), permissionHandler.GetUserAccess)
	admin.PUT("/users/:id/permissions", requirePermission("user_management.manage_permissions"), permissionHandler.UpdateUserAccess)
	admin.PUT("/users/:id", userHandler.Update)
	admin.DELETE("/users/:id", requirePermission("user_management.delete"), userHandler.Delete)
	admin.PATCH("/users/:id/status", userHandler.UpdateStatus)
	admin.POST("/users/:id/reset-password", userHandler.ResetPassword)

	admin.GET("/audit-logs", auditLogHandler.GetAll)
	admin.GET("/audit-logs/actors", requirePermission("audit_log.view"), auditLogHandler.GetActors)
	admin.GET("/notification-settings", notificationSettingHandler.GetAll)
	admin.PUT("/notification-settings/:id", notificationSettingHandler.Update)

	admin.GET("/reports/legal-cases", requirePermission("report.legal_case.view"), reportHandler.GetLegalCaseReport)
	admin.GET("/reports/legal-opinions", requirePermission("report.legal_opinion.view"), reportHandler.GetLegalOpinionReport)
	admin.GET("/reports/document-reviews", requirePermission("report.document_review.view"), reportHandler.GetDocumentReviewReport)

	// ── Legal ─────────────────────────────────────────────────────────────────
	legal := api.Group("/legal")
	legal.Use(middleware.AuthMiddleware(cfg, userRepo), middleware.PermissionContextMiddleware(permissionSvc), middleware.RoleMiddleware("LEGAL"), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	legal.GET("/dashboard/stats", requirePermission("dashboard.legal.view"), dashHandler.LegalStats)
	legal.GET("/dashboard/recent", requirePermission("dashboard.legal.view"), dashHandler.LegalRecent)

	legal.GET("/legal-opinions/download", requirePermission("legal_opinion.download.all", "legal_opinion.view.all"), loHandler.Download)
	legal.GET("/legal-opinions", requirePermission("legal_opinion.view.all"), loHandler.GetAll)
	legal.GET("/legal-opinions/:id", requirePermission("legal_opinion.view.all"), loHandler.GetByID)
	legal.GET("/legal-opinions/:id/history", requirePermission("legal_opinion.view.all"), auditLogHandler.Riwayat("legal_opinion", loHandler.BolehBuka))
	legal.PATCH("/legal-opinions/:id/status", requirePermission("legal_opinion.update_status.all"), loHandler.AdminUpdateStatus)
	legal.POST("/legal-opinions/:id/result", requirePermission("legal_opinion.upload_result.all"), loHandler.AdminUploadResult)
	legal.GET("/legal-opinions/:id/pdf", requirePermission("legal_opinion.download.all", "legal_opinion.view.all"), loHandler.GeneratePDF)
	legal.GET("/review-documents/download", requirePermission("document_review.download.all", "document_review.view.all"), drHandler.Download)
	legal.GET("/review-documents", requirePermission("document_review.view.all"), drHandler.GetAll)
	legal.GET("/review-documents/:id", requirePermission("document_review.view.all"), drHandler.GetByID)
	legal.GET("/review-documents/:id/history", requirePermission("document_review.view.all"), auditLogHandler.Riwayat("document_review", drHandler.BolehBuka))
	legal.PATCH("/review-documents/:id/status", requirePermission("document_review.update_status.all"), drHandler.AdminUpdateStatus)
	legal.POST("/review-documents/:id/result", requirePermission("document_review.upload_result.all"), drHandler.AdminUploadResult)
	legal.GET("/review-documents/:id/pdf", requirePermission("document_review.view.all"), drHandler.GeneratePDF)
	// Isian awal persetujuan (nomor, tempat, penandatangan); mengubah master tetap di /admin.
	legal.GET("/agreement-company-master", requirePermission("agreement_company_master.view.all"), agreementHandler.GetMaster)
	legal.GET("/agreement-signatories", requirePermission("agreement_company_master.view.all"), agreementHandler.ListSignatories)
	legal.GET("/agreement-documents", requirePermission("agreement_document.view.all"), agreementHandler.GetAll)
	legal.GET("/agreement-documents/:id", requirePermission("agreement_document.view.all"), agreementHandler.GetByID)
	legal.GET("/agreement-documents/:id/history", requirePermission("agreement_document.view.all"), auditLogHandler.Riwayat("agreement_document", agreementHandler.BolehBuka))
	legal.GET("/agreement-documents/:id/schema", requirePermission("agreement_document.view.all"), agreementHandler.DocumentSchema)
	legal.PATCH("/agreement-documents/:id/meta", requirePermission("agreement_document.update_meta.all"), agreementHandler.UpdateMeta)
	legal.PATCH("/agreement-documents/:id/status", requirePermission("agreement_document.update_status.all"), agreementHandler.UpdateStatus)
	legal.GET("/agreement-documents/:id/preview", requirePermission("agreement_document.preview.all"), agreementHandler.Preview)
	legal.GET("/agreement-documents/:id/pdf", requirePermission("agreement_document.download_pdf.all"), agreementHandler.DownloadPDF)
	legal.GET("/agreement-documents/:id/docx", requirePermission("agreement_document.download_docx.all"), agreementHandler.DownloadDOCX)
	legal.GET("/agreement-documents/:id/attachments/:attachmentId", requirePermission("agreement_document.download_attachment.all"), agreementHandler.DownloadAttachment)
	registerAgreementTemplateRoutes(legal, requirePermission, agreementTemplateHandler)

	registerLegalCaseRoutes(legal, legalCaseHandler, auditLogHandler, requirePermission)

	legal.GET("/audit-logs", requirePermission("audit_log.view"), auditLogHandler.GetAll)
	legal.GET("/audit-logs/actors", requirePermission("audit_log.view"), auditLogHandler.GetActors)

	legal.GET("/materials", requirePermission("legal_material.view", "legal_material.manage"), materialHandler.GetAll)
	legal.GET("/materials/:id", requirePermission("legal_material.view", "legal_material.manage"), materialHandler.GetByID)
	legal.POST("/materials", requirePermission("legal_material.manage"), materialHandler.Create)
	legal.PUT("/materials/:id", requirePermission("legal_material.manage"), materialHandler.Update)
	legal.DELETE("/materials/:id", requirePermission("legal_material.manage"), materialHandler.Delete)

	legal.GET("/reports/legal-cases", requirePermission("report.legal_case.view"), reportHandler.GetLegalCaseReport)
	legal.GET("/reports/legal-opinions", requirePermission("report.legal_opinion.view"), reportHandler.GetLegalOpinionReport)
	legal.GET("/reports/document-reviews", requirePermission("report.document_review.view"), reportHandler.GetDocumentReviewReport)

	// ─── Legal AU ─────────────────────────────────────────────────────────────
	legalAU := api.Group("/legal-au")
	legalAU.Use(middleware.AuthMiddleware(cfg, userRepo), middleware.PermissionContextMiddleware(permissionSvc), middleware.RoleMiddleware("LEGAL_AU"), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	legalAU.GET("/cases", requirePermission("case_management.view"), legalCaseHandler.GetAll)
	legalAU.POST("/cases", requirePermission("case_management.create"), legalCaseHandler.Create)
	// Daftar pilihan yang dimuat form perkara
	legalAU.GET("/cases/regencies", requirePermission("case_management.view", "case_management.create"), legalCaseHandler.ListRegencies)
	legalAU.GET("/cases/cedants", requirePermission("case_management.view", "case_management.create"), legalCaseHandler.ListCedants)
	legalAU.GET("/cases/:id", requirePermission("case_management.view"), legalCaseHandler.GetByID)
	legalAU.GET("/cases/:id/history", requirePermission("case_management.view"), auditLogHandler.Riwayat("legal_case", legalCaseHandler.BolehBuka))
	legalAU.PATCH("/cases/:id/status", requirePermission("case_management.update_status"), legalCaseHandler.UpdateStatus)
	legalAU.POST("/cases/:id/chronology", requirePermission("case_management.manage_chronology"), legalCaseHandler.CreateChronology)
	legalAU.PUT("/cases/:id/chronology/:chronId", requirePermission("case_management.manage_chronology"), legalCaseHandler.UpdateChronology)
	legalAU.DELETE("/cases/:id/chronology/:chronId", requirePermission("case_management.manage_chronology"), legalCaseHandler.DeleteChronology)
	legalAU.POST("/cases/:id/chronology/import", requirePermission("case_management.manage_chronology"), legalCaseHandler.ImportChronologies)
	legalAU.GET("/cases/chronology/template", requirePermission("case_management.view"), legalCaseHandler.DownloadChronologyTemplate)
	legalAU.POST("/cases/:id/upload-document", requirePermission("case_management.manage_document"), legalCaseHandler.UploadDocument)
	legalAU.DELETE("/cases/:id/document", requirePermission("case_management.manage_document"), legalCaseHandler.DeleteDocument)
	legalAU.GET("/cases/download", requirePermission("case_management.view", "case_management.manage_document"), legalCaseHandler.Download)
	legalAU.POST("/cases/:id/photo", requirePermission("case_management.manage_document"), legalCaseHandler.UploadPhoto)
	legalAU.DELETE("/cases/:id/photo", requirePermission("case_management.manage_document"), legalCaseHandler.DeletePhoto)
	legalAU.GET("/cases/file", requirePermission("case_management.view"), legalCaseHandler.ViewFile)
	legalAU.GET("/cases/:id/pdf", requirePermission("case_management.view"), legalCaseHandler.GeneratePDF)

	legalAU.GET("/materials", requirePermission("legal_material.view", "legal_material.manage"), materialHandler.GetAll)
	legalAU.GET("/materials/:id", requirePermission("legal_material.view", "legal_material.manage"), materialHandler.GetByID)
	legalAU.POST("/materials", requirePermission("legal_material.manage"), materialHandler.Create)
	legalAU.PUT("/materials/:id", requirePermission("legal_material.manage"), materialHandler.Update)
	legalAU.DELETE("/materials/:id", requirePermission("legal_material.manage"), materialHandler.Delete)

	// ─── External ─────────────────────────────────────────────────────────────
	external := api.Group("/external")
	external.Use(middleware.AuthMiddleware(cfg, userRepo), middleware.PermissionContextMiddleware(permissionSvc), middleware.RoleMiddleware("EXTERNAL"), middleware.AuditMiddleware(auditLogSvc), middleware.CSRFProtection())

	registerLegalCaseRoutes(external, legalCaseHandler, auditLogHandler, requirePermission)

	// ── Public materials ──────────────────────────────────────────────────────
	protectedMaterials := api.Group("")
	protectedMaterials.Use(middleware.AuthMiddleware(cfg, userRepo), middleware.PermissionContextMiddleware(permissionSvc))
	protectedMaterials.GET("/materials", requirePermission("legal_material.view", "legal_material.manage"), materialHandler.GetAll)
	protectedMaterials.GET("/materials/:id", requirePermission("legal_material.view", "legal_material.manage"), materialHandler.GetByID)

	return r, nil
}

func registerAgreementTemplateRoutes(group *gin.RouterGroup, requirePermission func(...string) gin.HandlerFunc, h *handler.AgreementTemplateHandler) {
	group.GET("/agreement-templates", requirePermission("agreement_template.view.all"), h.List)
	group.GET("/agreement-templates/placeholders", requirePermission("agreement_template.view.all"), h.Placeholders)
	group.GET("/agreement-templates/:id", requirePermission("agreement_template.view.all"), h.GetByID)
	group.GET("/agreement-templates/:id/preview", requirePermission("agreement_template.view.all"), h.Preview)
	group.GET("/agreement-templates/:id/download", requirePermission("agreement_template.view.all"), h.Download)
	group.POST("/agreement-templates", requirePermission("agreement_template.manage.all"), h.Upload)
	group.POST("/agreement-templates/new-version", requirePermission("agreement_template.manage.all"), h.NewVersion)
	group.POST("/agreement-templates/:id/activate", requirePermission("agreement_template.manage.all"), h.Activate)
	group.DELETE("/agreement-templates/:id", requirePermission("agreement_template.manage.all"), h.Delete)
}

func registerLegalCaseRoutes(group *gin.RouterGroup, legalCaseHandler *handler.LegalCaseHandler, auditLogHandler *handler.AuditLogHandler, requirePermission func(...string) gin.HandlerFunc) {
	group.GET("/legal-cases/regencies", requirePermission("case_management.view", "case_management.create"), legalCaseHandler.ListRegencies)
	group.GET("/legal-cases/cedants", requirePermission("case_management.view", "case_management.create"), legalCaseHandler.ListCedants)
	group.POST("/legal-cases/cedants", requirePermission("case_management.manage_reference"), legalCaseHandler.CreateCedant)
	group.PUT("/legal-cases/cedants/:id", requirePermission("case_management.manage_reference"), legalCaseHandler.UpdateCedant)
	group.DELETE("/legal-cases/cedants/:id", requirePermission("case_management.manage_reference"), legalCaseHandler.DeleteCedant)
	group.POST("/legal-cases/cedants/import", requirePermission("case_management.manage_reference"), legalCaseHandler.ImportCedants)
	group.GET("/legal-cases/cedants/import/template", requirePermission("case_management.manage_reference"), legalCaseHandler.DownloadCedantTemplate)
	group.POST("/legal-cases/regencies/import", requirePermission("case_management.manage_reference"), legalCaseHandler.ImportRegencies)
	group.GET("/legal-cases/regencies/import/template", requirePermission("case_management.manage_reference"), legalCaseHandler.DownloadRegencyTemplate)
	group.GET("/legal-cases/download", requirePermission("case_management.view", "case_management.manage_document"), legalCaseHandler.Download)
	group.GET("/legal-cases/latest", requirePermission("case_management.view"), legalCaseHandler.GetLatest)
	group.GET("/legal-cases", requirePermission("case_management.view"), legalCaseHandler.GetAll)
	group.POST("/legal-cases", requirePermission("case_management.create"), legalCaseHandler.Create)
	group.GET("/legal-cases/:id", requirePermission("case_management.view"), legalCaseHandler.GetByID)
	group.GET("/legal-cases/:id/history", requirePermission("case_management.view"), auditLogHandler.Riwayat("legal_case", legalCaseHandler.BolehBuka))
	group.PUT("/legal-cases/:id", requirePermission("case_management.update"), legalCaseHandler.Update)
	group.DELETE("/legal-cases/:id", requirePermission("case_management.delete"), legalCaseHandler.Delete)
	group.POST("/legal-cases/:id/upload-document", requirePermission("case_management.manage_document"), legalCaseHandler.UploadDocument)
	group.DELETE("/legal-cases/:id/document", requirePermission("case_management.manage_document"), legalCaseHandler.DeleteDocument)
	group.POST("/legal-cases/:id/photo", requirePermission("case_management.manage_document"), legalCaseHandler.UploadPhoto)
	group.DELETE("/legal-cases/:id/photo", requirePermission("case_management.manage_document"), legalCaseHandler.DeletePhoto)
	group.GET("/legal-cases/:id/chronology", requirePermission("case_management.view", "case_management.manage_chronology"), legalCaseHandler.ListChronologies)
	group.POST("/legal-cases/:id/chronology", requirePermission("case_management.manage_chronology"), legalCaseHandler.CreateChronology)
	group.PUT("/legal-cases/:id/chronology/:chronId", requirePermission("case_management.manage_chronology"), legalCaseHandler.UpdateChronology)
	group.DELETE("/legal-cases/:id/chronology/:chronId", requirePermission("case_management.manage_chronology"), legalCaseHandler.DeleteChronology)
	group.POST("/legal-cases/:id/chronology/import", requirePermission("case_management.manage_chronology"), legalCaseHandler.ImportChronologies)
	group.GET("/legal-cases/chronology/template", requirePermission("case_management.view"), legalCaseHandler.DownloadChronologyTemplate)
	group.GET("/legal-cases/file", requirePermission("case_management.view"), legalCaseHandler.ViewFile)
	group.GET("/legal-cases/:id/pdf", requirePermission("case_management.view"), legalCaseHandler.GeneratePDF)
}
