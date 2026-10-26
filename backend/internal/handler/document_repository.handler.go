package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

var jenisDokumenRepositori = map[string]bool{
	"legal_opinion_attachments":   true,
	"legal_opinion_results":       true,
	"document_review_attachments": true,
	"document_review_results":     true,
	"agreement_attachments":       true,
	"agreement_documents":         true,
	"legal_cases":                 true,
	"legal_cases_photo":           true,
	"case_chronologies":           true,
}

type RepositoryDocumentHandler struct {
	svc      service.RepositoryDocumentService
	userRepo repository.UserRepository
}

func NewRepositoryDocumentHandler(svc service.RepositoryDocumentService, userRepo repository.UserRepository) *RepositoryDocumentHandler {
	return &RepositoryDocumentHandler{svc: svc, userRepo: userRepo}
}

// access menyusun apa yang boleh dilihat si pemanggil dari permission yang
// sama dengan yang dipakai modul asal tiap dokumen, supaya repositori tidak
// menjadi jalan pintas ke dokumen yang di modul asalnya tertutup baginya.
func (h *RepositoryDocumentHandler) access(c *gin.Context) (repository.DocumentAccess, bool) {
	reach := func(feature, action string) string {
		switch {
		case middleware.HasPermission(c, feature+"."+action+".all"):
			return repository.ReachAll
		case middleware.HasPermission(c, feature+"."+action+".own"):
			return repository.ReachOwn
		}
		return ""
	}
	access := repository.DocumentAccess{
		UserID:               mustParseUUID(middleware.GetUserID(c)),
		Opinions:             reach("legal_opinion", "view"),
		Reviews:              reach("document_review", "view"),
		AgreementAttachments: reach("agreement_document", "download_attachment"),
		AgreementPDFs:        reach("agreement_document", "download_pdf"),
	}
	// DOCX final perjanjian tidak punya izin "milik sendiri".
	if middleware.HasPermission(c, "agreement_document.download_docx.all") {
		access.AgreementDOCX = repository.ReachAll
	}
	if middleware.HasPermission(c, "case_management.view") {
		company, ok := companyScope(c, h.userRepo)
		if !ok {
			return access, false
		}
		access.Cases, access.CaseCompany = true, company
	}
	return access, true
}

func (h *RepositoryDocumentHandler) GetAll(c *gin.Context) {
	access, ok := h.access(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		limit = 100
	}
	filters := dto.RepositoryDocumentFilters{
		Page:        page,
		Limit:       limit,
		FeatureCode: strings.TrimSpace(c.Query("feature_code")),
		SourceType:  strings.TrimSpace(c.Query("source_type")),
		MIMEType:    strings.TrimSpace(c.Query("mime_type")),
		Search:      strings.TrimSpace(c.Query("search")),
	}
	if filters.SourceType != "" && !jenisDokumenRepositori[filters.SourceType] {
		utils.BadRequest(c, "Jenis dokumen tidak valid", nil)
		return
	}
	if filters.MIMEType != "" {
		if filters.SourceType != "agreement_documents" ||
			(filters.MIMEType != "application/pdf" &&
				filters.MIMEType != "application/vnd.openxmlformats-officedocument.wordprocessingml.document") {
			utils.BadRequest(c, "Tipe file tidak valid", nil)
			return
		}
	}
	wib := time.FixedZone("WIB", 7*60*60)
	if value := c.Query("date_from"); value != "" {
		tanggal, err := time.ParseInLocation("2006-01-02", value, wib)
		if err != nil {
			utils.BadRequest(c, "Tanggal awal tidak valid", nil)
			return
		}
		filters.DateFrom = &tanggal
	}
	if value := c.Query("date_to"); value != "" {
		tanggal, err := time.ParseInLocation("2006-01-02", value, wib)
		if err != nil {
			utils.BadRequest(c, "Tanggal akhir tidak valid", nil)
			return
		}
		tanggal = tanggal.Add(24*time.Hour - time.Nanosecond)
		filters.DateTo = &tanggal
	}
	if filters.DateFrom != nil && filters.DateTo != nil && filters.DateFrom.After(*filters.DateTo) {
		utils.BadRequest(c, "Rentang tanggal tidak valid", nil)
		return
	}
	result, err := h.svc.GetAllDocuments(access, filters)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", result)
}

func (h *RepositoryDocumentHandler) GetByID(c *gin.Context) {
	access, ok := h.access(c)
	if !ok {
		return
	}
	id := c.Param("id")
	item, err := h.svc.GetDocumentByID(access, id)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.OK(c, "Success", item)
}

func (h *RepositoryDocumentHandler) Download(c *gin.Context) {
	access, ok := h.access(c)
	if !ok {
		return
	}
	id := c.Param("id")
	if id == "" {
		utils.BadRequest(c, "ID tidak boleh kosong", nil)
		return
	}
	obj, fileName, err := h.svc.DownloadDocument(access, id)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	defer obj.Close()
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", obj, map[string]string{
		"Content-Disposition": utils.HeaderUnduhan("attachment", fileName),
	})
}
