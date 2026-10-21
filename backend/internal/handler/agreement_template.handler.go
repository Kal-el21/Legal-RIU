package handler

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type AgreementTemplateHandler struct {
	svc service.AgreementTemplateService
}

func NewAgreementTemplateHandler(s service.AgreementTemplateService) *AgreementTemplateHandler {
	return &AgreementTemplateHandler{s}
}

func (h *AgreementTemplateHandler) List(c *gin.Context) {
	items, e := h.svc.List(c.Query("code"))
	if e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	utils.OK(c, "Success", items)
}

func (h *AgreementTemplateHandler) GetByID(c *gin.Context) {
	v, e := h.svc.Get(c.Param("id"))
	if e != nil {
		utils.NotFound(c, e.Error())
		return
	}
	utils.OK(c, "Success", v)
}

func (h *AgreementTemplateHandler) Placeholders(c *gin.Context) {
	v, e := h.svc.Placeholders(c.DefaultQuery("code", "PKS"))
	if e != nil {
		utils.NotFound(c, e.Error())
		return
	}
	utils.OK(c, "Success", v)
}

func (h *AgreementTemplateHandler) Upload(c *gin.Context) {
	file, e := c.FormFile("file")
	if e != nil {
		utils.BadRequest(c, "File template wajib diunggah", nil)
		return
	}
	req := dto.UploadAgreementTemplateRequest{
		Code:          c.PostForm("code"),
		Name:          c.PostForm("name"),
		Note:          c.PostForm("note"),
		EffectiveDate: c.PostForm("effective_date"),
	}
	v, hasil, e := h.svc.Upload(c, middleware.GetUserID(c), req, file)
	h.versiBaru(c, v, hasil, e, "Template berhasil diunggah dan lolos uji coba")
}

// NewVersion: "Simpan sebagai versi baru" dengan DOCX versi terakhir.
func (h *AgreementTemplateHandler) NewVersion(c *gin.Context) {
	var req dto.UploadAgreementTemplateRequest
	if e := c.ShouldBindJSON(&req); e != nil {
		utils.BadRequest(c, "Data tidak valid", e.Error())
		return
	}
	v, hasil, e := h.svc.NewVersion(c, middleware.GetUserID(c), req)
	h.versiBaru(c, v, hasil, e, "Versi baru berhasil dibuat dan lolos uji coba")
}

// versiBaru menjawab hasil pembuatan versi: penolakan pencocokan membawa daftar
// placeholder tanpa field; keberhasilan membawa peringatan pencocokan.
func (h *AgreementTemplateHandler) versiBaru(c *gin.Context, v *entity.AgreementTemplate, hasil dto.AgreementTemplateValidation, e error, pesan string) {
	var gagal *service.GagalPencocokan
	if errors.As(e, &gagal) {
		utils.BadRequest(c, e.Error(), gagal.Hasil)
		return
	}
	if e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionFileUpload, "agreement_template", v.ID)
	c.Set("audit_description", fmt.Sprintf("Template %s versi %d dibuat", v.Code, v.Version))
	utils.Created(c, pesan, gin.H{"template": v, "validation": hasil})
}

func (h *AgreementTemplateHandler) Delete(c *gin.Context) {
	e := h.svc.Delete(c.Param("id"))
	if errors.Is(e, service.ErrTemplateTidakAda) {
		utils.NotFound(c, e.Error())
		return
	}
	if e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionDelete, "agreement_template", c.Param("id"))
	utils.OK(c, "Versi template dihapus", nil)
}

func (h *AgreementTemplateHandler) Activate(c *gin.Context) {
	v, e := h.svc.Activate(c, c.Param("id"), middleware.GetUserID(c))
	if e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionStatusChange, "agreement_template", v.ID)
	c.Set("audit_description", fmt.Sprintf("Template %s versi %d diaktifkan", v.Code, v.Version))
	utils.OK(c, "Template diaktifkan", v)
}

func (h *AgreementTemplateHandler) Download(c *gin.Context) {
	data, name, e := h.svc.Download(c, c.Param("id"))
	if e != nil {
		utils.NotFound(c, e.Error())
		return
	}
	c.DataFromReader(http.StatusOK, int64(len(data)),
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		bytes.NewReader(data),
		map[string]string{"Content-Disposition": utils.HeaderUnduhan("attachment", name)})
}

func (h *AgreementTemplateHandler) Preview(c *gin.Context) {
	data, e := h.svc.Preview(c, c.Param("id"))
	if errors.Is(e, service.ErrTemplateLama) {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	if e != nil {
		utils.InternalError(c, e.Error())
		return
	}
	c.DataFromReader(http.StatusOK, int64(len(data)), "application/pdf", bytes.NewReader(data),
		map[string]string{"Content-Disposition": "inline; filename=template-preview.pdf"})
}
