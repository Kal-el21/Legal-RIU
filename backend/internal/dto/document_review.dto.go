package dto

import (
	"errors"
	"net/mail"
	"strings"
)

type CreateDocumentReviewRequest struct {
	RequestorName     string `json:"requestor_name"`
	RequestorPosition string `json:"requestor_position"`
	RequestorDivision string `json:"requestor_division"`
	RequestorEmail    string `json:"requestor_email"`
	RequestorPhone    string `json:"requestor_phone"`
	DocumentName      string `json:"document_name"`
	SecondParty       string `json:"second_party"`
	ThirdParty        string `json:"third_party"`
	DocumentType      string `json:"document_type"`
	DocumentTypeOther string `json:"document_type_other"`
	AdditionalNote    string `json:"additional_note"`
}

func (r *CreateDocumentReviewRequest) Normalize() {
	r.RequestorName = strings.TrimSpace(r.RequestorName)
	r.RequestorPosition = strings.TrimSpace(r.RequestorPosition)
	r.RequestorDivision = strings.TrimSpace(r.RequestorDivision)
	r.RequestorEmail = strings.TrimSpace(r.RequestorEmail)
	r.RequestorPhone = strings.TrimSpace(r.RequestorPhone)
	r.DocumentName = strings.TrimSpace(r.DocumentName)
	r.SecondParty = strings.TrimSpace(r.SecondParty)
	r.ThirdParty = strings.TrimSpace(r.ThirdParty)
	r.DocumentType = strings.TrimSpace(r.DocumentType)
	r.DocumentTypeOther = strings.TrimSpace(r.DocumentTypeOther)
	r.AdditionalNote = strings.TrimSpace(r.AdditionalNote)
}

func (r CreateDocumentReviewRequest) Validate() error {
	for _, field := range []struct{ value, label string }{
		{r.RequestorName, "Nama Lengkap"},
		{r.RequestorPosition, "Posisi Jabatan"},
		{r.RequestorDivision, "Divisi pada RIU"},
		{r.RequestorEmail, "Email Kantor"},
		{r.RequestorPhone, "Nomor WhatsApp"},
		{r.DocumentName, "Nama Dokumen"},
		{r.SecondParty, "Pihak Kedua / Vendor / Broker / Ceding"},
		{r.DocumentType, "Jenis Dokumen"},
	} {
		if strings.TrimSpace(field.value) == "" {
			return errors.New(field.label + " wajib diisi")
		}
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(r.RequestorEmail)); err != nil {
		return errors.New("Email pemohon tidak valid")
	}
	if r.DocumentType == "Lain-Lain" && strings.TrimSpace(r.DocumentTypeOther) == "" {
		return errors.New("Jenis dokumen lainnya wajib diisi")
	}
	return nil
}

type UpdateDocumentReviewRequest struct {
	RequestorName     string `json:"requestor_name"`
	RequestorPosition string `json:"requestor_position"`
	RequestorDivision string `json:"requestor_division"`
	RequestorEmail    string `json:"requestor_email"`
	RequestorPhone    string `json:"requestor_phone"`
	DocumentName      string `json:"document_name"`
	SecondParty       string `json:"second_party"`
	ThirdParty        string `json:"third_party"`
	DocumentType      string `json:"document_type"`
	DocumentTypeOther string `json:"document_type_other"`
	AdditionalNote    string `json:"additional_note"`
}

func (r *UpdateDocumentReviewRequest) Normalize() {
	req := CreateDocumentReviewRequest(*r)
	req.Normalize()
	*r = UpdateDocumentReviewRequest(req)
}

func (r UpdateDocumentReviewRequest) Validate() error {
	return CreateDocumentReviewRequest(r).Validate()
}

type DocumentReviewListQuery struct {
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
	Status string `form:"status"`
}
