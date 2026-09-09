package handler

import (
	"bytes"
	"fmt"
	"net/http"

	"legal-riu-portal/internal/dto"
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
	utils.OK(c, "Success", h.svc.Placeholders())
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
	v, e := h.svc.Upload(c, middleware.GetUserID(c), req, file)
	if e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	utils.Created(c, "Template berhasil diunggah dan lolos uji coba", v)
}

func (h *AgreementTemplateHandler) Activate(c *gin.Context) {
	v, e := h.svc.Activate(c, c.Param("id"), middleware.GetUserID(c))
	if e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
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
		map[string]string{"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, name)})
}

func (h *AgreementTemplateHandler) Preview(c *gin.Context) {
	data, e := h.svc.Preview(c, c.Param("id"))
	if e != nil {
		utils.InternalError(c, e.Error())
		return
	}
	c.DataFromReader(http.StatusOK, int64(len(data)), "application/pdf", bytes.NewReader(data),
		map[string]string{"Content-Disposition": "inline; filename=template-preview.pdf"})
}
