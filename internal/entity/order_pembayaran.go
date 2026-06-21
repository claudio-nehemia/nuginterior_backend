package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderPembayaran tracks payment per termin step (entity only, no API yet).
type OrderPembayaran struct {
	ID                 uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID            uint            `gorm:"not null" json:"order_id"`
	TahapanStep        int             `gorm:"not null" json:"tahapan_step"`
	Jumlah             decimal.Decimal `gorm:"type:decimal(18,2);not null" json:"jumlah"`
	TanggalJatuhTempo  *time.Time      `json:"tanggal_jatuh_tempo,omitempty"`
	TanggalBayar       *time.Time      `json:"tanggal_bayar,omitempty"`
	Status             string          `gorm:"size:50;default:'belum_bayar'" json:"status"`
	BuktiBayar         string          `gorm:"size:500" json:"bukti_bayar"`
	Catatan            string          `gorm:"type:text" json:"catatan"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`

	Order *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

func (OrderPembayaran) TableName() string {
	return "order_pembayaran" 
}
