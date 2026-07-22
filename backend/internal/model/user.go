package model

import "time"

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	DisplayName     string     `json:"display_name"`
	Description     string     `json:"description"`
	Role            UserRole   `json:"role"`
	IsActive        bool       `json:"is_active"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	Language        string     `json:"language"`
	Theme           string     `json:"theme"`
	PasswordHash    string     `json:"-"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateUserRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

type UpdateUserRequest struct {
	Email       *string `json:"email,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type UpdateProfileRequest struct {
	Email       *string `json:"email,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	Description *string `json:"description,omitempty"`
	Password    *string `json:"password,omitempty"`
	Language    *string `json:"language,omitempty"`
	Theme       *string `json:"theme,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	DisplayName     string     `json:"display_name"`
	Description     string     `json:"description"`
	Role            UserRole   `json:"role"`
	IsActive        bool       `json:"is_active"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	Language        string     `json:"language"`
	Theme           string     `json:"theme"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:              u.ID,
		Email:           u.Email,
		DisplayName:     u.DisplayName,
		Description:     u.Description,
		Role:            u.Role,
		IsActive:        u.IsActive,
		EmailVerifiedAt: u.EmailVerifiedAt,
		Language:        u.Language,
		Theme:           u.Theme,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}
