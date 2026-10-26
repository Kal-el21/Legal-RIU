package dto

import (
	"errors"
	"net/mail"
	"strings"
)

// CreateLegalOpinionRequest memuat isian untuk pengajuan langsung dan draf.
type CreateLegalOpinionRequest struct {
	RequestorName     string `json:"requestor_name"`
	RequestorPosition string `json:"requestor_position"`
	RequestorDivision string `json:"requestor_division"`
	RequestorEmail    string `json:"requestor_email"`
	RequestorPhone    string `json:"requestor_phone"`
	LegalType         string `json:"legal_type"`
	LegalTypeOther    string `json:"legal_type_other"`
	Title             string `json:"title"`
	Chronology        string `json:"chronology"`
	Question          string `json:"question"`
}

func (r *CreateLegalOpinionRequest) Normalize() {
	r.RequestorName = strings.TrimSpace(r.RequestorName)
	r.RequestorPosition = strings.TrimSpace(r.RequestorPosition)
	r.RequestorDivision = strings.TrimSpace(r.RequestorDivision)
	r.RequestorEmail = strings.TrimSpace(r.RequestorEmail)
	r.RequestorPhone = strings.TrimSpace(r.RequestorPhone)
	r.LegalType = strings.TrimSpace(r.LegalType)
	r.LegalTypeOther = strings.TrimSpace(r.LegalTypeOther)
	r.Title = strings.TrimSpace(r.Title)
	r.Chronology = strings.TrimSpace(r.Chronology)
	r.Question = strings.TrimSpace(r.Question)
}

func (r CreateLegalOpinionRequest) Validate() error {
	for _, field := range []struct{ value, label string }{
		{r.RequestorName, "Nama Lengkap"},
		{r.RequestorPosition, "Posisi Jabatan"},
		{r.RequestorDivision, "Divisi pada RIU"},
		{r.RequestorEmail, "Email Kantor"},
		{r.RequestorPhone, "Nomor WhatsApp"},
		{r.LegalType, "Jenis Kajian Hukum"},
		{r.Title, "Judul Permasalahan Hukum"},
		{r.Chronology, "Kronologis Permasalahan"},
		{r.Question, "Pertanyaan Permasalahan Hukum"},
	} {
		if strings.TrimSpace(field.value) == "" {
			return errors.New(field.label + " wajib diisi")
		}
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(r.RequestorEmail)); err != nil {
		return errors.New("Email pemohon tidak valid")
	}
	if r.LegalType == "Lain-Lain" && strings.TrimSpace(r.LegalTypeOther) == "" {
		return errors.New("Jenis kajian lainnya wajib diisi")
	}
	return nil
}

type UpdateLegalOpinionRequest struct {
	RequestorName     string `json:"requestor_name"`
	RequestorPosition string `json:"requestor_position"`
	RequestorDivision string `json:"requestor_division"`
	RequestorEmail    string `json:"requestor_email"`
	RequestorPhone    string `json:"requestor_phone"`
	LegalType         string `json:"legal_type"`
	LegalTypeOther    string `json:"legal_type_other"`
	Title             string `json:"title"`
	Chronology        string `json:"chronology"`
	Question          string `json:"question"`
}

func (r *UpdateLegalOpinionRequest) Normalize() {
	req := CreateLegalOpinionRequest(*r)
	req.Normalize()
	*r = UpdateLegalOpinionRequest(req)
}

func (r UpdateLegalOpinionRequest) Validate() error {
	return CreateLegalOpinionRequest(r).Validate()
}

type UpdateStatusRequest struct {
	Status    string `json:"status" binding:"required"`
	AdminNote string `json:"admin_note"`
}

type UploadResultRequest struct {
	Notes string `json:"notes"`
}

type LegalOpinionListQuery struct {
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
	Status string `form:"status"`
}
