package service

import (
	"github.com/google/uuid"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
)

func parseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

func toUserResponse(user *entity.User) dto.UserResponse {
	return dto.UserResponse{
		ID:                  user.ID.String(),
		FullName:            user.FullName,
		Email:               user.Email,
		Position:            user.Position,
		Division:            user.Division,
		Role:                string(user.Role),
		Status:              string(user.Status),
		EmailNotifications:  user.EmailNotifications,
		TwoFAEnabled:        user.TwoFAEnabled,
	}
}
