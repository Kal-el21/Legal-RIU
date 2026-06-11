package repository

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DashboardRepository interface {
	// User stats
	CountLegalOpinionsByUser(userID uuid.UUID) (int64, error)
	CountDocumentReviewsByUser(userID uuid.UUID) (int64, error)
	CountLegalOpinionsByUserAndStatus(userID uuid.UUID, status entity.SubmissionStatus) (int64, error)
	CountDocumentReviewsByUserAndStatus(userID uuid.UUID, status entity.SubmissionStatus) (int64, error)

	// Admin stats
	CountAllUsers() (int64, error)
	CountAllLegalOpinions() (int64, error)
	CountAllDocumentReviews() (int64, error)
	CountLegalOpinionsByStatus(status entity.SubmissionStatus) (int64, error)
	CountDocumentReviewsByStatus(status entity.SubmissionStatus) (int64, error)

	// Recent activities
	RecentLegalOpinionsByUser(userID uuid.UUID, limit int) ([]entity.LegalOpinion, error)
	RecentDocumentReviewsByUser(userID uuid.UUID, limit int) ([]entity.DocumentReview, error)
	RecentAllLegalOpinions(limit int) ([]entity.LegalOpinion, error)
	RecentAllDocumentReviews(limit int) ([]entity.DocumentReview, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) CountLegalOpinionsByUser(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&entity.LegalOpinion{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) CountDocumentReviewsByUser(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&entity.DocumentReview{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) CountLegalOpinionsByUserAndStatus(userID uuid.UUID, status entity.SubmissionStatus) (int64, error) {
	var count int64
	err := r.db.Model(&entity.LegalOpinion{}).Where("user_id = ? AND status = ?", userID, status).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) CountDocumentReviewsByUserAndStatus(userID uuid.UUID, status entity.SubmissionStatus) (int64, error) {
	var count int64
	err := r.db.Model(&entity.DocumentReview{}).Where("user_id = ? AND status = ?", userID, status).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) CountAllUsers() (int64, error) {
	var count int64
	err := r.db.Model(&entity.User{}).Where("status = ?", entity.UserActive).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) CountAllLegalOpinions() (int64, error) {
	var count int64
	err := r.db.Model(&entity.LegalOpinion{}).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) CountAllDocumentReviews() (int64, error) {
	var count int64
	err := r.db.Model(&entity.DocumentReview{}).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) CountLegalOpinionsByStatus(status entity.SubmissionStatus) (int64, error) {
	var count int64
	err := r.db.Model(&entity.LegalOpinion{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) CountDocumentReviewsByStatus(status entity.SubmissionStatus) (int64, error) {
	var count int64
	err := r.db.Model(&entity.DocumentReview{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *dashboardRepository) RecentLegalOpinionsByUser(userID uuid.UUID, limit int) ([]entity.LegalOpinion, error) {
	var items []entity.LegalOpinion
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&items).Error
	return items, err
}

func (r *dashboardRepository) RecentDocumentReviewsByUser(userID uuid.UUID, limit int) ([]entity.DocumentReview, error) {
	var items []entity.DocumentReview
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&items).Error
	return items, err
}

func (r *dashboardRepository) RecentAllLegalOpinions(limit int) ([]entity.LegalOpinion, error) {
	var items []entity.LegalOpinion
	err := r.db.Preload("User").Order("created_at DESC").Limit(limit).Find(&items).Error
	return items, err
}

func (r *dashboardRepository) RecentAllDocumentReviews(limit int) ([]entity.DocumentReview, error) {
	var items []entity.DocumentReview
	err := r.db.Preload("User").Order("created_at DESC").Limit(limit).Find(&items).Error
	return items, err
}
