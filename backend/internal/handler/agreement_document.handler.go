package handler

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	tmplerrors "legal-riu-portal/internal/errors"
	"legal-riu-portal/internal/middleware"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type AgreementDocumentHandler struct {
	svc         service.AgreementDocumentService
	auditLogSvc service.AuditLogService
}

func NewAgreementDocumentHandler(svc service.AgreementDocumentService, auditLogSvc service.AuditLogService) *AgreementDocumentHandler {
	return &AgreementDocumentHandler{svc: svc, auditLogSvc: auditLogSvc}
}

// GET /api/v1/agreement-documents
func (h *AgreementDocumentHandler) GetAll(c *gin.Context) {
	var query dto.AgreementListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.BadRequest(c, "Query tidak valid", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	role := middleware.RoleWithAllAccess(c, "agreement_document.view.all")
	items, total, err := h.svc.GetAll(userID, role, query)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", gin.H{
		"items":       items,
		"total":       total,
		"page":        query.Page,
		"limit":       query.Limit,
		"total_pages": (total + int64(query.Limit) - 1) / int64(query.Limit),
	})
}

// GET /api/v1/agreement-documents/:id
func (h *AgreementDocumentHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	role := middleware.RoleWithAllAccess(c, "agreement_document.view.all")
	doc, err := h.svc.GetByID(id, userID, role)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.OK(c, "Success", doc)
}

// POST /api/v1/agreement-documents
func (h *AgreementDocumentHandler) Create(c *gin.Context) {
	var files []*multipart.FileHeader
	var req dto.CreateAgreementRequest

	if err := c.Request.ParseMultipartForm(110 << 20); err == nil {
		req = dto.CreateAgreementRequest{
			NomorPihakKedua:       c.PostForm("nomor_pihak_kedua"),
			TempatTtd:             c.PostForm("tempat_ttd"),
			TanggalTtd:            c.PostForm("tanggal_ttd"),
			PihakKeduaNama:        c.PostForm("pihak_kedua_nama"),
			PihakKeduaBidang:      c.PostForm("pihak_kedua_bidang"),
			PihakKeduaAlamat:      c.PostForm("pihak_kedua_alamat"),
			PihakKeduaTelepon:     c.PostForm("pihak_kedua_telepon"),
			PihakKeduaEmail:       c.PostForm("pihak_kedua_email"),
			PihakKeduaPic:         c.PostForm("pihak_kedua_pic"),
			PihakKeduaPejabat:     c.PostForm("pihak_kedua_pejabat"),
			PihakKeduaJabatan:     c.PostForm("pihak_kedua_jabatan"),
			JenisPekerjaan:        c.PostForm("jenis_pekerjaan"),
			SuratPenawaranNomor:   c.PostForm("surat_penawaran_nomor"),
			SuratPenawaranPerihal: c.PostForm("surat_penawaran_perihal"),
			SuratPenawaranTanggal: c.PostForm("surat_penawaran_tanggal"),
			SuratPenunjukanNomor:  c.PostForm("surat_penunjukan_nomor"),
			SuratPenunjukanPerihal: c.PostForm("surat_penunjukan_perihal"),
			SuratPenunjukanTanggal: c.PostForm("surat_penunjukan_tanggal"),
			RuangLingkup:          c.PostForm("ruang_lingkup"),
			JangkaWaktuMulai:      c.PostForm("jangka_waktu_mulai"),
			JangkaWaktuSelesai:    c.PostForm("jangka_waktu_selesai"),
			Bank:                  c.PostForm("bank"),
			NomorRekening:         c.PostForm("nomor_rekening"),
			AtasNama:              c.PostForm("atas_nama"),
		}
		if v := c.PostForm("nilai_kontrak"); v != "" {
			if f, perr := strconv.ParseFloat(v, 64); perr == nil {
				req.NilaiKontrak = f
			}
		}
		if v := c.PostForm("termin1_persen"); v != "" {
			if f, perr := strconv.ParseFloat(v, 64); perr == nil {
				req.Termin1Persen = f
			}
		}
		if v := c.PostForm("termin1_nilai"); v != "" {
			if f, perr := strconv.ParseFloat(v, 64); perr == nil {
				req.Termin1Nilai = f
			}
		}
		if v := c.PostForm("termin2_persen"); v != "" {
			if f, perr := strconv.ParseFloat(v, 64); perr == nil {
				req.Termin2Persen = f
			}
		}
		if v := c.PostForm("termin2_nilai"); v != "" {
			if f, perr := strconv.ParseFloat(v, 64); perr == nil {
				req.Termin2Nilai = f
			}
		}
		if form := c.Request.MultipartForm; form != nil {
			files = form.File["attachments"]
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.BadRequest(c, "Validasi gagal", err.Error())
			return
		}
	}

	if req.PihakKeduaNama == "" || req.PihakKeduaPejabat == "" || req.PihakKeduaJabatan == "" ||
		req.JenisPekerjaan == "" || req.RuangLingkup == "" {
		utils.BadRequest(c, "Field wajib diisi: pihak_kedua_nama, pihak_kedua_pejabat, pihak_kedua_jabatan, jenis_pekerjaan, ruang_lingkup", nil)
		return
	}

	userID := middleware.GetUserID(c)
	doc, err := h.svc.Create(userID, req, files)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "agreement_document", doc.ID.String())
	desc := "Agreement document created"
	if len(files) > 0 {
		desc = "Agreement document created with " + strconv.Itoa(len(files)) + " file(s)"
	}
	c.Set("audit_description", desc)
	utils.Created(c, "Pengajuan berhasil dibuat", doc)
}

