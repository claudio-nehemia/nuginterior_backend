package entity

import (
	"time"
	"gorm.io/gorm"
)

type DesainFinal struct {
	ID                    uint              `gorm:"primaryKey" json:"id"`
	OrderID               uint              `gorm:"not null;uniqueIndex" json:"order_id"`
	ResponseTime          *time.Time        `json:"response_time"`
	ResponseBy            string            `gorm:"type:varchar(100)" json:"response_by"`
	MarketingResponseTime *time.Time        `json:"marketing_response_time"`
	MarketingResponseBy   string            `gorm:"type:varchar(100)" json:"marketing_response_by"`
	Status                string            `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, uploaded, revision, accepted
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`
	DeletedAt             gorm.DeletedAt    `gorm:"index" json:"-"`

	// Relationships
	Order *Order            `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Files []DesainFinalFile `gorm:"foreignKey:DesainFinalID" json:"files,omitempty"`
}

func (DesainFinal) TableName() string {
	return "desain_finals"
}

type DesainFinalFile struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	DesainFinalID uint           `gorm:"not null;index" json:"desain_final_id"`
	FilePath      string         `gorm:"type:varchar(255);not null" json:"file_path"`
	OriginalName  string         `gorm:"type:varchar(255);not null" json:"original_name"`
	Status        string         `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, approved, revisi
	Revisi        string         `gorm:"type:text" json:"revisi"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DesainFinalFile) TableName() string {
	return "desain_final_files"
}
