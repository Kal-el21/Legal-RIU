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
type LegalCaseType string

const (
	RoleUser     UserRole = "USER"
	RoleAdmin    UserRole = "ADMIN"
	RoleLegal    UserRole = "LEGAL"
	RoleExternal UserRole = "EXTERNAL"

	UserActive   UserStatus = "ACTIVE"
	UserInactive UserStatus = "INACTIVE"

	StatusSubmitted    SubmissionStatus = "SUBMITTED"
	StatusUnderReview  SubmissionStatus = "UNDER_REVIEW"
	StatusNeedRevision SubmissionStatus = "NEED_REVISION"
	StatusRejected     SubmissionStatus = "REJECTED"
	StatusResubmitted  SubmissionStatus = "RESUBMITTED"
	StatusCompleted    SubmissionStatus = "COMPLETED"

	CaseTypeNonLitigasi LegalCaseType = "NON_LITIGASI"
	CaseTypePerdata     LegalCaseType = "PERDATA"
	CaseTypePidana      LegalCaseType = "PIDANA"
	CaseTypeTipekor     LegalCaseType = "TIPEKOR"
	CaseTypeArbitrase   LegalCaseType = "ARBITRASE"
	CaseTypeTUN         LegalCaseType = "TUN"

	CaseCategoryLife     string = "Life"
	CaseCategoryBPPDAN   string = "BPPDAN"
	CaseCategoryProperty string = "Property"
	CaseCategoryCOB      string = "COB"
)

// ─── User ─────────────────────────────────────────────────────────────────────

type User struct {
	Base
	FullName           string     `gorm:"not null" json:"full_name"`
	Email              string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash       string     `gorm:"not null" json:"-"`
	Position           string     `gorm:"not null" json:"position"`
	Division           string     `gorm:"not null" json:"division"`
	DivisionID         *uuid.UUID `gorm:"type:uuid;index" json:"division_id,omitempty"`
	DivisionRef        Division   `gorm:"foreignKey:DivisionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"division_ref,omitempty"`
	Role               UserRole   `gorm:"not null;default:'USER'" json:"role"`
	Status             UserStatus `gorm:"not null;default:'ACTIVE'" json:"status"`
	EmailNotifications bool       `gorm:"not null;default:true" json:"email_notifications"`
	TwoFAEnabled       bool       `gorm:"not null;default:false" json:"two_fa_enabled"`
}

type RefreshToken struct {
	Base
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	User       User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TokenHash  string     `gorm:"size:64;uniqueIndex;not null" json:"-"`
	ExpiresAt  time.Time  `gorm:"not null;index" json:"expires_at"`
	RevokedAt  *time.Time `gorm:"index" json:"revoked_at"`
	IPAddress  string     `gorm:"size:45" json:"ip_address"`
	UserAgent  string     `gorm:"size:255" json:"user_agent"`
	LastUsedAt time.Time  `json:"last_used_at"`
}

