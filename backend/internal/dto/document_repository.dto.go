package dto

import "time"

type RepositoryDocumentFilters struct {
	Page        int
	Limit       int
	FeatureCode string
	SourceType  string
	MIMEType    string
	DateFrom    *time.Time
	DateTo      *time.Time
	Search      string
}

type RepositoryDocumentResponse struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	FeatureCode  string     `json:"feature_code"`
	SourceType   string     `json:"source_type"`
	SourceID     string     `json:"source_id"`
	SourceNumber *string    `json:"source_number,omitempty"`
	SourceTitle  *string    `json:"source_title,omitempty"`
	UploadedAt   *time.Time `json:"uploaded_at,omitempty"`
	UploaderType string     `json:"uploader_type"`
	FileName     string     `json:"file_name"`
	MIMEType     string     `json:"mime_type"`
	FileSize     int64      `json:"file_size"`
	IsActive     bool       `json:"is_active"`
	CreatedBy    string     `json:"created_by"`
	UpdatedBy    string     `json:"updated_by"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type RepositoryDocumentListResponse struct {
	Items      []RepositoryDocumentResponse `json:"items"`
	Total      int64                        `json:"total"`
	Page       int                          `json:"page"`
	Limit      int                          `json:"limit"`
	TotalPages int                          `json:"total_pages"`
}
