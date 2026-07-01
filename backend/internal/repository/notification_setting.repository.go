package repository

import (
	"time"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NotificationSettingRepository interface {
	GetAll() ([]entity.NotificationSetting, error)
	GetByID(id uuid.UUID) (*entity.NotificationSetting, error)
	GetByTypeAndLevel(submissionType, warningLevel string) ([]entity.NotificationSetting, error)
	Update(setting *entity.NotificationSetting) error
	GetNotificationReadsByUser(userID uuid.UUID) ([]entity.NotificationRead, error)
	MarkNotificationRead(userID uuid.UUID, submissionType string, submissionID uuid.UUID) error
	MarkNotificationsRead(userID uuid.UUID, items []entity.NotificationRead) error
}

type notificationSettingRepository struct {
	db *gorm.DB
}

func NewNotificationSettingRepository(db *gorm.DB) NotificationSettingRepository {
	return &notificationSettingRepository{db: db}
}

func (r *notificationSettingRepository) GetAll() ([]entity.NotificationSetting, error) {
	var settings []entity.NotificationSetting
	err := r.db.Find(&settings).Error
	return settings, err
}

func (r *notificationSettingRepository) GetByID(id uuid.UUID) (*entity.NotificationSetting, error) {
	var setting entity.NotificationSetting
	err := r.db.First(&setting, id).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *notificationSettingRepository) GetByTypeAndLevel(submissionType, warningLevel string) ([]entity.NotificationSetting, error) {
	var settings []entity.NotificationSetting
	err := r.db.Where("submission_type = ? AND warning_level = ?", submissionType, warningLevel).Find(&settings).Error
	return settings, err
}

func (r *notificationSettingRepository) Update(setting *entity.NotificationSetting) error {
	return r.db.Save(setting).Error
}

func (r *notificationSettingRepository) GetNotificationReadsByUser(userID uuid.UUID) ([]entity.NotificationRead, error) {
	var reads []entity.NotificationRead
	err := r.db.Where("user_id = ? AND is_read = ?", userID, true).Find(&reads).Error
	return reads, err
}

func (r *notificationSettingRepository) MarkNotificationRead(userID uuid.UUID, submissionType string, submissionID uuid.UUID) error {
	now := time.Now()
	read := entity.NotificationRead{
		UserID:         userID,
		SubmissionType: submissionType,
		SubmissionID:   submissionID,
		IsRead:         true,
		ReadAt:         &now,
	}

	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"},
			{Name: "submission_type"},
			{Name: "submission_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"is_read":    true,
			"read_at":    now,
			"updated_at": now,
		}),
	}).Create(&read).Error
}

func (r *notificationSettingRepository) MarkNotificationsRead(userID uuid.UUID, items []entity.NotificationRead) error {
	if len(items) == 0 {
		return nil
	}

	now := time.Now()
	for i := range items {
		items[i].UserID = userID
		items[i].IsRead = true
		items[i].ReadAt = &now
	}

	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"},
			{Name: "submission_type"},
			{Name: "submission_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"is_read":    true,
			"read_at":    now,
			"updated_at": now,
		}),
	}).Create(&items).Error
}
