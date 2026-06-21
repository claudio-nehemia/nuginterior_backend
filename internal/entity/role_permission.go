package entity

import (
	"time"
)

// RolePermission represents the role_permission pivot table.
type RolePermission struct {
	ID           uint `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleID       uint `gorm:"not null;uniqueIndex:idx_role_perm" json:"role_id"`
	PermissionID uint `gorm:"not null;uniqueIndex:idx_role_perm" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (RolePermission) TableName() string {
	return "role_permission"
}
