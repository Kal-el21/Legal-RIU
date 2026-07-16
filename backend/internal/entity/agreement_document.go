package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AgreementCompanyMaster struct {
	Base
	Name                     string `gorm:"size:255;not null" json:"name"`
	Address                  string `gorm:"type:text;not null" json:"address"`
	NPWP                     string `gorm:"size:80" json:"npwp"`
	Phone                    string `gorm:"size:80" json:"phone"`
	Email                    string `gorm:"size:150" json:"email"`
	PIC                      string `gorm:"size:255" json:"pic"`
	DefaultSignatoryName     string `gorm:"size:255;not null" json:"default_signatory_name"`
	DefaultSignatoryPosition string `gorm:"size:255;not null" json:"default_signatory_position"`
	DefaultSigningPlace      string `gorm:"size:255;not null" json:"default_signing_place"`
	IsActive                 bool   `gorm:"not null;default:true;index" json:"is_active"`
}

type AgreementDocument struct {
	Base
	TicketNumber      string                `gorm:"size:80;uniqueIndex;not null" json:"ticket_number"`
	UserID       uuid.UUID             `gorm:"type:uuid;not null;index" json:"user_id"`
	User         User                  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	DocumentTypeCode  string                `gorm:"size:50;not null;index" json:"document_type_code"`
	FormData          json.RawMessage       `gorm:"type:jsonb;not null" json:"form_data"`
	PartyOneSnapshot  json.RawMessage       `gorm:"type:jsonb" json:"party_one_snapshot,omitempty"`
	AgreementNumber   string                `gorm:"size:150;uniqueIndex" json:"agreement_number"`
	Status            SubmissionStatus      `gorm:"not null;default:'SUBMITTED';index" json:"status"`
	StatusUpdatedAt   *time.Time            `json:"status_updated_at,omitempty"`
	ApproverNote      string                `gorm:"type:text" json:"approver_note"`
	GeneratedDOCXPath string                `gorm:"size:500" json:"-"`
	GeneratedPDFPath  string                `gorm:"size:500" json:"-"`
	GeneratedFileName string                `gorm:"size:255" json:"generated_file_name,omitempty"`
	TemplateChecksum  string                `gorm:"size:64" json:"template_checksum,omitempty"`
	ApprovedBy        *uuid.UUID            `gorm:"type:uuid" json:"approved_by,omitempty"`
	Approver          *User                 `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
	ApprovedAt        *time.Time            `json:"approved_at,omitempty"`
	Attachments       []AgreementAttachment `gorm:"foreignKey:AgreementDocumentID" json:"attachments,omitempty"`
}

type AgreementAttachment struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AgreementDocumentID uuid.UUID `gorm:"type:uuid;not null;index" json:"agreement_document_id"`
	FileName            string    `gorm:"size:255;not null" json:"file_name"`
	FilePath            string    `gorm:"size:500;not null" json:"-"`
	MIMEType            string    `gorm:"size:150" json:"mime_type"`
	FileSize            int64     `json:"file_size"`
	Description         string    `gorm:"size:500" json:"description"`
	UploadRound         int       `gorm:"not null;default:1" json:"upload_round"`
	UploadedBy          uuid.UUID `gorm:"type:uuid;not null" json:"uploaded_by"`
	CreatedAt           time.Time `json:"created_at"`
}

func (a *AgreementAttachment) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
