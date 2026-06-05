package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── Base ────────────────────────────────────────────────────────────────────

type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// ─── Role & Status constants ──────────────────────────────────────────────────

type UserRole string
type UserStatus string
type SubmissionStatus string

const (
	RoleUser  UserRole = "USER"
	RoleAdmin UserRole = "ADMIN"

	UserActive   UserStatus = "ACTIVE"
	UserInactive UserStatus = "INACTIVE"

	StatusSubmitted    SubmissionStatus = "SUBMITTED"
	StatusUnderReview  SubmissionStatus = "UNDER_REVIEW"
	StatusNeedRevision SubmissionStatus = "NEED_REVISION"
	StatusRejected     SubmissionStatus = "REJECTED"
	StatusResubmitted  SubmissionStatus = "RESUBMITTED"
	StatusCompleted    SubmissionStatus = "COMPLETED"
)

// ─── User ─────────────────────────────────────────────────────────────────────

type User struct {
	Base
	FullName     string     `gorm:"not null" json:"full_name"`
	Email        string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"not null" json:"-"`
	Position     string     `gorm:"not null" json:"position"`
	Division     string     `gorm:"not null" json:"division"`
	Role         UserRole   `gorm:"not null;default:'USER'" json:"role"`
	Status       UserStatus `gorm:"not null;default:'ACTIVE'" json:"status"`
}

// ─── Legal Opinion ────────────────────────────────────────────────────────────

type LegalOpinion struct {
	Base
	TicketNumber      string                   `gorm:"uniqueIndex;not null" json:"ticket_number"`
	UserID            uuid.UUID                `gorm:"type:uuid;not null" json:"user_id"`
	User              User                     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	RequestorName     string                   `gorm:"not null" json:"requestor_name"`
	RequestorPosition string                   `gorm:"not null" json:"requestor_position"`
	RequestorDivision string                   `gorm:"not null" json:"requestor_division"`
	RequestorEmail    string                   `gorm:"not null" json:"requestor_email"`
	RequestorPhone    string                   `gorm:"not null" json:"requestor_phone"`
	LegalType         string                   `gorm:"not null" json:"legal_type"`
	LegalTypeOther    string                   `json:"legal_type_other"`
	Title             string                   `gorm:"not null" json:"title"`
	Chronology        string                   `gorm:"type:text;not null" json:"chronology"`
	Question          string                   `gorm:"type:text;not null" json:"question"`
	Status            SubmissionStatus         `gorm:"not null;default:'SUBMITTED'" json:"status"`
	AdminNote         string                   `gorm:"type:text" json:"admin_note"`
	Attachments       []LegalOpinionAttachment `gorm:"foreignKey:LegalOpinionID" json:"attachments,omitempty"`
	Results           []LegalOpinionResult     `gorm:"foreignKey:LegalOpinionID" json:"results,omitempty"`
}

type LegalOpinionAttachment struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	LegalOpinionID uuid.UUID `gorm:"type:uuid;not null" json:"legal_opinion_id"`
	FileName       string    `gorm:"not null" json:"file_name"`
	FilePath       string    `gorm:"not null" json:"file_path"`
	FileSize       int64     `json:"file_size"`
	UploadRound    int       `gorm:"not null;default:1" json:"upload_round"`
	CreatedAt      time.Time `json:"created_at"`
}

func (a *LegalOpinionAttachment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type LegalOpinionResult struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	LegalOpinionID uuid.UUID `gorm:"type:uuid;not null" json:"legal_opinion_id"`
	UploadedBy     uuid.UUID `gorm:"type:uuid;not null" json:"uploaded_by"`
	Uploader       User      `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
	FileName       string    `json:"file_name"`
	FilePath       string    `json:"file_path"`
	Notes          string    `gorm:"type:text" json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
}

func (r *LegalOpinionResult) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// ─── Document Review ──────────────────────────────────────────────────────────

type DocumentReview struct {
	Base
	TicketNumber      string                     `gorm:"uniqueIndex;not null" json:"ticket_number"`
	UserID            uuid.UUID                  `gorm:"type:uuid;not null" json:"user_id"`
	User              User                       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	RequestorName     string                     `gorm:"not null" json:"requestor_name"`
	RequestorPosition string                     `gorm:"not null" json:"requestor_position"`
	RequestorDivision string                     `gorm:"not null" json:"requestor_division"`
	RequestorEmail    string                     `gorm:"not null" json:"requestor_email"`
	RequestorPhone    string                     `gorm:"not null" json:"requestor_phone"`
	DocumentName      string                     `gorm:"not null" json:"document_name"`
	SecondParty       string                     `gorm:"not null" json:"second_party"`
	ThirdParty        string                     `json:"third_party"`
	DocumentType      string                     `gorm:"not null" json:"document_type"`
	DocumentTypeOther string                     `json:"document_type_other"`
	AdditionalNote    string                     `gorm:"type:text" json:"additional_note"`
	Status            SubmissionStatus           `gorm:"not null;default:'SUBMITTED'" json:"status"`
	AdminNote         string                     `gorm:"type:text" json:"admin_note"`
	Attachments       []DocumentReviewAttachment `gorm:"foreignKey:DocumentReviewID" json:"attachments,omitempty"`
	Results           []DocumentReviewResult     `gorm:"foreignKey:DocumentReviewID" json:"results,omitempty"`
}

type DocumentReviewAttachment struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DocumentReviewID uuid.UUID `gorm:"type:uuid;not null" json:"document_review_id"`
	FileName         string    `gorm:"not null" json:"file_name"`
	FilePath         string    `gorm:"not null" json:"file_path"`
	FileSize         int64     `json:"file_size"`
	UploadRound      int       `gorm:"not null;default:1" json:"upload_round"`
	CreatedAt        time.Time `json:"created_at"`
}

func (a *DocumentReviewAttachment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type DocumentReviewResult struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DocumentReviewID uuid.UUID `gorm:"type:uuid;not null" json:"document_review_id"`
	UploadedBy       uuid.UUID `gorm:"type:uuid;not null" json:"uploaded_by"`
	Uploader         User      `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
	FileName         string    `json:"file_name"`
	FilePath         string    `json:"file_path"`
	Notes            string    `gorm:"type:text" json:"notes"`
	CreatedAt        time.Time `json:"created_at"`
}

func (r *DocumentReviewResult) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
