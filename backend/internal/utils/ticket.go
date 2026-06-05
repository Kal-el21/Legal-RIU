package utils

import (
	"fmt"
	"time"
)

type TicketPrefix string

const (
	PrefixLegalOpinion   TicketPrefix = "LO"
	PrefixDocumentReview TicketPrefix = "RD"
)

// GenerateTicketNumber creates a ticket number in format: LO-202506-0001
// The sequence number comes from counting existing records for that month
func GenerateTicketNumber(prefix TicketPrefix, sequence int) string {
	now := time.Now()
	return fmt.Sprintf("%s-%s-%04d", prefix, now.Format("200601"), sequence)
}
