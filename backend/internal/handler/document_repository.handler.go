package handler

import (
	"net/http"

	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

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
	featureCode := c.Query("feature_code")
	search := c.Query("search")
	items, err := h.svc.GetAllDocuments(access, featureCode, search)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", items)
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
		"Content-Disposition": `attachment; filename="` + fileName + `"`,
	})
}
