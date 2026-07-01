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
	Category          string  `json:"category" binding:"required,oneof=Life BPPDAN Property COB"`
	Specification     string  `json:"specification"`
	CaseType          string  `json:"case_type" binding:"required,oneof=NON_LITIGASI PERDATA PIDANA TIPEKOR ARBITRASE TUN"`
	TechnicalReserve  string  `json:"technical_reserve"`
	CaseValue         float64 `json:"case_value"`
	PIC               string  `json:"pic" binding:"required"`
	DocumentLink      string  `json:"document_link"`
	CurrentStatus     string  `json:"current_status"`
	CaseDate          string  `json:"case_date" binding:"required"`
	Level             string  `json:"level" binding:"required"`
	AdditionalNotes   string  `json:"additional_notes"`
	LocationRegencyID string  `json:"location_regency_id" binding:"required"`
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
	Category          string                   `json:"category"`
	Specification     string                   `json:"specification"`
	CaseType          string                   `json:"case_type"`
	TechnicalReserve  string                   `json:"technical_reserve"`
	CaseValue         float64                  `json:"case_value"`
	PIC               string                   `json:"pic"`
	PICDivision       *DivisionResponse        `json:"pic_division,omitempty"`
	DocumentLink      string                   `json:"document_link"`
	CurrentStatus     string                   `json:"current_status"`
	CaseDate          time.Time                `json:"case_date"`
	Level             string                   `json:"level"`
	AdditionalNotes   string                   `json:"additional_notes"`
	LocationRegencyID string                   `json:"location_regency_id"`
	LocationRegency   *RegencyResponse         `json:"location_regency,omitempty"`
	Chronologies      []CaseChronologyResponse `json:"chronologies,omitempty"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
}
