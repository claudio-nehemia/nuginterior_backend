package entity

import (
	"time"
	"gorm.io/gorm"
)

type Estimasi struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	MoodboardID    uint           `gorm:"not null;uniqueIndex" json:"moodboard_id"`
	EstimatedCost  string         `gorm:"type:varchar(255)" json:"estimated_cost"` // Locked/Approved Estimasi Excel File Path
	ResponseBy     string         `gorm:"type:varchar(100)" json:"response_by"`
	ResponseTime   *time.Time     `json:"response_time"`
	PmResponseBy   string         `gorm:"type:varchar(100)" json:"pm_response_by"`
	PmResponseTime *time.Time     `json:"pm_response_time"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Files []EstimasiFile `gorm:"foreignKey:EstimasiID" json:"files,omitempty"`
}

func (Estimasi) TableName() string {
	return "estimasis"
}
