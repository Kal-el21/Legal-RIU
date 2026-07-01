package service

import (
	"errors"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
)

type DashboardService interface {
	GetUserStats(userID string) (*dto.UserDashboardStats, error)
	GetUserRecentActivity(userID string) (map[string]interface{}, error)
	GetAdminStats() (*dto.AdminDashboardStats, error)
	GetAdminRecentActivity() (map[string]interface{}, error)
	GetLegalStats() (*dto.AdminDashboardStats, error)
	GetLegalRecentActivity() (map[string]interface{}, error)
	GetExternalStats() (*dto.AdminDashboardStats, error)
	GetExternalRecentActivity() (map[string]interface{}, error)
	GetReminders(userID string, role string) (*dto.RemindersResponse, error)
}

type dashboardService struct {
	repo                repository.DashboardRepository
	notificationSetting NotificationSettingService
}

func NewDashboardService(repo repository.DashboardRepository, notificationSettingService NotificationSettingService) DashboardService {
	return &dashboardService{repo: repo, notificationSetting: notificationSettingService}
}

func (s *dashboardService) GetUserStats(userID string) (*dto.UserDashboardStats, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, errors.New("user tidak valid")
	}

	totalLO, _ := s.repo.CountLegalOpinionsByUser(uid)
	totalDR, _ := s.repo.CountDocumentReviewsByUser(uid)

	// Pending = SUBMITTED + UNDER_REVIEW + RESUBMITTED
	loPending, _ := s.repo.CountLegalOpinionsByUserAndStatus(uid, entity.StatusSubmitted)
	loUnder, _ := s.repo.CountLegalOpinionsByUserAndStatus(uid, entity.StatusUnderReview)
	loResub, _ := s.repo.CountLegalOpinionsByUserAndStatus(uid, entity.StatusResubmitted)
	drPending, _ := s.repo.CountDocumentReviewsByUserAndStatus(uid, entity.StatusSubmitted)
	drUnder, _ := s.repo.CountDocumentReviewsByUserAndStatus(uid, entity.StatusUnderReview)
	drResub, _ := s.repo.CountDocumentReviewsByUserAndStatus(uid, entity.StatusResubmitted)

	loRevision, _ := s.repo.CountLegalOpinionsByUserAndStatus(uid, entity.StatusNeedRevision)
	drRevision, _ := s.repo.CountDocumentReviewsByUserAndStatus(uid, entity.StatusNeedRevision)

	loCompleted, _ := s.repo.CountLegalOpinionsByUserAndStatus(uid, entity.StatusCompleted)
	drCompleted, _ := s.repo.CountDocumentReviewsByUserAndStatus(uid, entity.StatusCompleted)

	return &dto.UserDashboardStats{
		TotalLegalOpinions:   totalLO,
		TotalDocumentReviews: totalDR,
		Pending:              loPending + loUnder + loResub + drPending + drUnder + drResub,
		NeedRevision:         loRevision + drRevision,
		Completed:            loCompleted + drCompleted,
	}, nil
}

func (s *dashboardService) GetUserRecentActivity(userID string) (map[string]interface{}, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, errors.New("user tidak valid")
	}

	legalOpinions, _ := s.repo.RecentLegalOpinionsByUser(uid, 5)
	documentReviews, _ := s.repo.RecentDocumentReviewsByUser(uid, 5)

	return map[string]interface{}{
		"legal_opinions":   legalOpinions,
		"document_reviews": documentReviews,
	}, nil
}

func (s *dashboardService) GetAdminStats() (*dto.AdminDashboardStats, error) {
	totalUsers, _ := s.repo.CountAllUsers()
	totalLO, _ := s.repo.CountAllLegalOpinions()
	totalDR, _ := s.repo.CountAllDocumentReviews()

	// Pending Review = SUBMITTED dari semua
	loSubmitted, _ := s.repo.CountLegalOpinionsByStatus(entity.StatusSubmitted)
	drSubmitted, _ := s.repo.CountDocumentReviewsByStatus(entity.StatusSubmitted)

	loRevision, _ := s.repo.CountLegalOpinionsByStatus(entity.StatusNeedRevision)
	drRevision, _ := s.repo.CountDocumentReviewsByStatus(entity.StatusNeedRevision)

	loResub, _ := s.repo.CountLegalOpinionsByStatus(entity.StatusResubmitted)
	drResub, _ := s.repo.CountDocumentReviewsByStatus(entity.StatusResubmitted)

	return &dto.AdminDashboardStats{
		TotalUsers:           totalUsers,
		TotalLegalOpinions:   totalLO,
		TotalDocumentReviews: totalDR,
		PendingReview:        loSubmitted + drSubmitted,
		NeedRevision:         loRevision + drRevision,
		Resubmitted:          loResub + drResub,
	}, nil
}

