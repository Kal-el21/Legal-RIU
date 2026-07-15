package dto

import "encoding/json"

type CreateAgreementDocumentRequest struct {
	DocumentTypeCode string                 `json:"document_type_code" binding:"required"`
	FormData         map[string]interface{} `json:"form_data" binding:"required"`
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
}

type AgreementTypeResponse struct {
	Code   string          `json:"code"`
	Name   string          `json:"name"`
	Schema json.RawMessage `json:"schema,omitempty"`
}
