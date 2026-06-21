package entity

import (
	"time"
)

// User represents the users table.
type User struct {
	ID                uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	CompanyID         uint       `gorm:"not null;default:1" json:"company_id"`
	Name              string     `gorm:"size:255;not null" json:"name"`
	Email             string     `gorm:"size:255;not null;uniqueIndex" json:"email"`
	EmailVerifiedAt   *time.Time `json:"email_verified_at,omitempty"`
	Password          string     `gorm:"size:255;not null" json:"-"`
	RoleID            *uint      `json:"role_id,omitempty"`
	FcmToken          string     `gorm:"type:text" json:"fcm_token,omitempty"`
	DevicePlatform    string     `gorm:"size:50" json:"device_platform,omitempty"`
	FcmTokenUpdatedAt *time.Time `json:"fcm_token_updated_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	// Relations
	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Role    *Role    `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

func (User) TableName() string {
	return "users"
}
