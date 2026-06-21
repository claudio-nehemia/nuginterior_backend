package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderProduk is the pivot table linking orders to produks (entity only, no API yet).
type OrderProduk struct {
	ID          uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID     uint            `gorm:"not null" json:"order_id"`
	ProdukID    uint            `gorm:"not null" json:"produk_id"`
	Qty         int             `gorm:"not null;default:1" json:"qty"`
	HargaSatuan decimal.Decimal `gorm:"type:decimal(18,2);default:0" json:"harga_satuan"`
	HargaJasa   decimal.Decimal `gorm:"type:decimal(18,2);default:0" json:"harga_jasa"`
	Catatan     string          `gorm:"type:text" json:"catatan"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`

	Order  *Order  `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Produk *Produk `gorm:"foreignKey:ProdukID" json:"produk,omitempty"`
}

func (OrderProduk) TableName() string {
	return "order_produk"
}
