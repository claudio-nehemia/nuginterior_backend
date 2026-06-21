package entity

import (
	"time"

	"gorm.io/gorm"
)

type GambarKerjaFile struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	GambarKerjaID uint           `gorm:"not null;index" json:"gambar_kerja_id"`
	FilePath      string         `gorm:"size:255;not null" json:"file_path"`
	OriginalName  string         `gorm:"size:255;not null" json:"original_name"`
	Status        string         `gorm:"size:50;not null;default:'pending'" json:"status"`
	Revisi        string         `gorm:"type:text" json:"revisi"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (GambarKerjaFile) TableName() string {
	return "gambar_kerja_files"
}
