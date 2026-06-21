package entity

import (
	"time"

	"gorm.io/gorm"
)

// Invoice represents the invoices table.
type Invoice struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ContractID   uint           `gorm:"not null;index" json:"contract_id"`
	OrderID      uint           `gorm:"not null;index" json:"order_id"`
	Step         int            `gorm:"not null" json:"step"`
	Keterangan   string         `gorm:"size:255;not null" json:"keterangan"`
	Persentase   float64        `gorm:"type:decimal(5,2);not null" json:"persentase"`
	Amount       float64        `gorm:"type:decimal(18,2);not null" json:"amount"`
	Deadline     *time.Time     `json:"deadline"`
	Status       string         `gorm:"size:50;not null;default:'belum_bayar'" json:"status"` // 'belum_bayar', 'terbayar'
	PaymentProof string         `gorm:"size:500" json:"payment_proof"`
	PaidAt       *time.Time     `json:"paid_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Contract *Contract `gorm:"foreignKey:ContractID" json:"contract,omitempty"`
	Order    *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

func (Invoice) TableName() string {
	return "invoices"
}
