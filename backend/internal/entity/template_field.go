package entity

// TemplateFieldPosition stores the calibrated position of a fillable field on a
// given agreement template version. Coordinates are in millimetres relative to
// the top-left corner of an A4 page (210×297mm). Multiple rows share the same
// template_version, one per field name.
type TemplateFieldPosition struct {
	Base
	TemplateVersion string  `gorm:"size:50;not null;index:idx_template_field_version" json:"template_version"`
	FieldName       string  `gorm:"size:100;not null" json:"field_name"`
	X               float64 `gorm:"not null" json:"x"`
	Y               float64 `gorm:"not null" json:"y"`
	Font            string  `gorm:"size:50" json:"font"`
	Style           string  `gorm:"size:20" json:"style"`
	Size            float64 `gorm:"not null" json:"size"`
	Align           string  `gorm:"size:20;not null" json:"align"`
	PageNumber      int     `gorm:"not null;default:1" json:"page_number"`
}
