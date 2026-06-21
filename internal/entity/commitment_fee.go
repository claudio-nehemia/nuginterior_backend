package entity

import (
	"time"
	"gorm.io/gorm"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
)

type CommitmentFee struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	MoodboardID    uint           `gorm:"not null;uniqueIndex" json:"moodboard_id"`
	TotalFee       *float64       `gorm:"type:decimal(15,2)" json:"total_fee"`
	PaymentProof   string         `gorm:"type:varchar(255)" json:"payment_proof"`
	PaymentStatus  PaymentStatus  `gorm:"type:varchar(50);default:'pending'" json:"payment_status"`
	ResponseBy     string         `gorm:"type:varchar(100)" json:"response_by"`
	ResponseTime   *time.Time     `json:"response_time"`
	PmResponseBy   string         `gorm:"type:varchar(100)" json:"pm_response_by"`
	PmResponseTime *time.Time     `json:"pm_response_time"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CommitmentFee) TableName() string {
	return "commitment_fees"
}
