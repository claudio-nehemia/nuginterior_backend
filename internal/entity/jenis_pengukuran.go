package entity

import (
	"time"

	"gorm.io/gorm"
)

// JenisPengukuran represents the jenis_pengukuran table (soft delete enabled).
type JenisPengukuran struct {
	ID              uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	NamaPengukuran  string         `gorm:"size:255;not null" json:"nama_pengukuran"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (JenisPengukuran) TableName() string {
	return "jenis_pengukuran"
}
