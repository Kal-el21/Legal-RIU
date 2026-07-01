package repository

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationSettingRepository interface {
	GetAll() ([]entity.NotificationSetting, error)
	GetByID(id uuid.UUID) (*entity.NotificationSetting, error)
	GetByTypeAndLevel(submissionType, warningLevel string) ([]entity.NotificationSetting, error)
	Update(setting *entity.NotificationSetting) error
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
