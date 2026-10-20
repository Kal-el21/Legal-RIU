package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"
)

type AgreementTypeHandler struct{ svc service.AgreementTypeService }

func NewAgreementTypeHandler(svc service.AgreementTypeService) *AgreementTypeHandler {
	return &AgreementTypeHandler{svc}
}

func (h *AgreementTypeHandler) List(c *gin.Context) {
	items, e := h.svc.List()
	if e != nil {
		utils.InternalError(c, "Gagal memuat jenis dokumen")
		return
	}
	utils.OK(c, "Success", items)
}

func (h *AgreementTypeHandler) Create(c *gin.Context) {
	var in service.AgreementTypeInput
	if e := c.ShouldBindJSON(&in); e != nil {
		utils.BadRequest(c, "Data tidak valid", e.Error())
		return
	}
	v, e := h.svc.Create(in)
	if e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "agreement_type", v.ID)
	c.Set("audit_description", "Jenis dokumen "+v.Code+" dibuat")
	utils.Created(c, "Jenis dokumen berhasil dibuat", v)
}

func (h *AgreementTypeHandler) Update(c *gin.Context) {
	var in service.AgreementTypeInput
	if e := c.ShouldBindJSON(&in); e != nil {
		utils.BadRequest(c, "Data tidak valid", e.Error())
		return
	}
	v, e := h.svc.Update(c.Param("code"), in)
	if errors.Is(e, service.ErrJenisTidakAda) {
		utils.NotFound(c, e.Error())
		return
	}
	if e != nil {
		utils.BadRequest(c, e.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "agreement_type", v.ID)
	c.Set("audit_description", "Jenis dokumen "+v.Code+" diubah")
	utils.OK(c, "Jenis dokumen berhasil diubah", v)
}
