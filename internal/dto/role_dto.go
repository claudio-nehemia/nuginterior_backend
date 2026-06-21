package dto

import "time"

// ── Role Request DTOs ──

type CreateRoleRequest struct {
	NamaRole    string `json:"nama_role" binding:"required,min=2,max=255"`
	DivisiID    uint   `json:"divisi_id" binding:"required"`
	Permissions []uint `json:"permissions" binding:"omitempty"`
}

type UpdateRoleRequest struct {
	NamaRole    string `json:"nama_role" binding:"required,min=2,max=255"`
	DivisiID    uint   `json:"divisi_id" binding:"required"`
	Permissions []uint `json:"permissions" binding:"omitempty"`
}

type SyncPermissionsRequest struct {
	Permissions []uint `json:"permissions"`
}

// ── Role Response DTOs ──

type RoleResponse struct {
	ID               uint                `json:"id"`
	NamaRole         string              `json:"nama_role"`
	DivisiID         uint                `json:"divisi_id"`
	Divisi           *DivisiResponse     `json:"divisi,omitempty"`
	UsersCount       int64               `json:"users_count"`
	PermissionsCount int64               `json:"permissions_count"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

type RoleDetailResponse struct {
	ID               uint                   `json:"id"`
	NamaRole         string                 `json:"nama_role"`
	DivisiID         uint                   `json:"divisi_id"`
	Divisi           *DivisiResponse        `json:"divisi,omitempty"`
	Users            []UserResponse         `json:"users,omitempty"`
	Permissions      []PermissionResponse   `json:"permissions,omitempty"`
	PermissionsGrouped map[string][]PermissionResponse `json:"permissions_grouped,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}
