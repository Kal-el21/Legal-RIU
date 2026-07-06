package dto

import "time"

type DocumentTypeResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Label     string    `json:"label"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}