package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"
	"mime/multipart"
	"net/http"
	"strings"
)

type AgreementDocumentHandler struct {
	svc service.AgreementDocumentService
}

func NewAgreementDocumentHandler(s service.AgreementDocumentService) *AgreementDocumentHandler {
	return &AgreementDocumentHandler{s}
}
func (h *AgreementDocumentHandler) ListTypes(c *gin.Context) {
	utils.OK(c, "Success", h.svc.ListTypes())
}
func (h *AgreementDocumentHandler) GetType(c *gin.Context) {
	v, e := h.svc.GetType(c.Param("code"))
	if e != nil {
		utils.NotFound(c, e.Error())
		return
	}
	utils.OK(c, "Success", v)
}
func (h *AgreementDocumentHandler) Create(c *gin.Context) {
	var req dto.CreateAgreementDocumentRequest
	var files []*multipart.FileHeader
	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/") {
		if e := json.Unmarshal([]byte(c.PostForm("data")), &req); e != nil {
			utils.BadRequest(c, "Data tidak valid", e.Error())
			return
		}
		if f, _ := c.MultipartForm(); f != nil {
			files = f.File["attachments"]
		}
	} else if e := c.ShouldBindJSON(&req); e != nil {
		utils.BadRequest(c, "Data tidak valid", e.Error())
		return
	}
	v, e := h.svc.Create(middleware.GetUserID(c), req, files)
	if e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	utils.Created(c, "Pengajuan berhasil dibuat", v)
}
func (h *AgreementDocumentHandler) GetAll(c *gin.Context) {
	var q dto.AgreementListQuery
	if e := c.ShouldBindQuery(&q); e != nil {
		utils.BadRequest(c, "Query tidak valid", nil)
		return
	}
	all := middleware.RoleWithAllAccess(c, "agreement_document.view.all") != "USER"
	items, total, e := h.svc.GetAll(middleware.GetUserID(c), all, q)
	if e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	utils.OK(c, "Success", gin.H{"items": items, "total": total, "page": q.Page, "limit": q.Limit})
}
func (h *AgreementDocumentHandler) GetByID(c *gin.Context) {
	all := middleware.RoleWithAllAccess(c, "agreement_document.view.all") != "USER"
	v, e := h.svc.GetByID(c.Param("id"), middleware.GetUserID(c), all)
	if e != nil {
		utils.NotFound(c, e.Error())
		return
	}
	utils.OK(c, "Success", v)
}
func (h *AgreementDocumentHandler) Update(c *gin.Context) {
	var req dto.UpdateAgreementDocumentRequest
	if e := c.ShouldBindJSON(&req); e != nil {
		utils.BadRequest(c, "Data tidak valid", e.Error())
		return
	}
	v, e := h.svc.Update(c.Param("id"), middleware.GetUserID(c), req)
	if e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	utils.OK(c, "Pengajuan diperbarui", v)
}
func (h *AgreementDocumentHandler) Delete(c *gin.Context) {
	if e := h.svc.Delete(c.Param("id"), middleware.GetUserID(c)); e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	utils.OK(c, "Pengajuan dihapus", nil)
}
func (h *AgreementDocumentHandler) Resubmit(c *gin.Context) {
	var files []*multipart.FileHeader
	if f, _ := c.MultipartForm(); f != nil {
		files = f.File["attachments"]
	}
	v, e := h.svc.Resubmit(c.Param("id"), middleware.GetUserID(c), files)
	if e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	utils.OK(c, "Pengajuan dikirim ulang", v)
}
func (h *AgreementDocumentHandler) UpdateMeta(c *gin.Context) {
	var req dto.AgreementMetaRequest
	if e := c.ShouldBindJSON(&req); e != nil {
		utils.BadRequest(c, "Data tidak valid", e.Error())
		return
	}
	v, e := h.svc.UpdateMeta(c.Param("id"), req)
	if e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	utils.OK(c, "Metadata diperbarui", v)
}
func (h *AgreementDocumentHandler) UpdateStatus(c *gin.Context) {
	var req dto.AgreementStatusRequest
	if e := c.ShouldBindJSON(&req); e != nil {
		utils.BadRequest(c, "Data tidak valid", e.Error())
		return
	}
	v, e := h.svc.UpdateStatus(c, c.Param("id"), middleware.GetUserID(c), req)
	if e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	utils.OK(c, "Status diperbarui", v)
}
func (h *AgreementDocumentHandler) Preview(c *gin.Context) {
	all := middleware.RoleWithAllAccess(c, "agreement_document.preview.all") != "USER"
	b, e := h.svc.Preview(c, c.Param("id"), middleware.GetUserID(c), all)
	if e != nil {
		utils.InternalError(c, e.Error())
		return
	}
	c.DataFromReader(http.StatusOK, int64(len(b)), "application/pdf", bytes.NewReader(b), map[string]string{"Content-Disposition": "inline; filename=preview.pdf"})
}
func (h *AgreementDocumentHandler) DownloadPDF(c *gin.Context)  { h.downloadGenerated(c, "pdf") }
func (h *AgreementDocumentHandler) DownloadDOCX(c *gin.Context) { h.downloadGenerated(c, "docx") }
func (h *AgreementDocumentHandler) downloadGenerated(c *gin.Context, kind string) {
	all := middleware.RoleWithAllAccess(c, "agreement_document.download_"+kind+".all") != "USER"
	o, n, e := h.svc.GetGeneratedFile(c.Param("id"), middleware.GetUserID(c), all, kind)
	if e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	defer o.Close()
	ct := "application/pdf"
	if kind == "docx" {
		ct = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}
	c.DataFromReader(http.StatusOK, -1, ct, o, map[string]string{"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, n)})
}
func (h *AgreementDocumentHandler) DownloadAttachment(c *gin.Context) {
	all := middleware.RoleWithAllAccess(c, "agreement_document.download_attachment.all") != "USER"
	o, n, ct, e := h.svc.GetAttachment(c.Param("id"), c.Param("attachmentId"), middleware.GetUserID(c), all)
	if e != nil {
		utils.NotFound(c, e.Error())
		return
	}
	defer o.Close()
	if ct == "" {
		ct = "application/octet-stream"
	}
	c.DataFromReader(http.StatusOK, -1, ct, o, map[string]string{"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, n)})
}
func (h *AgreementDocumentHandler) GetMaster(c *gin.Context) {
	v, e := h.svc.GetMaster()
	if e != nil {
		utils.NotFound(c, "Master belum tersedia")
		return
	}
	utils.OK(c, "Success", v)
}
func (h *AgreementDocumentHandler) UpdateMaster(c *gin.Context) {
	var req dto.AgreementCompanyMasterRequest
	if e := c.ShouldBindJSON(&req); e != nil {
		utils.BadRequest(c, "Data tidak valid", e.Error())
		return
	}
	v, e := h.svc.UpdateMaster(req)
	if e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	utils.OK(c, "Master diperbarui", v)
}
