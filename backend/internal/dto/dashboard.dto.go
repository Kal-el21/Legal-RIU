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