func (s *dashboardService) GetAdminRecentActivity() (map[string]interface{}, error) {
	legalOpinions, _ := s.repo.RecentAllLegalOpinions(5)
	documentReviews, _ := s.repo.RecentAllDocumentReviews(5)

	return map[string]interface{}{
		"legal_opinions":   legalOpinions,
		"document_reviews": documentReviews,
	}, nil
}

func (s *dashboardService) GetLegalStats() (*dto.AdminDashboardStats, error) {
	totalLO, _ := s.repo.CountAllLegalOpinions()
	totalDR, _ := s.repo.CountAllDocumentReviews()

	// Pending Review = SUBMITTED dari semua
	loSubmitted, _ := s.repo.CountLegalOpinionsByStatus(entity.StatusSubmitted)
	drSubmitted, _ := s.repo.CountDocumentReviewsByStatus(entity.StatusSubmitted)

	loRevision, _ := s.repo.CountLegalOpinionsByStatus(entity.StatusNeedRevision)
	drRevision, _ := s.repo.CountDocumentReviewsByStatus(entity.StatusNeedRevision)

	loResub, _ := s.repo.CountLegalOpinionsByStatus(entity.StatusResubmitted)
	drResub, _ := s.repo.CountDocumentReviewsByStatus(entity.StatusResubmitted)

	return &dto.AdminDashboardStats{
		TotalLegalOpinions:   totalLO,
		TotalDocumentReviews: totalDR,
		PendingReview:        loSubmitted + drSubmitted,
		NeedRevision:         loRevision + drRevision,
		Resubmitted:          loResub + drResub,
	}, nil
}

func (s *dashboardService) GetLegalRecentActivity() (map[string]interface{}, error) {
	legalOpinions, _ := s.repo.RecentAllLegalOpinions(5)
	documentReviews, _ := s.repo.RecentAllDocumentReviews(5)

	return map[string]interface{}{
		"legal_opinions":   legalOpinions,
		"document_reviews": documentReviews,
	}, nil
}

func (s *dashboardService) GetExternalStats() (*dto.AdminDashboardStats, error) {
	totalLO, _ := s.repo.CountAllLegalOpinions()
	totalDR, _ := s.repo.CountAllDocumentReviews()

	// Pending Review = SUBMITTED dari semua
	loSubmitted, _ := s.repo.CountLegalOpinionsByStatus(entity.StatusSubmitted)
	drSubmitted, _ := s.repo.CountDocumentReviewsByStatus(entity.StatusSubmitted)

	loRevision, _ := s.repo.CountLegalOpinionsByStatus(entity.StatusNeedRevision)
	drRevision, _ := s.repo.CountDocumentReviewsByStatus(entity.StatusNeedRevision)

	loResub, _ := s.repo.CountLegalOpinionsByStatus(entity.StatusResubmitted)
	drResub, _ := s.repo.CountDocumentReviewsByStatus(entity.StatusResubmitted)

	return &dto.AdminDashboardStats{
		TotalLegalOpinions:   totalLO,
		TotalDocumentReviews: totalDR,
		PendingReview:        loSubmitted + drSubmitted,
		NeedRevision:         loRevision + drRevision,
		Resubmitted:          loResub + drResub,
	}, nil
}

func (s *dashboardService) GetExternalRecentActivity() (map[string]interface{}, error) {
	legalOpinions, _ := s.repo.RecentAllLegalOpinions(5)
	documentReviews, _ := s.repo.RecentAllDocumentReviews(5)

	return map[string]interface{}{
		"legal_opinions":   legalOpinions,
		"document_reviews": documentReviews,
	}, nil
}

func (s *dashboardService) GetReminders(userID string, role string) (*dto.RemindersResponse, error) {
	return s.notificationSetting.GetReminders(userID, role)
}
