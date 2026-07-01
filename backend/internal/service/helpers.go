package service

import (
	"errors"
	"strings"

	"legal-riu-portal/internal/repository"

	"github.com/google/uuid"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
)

func parseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

func toUserResponse(user *entity.User) dto.UserResponse {
	division := user.Division
	divisionID := ""
	var divisionDetail *dto.DivisionResponse
	if user.DivisionRef.ID != uuid.Nil {
		division = user.DivisionRef.Name
		divisionID = user.DivisionRef.ID.String()
		resp := toDivisionResponse(&user.DivisionRef)
		divisionDetail = &resp
	} else if user.DivisionID != nil && *user.DivisionID != uuid.Nil {
		divisionID = user.DivisionID.String()
	}

	return dto.UserResponse{
		ID:                 user.ID.String(),
		FullName:           user.FullName,
		Email:              user.Email,
		Position:           user.Position,
		Division:           division,
		DivisionID:         divisionID,
		DivisionDetail:     divisionDetail,
		Role:               string(user.Role),
		Status:             string(user.Status),
		EmailNotifications: user.EmailNotifications,
		TwoFAEnabled:       user.TwoFAEnabled,
	}
}

func resolveDivisionSelection(userRepo repository.UserRepository, value string) (string, *uuid.UUID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil, errors.New("divisi wajib diisi")
	}

	if id, err := uuid.Parse(value); err == nil {
		division, err := userRepo.FindDivisionByID(id)
		if err != nil {
			return "", nil, errors.New("divisi tidak ditemukan")
		}
		divisionID := division.ID
		return division.Name, &divisionID, nil
	}

	division, err := userRepo.FindDivisionByName(value)
	if err != nil {
		return "", nil, errors.New("divisi tidak ditemukan")
	}
	divisionID := division.ID
	return division.Name, &divisionID, nil
}
