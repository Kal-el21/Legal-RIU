package dto

// ─── Auth ─────────────────────────────────────────────────────────────────────

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	Token        string       `json:"token"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// ─── User ─────────────────────────────────────────────────────────────────────

type UserResponse struct {
	ID                 string            `json:"id"`
	FullName           string            `json:"full_name"`
	Email              string            `json:"email"`
	AuthType           string            `json:"auth_type"`
	Position           string            `json:"position"`
	Division           string            `json:"division"`
	DivisionID         string            `json:"division_id,omitempty"`
	DivisionDetail     *DivisionResponse `json:"division_detail,omitempty"`
	Role               string            `json:"role"`
	Status             string            `json:"status"`
	EmailNotifications bool              `json:"email_notifications"`
	TwoFAEnabled       bool              `json:"two_fa_enabled"`
}

type CreateUserRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Position string `json:"position" binding:"required"`
	Division string `json:"division" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=USER ADMIN LEGAL EXTERNAL"`
}

type UpdateUserRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Position string `json:"position" binding:"required"`
	Division string `json:"division" binding:"required"`
	Role     string `json:"role" binding:"omitempty,oneof=USER ADMIN LEGAL EXTERNAL"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type PaginationQuery struct {
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
	Search string `form:"search"`
}

// ─── Settings ─────────────────────────────────────────────────────────────────

type UpdateProfileRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Position string `json:"position" binding:"required"`
	Division string `json:"division" binding:"required"`
}

type UpdateNotificationRequest struct {
	EmailNotifications bool `json:"email_notifications"`
}

type Toggle2FARequest struct {
	Enabled  bool   `json:"enabled"`
	Password string `json:"password" binding:"required"`
}
