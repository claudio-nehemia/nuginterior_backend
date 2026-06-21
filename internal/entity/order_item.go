package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderItem is the pivot table linking orders to items (entity only, no API yet).
type OrderItem struct {
	ID          uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID     uint            `gorm:"not null" json:"order_id"`
	ItemID      uint            `gorm:"not null" json:"item_id"`
	Qty         int             `gorm:"not null;default:1" json:"qty"`
	HargaSatuan decimal.Decimal `gorm:"type:decimal(18,2);default:0" json:"harga_satuan"`
	Catatan     string          `gorm:"type:text" json:"catatan"`
	CreatedAt   time.Time       `json:"created_at"`

	Order *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Item  *Item  `gorm:"foreignKey:ItemID" json:"item,omitempty"`
}

func (OrderItem) TableName() string {
	return "order_item"
}
