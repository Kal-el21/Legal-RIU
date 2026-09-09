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
	DefaultAgreementNumber   string `gorm:"size:150" json:"default_agreement_number"`
	IsActive                 bool   `gorm:"not null;default:true;index" json:"is_active"`
}

type TemplateStatus string

const (
	TemplateStatusDraft    TemplateStatus = "DRAFT"
	TemplateStatusActive   TemplateStatus = "ACTIVE"
	TemplateStatusArchived TemplateStatus = "ARCHIVED"
)

// AgreementTemplate menyimpan setiap versi template DOCX yang pernah diunggah.
// Versi tidak pernah ditimpa agar dokumen lama tetap dapat ditelusuri ke
// template persis yang menghasilkannya.
type AgreementTemplate struct {
	Base
	Code          string          `gorm:"size:50;not null;uniqueIndex:idx_agreement_template_code_version" json:"code"`
	Version       int             `gorm:"not null;uniqueIndex:idx_agreement_template_code_version" json:"version"`
	Name          string          `gorm:"size:255;not null" json:"name"`
	FileName      string          `gorm:"size:255;not null" json:"file_name"`
	FilePath      string          `gorm:"size:500;not null" json:"-"`
	PreviewPath   string          `gorm:"size:500" json:"-"`
	Checksum      string          `gorm:"size:64;not null;index" json:"checksum"`
	Status        TemplateStatus  `gorm:"size:20;not null;default:'DRAFT';index" json:"status"`
	IsLegacy      bool            `gorm:"not null;default:false" json:"is_legacy"`
	Placeholders  json.RawMessage `gorm:"type:jsonb" json:"placeholders,omitempty"`
	Note          string          `gorm:"type:text" json:"note"`
	EffectiveDate *time.Time      `json:"effective_date,omitempty"`
	UploadedBy    *uuid.UUID      `gorm:"type:uuid" json:"uploaded_by,omitempty"`
	Uploader      *User           `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
	ActivatedBy   *uuid.UUID      `gorm:"type:uuid" json:"activated_by,omitempty"`
	ActivatedAt   *time.Time      `json:"activated_at,omitempty"`
}

type AgreementDocument struct {
	Base
	TicketNumber      string                `gorm:"size:80;uniqueIndex;not null" json:"ticket_number"`
	UserID       uuid.UUID             `gorm:"type:uuid;not null;index" json:"user_id"`
	User         User                  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	DocumentTypeCode  string                `gorm:"size:50;not null;index" json:"document_type_code"`
	FormData          json.RawMessage       `gorm:"type:jsonb;not null" json:"form_data"`
	PartyOneSnapshot  json.RawMessage       `gorm:"type:jsonb" json:"party_one_snapshot,omitempty"`
	AgreementNumber   *string               `gorm:"size:150;uniqueIndex" json:"agreement_number"`
	Status            SubmissionStatus      `gorm:"not null;default:'SUBMITTED';index" json:"status"`
	StatusUpdatedAt   *time.Time            `json:"status_updated_at,omitempty"`
	ApproverNote      string                `gorm:"type:text" json:"approver_note"`
	GeneratedDOCXPath string                `gorm:"size:500" json:"-"`
	GeneratedPDFPath  string                `gorm:"size:500" json:"-"`
	GeneratedFileName string                `gorm:"size:255" json:"generated_file_name,omitempty"`
	TemplateID        *uuid.UUID            `gorm:"type:uuid;index" json:"template_id,omitempty"`
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
