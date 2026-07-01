package dto

type NotificationSettingResponse struct {
	ID             string `json:"id"`
	SubmissionType string `json:"submission_type"`
	WarningLevel   string `json:"warning_level"`
	DaysThreshold  int    `json:"days_threshold"`
	IsActive       bool   `json:"is_active"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type UpdateNotificationSettingRequest struct {
	DaysThreshold int   `json:"days_threshold" binding:"required,min=1"`
	IsActive      *bool `json:"is_active"`
}