// PUT /api/v1/agreement-documents/:id
func (h *AgreementDocumentHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateAgreementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	doc, err := h.svc.Update(id, userID, req)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "agreement_document", id)
	utils.OK(c, "Dokumen berhasil diupdate", doc)
}

// DELETE /api/v1/agreement-documents/:id
func (h *AgreementDocumentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	if err := h.svc.Delete(id, userID); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionDelete, "agreement_document", id)
	utils.OK(c, "Dokumen berhasil dihapus", nil)
}

// POST /api/v1/agreement-documents/:id/resubmit
func (h *AgreementDocumentHandler) Resubmit(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	var files []*multipart.FileHeader
	if form, _ := c.MultipartForm(); form != nil {
		files = form.File["attachments"]
	}
	doc, err := h.svc.Resubmit(id, userID, files)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "agreement_document", id)
	desc := "Agreement document resubmitted"
	if len(files) > 0 {
		desc = "Agreement document resubmitted with " + strconv.Itoa(len(files)) + " file(s)"
	}
	c.Set("audit_description", desc)
	utils.OK(c, "Pengajuan berhasil diajukan ulang", doc)
}

// GET /api/v1/agreement-documents/:id/preview  (owner or approver)
func (h *AgreementDocumentHandler) PreviewDocument(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	role := middleware.RoleWithAllAccess(c, "agreement_document.preview.all")
	pdfData, err := h.svc.GeneratePreview(c.Request.Context(), id, userID, role)
	if err != nil {
		if errors.Is(err, tmplerrors.ErrTemplateNotFound) {
			utils.BadRequest(c, "Template perjanjian belum diupload. Silakan hubungi admin.", map[string]string{
				"hint":     "Silakan upload template .docx di halaman Admin > Company Masters",
				"docs_url": "/admin/company-masters",
			})
			return
		}
		if errors.Is(err, tmplerrors.ErrConversionFailed) {
			utils.InternalError(c, "Gagal mengkonversi template PDF")
			return
		}
		if errors.Is(err, tmplerrors.ErrBasePDFNotFound) {
			utils.InternalError(c, "Base PDF template tidak ditemukan")
			return
		}
		if errors.Is(err, tmplerrors.ErrPdftoppmFailed) {
			utils.InternalError(c, "Gagal memproses PDF template")
			return
		}
		utils.InternalError(c, err.Error())
		return
	}
	c.DataFromReader(-1, -1, "application/pdf", bytes.NewReader(pdfData), map[string]string{
		"Content-Disposition": fmt.Sprintf(`inline; filename="preview-%s.pdf"`, id),
	})
}

// GET /api/v1/agreement-documents/:id/pdf  (final, owner/approver)
func (h *AgreementDocumentHandler) DownloadFinal(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	role := middleware.RoleWithAllAccess(c, "agreement_document.view.all")
	doc, err := h.svc.GetByID(id, userID, role)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	if doc.Status != entity.StatusCompleted {
		utils.BadRequest(c, "Dokumen final belum tersedia", nil)
		return
	}
	url, err := h.svc.GetDownloadURL(id)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	c.Redirect(302, url)
}

// PUT /api/v1/admin/agreement-documents/:id/meta
func (h *AgreementDocumentHandler) UpdateMeta(c *gin.Context) {
	id := c.Param("id")
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	doc, err := h.svc.UpdateMeta(id, req)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "agreement_document", id)
	utils.OK(c, "Data berhasil diupdate", doc)
}

// PUT /api/v1/admin/agreement-documents/:id/pihak-pertama
func (h *AgreementDocumentHandler) UpdatePihakPertama(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdatePihakPertamaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	doc, err := h.svc.UpdatePihakPertama(id, req)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionUserUpdate, "agreement_document", id)
	utils.OK(c, "Data Pihak Pertama berhasil diupdate", doc)
}

// POST /api/v1/admin/agreement-documents/:id/approve
func (h *AgreementDocumentHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	doc, err := h.svc.Approve(id)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionStatusChange, "agreement_document", id)
	utils.OK(c, "Dokumen berhasil disetujui", doc)
}

// POST /api/v1/admin/agreement-documents/:id/return
func (h *AgreementDocumentHandler) ReturnForRevision(c *gin.Context) {
	id := c.Param("id")
	var req dto.AgreementDecisionRequest
	_ = c.ShouldBindJSON(&req)
	doc, err := h.svc.ReturnForRevision(id, req.AdminNote)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionStatusChange, "agreement_document", id)
	utils.OK(c, "Dokumen dikembalikan untuk revisi", doc)
}

// POST /api/v1/admin/agreement-documents/:id/reject
func (h *AgreementDocumentHandler) Reject(c *gin.Context) {
	id := c.Param("id")
	var req dto.AgreementDecisionRequest
	_ = c.ShouldBindJSON(&req)
	doc, err := h.svc.Reject(id, req.AdminNote)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	middleware.SetAuditContext(c, entity.ActionStatusChange, "agreement_document", id)
	utils.OK(c, "Dokumen ditolak", doc)
}
