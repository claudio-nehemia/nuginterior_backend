package entity

import (
	"time"
)

// InputItem represents the parent planning record linked to DesainFinal.
type InputItem struct {
	ID                    uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	DesainFinalID         uint           `gorm:"not null;uniqueIndex" json:"desain_final_id"`
	OrderID               uint           `gorm:"not null" json:"order_id"`
	Status                string         `gorm:"type:varchar(50);default:'draft'" json:"status"` // draft, approved
	ResponseTime          *time.Time     `json:"response_time"`
	ResponseBy            string         `gorm:"type:varchar(100)" json:"response_by"`
	MarketingResponseTime *time.Time     `json:"marketing_response_time"`
	MarketingResponseBy   string         `gorm:"type:varchar(100)" json:"marketing_response_by"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`

	// Relationships
	DesainFinal *DesainFinal     `gorm:"foreignKey:DesainFinalID" json:"desain_final,omitempty"`
	Order       *Order           `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Rooms       []InputItemRoom  `gorm:"foreignKey:InputItemID;constraint:OnDelete:CASCADE" json:"rooms,omitempty"`
}

func (InputItem) TableName() string {
	return "input_items"
}

// InputItemRoom represents a room in the plan.
type InputItemRoom struct {
	ID          uint                     `gorm:"primaryKey;autoIncrement" json:"id"`
	InputItemID uint                     `gorm:"not null;index" json:"input_item_id"`
	NamaRuangan string                   `gorm:"size:255;not null" json:"nama_ruangan"`
	ProdukID    *uint                    `gorm:"null" json:"produk_id"`
	Qty         int                      `gorm:"not null;default:1" json:"qty"`
	Panjang     float64                  `gorm:"type:decimal(10,2);not null;default:0" json:"panjang"`
	Lebar       float64                  `gorm:"type:decimal(10,2);not null;default:0" json:"lebar"`
	Tinggi      float64                  `gorm:"type:decimal(10,2);not null;default:0" json:"tinggi"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`

	// Relationships
	Produk          *Produk                  `gorm:"foreignKey:ProdukID" json:"produk,omitempty"`
	BahanBakus      []InputItemRoomBahanBaku `gorm:"foreignKey:InputItemRoomID;constraint:OnDelete:CASCADE" json:"bahan_bakus,omitempty"`
	FinishingDalams []InputItemRoomFinishing `gorm:"foreignKey:InputItemRoomID;constraint:OnDelete:CASCADE" json:"finishing_dalams,omitempty"`
	FinishingLuars  []InputItemRoomFinishing `gorm:"foreignKey:InputItemRoomID;constraint:OnDelete:CASCADE" json:"finishing_luars,omitempty"`
	Aksesoris       []InputItemRoomAksesoris `gorm:"foreignKey:InputItemRoomID;constraint:OnDelete:CASCADE" json:"aksesoris,omitempty"`
}

func (InputItemRoom) TableName() string {
	return "input_item_rooms"
}

// InputItemRoomBahanBaku represents a raw material chosen for the product in a room.
type InputItemRoomBahanBaku struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	InputItemRoomID uint      `gorm:"not null;index" json:"input_item_room_id"`
	BahanBakuID     uint      `gorm:"not null" json:"bahan_baku_id"`
	CreatedAt       time.Time `json:"created_at"`

	// Relationships
	BahanBaku *BahanBaku `gorm:"foreignKey:BahanBakuID" json:"bahan_baku,omitempty"`
}

func (InputItemRoomBahanBaku) TableName() string {
	return "input_item_room_bahan_bakus"
}

// InputItemRoomFinishing represents finishing materials chosen for a room (inside or outside).
type InputItemRoomFinishing struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	InputItemRoomID uint      `gorm:"not null;index" json:"input_item_room_id"`
	ItemID          uint      `gorm:"not null" json:"item_id"`
	Type            string    `gorm:"size:50;not null" json:"type"` // "dalam" or "luar"
	Notes           string    `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Relationships
	Item *Item `gorm:"foreignKey:ItemID" json:"item,omitempty"`
}

func (InputItemRoomFinishing) TableName() string {
	return "input_item_room_finishings"
}

// InputItemRoomAksesoris represents accessory items chosen for a room.
type InputItemRoomAksesoris struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	InputItemRoomID uint      `gorm:"not null;index" json:"input_item_room_id"`
	ItemID          uint      `gorm:"not null" json:"item_id"`
	Qty             int       `gorm:"not null;default:1" json:"qty"`
	Notes           string    `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Relationships
	Item *Item `gorm:"foreignKey:ItemID" json:"item,omitempty"`
}

func (InputItemRoomAksesoris) TableName() string {
	return "input_item_room_aksesoris"
}
