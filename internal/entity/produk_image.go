package entity

import (
	"time"
)

// ProdukImage represents the produk_images table.
type ProdukImage struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ProdukID  uint      `gorm:"not null" json:"produk_id"`
	Image     string    `gorm:"size:255;not null" json:"image"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ProdukImage) TableName() string {
	return "produk_images"
}
