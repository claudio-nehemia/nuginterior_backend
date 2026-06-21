package entity

import (
	"time"
)

// BahanBaku represents the bahan_bakus master table.
type BahanBaku struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	NamaBahanBaku  string    `gorm:"size:255;not null" json:"nama_bahan_baku"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (BahanBaku) TableName() string {
	return "bahan_bakus"
}
