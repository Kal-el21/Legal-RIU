package dto

import (
	"time"

	"github.com/google/uuid"
)

type AuditLogFilters struct {
	Page       int
	Limit      int
	Action     *string
	EntityType *string
	EntityID   *uuid.UUID
	UserID     *uuid.UUID
	DateFrom   *time.Time
	DateTo     *time.Time
	Search     *string
}

// PelakuAudit: pelaku yang ditampilkan di Audit Log, tanpa relasi dan data akun lain
// (PR Audit Log tahap L.1).
type PelakuAudit struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type AuditLogResponse struct {
	ID          string       `json:"id"`
	UserID      string       `json:"user_id"`
	User        *PelakuAudit `json:"user,omitempty"`
	Action      string       `json:"action"`
	EntityType  string       `json:"entity_type"`
	EntityID    string       `json:"entity_id"`
	OldValue    *string      `json:"old_value,omitempty"`
	NewValue    *string      `json:"new_value,omitempty"`
	Description *string      `json:"description,omitempty"`
	IPAddress   string       `json:"ip_address"`
	UserAgent   string       `json:"user_agent"`
	CreatedAt   time.Time    `json:"created_at"`
}
