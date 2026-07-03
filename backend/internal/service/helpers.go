package service

import (
	"errors"
	"strings"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"

	"github.com/google/uuid"
)

func parseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

func canAccessAllSubmissions(role string) bool {
	return role == string(entity.RoleAdmin) || role == string(entity.RoleLegal)
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

	resp := dto.UserResponse{
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
		CompanyID:          "",
		PurposeTypeID:      "",
	}
	if user.CompanyID != nil {
		resp.CompanyID = user.CompanyID.String()
	}
	if user.Company.ID != uuid.Nil {
		company := toCompanyResponse(&user.Company)
		resp.Company = &company
	}
	if user.PurposeTypeID != nil {
		resp.PurposeTypeID = user.PurposeTypeID.String()
	}
	if user.PurposeType.ID != uuid.Nil {
		pt := toPurposeTypeResponse(&user.PurposeType)
		resp.PurposeType = &pt
	}
	return resp
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

func resolveCompanySelection(userRepo repository.UserRepository, role string, companyIDStr string) (*uuid.UUID, error) {
	switch role {
	case "EXTERNAL":
		return nil, nil
	case "LEGAL_AU":
		if companyIDStr == "" {
			return nil, errors.New("perusahaan wajib dipilih untuk role LEGAL_AU")
		}
		companyID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return nil, errors.New("ID perusahaan tidak valid")
		}
		company, err := userRepo.FindCompanyByID(companyID)
		if err != nil {
			return nil, errors.New("perusahaan tidak ditemukan")
		}
		return &company.ID, nil
	default:
		if companyIDStr != "" {
			companyID, err := uuid.Parse(companyIDStr)
			if err != nil {
				return nil, errors.New("ID perusahaan tidak valid")
			}
			company, err := userRepo.FindCompanyByID(companyID)
			if err != nil {
				return nil, errors.New("perusahaan tidak ditemukan")
			}
			return &company.ID, nil
		}
		company, err := userRepo.FindCompanyByDomain("indonesiare.co.id")
		if err == nil {
			return &company.ID, nil
		}
		company, err = userRepo.FindFirstCompany()
		if err != nil {
			return nil, errors.New("tidak ada perusahaan yang tersedia")
		}
		return &company.ID, nil
	}
}

func toDivisionResponse(division *entity.Division) dto.DivisionResponse {
	return dto.DivisionResponse{
		ID:          division.ID.String(),
		Name:        division.Name,
		Description: division.Description,
		CreatedAt:   division.CreatedAt,
		UpdatedAt:   division.UpdatedAt,
	}
}

func toDivisionResponsePointer(division *entity.Division) *dto.DivisionResponse {
	if division == nil || division.ID == uuid.Nil {
		return nil
	}
	resp := toDivisionResponse(division)
	return &resp
}

func toCompanyResponse(company *entity.Company) dto.CompanyResponse {
	return dto.CompanyResponse{
		ID:          company.ID.String(),
		Name:        company.Name,
		EmailDomain: company.EmailDomain,
		IsInternal:  company.IsInternal,
		CreatedAt:   company.CreatedAt,
		UpdatedAt:   company.UpdatedAt,
	}
}

func toPurposeTypeResponse(pt *entity.PurposeType) dto.PurposeTypeResponse {
	return dto.PurposeTypeResponse{
		ID:          pt.ID.String(),
		Name:        pt.Name,
		Description: pt.Description,
		IsActive:    pt.IsActive,
		CreatedAt:   pt.CreatedAt,
		UpdatedAt:   pt.UpdatedAt,
	}
}

func toCaseTypeResponse(caseType *entity.CaseType) dto.CaseTypeResponse {
	return dto.CaseTypeResponse{
		ID:        caseType.ID.String(),
		Code:      caseType.Code,
		Label:     caseType.Label,
		IsActive:  caseType.IsActive,
		CreatedAt: caseType.CreatedAt,
		UpdatedAt: caseType.UpdatedAt,
	}
}

func toCaseCategoryResponse(caseCategory *entity.CaseCategory) dto.CaseCategoryResponse {
	return dto.CaseCategoryResponse{
		ID:        caseCategory.ID.String(),
		Code:      caseCategory.Code,
		Label:     caseCategory.Label,
		IsActive:  caseCategory.IsActive,
		CreatedAt: caseCategory.CreatedAt,
		UpdatedAt: caseCategory.UpdatedAt,
	}
}
