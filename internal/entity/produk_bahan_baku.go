package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

// ProdukBahanBaku represents the produk_bahan_bakus pivot table.
type ProdukBahanBaku struct {
	ID          uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	ProdukID    uint            `gorm:"not null" json:"produk_id"`
	BahanBakuID uint            `gorm:"not null" json:"bahan_baku_id"`
	HargaDasar  decimal.Decimal `gorm:"type:decimal(18,2);default:0" json:"harga_dasar"`
	HargaJasa   decimal.Decimal `gorm:"type:decimal(18,2);default:0" json:"harga_jasa"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`

	// Relations
	BahanBaku BahanBaku `gorm:"foreignKey:BahanBakuID" json:"bahan_baku,omitempty"`
}

func (ProdukBahanBaku) TableName() string {
	return "produk_bahan_bakus"
}
