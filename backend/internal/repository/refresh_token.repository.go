package repository

import (
	"time"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token *entity.RefreshToken) error
	FindActiveByHash(tokenHash string) (*entity.RefreshToken, error)
	RevokeByHash(tokenHash string) error
	RevokeAllByUserID(userID uuid.UUID) error
	UpdateLastUsed(tokenHash string) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token *entity.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepository) FindActiveByHash(tokenHash string) (*entity.RefreshToken, error) {
	var token entity.RefreshToken
	err := r.db.
		Preload("User").
		Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", tokenHash, time.Now()).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *refreshTokenRepository) RevokeByHash(tokenHash string) error {
	now := time.Now()
	return r.db.Model(&entity.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", &now).Error
}

func (r *refreshTokenRepository) RevokeAllByUserID(userID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&entity.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", &now).Error
}

func (r *refreshTokenRepository) UpdateLastUsed(tokenHash string) error {
	now := time.Now()
	return r.db.Model(&entity.RefreshToken{}).
		Where("token_hash = ?", tokenHash).
		Update("last_used_at", &now).Error
}