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

// AgreementSignatory adalah penandatangan Pihak Pertama untuk satu rentang nilai
// kontrak (PR 11). MaxValue nil berarti tak terbatas. Batas bawah tidak disimpan:
// diturunkan dari batas baris sebelumnya, jadi rentang tidak bisa tumpang tindih atau
// berlubang. IsDefault dipakai jenis dokumen tanpa nilai utama.
type AgreementSignatory struct {
	Base
	Name      string `gorm:"size:255;not null" json:"name"`
	Position  string `gorm:"size:255;not null" json:"position"`
	MaxValue  *int64 `json:"max_value"`
	IsDefault bool   `gorm:"not null;default:false" json:"is_default"`
	IsActive  bool   `gorm:"not null;default:true" json:"is_active"`
}

// AgreementType adalah jenis dokumen perjanjian (mis. PKS). Fields di sini adalah
// field kerja yang sedang disusun Admin; pengajuan tidak membacanya. Saat versi
// template dibuat, field ini disalin ke versi itu (AgreementTemplate.Fields), dan
// pengajuan memakai salinan beku di versinya.
type AgreementType struct {
	Base
	Code     string          `gorm:"size:50;not null;uniqueIndex" json:"code"`
	Name     string          `gorm:"size:255;not null" json:"name"`
	IsActive bool            `gorm:"not null;default:true" json:"is_active"`
	Fields   json.RawMessage `gorm:"type:jsonb" json:"fields"`
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
	Fields        json.RawMessage `gorm:"type:jsonb" json:"fields,omitempty"` // definisi field per bagian, dipakai pengajuan dengan versi ini
	Note          string          `gorm:"type:text" json:"note"`
	EffectiveDate *time.Time      `json:"effective_date,omitempty"`
	UploadedBy    *uuid.UUID      `gorm:"type:uuid" json:"uploaded_by,omitempty"`
	Uploader      *User           `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
	ActivatedBy   *uuid.UUID      `gorm:"type:uuid" json:"activated_by,omitempty"`
	Activator     *User           `gorm:"foreignKey:ActivatedBy" json:"activator,omitempty"`
	ActivatedAt   *time.Time      `json:"activated_at,omitempty"`
	// RunningDocuments: jumlah pengajuan belum selesai yang dikunci ke versi ini;
	// diisi saat daftar versi dibaca, tidak disimpan.
	RunningDocuments int64 `gorm:"-" json:"running_documents"`
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
	// OldTemplate diisi untuk approver kalau pengajuan berjalan masih memakai versi
	// template lama; tidak disimpan.
	OldTemplate *TemplateLama `gorm:"-" json:"old_template,omitempty"`
	// SignatorySuggestion: penandatangan Pihak Pertama usulan untuk approver, menurut
	// nilai utama dan Daftar Penandatangan terbaru; dihitung saat dibaca (PR 11, P2).
	SignatorySuggestion *UsulanPenandatangan `gorm:"-" json:"signatory_suggestion,omitempty"`
}

// TemplateLama: versi template pengajuan dibanding versi aktif jenisnya (PR 9 tahap 9.7).
// UsulanPenandatangan: Value adalah nilai utama pengajuan (kosong untuk jenis tanpa
// nilai utama). ID kosong kalau tidak ada penandatangan yang cocok.
type UsulanPenandatangan struct {
	ID       *uuid.UUID `json:"id"`
	Name     string     `json:"name"`
	Position string     `json:"position"`
	Value    *int64     `json:"value"`
}

type TemplateLama struct {
	Version       int        `json:"version"`
	ActiveVersion int        `json:"active_version"`
	ActiveNote    string     `json:"active_note"`
	ActivatedAt   *time.Time `json:"activated_at,omitempty"`
	ActivatedBy   string     `json:"activated_by"`
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
