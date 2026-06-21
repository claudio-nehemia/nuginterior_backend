package dto

import "time"

// ── Permission Response DTOs ──

type PermissionResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Group       string `json:"group"`
}

type PermissionGroupedResponse struct {
	Groups map[string][]PermissionResponse `json:"groups"`
}

// ── User Request DTOs ──

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=255"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	RoleID   *uint  `json:"role_id" binding:"omitempty"`
}

type UpdateUserRequest struct {
	Name     string `json:"name" binding:"omitempty,min=2,max=255"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"omitempty,min=8"`
	RoleID   *uint  `json:"role_id" binding:"omitempty"`
}

// ── User Response DTOs ──

type UserResponse struct {
	ID              uint              `json:"id"`
	Name            string            `json:"name"`
	Email           string            `json:"email"`
	EmailVerifiedAt *time.Time        `json:"email_verified_at,omitempty"`
	RoleID          *uint             `json:"role_id,omitempty"`
	Role            *RoleSimple       `json:"role,omitempty"`
	DivisiName      string            `json:"divisi_name,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}
