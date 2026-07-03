package dto

import "time"

type PermissionResponse struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Feature     string    `json:"feature"`
	Action      string    `json:"action"`
	Scope       string    `json:"scope"`
	Label       string    `json:"label"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserPermissionOverrideRequest struct {
	Code   string `json:"code" binding:"required"`
	Effect string `json:"effect" binding:"required,oneof=ALLOW DENY"`
}

type UpdateUserPermissionOverridesRequest struct {
	Overrides []UserPermissionOverrideRequest `json:"overrides"`
}

type UserPermissionOverrideResponse struct {
	Code      string    `json:"code"`
	Effect    string    `json:"effect"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserPermissionAccessResponse struct {
	UserID               string                           `json:"user_id"`
	Role                 string                           `json:"role"`
	Permissions          []PermissionResponse             `json:"permissions"`
	RolePermissions      []string                         `json:"role_permissions"`
	Overrides            []UserPermissionOverrideResponse `json:"overrides"`
	EffectivePermissions []string                         `json:"effective_permissions"`
}
