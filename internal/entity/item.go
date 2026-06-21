package entity

import (
	"time"
)

// Item represents the items table.
// jenis_item can be: finishing_dalam, finishing_luar, aksesoris
type Item struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	NamaItem  string    `gorm:"size:255;not null" json:"nama_item"`
	JenisItem string    `gorm:"size:50;not null" json:"jenis_item"`
	Harga     float64   `gorm:"type:decimal(18,2);not null;default:0" json:"harga"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Item) TableName() string {
	return "items"
}

// Valid jenis_item values.
const (
	JenisFinishingDalam = "finishing_dalam"
	JenisFinishingLuar  = "finishing_luar"
	JenisAksesoris      = "aksesoris"
)

// ValidJenisItem checks if the jenis_item value is valid.
func ValidJenisItem(jenis string) bool {
	switch jenis {
	case JenisFinishingDalam, JenisFinishingLuar, JenisAksesoris:
		return true
	}
	return false
}
