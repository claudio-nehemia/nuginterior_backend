package entity

import (
	"time"

	"gorm.io/gorm"
)

type GambarKerja struct {
	ID                    uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID               uint            `gorm:"not null;uniqueIndex" json:"order_id"`
	Status                string          `gorm:"size:50;not null;default:'pending'" json:"status"`
	ResponseBy            string          `gorm:"size:255" json:"response_by"`
	ResponseTime          *time.Time      `json:"response_time"`
	MarketingResponseBy   string          `gorm:"size:255" json:"marketing_response_by"`
	MarketingResponseTime *time.Time      `json:"marketing_response_time"`
	RevisiGeneral         string          `gorm:"type:text" json:"revisi_general"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	DeletedAt             gorm.DeletedAt  `gorm:"index" json:"-"`

	// Relations
	Order *Order            `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Files []GambarKerjaFile `gorm:"foreignKey:GambarKerjaID" json:"files,omitempty"`
}

func (GambarKerja) TableName() string {
	return "gambar_kerja"
}
