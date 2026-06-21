package entity

import (
	"time"
	"gorm.io/gorm"
)

type MoodboardFile struct {
	ID           uint              `gorm:"primaryKey" json:"id"`
	MoodboardID  uint              `gorm:"not null;index" json:"moodboard_id"`
	FilePath     string            `gorm:"type:varchar(255);not null" json:"file_path"`
	OriginalName string            `gorm:"type:varchar(255);not null" json:"original_name"`
	Status       string            `gorm:"type:varchar(50);default:'pending'" json:"status"`
	Revisi       string            `gorm:"type:text" json:"revisi"`
	CreatedAt    time.Time         `json:"created_at"`	
	UpdatedAt    time.Time         `json:"updated_at"`
	DeletedAt    gorm.DeletedAt    `gorm:"index" json:"-"`
}

func (MoodboardFile) TableName() string {
	return "moodboard_files"
}
