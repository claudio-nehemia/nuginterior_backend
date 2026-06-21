package entity

import (
	"time"

	"gorm.io/gorm"
)

// Contract represents the contracts table.
type Contract struct {
	ID                 uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	RABID              uint           `gorm:"not null;uniqueIndex" json:"rab_id"`
	OrderID            uint           `gorm:"not null;index" json:"order_id"`
	TerminID           *uint          `gorm:"index" json:"termin_id"`
	LamaKontrak        string         `gorm:"size:255" json:"lama_kontrak"`
	Status             string         `gorm:"size:50;not null;default:'belum_dibuat'" json:"status"` // 'belum_dibuat', 'draft', 'deal'
	SignedContractFile string         `gorm:"size:500" json:"signed_contract_file"`
	ResponseBy         string         `gorm:"size:255" json:"response_by"`
	ResponseTime       *time.Time     `json:"response_time"`
	InvoiceResponseBy  string         `gorm:"size:255" json:"invoice_response_by"`
	InvoiceResponseTime *time.Time    `json:"invoice_response_time"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	RAB      *RAB      `gorm:"foreignKey:RABID" json:"rab,omitempty"`
	Order    *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Termin   *Termin   `gorm:"foreignKey:TerminID" json:"termin,omitempty"`
	Invoices []Invoice `gorm:"foreignKey:ContractID;constraint:OnDelete:CASCADE" json:"invoices,omitempty"`
}

func (Contract) TableName() string {
	return "contracts"
}
