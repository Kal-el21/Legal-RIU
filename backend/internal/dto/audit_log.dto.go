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
