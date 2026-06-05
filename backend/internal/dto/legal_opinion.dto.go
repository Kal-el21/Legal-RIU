package dto

type CreateLegalOpinionRequest struct {
	RequestorName     string `json:"requestor_name" binding:"required"`
	RequestorPosition string `json:"requestor_position" binding:"required"`
	RequestorDivision string `json:"requestor_division" binding:"required"`
	RequestorEmail    string `json:"requestor_email" binding:"required,email"`
	RequestorPhone    string `json:"requestor_phone" binding:"required"`
	LegalType         string `json:"legal_type" binding:"required"`
	LegalTypeOther    string `json:"legal_type_other"`
	Title             string `json:"title" binding:"required"`
	Chronology        string `json:"chronology" binding:"required"`
	Question          string `json:"question" binding:"required"`
}

type UpdateLegalOpinionRequest struct {
	RequestorName     string `json:"requestor_name" binding:"required"`
	RequestorPosition string `json:"requestor_position" binding:"required"`
	RequestorDivision string `json:"requestor_division" binding:"required"`
	RequestorEmail    string `json:"requestor_email" binding:"required,email"`
	RequestorPhone    string `json:"requestor_phone" binding:"required"`
	LegalType         string `json:"legal_type" binding:"required"`
	LegalTypeOther    string `json:"legal_type_other"`
	Title             string `json:"title" binding:"required"`
	Chronology        string `json:"chronology" binding:"required"`
	Question          string `json:"question" binding:"required"`
}

type UpdateStatusRequest struct {
	Status    string `json:"status" binding:"required"`
	AdminNote string `json:"admin_note"`
}

type UploadResultRequest struct {
	Notes string `json:"notes"`
}

type LegalOpinionListQuery struct {
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
	Status string `form:"status"`
}
