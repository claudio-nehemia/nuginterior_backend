package entity

import (
	"time"
	"gorm.io/gorm"
)

type Moodboard struct {
	ID                    uint           `gorm:"primaryKey" json:"id"`
	OrderID               uint           `gorm:"not null;uniqueIndex" json:"order_id"`
	ResponseTime          *time.Time     `json:"response_time"`
	ResponseBy            string         `gorm:"type:varchar(100)" json:"response_by"`
	MarketingResponse     string         `gorm:"type:text" json:"marketing_response"`
	MarketingResponseBy   string         `gorm:"type:varchar(100)" json:"marketing_response_by"`
	MarketingResponseTime *time.Time     `json:"marketing_response_time"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Order         *Order          `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Files         []MoodboardFile `gorm:"foreignKey:MoodboardID" json:"files,omitempty"`
	Estimasi      *Estimasi       `gorm:"foreignKey:MoodboardID" json:"estimasi,omitempty"`
	CommitmentFee *CommitmentFee  `gorm:"foreignKey:MoodboardID" json:"commitment_fee,omitempty"`
}

func (Moodboard) TableName() string {
	return "moodboards"
}
