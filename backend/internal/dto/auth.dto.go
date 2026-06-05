package dto

// ─── Auth ─────────────────────────────────────────────────────────────────────

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// ─── User ─────────────────────────────────────────────────────────────────────

type UserResponse struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Position string `json:"position"`
	Division string `json:"division"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

type CreateUserRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Position string `json:"position" binding:"required"`
	Division string `json:"division" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=USER ADMIN"`
}

type UpdateUserRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Position string `json:"position" binding:"required"`
	Division string `json:"division" binding:"required"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type PaginationQuery struct {
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
	Search string `form:"search"`
}
