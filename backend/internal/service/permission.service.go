package service

import (
	"errors"
	"sort"
	"strings"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"

	"github.com/google/uuid"
)

type PermissionService interface {
	GetCatalog() ([]dto.PermissionResponse, error)
	GetEffectivePermissionCodes(userID string, role string) ([]string, error)
	HasPermission(userID string, role string, code string) bool
	HasAnyPermission(userID string, role string, codes ...string) bool
	GetUserAccess(userID string) (*dto.UserPermissionAccessResponse, error)
	UpdateUserOverrides(userID string, req dto.UpdateUserPermissionOverridesRequest, adminID string) (*dto.UserPermissionAccessResponse, error)
}

type permissionService struct {
	permissionRepo repository.PermissionRepository
	userRepo       repository.UserRepository
}

func NewPermissionService(permissionRepo repository.PermissionRepository, userRepo repository.UserRepository) PermissionService {
	return &permissionService{permissionRepo: permissionRepo, userRepo: userRepo}
}

func (s *permissionService) GetCatalog() ([]dto.PermissionResponse, error) {
	permissions, err := s.permissionRepo.FindAll()
	if err != nil {
		return nil, err
	}
	return toPermissionResponses(permissions), nil
}

func (s *permissionService) GetEffectivePermissionCodes(userID string, role string) ([]string, error) {
	permissions, err := s.resolveEffectivePermissions(userID, entity.UserRole(role))
	if err != nil {
		return nil, err
	}
	return permissionCodeList(permissions), nil
}

func (s *permissionService) HasPermission(userID string, role string, code string) bool {
	return s.HasAnyPermission(userID, role, code)
}

func (s *permissionService) HasAnyPermission(userID string, role string, codes ...string) bool {
	if len(codes) == 0 {
		return true
	}
	if role == string(entity.RoleAdmin) {
		return true
	}

	effective, err := s.GetEffectivePermissionCodes(userID, role)
	if err != nil {
		return false
	}

	allowed := make(map[string]bool, len(effective))
	for _, code := range effective {
		allowed[code] = true
	}
	for _, code := range codes {
		if allowed[code] {
			return true
		}
	}
	return false
}

func (s *permissionService) GetUserAccess(userID string) (*dto.UserPermissionAccessResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user tidak valid")
	}

	user, err := s.userRepo.FindByID(uid)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	return s.buildUserAccessResponse(user.ID, user.Role)
}

func (s *permissionService) UpdateUserOverrides(userID string, req dto.UpdateUserPermissionOverridesRequest, adminID string) (*dto.UserPermissionAccessResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user tidak valid")
	}

	user, err := s.userRepo.FindByID(uid)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	adminUUID, _ := uuid.Parse(adminID)
	codeEffect := map[string]entity.PermissionEffect{}
	var codes []string
	for _, item := range req.Overrides {
		code := strings.TrimSpace(item.Code)
		if code == "" {
			return nil, errors.New("kode permission wajib diisi")
		}

		effect := entity.PermissionEffect(strings.ToUpper(strings.TrimSpace(item.Effect)))
		if effect != entity.PermissionAllow && effect != entity.PermissionDeny {
			return nil, errors.New("effect permission tidak valid")
		}

		if _, exists := codeEffect[code]; !exists {
			codes = append(codes, code)
		}
		codeEffect[code] = effect
	}

	permissions, err := s.permissionRepo.FindByCodes(codes)
	if err != nil {
		return nil, err
	}
	permissionByCode := make(map[string]entity.Permission, len(permissions))
	for _, permission := range permissions {
		permissionByCode[permission.Code] = permission
	}
	if len(permissionByCode) != len(codeEffect) {
		return nil, errors.New("terdapat permission yang tidak valid")
	}

	overrides := make([]entity.UserPermissionOverride, 0, len(codeEffect))
	sort.Strings(codes)
	for _, code := range codes {
		permission := permissionByCode[code]
		override := entity.UserPermissionOverride{
			UserID:       user.ID,
			PermissionID: permission.ID,
			Effect:       codeEffect[code],
		}
		if adminUUID != uuid.Nil {
			override.CreatedBy = &adminUUID
			override.UpdatedBy = &adminUUID
		}
		overrides = append(overrides, override)
	}

	if err := s.permissionRepo.ReplaceUserOverrides(user.ID, overrides); err != nil {
		return nil, err
	}

	return s.buildUserAccessResponse(user.ID, user.Role)
}

