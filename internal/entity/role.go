package entity

import (
	"time"
)

// Role represents the roles table.
type Role struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	CompanyID uint   `gorm:"not null;default:1;uniqueIndex:idx_company_role" json:"company_id"`
	NamaRole  string `gorm:"size:255;not null;uniqueIndex:idx_company_role" json:"nama_role"`
	DivisiID  uint   `gorm:"not null" json:"divisi_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Divisi      Divisi       `gorm:"foreignKey:DivisiID" json:"divisi,omitempty"`
	Users       []User       `gorm:"foreignKey:RoleID" json:"users,omitempty"`
	Permissions []Permission `gorm:"many2many:role_permission;" json:"permissions,omitempty"`
}

func (Role) TableName() string {
	return "roles"
}
