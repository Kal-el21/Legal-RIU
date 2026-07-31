package dto

import "time"

type RepositoryDocumentResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	FeatureCode string    `json:"feature_code"`
	SourceType  string    `json:"source_type"`
	SourceID    string    `json:"source_id"`
	UploaderType string   `json:"uploader_type"`
	FileName    string    `json:"file_name"`
	MIMEType    string    `json:"mime_type"`
	FileSize    int64     `json:"file_size"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   string    `json:"created_by"`
	UpdatedBy   string    `json:"updated_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}