package entity

import (
	"time"
	"gorm.io/gorm"
)

type EstimasiFile struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	EstimasiID      uint           `gorm:"not null;index" json:"estimasi_id"`
	MoodboardFileID uint           `gorm:"not null;index" json:"moodboard_file_id"` // Matches specific Desain Kasar option
	FilePath        string         `gorm:"type:varchar(255);not null" json:"file_path"`
	OriginalName    string         `gorm:"type:varchar(255);not null" json:"original_name"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (EstimasiFile) TableName() string {
	return "estimasi_files"
}