func (r *RefreshToken) IsValid() bool {
	return r.RevokedAt == nil && time.Now().Before(r.ExpiresAt)
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

// ─── Audit Log ─────────────────────────────────────────────────────────────────

// Case Management

type Regency struct {
	Base
	Name     string `gorm:"size:255;not null;uniqueIndex:idx_regencies_name_province" json:"name"`
	Province string `gorm:"size:255;not null;uniqueIndex:idx_regencies_name_province;index" json:"province"`
	Type     string `gorm:"size:20;not null;index" json:"type"`
}

type Cedant struct {
	Base
	Name        string `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description string `gorm:"type:text" json:"description"`
}

type Division struct {
	Base
	Name        string `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description string `gorm:"type:text" json:"description"`
}

type LegalCase struct {
	Base
	CaseName          string           `gorm:"size:255;not null;index" json:"case_name"`
	CaseSummary       string           `gorm:"type:text" json:"case_summary"`
	RelatedPartyID    uuid.UUID        `gorm:"type:uuid;not null;index" json:"related_party_id"`
	RelatedParty      Cedant           `gorm:"foreignKey:RelatedPartyID" json:"related_party,omitempty"`
	Category          string           `gorm:"size:100;not null;index" json:"category"`
	Specification     string           `gorm:"type:text" json:"specification"`
	CaseType          LegalCaseType    `gorm:"size:20;not null;index" json:"case_type"`
	TechnicalReserve  string           `gorm:"size:255" json:"technical_reserve"`
	CaseValue         float64          `gorm:"type:decimal(18,2)" json:"case_value"`
	PIC               uuid.UUID        `gorm:"type:uuid;index" json:"pic"`
	PICDivision       Division         `gorm:"foreignKey:PIC" json:"pic_division,omitempty"`
	DocumentLink      string           `gorm:"size:500" json:"document_link"`
	CurrentStatus     string           `gorm:"size:100;index" json:"current_status"`
	CaseDate          time.Time        `gorm:"not null;index" json:"case_date"`
	Level             string           `gorm:"size:100;not null;index" json:"level"`
	AdditionalNotes   string           `gorm:"type:text" json:"additional_notes"`
	LocationRegencyID uuid.UUID        `gorm:"type:uuid;not null;index" json:"location_regency_id"`
	LocationRegency   Regency          `gorm:"foreignKey:LocationRegencyID" json:"location_regency,omitempty"`
	Chronologies      []CaseChronology `gorm:"foreignKey:CaseID" json:"chronologies,omitempty"`
}

type CaseChronology struct {
	Base
	CaseID      uuid.UUID `gorm:"type:uuid;not null;index" json:"case_id"`
	LegalCase   LegalCase `gorm:"foreignKey:CaseID" json:"legal_case,omitempty"`
	AgendaDate  time.Time `gorm:"not null;index" json:"agenda_date"`
	Agenda      string    `gorm:"type:text;not null" json:"agenda"`
	Description string    `gorm:"type:text" json:"description"`
	Documents   string    `gorm:"type:text" json:"-"`
}

// Audit Log

type AuditAction string

const (
	ActionStatusChange AuditAction = "STATUS_CHANGE"
	ActionFileUpload   AuditAction = "FILE_UPLOAD"
	ActionUserUpdate   AuditAction = "USER_UPDATE"
	ActionLogin        AuditAction = "LOGIN"
	ActionLogout       AuditAction = "LOGOUT"
	ActionDelete       AuditAction = "DELETE"
	ActionFileDelete   AuditAction = "FILE_DELETE"
)

type AuditLog struct {
	Base
	UserID      uuid.UUID   `gorm:"type:uuid;not null;index" json:"user_id"`
	User        User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action      AuditAction `gorm:"not null;index" json:"action"`
	EntityType  string      `gorm:"not null;index" json:"entity_type"`
	EntityID    uuid.UUID   `gorm:"type:uuid;not null;index" json:"entity_id"`
	OldValue    *string     `gorm:"type:text" json:"old_value,omitempty"`
	NewValue    *string     `gorm:"type:text" json:"new_value,omitempty"`
	Description *string     `gorm:"type:text" json:"description,omitempty"`
	IPAddress   string      `gorm:"size:45" json:"ip_address"`
	UserAgent   string      `gorm:"size:500" json:"user_agent"`
}

// ─── Settings ─────────────────────────────────────────────────────────────────

type UserSettings struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID            uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	EmailNotification bool      `gorm:"default:true" json:"email_notification"`
	TwoFactorEnabled  bool      `gorm:"default:false" json:"two_factor_enabled"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (s *UserSettings) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type NotificationSetting struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SubmissionType string    `gorm:"size:50;not null" json:"submission_type"`
	WarningLevel   string    `gorm:"size:20;not null" json:"warning_level"`
	DaysThreshold  int       `gorm:"not null" json:"days_threshold"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (n *NotificationSetting) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

type NotificationRead struct {
	Base
	UserID         uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_notification_reads_user_submission" json:"user_id"`
	User           User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	SubmissionType string     `gorm:"size:50;not null;uniqueIndex:idx_notification_reads_user_submission" json:"submission_type"`
	SubmissionID   uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_notification_reads_user_submission" json:"submission_id"`
	IsRead         bool       `gorm:"not null;default:true" json:"is_read"`
	ReadAt         *time.Time `json:"read_at"`
}
