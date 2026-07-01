package dto

type UserDashboardStats struct {
	TotalLegalOpinions   int64 `json:"total_legal_opinions"`
	TotalDocumentReviews int64 `json:"total_document_reviews"`
	Pending              int64 `json:"pending"`
	NeedRevision         int64 `json:"need_revision"`
	Completed            int64 `json:"completed"`
}

type AdminDashboardStats struct {
	TotalUsers           int64 `json:"total_users"`
	TotalLegalOpinions   int64 `json:"total_legal_opinions"`
	TotalDocumentReviews int64 `json:"total_document_reviews"`
	PendingReview        int64 `json:"pending_review"`
	NeedRevision         int64 `json:"need_revision"`
	Resubmitted          int64 `json:"resubmitted"`
}

type ReminderItem struct {
	ID                  string  `json:"id"`
	SubmissionType      string  `json:"submission_type"`
	TicketNumber        string  `json:"ticket_number"`
	Title               string  `json:"title"`
	Status              string  `json:"status"`
	SubmittedAt         string  `json:"submitted_at"`
	LastUpdatedAt       *string `json:"last_updated_at"`
	DaysSinceSubmission int     `json:"days_since_submission"`
	DaysSinceLastUpdate int     `json:"days_since_last_update"`
	WarningLevel        string  `json:"warning_level"`
	WarningColor        string  `json:"warning_color"`
	IsRead              bool    `json:"is_read"`
	AssignedLegalName   string  `json:"assigned_legal_name,omitempty"`
}

type RemindersResponse struct {
	Yellow      []ReminderItem `json:"yellow"`
	Red         []ReminderItem `json:"red"`
	None        []ReminderItem `json:"none"`
	Items       []ReminderItem `json:"items"`
	Total       int            `json:"total"`
	UnreadTotal int            `json:"unread_total"`
	Page        int            `json:"page"`
	Limit       int            `json:"limit"`
	TotalPages  int            `json:"total_pages"`
}

type MarkReminderReadRequest struct {
	SubmissionType string `json:"submission_type" binding:"required"`
	SubmissionID   string `json:"submission_id" binding:"required"`
}