func (s *permissionService) buildUserAccessResponse(userID uuid.UUID, role entity.UserRole) (*dto.UserPermissionAccessResponse, error) {
	catalog, err := s.permissionRepo.FindAll()
	if err != nil {
		return nil, err
	}
	rolePermissions, err := s.rolePermissionCodes(role, catalog)
	if err != nil {
		return nil, err
	}
	overrides, err := s.permissionRepo.FindUserOverrides(userID)
	if err != nil {
		return nil, err
	}
	effective, err := s.resolveEffectivePermissions(userID.String(), role)
	if err != nil {
		return nil, err
	}

	return &dto.UserPermissionAccessResponse{
		UserID:               userID.String(),
		Role:                 string(role),
		Permissions:          toPermissionResponses(catalog),
		RolePermissions:      rolePermissions,
		Overrides:            toUserPermissionOverrideResponses(overrides),
		EffectivePermissions: permissionCodeList(effective),
	}, nil
}

func (s *permissionService) resolveEffectivePermissions(userID string, role entity.UserRole) (map[string]entity.Permission, error) {
	catalog, err := s.permissionRepo.FindAll()
	if err != nil {
		return nil, err
	}

	baseCodes, err := s.rolePermissionCodes(role, catalog)
	if err != nil {
		return nil, err
	}

	permissionByCode := make(map[string]entity.Permission, len(catalog))
	for _, permission := range catalog {
		permissionByCode[permission.Code] = permission
	}

	effective := make(map[string]entity.Permission, len(baseCodes))
	for _, code := range baseCodes {
		if permission, ok := permissionByCode[code]; ok {
			effective[code] = permission
		}
	}

	if role == entity.RoleAdmin {
		return effective, nil
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return effective, nil
	}

	overrides, err := s.permissionRepo.FindUserOverrides(uid)
	if err != nil {
		return nil, err
	}
	for _, override := range overrides {
		code := override.Permission.Code
		if code == "" {
			continue
		}
		if override.Effect == entity.PermissionDeny {
			delete(effective, code)
			continue
		}
		if permission, ok := permissionByCode[code]; ok {
			effective[code] = permission
		}
	}

	return effective, nil
}

func (s *permissionService) rolePermissionCodes(role entity.UserRole, catalog []entity.Permission) ([]string, error) {
	if role == entity.RoleAdmin {
		codes := make([]string, 0, len(catalog))
		for _, permission := range catalog {
			codes = append(codes, permission.Code)
		}
		sort.Strings(codes)
		return codes, nil
	}

	codes, err := s.permissionRepo.FindRolePermissionCodes(role)
	if err != nil {
		return nil, err
	}
	sort.Strings(codes)
	return codes, nil
}

func toPermissionResponses(permissions []entity.Permission) []dto.PermissionResponse {
	responses := make([]dto.PermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		responses = append(responses, dto.PermissionResponse{
			ID:          permission.ID.String(),
			Code:        permission.Code,
			Feature:     permission.Feature,
			Action:      permission.Action,
			Scope:       permission.Scope,
			Label:       permission.Label,
			Description: permission.Description,
			IsActive:    permission.IsActive,
			CreatedAt:   permission.CreatedAt,
			UpdatedAt:   permission.UpdatedAt,
		})
	}
	return responses
}

func toUserPermissionOverrideResponses(overrides []entity.UserPermissionOverride) []dto.UserPermissionOverrideResponse {
	responses := make([]dto.UserPermissionOverrideResponse, 0, len(overrides))
	for _, override := range overrides {
		responses = append(responses, dto.UserPermissionOverrideResponse{
			Code:      override.Permission.Code,
			Effect:    string(override.Effect),
			UpdatedAt: override.UpdatedAt,
		})
	}
	sort.Slice(responses, func(i, j int) bool {
		return responses[i].Code < responses[j].Code
	})
	return responses
}

func permissionCodeList(permissions map[string]entity.Permission) []string {
	codes := make([]string, 0, len(permissions))
	for code := range permissions {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}
