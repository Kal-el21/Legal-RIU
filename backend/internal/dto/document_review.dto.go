package dto

type CreateDocumentReviewRequest struct {
	RequestorName     string `json:"requestor_name"`
	RequestorPosition string `json:"requestor_position"`
	RequestorDivision string `json:"requestor_division"`
	RequestorEmail    string `json:"requestor_email"`
	RequestorPhone    string `json:"requestor_phone"`
	DocumentName      string `json:"document_name"`
	SecondParty       string `json:"second_party"`
	ThirdParty        string `json:"third_party"`
	DocumentType      string `json:"document_type"`
	DocumentTypeOther string `json:"document_type_other"`
	AdditionalNote    string `json:"additional_note"`
}

type UpdateDocumentReviewRequest struct {
	RequestorName     string `json:"requestor_name" binding:"required"`
	RequestorPosition string `json:"requestor_position" binding:"required"`
	RequestorDivision string `json:"requestor_division" binding:"required"`
	RequestorEmail    string `json:"requestor_email" binding:"required,email"`
	RequestorPhone    string `json:"requestor_phone" binding:"required"`
	DocumentName      string `json:"document_name" binding:"required"`
	SecondParty       string `json:"second_party" binding:"required"`
	ThirdParty        string `json:"third_party"`
	DocumentType      string `json:"document_type" binding:"required"`
	DocumentTypeOther string `json:"document_type_other"`
	AdditionalNote    string `json:"additional_note"`
}

type DocumentReviewListQuery struct {
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
	Status string `form:"status"`
}
