package dto

import "encoding/json"

type CreateAgreementDocumentRequest struct {
	DocumentTypeCode string                 `json:"document_type_code" binding:"required"`
	TemplateID       string                 `json:"template_id"` // versi template form, dari skema jenis dokumen
	FormData         map[string]interface{} `json:"form_data" binding:"required"`
}

// PreviewDraftRequest: isian wizard yang belum dikirim. DocumentID diisi saat
// merevisi; versi template harus versi terkunci pengajuan itu. AttachmentNames
// hanya nama file untuk Daftar Lampiran; file baru diunggah saat Kirim.
type PreviewDraftRequest struct {
	DocumentTypeCode string                 `json:"document_type_code"`
	TemplateID       string                 `json:"template_id"`
	DocumentID       string                 `json:"document_id"`
	FormData         map[string]interface{} `json:"form_data" binding:"required"`
	AttachmentNames  []string               `json:"attachment_names"`
}

type UpdateAgreementDocumentRequest struct {
	FormData map[string]interface{} `json:"form_data" binding:"required"`
}

type AgreementListQuery struct {
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=10"`
	Status   string `form:"status"`
	DateFrom string `form:"date_from"`
	Search   string `form:"search"`
}

type AgreementMetaRequest struct {
	AgreementNumber           *string `json:"agreement_number"`
	SigningPlace              *string `json:"signing_place"`
	SigningDate               *string `json:"signing_date"`
	PartyOneSignatoryName     *string `json:"party_one_signatory_name"`
	PartyOneSignatoryPosition *string `json:"party_one_signatory_position"`
	Pic                       *string `json:"pic"`
	Phone                     *string `json:"phone"`
	Email                     *string `json:"email"`
}

type AgreementStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Note   string `json:"note"`
}

type AgreementCompanyMasterRequest struct {
	Name                     string `json:"name" binding:"required"`
	Address                  string `json:"address" binding:"required"`
	NPWP                     string `json:"npwp"`
	Phone                    string `json:"phone"`
	Email                    string `json:"email"`
	PIC                      string `json:"pic"`
	DefaultSignatoryName     string `json:"default_signatory_name" binding:"required"`
	DefaultSignatoryPosition string `json:"default_signatory_position" binding:"required"`
	DefaultSigningPlace      string `json:"default_signing_place" binding:"required"`
	DefaultAgreementNumber   string `json:"default_agreement_number"`
}

// AgreementSignatoryRequest: satu baris Daftar Penandatangan. MaxValue nil berarti
// tak terbatas. IsActive hanya dipakai saat mengubah; penandatangan baru selalu aktif.
type AgreementSignatoryRequest struct {
	Name      string `json:"name"`
	Position  string `json:"position"`
	MaxValue  *int64 `json:"max_value"`
	IsDefault bool   `json:"is_default"`
	IsActive  *bool  `json:"is_active"`
}

type AgreementTypeResponse struct {
	Code   string          `json:"code"`
	Name   string          `json:"name"`
	Schema json.RawMessage `json:"schema,omitempty"`
}
