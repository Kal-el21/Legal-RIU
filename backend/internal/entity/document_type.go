package entity

type DocumentType struct {
	Base
	Name     string `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Label    string `gorm:"size:100;not null" json:"label"`
	IsActive bool   `gorm:"not null;default:true" json:"is_active"`
}