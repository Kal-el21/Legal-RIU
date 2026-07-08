package dto

import "time"

type LegalCaseListQuery struct {
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=10"`
	Search   string `form:"search"`
	Status   string `form:"status"`
	CaseType string `form:"case_type"`
	Level    string `form:"level"`
	DateFrom string `form:"date_from"`
	DateTo   string `form:"date_to"`
}

type CreateLegalCaseRequest struct {
	CaseName          string  `json:"case_name" binding:"required"`
	CaseSummary       string  `json:"case_summary"`
	RelatedPartyID    string  `json:"related_party_id" binding:"required"`
	CategoryID        string  `json:"category_id" binding:"required"`
	Specification     string  `json:"specification"`
	CaseTypeID        string  `json:"case_type_id" binding:"required"`
	TechnicalReserve  float64 `json:"technical_reserve"`
	CaseValue         float64 `json:"case_value"`
	PIC               string  `json:"pic" binding:"required"`
	DocumentLink      string  `json:"document_link"`
	Photo             string  `json:"photo"`
	CurrentStatus     string  `json:"current_status"`
	CaseDate          string  `json:"case_date" binding:"required"`
	Level             string  `json:"level" binding:"required"`
	AdditionalNotes   string  `json:"additional_notes"`
	LocationRegencyID string  `json:"location_regency_id" binding:"required"`
	CompanyID         string  `json:"company_id" binding:"required"`
}

type UpdateLegalCaseRequest = CreateLegalCaseRequest



type CreateCaseChronologyRequest struct {
	AgendaDate  string   `json:"agenda_date" binding:"required"`
	Agenda      string   `json:"agenda" binding:"required"`
	Description string   `json:"description"`
	Documents   []string `json:"documents"`
}

type UpdateCaseChronologyRequest = CreateCaseChronologyRequest

type CreateCedantRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateCedantRequest = CreateCedantRequest

type CompanyResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	EmailDomain string `json:"email_domain"`
	IsInternal  bool   `json:"is_internal"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PurposeTypeResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CaseTypeResponse struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Label     string `json:"label"`
	IsActive  bool   `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CaseCategoryResponse struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Label     string `json:"label"`
	IsActive  bool   `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LegalMaterialResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Excerpt   string `json:"excerpt"`
	Content   string `json:"content"`
	CreatedBy string `json:"created_by"`
	UpdatedBy string `json:"updated_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateLegalMaterialRequest struct {
	Title   string `json:"title" binding:"required"`
	Excerpt string `json:"excerpt"`
	Content string `json:"content" binding:"required"`
}

type UpdateLegalMaterialRequest struct {
	Title   string `json:"title" binding:"required"`
	Excerpt string `json:"excerpt"`
	Content string `json:"content" binding:"required"`
}

type RegencyResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Province string `json:"province"`
	Type     string `json:"type"`
	Label    string `json:"label"`
}

type DivisionResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CedantResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CaseChronologyResponse struct {
	ID          string    `json:"id"`
	CaseID      string    `json:"case_id"`
	AgendaDate  time.Time `json:"agenda_date"`
	Agenda      string    `json:"agenda"`
	Description string    `json:"description"`
	Documents   []string  `json:"documents"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LegalCaseResponse struct {
	ID                string                   `json:"id"`
	CaseName          string                   `json:"case_name"`
	CaseSummary       string                   `json:"case_summary"`
	RelatedPartyID    string                   `json:"related_party_id"`
	RelatedParty      *CedantResponse          `json:"related_party,omitempty"`
	CategoryID        string                   `json:"category_id"`
	Category          *CaseCategoryResponse    `json:"category,omitempty"`
	Specification     string                   `json:"specification"`
	CaseTypeID        string                   `json:"case_type_id"`
	CaseType          *CaseTypeResponse        `json:"case_type,omitempty"`
	TechnicalReserve  float64                  `json:"technical_reserve"`
	CaseValue         float64                  `json:"case_value"`
	PIC               string                   `json:"pic"`
	PICDivision       *DivisionResponse        `json:"pic_division,omitempty"`
	DocumentLink      string                   `json:"document_link"`
	Photo             string                   `json:"photo"`
	CurrentStatus     string                   `json:"current_status"`
	CaseDate          time.Time                `json:"case_date"`
	Level             string                   `json:"level"`
	AdditionalNotes   string                   `json:"additional_notes"`
	LocationRegencyID string                   `json:"location_regency_id"`
	LocationRegency   *RegencyResponse         `json:"location_regency,omitempty"`
	CompanyID         string                   `json:"company_id"`
	Company           *CompanyResponse         `json:"company,omitempty"`
	Chronologies      []CaseChronologyResponse `json:"chronologies,omitempty"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
}


