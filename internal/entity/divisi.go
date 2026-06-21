package entity

import (
	"time"
)

// Divisi represents the divisis table.
type Divisi struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CompanyID  uint      `gorm:"not null;default:1;uniqueIndex:idx_company_divisi" json:"company_id"`
	NamaDivisi string    `gorm:"size:255;not null;uniqueIndex:idx_company_divisi" json:"nama_divisi"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relations
	Roles []Role `gorm:"foreignKey:DivisiID" json:"roles,omitempty"`
}

func (Divisi) TableName() string {
	return "divisis"
}
