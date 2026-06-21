package entity

import (
	"time"
)

// Permission represents the permissions table.
type Permission struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"size:255;not null;uniqueIndex" json:"name"`
	DisplayName string `gorm:"size:255" json:"display_name"`
	Group       string `gorm:"size:255;column:group" json:"group"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations
	Roles []Role `gorm:"many2many:role_permission;" json:"roles,omitempty"`
}

func (Permission) TableName() string {
	return "permissions"
}
